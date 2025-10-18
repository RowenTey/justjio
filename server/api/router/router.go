package router

import (
	"github.com/RowenTey/JustJio/server/api/config"
	pushNotificationModel "github.com/RowenTey/JustJio/server/api/dto/push_notifications"
	"github.com/RowenTey/JustJio/server/api/dto/request"
	"github.com/RowenTey/JustJio/server/api/handlers"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type Repositories struct {
	UserRepository         repository.UserRepository
	RoomRepository         repository.RoomRepository
	MessageRepository      repository.MessageRepository
	BillRepository         repository.BillRepository
	TransactionRepository  repository.TransactionRepository
	SubscriptionRepository repository.SubscriptionRepository
	NotificationRepository repository.NotificationRepository
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

func Initalize(
	router *fiber.App,
	env string,
	conf *config.Config,
	logger *logrus.Logger,
	dbConn *gorm.DB,
	kafkaService services.KafkaService,
	notificationsChan chan<- pushNotificationModel.NotificationData,
) {
	// initialize repositories
	repositories := initRepositories(dbConn)

	// initialize services
	appServices := initServices(
		dbConn,
		conf,
		repositories,
		kafkaService,
		notificationsChan,
		logger,
	)

	// custom middlewares
	roomMiddleware := func(c *fiber.Ctx) error {
		return middleware.IsUserInRoom(c, appServices.RoomService)
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
	if env == "dev" || env == "staging" {
		if err := services.SeedDB(
			dbConn,
			appServices.UserService,
			appServices.RoomService,
			appServices.BillService,
			logger,
		); err != nil {
			logger.Fatal("Error seeding database: ", err)
		}
	}
}

func initRepositories(dbConn *gorm.DB) *Repositories {
	return &Repositories{
		UserRepository:         repository.NewUserRepository(dbConn),
		RoomRepository:         repository.NewRoomRepository(dbConn),
		MessageRepository:      repository.NewMessageRepository(dbConn),
		BillRepository:         repository.NewBillRepository(dbConn),
		TransactionRepository:  repository.NewTransactionRepository(dbConn),
		SubscriptionRepository: repository.NewSubscriptionRepository(dbConn),
		NotificationRepository: repository.NewNotificationRepository(dbConn),
	}
}

func initServices(
	dbConn *gorm.DB,
	conf *config.Config,
	repositories *Repositories,
	kafkaService services.KafkaService,
	notificationsChan chan<- pushNotificationModel.NotificationData,
	logger *logrus.Logger,
) *Services {
	userService := services.NewUserService(
		dbConn,
		repositories.UserRepository,
		logger,
	)
	authService := services.NewAuthService(
		userService,
		kafkaService,
		utils.HashPassword,
		utils.SendSMTPEmail,
		conf,
		logger,
	)
	roomService := services.NewRoomService(
		dbConn,
		repositories.RoomRepository,
		repositories.UserRepository,
		utils.NewHTTPClient(), // Instrumented HTTP client with OpenTelemetry
		conf.GoogleMapsApiKey,
		logger,
	)
	messageService := services.NewMessageService(
		dbConn,
		repositories.MessageRepository,
		repositories.RoomRepository,
		repositories.UserRepository,
		kafkaService,
		logger,
	)
	transactionService := services.NewTransactionService(
		repositories.TransactionRepository,
		repositories.BillRepository,
		logger,
	)
	billService := services.NewBillService(
		dbConn,
		repositories.BillRepository,
		repositories.UserRepository,
		repositories.RoomRepository,
		repositories.TransactionRepository,
		transactionService,
		logger,
	)
	subscriptionService := services.NewSubscriptionService(
		repositories.SubscriptionRepository,
		notificationsChan,
		logger,
	)
	notificationService := services.NewNotificationService(
		repositories.NotificationRepository,
		repositories.SubscriptionRepository,
		notificationsChan,
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
		middleware.ParseAndValidate[request.LoginRequest](),
		handlers.AuthHandler.Login,
	)
	auth.Post("/google",
		middleware.ParseAndValidate[request.GoogleAuthRequest](),
		handlers.AuthHandler.GoogleLogin,
	)
	auth.Post("/signup",
		middleware.ParseAndValidate[request.SignUpRequest](),
		handlers.AuthHandler.SignUp,
	)
	auth.Post("/verify",
		middleware.ParseAndValidate[request.VerifyOTPRequest](),
		handlers.AuthHandler.VerifyOTP,
	)
	auth.Post("/otp",
		middleware.ParseAndValidate[request.SendOTPEmailRequest](),
		handlers.AuthHandler.SendOTPEmail,
	)
	auth.Patch("/reset",
		middleware.ParseAndValidate[request.ResetPasswordRequest](),
		handlers.AuthHandler.ResetPassword,
	)
}

func setupUserRoutes(v1 fiber.Router, handlers *Handlers) {
	users := v1.Group("/users")
	users.Get("/:userId", handlers.UserHandler.GetUser)
	users.Patch("/:userId/username",
		middleware.ParseAndValidate[request.UpdateUsernameRequest](),
		handlers.UserHandler.UpdateUsername,
	)

	friends := users.Group("/:userId/friends")
	friends.Get("/", handlers.UserHandler.GetFriends)
	friends.Get("/count", handlers.UserHandler.GetNumFriends)
	friends.Get("/search", handlers.UserHandler.SearchNonFriends)
	friends.Delete("/:friendId", handlers.UserHandler.RemoveFriend)

	friendRequests := users.Group("/:userId/friendRequests")
	friendRequests.Get("/", handlers.UserHandler.GetFriendRequestsByStatus)
	friendRequests.Get("/count", handlers.UserHandler.CountPendingFriendRequests)
	friendRequests.Post("/",
		middleware.ParseAndValidate[request.ModifyFriendRequest](),
		handlers.UserHandler.SendFriendRequest,
	)
	friendRequests.Patch("/",
		middleware.ParseAndValidate[request.RespondToFriendRequestRequest](),
		handlers.UserHandler.RespondToFriendRequest,
	)

	userNotifications := users.Group("/:userId/notifications")
	userNotifications.Get("/", handlers.NotificationHandler.GetNotifications)
	userNotifications.Patch("/:id", handlers.NotificationHandler.MarkNotificationAsRead)
}

func setupRoomRoutes(
	v1 fiber.Router,
	handlers *Handlers,
	roomMiddleware func(c *fiber.Ctx) error,
) {
	rooms := v1.Group("/rooms")
	rooms.Get("/", handlers.RoomHandler.GetRooms)
	rooms.Get("/public", handlers.RoomHandler.GetUnjoinedPublicRooms)
	rooms.Get("/count", handlers.RoomHandler.GetNumRooms)
	rooms.Get("/invites", handlers.RoomHandler.GetRoomInvites)
	rooms.Get("/invites/count", handlers.RoomHandler.GetNumRoomInvites)
	rooms.Get("/venues/search", handlers.RoomHandler.QueryVenue)
	rooms.Get("/:roomId", roomMiddleware, handlers.RoomHandler.GetRoom)
	rooms.Get("/:roomId/uninvited", roomMiddleware, handlers.RoomHandler.GetUninvitedFriendsForRoom)
	rooms.Post("/",
		middleware.ParseAndValidate[request.CreateRoomRequest](),
		handlers.RoomHandler.CreateRoom,
	)
	rooms.Post("/:roomId",
		roomMiddleware,
		middleware.ParseAndValidate[request.InviteUserRequest](),
		handlers.RoomHandler.InviteUser,
	)
	rooms.Patch("/:roomId",
		middleware.ParseAndValidate[request.RespondToRoomInviteRequest](),
		handlers.RoomHandler.RespondToRoomInvite,
	)
	rooms.Patch("/:roomId/edit",
		middleware.ParseAndValidate[request.EditRoomRequest](),
		handlers.RoomHandler.EditRoom,
	)
	rooms.Patch("/:roomId/join", handlers.RoomHandler.JoinRoom)
	rooms.Patch("/:roomId/close", roomMiddleware, handlers.RoomHandler.CloseRoom)
	rooms.Delete("/:roomId/leave", roomMiddleware, handlers.RoomHandler.LeaveRoom)

	messages := rooms.Group("/:roomId/messages")
	messages.Use(roomMiddleware)
	messages.Get("/", handlers.MessageHandler.GetMessages)
	messages.Get("/:msgId", handlers.MessageHandler.GetMessage)
	messages.Post("/",
		middleware.ParseAndValidate[request.CreateMessageRequest](),
		handlers.MessageHandler.CreateMessage,
	)
}

func setupBillRoutes(v1 fiber.Router, handlers *Handlers) {
	bills := v1.Group("/bills")
	bills.Get("/", handlers.BillHandler.GetBillsByRoom)
	bills.Post("/",
		middleware.ParseAndValidate[request.CreateBillRequest](),
		handlers.BillHandler.CreateBill,
	)
	bills.Post("/consolidate",
		middleware.ParseAndValidate[request.ConsolidateBillsRequest](),
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
		middleware.ParseAndValidate[request.CreateNotificationRequest](),
		handlers.NotificationHandler.CreateNotification,
	)
}

func setupSubscriptionRoutes(v1 fiber.Router, handlers *Handlers) {
	subscriptions := v1.Group("/subscriptions")
	subscriptions.Get("/:endpoint", handlers.SubscriptionHandler.GetSubscriptionByEndpoint)
	subscriptions.Post("/",
		middleware.ParseAndValidate[request.CreateSubscriptionRequest](),
		handlers.SubscriptionHandler.CreateSubscription,
	)
	subscriptions.Delete("/:subId", handlers.SubscriptionHandler.DeleteSubscription)
}
