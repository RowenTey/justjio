package handlers

import (
	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/dto/response"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type MessageHandler struct {
	messageService *services.MessageService
	logger         *log.Entry
}

func NewMessageHandler(
	messageService *services.MessageService,
	logger *log.Logger,
) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		logger:         utils.AddServiceField(logger, "MessageHandler"),
	}
}

func (h *MessageHandler) GetMessage(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	msgId := c.Params("msgId")

	message, err := h.messageService.GetMessageById(msgId, roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No message found")
	}

	return utils.HandleSuccess(c, "Retrieved message successfully", message)
}

func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	page := c.QueryInt("page", 1)
	asc := c.QueryBool("asc", true)

	messages, pageCount, err := h.messageService.GetMessagesByRoomId(roomId, page, asc)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No messages found")
	}

	response := response.GetMessagesResponse{
		Messages:  *messages,
		Page:      page,
		PageCount: pageCount,
	}
	return utils.HandleSuccess(c, "Retrieved messages successfully", response)
}

func (h *MessageHandler) CreateMessage(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	var request request.CreateMessageRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomUserIds := c.Locals("roomUserIds").(*[]string)

	err := h.messageService.SaveMessage(roomId, userId, roomUserIds, request.Content)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room or user not found")
	}

	return utils.HandleSuccess(c, "Message saved successfully", nil)
}
