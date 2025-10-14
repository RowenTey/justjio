package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	WithTx(tx *gorm.DB) UserRepository

	Create(user *model.User) (*model.User, error)
	FindByID(id string) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByIDs(ids []string) ([]model.User, error)
	Update(user *model.User) error
	Delete(id string) error

	// Friends relationships
	CreateFriendRequest(request *model.FriendRequest) error
	FindFriendRequest(id uint) (*model.FriendRequest, error)
	UpdateFriendRequest(requestID uint, values any) error
	FindFriendRequestsByReceiver(receiverID uint, status string) ([]model.FriendRequest, error)
	CountFriendRequestsByReceiver(receiverID uint, status string) (int64, error)
	CheckFriendRequestExists(senderID, receiverID uint) (bool, error)

	// Friends operations
	AddFriend(userID, friendID uint) error
	RemoveFriend(userID, friendID uint) error
	GetFriends(userID uint) ([]model.User, error)
	CountFriends(userID uint) (int64, error)
	CheckFriendship(userID, friendID uint) (bool, error)
	GetUninvitedFriends(roomID, userID string) ([]model.User, error)

	// Search
	SearchNonFriendUsers(currentUserId, query string, limit int) ([]model.User, error)

	UpdateNoOfPendingRoomInvites(userIDs []string, delta int) error
	UpdateNoOfPendingFriendRequests(userIDs []uint, delta int) error
	UpdateNoOfFriends(userIDs []uint, delta int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// WithTx returns a new UserRepository with the provided transaction
func (r *userRepository) WithTx(tx *gorm.DB) UserRepository {
	if tx == nil {
		return r
	}
	return &userRepository{db: tx}
}

// Create inserts a new user into the database.
func (r *userRepository) Create(user *model.User) (*model.User, error) {
	err := r.db.Create(user).Error
	return user, err
}

// FindByID retrieves a user by their ID.
func (r *userRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

// FindByUsername retrieves a user by their username.
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// FindByEmail retrieves a user by their email address.
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// FindByIDs retrieves users by their IDs.
func (r *userRepository) FindByIDs(ids []string) ([]model.User, error) {
	if len(ids) == 0 {
		return []model.User{}, nil
	}

	var users []model.User
	err := r.db.Find(&users, ids).Error
	if len(users) != len(ids) {
		return nil, gorm.ErrRecordNotFound
	}

	return users, err
}

// Update modifies an existing user.
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete removes a user by ID.
func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&model.User{}, id).Error
}

// CreateFriendRequest creates a new friend request.
func (r *userRepository) CreateFriendRequest(request *model.FriendRequest) error {
	return r.db.Create(request).Error
}

// FindFriendRequest retrieves a friend request by its ID.
func (r *userRepository) FindFriendRequest(id uint) (*model.FriendRequest, error) {
	var request model.FriendRequest
	err := r.db.First(&request, id).Error
	return &request, err
}

// UpdateFriendRequest updates an existing friend request.
func (r *userRepository) UpdateFriendRequest(requestID uint, values any) error {
	return r.db.Model(model.FriendRequest{ID: requestID}).Updates(values).Error
}

// FindFriendRequestsByReceiver retrieves friend requests for a specific receiver with a given status.
func (r *userRepository) FindFriendRequestsByReceiver(receiverID uint, status string) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := r.db.
		Where("receiver_id = ? AND status = ?", receiverID, status).
		Joins("Sender").
		Joins("Receiver").
		Find(&requests).Error
	return requests, err
}

// CountFriendRequestsByReceiver counts the number of friend requests for a specific receiver with a given status.
func (r *userRepository) CountFriendRequestsByReceiver(receiverID uint, status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.FriendRequest{}).
		Where("receiver_id = ? AND status = ?", receiverID, status).
		Count(&count).Error
	return count, err
}

