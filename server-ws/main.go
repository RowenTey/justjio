package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RowenTey/JustJio/server-ws/services"
	"github.com/RowenTey/JustJio/server-ws/utils"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

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

	log.Println("Kafka client created")
	kafkaClient, err := services.NewKafkaService(utils.Config("KAFKA_URL"), consumerName)
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaClient.Close()

	// handle websocket upgrade
	app.Use(webSocketUpgrade)

	app.Get("/", websocket.New(func(c *websocket.Conn) {
		user, err := services.GetCurrentUser(c)
		if err != nil {
			log.Println(err)
			c.WriteJSON(fiber.Map{
				"status": "Unauthorized",
			})
			c.Close()
			return
		}
		channel := services.GetUserChannel(user.ID, env)

		log.Println("Channel: ", channel)
		forAllConns, remove, isInit := connMap.Add(user.ID, c)

		onMessage := func(message kafka.Message) {
			forAllConns(func(conn *websocket.Conn) {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message.Value)); err != nil {
					log.Println("WebSocket Error:", err)
				}
			})
		}

		onClose := func(code int, text string) error {
			log.Printf("%s disconnected\n", user.ID)
			remove(func() {
				if err := kafkaClient.Unsubscribe(); err != nil {
					log.Println(err)
				}
			})
			return nil
		}
		c.SetCloseHandler(onClose)

		if isInit {
			kafkaClient.Subscribe([]string{channel})
			go kafkaClient.ConsumeMessages(onMessage)
		}

		log.Printf("User %s connected\n", user.ID)

		var (
			mt    int
			msg   []byte
			wsErr error
		)
		for {
			if mt, msg, wsErr = c.ReadMessage(); wsErr != nil {
				log.Println("WebSocket Error:", wsErr)
				break
			}

			log.Printf("[WebSocket] Received (%d): %s\n", mt, msg)
			onMessage(kafka.Message{
				Value: msg,
			})
		}
	}))

	log.Fatal(app.Listen(":" + utils.Config("PORT")))
}
