package services

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pushNotificationModel "github.com/RowenTey/JustJio/server/api/dto/push_notifications"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	notificationService *NotificationService

	// Mock repositories
	mockNotificationRepo *repository.MockNotificationRepository
	mockSubscriptionRepo *repository.MockSubscriptionRepository

	// Mock channel
	mockNotificationsChan chan pushNotificationModel.NotificationData
}

func TestNotificationServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(NotificationServiceTestSuite))
}

func (s *NotificationServiceTestSuite) SetupTest() {
	// Initialize mock repositories
	s.mockNotificationRepo = new(repository.MockNotificationRepository)
	s.mockSubscriptionRepo = new(repository.MockSubscriptionRepository)

	// Create buffered channel for testing
	s.mockNotificationsChan = make(chan pushNotificationModel.NotificationData, 10)

	// Create notificationService with mock dependencies
	s.notificationService = NewNotificationService(
		s.mockNotificationRepo,
		s.mockSubscriptionRepo,
		s.mockNotificationsChan,
		logrus.New(),
	)
}

func (s *NotificationServiceTestSuite) TearDownTest() {
	close(s.mockNotificationsChan)
}

func (s *NotificationServiceTestSuite) TestCreateNotification_Success() {
	// Setup test data
	userId := "1"
	title := "Test Title"
	content := "Test Content"
	userIdUint := uint(1)

	expectedNotification := &model.Notification{
		UserID:  userIdUint,
		Title:   title,
		Content: content,
		IsRead:  false,
	}

	// Mock expectations
	s.mockNotificationRepo.On("Create", expectedNotification).Return(expectedNotification, nil)

	// Execute
	result, err := s.notificationService.CreateNotification(userId, title, content)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedNotification, result)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestCreateNotification_EmptyContent() {
	// Setup test data
	userId := "1"
	title := "Test Title"
	content := ""

	// Execute
	result, err := s.notificationService.CreateNotification(userId, title, content)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrEmptyContent, err)
	assert.Nil(s.T(), result)
	s.mockNotificationRepo.AssertNotCalled(s.T(), "Create")
}

func (s *NotificationServiceTestSuite) TestCreateNotification_InvalidUserID() {
	// Setup test data
	userId := "invalid"
	title := "Test Title"
	content := "Test Content"

	// Execute
	result, err := s.notificationService.CreateNotification(userId, title, content)

	// Assertions
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	s.mockNotificationRepo.AssertNotCalled(s.T(), "Create")
}

func (s *NotificationServiceTestSuite) TestMarkNotificationAsRead_Success() {
	// Setup test data
	notificationId := uint(1)
	userId := uint(1)

	// Mock expectations
	s.mockNotificationRepo.On("FindByIDAndUser", notificationId, userId).Return(&model.Notification{
		ID:     notificationId,
		UserID: userId,
	}, nil)
	s.mockNotificationRepo.On("MarkAsRead", notificationId, userId).Return(nil)

	// Execute
	err := s.notificationService.MarkNotificationAsRead(notificationId, userId)

	// Assertions
	assert.NoError(s.T(), err)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestGetNotification_Success() {
	// Setup test data
	notificationId := uint(1)
	userId := uint(1)
	expectedNotification := &model.Notification{
		ID:      notificationId,
		UserID:  userId,
		Title:   "Test",
		Content: "Test Content",
		IsRead:  false,
	}

	// Mock expectations
	s.mockNotificationRepo.On("FindByIDAndUser", notificationId, userId).Return(expectedNotification, nil)

	// Execute
	result, err := s.notificationService.GetNotification(notificationId, userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedNotification, result)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestGetNotifications_Success() {
	// Setup test data
	userId := uint(1)
	expectedNotifications := []model.Notification{
		{
			ID:      1,
			UserID:  userId,
			Title:   "Test 1",
			Content: "Content 1",
			IsRead:  false,
		},
		{
			ID:      2,
			UserID:  userId,
			Title:   "Test 2",
			Content: "Content 2",
			IsRead:  true,
		},
	}

	// Mock expectations
	s.mockNotificationRepo.On("FindByUser", userId).Return(&expectedNotifications, nil)

	// Execute
	result, err := s.notificationService.GetNotifications(userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedNotifications, result)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestSendNotification_Success() {
	// Setup test data
	userId := "1"
	title := "Test Title"
	message := "Test Message"
	userIdUint := uint(1)

	subscriptions := []model.Subscription{
		{ID: "1", UserID: userIdUint, Endpoint: "endpoint1", P256dh: "key1", Auth: "auth1"},
		{ID: "2", UserID: userIdUint, Endpoint: "endpoint2", P256dh: "key2", Auth: "auth2"},
	}

	// Mock expectations
	s.mockNotificationRepo.On("Create", mock.AnythingOfType("*model.Notification")).Return(&model.Notification{}, nil)
	s.mockSubscriptionRepo.On("FindByUserID", userId).Return(&subscriptions, nil)

	// Execute
	err := s.notificationService.SendNotification(userId, title, message)

	// Assertions
	assert.NoError(s.T(), err)

	// Verify notifications were sent to channel
	require.Equal(s.T(), 2, len(s.mockNotificationsChan))
	for range subscriptions {
		<-s.mockNotificationsChan // Drain the channel
	}

	s.mockNotificationRepo.AssertExpectations(s.T())
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestSendNotification_CreateFails() {
	// Setup test data
	userId := "1"
	title := "Test Title"
	message := "Test Message"

	// Mock expectations
	s.mockNotificationRepo.On("Create", mock.AnythingOfType("*model.Notification")).Return((*model.Notification)(nil), errors.New("create error"))

	// Execute
	err := s.notificationService.SendNotification(userId, title, message)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), 0, len(s.mockNotificationsChan)) // No messages should be sent
	s.mockNotificationRepo.AssertExpectations(s.T())
	s.mockSubscriptionRepo.AssertNotCalled(s.T(), "FindByUserID")
}

func (s *NotificationServiceTestSuite) TestSendNotification_NoSubscriptions() {
	// Setup test data
	userId := "1"
	title := "Test Title"
	message := "Test Message"

	// Mock expectations
	s.mockNotificationRepo.On("Create", mock.AnythingOfType("*model.Notification")).Return(&model.Notification{}, nil)
	s.mockSubscriptionRepo.On("FindByUserID", userId).Return(&[]model.Subscription{}, nil)

	// Execute
	err := s.notificationService.SendNotification(userId, title, message)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, len(s.mockNotificationsChan)) // No messages should be sent
	s.mockNotificationRepo.AssertExpectations(s.T())
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}
