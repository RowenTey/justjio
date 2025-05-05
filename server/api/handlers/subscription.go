package handlers

import (
	"errors"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	model_push_notifications "github.com/RowenTey/JustJio/server/api/model/push_notifications"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
)

type NotificationData = model_push_notifications.NotificationData

var subscriptionLogger = log.WithField("service", "SubscriptionHandler")

func CreateSubscription(c *fiber.Ctx, notificationsChan chan<- NotificationData) error {
	var subscription model.Subscription
	if err := c.BodyParser(&subscription); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if subscription.Auth == "" || subscription.P256dh == "" || subscription.Endpoint == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing required fields"))
	}

	subscriptionService := services.NewSubscriptionService(database.DB)
	createdSubscription, err := subscriptionService.CreateSubscription(&subscription)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	notificationsChan <- NotificationData{
		Subscription: subscriptionService.NewWebPushSubscriptionObj(createdSubscription),
		Title:        "Welcome",
		Message:      "Subscribed to JustJio! You will now receive notifications for app events.",
	}

	subscriptionLogger.Info("Subscription created successfully: ", createdSubscription.ID)
	return utils.HandleSuccess(c, "Subscription created successfully", createdSubscription)
}

func GetSubscriptionByEndpoint(c *fiber.Ctx) error {
	endpoint := c.Params("endpoint")
	decodedEndpoint, err := url.QueryUnescape(endpoint)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	subscriptionService := services.NewSubscriptionService(database.DB)
	subscription, err := subscriptionService.GetSubscriptionsByEndpoint(decodedEndpoint)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Subscription not found")
	}

	return utils.HandleSuccess(c, "Subscription retrieved successfully", subscription)
}

func DeleteSubscription(c *fiber.Ctx) error {
	subId := c.Params("subId")
	subscriptionService := services.NewSubscriptionService(database.DB)
	if err := subscriptionService.DeleteSubscription(subId); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Subscription not found")
	}

	return utils.HandleSuccess(c, "Subscription deleted successfully", nil)
}
