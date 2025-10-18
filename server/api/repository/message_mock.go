package repository

import (
	"context"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) WithTx(tx *gorm.DB) MessageRepository {
	args := m.Called(tx)
	return args.Get(0).(MessageRepository)
}

func (m *MockMessageRepository) Create(ctx context.Context, message *model.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(ctx context.Context, msgID string) (*model.Message, error) {
	args := m.Called(ctx, msgID)
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockMessageRepository) Delete(ctx context.Context, msgID string) error {
	args := m.Called(ctx, msgID)
	return args.Error(0)
}

func (m *MockMessageRepository) DeleteByRoom(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

func (m *MockMessageRepository) CountByRoom(ctx context.Context, roomID string) (int64, error) {
	args := m.Called(ctx, roomID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMessageRepository) FindByRoom(ctx context.Context, roomId string, page int, pageSize int, asc bool) ([]model.Message, error) {
	args := m.Called(ctx, roomId, page, pageSize, asc)
	return args.Get(0).([]model.Message), args.Error(1)
}
