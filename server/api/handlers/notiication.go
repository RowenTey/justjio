package handlers

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var notificationLogger = log.WithFields(log.Fields{"service": "NotificationHandler"})

// CreateNotification handles the creation of a new notification
func CreateNotification(c *fiber.Ctx, notificationsChan chan<- NotificationData) error {
	var request request.CreateNotificationRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	notification, err := services.NewNotificationService(database.DB).CreateNotification(
		request.UserId, request.Title, request.Content)
	if err != nil {
		if err.Error() == "content cannot be empty" {
			return utils.HandleInvalidInputError(c, err)
		}
		return utils.HandleInternalServerError(c, err)
	}

	go func() {
		subscriptionService := services.NewSubscriptionService(database.DB)
		subscriptions, err := subscriptionService.GetSubscriptionsByUserID(request.UserId)
		if err != nil {
			notificationLogger.Error("Error getting subscriptions: ", err)
			return
		}

		for _, sub := range *subscriptions {
			notificationsChan <- NotificationData{
				Subscription: subscriptionService.NewWebPushSubscriptionObj(&sub),
				Title:        notification.Title,
				Message:      notification.Content,
			}
		}
	}()

	return utils.HandleSuccess(c, "Notification created successfully", notification)
}

// MarkNotificationAsRead handles marking a notification as read
func MarkNotificationAsRead(c *fiber.Ctx) error {
	notificationId, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	userId, err := strconv.ParseUint(c.Params("userId"), 10, 32)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	notificationLogger.Infof("Marking notification %d as read for user %d", notificationId, userId)
	if err := services.
		NewNotificationService(database.DB).
		MarkNotificationAsRead(uint(notificationId), uint(userId)); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Notification not found")
	}

	return utils.HandleSuccess(c, "Notification marked as read successfully", nil)
}

// GetNotification handles retrieving a single notification
func GetNotification(c *fiber.Ctx) error {
	notificationId, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	userId, err := strconv.ParseUint(c.Params("userId"), 10, 32)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	notification, err := services.
		NewNotificationService(database.DB).
		GetNotification(uint(notificationId), uint(userId))
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Notification not found")
	}

	return utils.HandleSuccess(c, "Retrieved notification successfully", notification)
}

// GetNotifications handles retrieving all notifications for a user
func GetNotifications(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	userIdInt, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	notifications, err := services.NewNotificationService(database.DB).GetNotifications(uint(userIdInt))
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "User not found")
	}

	return utils.HandleSuccess(c, "Retrieved notifications successfully", notifications)
}
