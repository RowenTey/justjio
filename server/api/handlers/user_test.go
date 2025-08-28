package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	mockJWTSecret string

	userService *services.UserService

	testUserID      uint
	testUserToken   string
	testFriendID    uint
	testFriendToken string
	testRequestID   uint
}

func (suite *UserHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "user_test")
	assert.NoError(suite.T(), err)

	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	userRepository := repository.NewUserRepository(suite.db)
	suite.userService = services.NewUserService(suite.db, userRepository, suite.logger)
	userHandler := NewUserHandler(suite.userService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(suite.mockJWTSecret))

	// Register User routes
	userRoutes := suite.app.Group("/users/:userId")
	userRoutes.Get("/", userHandler.GetUser)
	userRoutes.Patch("/", userHandler.UpdateUser)
	userRoutes.Delete("/", userHandler.DeleteUser)
	userRoutes.Get("/friends", userHandler.GetFriends)
	userRoutes.Get("/friends/count", userHandler.GetNumFriends)
	userRoutes.Get("/friends/search", userHandler.SearchFriends)
	userRoutes.Post("/friends", userHandler.SendFriendRequest)
	userRoutes.Delete("/friends/:friendId", userHandler.RemoveFriend)
	userRoutes.Get("/friends/check", userHandler.IsFriend)
	userRoutes.Get("/friends/requests", userHandler.GetFriendRequestsByStatus)
	userRoutes.Get("/friends/requests/count", userHandler.CountPendingFriendRequests)
	userRoutes.Patch("/friends/requests/respond", userHandler.RespondToFriendRequest)
}

func (suite *UserHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	suite.logger.Info("Tore down test suite dependencies")
}

func (suite *UserHandlerTestSuite) SetupTest() {
	// Create test user
	hashedPassword1, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)
	user := model.User{
		Username: "testuser",
		Email:    "user@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&user)
	assert.NoError(suite.T(), result.Error)
	suite.testUserID = user.ID
	userToken, err := tests.GenerateTestToken(user.ID, user.Username, user.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testUserToken = userToken

	// Create test friend
	hashedPassword2, err := utils.HashPassword("password456")
	assert.NoError(suite.T(), err)
	friend := model.User{
		Username: "testfriend",
		Email:    "friend@example.com",
		Password: hashedPassword2,
	}
	result = suite.db.Create(&friend)
	assert.NoError(suite.T(), result.Error)
	suite.testFriendID = friend.ID
	friendToken, err := tests.GenerateTestToken(friend.ID, friend.Username, friend.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testFriendToken = friendToken

	// Create test friend request
	friendRequest := model.FriendRequest{
		SenderID:   suite.testFriendID,
		ReceiverID: suite.testUserID,
		Status:     "pending",
	}
	result = suite.db.Create(&friendRequest)
	assert.NoError(suite.T(), result.Error)
	suite.testRequestID = friendRequest.ID

	suite.logger.Infof("SetupTest complete: User ID=%d, Friend ID=%d, Request ID=%d",
		suite.testUserID, suite.testFriendID, suite.testRequestID)
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE friend_requests RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE user_friends RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	suite.logger.Info("Tore down test data")
}

func TestUserHandlerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UserHandlerTestSuite))
}

func (suite *UserHandlerTestSuite) TestGetUser_Success() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User found successfully", responseBody["message"])

	userData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), "testuser", userData["username"])
}

