package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) WithTx(tx *gorm.DB) NotificationRepository {
	args := m.Called(tx)
	return args.Get(0).(NotificationRepository)
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	args := m.Called(ctx, notification)
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, notificationID uint) (*models.Notification, error) {
	args := m.Called(ctx, notificationID)
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByUser(ctx context.Context, userID string) ([]models.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, notificationID uint) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}
