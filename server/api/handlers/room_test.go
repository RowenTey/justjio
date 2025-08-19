package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RoomHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	mockJWTSecret string

	mockHttpClient *utils.MockHTTPClient

	roomService *services.RoomService

	testHostID    uint
	testHostToken string
	testUserID    uint
	testUserToken string
	testRoomID    string
	testRoom      *model.Room
	testInviteID  uint
}

func (suite *RoomHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	suite.mockHttpClient = new(utils.MockHTTPClient)
	roomRepository := repository.NewRoomRepository(suite.db)
	userRepository := repository.NewUserRepository(suite.db)
	billRepository := repository.NewBillRepository(suite.db)
	suite.roomService = services.NewRoomService(
		suite.db,
		roomRepository,
		userRepository,
		billRepository,
		suite.mockHttpClient,
		"test-api-key",
		suite.logger,
	)
	roomHandler := NewRoomHandler(suite.roomService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(suite.mockJWTSecret))

	// Register Room routes
	roomRoutes := suite.app.Group("/rooms")
	roomRoutes.Get("/", roomHandler.GetRooms)
	roomRoutes.Get("/count", roomHandler.GetNumRooms)
	roomRoutes.Get("/invites", roomHandler.GetRoomInvitations)
	roomRoutes.Get("/invites/count", roomHandler.GetNumRoomInvitations)
	roomRoutes.Get("/:roomId", roomHandler.GetRoom)
	roomRoutes.Get("/:roomId/attendees", roomHandler.GetRoomAttendees)
	roomRoutes.Get("/:roomId/uninvited", roomHandler.GetUninvitedFriendsForRoom)
	roomRoutes.Post("/", roomHandler.CreateRoom)
	roomRoutes.Post("/:roomId", roomHandler.InviteUser)
	roomRoutes.Patch("/:roomId", roomHandler.RespondToRoomInvite)
	roomRoutes.Patch("/:roomId/edit", roomHandler.EditRoom)
	roomRoutes.Patch("/:roomId/close", roomHandler.CloseRoom)
	roomRoutes.Patch("/:roomId/join", roomHandler.JoinRoom)
	roomRoutes.Patch("/:roomId/leave", roomHandler.LeaveRoom)
}

func (suite *RoomHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	suite.logger.Info("Tore down test suite dependencies")
}

