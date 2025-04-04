package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	RoomID   string    `gorm:"primaryKey; autoIncrement:false; type:uuid" json:"roomId"`
	SenderID uint      `gorm:"not null" json:"senderId"`
	Content  string    `gorm:"not null" json:"content"`
	SentAt   time.Time `gorm:"autoCreateTime" json:"sentAt"`

	// Associations
	Sender User `gorm:"not null" json:"sender"`
	Room   Room `gorm:"not null" json:"room"`
}

func (msg *Message) BeforeCreate(tx *gorm.DB) error {
	msg.RoomID = msg.Room.ID
	return nil
}
