package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) WithTx(tx *gorm.DB) RoomRepository {
	args := m.Called(tx)
	return args.Get(0).(RoomRepository)
}

func (m *MockRoomRepository) Create(room *model.Room) error {
	args := m.Called(room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetByID(roomID string) (*model.Room, error) {
	args := m.Called(roomID)
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetByIDWithAttendees(roomID string) (*model.Room, error) {
	args := m.Called(roomID)
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetUserRooms(userID string, page int, pageSize int) (*[]model.Room, error) {
	args := m.Called(userID, page, pageSize)
	return args.Get(0).(*[]model.Room), args.Error(1)
}

func (m *MockRoomRepository) CountUserRooms(userID string) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoomRepository) GetUnjoinedRoomsByIsPrivate(userID string, isPrivate bool) (*[]model.Room, error) {
	args := m.Called(userID, isPrivate)
	return args.Get(0).(*[]model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetRoomAttendees(roomID string) (*[]model.User, error) {
	args := m.Called(roomID)
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (m *MockRoomRepository) GetRoomAttendeeIDs(roomID string) (*[]string, error) {
	args := m.Called(roomID)
	return args.Get(0).(*[]string), args.Error(1)
}

func (m *MockRoomRepository) CloseRoom(roomID string) error {
	args := m.Called(roomID)
	return args.Error(0)
}

func (m *MockRoomRepository) UpdateRoom(room *model.Room) error {
	args := m.Called(room)
	return args.Error(0)
}

func (m *MockRoomRepository) AddUserToRoom(roomID string, user *model.User) error {
	args := m.Called(roomID, user)
	return args.Error(0)
}

func (m *MockRoomRepository) RemoveUserFromRoom(roomID, userID string) error {
	args := m.Called(roomID, userID)
	return args.Error(0)
}

func (m *MockRoomRepository) IsUserInRoom(roomID, userID string) (bool, error) {
	args := m.Called(roomID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoomRepository) GetPendingInvites(userID string) (*[]model.RoomInvite, error) {
	args := m.Called(userID)
	return args.Get(0).(*[]model.RoomInvite), args.Error(1)
}

func (m *MockRoomRepository) CountPendingInvites(userID string) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoomRepository) UpdateInviteStatus(roomID, userID, status string) error {
	args := m.Called(roomID, userID, status)
	return args.Error(0)
}

func (m *MockRoomRepository) CreateInvites(invites *[]model.RoomInvite) error {
	args := m.Called(invites)
	return args.Error(0)
}

func (m *MockRoomRepository) DeletePendingInvites(roomID string) error {
	args := m.Called(roomID)
	return args.Error(0)
}

func (m *MockRoomRepository) HasPendingInvites(roomID, userID string) (bool, error) {
	args := m.Called(roomID, userID)
	return args.Bool(0), args.Error(1)
}
