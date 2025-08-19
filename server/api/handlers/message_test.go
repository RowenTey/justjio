package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MessageHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	mockJWTSecret string

	messageService *services.MessageService
	kafkaService   services.KafkaService

	testUser1ID    uint
	testUser2ID    uint
	testRoomID     string
	testRoom       *model.Room
	testUser1Token string
	testUser2Token string
}

func (suite *MessageHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	// Get Kafka broker address
	kafkaBrokers, err := suite.dependencies.KafkaContainer.Brokers(suite.ctx)
	assert.NoError(suite.T(), err)

	// Initialize Kafka service
	config := &config.Config{
		Kafka: config.KafkaConfig{
			Host:        strings.Split(kafkaBrokers[0], ":")[0],
			Port:        strings.Split(kafkaBrokers[0], ":")[1],
			TopicPrefix: "test-",
		},
	}
	suite.kafkaService, err = services.NewKafkaService(config, suite.logger, "test")
	assert.NoError(suite.T(), err)

	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	mockHttpClient := new(utils.MockHTTPClient)
	roomRepository := repository.NewRoomRepository(suite.db)
	userRepository := repository.NewUserRepository(suite.db)
	billRepository := repository.NewBillRepository(suite.db)
	roomService := services.NewRoomService(
		suite.db,
		roomRepository,
		userRepository,
		billRepository,
		mockHttpClient,
		"test-api-key",
		suite.logger,
	)
	messageRepo := repository.NewMessageRepository(suite.db)
	roomRepo := repository.NewRoomRepository(suite.db)
	userRepo := repository.NewUserRepository(suite.db)
	suite.messageService = services.NewMessageService(
		suite.db,
		messageRepo,
		roomRepo,
		userRepo,
		suite.kafkaService,
		suite.logger,
	)
	messageHandler := NewMessageHandler(suite.messageService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(suite.mockJWTSecret))

	// Register Message routes
	roomMiddleware := func(c *fiber.Ctx) error {
		return middleware.IsUserInRoom(c, roomService)
	}
	messageRoutes := suite.app.Group("/rooms/:roomId/messages")
	messageRoutes.Use(roomMiddleware)
	messageRoutes.Get("/:msgId", messageHandler.GetMessage)
	messageRoutes.Get("/", messageHandler.GetMessages)
	messageRoutes.Post("/", messageHandler.CreateMessage)
}

func (suite *MessageHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	suite.logger.Info("Tore down test suite dependencies")
}

func (suite *MessageHandlerTestSuite) SetupTest() {
	// Create User 1
	hashedPassword1, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)
	user1 := model.User{
		Username: "user1",
		Email:    "user1@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&user1)
	assert.NoError(suite.T(), result.Error)
	suite.testUser1ID = user1.ID
	token1, err := tests.GenerateTestToken(user1.ID, user1.Username, user1.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testUser1Token = token1

	// Create User 2
	hashedPassword2, err := utils.HashPassword("password456")
	assert.NoError(suite.T(), err)
	user2 := model.User{
		Username: "user2",
		Email:    "user2@example.com",
		Password: hashedPassword2,
	}
	result = suite.db.Create(&user2)
	assert.NoError(suite.T(), result.Error)
	suite.testUser2ID = user2.ID
	token2, err := tests.GenerateTestToken(user2.ID, user2.Username, user2.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testUser2Token = token2

	// Create Room and add users
	room := model.Room{
		Name:   "Test Room",
		HostID: suite.testUser1ID,
		Users:  []model.User{user1, user2},
	}
	result = suite.db.Create(&room)
	assert.NoError(suite.T(), result.Error)
	suite.testRoomID = room.ID
	suite.testRoom = &room

	log.Infof("SetupTest complete: User1 ID=%d, User2 ID=%d, Room ID=%s", suite.testUser1ID, suite.testUser2ID, suite.testRoomID)
}

func (suite *MessageHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE messages RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	suite.logger.Info("Tore down test data")
}

func TestMessageHandlerSuite(t *testing.T) {
	suite.Run(t, new(MessageHandlerTestSuite))
}

func (suite *MessageHandlerTestSuite) TestGetMessage_Success() {
	// Create a message
	suite.logger.Info("Creating test message for room id: ", suite.testRoomID)
	message := model.Message{
		// Room: *suite.testRoom,
		RoomID:   suite.testRoomID,
		SenderID: suite.testUser1ID,
		Content:  "Hello World!",
	}
	log.Infof("Message to be created: %+v\n", message)
	result := suite.db.Create(&message)
	assert.NoError(suite.T(), result.Error)

	var messages []model.Message
	result = suite.db.Find(&messages)
	assert.NoError(suite.T(), result.Error)
	assert.Len(suite.T(), messages, 1)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rooms/%s/messages/%d", suite.testRoomID, message.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved message successfully", responseBody["message"])
	assert.NotNil(suite.T(), responseBody["data"])
	messageData := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), message.Content, messageData["content"])
	assert.Equal(suite.T(), message.SenderID, uint(messageData["senderId"].(float64)))
}

func (suite *MessageHandlerTestSuite) TestGetMessage_NotFound() {
	nonExistentMsgID := 9999
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rooms/%s/messages/%d", suite.testRoomID, nonExistentMsgID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "No message found", responseBody["message"])
}

