package services

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/sirupsen/logrus"

	pushNotificationModel "github.com/RowenTey/JustJio/server/api/dto/push_notifications"
)

type SubscriptionService struct {
	subscriptionRepo  repository.SubscriptionRepository
	notificationsChan chan<- pushNotificationModel.NotificationData
	logger            *logrus.Entry
}

func NewSubscriptionService(
	subscriptionRepo repository.SubscriptionRepository,
	notificationsChan chan<- pushNotificationModel.NotificationData,
	logger *logrus.Logger,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo:  subscriptionRepo,
		notificationsChan: notificationsChan,
		logger:            utils.AddServiceField(logger, "SubscriptionService"),
	}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, subscription *model.Subscription) (*model.Subscription, error) {
	subscription, err := s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		return nil, err
	}

	s.notificationsChan <- pushNotificationModel.NotificationData{
		Subscription: NewWebPushSubscriptionObj(subscription),
		Title:        "Welcome",
		Message:      "Subscribed to JustJio! You will now receive notifications for app events.",
	}

	return subscription, nil
}

func (s *SubscriptionService) GetSubscriptionsByUserID(ctx context.Context, userID string) ([]model.Subscription, error) {
	return s.subscriptionRepo.FindByUserID(ctx, userID)
}

func (s *SubscriptionService) GetSubscriptionsByEndpoint(ctx context.Context, endpoint string) (*model.Subscription, error) {
	return s.subscriptionRepo.FindByEndpoint(ctx, endpoint)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, subId string) error {
	if _, err := s.subscriptionRepo.FindByID(ctx, subId); err != nil {
		return err
	}
	return s.subscriptionRepo.Delete(ctx, subId)
}

func NewWebPushSubscriptionObj(subscription *model.Subscription) *webpush.Subscription {
	return &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			Auth:   subscription.Auth,
			P256dh: subscription.P256dh,
		},
	}
}
