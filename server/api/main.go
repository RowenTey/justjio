package main

import (
	"os"

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

	log.Info("Server running on port ", config.Config("PORT"))
	log.Fatal(app.Listen(":" + config.Config("PORT")))
}
