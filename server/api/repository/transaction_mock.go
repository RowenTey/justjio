package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) WithTx(tx *gorm.DB) TransactionRepository {
	args := m.Called(tx)
	return args.Get(0).(TransactionRepository)
}

func (m *MockTransactionRepository) Create(transactions *[]model.Transaction) error {
	args := m.Called(transactions)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByUser(isPaid bool, userID string) (*[]model.Transaction, error) {
	args := m.Called(isPaid, userID)
	return args.Get(0).(*[]model.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByID(transactionID string) (*model.Transaction, error) {
	args := m.Called(transactionID)
	return args.Get(0).(*model.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(transaction *model.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}
