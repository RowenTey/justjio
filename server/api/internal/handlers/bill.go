package handlers

import (
	"errors"

	"github.com/RowenTey/JustJio/server/api/internal/middlewares"
	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/request"
	"github.com/RowenTey/JustJio/server/api/pkg/otel"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
	"github.com/sirupsen/logrus"
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
		logger:      logger.WithFields(logrus.Fields{"handler": "BillHandler"}),
	}
}

// CreateBill creates a new bill for a room
// @Summary Create bill
// @Description Creates a new bill for a room with specified payers and amount
// @Tags Bills
// @Accept json
// @Produce json
// @Param bill body request.CreateBillRequest true "Bill creation details"
// @Success 200 {object} object{status=string,message=string,data=string} "Created bill successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input, empty payers, or room already consolidated"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found or payers not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /bills [post]
func (h *BillHandler) CreateBill(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	req := middlewares.GetValidatedRequest[request.CreateBillRequest](c)

	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	createdBillId, err := h.billService.CreateBill(
		ctx,
		req.RoomID,
		userId,
		req.Payers,
		req.Name,
		req.Amount,
		req.IncludeOwner,
	)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyConsolidated) {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		} else if errors.Is(err, services.ErrPayersNotFound) {
			return utils.HandleError(c, fiber.StatusNotFound, err.Error(), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	h.logger.Info("Created bill successfully: ", createdBillId)
	return utils.HandleSuccess(c, "Created bill successfully", createdBillId)
}

// GetBillsByRoom retrieves bills for a specific room
// @Summary Get bills by room
// @Description Retrieves all bills for a specified room
// @Tags Bills
// @Accept json
// @Produce json
// @Param roomId query string true "Room ID"
// @Success 200 {object} object{status=string,message=string,data=[]response.BillDto} "Retrieved bills successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Missing roomId in query parameter"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /bills [get]
func (h *BillHandler) GetBillsByRoom(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	roomId := c.Query("roomId")
	if roomId == "" {
		return utils.HandleInvalidInputError(c, errors.New("missing roomId in query param"))
	}

	bills, err := h.billService.GetBillsForRoom(ctx, roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	return utils.HandleSuccess(c, "Retrieved bills successfully", bills)
}

// ConsolidateBills consolidates all bills for a room
// @Summary Consolidate bills
// @Description Consolidates all bills for a room into transactions (host only)
// @Tags Bills
// @Accept json
// @Produce json
// @Param consolidation body request.ConsolidateBillsRequest true "Consolidation request"
// @Success 200 {object} utils.EmptyApiResponse "Bill consolidated successfully"
// @Failure 400 {object} utils.EmptyApiResponse "Invalid input or bills already consolidated"
// @Failure 403 {object} utils.EmptyApiResponse "Only host can consolidate bills"
// @Failure 404 {object} utils.EmptyApiResponse "Room not found"
// @Failure 500 {object} utils.EmptyApiResponse "Internal server error"
// @Security BearerAuth
// @Router /bills/consolidate [post]
func (h *BillHandler) ConsolidateBills(c *fiber.Ctx) error {
	ctx := otel.GetOtelContext(c)
	req := middlewares.GetValidatedRequest[request.ConsolidateBillsRequest](c)

	userId := utils.GetUserInfoFromToken(c.Locals("user").(*jwt.Token), "user_id")

	if err := h.billService.ConsolidateBills(ctx, req.RoomID, userId); err != nil {
		if errors.Is(err, services.ErrAlreadyConsolidated) {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), nil)
		} else if errors.Is(err, services.ErrOnlyHostCanConsolidate) {
			return utils.HandleError(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return utils.HandleNotFoundOrInternalError(c, err, RoomNotFoundErrorMsg)
	}

	return utils.HandleSuccess[any](c, "Bill consolidated successfully", nil)
}
