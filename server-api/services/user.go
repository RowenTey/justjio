package services

import (
	"errors"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/model"

	"gorm.io/gorm"
)

type UserService struct {
	DB     *gorm.DB
	logger *log.Entry
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		DB:     db,
		logger: log.WithFields(log.Fields{"service": "UserService"}),
	}
}

func (s *UserService) GetUserByID(userId string) (*model.User, error) {
	db := s.DB
	var user model.User
	if err := db.First(&user, userId).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	db := s.DB.Table("users")
	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	db := s.DB.Table("users")
	var user model.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUsersByID(userIds []uint) (*[]model.User, error) {
	db := s.DB.Table("users")
	var users []model.User
	if err := db.Where("id IN ?", userIds).Find(&users).Error; err != nil {
		return nil, err
	}
	return &users, nil
}

func (s *UserService) UpdateUserField(userid string, field string, value interface{}) error {
	db := s.DB
	var user model.User

	if err := db.First(&user, userid).Error; err != nil {
		return err
	}

	switch field {
	// case "name":
	// 	user.Name = value.(string)
	// case "phoneNum":
	// 	user.PhoneNum = value.(string)
	case "username":
		user.Username = value.(string)
	case "isEmailValid":
		user.IsEmailValid = value.(bool)
	case "isOnline":
		user.IsOnline = value.(bool)
	case "lastSeen":
		user.LastSeen = value.(time.Time)
	default:
		return errors.New("User field (" + field + ") not supported for update")
	}

	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func (s *UserService) CreateOrUpdateUser(user *model.User, isCreate bool) (*model.User, error) {
	db := s.DB.Table("users")

	if isCreate {
		user.RegisteredAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	if err := db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) DeleteUser(userId string) error {
	db := s.DB.Table("users")
	var user model.User

	userIDUint, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		return err
	}
	user.ID = uint(userIDUint)

	if err := db.Delete(&user).Error; err != nil {
		return err
	}

	return nil
}

func (s *UserService) ValidateUsers(userIds []string) (*[]model.User, error) {
	if len(userIds) == 0 {
		return &[]model.User{}, nil
	}

	db := s.DB.Table("users")
	var users []model.User

	if err := db.Find(&users, userIds).Error; err != nil {
		return nil, err
	}

	return &users, nil
}

func (s *UserService) MarkOnline(userId string) error {
	if err := s.UpdateUserField(userId, "isOnline", true); err != nil {
		return err
	}
	return nil
}

func (s *UserService) MarkOffline(userId string) error {
	if err := s.UpdateUserField(userId, "isOnline", false); err != nil {
		return err
	}

	if err := s.UpdateUserField(userId, "lastSeen", time.Now()); err != nil {
		return err
	}
	return nil
}

func (s *UserService) SearchUsers(currentUserID, query string) (*[]model.User, error) {
	db := s.DB
	var users []model.User

	// Use LEFT JOIN to exclude friends
	if err := db.
		Table("users").
		Joins("LEFT JOIN user_friends ON users.id = user_friends.friend_id AND user_friends.user_id = ?", currentUserID).
		Where("users.username LIKE ?", "%"+query+"%").
		Where("user_friends.friend_id IS NULL").
		Where("users.id != ?", currentUserID).
		Limit(10).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return &users, nil
}

func (s *UserService) SendFriendRequest(senderID, receiverID uint) error {
	db := s.DB

	if senderID == receiverID {
		return errors.New("cannot send friend request to yourself")
	}

	// Check if they are already friends
	var sender model.User
	if err := db.Preload("Friends").First(&sender, senderID).Error; err != nil {
		return errors.New("sender not found")
	}
	for _, friend := range sender.Friends {
		if friend.ID == receiverID {
			return errors.New("already friends")
		}
	}

	// Check if a friend request already exists (either sent by userA or userB)
	var existing model.FriendRequest
	if err := db.Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?", senderID, receiverID, receiverID, senderID, "pending").First(&existing).Error; err == nil {
		return errors.New("friend request already sent")
	}

	// Create new friend request
	request := model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     "pending",
	}
	return db.Create(&request).Error
}

func (s *UserService) AcceptFriendRequest(requestID uint) error {
	db := s.DB
	var request model.FriendRequest

	if err := db.First(&request, requestID).Error; err != nil {
		return err
	}

	if request.Status != "pending" {
		return errors.New("friend request already processed")
	}

	request.Status = "accepted"
	request.RespondedAt = time.Now()
	if err := db.Save(&request).Error; err != nil {
		return err
	}

	// Add each user to the other's friend list
	var sender, receiver model.User
	if err := db.First(&sender, request.SenderID).Error; err != nil {
		return err
	}
	if err := db.First(&receiver, request.ReceiverID).Error; err != nil {
		return err
	}

	if err := db.Model(&sender).Association("Friends").Append(&receiver); err != nil {
		return err
	}
	if err := db.Model(&receiver).Association("Friends").Append(&sender); err != nil {
		return err
	}

	return nil
}

func (s *UserService) RejectFriendRequest(requestID uint) error {
	db := s.DB
	var request model.FriendRequest

	if err := db.First(&request, requestID).Error; err != nil {
		return err
	}

	if request.Status != "pending" {
		return errors.New("friend request already processed")
	}

	request.Status = "rejected"
	request.RespondedAt = time.Now()
	return db.Save(&request).Error
}

func (s *UserService) RemoveFriend(userID, friendID uint) error {
	db := s.DB
	var user, friend model.User

	if err := db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := db.First(&friend, friendID).Error; err != nil {
		return err
	}

	if err := db.Model(&user).Association("Friends").Delete(&friend); err != nil {
		return err
	}

	if err := db.Model(&friend).Association("Friends").Delete(&user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetFriends(userID string) ([]model.User, error) {
	db := s.DB

	var user model.User
	var friends []model.User

	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&user).Association("Friends").Find(&friends); err != nil {
		return nil, err
	}

	return friends, nil
}

func (s *UserService) GetFriendRequestsByStatus(userID uint, status string) (*[]model.FriendRequest, error) {
	db := s.DB
	var requests []model.FriendRequest

	// Validate status
	validStatuses := map[string]bool{"pending": true, "accepted": true, "rejected": true}
	if !validStatuses[status] {
		return nil, errors.New("invalid status")
	}

	// Fetch friend requests where the user is the receiver
	if err := db.Where("receiver_id = ? AND status = ?", userID, status).
		Preload("Sender").
		Preload("Receiver").
		Find(&requests).Error; err != nil {
		return nil, err
	}

	return &requests, nil
}

func (s *UserService) CountPendingFriendRequests(userID uint) (int64, error) {
	db := s.DB
	var count int64

	// Count pending friend requests where the user is the receiver
	if err := db.Model(&model.FriendRequest{}).
		Where("receiver_id = ? AND status = ?", userID, "pending").
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (s *UserService) GetNumFriends(userID string) (int64, error) {
	db := s.DB
	var user model.User

	if err := db.First(&user, userID).Error; err != nil {
		return 0, err
	}

	return db.Model(&user).Association("Friends").Count(), nil
}

func (s *UserService) IsFriend(userID uint, friendID uint) bool {
	db := s.DB

	var user, friend model.User
	if err := db.First(&user, userID).Error; err != nil {
		return false
	}

	if err := db.First(&friend, friendID).Error; err != nil {
		return false
	}

	if err := db.
		Model(&user).
		Where("id = ?", friendID).
		Association("Friends").
		Find(&friend); err != nil {
		return false
	}

	return true
}
