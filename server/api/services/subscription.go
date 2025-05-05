package services

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/SherClockHolmes/webpush-go"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SubscriptionService struct {
	DB     *gorm.DB
	Logger *log.Entry
}

// NOTE: used var instead of func to enable mocking in tests
var NewSubscriptionService = func(db *gorm.DB) *SubscriptionService {
	return &SubscriptionService{
		DB:     db,
		Logger: log.WithFields(log.Fields{"service": "SubscriptionService"}),
	}
}

func (s *SubscriptionService) CreateSubscription(subscription *model.Subscription) (*model.Subscription, error) {
	if err := s.DB.Create(subscription).Error; err != nil {
		return nil, err
	}
	return subscription, nil
}

func (s *SubscriptionService) GetSubscriptionsByUserID(userID uint) (*[]model.Subscription, error) {
	var subscriptions []model.Subscription
	if err := s.DB.Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return &subscriptions, nil
}

func (s *SubscriptionService) GetSubscriptionsByEndpoint(endpoint string) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := s.DB.Where("endpoint = ?", endpoint).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (s *SubscriptionService) DeleteSubscription(subId string) error {
	if err := s.DB.Where("id = ?", subId).First(&model.Subscription{}).Error; err != nil {
		return err
	}

	if err := s.DB.Where("id = ?", subId).Delete(&model.Subscription{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) NewWebPushSubscriptionObj(subscription *model.Subscription) *webpush.Subscription {
	return &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			Auth:   subscription.Auth,
			P256dh: subscription.P256dh,
		},
	}
}
