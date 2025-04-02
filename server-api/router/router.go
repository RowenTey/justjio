package router

import (
	"github.com/RowenTey/JustJio/handlers"
	"github.com/RowenTey/JustJio/middleware"
	model_push_notifications "github.com/RowenTey/JustJio/model/push_notifications"
	"github.com/RowenTey/JustJio/services"

	"github.com/gofiber/fiber/v2"
)

type NotificationData = model_push_notifications.NotificationData

func Initalize(router *fiber.App, kafkaSvc *services.KafkaService, notificationsChan chan<- NotificationData) {
	// healthcheck endpoint
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello world from JustJio API :)")
	})

	v1 := router.Group("/v1")

	/* public routes */

	auth := v1.Group("/auth")
	auth.Post("/", func(c *fiber.Ctx) error {
		return handlers.Login(c, kafkaSvc)
	})
	auth.Post("/google", handlers.GoogleLogin)
	auth.Post("/signup", handlers.SignUp)
	auth.Post("/verify", handlers.VerifyOTP)
	auth.Post("/otp", handlers.SendOTPEmail)
	auth.Patch("/reset", handlers.ResetPassword)

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
	friends.Delete("/:friendId", handlers.RemoveFriend)

	friendRequests := users.Group("/:userId/friendRequests")
	friendRequests.Get("/", handlers.GetFriendRequestsByStatus)
	friendRequests.Get("/count", handlers.CountPendingFriendRequests)
	friendRequests.Post("/", handlers.SendFriendRequest)
	friendRequests.Patch("/", handlers.RespondToFriendRequest)

	userNotifications := users.Group("/:userId/notifications")
	userNotifications.Get("/:id", handlers.GetNotification)
	userNotifications.Patch("/:id", handlers.MarkNotificationAsRead)

	rooms := v1.Group("/rooms")
	rooms.Get("/", handlers.GetRooms)
	rooms.Get("/count", handlers.GetNumRooms)
	rooms.Get("/invites", handlers.GetRoomInvitations)
	rooms.Get("/invites/count", handlers.GetNumRoomInvitations)
	rooms.Get("/:roomId", middleware.IsUserInRoom, handlers.GetRoom)
	rooms.Get("/:roomId/attendees", middleware.IsUserInRoom, handlers.GetRoomAttendees)
	rooms.Get("/:roomId/uninvited", middleware.IsUserInRoom, handlers.GetUninvitedFriendsForRoom)
	rooms.Post("/", handlers.CreateRoom)
	rooms.Post("/:roomId", middleware.IsUserInRoom, handlers.InviteUser)
	rooms.Patch("/:roomId", handlers.RespondToRoomInvite)
	rooms.Patch("/:roomId/join", handlers.JoinRoom)
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
	bills.Get("/consolidate/:roomId", handlers.IsRoomBillConsolidated)
	bills.Post("/", handlers.CreateBill)
	bills.Post("/consolidate", handlers.ConsolidateBills)

	transactions := v1.Group("/transactions")
	transactions.Get("/", handlers.GetTransactionsByUser)
	transactions.Patch("/:txId/settle", func(c *fiber.Ctx) error {
		return handlers.SettleTransaction(c, notificationsChan)
	})

	notifications := v1.Group("/notifications")
	notifications.Get("/", handlers.GetNotifications)
	notifications.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateNotification(c, notificationsChan)
	})

	subscriptions := v1.Group("/subscriptions")
	subscriptions.Get("/:endpoint", handlers.GetSubscriptionByEndpoint)
	subscriptions.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateSubscription(c, notificationsChan)
	})
	subscriptions.Delete("/:subId", handlers.DeleteSubscription)

	// 404 handler
	router.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "404: Endpoint Not Found",
		})
	})
}
