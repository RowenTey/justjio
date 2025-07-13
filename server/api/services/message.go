package services

import (
	"math"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	kafkaModel "github.com/RowenTey/JustJio/server/api/model/kafka"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
)

const (
	MESSAGE_PAGE_SIZE = 10
)

type MessageService struct {
	db           *gorm.DB
	messageRepo  repository.MessageRepository
	roomRepo     repository.RoomRepository
	userRepo     repository.UserRepository
	kafkaService KafkaService
	logger       *logrus.Entry
}

func NewMessageService(
	db *gorm.DB,
	messageRepo repository.MessageRepository,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	kafkaService KafkaService,
	logger *logrus.Logger) *MessageService {
	return &MessageService{
		db:           db,
		messageRepo:  messageRepo,
		roomRepo:     roomRepo,
		userRepo:     userRepo,
		kafkaService: kafkaService,
		logger:       utils.AddServiceField(logger, "MessageService"),
	}
}

func (ms *MessageService) SaveMessage(
	roomId string, senderId string, roomUserIds *[]string, content string) error {
	return database.RunInTransaction(ms.db, func(tx *gorm.DB) error {
		roomRepoTx := ms.roomRepo.WithTx(tx)
		userRepoTx := ms.userRepo.WithTx(tx)
		messageRepoTx := ms.messageRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(roomId)
		if err != nil {
			return err
		}

		sender, err := userRepoTx.FindByID(senderId)
		if err != nil {
			return err
		}

		msg := model.Message{
			RoomID:   room.ID,
			SenderID: sender.ID,
			Content:  content,
		}
		if err := messageRepoTx.Create(&msg); err != nil {
			return err
		}

		// TODO: Outbox pattern?
		broadcastPayload := kafkaModel.KafkaMessage{
			MsgType: "CREATE_MESSAGE",
			Data: struct {
				RoomID     string `json:"roomId"`
				SenderID   string `json:"senderId"`
				SenderName string `json:"senderName"`
				Content    string `json:"content"`
				SentAt     string `json:"sentAt"`
			}{
				RoomID:     roomId,
				SenderID:   senderId,
				SenderName: sender.Username,
				Content:    content,
				SentAt:     time.Now().Format(time.RFC3339),
			},
		}
		if err := ms.kafkaService.BroadcastMessage(roomUserIds, broadcastPayload); err != nil {
			ms.logger.Error("Failed to broadcast message:", err)
			return err
		}
		ms.logger.Debug("Broadcasted message to Kafka!")

		ms.logger.Infof("Saved message to room %s", msg.RoomID)
		return nil
	})
}

func (ms *MessageService) GetMessageById(msgId, roomId string) (*model.Message, error) {
	return ms.messageRepo.FindByID(msgId, roomId)
}

func (ms *MessageService) DeleteMessage(msgId, roomId string) error {
	return ms.messageRepo.Delete(msgId, roomId)
}

func (ms *MessageService) DeleteRoomMessages(roomId string) error {
	return ms.messageRepo.DeleteByRoom(roomId)
}

func (ms *MessageService) CountNumMessagesPages(roomId string) (int, error) {
	count, err := ms.messageRepo.CountByRoom(roomId)
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(count) / float64(MESSAGE_PAGE_SIZE))), nil
}

func (ms *MessageService) GetMessagesByRoomId(roomId string, page int, asc bool) (*[]model.Message, int, error) {
	messages, err := ms.messageRepo.FindByRoom(roomId, page, MESSAGE_PAGE_SIZE, asc)
	if err != nil {
		return nil, 0, err
	}

	pageCount, err := ms.CountNumMessagesPages(roomId)
	if err != nil {
		return nil, 0, err
	}

	return messages, pageCount, nil
}
