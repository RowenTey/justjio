package repository

import (
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

func (m *MockMessageRepository) Create(message *model.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(msgID, roomID string) (*model.Message, error) {
	args := m.Called(msgID, roomID)
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockMessageRepository) Delete(msgID, roomID string) error {
	args := m.Called(msgID, roomID)
	return args.Error(0)
}

func (m *MockMessageRepository) DeleteByRoom(roomID string) error {
	args := m.Called(roomID)
	return args.Error(0)
}

func (m *MockMessageRepository) CountByRoom(roomID string) (int64, error) {
	args := m.Called(roomID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMessageRepository) FindByRoom(roomId string, page int, pageSize int, asc bool) (*[]model.Message, error) {
	args := m.Called(roomId, page, pageSize, asc)
	return args.Get(0).(*[]model.Message), args.Error(1)
}
