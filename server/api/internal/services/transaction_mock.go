package services

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/response"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) GenerateTransactions(bills []models.Bill, consolidation *models.Consolidation) ([]models.Transaction, error) {
	args := m.Called(bills, consolidation)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByUser(ctx context.Context, isPaid bool, userId string) ([]response.TransactionDto, error) {
	args := m.Called(ctx, isPaid, userId)
	return args.Get(0).([]response.TransactionDto), args.Error(1)
}

func (m *MockTransactionService) SettleTransaction(ctx context.Context, transactionId string, userId string) (*models.Transaction, error) {
	args := m.Called(ctx, transactionId, userId)
	return args.Get(0).(*models.Transaction), args.Error(1)
}
