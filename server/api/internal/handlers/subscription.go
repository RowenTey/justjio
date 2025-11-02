package handlers

import (
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/internal/middlewares"
	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/request"
	"github.com/RowenTey/JustJio/server/api/pkg/otel"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
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
		logger:              logger.WithFields(log.Fields{"handler": "SubscriptionHandler"}),
	}
}

// CreateSubscription creates a new push notification subscription
// @Summary Create subscription
// @Description Creates a new push notification subscription for the user
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscriptionRequest body request.CreateSubscriptionRequest true "Subscription details"
// @Success 200 {object} object{status=string,message=string,data=string} "Subscription created successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input or missing required fields"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	req := middlewares.GetValidatedRequest[request.CreateSubscriptionRequest](c)

	subscription := &models.Subscription{
		UserID:   req.UserID,
		Endpoint: req.Endpoint,
		Auth:     req.Auth,
		P256dh:   req.P256dh,
	}

	createdSubscriptionID, err := h.subscriptionService.CreateSubscription(ctx, subscription)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	h.logger.Info("Subscription created successfully: ", createdSubscriptionID)
	return utils.HandleSuccess(c, "Subscription created successfully", createdSubscriptionID)
}

// GetSubscriptionByEndpoint retrieves a subscription by endpoint
// @Summary Get subscription by endpoint
// @Description Retrieves a push notification subscription by its endpoint URL
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param endpoint path string true "URL-encoded subscription endpoint"
// @Success 200 {object} object{status=string,message=string,data=response.SubscriptionDto} "Subscription retrieved successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid endpoint URL"
// @Failure 404 {object} utils.EmptyApiResponse "Subscription not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /subscriptions/{endpoint} [get]
func (h *SubscriptionHandler) GetSubscriptionByEndpoint(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	endpoint := c.Params("endpoint")
	decodedEndpoint, err := url.QueryUnescape(endpoint)
	if err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	subscription, err := h.subscriptionService.GetSubscriptionsByEndpoint(ctx, decodedEndpoint)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Subscription not found")
	}

	return utils.HandleSuccess(c, "Subscription retrieved successfully", subscription)
}

// DeleteSubscription deletes a subscription
// @Summary Delete subscription
// @Description Deletes a push notification subscription by ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subId path string true "Subscription ID"
// @Success 200 {object} utils.EmptyApiResponse "Subscription deleted successfully"
// @Failure 404 {object} utils.EmptyApiResponse "Subscription not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /subscriptions/{subId} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	subId := c.Params("subId")

	if err := h.subscriptionService.DeleteSubscription(ctx, subId); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Subscription not found")
	}

	return utils.HandleSuccess[any](c, "Subscription deleted successfully", nil)
}
