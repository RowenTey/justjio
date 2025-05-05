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
	Date           time.Time `gorm:"not null" json:"date"`
	HostID         uint      `gorm:"not null" json:"hostId"`
	AttendeesCount int       `gorm:"default:1" json:"attendeesCount"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	IsClosed       bool      `gorm:"default:false" json:"isClosed"`

	// Associations
	Host  User   `gorm:"not null; foreignKey:host_id" json:"host"`
	Users []User `gorm:"many2many:room_users" json:"users"`
}

type RoomInvite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    string    `gorm:"not null; type:uuid" json:"roomId"`
	UserID    uint      `gorm:"not null" json:"userId"`
	InviterID uint      `gorm:"not null" json:"inviterId"`
	Status    string    `gorm:"not null; default:'pending'" json:"status"` // Invite status (pending, accepted, rejected)
	Message   string    `json:"message"`                                   // Optional message from the inviter
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Associations
	User    User `gorm:"not null" json:"user"`
	Inviter User `gorm:"not null; foreignKey:inviter_id" json:"inviter"`
	Room    Room `gorm:"not null" json:"room"`
}

func (room *Room) BeforeCreate(tx *gorm.DB) error {
	// Generate ULID first
	ulid := utils.CreateULID()

	// Convert to UUID for storage
	uuid := utils.ULIDToUUID(ulid)
	room.ID = uuid.String()
	return nil
}

// func (invite *RoomInvite) BeforeCreate(tx *gorm.DB) error {
// 	invite.RoomID = invite.Room.ID
// 	return nil
// }
