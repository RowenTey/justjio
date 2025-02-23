package handlers

import (
	"fmt"
	"log"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func GetTransactionsByUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	isPaid := c.QueryBool("isPaid", false)

	transactions, err := (&services.TransactionService{DB: database.DB}).GetTransactionsByUser(isPaid, userId)
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

	transaction, err := (&services.TransactionService{DB: database.DB}).SettleTransaction(txId, userId)
	if err != nil {
		if err.Error() == "invalid payer" {
			return util.HandleError(c, fiber.StatusUnauthorized, err.Error(), nil)
		}
		if err.Error() == "transaction already settled" {
			return util.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return util.HandleNotFoundOrInternalError(c, err, "Transaction not found")
	}

	// Send notification to payee
	go func() {
		title := "Settled"
		message := fmt.Sprintf("%s paid you $%.2f!", username, transaction.Amount)

		notificationService := &services.NotificationService{DB: database.DB}
		_, err := notificationService.CreateNotification(transaction.PayeeID, title, message)
		if err != nil {
			log.Println("[TRANSACTION] Error creating notification: ", err)
			return
		}

		subscriptionService := &services.SubscriptionService{DB: database.DB}
		subscriptions, err := subscriptionService.GetSubscriptionsByUserID(transaction.PayeeID)
		if err != nil {
			log.Println("[TRANSACTION] Error getting subscriptions: ", err)
			return
		}

		for _, sub := range *subscriptions {
			notificationsChan <- NotificationData{
				Subscription: subscriptionService.NewWebPushSubscriptionObj(&sub),
				Title:        title,
				Message:      message,
			}
		}
	}()

	return util.HandleSuccess(c, "Paid transactions successfully", nil)
}
