package services

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/sirupsen/logrus"

	pushNotificationModel "github.com/RowenTey/JustJio/server/api/model/push_notifications"
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

func (s *SubscriptionService) CreateSubscription(subscription *model.Subscription) (*model.Subscription, error) {
	subscription, err := s.subscriptionRepo.Create(subscription)
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

func (s *SubscriptionService) GetSubscriptionsByUserID(userID string) (*[]model.Subscription, error) {
	return s.subscriptionRepo.FindByUserID(userID)
}

func (s *SubscriptionService) GetSubscriptionsByEndpoint(endpoint string) (*model.Subscription, error) {
	return s.subscriptionRepo.FindByEndpoint(endpoint)
}

func (s *SubscriptionService) DeleteSubscription(subId string) error {
	return s.subscriptionRepo.Delete(subId)
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
