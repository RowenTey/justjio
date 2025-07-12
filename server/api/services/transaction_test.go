package services

import (
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/tests"
)

type TransactionServiceTestSuite struct {
	suite.Suite
	transactionService *transactionService

	// DB mocks
	db      *gorm.DB
	sqlMock sqlmock.Sqlmock

	// Mock repositories
	mockTransactionRepo *repository.MockTransactionRepository
	mockBillRepo        *repository.MockBillRepository
}

func TestTransactionServiceSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceTestSuite))
}

func (s *TransactionServiceTestSuite) SetupTest() {
	var err error
	s.db, s.sqlMock, err = tests.SetupTestDB()
	require.NoError(s.T(), err)

	// Initialize mock repositories
	s.mockTransactionRepo = new(repository.MockTransactionRepository)
	s.mockBillRepo = new(repository.MockBillRepository)

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
	bills := []model.Bill{
		{
			ID:           1,
			Name:         "Dinner",
			Amount:       100.0,
			IncludeOwner: true,
			OwnerID:      1,
			Owner:        model.User{ID: 1, Username: "owner"},
			Payers: []model.User{
				{ID: 2, Username: "payer1"},
				{ID: 3, Username: "payer2"},
				{ID: 4, Username: "payer4"},
			},
		},
	}
	consolidation := &model.Consolidation{ID: 1}

	// Expected transactions before consolidation
	expectedAmount := float32(math.Floor(float64(100.0 / 4))) // 25 per person

	// Mock expectations
	s.mockTransactionRepo.On("Create", mock.AnythingOfType("*model.Transaction")).
		Run(func(args mock.Arguments) {
			tx := args.Get(0).(*model.Transaction)
			tx.ID = uint(len(bills) * len(bills[0].Payers)) // Simulate ID generation
		}).
		Return(nil)

	// Execute
	transactions, err := s.transactionService.GenerateTransactions(&bills, consolidation)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), transactions)
	assert.Len(s.T(), *transactions, 3) // After consolidation

	for _, tx := range *transactions {
		assert.Equal(s.T(), consolidation.ID, tx.ConsolidationID)
		assert.True(s.T(), tx.Amount == expectedAmount)
	}
}

func (s *TransactionServiceTestSuite) TestGenerateTransactions_Consolidation() {
	// Setup test data with circular debts
	bills := []model.Bill{
		{
			ID:           1,
			Name:         "Bill1",
			Amount:       100.0,
			IncludeOwner: true,
			OwnerID:      1,
			Owner:        model.User{ID: 1, Username: "user1"},
			Payers: []model.User{
				{ID: 2, Username: "user2"},
			},
		},
		{
			ID:           2,
			Name:         "Bill2",
			Amount:       50.0,
			IncludeOwner: true,
			OwnerID:      2,
			Owner:        model.User{ID: 2, Username: "user2"},
			Payers: []model.User{
				{ID: 1, Username: "user1"},
			},
		},
	}
	consolidation := &model.Consolidation{ID: 1}

	// Mock expectations
	s.mockTransactionRepo.On("Create", mock.AnythingOfType("*model.Transaction")).
		Return(nil)

	// Execute
	transactions, err := s.transactionService.GenerateTransactions(&bills, consolidation)

	// Assertions
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), transactions)
	assert.Len(s.T(), *transactions, 1) // Should be consolidated to one transaction

	finalTx := (*transactions)[0]
	assert.Equal(s.T(), float32(25.0), finalTx.Amount) // 100/2 - 50/2 = 25
}

func (s *TransactionServiceTestSuite) TestGetTransactionsByUser_Success() {
	// Setup test data
	isPaid := true
	userId := "1"
	expectedTransactions := []model.Transaction{
		{ID: 1, PayerID: 1, PayeeID: 2, Amount: 50.0, IsPaid: true},
		{ID: 2, PayerID: 1, PayeeID: 3, Amount: 30.0, IsPaid: true},
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByUser", isPaid, userId).Return(&expectedTransactions, nil)

	// Execute
	transactions, err := s.transactionService.GetTransactionsByUser(isPaid, userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), &expectedTransactions, transactions)
	s.mockTransactionRepo.AssertExpectations(s.T())
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_Success() {
	// Setup test data
	transactionId := "1"
	userId := "1"
	transaction := &model.Transaction{
		ID:      1,
		PayerID: 1,
		PayeeID: 2,
		Amount:  50.0,
		IsPaid:  false,
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByID", transactionId).Return(transaction, nil)
	s.mockTransactionRepo.On("Update", transaction).Return(nil)

	// Execute
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

	// Assertions
	assert.NoError(s.T(), err)
	assert.True(s.T(), result.IsPaid)
	s.mockTransactionRepo.AssertExpectations(s.T())
}

func (s *TransactionServiceTestSuite) TestSettleTransaction_AlreadySettled() {
	// Setup test data
	transactionId := "1"
	userId := "1"
	transaction := &model.Transaction{
		ID:      1,
		PayerID: 1,
		IsPaid:  true,
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByID", transactionId).Return(transaction, nil)

	// Execute
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

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
	transaction := &model.Transaction{
		ID:      1,
		PayerID: 1,
		IsPaid:  false,
	}

	// Mock expectations
	s.mockTransactionRepo.On("FindByID", transactionId).Return(transaction, nil)

	// Execute
	result, err := s.transactionService.SettleTransaction(transactionId, userId)

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
	transactions := []model.Transaction{
		{PayerID: 1, PayeeID: 2, Amount: 100.0},
		{PayerID: 2, PayeeID: 1, Amount: 50.0},
	}
	consolidation := &model.Consolidation{ID: 1}

	// Execute
	result := s.transactionService.consolidateTransactions(&transactions, consolidation)

	// Assertions
	assert.Len(s.T(), *result, 1)
	assert.Equal(s.T(), float32(50.0), (*result)[0].Amount)
	assert.Equal(s.T(), uint(1), (*result)[0].PayerID)
	assert.Equal(s.T(), uint(2), (*result)[0].PayeeID)
}
