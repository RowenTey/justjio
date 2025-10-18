package services

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	pushNotificationModel "github.com/RowenTey/JustJio/server/api/dto/push_notifications"
	"github.com/RowenTey/JustJio/server/api/model"
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

func NewNotificationService(
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
func (s *NotificationService) CreateNotification(ctx context.Context, userId, title, content string) (*model.Notification, error) {
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
	return s.notificationRepo.Create(ctx, notification)
}

// MarkNotificationAsRead updates a notification's read status
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, notificationId uint) error {
	if _, err := s.notificationRepo.FindByID(ctx, notificationId); err != nil {
		return err
	}
	return s.notificationRepo.MarkAsRead(ctx, notificationId)
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, notificationId uint) (*model.Notification, error) {
	return s.notificationRepo.FindByID(ctx, notificationId)
}

// GetNotifications retrieves all notifications for a user
func (s *NotificationService) GetNotifications(ctx context.Context, userId string) ([]model.Notification, error) {
	return s.notificationRepo.FindByUser(ctx, userId)
}

// SendNotification sends a notification to a user and their subscriptions
func (s *NotificationService) SendNotification(ctx context.Context, userId, title, message string) error {
	if _, err := s.CreateNotification(ctx, userId, title, message); err != nil {
		s.logger.Error("Error creating notification: ", err)
		return err
	}

	subscriptions, err := s.subscriptionRepo.FindByUserID(ctx, userId)
	if err != nil {
		s.logger.Error("Error getting subscriptions: ", err)
		return err
	}

	for _, sub := range subscriptions {
		s.notificationsChan <- pushNotificationModel.NotificationData{
			Subscription: NewWebPushSubscriptionObj(&sub),
			Title:        title,
			Message:      message,
		}
	}

	return nil
}
