package services

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/RowenTey/JustJio/server/api/model"
	pushNotificationModel "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/RowenTey/JustJio/server/api/repository"
)

type SubscriptionServiceTestSuite struct {
	suite.Suite
	subscriptionService *SubscriptionService

	// Mock repositories
	mockSubscriptionRepo *repository.MockSubscriptionRepository

	// Mock channel
	mockNotificationsChan chan pushNotificationModel.NotificationData
}

func TestSubscriptionServiceSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionServiceTestSuite))
}

func (s *SubscriptionServiceTestSuite) SetupTest() {
	// Initialize mock repository
	s.mockSubscriptionRepo = new(repository.MockSubscriptionRepository)

	// Create buffered channel for testing
	s.mockNotificationsChan = make(chan pushNotificationModel.NotificationData, 1)

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
	subscription := &model.Subscription{
		UserID:   1,
		Endpoint: "https://example.com",
		P256dh:   "p256dh_key",
		Auth:     "auth_key",
	}

	// Mock expectations
	s.mockSubscriptionRepo.On("Create", subscription).Run(func(args mock.Arguments) {
		sub := args.Get(0).(*model.Subscription)
		sub.ID = "1"
	}).Return(subscription, nil)

	// Execute
	result, err := s.subscriptionService.CreateSubscription(subscription)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "1", result.ID)
	assert.Equal(s.T(), subscription.UserID, result.UserID)
	assert.Equal(s.T(), subscription.Endpoint, result.Endpoint)
	assert.Equal(s.T(), subscription.P256dh, result.P256dh)
	assert.Equal(s.T(), subscription.Auth, result.Auth)

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
	subscription := &model.Subscription{
		UserID:   1,
		Endpoint: "https://example.com",
	}
	expectedErr := errors.New("database error")

	// Mock expectations
	s.mockSubscriptionRepo.On("Create", subscription).Return((*model.Subscription)(nil), expectedErr)

	// Execute
	result, err := s.subscriptionService.CreateSubscription(subscription)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	assert.Nil(s.T(), result)

	// Verify no notification was sent
	assert.Equal(s.T(), 0, len(s.mockNotificationsChan))

	// Verify mock calls
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_Success() {
	// Setup test data
	userID := "1"
	expectedSubscriptions := []model.Subscription{
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
	s.mockSubscriptionRepo.On("FindByUserID", userID).Return(&expectedSubscriptions, nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByUserID(userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedSubscriptions, result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_NotFound() {
	// Setup test data
	userID := "999"

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByUserID", userID).Return(&[]model.Subscription{}, nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByUserID(userID)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), *result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionByEndpoint_Success() {
	// Setup test data
	endpoint := "https://example.com"
	expectedSubscription := &model.Subscription{
		ID:       "1",
		UserID:   1,
		Endpoint: endpoint,
	}

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByEndpoint", endpoint).Return(expectedSubscription, nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByEndpoint(endpoint)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedSubscription, result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionByEndpoint_NotFound() {
	// Setup test data
	endpoint := "https://nonexistent.com"

	// Mock expectations
	s.mockSubscriptionRepo.On("FindByEndpoint", endpoint).Return((*model.Subscription)(nil), nil)

	// Execute
	result, err := s.subscriptionService.GetSubscriptionsByEndpoint(endpoint)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), result)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_Success() {
	// Setup test data
	subID := "1"

	// Mock expectations
	s.mockSubscriptionRepo.On("Delete", subID).Return(nil)

	// Execute
	err := s.subscriptionService.DeleteSubscription(subID)

	// Assertions
	assert.NoError(s.T(), err)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_Failure() {
	// Setup test data
	subID := "999"
	expectedErr := errors.New("delete failed")

	// Mock expectations
	s.mockSubscriptionRepo.On("Delete", subID).Return(expectedErr)

	// Execute
	err := s.subscriptionService.DeleteSubscription(subID)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	s.mockSubscriptionRepo.AssertExpectations(s.T())
}
