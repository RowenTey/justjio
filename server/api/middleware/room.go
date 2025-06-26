package middleware

import (
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func IsUserInRoom(c *fiber.Ctx, roomService *services.RoomService) error {
	// Check if user is in room
	token := c.Locals("user").(*jwt.Token)
	userId := utils.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	userIds, err := roomService.GetRoomAttendeesIds(roomId)
	if err != nil {
		return utils.HandleNotFoundOrInternalError(c, err, "Room not found")
	}

	// Check if user is in room
	for _, id := range *userIds {
		if id == userId {
			c.Locals("roomUserIds", userIds)
			return c.Next()
		}
	}

	return utils.HandleError(c, fiber.StatusUnauthorized, "User is not in room", nil)
}
