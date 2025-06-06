package services

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"

	"gorm.io/gorm"
)

const (
	MESSAGE_PAGE_SIZE = 10
)

type MessageService struct {
	DB     *gorm.DB
	Logger *log.Entry
}

// NOTE: used var instead of func to enable mocking in tests
var NewMessageService = func(db *gorm.DB) *MessageService {
	return &MessageService{
		DB:     db,
		Logger: log.WithFields(log.Fields{"service": "MessageService"}),
	}
}

func (ms *MessageService) SaveMessage(room *model.Room, sender *model.User, content string) error {
	db := ms.DB.Table("messages")

	msg := model.Message{
		RoomID:   room.ID,
		SenderID: sender.ID,
		Content:  content,
		SentAt:   time.Now(),
	}

	// Omit to avoid creating new room
	if err := db.Omit("Room", "Sender").Create(&msg).Error; err != nil {
		return err
	}

	ms.Logger.Infof("Saved message to room %s", msg.RoomID)
	return nil
}

func (ms *MessageService) GetMessageById(msgId, roomId string) (*model.Message, error) {
	db := ms.DB.Table("messages")
	var message model.Message

	if err := db.Where("id = ? AND room_id = ?", msgId, roomId).First(&message).Error; err != nil {
		return &model.Message{}, err
	}

	return &message, nil
}

func (ms *MessageService) DeleteMessage(msgId, roomId string) error {
	db := ms.DB.Table("messages")

	if err := db.Where("id = ? AND room_id = ?", msgId, roomId).Delete(&model.Message{}).Error; err != nil {
		return err
	}

	return nil
}

func (ms *MessageService) DeleteRoomMessages(roomId string) error {
	db := ms.DB.Table("messages")

	if err := db.Where("room_id = ?", roomId).Delete(&model.Message{}).Error; err != nil {
		return err
	}

	return nil
}

func (ms *MessageService) CountNumMessagesPages(roomId string) (int, error) {
	db := ms.DB.Table("messages")

	var count int64
	err := db.Where("room_id = ?", roomId).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(float64(count) / float64(MESSAGE_PAGE_SIZE))), nil
}

func (ms *MessageService) GetMessagesByRoomId(roomId string, page int, asc bool) (*[]model.Message, error) {
	db := ms.DB.Table("messages")
	var message []model.Message

	// sorted by
	order := "sent_at ASC"
	if !asc {
		order = "sent_at DESC"
	}

	if err := db.
		Where("room_id = ?", roomId).
		Order(order).
		Scopes(database.Paginate(page, MESSAGE_PAGE_SIZE)).
		Preload("Room").
		Preload("Sender").
		Find(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}
