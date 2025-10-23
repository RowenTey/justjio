package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	WithTx(tx *gorm.DB) UserRepository

	Create(ctx context.Context, user *model.User) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByIDs(ctx context.Context, ids []string) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error

	// Friends relationships
	CreateFriendRequest(ctx context.Context, request *model.FriendRequest) error
	FindFriendRequest(ctx context.Context, id uint) (*model.FriendRequest, error)
	UpdateFriendRequest(ctx context.Context, requestID uint, values any) error
	FindFriendRequestsByReceiver(ctx context.Context, receiverID uint, status string) ([]model.FriendRequest, error)
	CountPendingFriendRequestsByReceiver(ctx context.Context, receiverID uint) (int64, error)
	CheckFriendRequestExists(ctx context.Context, senderID, receiverID uint) (bool, error)

	// Friends operations
	AddFriend(ctx context.Context, userID, friendID uint) error
	RemoveFriend(ctx context.Context, userID, friendID uint) error
	GetFriends(ctx context.Context, userID uint) ([]model.User, error)
	CountFriends(ctx context.Context, userID uint) (int64, error)
	CheckFriendship(ctx context.Context, userID, friendID uint) (bool, error)
	GetUninvitedFriends(ctx context.Context, roomID, userID string) ([]model.User, error)

	// Search
	SearchNonFriendUsers(ctx context.Context, currentUserId, query string, limit int) ([]model.User, error)

	UpdateNoOfPendingRoomInvites(ctx context.Context, userIDs []string, delta int) error
	UpdateNoOfPendingFriendRequests(ctx context.Context, userIDs []uint, delta int) error
	UpdateNoOfFriends(ctx context.Context, userIDs []uint, delta int) error
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
func (r *userRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	return user, err
}

// FindByID retrieves a user by their ID.
func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

// FindByUsername retrieves a user by their username.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

// FindByEmail retrieves a user by their email address.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

// FindByIDs retrieves users by their IDs.
func (r *userRepository) FindByIDs(ctx context.Context, ids []string) ([]model.User, error) {
	if len(ids) == 0 {
		return []model.User{}, nil
	}

	var users []model.User
	err := r.db.WithContext(ctx).Find(&users, ids).Error
	if len(users) != len(ids) {
		return nil, gorm.ErrRecordNotFound
	}

	return users, err
}

// Update modifies an existing user.
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete removes a user by ID.
func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// CreateFriendRequest creates a new friend request.
func (r *userRepository) CreateFriendRequest(ctx context.Context, request *model.FriendRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// FindFriendRequest retrieves a friend request by its ID.
func (r *userRepository) FindFriendRequest(ctx context.Context, id uint) (*model.FriendRequest, error) {
	var request model.FriendRequest
	err := r.db.WithContext(ctx).First(&request, id).Error
	return &request, err
}

// UpdateFriendRequest updates an existing friend request.
func (r *userRepository) UpdateFriendRequest(ctx context.Context, requestID uint, values any) error {
	return r.db.WithContext(ctx).Model(model.FriendRequest{ID: requestID}).Updates(values).Error
}

// FindFriendRequestsByReceiver retrieves friend requests for a specific receiver with a given status.
func (r *userRepository) FindFriendRequestsByReceiver(ctx context.Context, receiverID uint, status string) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := r.db.WithContext(ctx).
		Preload("Sender", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Preload("Receiver", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Where("receiver_id = ? AND status = ?", receiverID, status).
		Find(&requests).Error
	return requests, err
}

// CountPendingFriendRequestsByReceiver counts the number of friend requests for a specific receiver with a given status.
func (r *userRepository) CountPendingFriendRequestsByReceiver(ctx context.Context, receiverID uint) (int64, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, receiverID).Error
	return int64(user.NoOfPendingFriendRequests), err
}

// CheckFriendRequestExists checks if a friend request exists between two users.
func (r *userRepository) CheckFriendRequestExists(ctx context.Context, senderID, receiverID uint) (bool, error) {
	var existing model.FriendRequest
	err := r.db.WithContext(ctx).
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
func (r *userRepository) AddFriend(ctx context.Context, userID, friendID uint) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO user_friends (user_id, friend_id) VALUES (?, ?), (?, ?)", userID, friendID, friendID, userID).Error
}

// RemoveFriend removes a friend from a user's friend list.
func (r *userRepository) RemoveFriend(ctx context.Context, userID, friendID uint) error {
	result := r.db.WithContext(ctx).Exec("DELETE FROM user_friends WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetFriends retrieves the friends of a user.
func (r *userRepository) GetFriends(ctx context.Context, userID uint) ([]model.User, error) {
	var friends []model.User
	err := r.db.WithContext(ctx).
		Model(model.User{ID: userID}).
		Association("Friends").
		Find(&friends)
	return friends, err
}

// CountFriends returns the number of friends a user has.
func (r *userRepository) CountFriends(ctx context.Context, userID uint) (int64, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	return int64(user.NoOfFriends), err
}

// CheckFriendship checks if a user is friends with another user.
func (r *userRepository) CheckFriendship(ctx context.Context, userID, friendID uint) (bool, error) {
	count := r.db.WithContext(ctx).
		Model(&model.User{ID: userID}).
		Where("id = ?", friendID).
		Association("Friends").
		Count()
	return count > 0, nil
}

// SearchNonFriendUsers retrieves users based on a search query, excluding the current user and their friends.
func (r *userRepository) SearchNonFriendUsers(ctx context.Context, currentUserId, query string, limit int) ([]model.User, error) {
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
	if err := r.db.WithContext(ctx).
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
func (r *userRepository) GetUninvitedFriends(ctx context.Context, roomID, userID string) ([]model.User, error) {
	var friends []model.User
	err := r.db.WithContext(ctx).
		Table("users u").
		Select("u.id, u.username, u.picture_url").
		Joins("JOIN user_friends uf ON uf.friend_id = u.id AND uf.user_id = ?", userID).
		Where("NOT EXISTS (SELECT 1 FROM room_users ru WHERE ru.user_id = u.id AND ru.room_id = ?)", roomID).
		Where("NOT EXISTS (SELECT 1 FROM room_invites ri WHERE ri.user_id = u.id AND ri.room_id = ? AND ri.status = 'pending')", roomID).
		Find(&friends).Error
	return friends, err
}

func (r *userRepository) UpdateNoOfPendingRoomInvites(ctx context.Context, userIDs []string, delta int) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("no_of_pending_room_invites", gorm.Expr("no_of_pending_room_invites + ?", delta)).Error
}

func (r *userRepository) UpdateNoOfPendingFriendRequests(ctx context.Context, userIDs []uint, delta int) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("no_of_pending_friend_requests", gorm.Expr("no_of_pending_friend_requests + ?", delta)).Error
}

func (r *userRepository) UpdateNoOfFriends(ctx context.Context, userIDs []uint, delta int) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id IN ?", userIDs).
		Update("no_of_friends", gorm.Expr("no_of_friends + ?", delta)).Error
}
