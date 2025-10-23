package services

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/dto/response"
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

func (s *UserService) GetUserByID(ctx context.Context, userId string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userId)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.userRepo.FindByUsername(ctx, username)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

func (s *UserService) GetUsersByID(ctx context.Context, userIds []string) ([]model.User, error) {
	return s.userRepo.FindByIDs(ctx, userIds)
}

func (s *UserService) UpdateUsername(ctx context.Context, userId string, newUsername string) error {
	user, err := s.userRepo.FindByID(ctx, userId)
	if err != nil {
		return err
	}

	user.Username = newUsername
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdateUserField(ctx context.Context, userid string, field string, value any) error {
	user, err := s.userRepo.FindByID(ctx, userid)
	if err != nil {
		return err
	}

	switch field {
	case "isEmailValid":
		user.IsEmailValid = value.(bool)
	case "isOnline":
		user.IsOnline = value.(bool)
	case "lastSeen":
		user.LastSeen = value.(time.Time)
	default:
		return ErrUserFieldNotSupported
	}

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpsertUser(ctx context.Context, user *model.User, isCreate bool) (*model.User, error) {
	if isCreate {
		return s.userRepo.Create(ctx, user)
	}

	err := s.userRepo.Update(ctx, user)
	return user, err
}

func (s *UserService) MarkOnline(ctx context.Context, userId string) error {
	return s.UpdateUserField(ctx, userId, "isOnline", true)
}

func (s *UserService) MarkOffline(ctx context.Context, userId string) error {
	user, err := s.userRepo.FindByID(ctx, userId)
	if err != nil {
		return err
	}

	user.IsOnline = false
	user.LastSeen = time.Now()
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) SearchNonFriendUsers(ctx context.Context, currentUserID, query string) ([]response.MinimalUserDto, error) {
	strangers, err := s.userRepo.SearchNonFriendUsers(ctx, currentUserID, query, 10)
	if err != nil {
		return nil, err
	}

	strangerDtos := make([]response.MinimalUserDto, len(strangers))
	for i, stranger := range strangers {
		strangerDtos[i] = response.MinimalUserDto{
			ID:         stranger.ID,
			Username:   stranger.Username,
			PictureUrl: stranger.PictureUrl,
		}
	}

	return strangerDtos, nil
}

func (s *UserService) SendFriendRequest(ctx context.Context, senderID, receiverID uint) error {
	if senderID == receiverID {
		return ErrNoSelfFriendRequest
	}

	// Check if they are already friends
	isFriend, err := s.userRepo.CheckFriendship(ctx, senderID, receiverID)
	if err != nil {
		return err
	} else if isFriend {
		return ErrAlreadyFriends
	}

	// Check if a friend request already exists
	exists, err := s.userRepo.CheckFriendRequestExists(ctx, senderID, receiverID)
	if err != nil {
		return err
	} else if exists {
		return ErrFriendRequestExists
	}

	return database.RunInTransaction(s.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)

		request := model.FriendRequest{
			SenderID:   senderID,
			ReceiverID: receiverID,
			Status:     "pending",
		}

		if err := userRepoTx.CreateFriendRequest(ctx, &request); err != nil {
			return err
		}

		// Increment receiver's pending friend requests count
		return userRepoTx.UpdateNoOfPendingFriendRequests(ctx, []uint{receiverID}, 1)
	})
}

