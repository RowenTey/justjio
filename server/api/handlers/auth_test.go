package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RowenTey/JustJio/server/api/config"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	mockJWTSecret string

	kafkaService services.KafkaService
	userService  *services.UserService
	authService  *services.AuthService

	authHandler *AuthHandler
}

func (suite *AuthHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx, suite.logger)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println(pgConnStr)

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
}

func (suite *AuthHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	suite.dependencies.Teardown(suite.ctx)
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	userRepository := repository.NewUserRepository(suite.db)
	suite.userService = services.NewUserService(suite.db, userRepository, suite.logger)
	suite.authService = services.NewAuthService(
		suite.userService,
		suite.kafkaService,
		func(password string) (string, error) {
			return utils.HashPassword(password)
		},
		func(from, to, subject, textBody string) error {
			return nil
		},
		suite.mockJWTSecret,
		"test@test.com",
		&oauth2.Config{},
		suite.logger,
	)
	suite.authHandler = NewAuthHandler(suite.authService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Post("/signup", suite.authHandler.SignUp)
	suite.app.Post("/login", suite.authHandler.Login)
	suite.app.Post("/verify", suite.authHandler.VerifyOTP)
	suite.app.Post("/otp", suite.authHandler.SendOTPEmail)
	suite.app.Patch("/reset", suite.authHandler.ResetPassword)
	suite.app.Post("/google", suite.authHandler.GoogleLogin)
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (suite *AuthHandlerTestSuite) TestSignUp_Success() {
	// Prepare request
	newUser := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	reqBody, err := json.Marshal(newUser)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "User signed up successfully", response["message"])
	assert.NotEmpty(suite.T(), response["data"].(map[string]any)["id"])

	// Verify database
	var dbUser model.User
	err = suite.db.Where("username = ?", "testuser").First(&dbUser).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test@example.com", dbUser.Email)
}

func (suite *AuthHandlerTestSuite) TestSignUp_DuplicateEmail() {
	// Create existing user
	existingUser := model.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
	}
	err := suite.db.Create(&existingUser).Error
	assert.NoError(suite.T(), err)

	// Prepare request with duplicate email
	newUser := model.User{
		Username: "testuser",
		Email:    "existing@example.com", // duplicate email
		Password: "password123",
	}

	reqBody, err := json.Marshal(newUser)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Username or email already exists", response["message"])
}

