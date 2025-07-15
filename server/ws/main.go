package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/RowenTey/JustJio/server/ws/services"
	"github.com/RowenTey/JustJio/server/ws/utils"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	logger := utils.InitLogger(env)
	kafkaLogger := logger.WithField("service", "Kafka")
	wsLogger := logger.WithField("service", "WebSocket")

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
	connMap := utils.NewConnMap()

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
		user, err := utils.GetCurrentUser(conf, c)
		if err != nil {
			logger.WithField("service", "AUTH").Error(err)

			if err := c.WriteJSON(fiber.Map{
				"status": "Unauthorized",
				"error":  err.Error(),
			}); err != nil {
				wsLogger.Error("Error writing JSON:", err)
			}

			c.Close()
			return
		}

		forAllConns, remove, isInit := connMap.Add(user.ID, c)

		var channel string
		var kafkaClient *services.KafkaService

		onMessage := func(message kafka.Message) {
			forAllConns(func(conn *websocket.Conn) {
				wsLogger.Info("Sending message to ", user.ID)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message.Value)); err != nil {
					wsLogger.Error("WebSocket Error:", err)
				}
				wsLogger.Info("Sent message to ", user.ID)
			})
		}

		// if the user is connecting for the first time, create a new Kafka client
		if isInit {
			kafkaClient, err = services.NewKafkaService(conf, consumerName, env)
			if err != nil {
				kafkaLogger.Fatal(err)
			}

			channel = kafkaClient.GetUserChannel(user.ID)
			kafkaLogger.Info("Channel: ", channel)

			userKafkaClients[user.ID] = &services.UserKafkaClient{
				Client: kafkaClient,
			}

			if err := kafkaClient.Subscribe([]string{channel}); err != nil {
				kafkaLogger.Error("Error subscribing to channel: ", err)
			}
			// consume messages in a separate goroutine
			go kafkaClient.ConsumeMessages(onMessage)
		} else {
			kafkaClient = userKafkaClients[user.ID].Client
		}

		onClose := func(code int, text string) error {
			wsLogger.Infof("User %s disconnected\n", user.ID)
			// only runs when the last connection is removed
			remove(func() {
				if err := kafkaClient.Unsubscribe(); err != nil {
					kafkaLogger.Error(err)
				}
				kafkaClient.Close()
				delete(userKafkaClients, user.ID)
			})
			return nil
		}
		c.SetCloseHandler(onClose)

		wsLogger.Infof("User %s connected\n", user.ID)

		// set up ping/pong handlers
		c.SetPingHandler(func(appData string) error {
			wsLogger.Debug("Received ping: ", appData)
			return c.WriteMessage(websocket.PongMessage, []byte(appData))
		})
		c.SetPongHandler(func(appData string) error {
			return nil
		})

		// send ping messages every 5 seconds (heartbeat) via a goroutine
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			heartbeat := time.NewTicker(5 * time.Second)
			defer heartbeat.Stop()
			for {
				select {
				case <-heartbeat.C:
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						wsLogger.Error("Ping error: ", err)
						c.Close()
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		var (
			mt    int
			msg   []byte
			wsErr error
		)
		for {
			if mt, msg, wsErr = c.ReadMessage(); wsErr != nil {
				if websocket.IsCloseError(wsErr, websocket.CloseNoStatusReceived) {
					wsLogger.Debug("Connection closed by client")
				} else {
					wsLogger.Error("Error: ", wsErr)
				}
				break
			}

			wsLogger.Infof("Received (%d): %s\n", mt, msg)
			onMessage(kafka.Message{
				Value: msg,
			})
		}
	}))

	logger.Info("Server running on port ", conf.Port)
	logger.Fatal(app.Listen(":" + conf.Port))
}
