package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Username   string `gorm:"index:username_idx, class:FULLTEXT, unique; not null" json:"username"`
	Email      string `gorm:"unique; not null" json:"email"`
	Password   string `gorm:"not null" json:"password"`
	PictureUrl string `gorm:"default:'https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg'" json:"pictureUrl"`
	// Name         string    `json:"name"`
	// PhoneNum     string    `gorm:"default:null" json:"phoneNum"`
	IsEmailValid bool      `gorm:"default:false" json:"isEmailValid"`
	IsOnline     bool      `gorm:"default:false" json:"isOnline"`
	LastSeen     time.Time `json:"lastSeen"`
	RegisteredAt time.Time `gorm:"autoCreateTime" json:"registeredAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Associations
	Rooms   []Room `gorm:"many2many:room_users" json:"rooms"`
	Friends []User `gorm:"many2many:user_friends" json:"friends"`
}

type FriendRequest struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SenderID    uint      `gorm:"not null" json:"senderId"`
	ReceiverID  uint      `gorm:"not null" json:"receiverId"`
	Status      string    `gorm:"default:'pending'" json:"status"`
	SentAt      time.Time `gorm:"autoCreateTime" json:"sentAt"`
	RespondedAt time.Time `json:"respondedAt,omitempty"` // Nullable, only set when accepted/rejected

	// Associations
	Sender   User `gorm:"foreignKey:sender_id; constraint:OnDelete:CASCADE" json:"sender"`
	Receiver User `gorm:"foreignKey:receiver_id; constraint:OnDelete:CASCADE" json:"receiver"`
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
