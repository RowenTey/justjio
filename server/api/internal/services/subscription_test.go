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

type SubscriptionServiceTestSuite struct {
	suite.Suite
	subscriptionService *SubscriptionService

	// Mock repositories
	mockSubscriptionRepo *repositories.MockSubscriptionRepository

	// Mock channel
	mockNotificationsChan chan notifications.NotificationData
}

func TestSubscriptionServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SubscriptionServiceTestSuite))
}

func (s *SubscriptionServiceTestSuite) SetupTest() {
	// Initialize mock repository
	s.mockSubscriptionRepo = new(repositories.MockSubscriptionRepository)

	// Create buffered channel for testing
	s.mockNotificationsChan = make(chan notifications.NotificationData, 1)

	// Create subscriptionService with mock dependencies
	s.subscriptionService = NewSubscriptionService(
		s.mockSubscriptionRepo,
		s.mockNotificationsChan,
		logrus.New(),
	)
}

func (s *SubscriptionServiceTestSuite) TearDownTest() {
	close(s.mockNotificationsChan)
}

func (s *SubscriptionServiceTestSuite) TestCreateSubscription_Success() {
	// Setup test data
	subscription := &models.Subscription{
		UserID:   1,
		Endpoint: "https://example.com",
		P256dh:   "p256dh_key",
		Auth:     "auth_key",
	}

	// Mock expectations
	s.mockSubscriptionRepo.On("Create", mock.Anything, subscription).Run(func(args mock.Arguments) {
		sub := args.Get(1).(*models.Subscription) // Get second argument (index 1) since first is context
		sub.ID = "1"
	}).Return(subscription, nil)

	// Execute
	createdSubscriptionID, err := s.subscriptionService.CreateSubscription(context.Background(), subscription)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "1", createdSubscriptionID)

	// Verify notification was sent
	require.Equal(s.T(), 1, len(s.mockNotificationsChan))
	notification := <-s.mockNotificationsChan
	assert.Equal(s.T(), "Welcome", notification.Title)
	assert.Equal(s.T(), subscription.Endpoint, notification.Subscription.Endpoint)

	// Verify mock calls
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestCreateSubscription_Failure() {
	// Setup test data
	subscription := &models.Subscription{
		UserID:   1,
		Endpoint: "https://example.com",
	}
	expectedErr := errors.New("database error")

	// Mock expectations
	s.mockSubscriptionRepo.On("Create", mock.Anything, subscription).Return((*models.Subscription)(nil), expectedErr)

	// Execute
	result, err := s.subscriptionService.CreateSubscription(context.Background(), subscription)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	assert.Equal(s.T(), "", result)

	// Verify no notification was sent
	assert.Equal(s.T(), 0, len(s.mockNotificationsChan))

	// Verify mock calls
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_Success() {
	// Setup test data
	userID := "1"
	expectedSubscriptions := []models.Subscription{
		{
			ID:       "1",
			UserID:   1,
			Endpoint: "https://example.com/1",
		},
		{
			ID:       "2",
			UserID:   1,
			Endpoint: "https://example.com/2",
		},
	}

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByUserID", mock.Anything, userID).Return(expectedSubscriptions, nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByUserID(context.Background(), userID)

	// Assertions
	assert.NoError(s.T(), err)
	expectedDto := make([]response.SubscriptionDto, len(expectedSubscriptions))
	for i, sub := range expectedSubscriptions {
		expectedDto[i] = response.SubscriptionDto{
			ID:       sub.ID,
			Endpoint: sub.Endpoint,
			Auth:     sub.Auth,
			P256dh:   sub.P256dh,
		}
	}
	assert.Equal(s.T(), expectedDto, result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_NotFound() {
	// Setup test data
	userID := "999"

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByUserID", mock.Anything, userID).Return([]models.Subscription{}, nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByUserID(context.Background(), userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionByEndpoint_Success() {
	// Setup test data
	endpoint := "https://example.com"
	expectedSubscription := &models.Subscription{
		ID:       "1",
		UserID:   1,
		Endpoint: endpoint,
	}

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByEndpoint", mock.Anything, endpoint).Return(expectedSubscription, nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByEndpoint(context.Background(), endpoint)

	// Assertions
	assert.NoError(s.T(), err)
	expectedDto := &response.SubscriptionDto{
		ID:       expectedSubscription.ID,
		Endpoint: expectedSubscription.Endpoint,
		Auth:     expectedSubscription.Auth,
		P256dh:   expectedSubscription.P256dh,
	}
	assert.Equal(s.T(), expectedDto, result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionByEndpoint_NotFound() {
	// Setup test data
	endpoint := "https://nonexistent.com"

	// Mock expectations
	expectedErr := errors.New("not found")
	s.mockSubscriptionRepo.On("FindByEndpoint", mock.Anything, endpoint).Return((*models.Subscription)(nil), expectedErr)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByEndpoint(context.Background(), endpoint)

	// Assertions
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_Success() {
	// Setup test data
	subID := "1"

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByID", mock.Anything, subID).Return(&models.Subscription{ID: subID}, nil)
	s.mockSubscriptionRepo.On("Delete", mock.Anything, subID).Return(nil)

	// Execute
	err := s.subscriptionService.DeleteSubscription(context.Background(), subID)

	// Assertions
	assert.NoError(s.T(), err)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_Failure() {
	// Setup test data
	subID := "999"
	expectedErr := errors.New("delete failed")

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByID", mock.Anything, subID).Return(&models.Subscription{ID: subID}, nil)
	s.mockSubscriptionRepo.On("Delete", mock.Anything, subID).Return(expectedErr)

	// Execute
	err := s.subscriptionService.DeleteSubscription(context.Background(), subID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}
