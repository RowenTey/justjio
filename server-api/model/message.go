package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	RoomID   string    `gorm:"primaryKey; autoIncrement:false; type:uuid" json:"room_id"`
	Room     Room      `gorm:"not null" json:"room"` // gorm feature -> not actually stored in DB
	SenderID uint      `gorm:"not null" json:"sender_id"`
	Sender   User      `gorm:"not null" json:"sender"`
	Content  string    `gorm:"not null" json:"content"`
	SentAt   time.Time `gorm:"autoCreateTime" json:"sent_at"`
}

func (msg *Message) BeforeCreate(tx *gorm.DB) error {
	msg.RoomID = msg.Room.ID
	return nil
}
