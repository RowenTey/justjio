package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockBillRepository struct {
	mock.Mock
}

func (m *MockBillRepository) WithTx(tx *gorm.DB) BillRepository {
	args := m.Called(tx)
	return args.Get(0).(BillRepository)
}

func (m *MockBillRepository) Create(ctx context.Context, bill *models.Bill) error {
	args := m.Called(ctx, bill)
	return args.Error(0)
}

func (m *MockBillRepository) FindByID(ctx context.Context, billID uint) (*models.Bill, error) {
	args := m.Called(ctx, billID)
	return args.Get(0).(*models.Bill), args.Error(1)
}

func (m *MockBillRepository) FindByRoom(ctx context.Context, roomID string) ([]models.Bill, error) {
	args := m.Called(ctx, roomID)
	return args.Get(0).([]models.Bill), args.Error(1)
}

func (m *MockBillRepository) DeleteByRoom(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

func (m *MockBillRepository) FindByConsolidation(ctx context.Context, consolidationID uint) ([]models.Bill, error) {
	args := m.Called(ctx, consolidationID)
	return args.Get(0).([]models.Bill), args.Error(1)
}

func (m *MockBillRepository) ConsolidateBills(ctx context.Context, roomID string) (*models.Consolidation, error) {
	args := m.Called(ctx, roomID)
	return args.Get(0).(*models.Consolidation), args.Error(1)
}
