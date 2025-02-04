package services

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/RowenTey/JustJio/model"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
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
		log.Println("Error marking user online:", err)
		return err
	}
	return nil
}

func (s *UserService) MarkOffline(userId string) error {
	if err := s.UpdateUserField(userId, "isOnline", false); err != nil {
		log.Println("Error marking user offline:", err)
		return err
	}

	if err := s.UpdateUserField(userId, "lastSeen", time.Now()); err != nil {
		log.Println("Error updating last seen:", err)
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

func (s *UserService) AddFriend(userID string, friendID string) error {
	db := s.DB
	var user, friend model.User

	if err := db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := db.First(&friend, friendID).Error; err != nil {
		return err
	}

	if userID == friendID {
		return errors.New("Cannot add self as friend")
	}

	return db.Model(&user).Association("Friends").Append(&friend)
}

func (s *UserService) RemoveFriend(userID string, friendID string) error {
	db := s.DB
	var user, friend model.User

	if err := db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := db.First(&friend, friendID).Error; err != nil {
		return err
	}

	return db.Model(&user).Association("Friends").Delete(&friend)
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

func (s *UserService) GetNumFriends(userID string) (int64, error) {
	db := s.DB
	var user model.User

	if err := db.First(&user, userID).Error; err != nil {
		return 0, err
	}

	return db.Model(&user).Association("Friends").Count(), nil
}

func (s *UserService) IsFriend(userID string, friendID string) (bool, error) {
	db := s.DB

	var user, friend model.User

	if err := db.First(&user, userID).Error; err != nil {
		return false, err
	}

	if err := db.First(&friend, friendID).Error; err != nil {
		return false, err
	}

	if err := db.
		Model(&user).
		Where("id = ?", friendID).
		Association("Friends").
		Find(&friend); err != nil {
		return false, err
	}

	return true, nil
}
