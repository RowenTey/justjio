package services

import (
	"errors"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
)

var (
	ErrUserFieldNotSupported         = errors.New("user field not supported for update")
	ErrNoSelfFriendRequest           = errors.New("cannot send friend request to yourself")
	ErrAlreadyFriends                = errors.New("already friends")
	ErrFriendRequestExists           = errors.New("friend request already sent")
	ErrFriendRequestAlreadyProcessed = errors.New("friend request already processed")
	ErrInvalidFriendRequestStatus    = errors.New("invalid status")
)

type UserService struct {
	db       *gorm.DB
	userRepo repository.UserRepository
	logger   *logrus.Entry
}

func NewUserService(db *gorm.DB, userRepo repository.UserRepository, logger *logrus.Logger) *UserService {
	return &UserService{
		db:       db,
		userRepo: userRepo,
		logger:   utils.AddServiceField(logger, "UserService"),
	}
}

func (s *UserService) GetUserByID(userId string) (*model.User, error) {
	return s.userRepo.FindByID(userId)
}

func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *UserService) GetUsersByID(userIds *[]uint) (*[]model.User, error) {
	return s.userRepo.FindByIDs(userIds)
}

func (s *UserService) UpdateUserField(userid string, field string, value any) error {
	user, err := s.userRepo.FindByID(userid)
	if err != nil {
		return err
	}

	switch field {
	case "username":
		user.Username = value.(string)
	case "isEmailValid":
		user.IsEmailValid = value.(bool)
	case "isOnline":
		user.IsOnline = value.(bool)
	case "lastSeen":
		user.LastSeen = value.(time.Time)
	default:
		return ErrUserFieldNotSupported
	}

	return s.userRepo.Update(user)
}

func (s *UserService) CreateOrUpdateUser(user *model.User, isCreate bool) (*model.User, error) {
	if isCreate {
		return s.userRepo.Create(user)
	}

	err := s.userRepo.Update(user)
	return user, err
}

func (s *UserService) DeleteUser(userId string) error {
	return s.userRepo.Delete(userId)
}

func (s *UserService) ValidateUsers(userIds *[]uint) (*[]model.User, error) {
	return s.userRepo.FindByIDs(userIds)
}

func (s *UserService) MarkOnline(userId string) error {
	return s.UpdateUserField(userId, "isOnline", true)
}

func (s *UserService) MarkOffline(userId string) error {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return err
	}

	user.IsOnline = false
	user.LastSeen = time.Now()
	return s.userRepo.Update(user)
}

func (s *UserService) SearchUsers(currentUserID, query string) (*[]model.User, error) {
	return s.userRepo.SearchUsers(currentUserID, query, 10)
}

func (s *UserService) SendFriendRequest(senderID, receiverID uint) error {
	if senderID == receiverID {
		return ErrNoSelfFriendRequest
	}

	// Check if they are already friends
	if isFriend, err := s.userRepo.CheckFriendship(senderID, receiverID); err != nil {
		return err
	} else if isFriend {
		return ErrAlreadyFriends
	}

	// Check if a friend request already exists
	if exists, err := s.userRepo.CheckFriendRequestExists(senderID, receiverID); err != nil {
		return err
	} else if exists {
		return ErrFriendRequestExists
	}

	request := model.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     "pending",
	}
	return s.userRepo.CreateFriendRequest(&request)
}

// TODO: Test if addFriend will throw error for non-existing users
func (s *UserService) AcceptFriendRequest(requestID uint) error {
	return database.RunInTransaction(s.db, func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)

		request, err := userRepoTx.FindFriendRequest(requestID)
		if err != nil {
			return err
		}

		if request.Status != "pending" {
			return ErrFriendRequestAlreadyProcessed
		}

		request.Status = "accepted"
		request.RespondedAt = time.Now()
		if err := userRepoTx.UpdateFriendRequest(request); err != nil {
			return err
		}

		// Add each user to the other's friend list
		return userRepoTx.AddFriend(request.SenderID, request.ReceiverID)
	})
}

func (s *UserService) RejectFriendRequest(requestID uint) error {
	request, err := s.userRepo.FindFriendRequest(requestID)
	if err != nil {
		return err
	}

	if request.Status != "pending" {
		return ErrFriendRequestAlreadyProcessed
	}

	request.Status = "rejected"
	request.RespondedAt = time.Now()
	return s.userRepo.UpdateFriendRequest(request)
}

func (s *UserService) RemoveFriend(userID, friendID uint) error {
	return s.userRepo.RemoveFriend(userID, friendID)
}

func (s *UserService) GetFriends(userID string) (*[]model.User, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetFriends(uint(userIDUint))
}

func (s *UserService) GetFriendRequestsByStatus(userID uint, status string) (*[]model.FriendRequest, error) {
	// Validate status
	validStatuses := map[string]bool{"pending": true, "accepted": true, "rejected": true}
	if !validStatuses[status] {
		return nil, ErrInvalidFriendRequestStatus
	}
	return s.userRepo.FindFriendRequestsByReceiver(userID, status)
}

func (s *UserService) CountPendingFriendRequests(userID uint) (int64, error) {
	return s.userRepo.CountFriendRequestsByReceiver(userID, "pending")
}

func (s *UserService) GetNumFriends(userID string) (int64, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, err
	}
	return s.userRepo.CountFriends(uint(userIDUint))
}

func (s *UserService) IsFriend(userID uint, friendID uint) bool {
	isFriend, err := s.userRepo.CheckFriendship(userID, friendID)
	if err != nil {
		return false
	}
	return isFriend
}
