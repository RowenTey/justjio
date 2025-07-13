package handlers

import (
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
	pushNotificationsModel "github.com/RowenTey/JustJio/server/api/model/push_notifications"
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

type TransactionHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	mockJWTSecret string

	transactionService services.TransactionService

	testNotificationChan chan pushNotificationsModel.NotificationData

	testPayerID         uint
	testPayeeID         uint
	testConsolidationID uint
	testPayerToken      string
	testPayeeToken      string
	testTransaction     *model.Transaction
}

func (suite *TransactionHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx, suite.logger)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println("Test DB Connection String:", pgConnStr)

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	// Initialize deps
	suite.mockJWTSecret = "test-secret"
	suite.testNotificationChan = make(chan pushNotificationsModel.NotificationData, 100)
	transactionRepo := repository.NewTransactionRepository(suite.db)
	billRepo := repository.NewBillRepository(suite.db)
	suite.transactionService = services.NewTransactionService(
		transactionRepo,
		billRepo,
		suite.logger,
	)
	notificationRepo := repository.NewNotificationRepository(suite.db)
	subscriptionRepo := repository.NewSubscriptionRepository(suite.db)
	notificationService := services.NewNotificationService(
		notificationRepo,
		subscriptionRepo,
		suite.testNotificationChan,
		suite.logger,
	)
	transactionHandler := NewTransactionHandler(suite.transactionService, notificationService, suite.logger)

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(suite.mockJWTSecret))

	// Register Transaction routes
	transactionRoutes := suite.app.Group("/transactions")
	transactionRoutes.Get("/", transactionHandler.GetTransactionsByUser)
	transactionRoutes.Patch("/:txId/settle", transactionHandler.SettleTransaction)
}

func (suite *TransactionHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	log.Info("Tore down test suite dependencies")
	close(suite.testNotificationChan)
}

