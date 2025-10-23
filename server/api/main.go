package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/router"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/RowenTey/JustJio/server/api/worker"

	"github.com/gofiber/fiber/v2"

	// Swagger docs
	_ "github.com/RowenTey/JustJio/server/api/docs"
)

// @title JustJio API
// @version 1.0
// @description API server for JustJio.
// @host localhost:8080
// @schemes http
// @BasePath /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	logger := utils.InitLogger(env)

	conf, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration!")
	}

	logger.Info("Starting API server...")

	tp, err := utils.InitTracer(env)
	if err != nil {
		logger.Warn("Failed to initialize tracer: ", err)
	}
	logger.Info("OpenTelemetry tracer initialized")
	defer func() {
		if err := utils.ShutdownTracer(context.Background(), tp); err != nil {
			logger.Error("Failed to shutdown tracer: ", err)
		}
	}()

	db := database.ConnectDB(conf, env, logger)
	notificationsChan := worker.StartWorkers(conf, logger, db)

	kafkaService, err := services.NewKafkaService(
		conf,
		logger,
		env,
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer kafkaService.Close()

	app := fiber.New()
	middleware.Fiber(app, conf, env)
	router.Initalize(
		app,
		env,
		conf,
		logger,
		db,
		kafkaService,
		notificationsChan,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Server running on port ", conf.Port)
		if err := app.Listen(":" + conf.Port); err != nil {
			logger.Fatal(err)
		}
	}()

	<-quit
	logger.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		logger.Error("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited gracefully")
}
