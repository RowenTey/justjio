package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model/request"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var billLogger = log.WithFields(log.Fields{"service": "BillHandler"})

func CreateBill(c *fiber.Ctx) error {
	var request request.CreateBillRequest
	if err := c.BodyParser(&request); err != nil {
		return util.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")

	roomService := services.NewRoomService(database.DB)
	userService := services.NewUserService(database.DB)
	billService := services.NewBillService(database.DB)

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
		if err.Error() == "payers of a bill can't be empty" {
			return util.HandleInvalidInputError(c, err)
		}
		return util.HandleInternalServerError(c, err)
	}

	billLogger.Info("Created bill successfully: ", bill.ID)
	return util.HandleSuccess(c, "Created bill successfully", bill)
}

func GetBillsByRoom(c *fiber.Ctx) error {
	roomId := c.Query("roomId")
	if roomId == "" {
		return util.HandleInvalidInputError(c, errors.New("missing roomId in query param"))
	}

	bills, err := services.NewBillService(database.DB).GetBillsForRoom(roomId)
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

	billService := services.NewBillService(tx)
	transactionService := services.NewTransactionService(tx)
	roomService := services.NewRoomService(tx)

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

	billLogger.Info("Bills consolidated successfully: ", consolidation.ID)
	return util.HandleSuccess(c, "Bill consolidated successfully", nil)
}

func IsRoomBillConsolidated(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	if roomId == "" {
		return util.HandleInvalidInputError(c, errors.New("missing roomId in path param"))
	}

	isConsolidated, err := services.NewBillService(database.DB).IsRoomBillConsolidated(roomId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return util.HandleInternalServerError(c, err)
	}

	return util.HandleSuccess(c, "Retrieved consolidation status successfully", fiber.Map{
		"isConsolidated": isConsolidated,
	})
}
