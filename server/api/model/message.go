package model

import (
	"time"
)

type Message struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	RoomID   string    `gorm:"primaryKey; autoIncrement:false; type:uuid" json:"roomId"`
	SenderID uint      `gorm:"not null" json:"senderId"`
	Content  string    `gorm:"not null" json:"content"`
	SentAt   time.Time `gorm:"autoCreateTime" json:"sentAt"`

	// Associations
	Sender User `gorm:"not null; foreignKey:sender_id" json:"sender"`
	Room   Room `gorm:"not null; foreignKey:room_id" json:"room"`
}
