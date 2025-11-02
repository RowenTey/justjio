package router

import (
	"github.com/RowenTey/JustJio/server/api/internal/handlers"
	"github.com/RowenTey/JustJio/server/api/internal/middlewares"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/internal/services"
	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/request"
	"github.com/RowenTey/JustJio/server/api/pkg/http"
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type Repositories struct {
	UserRepository         repositories.UserRepository
	RoomRepository         repositories.RoomRepository
	MessageRepository      repositories.MessageRepository
	BillRepository         repositories.BillRepository
	TransactionRepository  repositories.TransactionRepository
	SubscriptionRepository repositories.SubscriptionRepository
	NotificationRepository repositories.NotificationRepository
}

type Services struct {
	UserService         *services.UserService
	AuthService         *services.AuthService
	RoomService         *services.RoomService
	MessageService      *services.MessageService
	BillService         *services.BillService
	TransactionService  services.TransactionService
	SubscriptionService *services.SubscriptionService
	NotificationService *services.NotificationService
}

type Handlers struct {
	UserHandler         *handlers.UserHandler
	AuthHandler         *handlers.AuthHandler
	RoomHandler         *handlers.RoomHandler
	MessageHandler      *handlers.MessageHandler
	BillHandler         *handlers.BillHandler
	TransactionHandler  *handlers.TransactionHandler
	NotificationHandler *handlers.NotificationHandler
	SubscriptionHandler *handlers.SubscriptionHandler
}

func Initalize(ctx *app.Context) {
	router := ctx.App
	logger := ctx.Logger

	// initialize repositories
	repositories := initRepositories(ctx.DB)

	// initialize services
	appServices := initServices(
		ctx,
		repositories,
	)

	// custom middlewares
	roomMiddleware := func(c *fiber.Ctx) error {
		return middlewares.IsUserInRoom(c, appServices.RoomService)
	}

	// initialize handlers
	handlers := initHandlers(appServices, logger)

	// healthcheck endpoint
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello world from JustJio API :)")
	})

	// OpenAPI docs
	setupDocsRoutes(router)

	v1 := router.Group("/v1")

	/* public routes */

	setupAuthRoutes(v1, handlers)

	/* private routes */

	setupUserRoutes(v1, handlers)
	setupRoomRoutes(v1, handlers, roomMiddleware)
	setupBillRoutes(v1, handlers)
	setupTransactionRoutes(v1, handlers)
	setupNotificationRoutes(v1, handlers)
	setupSubscriptionRoutes(v1, handlers)

	// 404 handler
	router.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "404: Endpoint Not Found",
		})
	})

	// Seed the database if in dev or staging environment
	if ctx.Config.Environment != "production" {
		if err := services.SeedDB(
			ctx.DB,
			appServices.UserService,
			appServices.RoomService,
			appServices.BillService,
			logger,
		); err != nil {
			logger.Fatal("Error seeding database: ", err)
		}
	}
}

func initRepositories(conn *gorm.DB) *Repositories {
	return &Repositories{
		UserRepository:         repositories.NewUserRepository(conn),
		RoomRepository:         repositories.NewRoomRepository(conn),
		MessageRepository:      repositories.NewMessageRepository(conn),
		BillRepository:         repositories.NewBillRepository(conn),
		TransactionRepository:  repositories.NewTransactionRepository(conn),
		SubscriptionRepository: repositories.NewSubscriptionRepository(conn),
		NotificationRepository: repositories.NewNotificationRepository(conn),
	}
}

