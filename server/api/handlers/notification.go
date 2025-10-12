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

// CreateNotification creates a new notification
// @Summary Create notification
// @Description Creates and sends a new notification to a user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param notification body request.CreateNotificationRequest true "Notification details"
// @Success 200 {object} utils.EmptyApiResponse "Notification created successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input or empty content"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /notifications [post]
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

	return utils.HandleSuccess[any](c, "Notification created successfully", nil)
}

// MarkNotificationAsRead marks a notification as read
// @Summary Mark notification as read
// @Description Marks a specific notification as read by its ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} utils.EmptyApiResponse "Notification marked as read successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid notification ID"
// @Failure 404 {object} utils.EmptyApiResponse "Notification not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{userId}/notifications/{notificationId} [patch]
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

	return utils.HandleSuccess[any](c, "Notification marked as read successfully", nil)
}

// GetNotification retrieves a specific notification
// @Summary Get notification by ID
// @Description Retrieves a specific notification by its ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} object{status=string,message=string,data=model.Notification} "Retrieved notification successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid notification ID"
// @Failure 404 {object} utils.EmptyApiResponse "Notification not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /notifications/{id} [get]
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

// GetNotifications retrieves all notifications for a user
// @Summary Get user notifications
// @Description Retrieves all notifications for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} object{status=string,message=string,data=[]model.Notification} "Retrieved notifications successfully"
// @Failure 404 {object} utils.EmptyApiResponse "User not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{userId}/notifications [get]
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
