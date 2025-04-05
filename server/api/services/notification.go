package services

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type NotificationService struct {
	DB     *gorm.DB
	Logger *log.Entry
}

// NOTE: used var instead of func to enable mocking in tests
var NewNotificationService = func(db *gorm.DB) *NotificationService {
	return &NotificationService{
		DB:     db,
		Logger: log.WithFields(log.Fields{"service": "NotificationService"}),
	}
}

// CreateNotification creates a new notification for a user
func (s *NotificationService) CreateNotification(userId uint, title, content string) (*model.Notification, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	notification := &model.Notification{
		UserID:  userId,
		Title:   title,
		Content: content,
		IsRead:  false,
	}

	if err := s.DB.Create(notification).Error; err != nil {
		return nil, err
	}

	return notification, nil
}

// MarkNotificationAsRead updates a notification's read status
func (s *NotificationService) MarkNotificationAsRead(notificationId, user_id uint) error {
	return s.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationId, user_id).
		Update("is_read", true).Error
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(notificationId, user_id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := s.DB.Model(&model.Notification{}).Where("id = ? AND user_id = ?", notificationId, user_id).First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetNotifications retrieves all notifications for a user
func (s *NotificationService) GetNotifications(userId uint) (*[]model.Notification, error) {
	var notifications []model.Notification
	if err := s.DB.Where("user_id = ?", userId).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return &notifications, nil
}
