package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) WithTx(tx *gorm.DB) SubscriptionRepository {
	args := m.Called(tx)
	return args.Get(0).(SubscriptionRepository)
}

func (m *MockSubscriptionRepository) Create(subscription *model.Subscription) (*model.Subscription, error) {
	args := m.Called(subscription)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByUserID(userID string) (*[]model.Subscription, error) {
	args := m.Called(userID)
	return args.Get(0).(*[]model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByEndpoint(endpoint string) (*model.Subscription, error) {
	args := m.Called(endpoint)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Delete(subID string) error {
	args := m.Called(subID)
	return args.Error(0)
}
