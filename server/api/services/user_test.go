package services

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/tests"
)

type UserServiceTestSuite struct {
	suite.Suite
	userService *UserService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockUserRepo *repository.MockUserRepository
}

func TestUserServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) SetupTest() {
	var err error
	s.db, s.sqlMock, err = tests.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize mock repositories
	s.mockUserRepo = new(repository.MockUserRepository)

	// Create userService with mock dependencies
	s.userService = NewUserService(
		s.db,
		s.mockUserRepo,
		logrus.New(),
	)
}

func (s *UserServiceTestSuite) TestAcceptFriendRequest_Success() {
	// Setup test data
	requestID := uint(1)
	senderID := uint(2)
	receiverID := uint(3)

	request := &model.FriendRequest{
		ID:         requestID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     "pending",
	}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockUserRepo.On("FindFriendRequest", requestID).Return(request, nil)
	s.mockUserRepo.On("UpdateFriendRequest", requestID, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	s.mockUserRepo.On("AddFriend", senderID, receiverID).Return(nil)
	s.mockUserRepo.On("UpdateNoOfPendingFriendRequests", []uint{receiverID}, -1).Return(nil)
	s.mockUserRepo.On("UpdateNoOfFriends", []uint{senderID, receiverID}, 1).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.userService.AcceptFriendRequest(requestID)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestAcceptFriendRequest_AlreadyProcessed() {
	// Setup test data
	requestID := uint(1)
	processedRequest := &model.FriendRequest{
		ID:          requestID,
		Status:      "accepted",
		RespondedAt: time.Now(),
	}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockUserRepo.On("FindFriendRequest", requestID).Return(processedRequest, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.userService.AcceptFriendRequest(requestID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrFriendRequestAlreadyProcessed, err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestAcceptFriendRequest_UpdateFails() {
	// Setup test data
	requestID := uint(1)
	request := &model.FriendRequest{
		ID:         requestID,
		SenderID:   2,
		ReceiverID: 3,
		Status:     "pending",
	}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockUserRepo.On("FindFriendRequest", requestID).Return(request, nil)
	s.mockUserRepo.On("UpdateFriendRequest", requestID, mock.AnythingOfType("map[string]interface {}")).Return(errors.New("update failed"))

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.userService.AcceptFriendRequest(requestID)

	// Assertions
	assert.Error(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_Success() {
	// Setup test data
	senderID := uint(1)
	receiverID := uint(2)

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockUserRepo.On("CheckFriendship", senderID, receiverID).Return(false, nil)
	s.mockUserRepo.On("CheckFriendRequestExists", senderID, receiverID).Return(false, nil)
	s.mockUserRepo.On("CreateFriendRequest", mock.AnythingOfType("*model.FriendRequest")).Return(nil)
	s.mockUserRepo.On("UpdateNoOfPendingFriendRequests", []uint{receiverID}, 1).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_AlreadyFriends() {
	// Setup test data
	senderID := uint(1)
	receiverID := uint(2)

	// Mock expectations
	s.mockUserRepo.On("CheckFriendship", senderID, receiverID).Return(true, nil)

	// Execute
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrAlreadyFriends, err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestSendFriendRequest_RequestExists() {
	// Setup test data
	senderID := uint(1)
	receiverID := uint(2)

	// Mock expectations
	s.mockUserRepo.On("CheckFriendship", senderID, receiverID).Return(false, nil)
	s.mockUserRepo.On("CheckFriendRequestExists", senderID, receiverID).Return(true, nil)

	// Execute
	err := s.userService.SendFriendRequest(senderID, receiverID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrFriendRequestExists, err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestSearchUsers_Success() {
	currentUserID := "1"
	query := "test"
	expected := []model.User{{ID: 2, Username: "testuser"}}

	s.mockUserRepo.On("SearchNonFriendUsers", currentUserID, query, 10).Return(expected, nil)

	result, err := s.userService.SearchNonFriendUsers(currentUserID, query)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expected, result)
}

func (s *UserServiceTestSuite) TestGetFriendRequestsByStatus_ValidStatus() {
	// Setup test data
	userID := uint(1)
	status := "pending"
	requests := []model.FriendRequest{
		{ID: 1, SenderID: 2, ReceiverID: userID, Status: status},
	}

	// Mock expectations
	s.mockUserRepo.On("FindFriendRequestsByReceiver", userID, status).Return(requests, nil)

	// Execute
	result, err := s.userService.GetFriendRequestsByStatus(userID, status)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), requests, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetFriendRequestsByStatus_InvalidStatus() {
	// Setup test data
	userID := uint(1)
	status := "invalid"

	// Execute
	result, err := s.userService.GetFriendRequestsByStatus(userID, status)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrInvalidFriendRequestStatus, err)
	assert.Nil(s.T(), result)
}

func (s *UserServiceTestSuite) TestGetUserByID_Success() {
	// Setup test data
	userID := "1"
	expectedUser := &model.User{ID: 1, Username: "testuser"}

	// Mock expectations
	s.mockUserRepo.On("FindByID", userID).Return(expectedUser, nil)

	// Execute
	result, err := s.userService.GetUserByID(userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedUser, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetUserByID_NotFound() {
	// Setup test data
	userID := "999"

	// Mock expectations
	s.mockUserRepo.On("FindByID", userID).Return((*model.User)(nil), gorm.ErrRecordNotFound)

	// Execute
	result, err := s.userService.GetUserByID(userID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.True(s.T(), errors.Is(err, gorm.ErrRecordNotFound))

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestUpdateUserField_Success() {
	// Setup test data
	userID := "1"
	field := "isOnline"
	value := true

	// Mock repository calls
	existingUser := &model.User{ID: 1, IsOnline: false}
	s.mockUserRepo.On("FindByID", userID).Return(existingUser, nil)
	s.mockUserRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	// Execute
	err := s.userService.UpdateUserField(userID, field, value)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestUpdateUserField_InvalidField() {
	// Setup test data
	userID := "1"
	field := "invalid_field"
	value := "somevalue"

	// Mock expectations
	existingUser := &model.User{ID: 1}
	s.mockUserRepo.On("FindByID", userID).Return(existingUser, nil)

	// Execute
	err := s.userService.UpdateUserField(userID, field, value)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrUserFieldNotSupported, err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestCreateOrUpdateUser_Create() {
	// Setup test data
	newUser := &model.User{Username: "newuser"}

	// Mock expectations
	s.mockUserRepo.On("Create", newUser).Return(newUser, nil)

	// Execute
	result, err := s.userService.UpsertUser(newUser, true)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newUser, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestCreateOrUpdateUser_Update() {
	// Setup test data
	existingUser := &model.User{ID: 1, Username: "existinguser"}

	// Mock expectations
	s.mockUserRepo.On("Update", existingUser).Return(nil)

	// Execute
	result, err := s.userService.UpsertUser(existingUser, false)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), existingUser, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

// TestDeleteUser_Success - Removed because DeleteUser method doesn't exist in UserService

func (s *UserServiceTestSuite) TestMarkOnline_Success() {
	// Setup test data
	userID := "1"

	// Mock expectations
	s.mockUserRepo.On("FindByID", userID).Return(&model.User{ID: 1}, nil)
	s.mockUserRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	// Execute
	err := s.userService.MarkOnline(userID)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestMarkOffline_Success() {
	// Setup test data
	userID := "1"

	// Mock repository calls
	existingUser := &model.User{ID: 1, IsOnline: true}
	s.mockUserRepo.On("FindByID", userID).Return(existingUser, nil)
	s.mockUserRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	// Execute
	err := s.userService.MarkOffline(userID)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestRejectFriendRequest_Success() {
	// Setup test data
	requestID := uint(1)
	receiverID := uint(3)
	request := &model.FriendRequest{
		ID:         requestID,
		ReceiverID: receiverID,
		Status:     "pending",
	}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock repository calls
	s.mockUserRepo.On("FindFriendRequest", requestID).Return(request, nil)
	s.mockUserRepo.On("UpdateFriendRequest", requestID, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	s.mockUserRepo.On("UpdateNoOfPendingFriendRequests", []uint{receiverID}, -1).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.userService.RejectFriendRequest(requestID)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestRejectFriendRequest_AlreadyProcessed() {
	// Setup test data
	requestID := uint(1)
	request := &model.FriendRequest{
		ID:          requestID,
		Status:      "accepted",
		RespondedAt: time.Now(),
	}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockUserRepo.On("FindFriendRequest", requestID).Return(request, nil)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.userService.RejectFriendRequest(requestID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrFriendRequestAlreadyProcessed, err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestRemoveFriend_Success() {
	// Setup test data
	userID := uint(1)
	friendID := uint(2)

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)

	// Mock expectations
	s.mockUserRepo.On("RemoveFriend", userID, friendID).Return(nil)
	s.mockUserRepo.On("UpdateNoOfFriends", []uint{userID, friendID}, -1).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.userService.RemoveFriend(userID, friendID)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetFriends_Success() {
	// Setup test data
	userID := "1"
	expectedFriends := []model.User{
		{ID: 2, Username: "friend1"},
		{ID: 3, Username: "friend2"},
	}

	// Mock expectations
	s.mockUserRepo.On("GetFriends", uint(1)).Return(expectedFriends, nil)

	// Execute
	result, err := s.userService.GetFriends(userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedFriends, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetFriends_InvalidID() {
	// Setup test data
	userID := "invalid"

	// Execute
	result, err := s.userService.GetFriends(userID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
}

func (s *UserServiceTestSuite) TestCountPendingFriendRequests_Success() {
	// Setup test data
	userID := uint(1)
	expectedCount := int64(3)

	// Mock expectations
	s.mockUserRepo.On("CountFriendRequestsByReceiver", userID, "pending").Return(expectedCount, nil)

	// Execute
	result, err := s.userService.CountPendingFriendRequests(userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedCount, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestGetNumFriends_Success() {
	// Setup test data
	userID := "1"
	expectedCount := int64(5)

	// Mock expectations
	s.mockUserRepo.On("CountFriends", uint(1)).Return(expectedCount, nil)

	// Execute
	result, err := s.userService.GetNumFriends(userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedCount, result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestIsFriend_True() {
	// Setup test data
	userID := uint(1)
	friendID := uint(2)

	// Mock expectations
	s.mockUserRepo.On("CheckFriendship", userID, friendID).Return(true, nil)

	// Execute
	result := s.userService.IsFriend(userID, friendID)

	// Assertions
	assert.True(s.T(), result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestIsFriend_False() {
	// Setup test data
	userID := uint(1)
	friendID := uint(2)

	// Mock expectations
	s.mockUserRepo.On("CheckFriendship", userID, friendID).Return(false, nil)

	// Execute
	result := s.userService.IsFriend(userID, friendID)

	// Assertions
	assert.False(s.T(), result)

	// Verify mock calls
	s.mockUserRepo.AssertExpectations(s.T())
}
