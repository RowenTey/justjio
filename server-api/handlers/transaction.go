package handlers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var transactionLogger = log.WithFields(log.Fields{"service": "TransactionHandler"})

func GetTransactionsByUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	isPaid := c.QueryBool("isPaid", false)

	transactions, err := services.NewTransactionService(database.DB).GetTransactionsByUser(isPaid, userId)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "No transactions found")
	}

	return util.HandleSuccess(c, "Retrieved transactions successfully", transactions)
}

func SettleTransaction(c *fiber.Ctx, notificationsChan chan<- NotificationData) error {
	txId := c.Params("txId")
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	username := util.GetUserInfoFromToken(token, "username")

	tx := database.DB.Begin()

	transaction, err := services.NewTransactionService(tx).SettleTransaction(txId, userId)
	if err != nil {
		tx.Rollback()
		if err.Error() == "invalid payer" {
			return util.HandleError(c, fiber.StatusUnauthorized, err.Error(), nil)
		}
		if err.Error() == "transaction already settled" {
			return util.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return util.HandleNotFoundOrInternalError(c, err, "Transaction not found")
	}

	// Send notification to payee
	go func(dbTx *gorm.DB) {
		title := "Settled"
		message := fmt.Sprintf("%s paid you $%.2f!", username, transaction.Amount)

		notificationService := services.NewNotificationService(dbTx)
		_, err := notificationService.CreateNotification(transaction.PayeeID, title, message)
		if err != nil {
			tx.Rollback()
			transactionLogger.Error("Error creating notification: ", err)
			return
		}

		subscriptionService := services.NewSubscriptionService(dbTx)
		subscriptions, err := subscriptionService.GetSubscriptionsByUserID(transaction.PayeeID)
		if err != nil {
			tx.Rollback()
			transactionLogger.Error("Error getting subscriptions: ", err)
			return
		}

		for _, sub := range *subscriptions {
			notificationsChan <- NotificationData{
				Subscription: subscriptionService.NewWebPushSubscriptionObj(&sub),
				Title:        title,
				Message:      message,
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			transactionLogger.Error("Error committing transaction: ", err)
			return
		}
	}(tx)

	return util.HandleSuccess(c, "Paid transactions successfully", nil)
}
