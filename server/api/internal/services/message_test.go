package services

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
	"github.com/RowenTey/JustJio/server/api/pkg/kafka"
	"github.com/RowenTey/JustJio/server/api/pkg/tests"
)

type MessageServiceTestSuite struct {
	suite.Suite
	messageService *MessageService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockMessageRepo *repositories.MockMessageRepository
	mockRoomRepo    *repositories.MockRoomRepository
	mockUserRepo    *repositories.MockUserRepository
	mockKafkaClient *kafka.MockKafkaClient
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
	s.mockMessageRepo = new(repositories.MockMessageRepository)
	s.mockRoomRepo = new(repositories.MockRoomRepository)
	s.mockUserRepo = new(repositories.MockUserRepository)
	s.mockKafkaClient = new(kafka.MockKafkaClient)

	// Create messageService with mock dependencies
	s.messageService = NewMessageService(
		s.db,
		s.mockMessageRepo,
		s.mockRoomRepo,
		s.mockUserRepo,
		s.mockKafkaClient,
		logrus.New(),
	)
}

func (s *MessageServiceTestSuite) TestSaveMessage_Success() {
	// Setup test data
	roomID := "room1"
	senderID := "user1"
	content := "Hello world"
	roomUserIDs := []string{"user1", "user2"}

	room := &models.Room{ID: roomID}
	sender := &models.User{ID: 1, Username: "testuser"}

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockMessageRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockMessageRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	s.mockUserRepo.On("FindByID", mock.Anything, senderID).Return(sender, nil)
	s.mockMessageRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).Return(nil)
	s.mockKafkaClient.On("BroadcastMessage", mock.Anything, roomUserIDs, mock.AnythingOfType("kafka.KafkaMessage")).Return(nil)

	// Expect transaction commit
	s.sqlMock.ExpectCommit()

	// Execute
	err := s.messageService.SaveMessage(context.Background(), roomID, senderID, roomUserIDs, content)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockMessageRepo.AssertExpectations(s.T())
	s.mockKafkaClient.AssertExpectations(s.T())
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
	s.mockRoomRepo.On("GetByID", mock.Anything, roomID).Return((*models.Room)(nil), gorm.ErrRecordNotFound)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.messageService.SaveMessage(context.Background(), roomID, senderID, roomUserIDs, content)

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

	room := &models.Room{ID: roomID}
	sender := &models.User{ID: 1, Username: "testuser"}
	kafkaErr := errors.New("kafka error")

	// Expect transaction begin
	s.sqlMock.ExpectBegin()

	// Setup repository mocks with transaction support
	s.mockRoomRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockRoomRepo)
	s.mockUserRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockUserRepo)
	s.mockMessageRepo.On("WithTx", mock.AnythingOfType("*gorm.DB")).Return(s.mockMessageRepo)

	// Mock expectations
	s.mockRoomRepo.On("GetByID", mock.Anything, roomID).Return(room, nil)
	s.mockUserRepo.On("FindByID", mock.Anything, senderID).Return(sender, nil)
	s.mockMessageRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).Return(nil)
	s.mockKafkaClient.On("BroadcastMessage", mock.Anything, roomUserIDs, mock.AnythingOfType("kafka.KafkaMessage")).Return(kafkaErr)

	// Expect transaction rollback
	s.sqlMock.ExpectRollback()

	// Execute
	err := s.messageService.SaveMessage(context.Background(), roomID, senderID, roomUserIDs, content)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), kafkaErr, err)

	// Verify mock calls
	s.mockRoomRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockMessageRepo.AssertExpectations(s.T())
	s.mockKafkaClient.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestGetMessageById_Success() {
	msgID := "1"
	expectedMsg := &models.Message{
		ID:       1,
		RoomID:   "room1",
		SenderID: 1,
		Content:  "test message",
	}

	s.mockMessageRepo.On("FindByID", mock.Anything, msgID).Return(expectedMsg, nil)

	result, err := s.messageService.GetMessageById(context.Background(), msgID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedMsg, result)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestDeleteMessage_Success() {
	msgID := "1"

	s.mockMessageRepo.On("Delete", mock.Anything, msgID).Return(nil)

	err := s.messageService.DeleteMessage(context.Background(), msgID)

	assert.NoError(s.T(), err)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestDeleteRoomMessages_Success() {
	roomID := "room1"

	s.mockMessageRepo.On("DeleteByRoom", mock.Anything, roomID).Return(nil)

	err := s.messageService.DeleteRoomMessages(context.Background(), roomID)

	assert.NoError(s.T(), err)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestCountNumMessagesPages_Success() {
	roomID := "room1"
	totalMessages := int64(25)
	expectedPages := 3 // 25 messages / 10 per page = 2.5 → ceil to 3

	s.mockMessageRepo.On("CountByRoom", mock.Anything, roomID).Return(totalMessages, nil)

	result, err := s.messageService.CountNumMessagesPages(context.Background(), roomID)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedPages, result)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestGetMessagesByRoomId_Success() {
	roomID := "room1"
	page := 1
	expectedMessages := []models.Message{
		{ID: 1, Content: "message 1"},
		{ID: 2, Content: "message 2"},
	}
	totalPages := 2

	s.mockMessageRepo.On("FindByRoom", mock.Anything, roomID, page, MESSAGE_PAGE_SIZE, false).Return(expectedMessages, nil)
	s.mockMessageRepo.On("CountByRoom", mock.Anything, roomID).Return(int64(15), nil)

	messages, pages, err := s.messageService.GetMessagesByRoomId(context.Background(), roomID, page, false)

	assert.NoError(s.T(), err)
	expectedDto := make([]response.MessageDto, len(expectedMessages))
	for i, msg := range expectedMessages {
		expectedDto[i] = response.MessageDto{
			ID:      msg.ID,
			RoomID:  msg.RoomID,
			Content: msg.Content,
			SentAt:  msg.SentAt,
			Sender: response.MinimalUserDto{
				ID:         msg.Sender.ID,
				Username:   msg.Sender.Username,
				PictureUrl: msg.Sender.PictureUrl,
			},
		}
	}
	assert.Equal(s.T(), expectedDto, messages)
	assert.Equal(s.T(), totalPages, pages)
	s.mockMessageRepo.AssertExpectations(s.T())
}

func (s *MessageServiceTestSuite) TestGetMessagesByRoomId_EmptyRoom() {
	roomID := "empty-room"
	page := 1

	s.mockMessageRepo.On("FindByRoom", mock.Anything, roomID, page, MESSAGE_PAGE_SIZE, true).Return([]models.Message{}, nil)
	s.mockMessageRepo.On("CountByRoom", mock.Anything, roomID).Return(int64(0), nil)

	messages, pages, err := s.messageService.GetMessagesByRoomId(context.Background(), roomID, page, true)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), messages)
	assert.Equal(s.T(), 0, pages)
	s.mockMessageRepo.AssertExpectations(s.T())
}
