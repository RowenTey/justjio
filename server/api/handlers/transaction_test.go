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
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TransactionHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	dependencies *tests.TestDependencies

	testPayerID         uint
	testPayeeID         uint
	testConsolidationID uint
	testPayerToken      string
	testPayeeToken      string
	testTransaction     *model.Transaction
	testNotifChan       chan NotificationData
}

func (suite *TransactionHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx)
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

	// Setup Fiber app
	suite.app = fiber.New()
	suite.app.Use(middleware.Authenticated(mockJWTSecret))

	// Notification channel for testing
	suite.testNotifChan = make(chan NotificationData, 100)

	// Register Transaction routes
	transactionRoutes := suite.app.Group("/transactions")
	transactionRoutes.Get("/", GetTransactionsByUser)
	transactionRoutes.Patch("/:txId/settle", func(c *fiber.Ctx) error {
		return SettleTransaction(c, suite.testNotifChan)
	})
}

func (suite *TransactionHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	log.Info("Tore down test suite dependencies")
	close(suite.testNotifChan)
}

func (suite *TransactionHandlerTestSuite) SetupTest() {
	// Assign the test DB to the global variable used by handlers/services
	database.DB = suite.db
	assert.NotNil(suite.T(), database.DB, "Global DB should be set")

	// Create test payer user
	hashedPassword1, _ := utils.HashPassword("password123")
	payer := model.User{
		Username: "payeruser",
		Email:    "payer@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&payer)
	assert.NoError(suite.T(), result.Error)
	suite.testPayerID = payer.ID
	payerToken, err := generateTestToken(payer.ID, payer.Username, payer.Email)
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
	payeeToken, err := generateTestToken(payee.ID, payee.Username, payee.Email)
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
	suite.db.Exec("TRUNCATE TABLE transactions CASCADE")
	suite.db.Exec("TRUNCATE TABLE bills CASCADE")
	suite.db.Exec("TRUNCATE TABLE consolidations CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms CASCADE")
	suite.db.Exec("TRUNCATE TABLE users CASCADE")
	suite.db.Exec("TRUNCATE TABLE payers CASCADE")

	// Reset the global DB variable
	database.DB = nil
	log.Info("Tore down test data and reset global DB")
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
	testUserToken, err := generateTestToken(testUser.ID, testUser.Username, testUser.Email)
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
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

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
	assert.Equal(suite.T(), fiber.StatusUnauthorized, resp.StatusCode)

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
