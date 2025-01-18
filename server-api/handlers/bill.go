package handlers

import (
	"errors"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func CreateBill(c *fiber.Ctx) error {
	var request request.CreateBillRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	roomService := &services.RoomService{DB: database.DB}
	userService := &services.UserService{DB: database.DB}
	billService := &services.BillService{DB: database.DB}

	room, err := roomService.GetRoomById(request.RoomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Room not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	owner, err := userService.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Owner not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	payers, err := userService.GetUsersByID(request.Payers)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.HandleError(c, fiber.StatusNotFound, "Payer not found", err)
		}
		return util.HandleInternalServerError(c, err)
	}

	bill, err := billService.CreateBill(
		room,
		owner,
		request.Name,
		request.Amount,
		payers,
	)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Created bill successfully", bill)
}

func GetBillsByRoom(c *fiber.Ctx) error {
	roomId := c.Query("roomId")
	if roomId == "" {
		return util.HandleInvalidInputError(c, errors.New("missing roomId in query param"))
	}

	bills, err := (&services.BillService{DB: database.DB}).GetBillsForRoom(roomId)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved bills successfully", bills)
}

func ConsolidateBills(c *fiber.Ctx) error {
	var request request.ConsolidateBillsRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	tx := database.DB.Begin()

	billService := &services.BillService{DB: tx}
	transactionService := &services.TransactionService{DB: tx}

	consolidation, err := billService.ConsolidateBills(request.RoomID)
	if err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	// TODO: Should this be blocking or non-blocking
	if err := transactionService.GenerateTransactions(consolidation); err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Bill consolidated successfully", nil)
}
