package services

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/notifications"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	notificationService *NotificationService

	// Mock repositories
	mockNotificationRepo *repositories.MockNotificationRepository
	mockSubscriptionRepo *repositories.MockSubscriptionRepository

	// Mock channel
	mockNotificationsChan chan notifications.NotificationData
}

func TestNotificationServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(NotificationServiceTestSuite))
}

func (s *NotificationServiceTestSuite) SetupTest() {
	// Initialize mock repositories
	s.mockNotificationRepo = new(repositories.MockNotificationRepository)
	s.mockSubscriptionRepo = new(repositories.MockSubscriptionRepository)

	// Create buffered channel for testing
	s.mockNotificationsChan = make(chan notifications.NotificationData, 10)

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

	expectedNotification := &models.Notification{
		UserID:  userIdUint,
		Title:   title,
		Content: content,
		IsRead:  false,
	}

	// Mock expectations
	s.mockNotificationRepo.On("Create", mock.Anything, expectedNotification).Return(expectedNotification, nil)

	// Execute
	result, err := s.notificationService.CreateNotification(context.Background(), userId, title, content)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedNotification.ID, result)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestCreateNotification_InvalidUserID() {
	// Setup test data
	userId := "invalid"
	title := "Test Title"
	content := "Test Content"

	// Execute
	result, err := s.notificationService.CreateNotification(context.Background(), userId, title, content)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), uint(0), result)
	s.mockNotificationRepo.AssertNotCalled(s.T(), "Create")
}

func (s *NotificationServiceTestSuite) TestMarkNotificationAsRead_Success() {
	// Setup test data
	notificationId := uint(1)

	// Mock expectations
	s.mockNotificationRepo.On("MarkAsRead", mock.Anything, notificationId).Return(nil)

	// Execute
	err := s.notificationService.MarkNotificationAsRead(context.Background(), notificationId)

	// Assertions
	assert.NoError(s.T(), err)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestGetNotification_Success() {
	// Setup test data
	notificationId := uint(1)
	userId := uint(1)
	expectedNotification := &models.Notification{
		ID:      notificationId,
		UserID:  userId,
		Title:   "Test",
		Content: "Test Content",
		IsRead:  false,
	}

	// Mock expectations
	s.mockNotificationRepo.On("FindByID", mock.Anything, notificationId).Return(expectedNotification, nil)

	// Execute
	result, err := s.notificationService.GetNotification(context.Background(), notificationId)

	// Assertions
	assert.NoError(s.T(), err)
	expectedDto := &response.NotificationDto{
		ID:        expectedNotification.ID,
		Title:     expectedNotification.Title,
		Content:   expectedNotification.Content,
		IsRead:    expectedNotification.IsRead,
		CreatedAt: expectedNotification.CreatedAt,
	}
	assert.Equal(s.T(), expectedDto, result)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestGetNotifications_Success() {
	// Setup test data
	userId := uint(1)
	userIdStr := "1"
	expectedNotifications := []models.Notification{
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
	s.mockNotificationRepo.On("FindByUser", mock.Anything, userIdStr).Return(expectedNotifications, nil)

	// Execute
	result, err := s.notificationService.GetNotifications(context.Background(), userIdStr)

	// Assertions
	assert.NoError(s.T(), err)
	expectedDto := make([]response.NotificationDto, len(expectedNotifications))
	for i, notif := range expectedNotifications {
		expectedDto[i] = response.NotificationDto{
			ID:        notif.ID,
			Title:     notif.Title,
			Content:   notif.Content,
			IsRead:    notif.IsRead,
			CreatedAt: notif.CreatedAt,
		}
	}
	assert.Equal(s.T(), expectedDto, result)
	s.mockNotificationRepo.AssertExpectations(s.T())
}

func (s *NotificationServiceTestSuite) TestSendNotification_Success() {
	// Setup test data
	userId := "1"
	title := "Test Title"
	message := "Test Message"
	userIdUint := uint(1)

	subscriptions := []models.Subscription{
		{ID: "1", UserID: userIdUint, Endpoint: "endpoint1", P256dh: "key1", Auth: "auth1"},
		{ID: "2", UserID: userIdUint, Endpoint: "endpoint2", P256dh: "key2", Auth: "auth2"},
	}

	// Mock expectations
	s.mockNotificationRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Notification")).Return(&models.Notification{}, nil)
	s.mockSubscriptionRepo.On("FindByUserID", mock.Anything, userId).Return(subscriptions, nil)

	// Execute
	err := s.notificationService.SendNotification(context.Background(), userId, title, message)

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
	s.mockNotificationRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Notification")).Return((*models.Notification)(nil), errors.New("create error"))

	// Execute
	err := s.notificationService.SendNotification(context.Background(), userId, title, message)

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
	s.mockNotificationRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Notification")).Return(&models.Notification{}, nil)
	s.mockSubscriptionRepo.On("FindByUserID", mock.Anything, userId).Return([]models.Subscription{}, nil)

	// Execute
	err := s.notificationService.SendNotification(context.Background(), userId, title, message)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, len(s.mockNotificationsChan)) // No messages should be sent
	s.mockNotificationRepo.AssertExpectations(s.T())
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}
