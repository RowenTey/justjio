package repository

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) WithTx(tx *gorm.DB) UserRepository {
	args := m.Called(tx)
	return args.Get(0).(UserRepository)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDs(ctx context.Context, ids []string) ([]model.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Friends relationships
func (m *MockUserRepository) CreateFriendRequest(ctx context.Context, request *model.FriendRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockUserRepository) FindFriendRequest(ctx context.Context, id uint) (*model.FriendRequest, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.FriendRequest), args.Error(1)
}

func (m *MockUserRepository) UpdateFriendRequest(ctx context.Context, requestID uint, values any) error {
	args := m.Called(ctx, requestID, values)
	return args.Error(0)
}

func (m *MockUserRepository) FindFriendRequestsByReceiver(ctx context.Context, receiverID uint, status string) ([]model.FriendRequest, error) {
	args := m.Called(ctx, receiverID, status)
	return args.Get(0).([]model.FriendRequest), args.Error(1)
}

func (m *MockUserRepository) CountPendingFriendRequestsByReceiver(ctx context.Context, receiverID uint) (int64, error) {
	args := m.Called(ctx, receiverID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CheckFriendRequestExists(ctx context.Context, senderID, receiverID uint) (bool, error) {
	args := m.Called(ctx, senderID, receiverID)
	return args.Bool(0), args.Error(1)
}

// Friends operations
func (m *MockUserRepository) AddFriend(ctx context.Context, userID, friendID uint) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockUserRepository) RemoveFriend(ctx context.Context, userID, friendID uint) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockUserRepository) GetFriends(ctx context.Context, userID uint) ([]model.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) CountFriends(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CheckFriendship(ctx context.Context, userID, friendID uint) (bool, error) {
	args := m.Called(ctx, userID, friendID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUninvitedFriends(ctx context.Context, roomID, userID string) ([]model.User, error) {
	args := m.Called(ctx, roomID, userID)
	return args.Get(0).([]model.User), args.Error(1)
}

// Search
func (m *MockUserRepository) SearchNonFriendUsers(ctx context.Context, currentUserId, query string, limit int) ([]model.User, error) {
	args := m.Called(ctx, currentUserId, query, limit)
	return args.Get(0).([]model.User), args.Error(1)
}

// Pending invites
func (m *MockUserRepository) UpdateNoOfPendingRoomInvites(ctx context.Context, userIDs []string, delta int) error {
	args := m.Called(ctx, userIDs, delta)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateNoOfPendingFriendRequests(ctx context.Context, userIDs []uint, delta int) error {
	args := m.Called(ctx, userIDs, delta)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateNoOfFriends(ctx context.Context, userIDs []uint, delta int) error {
	args := m.Called(ctx, userIDs, delta)
	return args.Error(0)
}
