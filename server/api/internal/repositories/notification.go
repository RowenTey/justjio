package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	WithTx(tx *gorm.DB) NotificationRepository

	Create(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	FindByID(ctx context.Context, notificationID uint) (*models.Notification, error)
	FindByUser(ctx context.Context, userID string) ([]models.Notification, error)
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

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	err := r.db.WithContext(ctx).Create(notification).Error
	return notification, err
}

func (r *notificationRepository) FindByID(ctx context.Context, notificationID uint) (*models.Notification, error) {
	var notification models.Notification
	if err := r.db.
		WithContext(ctx).
		Where("id = ?", notificationID).
		First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) FindByUser(ctx context.Context, userID string) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID uint) error {
	return r.db.
		WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Update("is_read", true).Error
}
