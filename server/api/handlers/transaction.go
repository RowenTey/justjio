package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type TransactionHandler struct {
	transactionService  *services.TransactionService
	notificationService *services.NotificationService
	logger              *log.Entry
}

func NewTransactionHandler(
	transactionService *services.TransactionService,
	notificationService *services.NotificationService,
	logger *log.Logger,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService:  transactionService,
		notificationService: notificationService,
		logger:              utils.AddServiceField(logger, "TransactionHandler"),
	}
}

func (h *TransactionHandler) GetTransactionsByUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	isPaid := c.QueryBool("isPaid", false)

	transactions, err := h.transactionService.GetTransactionsByUser(isPaid, userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "No transactions found")
	}
	return utils.HandleSuccess(c, "Retrieved transactions successfully", transactions)
}

func (h *TransactionHandler) SettleTransaction(c *fiber.Ctx) error {
	txId := c.Params("txId")
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	username := utils.GetUserInfoFromToken(token, "username")

	transaction, err := h.transactionService.SettleTransaction(txId, userId)
	if err != nil {
		if errors.Is(err, services.ErrTransactionAlreadySettled) {
			return utils.HandleError(c, fiber.StatusUnauthorized, err.Error(), nil)
		} else if errors.Is(err, services.ErrInvalidPayer) {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err, "Transaction not found")
	}

	// Send notification to payee
	go func() {
		title := "Settled"
		message := fmt.Sprintf("%s paid you $%.2f!", username, transaction.Amount)
		if err := h.
			notificationService.
			SendNotification(utils.UIntToString(transaction.PayeeID), title, message); err != nil {
			h.logger.Errorf("Failed to send notification: %v", err)
		}
	}()

	return utils.HandleSuccess(c, "Paid transactions successfully", nil)
}
