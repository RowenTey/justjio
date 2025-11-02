package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
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

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) (*models.Subscription, error) {
	args := m.Called(ctx, subscription)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByID(ctx context.Context, subID string) (*models.Subscription, error) {
	args := m.Called(ctx, subID)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByUserID(ctx context.Context, userID string) ([]models.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByEndpoint(ctx context.Context, endpoint string) (*models.Subscription, error) {
	args := m.Called(ctx, endpoint)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, subID string) error {
	args := m.Called(ctx, subID)
	return args.Error(0)
}