func (suite *RoomHandlerTestSuite) SetupTest() {
	// Create test host user
	hashedPassword1, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)
	host := model.User{
		Username: "hostuser",
		Email:    "host@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&host)
	assert.NoError(suite.T(), result.Error)
	suite.testHostID = host.ID
	hostToken, err := tests.GenerateTestToken(host.ID, host.Username, host.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testHostToken = hostToken

	// Create test regular user
	hashedPassword2, err := utils.HashPassword("password456")
	assert.NoError(suite.T(), err)
	user := model.User{
		Username: "testuser",
		Email:    "user@example.com",
		Password: hashedPassword2,
	}
	result = suite.db.Create(&user)
	assert.NoError(suite.T(), result.Error)
	suite.testUserID = user.ID
	userToken, err := tests.GenerateTestToken(user.ID, user.Username, user.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testUserToken = userToken

	// Create test room
	room := model.Room{
		ID:     uuid.NewString(),
		Name:   "Test Room",
		HostID: suite.testHostID,
		Users:  []model.User{host},
	}
	result = suite.db.Create(&room)
	assert.NoError(suite.T(), result.Error)
	suite.testRoomID = room.ID
	suite.testRoom = &room

	// Create test invite
	invite := model.RoomInvite{
		RoomID:    suite.testRoomID,
		UserID:    suite.testUserID,
		InviterID: suite.testHostID,
		Status:    "pending",
	}
	result = suite.db.Create(&invite)
	assert.NoError(suite.T(), result.Error)
	suite.testInviteID = invite.ID

	suite.logger.Infof("SetupTest complete: Host ID=%d, User ID=%d, Room ID=%s",
		suite.testHostID, suite.testUserID, suite.testRoomID)
}

func (suite *RoomHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE room_invites RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	suite.logger.Info("Tore down test data")
}

func TestRoomHandlerSuite(t *testing.T) {
	suite.Run(t, new(RoomHandlerTestSuite))
}

func (suite *RoomHandlerTestSuite) TestGetRoom_Success() {
	req := httptest.NewRequest(http.MethodGet, "/rooms/"+suite.testRoomID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved room successfully", responseBody["message"])

	roomData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), suite.testRoom.Name, roomData["name"])
}

func (suite *RoomHandlerTestSuite) TestGetRoom_NotFound() {
	nonExistentRoomID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/rooms/"+nonExistentRoomID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *RoomHandlerTestSuite) TestGetRooms_Success() {
	// Create a second room for the host
	room2 := model.Room{
		ID:     uuid.NewString(),
		Name:   "Second Room",
		HostID: suite.testHostID,
		Users:  []model.User{{ID: suite.testHostID}},
	}
	err := suite.db.Create(&room2).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved rooms successfully", responseBody["message"])
	assert.Len(suite.T(), responseBody["data"].([]any), 2)
}

func (suite *RoomHandlerTestSuite) TestGetNumRooms_Success() {
	req := httptest.NewRequest(http.MethodGet, "/rooms/count", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved number of rooms successfully", responseBody["message"])
	assert.Equal(suite.T(), 1, int(responseBody["data"].(map[string]any)["count"].(float64)))
}

func (suite *RoomHandlerTestSuite) TestGetRoomInvitations_Success() {
	req := httptest.NewRequest(http.MethodGet, "/rooms/invites", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved room invitations successfully", responseBody["message"])

	invitationData := responseBody["data"].([]any)
	assert.Len(suite.T(), invitationData, 1)
	assert.Equal(suite.T(), suite.testInviteID, uint(invitationData[0].(map[string]any)["id"].(float64)))
}

func (suite *RoomHandlerTestSuite) TestGetNumRoomInvitations_Success() {
	req := httptest.NewRequest(http.MethodGet, "/rooms/invites/count", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved number of invitations successfully", responseBody["message"])
	assert.Equal(suite.T(), 1, int(responseBody["data"].(map[string]any)["count"].(float64)))
}

func (suite *RoomHandlerTestSuite) TestGetRoomAttendees_Success() {
	req := httptest.NewRequest(http.MethodGet, "/rooms/"+suite.testRoomID+"/attendees", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved room attendees successfully", responseBody["message"])

	userData := responseBody["data"].([]any)
	assert.Len(suite.T(), userData, 1)
	assert.Equal(suite.T(), suite.testHostID, uint(userData[0].(map[string]any)["id"].(float64)))
}

func (suite *RoomHandlerTestSuite) TestCreateRoom_Success() {
	invitees := []string{fmt.Sprintf("%d", suite.testUserID)}
	inviteesJSON, _ := json.Marshal(invitees)

	placeId := "ChIJN1t_tDeuEmsRUsoyG83frY4"
	expectedUri := "https://maps.google.com/?cid=123456789"

	// Create a mock response
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(
			`{"googleMapsUri": "` + expectedUri + `"}`,
		)),
		Header: make(http.Header),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	suite.mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" &&
			req.URL.String() == fmt.Sprintf("https://places.googleapis.com/v1/places/%s", placeId) &&
			req.Header.Get("X-Goog-Api-Key") == "test-api-key" &&
			req.Header.Get("X-Goog-FieldMask") == "googleMapsUri"
	})).Return(mockResponse, nil)

	createReq := request.CreateRoomRequest{
		Room: model.Room{
			Name:         "New Test Room",
			VenuePlaceId: placeId,
		},
		InviteesId: datatypes.JSON(inviteesJSON),
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Created room successfully", responseBody["message"])

	roomData := responseBody["data"].(map[string]any)["room"].(map[string]any)
	assert.Equal(suite.T(), "New Test Room", roomData["name"])

	invitesData := responseBody["data"].(map[string]any)["invites"].([]any)
	assert.Len(suite.T(), invitesData, 1)
}

func (suite *RoomHandlerTestSuite) TestInviteUser_Success() {
	// Create a new user to invite
	newUser := model.User{
		Username: "newuser",
		Email:    "newuser@test.com",
		Password: "password789",
	}
	result := suite.db.Create(&newUser)
	assert.NoError(suite.T(), result.Error)

	invitees := []string{fmt.Sprintf("%d", newUser.ID)}
	inviteesJSON, _ := json.Marshal(invitees)

	inviteReq := request.InviteUserRequest{
		InviteesId: datatypes.JSON(inviteesJSON),
	}
	reqBody, _ := json.Marshal(inviteReq)

	req := httptest.NewRequest(http.MethodPost,
		"/rooms/"+suite.testRoomID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invited users successfully", responseBody["message"])
	assert.Len(suite.T(), responseBody["data"].([]any), 1)
}

func (suite *RoomHandlerTestSuite) TestJoinRoom_Success() {
	req := httptest.NewRequest(http.MethodPatch,
		"/rooms/"+suite.testRoomID+"/join", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Joined room successfully", responseBody["message"])
	assert.Equal(suite.T(), suite.testRoomID, responseBody["data"].(map[string]any)["room"].(map[string]any)["id"])
	assert.Len(suite.T(), responseBody["data"].(map[string]any)["attendees"].([]any), 2) // Host + new user
}

func (suite *RoomHandlerTestSuite) TestRespondToRoomInvite_Accept() {
	respondReq := request.RespondToRoomInviteRequest{
		Accept: true,
	}
	reqBody, _ := json.Marshal(respondReq)

	req := httptest.NewRequest(http.MethodPatch,
		"/rooms/"+suite.testRoomID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Joined room successfully", responseBody["message"])
	assert.Equal(suite.T(), suite.testRoomID, responseBody["data"].(map[string]any)["room"].(map[string]any)["id"])
}

func (suite *RoomHandlerTestSuite) TestRespondToRoomInvite_Reject() {
	respondReq := request.RespondToRoomInviteRequest{
		Accept: false,
	}
	reqBody, _ := json.Marshal(respondReq)

	req := httptest.NewRequest(http.MethodPatch,
		"/rooms/"+suite.testRoomID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify invite was updated
	var invite model.RoomInvite
	err = suite.db.Where("id = ?", suite.testInviteID).First(&invite).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "rejected", invite.Status)
}

func (suite *RoomHandlerTestSuite) TestLeaveRoom_Success() {
	// First have the user join the room
	_, _, err := suite.roomService.JoinRoom(suite.testRoomID, fmt.Sprintf("%d", suite.testUserID))
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPatch,
		"/rooms/"+suite.testRoomID+"/leave", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify user was removed
	var count int64
	err = suite.db.Table("room_users").
		Where("room_id = ? AND user_id = ?", suite.testRoomID, suite.testUserID).
		Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RoomHandlerTestSuite) TestEditRoom_VenueChanged_Success() {
	updateReq := request.UpdateRoomRequest{
		Venue:       "New Awesome Place",
		PlaceId:     "newPlaceId123",
		Date:        time.Now(),
		Time:        "19:00:00",
		Description: "Updated event description.",
	}
	reqBody, _ := json.Marshal(updateReq)

	expectedUri := "https://maps.google.com/?cid=987654321"

	// Mock the Google Places API call
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"googleMapsUri": "` + expectedUri + `"}`)),
		Header:     make(http.Header),
	}
	mockResponse.Header.Set("Content-Type", "application/json")
	suite.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil).Once()

	req := httptest.NewRequest(http.MethodPatch, "/rooms/"+suite.testRoomID+"/edit", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Edited room successfully", responseBody["message"])

	roomData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), updateReq.Description, roomData["description"])
	assert.Equal(suite.T(), updateReq.Venue, roomData["venue"])
	assert.Equal(suite.T(), expectedUri, roomData["venueUrl"])

	suite.mockHttpClient.AssertExpectations(suite.T())
}

func (suite *RoomHandlerTestSuite) TestEditRoom_VenueNotChanged_Success() {
	updateReq := request.UpdateRoomRequest{
		PlaceId:     suite.testRoom.VenuePlaceId, // Same PlaceId
		Date:        time.Now(),
		Time:        "20:00:00",
		Description: "Only changing the date and time.",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPatch, "/rooms/"+suite.testRoomID+"/edit", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)

	roomData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), "Only changing the date and time.", roomData["description"])

	// Ensure the HTTP client was NOT called since PlaceId didn't change
	suite.mockHttpClient.AssertNotCalled(suite.T(), "Do")
}

func (suite *RoomHandlerTestSuite) TestEditRoom_NotHost() {
	updateReq := request.UpdateRoomRequest{Description: "Attempt by non-host"}
	reqBody, _ := json.Marshal(updateReq)

	// Use the non-host user's token
	req := httptest.NewRequest(http.MethodPatch, "/rooms/"+suite.testRoomID+"/edit", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusUnauthorized, resp.StatusCode)
}

func (suite *RoomHandlerTestSuite) TestEditRoom_RoomNotFound() {
	nonExistentRoomID := uuid.NewString()
	updateReq := request.UpdateRoomRequest{Description: "Attempt on non-existent room"}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPatch, "/rooms/"+nonExistentRoomID+"/edit", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)
}

func (suite *RoomHandlerTestSuite) TestEditRoom_InvalidBody() {
	req := httptest.NewRequest(http.MethodPatch, "/rooms/"+suite.testRoomID+"/edit", bytes.NewBufferString(`{"description": "bad json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)
}

func (suite *RoomHandlerTestSuite) TestEditRoom_GoogleAPIFailure() {
	updateReq := request.UpdateRoomRequest{
		PlaceId: "newPlaceIdThatWillFail",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Mock the Google Places API call to return an error
	suite.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{}, fmt.Errorf("google api is down")).Once()

	req := httptest.NewRequest(http.MethodPatch, "/rooms/"+suite.testRoomID+"/edit", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusInternalServerError, resp.StatusCode)

	suite.mockHttpClient.AssertExpectations(suite.T())
}

func (suite *RoomHandlerTestSuite) TestCloseRoom_Success() {
	req := httptest.NewRequest(http.MethodPatch,
		"/rooms/"+suite.testRoomID+"/close", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testHostToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify room was closed
	var room model.Room
	err = suite.db.Where("id = ?", suite.testRoomID).First(&room).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), room.IsClosed)
}