func (suite *UserHandlerTestSuite) TestGetUser_NotFound() {
	nonExistentUserID := "999999"
	req := httptest.NewRequest(http.MethodGet, "/users/"+nonExistentUserID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestUpdateUser_Success() {
	updateReq := request.UpdateUserRequest{
		Field: "username",
		Value: "updatedusername",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify update in database
	var user model.User
	err = suite.db.First(&user, suite.testUserID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updatedusername", user.Username)
}

func (suite *UserHandlerTestSuite) TestDeleteUser_Success() {
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify user was deleted
	var count int64
	err = suite.db.Model(&model.User{}).Where("id = ?", suite.testUserID).Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *UserHandlerTestSuite) TestSendFriendRequest_Success() {
	// Create a new user to send request to
	newUser := model.User{
		Username: "newuser",
		Email:    "newuser@test.com",
		Password: "password789",
	}
	result := suite.db.Create(&newUser)
	assert.NoError(suite.T(), result.Error)

	requestBody := request.ModifyFriendRequest{
		FriendID: newUser.ID,
	}
	reqBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/users/%d/friends", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify request was created
	var count int64
	err = suite.db.Model(model.FriendRequest{}).
		Where("sender_id = ? AND receiver_id = ?", suite.testUserID, newUser.ID).
		Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserHandlerTestSuite) TestSendFriendRequest_ToSelf() {
	requestBody := request.ModifyFriendRequest{
		FriendID: suite.testUserID,
	}
	reqBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/users/%d/friends", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestRemoveFriend_Success() {
	// First make the users friends
	err := suite.userService.AcceptFriendRequest(suite.testRequestID)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/users/%d/friends/%d", suite.testUserID, suite.testFriendID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify they are no longer friends
	var count int64
	suite.db.Table("user_friends").
		Where("user_id = ? AND friend_id = ?", suite.testUserID, suite.testFriendID).
		Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *UserHandlerTestSuite) TestGetFriends_Success() {
	// First make the users friends
	err := suite.userService.AcceptFriendRequest(suite.testRequestID)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Friends retrieved successfully", responseBody["message"])
	assert.Len(suite.T(), responseBody["data"].([]any), 1)
}

func (suite *UserHandlerTestSuite) TestIsFriend_True() {
	// First make the users friends
	err := suite.userService.AcceptFriendRequest(suite.testRequestID)
	assert.NoError(suite.T(), err)

	requestBody := request.ModifyFriendRequest{
		FriendID: suite.testFriendID,
	}
	reqBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/check", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), responseBody["data"].(map[string]any)["isFriend"].(bool))
}

func (suite *UserHandlerTestSuite) TestGetNumFriends_Success() {
	// First make the users friends
	err := suite.userService.AcceptFriendRequest(suite.testRequestID)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/count", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(1), responseBody["data"].(map[string]any)["numFriends"].(float64))
}

func (suite *UserHandlerTestSuite) TestSearchFriends_Success() {
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/search?query=test", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), responseBody["data"].([]any), 1)
}

func (suite *UserHandlerTestSuite) TestGetFriendRequestsByStatus_Success() {
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/requests?status=pending", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), responseBody["data"].([]any), 1)
}

func (suite *UserHandlerTestSuite) TestCountPendingFriendRequests_Success() {
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/requests/count", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(1), responseBody["data"].(map[string]any)["count"].(float64))
}

func (suite *UserHandlerTestSuite) TestRespondToFriendRequest_Accept() {
	respondReq := request.RespondToFriendRequestRequest{
		RequestID: suite.testRequestID,
		Action:    "accept",
	}
	reqBody, _ := json.Marshal(respondReq)

	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/users/%d/friends/requests/respond", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testFriendToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify they are now friends
	var count int64
	suite.db.Table("user_friends").
		Where("user_id = ? AND friend_id = ?", suite.testUserID, suite.testFriendID).
		Count(&count)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserHandlerTestSuite) TestRespondToFriendRequest_Reject() {
	respondReq := request.RespondToFriendRequestRequest{
		RequestID: suite.testRequestID,
		Action:    "reject",
	}
	reqBody, _ := json.Marshal(respondReq)

	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/users/%d/friends/requests/respond", suite.testFriendID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testFriendToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify request was rejected
	var request model.FriendRequest
	err = suite.db.First(&request, suite.testRequestID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "rejected", request.Status)
}