// TODO: Test if addFriend will throw error for non-existing users
func (s *UserService) AcceptFriendRequest(ctx context.Context, requestID uint) error {
	return database.RunInTransaction(s.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)

		request, err := userRepoTx.FindFriendRequest(ctx, requestID)
		if err != nil {
			return err
		}

		if request.Status != "pending" {
			return ErrFriendRequestAlreadyProcessed
		}

		if err := userRepoTx.UpdateFriendRequest(ctx, request.ID, map[string]any{
			"status":       "accepted",
			"responded_at": time.Now(),
		}); err != nil {
			return err
		}

		// Add each user to the other's friend list
		if err := userRepoTx.AddFriend(ctx, request.SenderID, request.ReceiverID); err != nil {
			return err
		}

		// Decrement receiver's pending friend requests count
		if err := userRepoTx.UpdateNoOfPendingFriendRequests(ctx, []uint{request.ReceiverID}, -1); err != nil {
			return err
		}

		// Increment friend count for both users
		return userRepoTx.UpdateNoOfFriends(ctx, []uint{request.SenderID, request.ReceiverID}, 1)
	})
}

func (s *UserService) RejectFriendRequest(ctx context.Context, requestID uint) error {
	return database.RunInTransaction(s.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)

		request, err := userRepoTx.FindFriendRequest(ctx, requestID)
		if err != nil {
			return err
		}

		if request.Status != "pending" {
			return ErrFriendRequestAlreadyProcessed
		}

		if err := userRepoTx.UpdateFriendRequest(ctx, request.ID, map[string]any{
			"status":       "rejected",
			"responded_at": time.Now(),
		}); err != nil {
			return err
		}

		// Decrement receiver's pending friend requests count
		return userRepoTx.UpdateNoOfPendingFriendRequests(ctx, []uint{request.ReceiverID}, -1)
	})
}

func (s *UserService) RemoveFriend(ctx context.Context, userID, friendID uint) error {
	return database.RunInTransaction(s.db, sql.LevelRepeatableRead, func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)

		if err := userRepoTx.RemoveFriend(ctx, userID, friendID); err != nil {
			return err
		}

		// Decrement friend count for both users
		return userRepoTx.UpdateNoOfFriends(ctx, []uint{userID, friendID}, -1)
	})
}

func (s *UserService) GetFriends(ctx context.Context, userID string) ([]response.MinimalUserDto, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, err
	}

	friends, err := s.userRepo.GetFriends(ctx, uint(userIDUint))

	friendDtos := make([]response.MinimalUserDto, len(friends))
	for i, friend := range friends {
		friendDtos[i] = response.MinimalUserDto{
			ID:         friend.ID,
			Username:   friend.Username,
			PictureUrl: friend.PictureUrl,
		}
	}

	return friendDtos, err
}

func (s *UserService) GetFriendRequestsByStatus(ctx context.Context, userID uint, status string) ([]response.FriendRequestDto, error) {
	friendRequests, err := s.userRepo.FindFriendRequestsByReceiver(ctx, userID, status)
	if err != nil {
		return nil, err
	}

	friendRequestsDto := make([]response.FriendRequestDto, len(friendRequests))
	for i, fr := range friendRequests {
		friendRequestsDto[i] = response.FriendRequestDto{
			ID:          fr.ID,
			Status:      fr.Status,
			SentAt:      fr.SentAt,
			RespondedAt: fr.RespondedAt,
			Sender: response.MinimalUserDto{
				ID:         fr.Sender.ID,
				Username:   fr.Sender.Username,
				PictureUrl: fr.Sender.PictureUrl,
			},
			Receiver: response.MinimalUserDto{
				ID:         fr.Receiver.ID,
				Username:   fr.Receiver.Username,
				PictureUrl: fr.Receiver.PictureUrl,
			},
		}
	}

	return friendRequestsDto, nil
}

func (s *UserService) CountPendingFriendRequests(ctx context.Context, userID uint) (int64, error) {
	return s.userRepo.CountPendingFriendRequestsByReceiver(ctx, userID)
}

func (s *UserService) GetNumFriends(ctx context.Context, userID string) (int64, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, err
	}
	return s.userRepo.CountFriends(ctx, uint(userIDUint))
}

func (s *UserService) IsFriend(ctx context.Context, userID uint, friendID uint) bool {
	isFriend, err := s.userRepo.CheckFriendship(ctx, userID, friendID)
	if err != nil {
		return false
	}
	return isFriend
}
