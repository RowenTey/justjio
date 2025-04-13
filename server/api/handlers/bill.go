package handlers

import (
	"errors"
	"fmt"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var billLogger = log.WithFields(log.Fields{"service": "BillHandler"})

func CreateBill(c *fiber.Ctx) error {
	var request request.CreateBillRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	roomService := services.NewRoomService(database.DB)
	userService := services.NewUserService(database.DB)
	billService := services.NewBillService(database.DB)

	isConsolidated, err := billService.IsRoomBillConsolidated(request.RoomID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.HandleInternalServerError(c, err)
	}

	if isConsolidated {
		return utils.HandleError(
			c, fiber.StatusBadRequest, "Bills for this room have already been consolidated", nil)
	}

	room, err := roomService.GetRoomById(request.RoomID)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	owner, err := userService.GetUserByID(userId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Owner not found")
	}

	payers, err := userService.GetUsersByID(request.Payers)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Payer(s) not found")
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
			return utils.HandleInvalidInputError(c, err)
		}
		return utils.HandleInternalServerError(c, err)
	}

	billLogger.Info("Created bill successfully: ", bill.ID)
	return utils.HandleSuccess(c, "Created bill successfully", bill)
}

func GetBillsByRoom(c *fiber.Ctx) error {
	roomId := c.Query("roomId")
	if roomId == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing roomId in query param"))
	}

	bills, err := services.NewBillService(database.DB).GetBillsForRoom(roomId)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	return utils.HandleSuccess(c, "Retrieved bills successfully", bills)
}

func ConsolidateBills(c *fiber.Ctx) error {
	var request request.ConsolidateBillsRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	tx := database.DB.Begin()

	billService := services.NewBillService(tx)
	transactionService := services.NewTransactionService(tx)
	roomService := services.NewRoomService(tx)

	room, err := roomService.GetRoomById(request.RoomID)
	if err != nil {
		tx.Rollback()
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	if fmt.Sprint(room.HostID) != userId {
		tx.Rollback()
		return utils.HandleError(c, fiber.StatusUnauthorized, "User is not the host of the room", nil)
	}

	isConsolidated, err := billService.IsRoomBillConsolidated(request.RoomID)
	if err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	if isConsolidated {
		tx.Rollback()
		return utils.HandleError(
			c, fiber.StatusBadRequest, "Bills for this room have already been consolidated", nil)
	}

	consolidation, err := billService.ConsolidateBills(tx, request.RoomID)
	if err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	if err := transactionService.GenerateTransactions(consolidation); err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.HandleInternalServerError(c, err)
	}

	billLogger.Info("Bills consolidated successfully: ", consolidation.ID)
	return utils.HandleSuccess(c, "Bill consolidated successfully", nil)
}

func IsRoomBillConsolidated(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	if roomId == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing roomId in path param"))
	}

	isConsolidated, err := services.NewBillService(database.DB).IsRoomBillConsolidated(roomId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.HandleInternalServerError(c, err)
	}

	return utils.HandleSuccess(c, "Retrieved consolidation status successfully", fiber.Map{
		"isConsolidated": isConsolidated,
	})
}
