package repository

import (
	"context"
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

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *model.Subscription) (*model.Subscription, error) {
	args := m.Called(ctx, subscription)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByID(ctx context.Context, subID string) (*model.Subscription, error) {
	args := m.Called(ctx, subID)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByUserID(ctx context.Context, userID string) ([]model.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByEndpoint(ctx context.Context, endpoint string) (*model.Subscription, error) {
	args := m.Called(ctx, endpoint)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, subID string) error {
	args := m.Called(ctx, subID)
	return args.Error(0)
}
