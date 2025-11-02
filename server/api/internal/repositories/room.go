package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/database"
	"gorm.io/gorm"
)

type RoomRepository interface {
	WithTx(tx *gorm.DB) RoomRepository

	Create(ctx context.Context, room *models.Room) error
	GetByID(ctx context.Context, roomID string) (*models.Room, error)
	GetByIDWithAttendees(ctx context.Context, roomID string) (*models.Room, error)
	GetUserRooms(ctx context.Context, userID string, page int, pageSize int) ([]models.Room, error)
	CountUserRooms(ctx context.Context, userID string) (int64, error)
	GetUnjoinedRoomsByIsPrivate(ctx context.Context, userID string, isPrivate bool) ([]models.Room, error)
	GetRoomAttendeeIDs(ctx context.Context, roomID string) ([]string, error)
	CloseRoom(ctx context.Context, roomID string) error
	Update(ctx context.Context, room *models.Room) error
	AddUserToRoom(ctx context.Context, roomID string, user *models.User) error
	RemoveUserFromRoom(ctx context.Context, roomID, userID string) error

	// Invite related methods
	GetPendingInvites(ctx context.Context, userID string) ([]models.RoomInvite, error)
	CountPendingInvites(ctx context.Context, userID string) (int64, error)
	UpdateInviteStatus(ctx context.Context, roomID, userID, status string) error
	CreateInvites(ctx context.Context, invites []models.RoomInvite) error
	DeletePendingInvites(ctx context.Context, roomID string) error
	HasPendingInvites(ctx context.Context, roomID, userID string) (bool, error)
	GetPendingInviteUsers(ctx context.Context, roomID string) ([]string, error)
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

func (r *roomRepository) Create(ctx context.Context, room *models.Room) error {
	return r.db.WithContext(ctx).Table("rooms").Create(&room).Error
}

func (r *roomRepository) GetByID(ctx context.Context, roomID string) (*models.Room, error) {
	var room models.Room
	err := r.db.
		WithContext(ctx).
		Table("rooms").
		Preload("Users", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		First(&room, "id = ?", roomID).Error
	return &room, err
}

func (r *roomRepository) GetByIDWithAttendees(ctx context.Context, roomID string) (*models.Room, error) {
	var room models.Room
	err := r.db.
		WithContext(ctx).
		Preload("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Preload("Users", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		First(&room, "id = ?", roomID).Error
	return &room, err
}

func (r *roomRepository) GetUserRooms(ctx context.Context, userID string, page int, pageSize int) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.
		WithContext(ctx).
		Model(&models.Room{}).
		Preload("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Joins("JOIN room_users ON rooms.id = room_users.room_id").
		Where("room_users.user_id = ?", userID).
		Where("rooms.is_closed = ?", false).
		Order("rooms.updated_at DESC").
		Scopes(database.Paginate(page, pageSize)).
		Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) CountUserRooms(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.
		WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Pluck("no_of_rooms", &count).Error
	return count, err
}

func (r *roomRepository) GetUnjoinedRoomsByIsPrivate(ctx context.Context, userID string, isPrivate bool) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.
		WithContext(ctx).
		Table("rooms").
		Joins("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Where("rooms.is_private = ?", isPrivate).
		Where("rooms.id NOT IN (SELECT room_id FROM room_users WHERE user_id = ?)", userID).
		Where("rooms.id NOT IN (SELECT room_id FROM room_invites WHERE user_id = ?)", userID).
		Order("rooms.updated_at DESC").
		Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) GetRoomAttendeeIDs(ctx context.Context, roomID string) ([]string, error) {
	var userIds []string
	err := r.db.
		WithContext(ctx).
		Table("room_users").
		Where("room_id = ?", roomID).
		Select("user_id").
		Find(&userIds).Error
	return userIds, err
}

func (r *roomRepository) CloseRoom(ctx context.Context, roomID string) error {
	return r.db.
		WithContext(ctx).
		Model(&models.Room{ID: roomID}).
		Update("is_closed", true).
		Error
}

func (r *roomRepository) Update(ctx context.Context, room *models.Room) error {
	return r.db.WithContext(ctx).Updates(room).Error
}

func (r *roomRepository) AddUserToRoom(ctx context.Context, roomID string, user *models.User) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO room_users (room_id, user_id) VALUES (?, ?)",
		roomID,
		user.ID,
	).Error
}

func (r *roomRepository) RemoveUserFromRoom(ctx context.Context, roomID, userID string) error {
	return r.db.
		WithContext(ctx).
		Exec("DELETE FROM room_users WHERE room_id = ? AND user_id = ?", roomID, userID).
		Error
}

func (r *roomRepository) DeletePendingInvites(ctx context.Context, roomID string) error {
	return r.db.
		WithContext(ctx).
		Where("room_id = ? AND status = ?", roomID, "pending").
		Delete(&models.RoomInvite{}).Error
}

func (r *roomRepository) GetPendingInviteUsers(ctx context.Context, roomID string) ([]string, error) {
	var userIds []string
	err := r.db.
		WithContext(ctx).
		Table("room_invites").
		Where("room_id = ? AND status = ?", roomID, "pending").
		Select("user_id").
		Find(&userIds).Error
	return userIds, err
}

func (r *roomRepository) GetPendingInvites(ctx context.Context, userID string) ([]models.RoomInvite, error) {
	var invites []models.RoomInvite
	err := r.db.
		WithContext(ctx).
		Joins("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, time, venue, venue_url, date, description, image_url, is_private, no_of_attendees, host_id")
		}).
		Joins("Room.Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Joins("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Joins("Inviter", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Find(&invites).Error
	return invites, err
}

func (r *roomRepository) CountPendingInvites(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Pluck("no_of_pending_room_invites", &count).Error
	return count, err
}

func (r *roomRepository) UpdateInviteStatus(ctx context.Context, roomID, userID, status string) error {
	return r.db.
		WithContext(ctx).
		Model(&models.RoomInvite{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("status", status).Error
}

func (r *roomRepository) CreateInvites(ctx context.Context, invites []models.RoomInvite) error {
	if len(invites) == 0 {
		return nil
	}

	return r.db.
		WithContext(ctx).
		Table("room_invites").
		Create(&invites).Error
}

func (r *roomRepository) HasPendingInvites(ctx context.Context, roomID, userID string) (bool, error) {
	var count int64
	err := r.db.
		WithContext(ctx).
		Model(&models.RoomInvite{}).
		Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "pending").
		Count(&count).Error
	return count > 0, err
}