func (suite *TransactionHandlerTestSuite) SetupTest() {
	// Create test payer user
	hashedPassword1, err := utils.HashPassword("password123")
	assert.NoError(suite.T(), err)
	payer := model.User{
		Username: "payeruser",
		Email:    "payer@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&payer)
	assert.NoError(suite.T(), result.Error)
	suite.testPayerID = payer.ID
	payerToken, err := tests.GenerateTestToken(payer.ID, payer.Username, payer.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testPayerToken = payerToken

	// Create test payee user
	hashedPassword2, _ := utils.HashPassword("password456")
	payee := model.User{
		Username: "payeeuser",
		Email:    "payee@example.com",
		Password: hashedPassword2,
	}
	result = suite.db.Create(&payee)
	assert.NoError(suite.T(), result.Error)
	suite.testPayeeID = payee.ID
	payeeToken, err := tests.GenerateTestToken(payee.ID, payee.Username, payee.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)
	suite.testPayeeToken = payeeToken

	// Create test room
	room := model.Room{
		ID:     uuid.NewString(),
		Name:   "Test Room",
		HostID: suite.testPayerID,
		Users:  []model.User{payer, payee},
	}
	result = suite.db.Create(&room)
	assert.NoError(suite.T(), result.Error)

	// Create test consolidation
	consolidation := model.Consolidation{}
	result = suite.db.Create(&consolidation)
	assert.NoError(suite.T(), result.Error)

	// Create test bill
	bill := model.Bill{
		Name:            "Test Bill",
		Amount:          100.00,
		Date:            time.Now(),
		IncludeOwner:    true,
		RoomID:          room.ID,
		OwnerID:         suite.testPayerID,
		ConsolidationID: consolidation.ID,
		Payers:          []model.User{payer, payee},
	}
	result = suite.db.Create(&bill)
	assert.NoError(suite.T(), result.Error)
	suite.testConsolidationID = consolidation.ID

	// Create test transaction
	transaction := model.Transaction{
		ConsolidationID: consolidation.ID,
		PayerID:         suite.testPayerID,
		PayeeID:         suite.testPayeeID,
		Amount:          50.00,
		IsPaid:          false,
		Consolidation:   consolidation,
		Payer:           payer,
		Payee:           payee,
	}
	result = suite.db.Create(&transaction)
	assert.NoError(suite.T(), result.Error)
	suite.testTransaction = &transaction

	log.Infof("SetupTest complete: Payer ID=%d, Payee ID=%d, Transaction ID=%d, Bill ID=%d, Consolidation ID=%d",
		suite.testPayerID, suite.testPayeeID, transaction.ID, bill.ID, consolidation.ID)
}

func (suite *TransactionHandlerTestSuite) TearDownTest() {
	// Clear database after each test
	suite.db.Exec("TRUNCATE TABLE transactions RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE bills RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE consolidations RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE payers RESTART IDENTITY CASCADE")
	log.Info("Tore down test data")
}

func TestTransactionHandlerSuite(t *testing.T) {
	suite.Run(t, new(TransactionHandlerTestSuite))
}

func (suite *TransactionHandlerTestSuite) TestGetTransactionsByUser_Success() {
	// Create another transaction for the payer
	extraTx := model.Transaction{
		ConsolidationID: suite.testConsolidationID,
		PayerID:         suite.testPayerID,
		PayeeID:         suite.testPayeeID,
		Amount:          30.00,
		IsPaid:          true,
	}
	err := suite.db.Create(&extraTx).Error
	assert.NoError(suite.T(), err)

	// Test getting unpaid transactions
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testPayerToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved transactions successfully", responseBody["message"])

	transactionData := responseBody["data"].([]any)
	assert.Len(suite.T(), transactionData, 1)

	// Test getting only paid transactions
	req = httptest.NewRequest(http.MethodGet, "/transactions?isPaid=true", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testPayerToken)

	resp, err = suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var paidTransactions map[string]any
	err = json.NewDecoder(resp.Body).Decode(&paidTransactions)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved transactions successfully", paidTransactions["message"])

	paidTransactionData := paidTransactions["data"].([]any)
	assert.Len(suite.T(), paidTransactionData, 1)
	// Convert the parsed float64 ID to uint
	txID := uint(paidTransactionData[0].(map[string]any)["id"].(float64))
	assert.Equal(suite.T(), extraTx.ID, txID)
}

func (suite *TransactionHandlerTestSuite) TestGetTransactionsByUser_Empty() {
	// Create extra test user
	hashedPassword, _ := utils.HashPassword("password123")
	testUser := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	result := suite.db.Create(&testUser)
	assert.NoError(suite.T(), result.Error)
	testUserToken, err := tests.GenerateTestToken(testUser.ID, testUser.Username, testUser.Email, suite.mockJWTSecret)
	assert.NoError(suite.T(), err)

	// Test with a user who has no transactions
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+testUserToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved transactions successfully", responseBody["message"])

	transactionData := responseBody["data"].([]any)
	assert.Empty(suite.T(), transactionData)
}

func (suite *TransactionHandlerTestSuite) TestSettleTransaction_Success() {
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/transactions/%d/settle", suite.testTransaction.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testPayerToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Paid transactions successfully", responseBody["message"])

	// Verify transaction is marked as paid
	var updatedTx model.Transaction
	err = suite.db.First(&updatedTx, "id = ?", suite.testTransaction.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedTx.IsPaid)
}

func (suite *TransactionHandlerTestSuite) TestSettleTransaction_AlreadySettled() {
	// Mark transaction as paid first
	err := suite.db.Model(&suite.testTransaction).Update("is_paid", true).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/transactions/%d/settle", suite.testTransaction.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testPayerToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusConflict, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "transaction already settled", responseBody["message"])
}

func (suite *TransactionHandlerTestSuite) TestSettleTransaction_InvalidPayer() {
	// Try to settle with payee's token (not the payer)
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/transactions/%d/settle", suite.testTransaction.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testPayeeToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid payer", responseBody["message"])
}

func (suite *TransactionHandlerTestSuite) TestSettleTransaction_NotFound() {
	nonExistentTxID := uint(999)
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/transactions/%d/settle", nonExistentTxID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testPayerToken)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Transaction not found", responseBody["message"])
}
