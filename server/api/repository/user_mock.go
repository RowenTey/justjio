package repository

import (
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

func (m *MockUserRepository) Create(user *model.User) (*model.User, error) {
	args := m.Called(user)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id string) (*model.User, error) {
	args := m.Called(id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDs(ids *[]uint) (*[]model.User, error) {
	args := m.Called(ids)
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateFriendRequest(request *model.FriendRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockUserRepository) FindFriendRequest(id uint) (*model.FriendRequest, error) {
	args := m.Called(id)
	return args.Get(0).(*model.FriendRequest), args.Error(1)
}

func (m *MockUserRepository) UpdateFriendRequest(request *model.FriendRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockUserRepository) FindFriendRequestsByReceiver(receiverID uint, status string) (*[]model.FriendRequest, error) {
	args := m.Called(receiverID, status)
	return args.Get(0).(*[]model.FriendRequest), args.Error(1)
}

func (m *MockUserRepository) CountFriendRequestsByReceiver(receiverID uint, status string) (int64, error) {
	args := m.Called(receiverID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CheckFriendRequestExists(senderID, receiverID uint) (bool, error) {
	args := m.Called(senderID, receiverID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) AddFriend(userID, friendID uint) error {
	args := m.Called(userID, friendID)
	return args.Error(0)
}

func (m *MockUserRepository) RemoveFriend(userID, friendID uint) error {
	args := m.Called(userID, friendID)
	return args.Error(0)
}

func (m *MockUserRepository) GetFriends(userID uint) (*[]model.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (m *MockUserRepository) CountFriends(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CheckFriendship(userID, friendID uint) (bool, error) {
	args := m.Called(userID, friendID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUninvitedFriends(roomID, userID string) (*[]model.User, error) {
	args := m.Called(roomID, userID)
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (m *MockUserRepository) SearchUsers(currentUserId, query string, limit int) (*[]model.User, error) {
	args := m.Called(currentUserId, query, limit)
	return args.Get(0).(*[]model.User), args.Error(1)
}
