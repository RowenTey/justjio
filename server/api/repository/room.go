package repository

import (
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	WithTx(tx *gorm.DB) RoomRepository

	Create(room *model.Room) error
	GetByID(roomID string) (*model.Room, error)
	GetUserRooms(userID string, page int, pageSize int) (*[]model.Room, error)
	CountUserRooms(userID string) (int64, error)
	GetRoomAttendees(roomID string) (*[]model.User, error)
	GetRoomAttendeeIDs(roomID string) (*[]string, error)
	CloseRoom(roomID string) error
	UpdateRoom(room *model.Room) error
	AddUserToRoom(roomID string, user *model.User) error
	RemoveUserFromRoom(roomID, userID string) error
	IsUserInRoom(roomID, userID string) (bool, error)

	// Invite related methods
	GetPendingInvites(userID string) (*[]model.RoomInvite, error)
	CountPendingInvites(userID string) (int64, error)
	UpdateInviteStatus(roomID, userID, status string) error
	CreateInvites(invites *[]model.RoomInvite) error
	DeletePendingInvites(roomID string) error
	HasPendingInvites(roomID, userID string) (bool, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

// WithTx returns a new RoomRepository with the provided transaction
func (r *roomRepository) WithTx(tx *gorm.DB) RoomRepository {
	if tx == nil {
		return r
	}
	return &roomRepository{db: tx}
}

func (r *roomRepository) Create(room *model.Room) error {
	return r.db.Table("rooms").Create(&room).Error
}

func (r *roomRepository) GetByID(roomID string) (*model.Room, error) {
	var room model.Room
	err := r.db.Table("rooms").Preload("Users").First(&room, "id = ?", roomID).Error
	return &room, err
}

func (r *roomRepository) GetUserRooms(userID string, page int, pageSize int) (*[]model.Room, error) {
	var rooms []model.Room
	err := r.db.
		Model(&model.Room{}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("room_users.user_id = ?", userID).
		Where("rooms.is_closed = ?", false).
		Order("rooms.updated_at DESC").
		Scopes(database.Paginate(page, pageSize)).
		Find(&rooms).Error
	return &rooms, err
}

func (r *roomRepository) CountUserRooms(userID string) (int64, error) {
	var count int64
	err := r.db.
		Model(&model.Room{}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("room_users.user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *roomRepository) GetRoomAttendees(roomID string) (*[]model.User, error) {
	var room model.Room
	err := r.db.Table("rooms").Preload("Users").First(&room, "id = ?", roomID).Error
	if err != nil {
		return nil, err
	}
	return &room.Users, nil
}

func (r *roomRepository) GetRoomAttendeeIDs(roomID string) (*[]string, error) {
	// TODO: test this function
	var userIds []string
	// var users []model.User

	err := r.db.
		Table("users").
		Joins("JOIN room_users ON room_users.user_id = users.id").
		Where("room_users.room_id = ?", roomID).
		Select("users.id").
		Find(&userIds).Error
	if err != nil {
		return nil, err
	}

	return &userIds, nil
}

func (r *roomRepository) CloseRoom(roomID string) error {
	return r.db.
		Model(&model.Room{}).
		Where("id = ?", roomID).
		Update("is_closed", true).
		Error
}

func (r *roomRepository) UpdateRoom(room *model.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) AddUserToRoom(roomID string, user *model.User) error {
	return r.db.Exec(
		"INSERT INTO room_users (room_id, user_id) VALUES (?, ?)",
		roomID,
		user.ID,
	).Error
}

func (r *roomRepository) RemoveUserFromRoom(roomID, userID string) error {
	return r.db.
		Exec("DELETE FROM room_users WHERE room_id = ? AND user_id = ?", roomID, userID).
		Error
}

func (r *roomRepository) IsUserInRoom(roomID, userID string) (bool, error) {
	var count int64
	err := r.db.
		Model(&model.Room{}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("rooms.id = ? AND room_users.user_id = ?", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *roomRepository) DeletePendingInvites(roomID string) error {
	return r.db.
		Where("room_id = ? AND status = ?", roomID, "pending").
		Delete(&model.RoomInvite{}).Error
}

func (r *roomRepository) GetPendingInvites(userID string) (*[]model.RoomInvite, error) {
	var invites []model.RoomInvite
	err := r.db.
		Preload("Room.Host").
		Preload("User").
		Preload("Inviter").
		Where("user_id = ? AND status = ?", userID, "pending").
		Find(&invites).Error
	return &invites, err
}

func (r *roomRepository) CountPendingInvites(userID string) (int64, error) {
	var count int64
	err := r.db.
		Model(&model.RoomInvite{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&count).Error
	return count, err
}

func (r *roomRepository) UpdateInviteStatus(roomID, userID, status string) error {
	return r.db.
		Model(&model.RoomInvite{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("status", status).Error
}

func (r *roomRepository) CreateInvites(invites *[]model.RoomInvite) error {
	return r.db.Table("room_invites").Omit("Room", "Inviter", "User").Create(invites).Error
}

func (r *roomRepository) HasPendingInvites(roomID, userID string) (bool, error) {
	var count int64
	err := r.db.
		Model(&model.RoomInvite{}).
		Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "pending").
		Count(&count).Error
	return count > 0, err
}
