package services

import (
	"context"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/internal/repositories"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
	"github.com/RowenTey/JustJio/server/api/pkg/tests"
)

type TransactionServiceTestSuite struct {
	suite.Suite
	transactionService *transactionService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockTransactionRepo *repositories.MockTransactionRepository
	mockBillRepo        *repositories.MockBillRepository
}

func TestTransactionServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TransactionServiceTestSuite))
}

func (s *TransactionServiceTestSuite) SetupTest() {
	var err error
	s.db, s.sqlMock, err = tests.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize mock repositories
	s.mockTransactionRepo = new(repositories.MockTransactionRepository)
	s.mockBillRepo = new(repositories.MockBillRepository)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create transactionService with mock dependencies
	s.transactionService = NewTransactionService(
		s.mockTransactionRepo,
		s.mockBillRepo,
		logger,
	).(*transactionService)
}

func (s *TransactionServiceTestSuite) TestGenerateTransactions_Success() {
	// Setup test data
	bills := []models.Bill{
		{
			ID:           1,
			Name:         "Dinner",
			Amount:       100.0,
			IncludeOwner: true,
			OwnerID:      1,
			Owner:        models.User{ID: 1, Username: "owner"},
			Payers: []models.User{
				{ID: 2, Username: "payer1"},
				{ID: 3, Username: "payer2"},
				{ID: 4, Username: "payer4"},
			},
		},
	}
	consolidation := &models.Consolidation{ID: 1}

	// Expected transactions before consolidation
	expectedAmount := float32(math.Floor(float64(100.0 / 4))) // 25 per person

	// Mock expectations
	s.mockTransactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Transaction")).
		Run(func(args mock.Arguments) {
			tx := args.Get(1).(*models.Transaction)         // Get second argument (index 1) since first is context
			tx.ID = uint(len(bills) * len(bills[0].Payers)) // Simulate ID generation
		}).
		Return(nil)

	// Execute
	transactions, err := s.transactionService.GenerateTransactions(bills, consolidation)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), transactions)
	assert.Len(s.T(), transactions, 3) // After consolidation

	for _, tx := range transactions {
		assert.Equal(s.T(), consolidation.ID, tx.ConsolidationID)
		assert.True(s.T(), tx.Amount == expectedAmount)
	}
}

func (s *TransactionServiceTestSuite) TestGenerateTransactions_Consolidation() {
	// Setup test data with circular debts
	bills := []models.Bill{
		{
			ID:           1,
			Name:         "Bill1",
			Amount:       100.0,
			IncludeOwner: true,
			OwnerID:      1,
			Owner:        models.User{ID: 1, Username: "user1"},
			Payers: []models.User{
				{ID: 2, Username: "user2"},
			},
		},
		{
			ID:           2,
			Name:         "Bill2",
			Amount:       50.0,
			IncludeOwner: true,
			OwnerID:      2,
			Owner:        models.User{ID: 2, Username: "user2"},
			Payers: []models.User{
				{ID: 1, Username: "user1"},
			},
		},
	}
	consolidation := &models.Consolidation{ID: 1}

	// Mock expectations
	s.mockTransactionRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Transaction")).
		Return(nil)

	// Execute
	transactions, err := s.transactionService.GenerateTransactions(bills, consolidation)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), transactions)
	assert.Len(s.T(), transactions, 1) // Should be consolidated to one transaction

	finalTx := transactions[0]
	assert.Equal(s.T(), float32(25.0), finalTx.Amount) // 100/2 - 50/2 = 25
}