func (suite *MessageHandlerTestSuite) TestGetMessages_Success_Ascending() {
	// Create messages
	now := time.Now()
	msg1 := model.Message{RoomID: suite.testRoomID, SenderID: suite.testUser1ID, Content: "Msg 1", SentAt: now.Add(-time.Minute)}
	msg2 := model.Message{RoomID: suite.testRoomID, SenderID: suite.testUser2ID, Content: "Msg 2", SentAt: now}

	err := suite.db.Create(&msg1).Error
	assert.NoError(suite.T(), err)

	err = suite.db.Create(&msg2).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rooms/%s/messages?asc=true", suite.testRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved messages successfully", responseBody["message"])

	msgResponse := responseBody["data"].(map[string]any)
	assert.NotNil(suite.T(), msgResponse)
	assert.Equal(suite.T(), 1, int(msgResponse["page"].(float64)))
	assert.Equal(suite.T(), 1, int(msgResponse["pageCount"].(float64)))

	msgData := msgResponse["messages"].([]any)
	assert.Len(suite.T(), msgData, 2)
	assert.Equal(suite.T(), "Msg 1", msgData[0].(map[string]any)["content"])
	assert.Equal(suite.T(), "Msg 2", msgData[1].(map[string]any)["content"])

}

func (suite *MessageHandlerTestSuite) TestGetMessages_Success_Descending() {
	// Create messages
	now := time.Now()
	msg1 := model.Message{RoomID: suite.testRoomID, SenderID: suite.testUser1ID, Content: "Msg 1", SentAt: now.Add(-time.Minute)}
	msg2 := model.Message{RoomID: suite.testRoomID, SenderID: suite.testUser2ID, Content: "Msg 2", SentAt: now}

	err := suite.db.Create(&msg1).Error
	assert.NoError(suite.T(), err)

	err = suite.db.Create(&msg2).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rooms/%s/messages?asc=false", suite.testRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved messages successfully", responseBody["message"])

	msgResponse := responseBody["data"].(map[string]any)
	assert.NotNil(suite.T(), msgResponse)
	assert.Equal(suite.T(), 1, int(msgResponse["page"].(float64)))
	assert.Equal(suite.T(), 1, int(msgResponse["pageCount"].(float64)))

	msgData := msgResponse["messages"].([]any)
	assert.Len(suite.T(), msgData, 2)
	assert.Equal(suite.T(), "Msg 2", msgData[0].(map[string]any)["content"])
	assert.Equal(suite.T(), "Msg 1", msgData[1].(map[string]any)["content"])
}

func (suite *MessageHandlerTestSuite) TestGetMessages_Empty() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rooms/%s/messages", suite.testRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved messages successfully", responseBody["message"])

	msgData := responseBody["data"].(map[string]any)["messages"].([]any)
	assert.Empty(suite.T(), msgData)
}

func (suite *MessageHandlerTestSuite) TestCreateMessage_Success() {
	suite.logger.Info("Creating test message for room id: ", suite.testRoomID)
	createReq := request.CreateMessageRequest{
		Content: "New message!",
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rooms/%s/messages", suite.testRoomID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Message saved successfully", responseBody["message"])

	// Verify message in database
	var message model.Message
	err = suite.db.Where("room_id = ? AND sender_id = ? AND content = ?", suite.testRoomID, suite.testUser1ID, createReq.Content).First(&message).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createReq.Content, message.Content)
	assert.Equal(suite.T(), suite.testUser1ID, message.SenderID)
}

func (suite *MessageHandlerTestSuite) TestCreateMessage_InvalidInput_BadJSON() {
	reqBody := bytes.NewBuffer([]byte(`{"content":`)) // Incomplete JSON
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rooms/%s/messages", suite.testRoomID), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])
}

func (suite *MessageHandlerTestSuite) TestCreateMessage_RoomNotFound() {
	suite.logger.Info("Creating test message for room id: ", suite.testRoomID)
	nonExistentRoomID := uuid.NewString()
	createReq := request.CreateMessageRequest{
		Content: "Will this be saved?",
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rooms/%s/messages", nonExistentRoomID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Room not found", responseBody["message"])

	// Ensure no message was saved
	var message model.Message
	err = suite.db.Where("room_id = ?", nonExistentRoomID).First(&message).Error
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, gorm.ErrRecordNotFound)
}

func (suite *MessageHandlerTestSuite) TestCreateMessage_UserNotInRoom() {
	hashedPassword, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)
	user := model.User{
		Username: "user",
		Email:    "user@example.com",
		Password: hashedPassword,
	}
	result := suite.db.Create(&user)
	assert.NoError(suite.T(), result.Error)

	// To simulate this, we can try creating a message with a token for a user we delete.
	token, err := tests.GenerateTestToken(user.ID, user.Username, user.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)

	createReq := request.CreateMessageRequest{
		Content: "Message from a ghost?",
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rooms/%s/messages", suite.testRoomID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusUnauthorized, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User is not in room", responseBody["message"])

	// Ensure no message was saved
	var message model.Message
	err = suite.db.Where("room_id = ? AND sender_id = ?", suite.testRoomID, user.ID).First(&message).Error
	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, gorm.ErrRecordNotFound)
}
