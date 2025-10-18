package repository

import (
	"context"
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

func (m *MockRoomRepository) Create(ctx context.Context, room *model.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetByID(ctx context.Context, roomID string) (*model.Room, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetByIDWithAttendees(ctx context.Context, roomID string) (*model.Room, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetUserRooms(ctx context.Context, userID string, page int, pageSize int) ([]model.Room, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).([]model.Room), args.Error(1)
}

func (m *MockRoomRepository) CountUserRooms(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoomRepository) GetUnjoinedRoomsByIsPrivate(ctx context.Context, userID string, isPrivate bool) ([]model.Room, error) {
	args := m.Called(ctx, userID, isPrivate)
	return args.Get(0).([]model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetRoomAttendeeIDs(ctx context.Context, roomID string) ([]string, error) {
	args := m.Called(ctx, roomID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoomRepository) CloseRoom(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

func (m *MockRoomRepository) Update(ctx context.Context, room *model.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) AddUserToRoom(ctx context.Context, roomID string, user *model.User) error {
	args := m.Called(ctx, roomID, user)
	return args.Error(0)
}

func (m *MockRoomRepository) RemoveUserFromRoom(ctx context.Context, roomID, userID string) error {
	args := m.Called(ctx, roomID, userID)
	return args.Error(0)
}

func (m *MockRoomRepository) IsUserInRoom(ctx context.Context, roomID, userID string) (bool, error) {
	args := m.Called(ctx, roomID, userID)
	return args.Bool(0), args.Error(1)
}

// Invite related methods
func (m *MockRoomRepository) GetPendingInvites(ctx context.Context, userID string) ([]model.RoomInvite, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.RoomInvite), args.Error(1)
}

func (m *MockRoomRepository) CountPendingInvites(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoomRepository) UpdateInviteStatus(ctx context.Context, roomID, userID, status string) error {
	args := m.Called(ctx, roomID, userID, status)
	return args.Error(0)
}

func (m *MockRoomRepository) CreateInvites(ctx context.Context, invites []model.RoomInvite) error {
	args := m.Called(ctx, invites)
	return args.Error(0)
}

func (m *MockRoomRepository) DeletePendingInvites(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

func (m *MockRoomRepository) HasPendingInvites(ctx context.Context, roomID, userID string) (bool, error) {
	args := m.Called(ctx, roomID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoomRepository) GetPendingInviteUsers(ctx context.Context, roomID string) ([]string, error) {
	args := m.Called(ctx, roomID)
	return args.Get(0).([]string), args.Error(1)
}
