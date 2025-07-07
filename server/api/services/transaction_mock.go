package services

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) GenerateTransactions(bills *[]model.Bill, consolidation *model.Consolidation) (*[]model.Transaction, error) {
	args := m.Called(bills, consolidation)
	return args.Get(0).(*[]model.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByUser(isPaid bool, userId string) (*[]model.Transaction, error) {
	args := m.Called(isPaid, userId)
	return args.Get(0).(*[]model.Transaction), args.Error(1)
}

func (m *MockTransactionService) SettleTransaction(transactionId string, userId string) (*model.Transaction, error) {
	args := m.Called(transactionId, userId)
	return args.Get(0).(*model.Transaction), args.Error(1)
}
