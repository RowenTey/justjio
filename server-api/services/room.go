package services

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/RowenTey/JustJio/database"
	"github.com/RowenTey/JustJio/model"

	"gorm.io/gorm"
)

const (
	ROOM_PAGE_SIZE = 6
)

type RoomService struct {
	DB *gorm.DB
}

func (rs *RoomService) CreateRoom(room *model.Room, host *model.User) (*model.Room, error) {
	db := rs.DB.Table("rooms")

	room.HostID = host.ID
	room.Users = append(room.Users, *host)
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	if err := db.Create(&room).Error; err != nil {
		return nil, err
	}

	log.Println("[ROOM] Created room with ID: ", room.ID)
	return room, nil
}

func (rs *RoomService) GetRooms(userId string, page int) (*[]model.Room, error) {
	db := rs.DB
	var rooms []model.Room

	if err := db.
		Model(&model.Room{}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("room_users.user_id = ?", userId).
		Where("rooms.is_closed = ?", false).
		Order("rooms.updated_at DESC").
		Scopes(database.Paginate(page, ROOM_PAGE_SIZE)).
		Find(&rooms).Error; err != nil {
		return nil, err
	}

	return &rooms, nil
}

func (rs *RoomService) GetNumRooms(userId string) (int64, error) {
	db := rs.DB
	var count int64

	if err := db.
		Model(&model.Room{}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("room_users.user_id = ?", userId).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (rs *RoomService) GetRoomById(roomId string) (*model.Room, error) {
	db := rs.DB.Table("rooms")
	var room model.Room

	if err := db.First(&room, "id = ?", roomId).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (rs *RoomService) GetRoomInvites(userId string) (*[]model.RoomInvite, error) {
	db := rs.DB.Table("room_invites")
	var invites []model.RoomInvite

	if err := db.
		Preload("Room.Host").
		Preload("User").
		Preload("Inviter").
		Where("user_id = ? AND status = ?", userId, "pending").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return &invites, nil
}

func (rs *RoomService) GetNumRoomInvites(userId string) (int64, error) {
	db := rs.DB.Table("room_invites")
	var count int64

	if err := db.
		Where("user_id = ? AND status = ?", userId, "pending").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (rs *RoomService) GetRoomAttendees(roomId string) (*[]model.User, error) {
	db := rs.DB.Table("rooms")
	var room model.Room

	if err := db.Preload("Users").First(&room, "id = ?", roomId).Error; err != nil {
		return nil, err
	}
	return &room.Users, nil
}

func (rs *RoomService) GetRoomAttendeesIds(roomId string) (*[]string, error) {
	db := rs.DB.Table("rooms")
	var room model.Room

	if err := db.Preload("Users").First(&room, "id = ?", roomId).Error; err != nil {
		return nil, err
	}

	var userIds []string
	for _, user := range room.Users {
		userIds = append(userIds, strconv.FormatUint(uint64(user.ID), 10))
	}

	return &userIds, nil
}

func (rs *RoomService) CloseRoom(roomId string, userId string) error {
	db := rs.DB
	var room model.Room

	userIdUint, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		return err
	}

	if err := db.First(&room, "id = ?", roomId).Error; err != nil {
		return err
	}

	if room.HostID != uint(userIdUint) {
		return errors.New("user is not the host of the room")
	}

	room.IsClosed = true
	room.UpdatedAt = time.Now()
	if err := db.Save(&room).Error; err != nil {
		return err
	}

	// Remove all pending invites
	if err := db.
		Where("room_id = ? AND status = ?", roomId, "pending").
		Delete(&model.RoomInvite{}).Error; err != nil {
		return err
	}

	return nil
}

func (rs *RoomService) UpdateRoomInviteStatus(roomId string, userId string, status string) error {
	if status != "accepted" && status != "rejected" {
		return errors.New("invalid status")
	}

	db := rs.DB

	if err := db.
		Model(&model.RoomInvite{}).
		Where("room_id = ? AND user_id = ?", roomId, userId).
		Update("status", status).Error; err != nil {
		return err
	}

	if status == "rejected" {
		return nil
	}

	var user model.User
	var room model.Room
	if err := db.First(&room, "id = ?", roomId).Error; err != nil {
		return err
	}
	if err := db.First(&user, userId).Error; err != nil {
		return err
	}

	// Update room info
	room.AttendeesCount++
	room.Users = append(room.Users, user)
	room.UpdatedAt = time.Now()

	if err := db.Save(&room).Error; err != nil {
		return err
	}

	return nil
}

func (rs *RoomService) JoinRoom(roomId, userId string) error {
	db := rs.DB
	var room model.Room
	var user model.User

	if err := db.First(&room, "id = ?", roomId).Error; err != nil {
		return err
	}
	if err := db.First(&user, userId).Error; err != nil {
		return err
	}

	// Check if user is already in room
	var count int64
	if err := db.
		Model(&model.Room{}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("rooms.id = ? AND room_users.user_id = ?", roomId, userId).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user is already in room")
	}

	// Update room info
	room.AttendeesCount++
	room.Users = append(room.Users, user)
	room.UpdatedAt = time.Now()

	if err := db.Save(&room).Error; err != nil {
		return err
	}

	return nil
}

func (rs *RoomService) InviteUserToRoom(
	roomId string,
	inviter *model.User,
	users *[]model.User,
	message string,
) (*[]model.RoomInvite, error) {
	if len(*users) == 0 {
		return &[]model.RoomInvite{}, nil
	}

	log.Printf("[ROOM] Inviting users (%v) to room %s", users, roomId)
	var room model.Room
	var roomInvites []model.RoomInvite

	if err := rs.DB.Table("rooms").First(&room, "id = ?", roomId).Error; err != nil {
		return nil, err
	}

	if room.HostID != inviter.ID {
		return nil, errors.New("user is not the host of the room")
	}

	// Check if users are already in room or have pending invites
	for _, user := range *users {
		var count int64

		// Check if user is already in room
		err := rs.DB.
			Model(&model.Room{}).
			Joins("JOIN room_users ON rooms.id = room_users.room_id").
			Where("rooms.id = ? AND room_users.user_id = ?", roomId, user.ID).
			Count(&count).Error
		if err != nil || count > 0 {
			return nil, errors.New("user is already in room")
		}

		// Check if user has pending invite
		if err := rs.DB.Table("room_invites").
			Where("room_id = ? AND user_id = ? AND status = ?", roomId, user.ID, "pending").
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("user already has pending invite")
		}
	}

	for _, user := range *users {
		roomInvite := model.RoomInvite{
			Room:      room,
			User:      user,
			UserID:    user.ID,
			InviterID: inviter.ID,
			Inviter:   *inviter,
			Message:   message,
			CreatedAt: time.Now(),
			Status:    "pending",
		}

		roomInvites = append(roomInvites, roomInvite)
	}

	if err := rs.DB.Table("room_invites").Omit("Room", "Inviter", "User").Create(roomInvites).Error; err != nil {
		return nil, err
	}

	return &roomInvites, nil
}

func (rs *RoomService) RemoveUserFromRoom(roomId string, userId string) error {
	db := rs.DB
	if err := db.
		Exec("DELETE FROM room_users WHERE room_id = ? AND user_id = ?", roomId, userId).Error; err != nil {
		return err
	}
	return nil
}

func (rs *RoomService) GetUninvitedFriendsForRoom(roomId string, userId string) (*[]model.User, error) {
	db := rs.DB
	var friends []model.User

	// Get friends who are not in the room and don't have pending invites
	if err := db.
		Distinct("users.*").
		Table("users").
		Joins("JOIN user_friends ON (user_friends.friend_id = ? AND user_friends.user_id = users.id)", userId).
		// Exclude users already in room
		Where("users.id NOT IN (SELECT user_id FROM room_users WHERE room_id = ?)", roomId).
		// Exclude users with pending invites
		Where("users.id NOT IN (SELECT user_id FROM room_invites WHERE room_id = ? AND status = 'pending')", roomId).
		Find(&friends).Error; err != nil {
		return nil, err
	}

	return &friends, nil
}
