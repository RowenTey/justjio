package main

import (
	"fmt"
	"os"

	"github.com/RowenTey/JustJio/server/ws/handlers"
	"github.com/RowenTey/JustJio/server/ws/services"
	"github.com/RowenTey/JustJio/server/ws/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	logger := utils.InitLogger(env)

	conf, err := utils.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration!")
	}

	consumerName := "chat-service"
	if env == "dev" || env == "staging" {
		consumerName = fmt.Sprintf("chat-service-%s", env)
	}
	consumerName = fmt.Sprintf("%s-%s", conf.Kafka.TopicPrefix, consumerName)
	logger.Infof("Consumer name: %s", consumerName)

	logger.Info("Starting WS server...")
	app := fiber.New()
	connMap := utils.NewConnMap(logger)

	// maps user ID to Kafka client
	var userKafkaClients = make(map[string]*services.UserKafkaClient)

	allowedOriginsMiddleware := func(c *fiber.Ctx) error {
		return utils.IsAllowedOrigins(conf, logger, c)
	}

	// healthcheck endpoint
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("pong")
	})

	// websocket endpoint with middleware to handle websocket upgrade
	app.Get("/", utils.WebSocketUpgrade, allowedOriginsMiddleware, websocket.New(func(c *websocket.Conn) {
		handlers.HandleWebsocketConn(conf, logger, connMap, userKafkaClients, env, consumerName, c)
	}))

	logger.Info("Server running on port ", conf.Port)
	logger.Fatal(app.Listen(":" + conf.Port))
}
