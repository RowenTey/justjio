package repository

import (
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

func (m *MockNotificationRepository) Create(notification *model.Notification) (*model.Notification, error) {
	args := m.Called(notification)
	return args.Get(0).(*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByID(notificationID uint) (*model.Notification, error) {
	args := m.Called(notificationID)
	return args.Get(0).(*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) FindByUser(userID uint) (*[]model.Notification, error) {
	args := m.Called(userID)
	return args.Get(0).(*[]model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(notificationID uint) error {
	args := m.Called(notificationID)
	return args.Error(0)
}
