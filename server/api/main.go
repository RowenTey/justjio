package main

import (
	"os"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/router"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/RowenTey/JustJio/server/api/worker"

	"github.com/gofiber/fiber/v2"
)

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

	db := database.ConnectDB(conf, logger)
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

	logger.Info("Server running on port ", conf.Port)
	logger.Fatal(app.Listen(":" + conf.Port))
}
