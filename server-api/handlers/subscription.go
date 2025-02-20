package handlers

import (
	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model"
	model_push_notifications "github.com/RowenTey/JustJio/model/push_notifications"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"
	"github.com/gofiber/fiber/v2"
)

type NotificationData = model_push_notifications.NotificationData

func CreateSubscription(c *fiber.Ctx, notificationsChan chan<- NotificationData) error {
	var subscription model.Subscription
	if err := c.BodyParser(&subscription); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	subscriptionService := &services.SubscriptionService{DB: database.DB}
	createdSubscription, err := subscriptionService.CreateSubscription(&subscription)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	notificationsChan <- NotificationData{
		Subscription: subscriptionService.NewWebPushSubscriptionObj(createdSubscription),
		Title:        "Welcome",
		Message:      "Subscribed to JustJio! You will now receive notifications for app events.",
	}

	return util.HandleSuccess(c, "Subscription created successfully", createdSubscription)
}

func DeleteSubscription(c *fiber.Ctx) error {
	subId := c.Params("subId")
	subscriptionService := &services.SubscriptionService{DB: database.DB}
	if err := subscriptionService.DeleteSubscription(subId); err != nil {
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Subscription deleted successfully", nil)
}
