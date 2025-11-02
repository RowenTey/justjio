package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/otel"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type TransactionHandler struct {
	transactionService  services.TransactionService
	notificationService *services.NotificationService
	logger              *log.Entry
}

func NewTransactionHandler(
	transactionService services.TransactionService,
	notificationService *services.NotificationService,
	logger *log.Logger,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService:  transactionService,
		notificationService: notificationService,
		logger:              logger.WithFields(logrus.Fields{"handler": "TransactionHandler"}),
	}
}

// GetTransactionsByUser retrieves transactions for a user
// @Summary Get user transactions
// @Description Retrieves transactions for the authenticated user, optionally filtered by payment status
// @Tags Transactions
// @Accept json
// @Produce json
// @Param isPaid query bool false "Filter by payment status" default(false)
// @Success 200 {object} object{status=string,message=string,data=[]response.TransactionDto} "Retrieved transactions successfully"
// @Failure 404 {object} utils.EmptyApiResponse "No transactions found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /transactions [get]
func (h *TransactionHandler) GetTransactionsByUser(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")
	isPaid := c.QueryBool("isPaid", false)

	transactions, err := h.transactionService.GetTransactionsByUser(ctx, isPaid, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No transactions found")
	}

	return utils.HandleSuccess(c, "Retrieved transactions successfully", transactions)
}

// SettleTransaction settles a transaction
// @Summary Settle transaction
// @Description Marks a transaction as paid by the authenticated user
// @Tags Transactions
// @Accept json
// @Produce json
// @Param txId path string true "Transaction ID"
// @Success 200 {object} utils.EmptyApiResponse "Paid transactions successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid payer or bad request"
// @Failure 404 {object} utils.EmptyApiResponse "Transaction not found"
// @Failure 409 {object} utils.EmptyApiResponse "Transaction already settled"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /transactions/{txId}/settle [patch]
func (h *TransactionHandler) SettleTransaction(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	txId := c.Params("txId")
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	username := utils.GetUserInfoFromToken(token, "username")

	transaction, err := h.transactionService.SettleTransaction(ctx, txId, userId)
	if err != nil {
		if errors.Is(err, services.ErrTransactionAlreadySettled) {
			return utils.HandleError(c, fiber.StatusConflict, err.Error(), nil)
		} else if errors.Is(err, services.ErrInvalidPayer) {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Transaction not found")
	}

	// TODO: Handle retries
	// Send notification to payee
	go func() {
		title := "Settled"
		message := fmt.Sprintf("%s paid you $%.2f!", username, transaction.Amount)
		if err := h.
			notificationService.
			SendNotification(ctx, utils.UIntToString(transaction.PayeeID), title, message); err != nil {
			h.logger.Errorf("Failed to send notification: %v", err)
		}
	}()

	return utils.HandleSuccess[any](c, "Paid transactions successfully", nil)
}
