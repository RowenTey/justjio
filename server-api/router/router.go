package router

import (
	"github.com/RowenTey/JustJio/handlers"
	"github.com/RowenTey/JustJio/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func Initalize(router *fiber.App) {
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello, World!")
	})

	v1 := router.Group("/v1")

	/* public routes */

	v1.Get("/swagger/*", swagger.HandlerDefault)

	auth := v1.Group("/auth")
	auth.Post("/", handlers.Login)
	auth.Post("/signup", handlers.SignUp)
	auth.Post("/verify", handlers.VerifyOTP)

	/* private routes */

	users := v1.Group("/users")
	users.Get("/:userId", handlers.GetUser)
	users.Patch("/:userId", handlers.UpdateUser)
	users.Delete("/:userId", handlers.DeleteUser)
	users.Get("/:userId/friends", handlers.GetFriends)
	users.Post("/:userId/friends", handlers.AddFriend)
	users.Post("/:userId/friends/check", handlers.IsFriend)
	users.Delete("/:userId/friends", handlers.RemoveFriend)

	rooms := v1.Group("/rooms")
	rooms.Get("/", handlers.GetRooms)
	rooms.Get("/invites", handlers.GetRoomInvitations)
	rooms.Get("/:roomId", middleware.IsUserInRoom, handlers.GetRoom)
	rooms.Get("/:roomId/attendees", middleware.IsUserInRoom, handlers.GetRoomAttendees)
	rooms.Post("/", handlers.CreateRoom)
	rooms.Post("/:roomId", middleware.IsUserInRoom, handlers.InviteUser)
	rooms.Patch("/:roomId", handlers.RespondToRoomInvite)
	rooms.Patch("/close/:roomId", middleware.IsUserInRoom, handlers.CloseRoom)
	rooms.Patch("/leave/:roomId", middleware.IsUserInRoom, handlers.LeaveRoom)

	messages := rooms.Group("/:roomId/messages")
	messages.Use(middleware.IsUserInRoom)
	messages.Get("/", handlers.GetMessages)
	messages.Get("/:msgId", handlers.GetMessage)
	messages.Post("/", handlers.CreateMessage)

	bills := v1.Group("/bills")
	bills.Get("/", handlers.GetBillsByRoom)
	bills.Post("/", handlers.CreateBill)
	bills.Post("/consolidate", handlers.ConsolidateBills)

	transactions := v1.Group("/transactions")
	transactions.Get("/", handlers.GetTransactionsByUser)
	transactions.Patch("/", handlers.SettleTransaction)

	// 404 handler
	router.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "404: Not Found",
		})
	})
}
