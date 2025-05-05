package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	model_push_notifications "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SubscriptionHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	dependencies *tests.TestDependencies

	testUserID       uint
	testUserToken    string
	testNotifChan    chan model_push_notifications.NotificationData
	testEndpoint     string
	testSubscription *model.Subscription
}

func (suite *SubscriptionHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println("Test DB Connection String:", pgConnStr)

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(mockJWTSecret))

	// Notification channel for testing
	suite.testNotifChan = make(chan model_push_notifications.NotificationData, 100)

	// Register Subscription routes
	subscriptionRoutes := suite.app.Group("/subscriptions")
	subscriptionRoutes.Post("/", func(c *fiber.Ctx) error {
		return CreateSubscription(c, suite.testNotifChan)
	})
	subscriptionRoutes.Get("/:endpoint", GetSubscriptionByEndpoint)
	subscriptionRoutes.Delete("/:subId", DeleteSubscription)
}

func (suite *SubscriptionHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	log.Info("Tore down test suite dependencies")
	close(suite.testNotifChan)
}

func (suite *SubscriptionHandlerTestSuite) SetupTest() {
	// Assign the test DB to the global variable used by handlers/services
	database.DB = suite.db
	assert.NotNil(suite.T(), database.DB, "Global DB should be set")

	// Create test user
	hashedPassword, _ := utils.HashPassword("password123")
	user := model.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: hashedPassword,
	}
	result := suite.db.Create(&user)
	assert.NoError(suite.T(), result.Error)
	suite.testUserID = user.ID

	// Generate JWT token for the user
	token, err := generateTestToken(user.ID, user.Username, user.Email)
	assert.NoError(suite.T(), err)
	suite.testUserToken = token

	// Test endpoint and subscription data
	suite.testEndpoint = "https://example.com/push/123"
	testSubscription := model.Subscription{
		UserID:   suite.testUserID,
		Endpoint: suite.testEndpoint,
		Auth:     "test_auth_key",
		P256dh:   "test_p256dh_key",
	}

	// Store test subscription in DB for some tests
	result = suite.db.Create(&testSubscription)
	assert.NoError(suite.T(), result.Error)
	suite.testSubscription = &testSubscription

	log.Infof("SetupTest complete: User ID=%d, Subscription ID=%s", suite.testUserID, suite.testSubscription.ID)
}

func (suite *SubscriptionHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE subscriptions CASCADE")
	suite.db.Exec("TRUNCATE TABLE users CASCADE")

	// Reset the global DB variable
	database.DB = nil
	log.Info("Tore down test data and reset global DB")
}

func TestSubscriptionHandlerSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionHandlerTestSuite))
}

func (suite *SubscriptionHandlerTestSuite) TestCreateSubscription_Success() {
	newSubscription := model.Subscription{
		UserID:   suite.testUserID,
		Endpoint: "https://example.com/push/new",
		P256dh:   "new_p256dh_key",
		Auth:     "new_auth_key",
	}
	reqBody, _ := json.Marshal(newSubscription)

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Subscription created successfully", responseBody["message"])

	// Verify subscription was created in database
	var subscription model.Subscription
	err = suite.db.Where("endpoint = ?", newSubscription.Endpoint).First(&subscription).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newSubscription.P256dh, subscription.P256dh)

	// Verify welcome notification was sent
	select {
	case notification := <-suite.testNotifChan:
		assert.Equal(suite.T(), "Welcome", notification.Title)
		assert.Equal(suite.T(), "Subscribed to JustJio! You will now receive notifications for app events.", notification.Message)
	case <-time.After(1 * time.Second):
		suite.T().Error("Expected welcome notification but none received")
	}
}

func (suite *SubscriptionHandlerTestSuite) TestCreateSubscription_InvalidInput() {
	// Test with missing required fields
	invalidSubscription := map[string]any{
		"userId": suite.testUserID,
		"p256dh": "test_key",
	}
	reqBody, _ := json.Marshal(invalidSubscription)

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])
}

func (suite *SubscriptionHandlerTestSuite) TestGetSubscriptionByEndpoint_Success() {
	encodedEndpoint := url.QueryEscape(suite.testSubscription.Endpoint)
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+encodedEndpoint, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Subscription retrieved successfully", responseBody["message"])

	subscriptionData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), suite.testSubscription.Endpoint, subscriptionData["endpoint"])
}

func (suite *SubscriptionHandlerTestSuite) TestGetSubscriptionByEndpoint_NotFound() {
	nonExistentEndpoint := url.QueryEscape("https://nonexistent.endpoint")
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+nonExistentEndpoint, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *SubscriptionHandlerTestSuite) TestDeleteSubscription_Success() {
	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+suite.testSubscription.ID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Subscription deleted successfully", responseBody["message"])

	// Verify subscription was deleted from database
	var subscription model.Subscription
	err = suite.db.Where("id = ?", suite.testSubscription.ID).First(&subscription).Error
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *SubscriptionHandlerTestSuite) TestDeleteSubscription_NotFound() {
	nonExistentSubId := uuid.NewString()
	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+nonExistentSubId, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}