func initServices(
	ctx *app.Context,
	repositories *Repositories,
) *Services {
	conn := ctx.DB
	logger := ctx.Logger

	userService := services.NewUserService(
		conn,
		repositories.UserRepository,
		logger,
	)
	authService := services.NewAuthService(
		userService,
		ctx.Kafka,
		utils.HashPassword,
		utils.SendSMTPEmail,
		ctx.Config,
		logger,
	)
	roomService := services.NewRoomService(
		conn,
		repositories.RoomRepository,
		repositories.UserRepository,
		http.NewHTTPClient(),
		ctx.Config.GoogleMapsApiKey,
		logger,
	)
	messageService := services.NewMessageService(
		conn,
		repositories.MessageRepository,
		repositories.RoomRepository,
		repositories.UserRepository,
		ctx.Kafka,
		logger,
	)
	transactionService := services.NewTransactionService(
		repositories.TransactionRepository,
		repositories.BillRepository,
		logger,
	)
	billService := services.NewBillService(
		conn,
		repositories.BillRepository,
		repositories.UserRepository,
		repositories.RoomRepository,
		repositories.TransactionRepository,
		transactionService,
		logger,
	)
	subscriptionService := services.NewSubscriptionService(
		repositories.SubscriptionRepository,
		ctx.NotificationsChan,
		logger,
	)
	notificationService := services.NewNotificationService(
		repositories.NotificationRepository,
		repositories.SubscriptionRepository,
		ctx.NotificationsChan,
		logger,
	)

	return &Services{
		UserService:         userService,
		AuthService:         authService,
		RoomService:         roomService,
		MessageService:      messageService,
		TransactionService:  transactionService,
		BillService:         billService,
		SubscriptionService: subscriptionService,
		NotificationService: notificationService,
	}
}

func initHandlers(services *Services, logger *logrus.Logger) *Handlers {
	return &Handlers{
		UserHandler:         handlers.NewUserHandler(services.UserService, logger),
		AuthHandler:         handlers.NewAuthHandler(services.AuthService, logger),
		RoomHandler:         handlers.NewRoomHandler(services.RoomService, logger),
		MessageHandler:      handlers.NewMessageHandler(services.MessageService, logger),
		BillHandler:         handlers.NewBillHandler(services.BillService, logger),
		TransactionHandler:  handlers.NewTransactionHandler(services.TransactionService, services.NotificationService, logger),
		NotificationHandler: handlers.NewNotificationHandler(services.NotificationService, logger),
		SubscriptionHandler: handlers.NewSubscriptionHandler(services.SubscriptionService, logger),
	}
}

func setupDocsRoutes(router *fiber.App) {
	router.Get("/swagger.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/yaml")
		return c.SendFile("./docs/swagger.yaml")
	})
	router.Get("/docs/*", swagger.New(swagger.Config{
		URL: "/swagger.yaml",
	}))
}

func setupAuthRoutes(v1 fiber.Router, handlers *Handlers) {
	auth := v1.Group("/auth")
	auth.Post("/",
		middlewares.ParseAndValidate[request.LoginRequest](),
		handlers.AuthHandler.Login,
	)
	auth.Post("/google",
		middlewares.ParseAndValidate[request.GoogleAuthRequest](),
		handlers.AuthHandler.GoogleLogin,
	)
	auth.Post("/signup",
		middlewares.ParseAndValidate[request.SignUpRequest](),
		handlers.AuthHandler.SignUp,
	)
	auth.Post("/verify",
		middlewares.ParseAndValidate[request.VerifyOTPRequest](),
		handlers.AuthHandler.VerifyOTP,
	)
	auth.Post("/otp",
		middlewares.ParseAndValidate[request.SendOTPEmailRequest](),
		handlers.AuthHandler.SendOTPEmail,
	)
	auth.Patch("/reset",
		middlewares.ParseAndValidate[request.ResetPasswordRequest](),
		handlers.AuthHandler.ResetPassword,
	)
}

func setupUserRoutes(v1 fiber.Router, handlers *Handlers) {
	users := v1.Group("/users")
	users.Get("/:userId", handlers.UserHandler.GetUser)
	users.Patch("/:userId/username",
		middlewares.ParseAndValidate[request.UpdateUsernameRequest](),
		handlers.UserHandler.UpdateUsername,
	)

	friends := users.Group("/:userId/friends")
	friends.Get("/", handlers.UserHandler.GetFriends)
	friends.Get("/count", handlers.UserHandler.CountFriends)
	friends.Get("/search", handlers.UserHandler.SearchNonFriends)
	friends.Delete("/:friendId", handlers.UserHandler.RemoveFriend)

	friendRequests := users.Group("/:userId/friendRequests")
	friendRequests.Get("/", handlers.UserHandler.GetFriendRequestsByStatus)
	friendRequests.Get("/count", handlers.UserHandler.CountPendingFriendRequests)
	friendRequests.Post("/",
		middlewares.ParseAndValidate[request.SendFriendRequest](),
		handlers.UserHandler.SendFriendRequest,
	)
	friendRequests.Patch("/",
		middlewares.ParseAndValidate[request.RespondToFriendRequestRequest](),
		handlers.UserHandler.RespondToFriendRequest,
	)

	userNotifications := users.Group("/:userId/notifications")
	userNotifications.Get("/", handlers.NotificationHandler.GetNotifications)
	userNotifications.Patch("/:id", handlers.NotificationHandler.MarkNotificationAsRead)
}

