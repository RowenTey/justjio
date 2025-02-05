package main

import (
	"log"
	"os"

	"github.com/RowenTey/JustJio/config"
	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/middleware"
	"github.com/RowenTey/JustJio/router"
	"github.com/RowenTey/JustJio/services"

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
		godotenv.Load(".env")
	}

	app := fiber.New()

	database.ConnectDB()
	if env == "dev" || env == "staging" {
		services.SeedDB(database.DB)
	}

	kafkaService, err := services.NewKafkaService(config.Config("KAFKA_URL"), env)
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaService.Close()

	middleware.Fiber(app)
	router.Initalize(app, kafkaService)
	log.Fatal(app.Listen(":" + config.Config("PORT")))
}
