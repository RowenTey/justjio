package repository

import (
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	WithTx(tx *gorm.DB) MessageRepository

	Create(message *model.Message) error
	FindByID(msgID string) (*model.Message, error)
	Delete(msgID string) error
	DeleteByRoom(roomID string) error
	CountByRoom(roomID string) (int64, error)
	FindByRoom(roomId string, page int, pageSize int, asc bool) (*[]model.Message, error)
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

func (r *messageRepository) Create(message *model.Message) error {
	return r.db.
		// Omit("Room", "Sender").
		Create(message).Error
}

func (r *messageRepository) FindByID(msgID string) (*model.Message, error) {
	var message model.Message
	err := r.db.
		Where("id = ?", msgID).
		First(&message).Error
	return &message, err
}

func (r *messageRepository) Delete(msgID string) error {
	return r.db.
		Where("id = ?", msgID).
		Delete(&model.Message{}).Error
}

func (r *messageRepository) DeleteByRoom(roomID string) error {
	return r.db.
		Where("room_id = ?", roomID).
		Delete(&model.Message{}).Error
}

func (r *messageRepository) CountByRoom(roomID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).
		Where("room_id = ?", roomID).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) FindByRoom(roomId string, page int, pageSize int, asc bool) (*[]model.Message, error) {
	var messages []model.Message

	order := "sent_at ASC"
	if !asc {
		order = "sent_at DESC"
	}

	err := r.db.
		Where("room_id = ?", roomId).
		Order(order).
		Scopes(database.Paginate(page, pageSize)).
		Preload("Room").
		Preload("Sender").
		Find(&messages).Error

	return &messages, err
}
