package handlers

import (
	"errors"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
	logger              *log.Entry
}

func NewSubscriptionHandler(
	subscriptionService *services.SubscriptionService,
	logger *log.Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		logger:              utils.AddServiceField(logger, "SubscriptionHandler"),
	}
}

func (h *SubscriptionHandler) CreateSubscription(c *fiber.Ctx) error {
	var subscription model.Subscription
	if err := c.BodyParser(&subscription); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	if subscription.Auth == "" || subscription.P256dh == "" || subscription.Endpoint == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing required fields"))
	}

	createdSubscription, err := h.subscriptionService.CreateSubscription(&subscription)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	h.logger.Info("Subscription created successfully: ", createdSubscription.ID)
	return utils.HandleSuccess(c, "Subscription created successfully", createdSubscription)
}

func (h *SubscriptionHandler) GetSubscriptionByEndpoint(c *fiber.Ctx) error {
	endpoint := c.Params("endpoint")
	decodedEndpoint, err := url.QueryUnescape(endpoint)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	subscription, err := h.subscriptionService.GetSubscriptionsByEndpoint(decodedEndpoint)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Subscription not found")
	}

	return utils.HandleSuccess(c, "Subscription retrieved successfully", subscription)
}

func (h *SubscriptionHandler) DeleteSubscription(c *fiber.Ctx) error {
	subId := c.Params("subId")

	if err := h.subscriptionService.DeleteSubscription(subId); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Subscription not found")
	}

	return utils.HandleSuccess(c, "Subscription deleted successfully", nil)
}
