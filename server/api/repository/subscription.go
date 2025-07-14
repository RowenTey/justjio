package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	WithTx(tx *gorm.DB) SubscriptionRepository

	Create(subscription *model.Subscription) (*model.Subscription, error)
	FindByUserID(userID string) (*[]model.Subscription, error)
	FindByEndpoint(endpoint string) (*model.Subscription, error)
	Delete(subID string) error
}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// WithTx returns a new SubscriptionRepository with the provided transaction
func (r *subscriptionRepository) WithTx(tx *gorm.DB) SubscriptionRepository {
	if tx == nil {
		return r
	}
	return &subscriptionRepository{db: tx}
}

func (r *subscriptionRepository) Create(subscription *model.Subscription) (*model.Subscription, error) {
	err := r.db.Create(subscription).Error
	return subscription, err
}

func (r *subscriptionRepository) FindByUserID(userID string) (*[]model.Subscription, error) {
	var subscriptions []model.Subscription
	err := r.db.Where("user_id = ?", userID).Find(&subscriptions).Error
	return &subscriptions, err
}

func (r *subscriptionRepository) FindByEndpoint(endpoint string) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.db.Where("endpoint = ?", endpoint).First(&subscription).Error
	return &subscription, err
}

func (r *subscriptionRepository) Delete(subID string) error {
	return r.db.Where("id = ?", subID).Delete(&model.Subscription{}).Error
}
