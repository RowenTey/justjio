package handlers

import (
	"strconv"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// CreateNotification handles the creation of a new notification
func CreateNotification(c *fiber.Ctx) error {
	var input struct {
		UserId  uint   `json:"userId"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&input); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	notification, err := (&services.NotificationService{DB: database.DB}).CreateNotification(input.UserId, input.Content)
	if err != nil {
		if err.Error() == "content cannot be empty" {
			return util.HandleInvalidInputError(c, err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Notification created successfully", notification)
}

// MarkNotificationAsRead handles marking a notification as read
func MarkNotificationAsRead(c *fiber.Ctx) error {
	notificationId, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userId, err := strconv.ParseUint(c.Params("userId"), 10, 32)
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	if err := (&services.NotificationService{DB: database.DB}).MarkNotificationAsRead(uint(notificationId), uint(userId)); err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "Notification not found")
	}

	return util.HandleSuccess(c, "Notification marked as read successfully", nil)
}

// GetNotification handles retrieving a single notification
func GetNotification(c *fiber.Ctx) error {
	notificationId, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	userId, err := strconv.ParseUint(c.Params("userId"), 10, 32)
	if err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	notification, err := (&services.NotificationService{DB: database.DB}).GetNotification(uint(notificationId), uint(userId))
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "Notification not found")
	}

	return util.HandleSuccess(c, "Retrieved notification successfully", notification)
}

// GetNotifications handles retrieving all notifications for a user
func GetNotifications(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	userIdInt, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	notifications, err := (&services.NotificationService{DB: database.DB}).GetNotifications(uint(userIdInt))
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	return util.HandleSuccess(c, "Retrieved notifications successfully", notifications)
}
