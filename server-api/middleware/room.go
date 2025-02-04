package middleware

import (
	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func IsUserInRoom(c *fiber.Ctx) error {
	// Check if user is in room
	token := c.Locals("user").(*jwt.Token)
	userId := util.GetUserInfoFromToken(token, "user_id")
	roomId := c.Params("roomId")

	// log.Println("[ROOM Middleware] roomId: ", roomId)

	userIds, err := (&services.RoomService{DB: database.DB}).GetRoomAttendeesIds(roomId)
	if err != nil {
		return util.HandleInternalServerError(c, err)
	}

	// log.Println("[ROOM Middleware] userId: ", userId)
	// log.Println("[ROOM Middleware] userIds: ", userIds)

	// Check if user is in room
	for _, id := range *userIds {
		if id == userId {
			c.Locals("roomUserIds", userIds)
			return c.Next()
		}
	}

	return util.HandleError(c, fiber.StatusUnauthorized, "User is not in room", nil)
}