func (s *TransactionServiceTestSuite) TestGetTransactionsByUser_Success() {
	// Setup test data
	isPaid := true
	userId := "1"
	expectedTransactions := []models.Transaction{
		{ID: 1, PayerID: 1, PayeeID: 2, Amount: 50.0, IsPaid: true},
		{ID: 2, PayerID: 1, PayeeID: 3, Amount: 30.0, IsPaid: true},
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByUser", mock.Anything, isPaid, userId).Return(expectedTransactions, nil)
	s.mockTransactionRepo.On("FindByUser", mock.Anything, isPaid, userId).Return(expectedTransactions, nil)

	// Execute
	transactions, err := s.transactionService.GetTransactionsByUser(context.Background(), isPaid, userId)

	// Assertions
	assert.NoError(s.T(), err)
	expectedDto := make([]response.TransactionDto, len(expectedTransactions))
	for i, tx := range expectedTransactions {
		expectedDto[i] = response.TransactionDto{
			ID:              tx.ID,
			ConsolidationID: tx.ConsolidationID,
			Amount:          tx.Amount,
			IsPaid:          tx.IsPaid,
			PaidOn:          tx.PaidOn,
			Payer: response.MinimalUserDto{
				ID:         tx.Payer.ID,
				Username:   tx.Payer.Username,
				PictureUrl: tx.Payer.PictureUrl,
			},
			Payee: response.MinimalUserDto{
				ID:         tx.Payee.ID,
				Username:   tx.Payee.Username,
				PictureUrl: tx.Payee.PictureUrl,
			},
		}
	}
	assert.Equal(s.T(), expectedDto, transactions)
	s.mockTransactionRepo.AssertExpectations(s.T())
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_Success() {
	// Setup test data
	transactionId := "1"
	userId := "1"
	transaction := &models.Transaction{
		ID:      1,
		PayerID: 1,
		PayeeID: 2,
		Amount:  50.0,
		IsPaid:  false,
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByID", mock.Anything, transactionId).Return(transaction, nil)
	s.mockTransactionRepo.On("Update", mock.Anything, transaction).Return(nil)
	s.mockTransactionRepo.On("FindByID", mock.Anything, transactionId).Return(transaction, nil)
	s.mockTransactionRepo.On("Update", mock.Anything, transaction).Return(nil)

	// Execute
	result, err := s.transactionService.SettleTransaction(context.Background(), transactionId, userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.True(s.T(), result.IsPaid)
	s.mockTransactionRepo.AssertExpectations(s.T())
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_AlreadySettled() {
	// Setup test data
	transactionId := "1"
	userId := "1"
	transaction := &models.Transaction{
		ID:      1,
		PayerID: 1,
		IsPaid:  true,
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByID", mock.Anything, transactionId).Return(transaction, nil)
	s.mockTransactionRepo.On("FindByID", mock.Anything, transactionId).Return(transaction, nil)

	// Execute
	result, err := s.transactionService.SettleTransaction(context.Background(), transactionId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrTransactionAlreadySettled, err)
	assert.Nil(s.T(), result)
	s.mockTransactionRepo.AssertExpectations(s.T())
	s.mockTransactionRepo.AssertNotCalled(s.T(), "Update")
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_InvalidPayer() {
	// Setup test data
	transactionId := "1"
	userId := "2" // Not the payer
	transaction := &models.Transaction{
		ID:      1,
		PayerID: 1,
		IsPaid:  false,
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByID", mock.Anything, transactionId).Return(transaction, nil)
	s.mockTransactionRepo.On("FindByID", mock.Anything, transactionId).Return(transaction, nil)

	// Execute
	result, err := s.transactionService.SettleTransaction(context.Background(), transactionId, userId)

	// Assertions
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrInvalidPayer, err)
	assert.Nil(s.T(), result)
	s.mockTransactionRepo.AssertExpectations(s.T())
	s.mockTransactionRepo.AssertNotCalled(s.T(), "Update")
}

func (s *TransactionServiceTestSuite) TestRemoveCycle_CycleExists() {
	// Setup test data
	graph := map[uint][]edge{
		1: {{userId: 2, amount: 100.0}},
		2: {{userId: 3, amount: 50.0}},
		3: {{userId: 1, amount: 30.0}},
	}
	visited := map[uint]bool{1: true, 2: false, 3: false}

	// Execute
	amount, _ := s.transactionService.removeCycle(1, graph, visited)

	// Assertions
	assert.Equal(s.T(), float32(30.0), amount)
	assert.Equal(s.T(), 0, len(graph[3]))
}

func (s *TransactionServiceTestSuite) TestConsolidateTransactions() {
	// Setup test data
	transactions := []models.Transaction{
		{PayerID: 1, PayeeID: 2, Amount: 100.0},
		{PayerID: 2, PayeeID: 1, Amount: 50.0},
	}
	consolidation := &models.Consolidation{ID: 1}

	// Execute
	result := s.transactionService.consolidateTransactions(transactions, consolidation)

	// Assertions
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), float32(50.0), result[0].Amount)
	assert.Equal(s.T(), uint(1), result[0].PayerID)
	assert.Equal(s.T(), uint(2), result[0].PayeeID)
}
