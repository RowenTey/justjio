package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/database"
	"gorm.io/gorm"
)

type MessageRepository interface {
	WithTx(tx *gorm.DB) MessageRepository

	Create(ctx context.Context, message *models.Message) error
	FindByID(ctx context.Context, msgID string) (*models.Message, error)
	FindByRoom(ctx context.Context, roomId string, page int, pageSize int, asc bool) ([]models.Message, error)
	Delete(ctx context.Context, msgID string) error
	DeleteByRoom(ctx context.Context, roomID string) error
	CountByRoom(ctx context.Context, roomID string) (int64, error)
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

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.
		WithContext(ctx).
		Create(message).Error
}

func (r *messageRepository) FindByID(ctx context.Context, msgID string) (*models.Message, error) {
	var message models.Message
	if err := r.db.
		WithContext(ctx).
		Where("id = ?", msgID).
		First(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *messageRepository) Delete(ctx context.Context, msgID string) error {
	result := r.db.
		WithContext(ctx).
		Where("id = ?", msgID).
		Delete(&models.Message{})

	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *messageRepository) DeleteByRoom(ctx context.Context, roomID string) error {
	result := r.db.
		WithContext(ctx).
		Where("room_id = ?", roomID).
		Delete(&models.Message{})

	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil

}

func (r *messageRepository) CountByRoom(ctx context.Context, roomID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Message{}).
		Where("room_id = ?", roomID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *messageRepository) FindByRoom(ctx context.Context, roomId string, page int, pageSize int, asc bool) ([]models.Message, error) {
	var messages []models.Message

	order := "sent_at ASC"
	if !asc {
		order = "sent_at DESC"
	}

	if err := r.db.
		WithContext(ctx).
		Joins("Sender", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Where("messages.room_id = ?", roomId).
		Order(order).
		Scopes(database.Paginate(page, pageSize)).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
