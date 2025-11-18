package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/router"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/RowenTey/JustJio/server/api/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	// initialize logger
	utils.InitLogger(env)

	// only load .env file if in dev environment
	if env == "dev" {
		log.Debug("Loading .env file...")
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	log.Info("Starting API server...")

	notificationsChan := worker.RunPushNotification()

	app := fiber.New()

	database.ConnectDB()
	if env == "dev" || env == "staging" {
		if err := services.SeedDB(database.DB); err != nil {
			log.Fatal("Error seeding database:", err)
		}
	}

	kafkaService, err := services.NewKafkaService(config.Config("KAFKA_URL"), env)
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaService.Close()

	middleware.Fiber(app, env, config.Config("ALLOWED_ORIGINS"))
	router.Initalize(app, kafkaService, notificationsChan)

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Info("Server running on port ", config.Config("PORT"))
		if err := app.Listen(":" + config.Config("PORT")); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Info("Gracefully shutting down server...")

	// Close Kafka service
	kafkaService.Close()

	// Shutdown server with timeout
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Error("Server forced to shutdown:", err)
	}

	log.Info("Server exited")
}
