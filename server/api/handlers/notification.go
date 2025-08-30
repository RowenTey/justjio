package handlers

import (
	"errors"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
	logger              *log.Entry
}

func NewNotificationHandler(
	notificationService *services.NotificationService,
	logger *log.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              utils.AddServiceField(logger, "NotificationHandler"),
	}
}

// CreateNotification handles the creation of a new notification
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	var request request.CreateNotificationRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	userId := utils.UIntToString(request.UserId)

	err := h.notificationService.SendNotification(userId, request.Title, request.Content)
	if err != nil {
		if errors.Is(err, services.ErrEmptyContent) {
			return utils.HandleInvalidInputError(c, err)
		}
		return utils.HandleInternalServerError(c, err)
	}

	return utils.HandleSuccess(c, "Notification created successfully", nil)
}

// MarkNotificationAsRead handles marking a notification as read
func (h *NotificationHandler) MarkNotificationAsRead(c *fiber.Ctx) error {
	notificationId, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	h.logger.Infof("Marking notification %d as read", notificationId)
	if err := h.
		notificationService.
		MarkNotificationAsRead(uint(notificationId)); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Notification not found")
	}

	return utils.HandleSuccess(c, "Notification marked as read successfully", nil)
}

// GetNotification handles retrieving a single notification
func (h *NotificationHandler) GetNotification(c *fiber.Ctx) error {
	notificationId, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	notification, err := h.
		notificationService.
		GetNotification(uint(notificationId))
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Notification not found")
	}

	return utils.HandleSuccess(c, "Retrieved notification successfully", notification)
}

// GetNotifications handles retrieving all notifications for a user
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	userIdInt, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	notifications, err := h.notificationService.GetNotifications(uint(userIdInt))
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	return utils.HandleSuccess(c, "Retrieved notifications successfully", notifications)
}
