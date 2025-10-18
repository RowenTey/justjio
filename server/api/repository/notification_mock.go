package repository

import (
	"context"
	"github.com/RowenTey/JustJio/server/api/model"
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

func (m *MockNotificationRepository) Create(ctx context.Context, notification *model.Notification) (*model.Notification, error) {
	args := m.Called(ctx, notification)
	return args.Get(0).(*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, notificationID uint) (*model.Notification, error) {
	args := m.Called(ctx, notificationID)
	return args.Get(0).(*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByUser(ctx context.Context, userID string) ([]model.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, notificationID uint) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}
