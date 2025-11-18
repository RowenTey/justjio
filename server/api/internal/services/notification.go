package services

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/notifications"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
)

type NotificationService struct {
	notificationRepo  repositories.NotificationRepository
	subscriptionRepo  repositories.SubscriptionRepository
	notificationsChan chan<- notifications.NotificationData
	logger            *logrus.Entry
}

func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	notificationsChan chan<- notifications.NotificationData,
	logger *logrus.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo:  notificationRepo,
		subscriptionRepo:  subscriptionRepo,
		notificationsChan: notificationsChan,
		logger:            logger.WithFields(logrus.Fields{"service": "NotificationService"}),
	}
}

// CreateNotification creates a new notification for a user
func (s *NotificationService) CreateNotification(ctx context.Context, userId, title, content string) (uint, error) {
	userIdUint, err := utils.StringToUint(userId)
	if err != nil {
		return 0, err
	}

	notification := &models.Notification{
		UserID:  userIdUint,
		Title:   title,
		Content: content,
		IsRead:  false,
	}

	createdNotification, err := s.notificationRepo.Create(ctx, notification)
	if err != nil {
		return 0, err
	}

	return createdNotification.ID, nil
}

// MarkNotificationAsRead updates a notification's read status
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, notificationId uint) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationId)
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, notificationId uint) (*response.NotificationDto, error) {
	notification, err := s.notificationRepo.FindByID(ctx, notificationId)
	if err != nil {
		return nil, err
	}

	return &response.NotificationDto{
		ID:        notification.ID,
		Title:     notification.Title,
		Content:   notification.Content,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
	}, nil
}

// GetNotifications retrieves all notifications for a user
func (s *NotificationService) GetNotifications(ctx context.Context, userId string) ([]response.NotificationDto, error) {
	notifications, err := s.notificationRepo.FindByUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	notificationsDto := make([]response.NotificationDto, len(notifications))
	for i, notification := range notifications {
		notificationsDto[i] = response.NotificationDto{
			ID:        notification.ID,
			Title:     notification.Title,
			Content:   notification.Content,
			IsRead:    notification.IsRead,
			CreatedAt: notification.CreatedAt,
		}
	}

	return notificationsDto, nil
}

// SendNotification sends a notification to a user and their subscriptions
func (s *NotificationService) SendNotification(ctx context.Context, userId, title, message string) error {
	notificationId, err := s.CreateNotification(ctx, userId, title, message)
	if err != nil {
		s.logger.Error("Error creating notification: ", err)
		return err
	}
	s.logger.Info("Notification created with ID: ", notificationId)

	subscriptions, err := s.subscriptionRepo.FindByUserID(ctx, userId)
	if err != nil {
		s.logger.Error("Error getting subscriptions: ", err)
		return err
	}

	for _, sub := range subscriptions {
		s.notificationsChan <- notifications.NotificationData{
			Subscription: NewWebPushSubscriptionObj(&sub),
			Title:        title,
			Message:      message,
		}
	}

	return nil
}
