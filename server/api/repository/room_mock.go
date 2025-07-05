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

func (r *MockRoomRepository) GetUserRooms(userID string, page int, pageSize int) (*[]model.Room, error) {
	args := r.Called(userID, page, pageSize)
	return args.Get(0).(*[]model.Room), args.Error(1)
}

func (r *MockRoomRepository) CountUserRooms(userID string) (int64, error) {
	args := r.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (r *MockRoomRepository) GetUnjoinedRoomsByIsPrivate(userID string, isPrivate bool) (*[]model.Room, error) {
	args := r.Called(userID, isPrivate)
	return args.Get(0).(*[]model.Room), args.Error(1)
}

func (r *MockRoomRepository) GetRoomAttendees(roomID string) (*[]model.User, error) {
	args := r.Called(roomID)
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (r *MockRoomRepository) GetRoomAttendeeIDs(roomID string) (*[]string, error) {
	args := r.Called(roomID)
	return args.Get(0).(*[]string), args.Error(1)
}

func (r *MockRoomRepository) CloseRoom(roomID string) error {
	args := r.Called(roomID)
	return args.Error(0)
}

func (r *MockRoomRepository) UpdateRoom(room *model.Room) error {
	args := r.Called(room)
	return args.Error(0)
}

func (r *MockRoomRepository) AddUserToRoom(roomID string, user *model.User) error {
	args := r.Called(roomID, user)
	return args.Error(0)
}

func (r *MockRoomRepository) RemoveUserFromRoom(roomID, userID string) error {
	args := r.Called(roomID, userID)
	return args.Error(0)
}

func (r *MockRoomRepository) IsUserInRoom(roomID, userID string) (bool, error) {
	args := r.Called(roomID, userID)
	return args.Bool(0), args.Error(1)
}

func (r *MockRoomRepository) GetPendingInvites(userID string) (*[]model.RoomInvite, error) {
	args := r.Called(userID)
	return args.Get(0).(*[]model.RoomInvite), args.Error(1)
}

func (r *MockRoomRepository) CountPendingInvites(userID string) (int64, error) {
	args := r.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (r *MockRoomRepository) UpdateInviteStatus(roomID, userID, status string) error {
	args := r.Called(roomID, userID, status)
	return args.Error(0)
}

func (r *MockRoomRepository) CreateInvites(invites *[]model.RoomInvite) error {
	args := r.Called(invites)
	return args.Error(0)
}

func (r *MockRoomRepository) DeletePendingInvites(roomID string) error {
	args := r.Called(roomID)
	return args.Error(0)
}

func (r *MockRoomRepository) HasPendingInvites(roomID, userID string) (bool, error) {
	args := r.Called(roomID, userID)
	return args.Bool(0), args.Error(1)
}
