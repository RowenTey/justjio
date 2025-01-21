package handlers

import (
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

func SettleTransaction(c *fiber.Ctx) error {
	txId := c.Params("txId")
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	err := (&services.TransactionService{DB: database.DB}).SettleTransaction(txId, userId)
	if err != nil {
		if err.Error() == "Invalid payer" {
			return util.HandleError(c, fiber.StatusUnauthorized, err.Error(), nil)
		}
		if err.Error() == "Transaction already settled" {
			return util.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return util.HandleNotFoundOrInternalError(c, err, "Transaction not found")
	}

	return util.HandleSuccess(c, "Paid transactions successfully", nil)
}
