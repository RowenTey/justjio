package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
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

func (m *MockBillRepository) Create(bill *model.Bill) error {
	args := m.Called(bill)
	return args.Error(0)
}

func (m *MockBillRepository) FindByID(billID uint) (*model.Bill, error) {
	args := m.Called(billID)
	return args.Get(0).(*model.Bill), args.Error(1)
}

func (m *MockBillRepository) FindByRoom(roomID string) (*[]model.Bill, error) {
	args := m.Called(roomID)
	return args.Get(0).(*[]model.Bill), args.Error(1)
}

func (m *MockBillRepository) DeleteByRoom(roomID string) error {
	args := m.Called(roomID)
	return args.Error(0)
}

func (m *MockBillRepository) HasUnconsolidatedBills(roomID string) (bool, error) {
	args := m.Called(roomID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBillRepository) FindByConsolidation(consolidationID uint) (*[]model.Bill, error) {
	args := m.Called(consolidationID)
	return args.Get(0).(*[]model.Bill), args.Error(1)
}

func (m *MockBillRepository) ConsolidateBills(roomID string) (*model.Consolidation, error) {
	args := m.Called(roomID)
	return args.Get(0).(*model.Consolidation), args.Error(1)
}
