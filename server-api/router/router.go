package router

import (
	"github.com/RowenTey/JustJio/handlers"
	"github.com/RowenTey/JustJio/middleware"
	"github.com/RowenTey/JustJio/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func Initalize(router *fiber.App, kafkaSvc *services.KafkaService) {
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello world from JustJio :)")
	})

	v1 := router.Group("/v1")

	/* public routes */

	v1.Get("/swagger/*", swagger.HandlerDefault)

	auth := v1.Group("/auth")
	auth.Post("/", func(c *fiber.Ctx) error {
		return handlers.Login(c, kafkaSvc)
	})
	auth.Post("/signup", handlers.SignUp)
	auth.Post("/verify", handlers.VerifyOTP)

	/* private routes */

	users := v1.Group("/users")
	users.Get("/:userId", handlers.GetUser)
	users.Patch("/:userId", handlers.UpdateUser)
	users.Delete("/:userId", handlers.DeleteUser)

	friends := users.Group("/:userId/friends")
	friends.Get("/", handlers.GetFriends)
	friends.Post("/check", handlers.IsFriend)
	friends.Get("/count", handlers.GetNumFriends)
	friends.Get("/search", handlers.SearchFriends)
	friends.Delete("/", handlers.RemoveFriend)

	friendRequests := users.Group("/:userId/friend-requests")
	friendRequests.Get("/", handlers.GetFriendRequestsByStatus)
	friendRequests.Post("/", handlers.SendFriendRequest)
	friendRequests.Patch("/", handlers.RespondToFriendRequest)

	rooms := v1.Group("/rooms")
	rooms.Get("/", handlers.GetRooms)
	rooms.Get("/count", handlers.GetNumRooms)
	rooms.Get("/invites", handlers.GetRoomInvitations)
	rooms.Get("/invites/count", handlers.GetNumRoomInvitations)
	rooms.Get("/:roomId", middleware.IsUserInRoom, handlers.GetRoom)
	rooms.Get("/:roomId/attendees", middleware.IsUserInRoom, handlers.GetRoomAttendees)
	rooms.Post("/", handlers.CreateRoom)
	rooms.Post("/:roomId", middleware.IsUserInRoom, handlers.InviteUser)
	rooms.Patch("/:roomId", handlers.RespondToRoomInvite)
	rooms.Patch("/:roomId/close", middleware.IsUserInRoom, handlers.CloseRoom)
	rooms.Patch("/:roomId/leave", middleware.IsUserInRoom, handlers.LeaveRoom)

	messages := rooms.Group("/:roomId/messages")
	messages.Use(middleware.IsUserInRoom)
	messages.Get("/", handlers.GetMessages)
	messages.Get("/:msgId", handlers.GetMessage)
	messages.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateMessage(c, kafkaSvc)
	})

	bills := v1.Group("/bills")
	bills.Get("/", handlers.GetBillsByRoom)
	bills.Post("/", handlers.CreateBill)
	bills.Post("/consolidate", handlers.ConsolidateBills)

	transactions := v1.Group("/transactions")
	transactions.Get("/", handlers.GetTransactionsByUser)
	transactions.Patch("/:txId/settle", handlers.SettleTransaction)

	// 404 handler
	router.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "404: Endpoint Not Found",
		})
	})
}
