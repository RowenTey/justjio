package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type NotificationHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	dependencies *tests.TestDependencies

	testUserID    uint
	testUserToken string
	testNotifChan chan NotificationData
}

func (suite *NotificationHandlerTestSuite) SetupSuite() {
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
	suite.testNotifChan = make(chan NotificationData, 100)

	// Register Notification routes
	notifRoutes := suite.app.Group("/notifications")
	notifRoutes.Get("/", GetNotifications)
	notifRoutes.Post("/", func(c *fiber.Ctx) error {
		return CreateNotification(c, suite.testNotifChan)
	})
	userNotifRoutes := suite.app.Group("/users/:userId/notifications")
	userNotifRoutes.Get("/:id", GetNotification)
	userNotifRoutes.Patch("/:id", MarkNotificationAsRead)
}

func (suite *NotificationHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	log.Info("Tore down test suite dependencies")
	close(suite.testNotifChan)
}

func (suite *NotificationHandlerTestSuite) SetupTest() {
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

	log.Infof("SetupTest complete: User ID=%d", suite.testUserID)
}

func (suite *NotificationHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE notifications CASCADE")
	suite.db.Exec("TRUNCATE TABLE users CASCADE")

	// Reset the global DB variable
	database.DB = nil
	log.Info("Tore down test data and reset global DB")
}

func TestNotificationHandlerSuite(t *testing.T) {
	suite.Run(t, new(NotificationHandlerTestSuite))
}

func (suite *NotificationHandlerTestSuite) TestCreateNotification_Success() {
	createReq := request.CreateNotificationRequest{
		UserId:  suite.testUserID,
		Title:   "Test Notification",
		Content: "This is a test notification",
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Notification created successfully", responseBody["message"])

	// Verify notification in database
	var notification model.Notification
	err = suite.db.Where("user_id = ? AND title = ?", suite.testUserID, createReq.Title).First(&notification).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createReq.Content, notification.Content)
	assert.Equal(suite.T(), suite.testUserID, notification.UserID)
	assert.False(suite.T(), notification.IsRead)

	// Verify notification was sent to channel (though we don't have subscriptions in this test)
	select {
	case <-suite.testNotifChan:
		// Notification was sent to channel
	case <-time.After(100 * time.Millisecond):
		// No notification sent (expected since we don't have subscriptions)
	}
}

func (suite *NotificationHandlerTestSuite) TestCreateNotification_InvalidInput() {
	// Test empty content
	createReq := request.CreateNotificationRequest{
		UserId:  suite.testUserID,
		Title:   "Empty Content",
		Content: "",
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])

	// Test bad JSON
	reqBody = []byte(`{"userId": "notanumber", "title": "Bad JSON", "content": "Test"}`)
	req = httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err = suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)
}

func (suite *NotificationHandlerTestSuite) TestGetNotification_Success() {
	// Create a test notification
	notification := model.Notification{
		UserID:  suite.testUserID,
		Title:   "Test Get",
		Content: "Test notification content",
	}
	err := suite.db.Create(&notification).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/notifications/%d", suite.testUserID, notification.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved notification successfully", responseBody["message"])

	notificationData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), notification.Content, notificationData["content"])
	assert.Equal(suite.T(), notification.Title, notificationData["title"])
}

func (suite *NotificationHandlerTestSuite) TestGetNotification_NotFound() {
	nonExistentID := uint(9999)
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/notifications/%d", suite.testUserID, nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Notification not found", responseBody["message"])
}

func (suite *NotificationHandlerTestSuite) TestGetNotifications_Success() {
	// Create test notifications
	notifications := []model.Notification{
		{
			UserID:  suite.testUserID,
			Title:   "Notification 1",
			Content: "Content 1",
			IsRead:  false,
		},
		{
			UserID:  suite.testUserID,
			Title:   "Notification 2",
			Content: "Content 2",
			IsRead:  true,
		},
	}
	err := suite.db.Create(&notifications).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved notifications successfully", responseBody["message"])

	notificationData := responseBody["data"].([]any)
	assert.Len(suite.T(), notificationData, 2)
	// Verify notification titles
	assert.Equal(suite.T(), "Notification 1", notificationData[0].(map[string]any)["title"])
	assert.Equal(suite.T(), "Notification 2", notificationData[1].(map[string]any)["title"])
}

func (suite *NotificationHandlerTestSuite) TestGetNotifications_Empty() {
	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved notifications successfully", responseBody["message"])

	notificationData := responseBody["data"].([]any)
	assert.Empty(suite.T(), notificationData)
}

func (suite *NotificationHandlerTestSuite) TestMarkNotificationAsRead_Success() {
	// Create an unread notification
	notification := model.Notification{
		UserID:  suite.testUserID,
		Title:   "Unread Notification",
		Content: "Please read me",
		IsRead:  false,
	}
	err := suite.db.Create(&notification).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/users/%d/notifications/%d", suite.testUserID, notification.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Notification marked as read successfully", responseBody["message"])

	// Verify notification is now marked as read
	var updatedNotification model.Notification
	err = suite.db.First(&updatedNotification, notification.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedNotification.IsRead)
}

func (suite *NotificationHandlerTestSuite) TestMarkNotificationAsRead_NotFound() {
	nonExistentID := uint(9999)
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/users/%d/notifications/%d", suite.testUserID, nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Notification not found", responseBody["message"])
}
