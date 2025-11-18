package repositories

import (
	"context"
	"database/sql"
	"time"

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
	if err := r.db.
		WithContext(ctx).
		First(&room, "id = ?", roomID).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *roomRepository) GetByIDWithAttendees(ctx context.Context, roomID string) (*models.Room, error) {
	type row struct {
		ID             string
		Name           string
		Time           string
		Venue          string
		VenuePlaceId   string
		VenueUrl       string
		Date           time.Time
		Description    string
		HostID         uint
		NoOfAttendees  int
		Consolidated   string
		IsClosed       bool
		IsPrivate      bool
		ImageUrl       string
		CreatedAt      time.Time
		UpdatedAt      time.Time
		HostUsername   sql.NullString
		HostPictureURL sql.NullString
		UserID         sql.NullInt64
		UserUsername   sql.NullString
		UserPictureURL sql.NullString
	}

	rows, err := r.db.WithContext(ctx).
		Table("rooms r").
		Select(`
			r.*, 
			h.username 		AS host_username,
			h.picture_url 	AS host_picture_url,
			u.id 			AS user_id,
			u.username 		AS user_username,
			u.picture_url 	AS user_picture_url
		`).
		Where("r.id = ?", roomID).
		Joins("LEFT JOIN users h ON r.host_id = h.id").
		Joins("LEFT JOIN room_users ru ON r.id = ru.room_id").
		Joins("LEFT JOIN users u ON ru.user_id = u.id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer func() error {
		return rows.Close()
	}()

	var result *models.Room
	for rows.Next() {
		var sc row
		if err := r.db.ScanRows(rows, &sc); err != nil {
			return nil, err
		}

		if result == nil {
			result = &models.Room{
				ID:            sc.ID,
				Name:          sc.Name,
				Time:          sc.Time,
				Venue:         sc.Venue,
				VenuePlaceId:  sc.VenuePlaceId,
				VenueUrl:      sc.VenueUrl,
				Date:          sc.Date,
				Description:   sc.Description,
				HostID:        sc.HostID,
				NoOfAttendees: sc.NoOfAttendees,
				Consolidated:  sc.Consolidated,
				IsClosed:      sc.IsClosed,
				IsPrivate:     sc.IsPrivate,
				ImageUrl:      sc.ImageUrl,
				CreatedAt:     sc.CreatedAt,
				UpdatedAt:     sc.UpdatedAt,
			}

			if sc.HostUsername.Valid {
				result.Host = models.User{
					ID:         sc.HostID,
					Username:   sc.HostUsername.String,
					PictureUrl: sc.HostPictureURL.String,
				}
			}
		}

		if sc.UserID.Valid {
			userID := uint(sc.UserID.Int64)
			result.Users = append(result.Users, models.User{
				ID:         userID,
				Username:   sc.UserUsername.String,
				PictureUrl: sc.UserPictureURL.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	} else if result == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return result, nil
}

func (r *roomRepository) GetUserRooms(ctx context.Context, userID string, page int, pageSize int) ([]models.Room, error) {
	var rooms []models.Room
	if err := r.db.
		WithContext(ctx).
		Table("rooms r").
		Joins("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Joins("LEFT JOIN room_users ru ON r.id = ru.room_id").
		Where("ru.user_id = ?", userID).
		Where("r.is_closed = ?", false).
		Order("r.updated_at DESC").
		Scopes(database.Paginate(page, pageSize)).
		Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *roomRepository) CountUserRooms(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.db.
		WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Pluck("no_of_rooms", &count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *roomRepository) GetUnjoinedRoomsByIsPrivate(ctx context.Context, userID string, isPrivate bool) ([]models.Room, error) {
	var rooms []models.Room
	if err := r.db.
		WithContext(ctx).
		Table("rooms r").
		Where("r.is_private = ?", isPrivate).
		Where("r.id NOT IN (SELECT room_id FROM room_users WHERE user_id = ?)", userID).
		Where("r.id NOT IN (SELECT room_id FROM room_invites WHERE user_id = ?)", userID).
		Order("r.updated_at DESC").
		Joins("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, picture_url")
		}).
		Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *roomRepository) GetRoomAttendeeIDs(ctx context.Context, roomID string) ([]string, error) {
	var userIds []string
	if err := r.db.
		WithContext(ctx).
		Table("room_users").
		Where("room_id = ?", roomID).
		Select("user_id").
		Find(&userIds).Error; err != nil {
		return nil, err
	}

	return userIds, nil
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
	if err := r.db.
		WithContext(ctx).
		Table("room_invites").
		Where("room_id = ? AND status = ?", roomID, "pending").
		Select("user_id").
		Find(&userIds).Error; err != nil {
		return nil, err
	}

	return userIds, nil
}

func (r *roomRepository) GetPendingInvites(ctx context.Context, userID string) ([]models.RoomInvite, error) {
	var invites []models.RoomInvite
	if err := r.db.
		WithContext(ctx).
		Table("room_invites ri").
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
		Where("ri.user_id = ? AND ri.status = ?", userID, "pending").
		Find(&invites).Error; err != nil {
		return nil, err
	}

	return invites, nil
}

func (r *roomRepository) CountPendingInvites(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Pluck("no_of_pending_room_invites", &count).Error; err != nil {
		return 0, err
	}

	return count, nil
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
	if err := r.db.
		WithContext(ctx).
		Model(&models.RoomInvite{}).
		Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "pending").
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
