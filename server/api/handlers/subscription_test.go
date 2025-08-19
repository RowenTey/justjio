package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	pushNotificationsModel "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SubscriptionHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	mockJWTSecret string

	subscriptionService *services.SubscriptionService

	testNotificationChan chan pushNotificationsModel.NotificationData

	testUserID       uint
	testUserToken    string
	testEndpoint     string
	testSubscription *model.Subscription
}

func (suite *SubscriptionHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "sub_test")
	assert.NoError(suite.T(), err)

	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	suite.testNotificationChan = make(chan pushNotificationsModel.NotificationData, 100)
	subscriptionRepo := repository.NewSubscriptionRepository(suite.db)
	suite.subscriptionService = services.NewSubscriptionService(
		subscriptionRepo,
		suite.testNotificationChan,
		suite.logger,
	)
	subscriptionHandler := NewSubscriptionHandler(suite.subscriptionService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(suite.mockJWTSecret))

	// Register Subscription routes
	subscriptionRoutes := suite.app.Group("/subscriptions")
	subscriptionRoutes.Post("/", subscriptionHandler.CreateSubscription)
	subscriptionRoutes.Get("/:endpoint", subscriptionHandler.GetSubscriptionByEndpoint)
	subscriptionRoutes.Delete("/:subId", subscriptionHandler.DeleteSubscription)
}

func (suite *SubscriptionHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	suite.logger.Info("Tore down test suite dependencies")
	close(suite.testNotificationChan)
}

func (suite *SubscriptionHandlerTestSuite) SetupTest() {
	hashedPassword, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)
	user := model.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: hashedPassword,
	}
	result := suite.db.Create(&user)
	assert.NoError(suite.T(), result.Error)
	suite.testUserID = user.ID

	// Generate JWT token for the user
	token, err := tests.GenerateTestToken(user.ID, user.Username, user.Email, suite.mockJWTSecret)
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
	suite.db.Exec("TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	suite.logger.Info("Tore down test data")
}

func TestSubscriptionHandlerSuite(t *testing.T) {
	t.Parallel()
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
	case notification := <-suite.testNotificationChan:
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
