package services

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/sirupsen/logrus"

	pushNotificationModel "github.com/RowenTey/JustJio/server/api/dto/push_notifications"
	"github.com/RowenTey/JustJio/server/api/dto/response"
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

func (s *SubscriptionService) CreateSubscription(ctx context.Context, subscription *model.Subscription) (string, error) {
	subscription, err := s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		return "", err
	}

	s.notificationsChan <- pushNotificationModel.NotificationData{
		Subscription: NewWebPushSubscriptionObj(subscription),
		Title:        "Welcome",
		Message:      "Subscribed to JustJio! You will now receive notifications for app events.",
	}

	return subscription.ID, nil
}

func (s *SubscriptionService) GetSubscriptionsByUserID(ctx context.Context, userID string) ([]response.SubscriptionDto, error) {
	subs, err := s.subscriptionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	subsDto := make([]response.SubscriptionDto, len(subs))
	for i, sub := range subs {
		subsDto[i] = response.SubscriptionDto{
			ID:       sub.ID,
			Endpoint: sub.Endpoint,
			Auth:     sub.Auth,
			P256dh:   sub.P256dh,
		}
	}

	return subsDto, nil
}

func (s *SubscriptionService) GetSubscriptionsByEndpoint(ctx context.Context, endpoint string) (*response.SubscriptionDto, error) {
	sub, err := s.subscriptionRepo.FindByEndpoint(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	return &response.SubscriptionDto{
		ID:       sub.ID,
		Endpoint: sub.Endpoint,
		Auth:     sub.Auth,
		P256dh:   sub.P256dh,
	}, nil
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
