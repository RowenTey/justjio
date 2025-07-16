package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RoomServiceTestSuite struct {
	suite.Suite
	roomService *RoomService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockRoomRepo *repository.MockRoomRepository
	mockUserRepo *repository.MockUserRepository
	mockBillRepo *repository.MockBillRepository

	// Mock HTTP client
	mockHTTPClient *utils.MockHTTPClient
}

func TestRoomServiceSuite(t *testing.T) {
	suite.Run(t, new(RoomServiceTestSuite))
}

func (s *RoomServiceTestSuite) SetupTest() {
	var err error
	s.db, s.sqlMock, err = tests.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize mock repositories
	s.mockRoomRepo = new(repository.MockRoomRepository)
	s.mockUserRepo = new(repository.MockUserRepository)
	s.mockBillRepo = new(repository.MockBillRepository)

	s.mockHTTPClient = new(utils.MockHTTPClient)

	// Create service with mock dependencies
	s.roomService = NewRoomService(
		s.db,
		s.mockRoomRepo,
		s.mockUserRepo,
		s.mockBillRepo,
		s.mockHTTPClient,
		"test-api-key",
		logrus.New(),
	)
}

func (s *RoomServiceTestSuite) TestCreateRoomWithInvites_Success() {
	// Setup test data
	host := &model.User{ID: 1, Username: "host"}
	invitees := []uint{2, 3}
	room := &model.Room{Name: "Test Room", VenuePlaceId: "ChIJN1t_tDeuEmsRUsoyG83frY4"}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)

	// Mock expectations
	s.mockUserRepo.On("FindByID", "1").Return(host, nil)
	s.mockUserRepo.On("FindByIDs", &invitees).Return(&[]model.User{
		{ID: 2}, {ID: 3},
	}, nil)

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

	s.mockHTTPClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" &&
			req.URL.String() == fmt.Sprintf("https://places.googleapis.com/v1/places/%s", room.VenuePlaceId) &&
			req.Header.Get("X-Goog-Api-Key") == "test-api-key" &&
			req.Header.Get("X-Goog-FieldMask") == "googleMapsUri"
	})).Return(mockResponse, nil)

	s.mockRoomRepo.On("Create", room).Return(nil)
	s.mockRoomRepo.On("CreateInvites", mock.Anything).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	resultRoom, resultInvites, err := s.roomService.CreateRoomWithInvites(
		room, "1", &invitees,
	)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resultRoom)
	assert.Equal(s.T(), uint(1), resultRoom.HostID)
	assert.Len(s.T(), *resultInvites, 2)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockHTTPClient.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestCreateRoomWithInvites_HostNotFound() {
	// Setup test data
	invitees := []uint{2, 3}
	room := &model.Room{Name: "Test Room", VenuePlaceId: "randomPlaceId"}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)

	// Simulate an error
	s.mockUserRepo.On("FindByID", "1").Return((*model.User)(nil), errors.New("user not found"))

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	_, _, err := s.roomService.CreateRoomWithInvites(
		room, "1", &invitees,
	)

	// Assertions
	assert.Error(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestGetRooms_Success() {
	// Setup test data
	userId := "1"
	page := 1
	expectedRooms := []model.Room{
		{ID: "1", Name: "Room 1"},
		{ID: "2", Name: "Room 2"},
	}

	// Mock expectations
	s.mockRoomRepo.On("GetUserRooms", userId, page, ROOM_PAGE_SIZE).Return(&expectedRooms, nil)

	// Execute
	rooms, err := s.roomService.GetRooms(userId, page)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedRooms, rooms)
	assert.Equal(s.T(), 2, len(*rooms))

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestGetUnjoinedPublicRooms_Success() {
	// Setup test data
	userId := "1"
	expectedRooms := []model.Room{
		{ID: "1", Name: "Room 1", IsPrivate: false},
		{ID: "2", Name: "Room 2", IsPrivate: false},
	}

	// Mock expectations
	s.mockRoomRepo.On("GetUnjoinedRoomsByIsPrivate", userId, false).Return(&expectedRooms, nil)

	// Execute
	rooms, err := s.roomService.GetUnjoinedPublicRooms(userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedRooms, rooms)
	assert.Equal(s.T(), 2, len(*rooms))

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestCloseRoom_Success() {
	// Setup test data
	roomId := "1"
	userId := "1"
	room := &model.Room{ID: "1", HostID: 1, IsClosed: false}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockRoomRepo.On("UpdateRoom", mock.AnythingOfType("*model.Room")).Return(nil)
	s.mockRoomRepo.On("DeletePendingInvites", roomId).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.roomService.CloseRoom(roomId, userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.True(s.T(), room.IsClosed)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestCloseRoom_NotHost() {
	// Setup test data
	roomId := "1"
	userId := "2"                           // Not the host
	room := &model.Room{ID: "1", HostID: 1} // Host ID is 1

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.roomService.CloseRoom(roomId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrInvalidHost, err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockBillRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestCloseRoom_UnconsolidatedBills() {
	// Setup test data
	roomId := "123"
	userId := "1"

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Mock expectations
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)

	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.UNCONSOLIDATED, nil) // Simulate unconsolidated bills

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.roomService.CloseRoom(roomId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrRoomHasUnconsolidatedBills, err)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestUpdateRoomInviteStatus_Accept() {
	// Setup test data
	roomId := "1"
	userId := "2"
	status := "accepted"
	room := &model.Room{ID: "1", AttendeesCount: 1}
	user := &model.User{ID: 2}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)

	// Mock expectations
	s.mockRoomRepo.On("UpdateInviteStatus", roomId, userId, status).Return(nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockUserRepo.On("FindByID", userId).Return(user, nil)
	s.mockRoomRepo.On("UpdateRoom", room).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.roomService.UpdateRoomInviteStatus(roomId, userId, status)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, room.AttendeesCount) // Attendee count should increment

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestRespondToRoomInvite_Rejected() {
	// Setup test data
	roomId := "123"
	userId := "2"

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockRoomRepo.On("UpdateInviteStatus", roomId, userId, "rejected").Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	room, attendees, err := s.roomService.RespondToRoomInvite(roomId, userId, false)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), room) // Room should be nil since the invite was rejected
	assert.Nil(s.T(), attendees)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestUpdateRoomInviteStatus_InvalidStatus() {
	// Execute
	err := s.roomService.UpdateRoomInviteStatus("1", "2", "invalid-status")

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrInvalidRoomStatus, err)
}

func (s *RoomServiceTestSuite) TestJoinRoom_Success() {
	// Setup test data
	roomId := "1"
	userId := "2"
	room := &model.Room{ID: "1", AttendeesCount: 1}
	user := &model.User{ID: 2}

	// Mock expectations
	s.mockRoomRepo.On("IsUserInRoom", roomId, userId).Return(false, nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockUserRepo.On("FindByID", userId).Return(user, nil)
	s.mockRoomRepo.On("UpdateRoom", room).Return(nil)
	s.mockRoomRepo.On("GetRoomAttendees", roomId).Return(&[]model.User{*user}, nil)

	// Execute
	resultRoom, resultAttendees, err := s.roomService.JoinRoom(roomId, userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, resultRoom.AttendeesCount)
	assert.Len(s.T(), *resultAttendees, 1)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestJoinRoom_AlreadyInRoom() {
	// Setup test data
	roomId := "1"
	userId := "2"

	// Mock expectations
	s.mockRoomRepo.On("IsUserInRoom", roomId, userId).Return(true, nil)

	// Execute
	_, _, err := s.roomService.JoinRoom(roomId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrAlreadyInRoom, err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestLeaveRoom_Success() {
	// Setup test data
	roomId := "1"
	userId := "2"                           // Not the host
	room := &model.Room{ID: "1", HostID: 1} // Host ID is 1

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockRoomRepo.On("RemoveUserFromRoom", roomId, userId).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.roomService.LeaveRoom(roomId, userId)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestLeaveRoom_AsHost() {
	// Setup test data
	roomId := "1"
	userId := "1" // Same as host ID
	room := &model.Room{ID: "1", HostID: 1}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockBillRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockBillRepo)

	// Mock expectations
	s.mockBillRepo.On("GetRoomBillConsolidationStatus", roomId).Return(repository.NO_BILLS, nil)
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.roomService.LeaveRoom(roomId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrLeaveRoomAsHost, err)

	// Verify mock calls
	s.mockBillRepo.AssertExpectations(s.T())
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestInviteUsersToRoom_Success() {
	// Setup test data
	roomId := "1"
	inviterId := "1"
	inviteesIds := []uint{2, 3}
	room := &model.Room{ID: "1", HostID: 1}
	inviter := &model.User{ID: 1}
	invitees := []model.User{{ID: 2}, {ID: 3}}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)
	s.mockUserRepo.On("FindByID", inviterId).Return(inviter, nil)
	s.mockUserRepo.On("FindByIDs", &inviteesIds).Return(&invitees, nil)
	s.mockRoomRepo.On("IsUserInRoom", roomId, "2").Return(false, nil)
	s.mockRoomRepo.On("IsUserInRoom", roomId, "3").Return(false, nil)
	s.mockRoomRepo.On("HasPendingInvites", roomId, "2").Return(false, nil)
	s.mockRoomRepo.On("HasPendingInvites", roomId, "3").Return(false, nil)
	s.mockRoomRepo.On("CreateInvites", mock.Anything).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	invites, err := s.roomService.InviteUsersToRoom(roomId, inviterId, &inviteesIds)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Len(s.T(), *invites, 2)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestInviteUsersToRoom_NotHost() {
	// Setup test data
	roomId := "123"
	inviterId := "2"
	invitees := []uint{3, 4}
	room := &model.Room{ID: "123", HostID: 1} // Host ID is 1, inviter ID is 2

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomId).Return(room, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	invites, err := s.roomService.InviteUsersToRoom(roomId, inviterId, &invitees)

	// Assertions
	assert.Equal(s.T(), ErrInvalidHost, err)
	assert.Len(s.T(), *invites, 0)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
}

func (s *RoomServiceTestSuite) TestQueryVenue_Success() {
	// Setup test data
	locationQuery := "pizza"
	expectedResponse := `{
        "suggestions": [
            {
                "placePrediction": {
					"placeId": "ChIJN1t_tDeuEmsRUcIaWtf4MzE",
                    "text": {
                        "text": "Pizza Hut, Jurong Point, Singapore"
                    },
                    "structuredFormat": {
                        "mainText": {
                            "text": "Pizza Hut"
                        }
                    }
                }
            }
        ]
    }`

	expectedRequest := func(req *http.Request) bool {
		return req.URL.String() == "https://places.googleapis.com/v1/places:autocomplete" &&
			req.Method == "POST" &&
			req.Header.Get("Content-Type") == "application/json" &&
			req.Header.Get("X-Goog-Api-Key") == "test-api-key" &&
			req.Header.Get("X-Goog-FieldMask") == "suggestions.placePrediction.text.text,suggestions.placePrediction.placeId,suggestions.placePrediction.structuredFormat.mainText.text"
	}

	// Mock the HTTP request and response
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(expectedResponse)),
		Header:     make(http.Header),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	s.mockHTTPClient.On("Do", mock.MatchedBy(expectedRequest)).Return(mockResponse, nil)

	// Execute
	predictions, err := s.roomService.QueryVenue(locationQuery)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), predictions)
	assert.Len(s.T(), *predictions, 1)
	assert.Equal(s.T(), "Pizza Hut", (*predictions)[0].Name)

	// Verify mock calls
	s.mockHTTPClient.AssertExpectations(s.T())
}
