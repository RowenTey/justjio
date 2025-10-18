package services

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	kafkaModel "github.com/RowenTey/JustJio/server/api/dto/kafka"
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
	ctx context.Context, roomId string, senderId string, roomUserIds []string, content string) error {
	return database.RunInTransaction(ms.db, sql.LevelDefault, func(tx *gorm.DB) error {
		roomRepoTx := ms.roomRepo.WithTx(tx)
		userRepoTx := ms.userRepo.WithTx(tx)
		messageRepoTx := ms.messageRepo.WithTx(tx)

		room, err := roomRepoTx.GetByID(ctx, roomId)
		if err != nil {
			return err
		}

		sender, err := userRepoTx.FindByID(ctx, senderId)
		if err != nil {
			return err
		}

		msg := model.Message{
			RoomID:   room.ID,
			SenderID: sender.ID,
			Content:  content,
		}
		if err := messageRepoTx.Create(ctx, &msg); err != nil {
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

func (ms *MessageService) GetMessageById(ctx context.Context, msgId string) (*model.Message, error) {
	return ms.messageRepo.FindByID(ctx, msgId)
}

func (ms *MessageService) DeleteMessage(ctx context.Context, msgId string) error {
	return ms.messageRepo.Delete(ctx, msgId)
}

func (ms *MessageService) DeleteRoomMessages(ctx context.Context, roomId string) error {
	return ms.messageRepo.DeleteByRoom(ctx, roomId)
}

func (ms *MessageService) CountNumMessagesPages(ctx context.Context, roomId string) (int, error) {
	count, err := ms.messageRepo.CountByRoom(ctx, roomId)
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(count) / float64(MESSAGE_PAGE_SIZE))), nil
}

func (ms *MessageService) GetMessagesByRoomId(ctx context.Context, roomId string, page int, asc bool) ([]model.Message, int, error) {
	messages, err := ms.messageRepo.FindByRoom(ctx, roomId, page, MESSAGE_PAGE_SIZE, asc)
	if err != nil {
		return nil, 0, err
	}

	pageCount, err := ms.CountNumMessagesPages(ctx, roomId)
	if err != nil {
		return nil, 0, err
	}

	return messages, pageCount, nil
}
