package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/services"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
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
	dependencies *tests.TestDependencies

	originalNewAuthService func(
		hashFunc func(string) (string, error),
		jwtSecret string,
		sendEmail func(string, string, string, string) error,
		googleConfig *oauth2.Config,
	) *services.AuthService
}

func (suite *AuthHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx)
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
	fmt.Println(kafkaBrokers)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Post("/signup", SignUp)
}

func (suite *AuthHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	suite.dependencies.Teardown(suite.ctx)
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	// Mock dependencies
	mockHash := func(password string) (string, error) {
		return "hashedpassword123", nil // Return known hash
	}
	mockJWTSecret := "test-secret" // Same as in test config
	mockSendEmail := func(otp, username, email, purpose string) error {
		return nil // Always succeed
	}
	mockGoogleConfig := &oauth2.Config{} // Empty config for tests

	// Replace the auth service creation in your handler
	suite.originalNewAuthService = services.NewAuthService
	services.NewAuthService = func(
		hashFunc func(string) (string, error),
		jwtSecret string,
		sendEmail func(string, string, string, string) error,
		googleConfig *oauth2.Config,
	) *services.AuthService {
		return &services.AuthService{
			HashFunc:      mockHash,
			JwtSecret:     mockJWTSecret,
			SendSMTPEmail: mockSendEmail,
			OAuthConfig:   mockGoogleConfig,
			Logger:        log.WithFields(log.Fields{"service": "AuthService"}),
		}
	}

	database.DB = suite.db // Set the database connection for the auth service
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE users CASCADE")

	database.DB = nil                                      // Reset the database connection
	services.NewAuthService = suite.originalNewAuthService // Restore original auth service creation
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
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "User signed up successfully", response["message"])
	assert.NotEmpty(suite.T(), response["data"].(map[string]interface{})["id"])

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
	resp, err := suite.app.Test(req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)

	// Verify response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Username or email already exists", response["message"])
}
