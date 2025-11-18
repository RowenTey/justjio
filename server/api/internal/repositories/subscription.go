package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	WithTx(tx *gorm.DB) SubscriptionRepository

	Create(ctx context.Context, subscription *models.Subscription) (*models.Subscription, error)
	FindByID(ctx context.Context, subID string) (*models.Subscription, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Subscription, error)
	FindByEndpoint(ctx context.Context, endpoint string) (*models.Subscription, error)
	Delete(ctx context.Context, subID string) error
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

func (r *subscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) (*models.Subscription, error) {
	err := r.db.WithContext(ctx).Create(subscription).Error
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (r *subscriptionRepository) FindByID(ctx context.Context, subID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).Where("id = ?", subID).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) FindByUserID(ctx context.Context, userID string) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *subscriptionRepository) FindByEndpoint(ctx context.Context, endpoint string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).Where("endpoint = ?", endpoint).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, subID string) error {
	return r.db.WithContext(ctx).Where("id = ?", subID).Delete(&models.Subscription{}).Error
}
