package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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

	isConsolidated, err := billService.IsRoomBillConsolidated(request.RoomID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return util.HandleInternalServerError(c, err)
	}

	if isConsolidated {
		return util.HandleError(
			c, fiber.StatusBadRequest, "Bills for this room have already been consolidated", nil)
	}

	room, err := roomService.GetRoomById(request.RoomID)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	owner, err := userService.GetUserByID(userId)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "Owner not found")
	}

	payers, err := userService.GetUsersByID(request.Payers)
	if err != nil {
		return util.HandleNotFoundOrInternalError(c, err, "Payer not found")
	}

	bill, err := billService.CreateBill(
		room,
		owner,
		request.Name,
		request.Amount,
		request.IncludeOwner,
		payers,
	)
	if err != nil {
		if err.Error() == "Payers of a bill can't be empty" {
			return util.HandleInvalidInputError(c, err)
		}
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Created bill successfully", bill)
}

func GetBillsByRoom(c *fiber.Ctx) error {
	roomId := c.Query("roomId")
	if roomId == "" {
		return util.HandleInvalidInputError(c, errors.New("Missing roomId in query param"))
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

	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	tx := database.DB.Begin()

	billService := &services.BillService{DB: tx}
	transactionService := &services.TransactionService{DB: tx}
	roomService := &services.RoomService{DB: tx}

	room, err := roomService.GetRoomById(request.RoomID)
	if err != nil {
		tx.Rollback()
		return util.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	if fmt.Sprint(room.HostID) != userId {
		tx.Rollback()
		return util.HandleError(c, fiber.StatusUnauthorized, "User is not the host of the room", nil)
	}

	isConsolidated, err := billService.IsRoomBillConsolidated(request.RoomID)
	if err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

	if isConsolidated {
		tx.Rollback()
		return util.HandleError(
			c, fiber.StatusBadRequest, "Bills for this room have already been consolidated", nil)
	}

	consolidation, err := billService.ConsolidateBills(request.RoomID)
	if err != nil {
		tx.Rollback()
		return util.HandleInternalServerError(c, err)
	}

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
