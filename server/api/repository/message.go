package repository

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	WithTx(tx *gorm.DB) MessageRepository

	Create(ctx context.Context, message *model.Message) error
	FindByID(ctx context.Context, msgID string) (*model.Message, error)
	Delete(ctx context.Context, msgID string) error
	DeleteByRoom(ctx context.Context, roomID string) error
	CountByRoom(ctx context.Context, roomID string) (int64, error)
	FindByRoom(ctx context.Context, roomId string, page int, pageSize int, asc bool) ([]model.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

// WithTx returns a new MessageRepository with the provided transaction
func (r *messageRepository) WithTx(tx *gorm.DB) MessageRepository {
	if tx == nil {
		return r
	}
	return &messageRepository{db: tx}
}

func (r *messageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.
		WithContext(ctx).
		// Omit("Room", "Sender").
		Create(message).Error
}

func (r *messageRepository) FindByID(ctx context.Context, msgID string) (*model.Message, error) {
	var message model.Message
	err := r.db.
		WithContext(ctx).
		Where("id = ?", msgID).
		First(&message).Error
	return &message, err
}

func (r *messageRepository) Delete(ctx context.Context, msgID string) error {
	return r.db.
		WithContext(ctx).
		Where("id = ?", msgID).
		Delete(&model.Message{}).Error
}

func (r *messageRepository) DeleteByRoom(ctx context.Context, roomID string) error {
	return r.db.
		WithContext(ctx).
		Where("room_id = ?", roomID).
		Delete(&model.Message{}).Error
}

func (r *messageRepository) CountByRoom(ctx context.Context, roomID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("room_id = ?", roomID).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) FindByRoom(ctx context.Context, roomId string, page int, pageSize int, asc bool) ([]model.Message, error) {
	var messages []model.Message

	order := "sent_at ASC"
	if !asc {
		order = "sent_at DESC"
	}

	err := r.db.
		WithContext(ctx).
		Where("room_id = ?", roomId).
		Order(order).
		Scopes(database.Paginate(page, pageSize)).
		Preload("Room").
		Preload("Sender").
		Find(&messages).Error

	return messages, err
}
