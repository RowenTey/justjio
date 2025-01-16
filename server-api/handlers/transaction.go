package handlers

import (
	"errors"
	"sc2006-JustJio/database"
	"sc2006-JustJio/model/request"
	"sc2006-JustJio/services"
	"sc2006-JustJio/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func GetTransactionsByUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	isPaid := c.QueryBool("isPaid", false)

	transactions, err := (&services.TransactionService{DB: database.DB}).GetTransactionsByUser(isPaid, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "No transactions found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved transactions successfully", transactions)
}

func SettleTransaction(c *fiber.Ctx) error {
	var request request.SettleTransactionRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	// TODO: Check if its the correct user making the request
	err := (&services.TransactionService{DB: database.DB}).SettleTransaction(request.TransactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "No transactions found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Paid transactions successfully", nil)
}
