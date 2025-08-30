package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	WithTx(tx *gorm.DB) NotificationRepository

	Create(notification *model.Notification) (*model.Notification, error)
	FindByID(notificationID uint) (*model.Notification, error)
	FindByUser(userID uint) (*[]model.Notification, error)
	MarkAsRead(notificationID uint) error
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

func (r *notificationRepository) Create(notification *model.Notification) (*model.Notification, error) {
	err := r.db.Create(notification).Error
	return notification, err
}

func (r *notificationRepository) FindByID(notificationID uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.
		Where("id = ?", notificationID).
		First(&notification).Error
	return &notification, err
}

func (r *notificationRepository) FindByUser(userID uint) (*[]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return &notifications, err
}

func (r *notificationRepository) MarkAsRead(notificationID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("id = ?", notificationID).
		Update("is_read", true).Error
}
