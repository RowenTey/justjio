package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MessageHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	dependencies *tests.TestDependencies
	kafkaService *services.KafkaService

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

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println("Test DB Connection String:", pgConnStr) // Log for debugging

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	// Get Kafka broker address
	kafkaBrokers, err := suite.dependencies.KafkaContainer.Brokers(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.kafkaService, err = services.NewKafkaService(kafkaBrokers[0], "test")
	assert.NoError(suite.T(), err)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(mockJWTSecret))

	// Register Message routes
	messageRoutes := suite.app.Group("/rooms/:roomId/messages")
	messageRoutes.Use(middleware.IsUserInRoom)
	messageRoutes.Get("/:msgId", GetMessage)
	messageRoutes.Get("/", GetMessages)
	messageRoutes.Post("/", func(c *fiber.Ctx) error {
		return CreateMessage(c, suite.kafkaService)
	})
}

func (suite *MessageHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	log.Info("Tore down test suite dependencies")
}

func (suite *MessageHandlerTestSuite) SetupTest() {
	// Assign the test DB to the global variable used by handlers/services
	database.DB = suite.db
	assert.NotNil(suite.T(), database.DB, "Global DB should be set")

	// Create User 1
	hashedPassword1, _ := utils.HashPassword("password123")
	user1 := model.User{
		Username: "user1",
		Email:    "user1@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&user1)
	assert.NoError(suite.T(), result.Error)
	suite.testUser1ID = user1.ID
	token1, err := generateTestToken(user1.ID, user1.Username, user1.Email)
	assert.NoError(suite.T(), err)
	suite.testUser1Token = token1

	// Create User 2
	hashedPassword2, _ := utils.HashPassword("password456")
	user2 := model.User{
		Username: "user2",
		Email:    "user2@example.com",
		Password: hashedPassword2,
	}
	result = suite.db.Create(&user2)
	assert.NoError(suite.T(), result.Error)
	suite.testUser2ID = user2.ID
	token2, err := generateTestToken(user2.ID, user2.Username, user2.Email)
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

	// message := model.Message{
	// 	RoomID:   suite.testRoomID,
	// 	SenderID: suite.testUser1ID,
	// 	Content:  "Hello World!",
	// }
	// result = suite.db.Create(&message)
	// assert.NoError(suite.T(), result.Error)

	log.Infof("SetupTest complete: User1 ID=%d, User2 ID=%d, Room ID=%s", suite.testUser1ID, suite.testUser2ID, suite.testRoomID)
}

func (suite *MessageHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE messages CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms CASCADE")
	suite.db.Exec("TRUNCATE TABLE users CASCADE")

	// Reset the global DB variable
	database.DB = nil
	log.Info("Tore down test data and reset global DB")
}

func TestMessageHandlerSuite(t *testing.T) {
	suite.Run(t, new(MessageHandlerTestSuite))
}

func (suite *MessageHandlerTestSuite) TestGetMessage_Success() {
	// Create a message
	log.Info("Creating test message for room id: ", suite.testRoomID)
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
	log.Info("Creating test message for room id: ", suite.testRoomID)
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
	log.Info("Creating test message for room id: ", suite.testRoomID)
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
	hashedPassword, _ := utils.HashPassword("password123")
	user := model.User{
		Username: "user",
		Email:    "user@example.com",
		Password: hashedPassword,
	}
	result := suite.db.Create(&user)
	assert.NoError(suite.T(), result.Error)

	// To simulate this, we can try creating a message with a token for a user we delete.
	token, err := generateTestToken(user.ID, user.Username, user.Email)
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
