package repository

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	WithTx(tx *gorm.DB) NotificationRepository

	Create(ctx context.Context, notification *model.Notification) (*model.Notification, error)
	FindByID(ctx context.Context, notificationID uint) (*model.Notification, error)
	FindByUser(ctx context.Context, userID string) ([]model.Notification, error)
	MarkAsRead(ctx context.Context, notificationID uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// WithTx returns a new NotificationRepository with the provided transaction
func (r *notificationRepository) WithTx(tx *gorm.DB) NotificationRepository {
	if tx == nil {
		return r
	}
	return &notificationRepository{db: tx}
}

func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) (*model.Notification, error) {
	err := r.db.WithContext(ctx).Create(notification).Error
	return notification, err
}

func (r *notificationRepository) FindByID(ctx context.Context, notificationID uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.
		WithContext(ctx).
		Where("id = ?", notificationID).
		First(&notification).Error
	return &notification, err
}

func (r *notificationRepository) FindByUser(ctx context.Context, userID string) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID uint) error {
	return r.db.
		WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ?", notificationID).
		Update("is_read", true).Error
}
