package services

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/model"
	pushNotificationModel "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
)

var (
	ErrEmptyContent = errors.New("content cannot be empty")
)

type NotificationService struct {
	notificationRepo  repository.NotificationRepository
	subscriptionRepo  repository.SubscriptionRepository
	notificationsChan chan<- pushNotificationModel.NotificationData
	logger            *logrus.Entry
}

// NOTE: used var instead of func to enable mocking in tests
var NewNotificationService = func(
	notificationRepo repository.NotificationRepository,
	subscriptionRepo repository.SubscriptionRepository,
	notificationsChan chan<- pushNotificationModel.NotificationData,
	logger *logrus.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo:  notificationRepo,
		subscriptionRepo:  subscriptionRepo,
		notificationsChan: notificationsChan,
		logger:            utils.AddServiceField(logger, "NotificationService"),
	}
}

// CreateNotification creates a new notification for a user
func (s *NotificationService) CreateNotification(userId, title, content string) (*model.Notification, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	userIdUint, err := utils.StringToUint(userId)
	if err != nil {
		return nil, err
	}

	notification := &model.Notification{
		UserID:  userIdUint,
		Title:   title,
		Content: content,
		IsRead:  false,
	}
	return s.notificationRepo.Create(notification)
}

// MarkNotificationAsRead updates a notification's read status
func (s *NotificationService) MarkNotificationAsRead(notificationId, userId uint) error {
	return s.notificationRepo.MarkAsRead(notificationId, userId)
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(notificationId, userId uint) (*model.Notification, error) {
	return s.notificationRepo.FindByIDAndUser(notificationId, userId)
}

// GetNotifications retrieves all notifications for a user
func (s *NotificationService) GetNotifications(userId uint) (*[]model.Notification, error) {
	return s.notificationRepo.FindByUser(userId)
}

func (s *NotificationService) SendNotification(userId, title, message string) error {
	if _, err := s.CreateNotification(userId, title, message); err != nil {
		s.logger.Error("Error creating notification: ", err)
		return err
	}

	subscriptions, err := s.subscriptionRepo.FindByUserID(userId)
	if err != nil {
		s.logger.Error("Error getting subscriptions: ", err)
		return err
	}

	for _, sub := range *subscriptions {
		s.notificationsChan <- pushNotificationModel.NotificationData{
			Subscription: NewWebPushSubscriptionObj(&sub),
			Title:        title,
			Message:      message,
		}
	}

	return nil
}
