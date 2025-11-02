package middlewares

import (
	"slices"

	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/otel"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func IsUserInRoom(c *fiber.Ctx, roomService *services.RoomService) error {
	ctx := otel.GetOtelContext(c)
	// Check if user is in room
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	if _, err := roomService.GetRoomById(ctx, roomId); err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	userIds, err := roomService.GetRoomAttendeesIds(ctx, roomId)
	if err != nil {
		return utils.HandleInternalServerError(c, err)
	}

	// Check if user is in room
	if !slices.Contains(userIds, userId) {
		return utils.HandleError(c, fiber.StatusUnauthorized, "User is not in room", nil)
	}

	c.Locals("roomUserIds", userIds)
	return c.Next()
}
