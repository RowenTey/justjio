package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/RowenTey/JustJio/server/ws/services"
	"github.com/RowenTey/JustJio/server/ws/utils"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
)

func HandleWebsocketConn(
	c *websocket.Conn,
	conf *utils.Config,
	logger *logrus.Logger,
	connMap *utils.ConnMap,
	kafkaClientMap map[string]*services.UserKafkaClient,
	env string,
) {
	wsLogger := logger.WithField("service", "WebSocket")

	user, err := utils.GetCurrentUser(conf, c)
	if err != nil {
		handleAuthError(c, logger, err)
		return
	}

	wsLogger.Infof("User %s connected\n", user.ID)
	forAllConns, onRemove, isInit := connMap.Add(user.ID, c)

	writeMessageFn := func(message kafka.Message) func(conn *websocket.Conn) {
		return func(conn *websocket.Conn) {
			wsLogger.Info("Sending message to ", user.ID)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message.Value)); err != nil {
				wsLogger.Error("WebSocket Error:", err)
			}
			wsLogger.Info("Sent message to ", user.ID)
		}
	}

	onMessage := func(message kafka.Message) {
		forAllConns(writeMessageFn(message))
	}

	kafkaClient := getKafkaClient(isInit, conf, logger, env, user, kafkaClientMap, onMessage)
	setupConnectionHandler(c, user, onRemove, logger, kafkaClient, kafkaClientMap)
	setupHeartbeat(c, logger)
	processWsMessage(c, logger, onMessage)
}

func processWsMessage(c *websocket.Conn, logger *logrus.Logger, onMessage func(message kafka.Message)) {
	wsLogger := logger.WithField("service", "WebSocket")

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
}

func setupHeartbeat(c *websocket.Conn, logger *logrus.Logger) {
	wsLogger := logger.WithField("service", "WebSocket")

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
}

func setupConnectionHandler(
	c *websocket.Conn,
	user *utils.User,
	onRemove func(func()) bool,
	logger *logrus.Logger,
	kafkaClient *services.KafkaService,
	kafkaClientMap map[string]*services.UserKafkaClient,
) {
	wsLogger := logger.WithField("service", "WebSocket")
	kafkaLogger := logger.WithField("service", "Kafka")

	onClose := func(code int, text string) error {
		wsLogger.Infof("User %s disconnected\n", user.ID)
		// only runs when the last connection is removed
		onRemove(func() {
			if err := kafkaClient.Unsubscribe(); err != nil {
				kafkaLogger.Error(err)
			}
			kafkaClient.Close()
			delete(kafkaClientMap, user.ID)
		})
		return nil
	}
	c.SetCloseHandler(onClose)

	// set up ping/pong handlers
	c.SetPingHandler(func(appData string) error {
		wsLogger.Debug("Received ping: ", appData)
		return c.WriteMessage(websocket.PongMessage, []byte(appData))
	})
	c.SetPongHandler(func(appData string) error {
		return nil
	})
}

func getKafkaClient(
	isInit bool,
	conf *utils.Config,
	logger *logrus.Logger,
	env string,
	user *utils.User,
	kafkaClientMap map[string]*services.UserKafkaClient,
	onMessage func(message kafka.Message),
) *services.KafkaService {
	if !isInit {
		return kafkaClientMap[user.ID].Client
	}

	kafkaLogger := logger.WithField("service", "Kafka")

	consumerName := "chat-service"
	if env == "dev" || env == "staging" {
		consumerName = fmt.Sprintf("chat-service-%s", env)
	}
	consumerName = fmt.Sprintf("%s-%s", conf.Kafka.TopicPrefix, consumerName)
	kafkaLogger.Infof("Consumer name: %s", consumerName)

	// if the user is connecting for the first time, create a new Kafka client
	kafkaClient, err := services.NewKafkaService(conf, consumerName, env)
	if err != nil {
		kafkaLogger.Fatal(err)
	}

	channel := kafkaClient.GetUserChannel(user.ID)
	kafkaLogger.Info("Channel: ", channel)

	kafkaClientMap[user.ID] = &services.UserKafkaClient{
		Client: kafkaClient,
	}

	if err := kafkaClient.Subscribe([]string{channel}); err != nil {
		kafkaLogger.Error("Error subscribing to channel: ", err)
	}
	// consume messages in a separate goroutine
	go kafkaClient.ConsumeMessages(onMessage)

	return kafkaClient
}

func handleAuthError(c *websocket.Conn, logger *logrus.Logger, err error) {
	logger.WithField("service", "AUTH").Error(err)

	if err := c.WriteJSON(fiber.Map{
		"status": "Unauthorized",
		"error":  err.Error(),
	}); err != nil {
		logger.WithField("service", "WebSocket").Error("Error writing JSON:", err)
	}

	c.Close()
}