func setupRoomRoutes(
	v1 fiber.Router,
	handlers *Handlers,
	roommiddlewares func(c *fiber.Ctx) error,
) {
	rooms := v1.Group("/rooms")
	rooms.Get("/", handlers.RoomHandler.GetRooms)
	rooms.Get("/public", handlers.RoomHandler.GetUnjoinedPublicRooms)
	rooms.Get("/count", handlers.RoomHandler.GetNumRooms)
	rooms.Get("/invites", handlers.RoomHandler.GetRoomInvites)
	rooms.Get("/invites/count", handlers.RoomHandler.GetNumRoomInvites)
	rooms.Get("/venues/search", handlers.RoomHandler.QueryVenue)
	rooms.Get("/:roomId", roommiddlewares, handlers.RoomHandler.GetRoom)
	rooms.Get("/:roomId/uninvited", roommiddlewares, handlers.RoomHandler.GetUninvitedFriendsForRoom)
	rooms.Post("/",
		middlewares.ParseAndValidate[request.CreateRoomRequest](),
		handlers.RoomHandler.CreateRoom,
	)
	rooms.Post("/:roomId",
		roommiddlewares,
		middlewares.ParseAndValidate[request.InviteUserRequest](),
		handlers.RoomHandler.InviteUser,
	)
	rooms.Patch("/:roomId",
		middlewares.ParseAndValidate[request.RespondToRoomInviteRequest](),
		handlers.RoomHandler.RespondToRoomInvite,
	)
	rooms.Patch("/:roomId/edit",
		middlewares.ParseAndValidate[request.EditRoomRequest](),
		handlers.RoomHandler.EditRoom,
	)
	rooms.Patch("/:roomId/join", handlers.RoomHandler.JoinRoom)
	rooms.Patch("/:roomId/close", roommiddlewares, handlers.RoomHandler.CloseRoom)
	rooms.Delete("/:roomId/leave", roommiddlewares, handlers.RoomHandler.LeaveRoom)

	messages := rooms.Group("/:roomId/messages")
	messages.Use(roommiddlewares)
	messages.Get("/", handlers.MessageHandler.GetMessages)
	messages.Get("/:msgId", handlers.MessageHandler.GetMessage)
	messages.Post("/",
		middlewares.ParseAndValidate[request.CreateMessageRequest](),
		handlers.MessageHandler.CreateMessage,
	)
}

func setupBillRoutes(v1 fiber.Router, handlers *Handlers) {
	bills := v1.Group("/bills")
	bills.Get("/", handlers.BillHandler.GetBillsByRoom)
	bills.Post("/",
		middlewares.ParseAndValidate[request.CreateBillRequest](),
		handlers.BillHandler.CreateBill,
	)
	bills.Post("/consolidate",
		middlewares.ParseAndValidate[request.ConsolidateBillsRequest](),
		handlers.BillHandler.ConsolidateBills,
	)
}

func setupTransactionRoutes(v1 fiber.Router, handlers *Handlers) {
	transactions := v1.Group("/transactions")
	transactions.Get("/", handlers.TransactionHandler.GetTransactionsByUser)
	transactions.Patch("/:txId/settle", handlers.TransactionHandler.SettleTransaction)
}

func setupNotificationRoutes(v1 fiber.Router, handlers *Handlers) {
	notifications := v1.Group("/notifications")
	notifications.Get("/:id", handlers.NotificationHandler.GetNotification)
	notifications.Post("/",
		middlewares.ParseAndValidate[request.CreateNotificationRequest](),
		handlers.NotificationHandler.CreateNotification,
	)
}

func setupSubscriptionRoutes(v1 fiber.Router, handlers *Handlers) {
	subscriptions := v1.Group("/subscriptions")
	subscriptions.Get("/:endpoint", handlers.SubscriptionHandler.GetSubscriptionByEndpoint)
	subscriptions.Post("/",
		middlewares.ParseAndValidate[request.CreateSubscriptionRequest](),
		handlers.SubscriptionHandler.CreateSubscription,
	)
	subscriptions.Delete("/:subId", handlers.SubscriptionHandler.DeleteSubscription)
}
