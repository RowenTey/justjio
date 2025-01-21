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
	room.Host = *host
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
	db := database.DB
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
		return errors.New("User is not the host of the room")
	}

	room.IsClosed = true
	room.UpdatedAt = time.Now()
	if err := db.Save(&room).Error; err != nil {
		return err
	}

	return nil
}

func (rs *RoomService) UpdateRoomInviteStatus(roomId string, userId string, status string) error {
	if status != "accepted" && status != "rejected" {
		return errors.New("Invalid status")
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
	roomsDB := rs.DB.Table("rooms")
	roomInvitesDB := rs.DB.Table("room_invites")
	var room model.Room
	var roomInvites []model.RoomInvite

	if err := roomsDB.First(&room, "id = ?", roomId).Error; err != nil {
		return nil, err
	}
	log.Printf("[ROOM] Room %s found", roomId)

	if room.HostID != inviter.ID {
		return nil, errors.New("User is not the host of the room")
	}

	for _, user := range *users {
		roomInvite := model.RoomInvite{
			Room:      room,
			User:      user,
			InviterID: inviter.ID,
			Inviter:   *inviter,
			Message:   message,
			CreatedAt: time.Now(),
			Status:    "pending",
		}
		log.Printf("[ROOM] Room ID of invite to be created: %v", roomInvite.Room.ID)
		roomInvites = append(roomInvites, roomInvite)
	}

	if err := roomInvitesDB.Omit("Room").Create(roomInvites).Error; err != nil {
		return nil, err
	}

	for _, invite := range roomInvites {
		log.Println("[ROOM] Invite created with room ID: ", invite.RoomID)
	}

	return &roomInvites, nil
}

func (rs *RoomService) RemoveUserFromRoom(roomId string, userId string) error {
	db := rs.DB
	if err := db.
		Exec("DELETE FROM room_users WHERE room_id = ? AND user_id = ?", roomId, userId).Error; err != nil {
		return errors.New("Error removing user from room: " + err.Error())
	}
	return nil
}
