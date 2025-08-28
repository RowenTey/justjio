package model

import (
	"time"

	"github.com/RowenTey/JustJio/server/api/utils"
	"gorm.io/gorm"
)

type Room struct {
	ID             string    `gorm:"primaryKey; type:uuid" json:"id"`
	Name           string    `gorm:"not null" json:"name"`
	Time           string    `gorm:"not null" json:"time"`
	Venue          string    `gorm:"not null" json:"venue"`
	VenuePlaceId   string    `gorm:"not null" json:"venuePlaceId"`
	VenueUrl       string    `gorm:"not null" json:"venueUrl"`
	Date           time.Time `gorm:"not null" json:"date"`
	Description    string    `gorm:"not null" json:"description"`
	HostID         uint      `gorm:"not null" json:"hostId"`
	AttendeesCount int       `gorm:"default:1" json:"attendeesCount"`
	Consolidated   bool      `gorm:"default:false" json:"consolidated"`
	IsClosed       bool      `gorm:"default:false" json:"isClosed"`
	IsPrivate      bool      `gorm:"default:false" json:"isPrivate"`
	ImageUrl       string    `gorm:"not null" json:"imageUrl"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Associations
	Host  User   `gorm:"not null; foreignKey:host_id" json:"host"`
	Users []User `gorm:"many2many:room_users" json:"users"`
}

func (room *Room) BeforeCreate(tx *gorm.DB) error {
	// Generate ULID first
	ulid := utils.CreateULID()

	// Convert to UUID for storage
	uuid := utils.ULIDToUUID(ulid)
	room.ID = uuid.String()
	return nil
}

type RoomInvite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    string    `gorm:"not null; type:uuid" json:"roomId"`
	UserID    uint      `gorm:"not null" json:"userId"`
	InviterID uint      `gorm:"not null" json:"inviterId"`
	Status    string    `gorm:"not null; default:'pending'" json:"status"` // Invite status (pending, accepted, rejected)
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Associations
	User    User `gorm:"not null" json:"user"`
	Inviter User `gorm:"not null; foreignKey:inviter_id" json:"inviter"`
	Room    Room `gorm:"not null" json:"room"`
}
