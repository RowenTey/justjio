package handlers

import (
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/internal/middlewares"
	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/request"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
	"github.com/RowenTey/JustJio/server/api/pkg/otel"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"

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
		logger:         logger.WithFields(logrus.Fields{"handler": "MessageHandler"}),
	}
}

// GetMessage retrieves a specific message by ID
// @Summary Get message by ID
// @Description Retrieves a specific message by its ID
// @Tags Messages
// @Accept json
// @Produce json
// @Param msgId path string true "Message ID"
// @Success 200 {object} object{status=string,message=string,data=models.Message} "Retrieved message successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No message found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/messages/{msgId} [get]
func (h *MessageHandler) GetMessage(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	msgId := c.Params("msgId")

	message, err := h.messageService.GetMessageById(ctx, msgId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No message found")
	}

	return utils.HandleSuccess(c, "Retrieved message successfully", message)
}

// GetMessages retrieves messages for a room
// @Summary Get room messages
// @Description Retrieves paginated messages for a specific room with optional sorting
// @Tags Messages
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param page query int false "Page number" default(1)
// @Param asc query bool false "Sort order (ascending if true)" default(true)
// @Success 200 {object} object{status=string,message=string,data=response.GetMessagesResponse} "Retrieved messages successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No messages found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/messages [get]
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	roomId := c.Params("roomId")
	page := c.QueryInt("page", 1)
	asc := c.QueryBool("asc", true)

	messages, pageCount, err := h.messageService.GetMessagesByRoomId(ctx, roomId, page, asc)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No messages found")
	}

	response := response.GetMessagesResponse{
		Messages:  messages,
		Page:      page,
		PageCount: pageCount,
	}
	return utils.HandleSuccess(c, "Retrieved messages successfully", response)
}

// CreateMessage creates a new message in a room
// @Summary Create message
// @Description Creates and saves a new message in the specified room
// @Tags Messages
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param message body request.CreateMessageRequest true "Message content"
// @Success 200 {object} utils.EmptyApiResponse "Message saved successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input"
// @Failure 404 {object} utils.EmptyApiResponse "Room or user not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /rooms/{roomId}/messages [post]
func (h *MessageHandler) CreateMessage(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	roomId := c.Params("roomId")

	req := middlewares.GetValidatedRequest[request.CreateMessageRequest](c)

	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")
	roomUserIds := c.Locals("roomUserIds").([]string)

	err := h.messageService.SaveMessage(ctx, roomId, userId, roomUserIds, req.Content)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room or user not found")
	}

	return utils.HandleSuccess[any](c, "Message saved successfully", nil)
}
