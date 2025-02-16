package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RowenTey/JustJio/server-ws/services"
	"github.com/RowenTey/JustJio/server-ws/utils"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

type UserKafkaClient struct {
	client *services.KafkaService
}

// maps user ID to Kafka client
var userKafkaClients = make(map[string]*UserKafkaClient)

func webSocketUpgrade(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("websocket", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func main() {
	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	// only load .env file if in dev environment
	if env == "dev" {
		godotenv.Load(".env")
	}

	app := fiber.New()
	connMap := utils.NewConnMap()

	consumerName := "chat-service"
	if env == "dev" || env == "staging" {
		consumerName = fmt.Sprintf("chat-service-%s", env)
	}
	consumerName = fmt.Sprintf("%s-%s", utils.Config("KAFKA_TOPIC_PREFIX"), consumerName)

	// handle websocket upgrade
	app.Use(webSocketUpgrade)

	app.Get("/", websocket.New(func(c *websocket.Conn) {
		user, err := services.GetCurrentUser(c)
		if err != nil {
			log.Println("[AUTH] ", err)
			c.WriteJSON(fiber.Map{
				"status": "Unauthorized",
			})
			c.Close()
			return
		}

		forAllConns, remove, isInit := connMap.Add(user.ID, c)

		var kafkaClient *services.KafkaService
		var channel string

		onMessage := func(message kafka.Message) {
			forAllConns(func(conn *websocket.Conn) {
				log.Println("[WebSocket] Sending message to", user.ID)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message.Value)); err != nil {
					log.Println("[WebSocket] WebSocket Error:", err)
				}
				log.Println("[WebSocket] Sent message to", user.ID)
			})
		}

		// if the user is connecting for the first time, create a new Kafka client
		if isInit {
			kafkaClient, err = services.NewKafkaService(utils.Config("KAFKA_URL"), consumerName)
			if err != nil {
				log.Fatal(err)
			}

			channel = services.GetUserChannel(user.ID, env)
			log.Println("[Kafka] Channel: ", channel)

			userKafkaClients[user.ID] = &UserKafkaClient{
				client: kafkaClient,
			}

			kafkaClient.Subscribe([]string{channel})
			go kafkaClient.ConsumeMessages(onMessage)
		} else {
			kafkaClient = userKafkaClients[user.ID].client
		}

		onClose := func(code int, text string) error {
			log.Printf("[WebSocket] User %s disconnected\n", user.ID)
			// only runs when the last connection is removed
			remove(func() {
				if err := kafkaClient.Unsubscribe(); err != nil {
					log.Println(err)
				}
				kafkaClient.Close()
				delete(userKafkaClients, user.ID)
			})
			return nil
		}
		c.SetCloseHandler(onClose)

		log.Printf("[WebSocket] User %s connected\n", user.ID)

		// Set up ping/pong handlers
		c.SetPingHandler(func(appData string) error {
			log.Println("[WebSocket] Received ping:", appData)
			return c.WriteMessage(websocket.PongMessage, []byte(appData))
		})

		c.SetPongHandler(func(appData string) error {
			return nil
		})

		// Send ping messages every 5 seconds (heartbeat)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			heartbeat := time.NewTicker(5 * time.Second)
			defer heartbeat.Stop()
			for {
				select {
				case <-heartbeat.C:
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Println("[WebSocket] Ping error:", err)
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
					log.Println("[WebSocket] Connection closed by client")
				} else {
					log.Println("[WebSocket] Error:", wsErr)
				}
				break
			}

			log.Printf("[WebSocket] Received (%d): %s\n", mt, msg)
			onMessage(kafka.Message{
				Value: msg,
			})
		}
	}))

	log.Println("Server running on port", utils.Config("PORT"))
	log.Fatal(app.Listen(":" + utils.Config("PORT")))
}
