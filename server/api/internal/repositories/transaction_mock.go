package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
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

func (m *MockTransactionRepository) Create(ctx context.Context, transactions []models.Transaction) error {
	args := m.Called(ctx, transactions)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByUser(ctx context.Context, isPaid bool, userID string) ([]models.Transaction, error) {
	args := m.Called(ctx, isPaid, userID)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByID(ctx context.Context, transactionID string) (*models.Transaction, error) {
	args := m.Called(ctx, transactionID)
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}
