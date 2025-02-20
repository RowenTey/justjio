package main

import (
	"log"
	"os"

	"github.com/RowenTey/JustJio/config"
	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/middleware"
	"github.com/RowenTey/JustJio/router"
	"github.com/RowenTey/JustJio/services"
	"github.com/RowenTey/JustJio/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting server...")

	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	// only load .env file if in dev environment
	if env == "dev" {
		log.Println("Loading .env file...")
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

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

	middleware.Fiber(app)
	router.Initalize(app, kafkaService, notificationsChan)

	log.Println("Server running on port", config.Config("PORT"))
	log.Fatal(app.Listen(":" + config.Config("PORT")))
}
