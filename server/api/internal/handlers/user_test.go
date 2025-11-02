package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RowenTey/JustJio/server/api/internal/middlewares"
	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/request"
	"github.com/RowenTey/JustJio/server/api/pkg/tests"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
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
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "user_test", "file://../../migrations")
	assert.NoError(suite.T(), err)

	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	userRepository := repositories.NewUserRepository(suite.db)
	suite.userService = services.NewUserService(suite.db, userRepository, suite.logger)
	userHandler := NewUserHandler(suite.userService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middlewares.Authenticated(suite.mockJWTSecret))

	// Register User routes
	userRoutes := suite.app.Group("/users/:userId")
	userRoutes.Get("/", userHandler.GetUser)
	userRoutes.Patch("/username",
		middlewares.ParseAndValidate[request.UpdateUsernameRequest](),
		userHandler.UpdateUsername)
	userRoutes.Get("/friends", userHandler.GetFriends)
	userRoutes.Get("/friends/count", userHandler.CountFriends)
	userRoutes.Get("/friends/search", userHandler.SearchNonFriends)
	userRoutes.Post("/friends",
		middlewares.ParseAndValidate[request.SendFriendRequest](),
		userHandler.SendFriendRequest)
	userRoutes.Delete("/friends/:friendId", userHandler.RemoveFriend)
	userRoutes.Get("/friends/requests", userHandler.GetFriendRequestsByStatus)
	userRoutes.Get("/friends/requests/count", userHandler.CountPendingFriendRequests)
	userRoutes.Patch("/friends/requests/respond",
		middlewares.ParseAndValidate[request.RespondToFriendRequestRequest](),
		userHandler.RespondToFriendRequest)
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
	user := models.User{
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
	friend := models.User{
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
	friendRequest := models.FriendRequest{
		SenderID:   suite.testFriendID,
		ReceiverID: suite.testUserID,
		Status:     "pending",
	}
	result = suite.db.Create(&friendRequest)
	assert.NoError(suite.T(), result.Error)
	suite.testRequestID = friendRequest.ID

	// Update the pending friend requests count
	result = suite.db.Model(&user).Update("no_of_pending_friend_requests", 1)
	assert.NoError(suite.T(), result.Error)

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
	updateReq := request.UpdateUsernameRequest{
		Username: "updatedusername",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d/username", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify update in database
	var user models.User
	err = suite.db.First(&user, suite.testUserID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updatedusername", user.Username)
}

func (suite *UserHandlerTestSuite) TestUpdateUser_InvalidInput() {
	// Test with empty username
	updateReq := request.UpdateUsernameRequest{
		Username: "",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d/username", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestUpdateUser_DuplicateUsername() {
	// Create another user with a username
	hashedPassword, _ := utils.HashPassword("password789")
	otherUser := models.User{
		Username: "existinguser",
		Email:    "other@example.com",
		Password: hashedPassword,
	}
	err := suite.db.Create(&otherUser).Error
	assert.NoError(suite.T(), err)

	// Try to update testUser to use the existing username
	updateReq := request.UpdateUsernameRequest{
		Username: "existinguser",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d/username", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestDeleteUser_Success() {
	suite.T().Skip("DeleteUser handler not implemented yet")
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify user was deleted
	var count int64
	err = suite.db.Model(&models.User{}).Where("id = ?", suite.testUserID).Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *UserHandlerTestSuite) TestSendFriendRequest_Success() {
	// Create a new user to send request to
	newUser := models.User{
		Username: "newuser",
		Email:    "newuser@test.com",
		Password: "password789",
	}
	result := suite.db.Create(&newUser)
	assert.NoError(suite.T(), result.Error)

	requestBody := request.SendFriendRequest{
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
	err = suite.db.Model(models.FriendRequest{}).
		Where("sender_id = ? AND receiver_id = ?", suite.testUserID, newUser.ID).
		Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserHandlerTestSuite) TestSendFriendRequest_ToSelf() {
	requestBody := request.SendFriendRequest{
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

func (suite *UserHandlerTestSuite) TestSendFriendRequest_AlreadyFriends() {
	// Create another user
	hashedPassword, _ := utils.HashPassword("password789")
	friendUser := models.User{
		Username: "frienduser",
		Email:    "friend@test.com",
		Password: hashedPassword,
	}
	err := suite.db.Create(&friendUser).Error
	assert.NoError(suite.T(), err)

	// Make them friends
	var testUser models.User
	err = suite.db.First(&testUser, suite.testUserID).Error
	assert.NoError(suite.T(), err)
	err = suite.db.Model(&testUser).Association("Friends").Append(&friendUser)
	assert.NoError(suite.T(), err)

	// Try to send friend request to existing friend
	requestBody := request.SendFriendRequest{
		FriendID: friendUser.ID,
	}
	reqBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/users/%d/friends", suite.testUserID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestSendFriendRequest_PendingRequest() {
	// Create another user
	hashedPassword, _ := utils.HashPassword("password789")
	otherUser := models.User{
		Username: "pendinguser",
		Email:    "pending@test.com",
		Password: hashedPassword,
	}
	err := suite.db.Create(&otherUser).Error
	assert.NoError(suite.T(), err)

	// Create a pending friend request
	friendRequest := models.FriendRequest{
		SenderID:   suite.testUserID,
		ReceiverID: otherUser.ID,
		Status:     "pending",
	}
	err = suite.db.Create(&friendRequest).Error
	assert.NoError(suite.T(), err)

	// Try to send another friend request to the same user
	requestBody := request.SendFriendRequest{
		FriendID: otherUser.ID,
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
	err := suite.userService.AcceptFriendRequest(context.Background(), suite.testRequestID)
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

func (suite *UserHandlerTestSuite) TestRemoveFriend_NotFriends() {
	// Create a user who is not a friend
	hashedPassword, _ := utils.HashPassword("password789")
	nonFriend := models.User{
		Username: "nonfriend",
		Email:    "nonfriend@test.com",
		Password: hashedPassword,
	}
	err := suite.db.Create(&nonFriend).Error
	assert.NoError(suite.T(), err)

	// Try to remove non-friend
	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/users/%d/friends/%d", suite.testUserID, nonFriend.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestRemoveFriend_UserNotFound() {
	// Try to remove friend with non-existent friend ID
	nonExistentID := uint(99999)

	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/users/%d/friends/%d", suite.testUserID, nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestGetFriends_Success() {
	// First make the users friends
	err := suite.userService.AcceptFriendRequest(context.Background(), suite.testRequestID)
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
	suite.T().Skip("IsFriend handler not implemented yet")
	// First make the users friends
	err := suite.userService.AcceptFriendRequest(context.Background(), suite.testRequestID)
	assert.NoError(suite.T(), err)

	requestBody := request.SendFriendRequest{
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
	err := suite.userService.AcceptFriendRequest(context.Background(), suite.testRequestID)
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
	assert.Equal(suite.T(), float64(1), responseBody["data"].(float64))
}

func (suite *UserHandlerTestSuite) TestSearchFriends_Success() {
	// This endpoint searches for NON-friends, not friends
	// Refresh the materialized view to include the test users
	err := suite.db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY user_non_friends").Error
	assert.NoError(suite.T(), err)

	// The friend request exists but hasn't been accepted yet, so testfriend should appear in search
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/search?query=test", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	// Should return 1 result (testfriend) since they're not friends yet
	assert.Len(suite.T(), responseBody["data"].([]any), 1)
}

func (suite *UserHandlerTestSuite) TestSearchFriends_EmptyQuery() {
	// Refresh the materialized view
	err := suite.db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY user_non_friends").Error
	assert.NoError(suite.T(), err)

	// Test with empty query string
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/search?query=", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	// Should return 200 with empty or all results depending on implementation
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestSearchFriends_NoResults() {
	// Refresh the materialized view
	err := suite.db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY user_non_friends").Error
	assert.NoError(suite.T(), err)

	// Search for a username that doesn't exist
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/search?query=nonexistentuser12345", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(responseBody["data"].([]any)))
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

func (suite *UserHandlerTestSuite) TestGetFriendRequestsByStatus_InvalidStatus() {
	// Test with invalid status value
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/users/%d/friends/requests?status=invalid_status", suite.testUserID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)
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
	assert.Equal(suite.T(), float64(1), responseBody["data"].(float64))
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
	var request models.FriendRequest
	err = suite.db.First(&request, suite.testRequestID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "rejected", request.Status)
}

func (suite *UserHandlerTestSuite) TestRespondToFriendRequest_NotFound() {
	// Test with non-existent request ID
	respondReq := request.RespondToFriendRequestRequest{
		RequestID: uint(99999),
		Action:    "accept",
	}
	reqBody, _ := json.Marshal(respondReq)

	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/users/%d/friends/requests/respond", suite.testFriendID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testFriendToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *UserHandlerTestSuite) TestRespondToFriendRequest_AlreadyResponded() {
	// First accept the request
	err := suite.userService.AcceptFriendRequest(context.Background(), suite.testRequestID)
	assert.NoError(suite.T(), err)

	// Try to accept the same request again
	respondReq := request.RespondToFriendRequestRequest{
		RequestID: suite.testRequestID,
		Action:    "accept",
	}
	reqBody, _ := json.Marshal(respondReq)

	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/users/%d/friends/requests/respond", suite.testFriendID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testFriendToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)
}
