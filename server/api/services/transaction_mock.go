package services

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/dto/response"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) GenerateTransactions(bills []model.Bill, consolidation *model.Consolidation) ([]model.Transaction, error) {
	args := m.Called(bills, consolidation)
	return args.Get(0).([]model.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByUser(ctx context.Context, isPaid bool, userId string) ([]response.TransactionDto, error) {
	args := m.Called(ctx, isPaid, userId)
	return args.Get(0).([]response.TransactionDto), args.Error(1)
}

func (m *MockTransactionService) SettleTransaction(ctx context.Context, transactionId string, userId string) (*model.Transaction, error) {
	args := m.Called(ctx, transactionId, userId)
	return args.Get(0).(*model.Transaction), args.Error(1)
}
