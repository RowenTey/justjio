package handlers

import (
	"errors"

	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var (
	RoomNotFoundErrorMsg = "Room not found"
)

type BillHandler struct {
	billService *services.BillService
	logger      *log.Entry
}

func NewBillHandler(
	billService *services.BillService,
	logger *log.Logger,
) *BillHandler {
	return &BillHandler{
		billService: billService,
		logger:      utils.AddServiceField(logger, "BillHandler"),
	}
}

func (h *BillHandler) CreateBill(c *fiber.Ctx) error {
	var request request.CreateBillRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	bill, err := h.billService.CreateBill(
		request.RoomID,
		userId,
		&request.Payers,
		request.Name,
		request.Amount,
		request.IncludeOwner,
	)
	if err != nil {
		if errors.Is(err, services.ErrEmptyPayers) {
			return utils.HandleInvalidInputError(c, err)
		} else if errors.Is(err, services.ErrAlreadyConsolidated) {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		} else if errors.Is(err, services.ErrPayersNotFound) {
			return utils.HandleError(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	h.logger.Info("Created bill successfully: ", bill.ID)
	return utils.HandleSuccess(c, "Created bill successfully", bill)
}

func (h *BillHandler) GetBillsByRoom(c *fiber.Ctx) error {
	roomId := c.Query("roomId")
	if roomId == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing roomId in query param"))
	}

	bills, err := h.billService.GetBillsForRoom(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	return utils.HandleSuccess(c, "Retrieved bills successfully", bills)
}

func (h *BillHandler) ConsolidateBills(c *fiber.Ctx) error {
	var request request.ConsolidateBillsRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.HandleInvalidInputError(c, err)
	}

	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")

	if err := h.billService.ConsolidateBills(request.RoomID, userId); err != nil {
		if errors.Is(err, services.ErrAlreadyConsolidated) {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		} else if errors.Is(err, services.ErrOnlyHostCanConsolidate) {
			return utils.HandleError(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	return utils.HandleSuccess(c, "Bill consolidated successfully", nil)
}

func (h *BillHandler) IsRoomBillConsolidated(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	if roomId == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing roomId in path param"))
	}

	status, err := h.billService.GetRoomBillConsolidationStatus(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	return utils.HandleSuccess(c,
		"Retrieved consolidation status successfully",
		fiber.Map{
			"isConsolidated": status == repository.CONSOLIDATED,
		})
}