// CheckFriendRequestExists checks if a friend request exists between two users.
func (r *userRepository) CheckFriendRequestExists(senderID, receiverID uint) (bool, error) {
	var existing model.FriendRequest
	err := r.db.
		Where(
			"((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?",
			senderID,
			receiverID,
			receiverID,
			senderID,
			"pending",
		).
		First(&existing).Error
	if err == nil {
		return existing.ID != 0, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

// AddFriend adds a friend relationship between two users.
func (r *userRepository) AddFriend(userID, friendID uint) error {
	return r.db.Exec("INSERT INTO user_friends (user_id, friend_id) VALUES (?, ?), (?, ?)", userID, friendID, friendID, userID).Error
}

// RemoveFriend removes a friend from a user's friend list.
func (r *userRepository) RemoveFriend(userID, friendID uint) error {
	result := r.db.Exec("DELETE FROM user_friends WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetFriends retrieves the friends of a user.
func (r *userRepository) GetFriends(userID uint) ([]model.User, error) {
	var friends []model.User
	err := r.db.
		Model(model.User{ID: userID}).
		Association("Friends").
		Find(&friends)
	return friends, err
}

// CountFriends returns the number of friends a user has.
func (r *userRepository) CountFriends(userID uint) (int64, error) {
	var user model.User
	err := r.db.First(&user, userID).Error
	return int64(user.NoOfFriends), err
}

// CheckFriendship checks if a user is friends with another user.
func (r *userRepository) CheckFriendship(userID, friendID uint) (bool, error) {
	count := r.db.
		Model(&model.User{ID: userID}).
		Where("id = ?", friendID).
		Association("Friends").
		Count()
	return count > 0, nil
}

// SearchNonFriendUsers retrieves users based on a search query, excluding the current user and their friends.
func (r *userRepository) SearchNonFriendUsers(currentUserId, query string, limit int) ([]model.User, error) {
	// Sanitize the query to prevent SQL injection in tsquery
	// Remove special characters that could break tsquery syntax
	sanitizedQuery := ""
	for _, char := range query {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == ' ' {
			sanitizedQuery += string(char)
		}
	}

	// Trim whitespace and ensure query is not empty
	sanitizedQuery = strings.TrimSpace(sanitizedQuery)
	if sanitizedQuery == "" {
		return []model.User{}, nil
	}

	tsQuery := fmt.Sprintf("%s:*", sanitizedQuery)
	var users []model.User
	if err := r.db.
		Table("users").
		Joins("JOIN user_non_friends ON users.id = user_non_friends.non_friend_id AND user_non_friends.user_id = ?", currentUserId).
		Where("users.search_vector @@ to_tsquery('english', ?)", tsQuery).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUninvitedFriends retrieves friends of a user who are not invited to a specific room.
func (r *userRepository) GetUninvitedFriends(roomID, userID string) ([]model.User, error) {
	var friends []model.User
	err := r.db.
		Table("users u").
		Select("u.id, u.username, u.picture_url").
		Joins("JOIN user_friends uf ON uf.friend_id = u.id AND uf.user_id = ?", userID).
		Where("NOT EXISTS (SELECT 1 FROM room_users ru WHERE ru.user_id = u.id AND ru.room_id = ?)", roomID).
		Where("NOT EXISTS (SELECT 1 FROM room_invites ri WHERE ri.user_id = u.id AND ri.room_id = ? AND ri.status = 'pending')", roomID).
		Find(&friends).Error
	return friends, err
}

func (r *userRepository) UpdateNoOfPendingRoomInvites(userIDs []string, delta int) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.
		Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("no_of_pending_room_invites", gorm.Expr("no_of_pending_room_invites + ?", delta)).Error
}

func (r *userRepository) UpdateNoOfPendingFriendRequests(userIDs []uint, delta int) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.
		Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("no_of_pending_friend_requests", gorm.Expr("no_of_pending_friend_requests + ?", delta)).Error
}

func (r *userRepository) UpdateNoOfFriends(userIDs []uint, delta int) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.
		Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("no_of_friends", gorm.Expr("no_of_friends + ?", delta)).Error
}
