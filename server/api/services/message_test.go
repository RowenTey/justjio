package services

import (
	"errors"
	"testing"

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

type MessageServiceTestSuite struct {
	suite.Suite
	messageService *MessageService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockMessageRepo *repository.MockMessageRepository
	mockRoomRepo    *repository.MockRoomRepository
	mockUserRepo    *repository.MockUserRepository
	mockKafkaSvc    *MockKafkaService
}

func TestMessageServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MessageServiceTestSuite))
}

func (s *MessageServiceTestSuite) SetupTest() {
	var err error
	s.db, s.sqlMock, err = tests.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize mock dependencies
	s.mockMessageRepo = new(repository.MockMessageRepository)
	s.mockRoomRepo = new(repository.MockRoomRepository)
	s.mockUserRepo = new(repository.MockUserRepository)
	s.mockKafkaSvc = new(MockKafkaService)

	// Create messageService with mock dependencies
	s.messageService = NewMessageService(
		s.db,
		s.mockMessageRepo,
		s.mockRoomRepo,
		s.mockUserRepo,
		s.mockKafkaSvc,
		logrus.New(),
	)
}

func (s *MessageServiceTestSuite) TestSaveMessage_Success() {
	// Setup test data
	roomID := "room1"
	senderID := "user1"
	content := "Hello world"
	roomUserIDs := []string{"user1", "user2"}

	room := &model.Room{ID: roomID}
	sender := &model.User{ID: 1, Username: "testuser"}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockMessageRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockMessageRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomID).Return(room, nil)
	s.mockUserRepo.On("FindByID", senderID).Return(sender, nil)
	s.mockMessageRepo.On("Create", mock.AnythingOfType("*model.Message")).Return(nil)
	s.mockKafkaSvc.On("BroadcastMessage", &roomUserIDs, mock.AnythingOfType("model_kafka.KafkaMessage")).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.messageService.SaveMessage(roomID, senderID, &roomUserIDs, content)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockMessageRepo.AssertExpectations(s.T())
	s.mockKafkaSvc.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestSaveMessage_RoomNotFound() {
	roomID := "invalid-room"
	senderID := "user1"
	content := "Hello world"
	roomUserIDs := []string{"user1", "user2"}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockMessageRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockMessageRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomID).Return((*model.Room)(nil), gorm.ErrRecordNotFound)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.messageService.SaveMessage(roomID, senderID, &roomUserIDs, content)

	// Assertions
	assert.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, gorm.ErrRecordNotFound))

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertNotCalled(s.T(), "FindByID")
	s.mockMessageRepo.AssertNotCalled(s.T(), "Create")
}

func (s *MessageServiceTestSuite) TestSaveMessage_KafkaBroadcastFailure() {
	roomID := "room1"
	senderID := "user1"
	content := "Hello world"
	roomUserIDs := []string{"user1", "user2"}

	room := &model.Room{ID: roomID}
	sender := &model.User{ID: 1, Username: "testuser"}
	kafkaErr := errors.New("kafka error")

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockMessageRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockMessageRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", roomID).Return(room, nil)
	s.mockUserRepo.On("FindByID", senderID).Return(sender, nil)
	s.mockMessageRepo.On("Create", mock.AnythingOfType("*model.Message")).Return(nil)
	s.mockKafkaSvc.On("BroadcastMessage", &roomUserIDs, mock.AnythingOfType("model_kafka.KafkaMessage")).Return(kafkaErr)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.messageService.SaveMessage(roomID, senderID, &roomUserIDs, content)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), kafkaErr, err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockMessageRepo.AssertExpectations(s.T())
	s.mockKafkaSvc.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestGetMessageById_Success() {
	msgID := "1"
	roomID := "room1"
	expectedMsg := &model.Message{
		ID:       1,
		RoomID:   roomID,
		SenderID: 1,
		Content:  "test message",
	}

	s.mockMessageRepo.On("FindByID", msgID, roomID).Return(expectedMsg, nil)

	result, err := s.messageService.GetMessageById(msgID, roomID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedMsg, result)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestDeleteMessage_Success() {
	msgID := "1"
	roomID := "room1"

	s.mockMessageRepo.On("Delete", msgID, roomID).Return(nil)

	err := s.messageService.DeleteMessage(msgID, roomID)

	assert.NoError(s.T(), err)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestDeleteRoomMessages_Success() {
	roomID := "room1"

	s.mockMessageRepo.On("DeleteByRoom", roomID).Return(nil)

	err := s.messageService.DeleteRoomMessages(roomID)

	assert.NoError(s.T(), err)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestCountNumMessagesPages_Success() {
	roomID := "room1"
	totalMessages := int64(25)
	expectedPages := 3 // 25 messages / 10 per page = 2.5 → ceil to 3

	s.mockMessageRepo.On("CountByRoom", roomID).Return(totalMessages, nil)

	result, err := s.messageService.CountNumMessagesPages(roomID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedPages, result)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestGetMessagesByRoomId_Success() {
	roomID := "room1"
	page := 1
	expectedMessages := []model.Message{
		{ID: 1, Content: "message 1"},
		{ID: 2, Content: "message 2"},
	}
	totalPages := 2

	s.mockMessageRepo.On("FindByRoom", roomID, page, MESSAGE_PAGE_SIZE, false).Return(&expectedMessages, nil)
	s.mockMessageRepo.On("CountByRoom", roomID).Return(int64(15), nil)

	messages, pages, err := s.messageService.GetMessagesByRoomId(roomID, page, false)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedMessages, messages)
	assert.Equal(s.T(), totalPages, pages)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestGetMessagesByRoomId_EmptyRoom() {
	roomID := "empty-room"
	page := 1

	s.mockMessageRepo.On("FindByRoom", roomID, page, MESSAGE_PAGE_SIZE, true).Return(&[]model.Message{}, nil)
	s.mockMessageRepo.On("CountByRoom", roomID).Return(int64(0), nil)

	messages, pages, err := s.messageService.GetMessagesByRoomId(roomID, page, true)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), *messages)
	assert.Equal(s.T(), 0, pages)
	s.mockMessageRepo.AssertExpectations(s.T())
}