func (suite *AuthHandlerTestSuite) TestSignUp_InvalidInput() {
	// Prepare request with invalid data
	reqBody := bytes.NewBuffer([]byte(`{"random", "password": "password123"}`))

	req := httptest.NewRequest("POST", "/signup", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Review your input", response["message"])
}

func (suite *AuthHandlerTestSuite) TestSignUp_InternalServerError() {
	// Mock the hash function to return an error
	mockHash := func(password string) (string, error) {
		return "", errors.New("hashing error")
	}

	// Re-initialize deps
	suite.authService = services.NewAuthService(
		suite.userService,
		suite.kafkaService,
		mockHash,
		func(from, to, subject, textBody string) error {
			return nil
		},
		suite.mockJWTSecret,
		"test@test.com",
		&oauth2.Config{},
		suite.logger,
	)
	suite.authHandler = NewAuthHandler(suite.authService, suite.logger)

	// Re-setup Fiber app
	suite.app = fiber.New()
	suite.app.Post("/signup", suite.authHandler.SignUp)

	// Prepare request
	newUser := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	reqBody, err := json.Marshal(newUser)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusInternalServerError, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Error occured in server", response["message"])
}

func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	// Create a user
	hashedPassword, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)

	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	err = suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	// Prepare request
	reqBody := bytes.NewBuffer([]byte(`{"username": "testuser", "password": "password123"}`))

	req := httptest.NewRequest("POST", "/login", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Login successfully", response["message"])
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidInput() {
	// Prepare request with invalid data
	reqBody := bytes.NewBuffer([]byte(`{"random", "password": "password123"}`))

	req := httptest.NewRequest("POST", "/login", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Review your input", response["message"])
}

func (suite *AuthHandlerTestSuite) TestLogin_UserNotFound() {
	// Prepare request with non-existent username
	reqBody := bytes.NewBuffer([]byte(`{"username": "nonexistentuser", "password": "password123"}`))

	req := httptest.NewRequest("POST", "/login", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "User not found", response["message"])
}

func (suite *AuthHandlerTestSuite) TestLogin_IncorrectPassword() {
	// Create a user
	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: utils.GenerateRandomString(32),
	}
	err := suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	// Prepare request with incorrect password
	reqBody := bytes.NewBuffer([]byte(`{"username": "testuser", "password": "wrongpassword"}`))

	req := httptest.NewRequest("POST", "/login", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusUnauthorized, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Invalid username or password", response["message"])
}

func (suite *AuthHandlerTestSuite) TestSendOTPEmail_InvalidInput() {
	// Prepare request with invalid data
	reqBody := bytes.NewBuffer([]byte(`{"random", "purpose": "verify-email"}`))

	req := httptest.NewRequest("POST", "/otp", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Review your input", response["message"])
}

func (suite *AuthHandlerTestSuite) TestSendOTPEmail_UserNotFound() {
	// Prepare request with non-existent email
	reqBody := bytes.NewBuffer([]byte(`{"email": "nonexistent@example.com", "purpose": "verify-email"}`))

	req := httptest.NewRequest("POST", "/otp", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "User not found", response["message"])
}

func (suite *AuthHandlerTestSuite) TestSendOTPEmail_EmailAlreadyVerified() {
	// Create a user with verified email
	user := model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		Password:     utils.GenerateRandomString(32),
		IsEmailValid: true,
	}
	err := suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	// Prepare request with verified email
	reqBody := bytes.NewBuffer([]byte(`{"email": "test@example.com", "purpose": "verify-email"}`))

	req := httptest.NewRequest("POST", "/otp", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Email already verified", response["message"])
}

func (suite *AuthHandlerTestSuite) TestVerifyOTP_InvalidInput() {
	// Prepare request with invalid data
	reqBody := bytes.NewBuffer([]byte(`{"test@email.com", "otp": "123456"}`))

	req := httptest.NewRequest("POST", "/verify", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Review your input", response["message"])
}

func (suite *AuthHandlerTestSuite) TestVerifyOTP_UserNotFound() {
	// Prepare request with non-existent email
	reqBody := bytes.NewBuffer([]byte(`{"email": "nonexistent@example.com", "otp": "123456"}`))

	req := httptest.NewRequest("POST", "/verify", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "User not found", response["message"])
}

func (suite *AuthHandlerTestSuite) TestVerifyOTP_OTPNotFound() {
	// Create a user
	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: utils.GenerateRandomString(32),
	}
	err := suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	// Prepare request with valid email but no OTP stored
	reqBody := bytes.NewBuffer([]byte(`{"email": "test@example.com", "otp": "123456"}`))

	req := httptest.NewRequest("POST", "/verify", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "OTP not found", response["message"])
}

func (suite *AuthHandlerTestSuite) TestVerifyOTP_InvalidOTP() {
	// Create a user
	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: utils.GenerateRandomString(32),
	}
	err := suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	// Store a valid OTP
	suite.authHandler.ClientOtpMap.Store("test@example.com", "123456")
	defer suite.authHandler.ClientOtpMap.Delete("test@example.com")

	// Prepare request with invalid OTP
	reqBody := bytes.NewBuffer([]byte(`{"email": "test@example.com", "otp": "654321"}`))

	req := httptest.NewRequest("POST", "/verify", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Invalid OTP", response["message"])
}

func (suite *AuthHandlerTestSuite) TestResetPassword_InvalidInput() {
	// Prepare request with invalid data
	reqBody := bytes.NewBuffer([]byte(`{"random", "password": "password123"}`))

	req := httptest.NewRequest("PATCH", "/reset", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Review your input", response["message"])
}

func (suite *AuthHandlerTestSuite) TestResetPassword_UserNotFound() {
	// Prepare request with non-existent email
	reqBody := bytes.NewBuffer([]byte(`{"email": "nonexistent@example.com", "password": "password123"}`))

	req := httptest.NewRequest("PATCH", "/reset", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "User not found", response["message"])
}

func (suite *AuthHandlerTestSuite) TestGoogleLogin_InvalidInput() {
	// Prepare request with invalid data
	reqBody := bytes.NewBuffer([]byte(`{"random"}`))

	req := httptest.NewRequest("POST", "/google", reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	// Verify response
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Review your input", response["message"])
}
