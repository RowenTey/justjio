package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"unique; not null" json:"username"`
	Email        string    `gorm:"unique; not null" json:"email"`
	Password     string    `gorm:"not null" json:"password"`
	Name         string    `json:"name"`
	PhoneNum     string    `gorm:"default:null" json:"phoneNum"`
	Rooms        []Room    `gorm:"many2many:room_users" json:"rooms"`
	Friends      []User    `gorm:"many2many:user_friends" json:"friends"`
	IsEmailValid bool      `gorm:"default:false" json:"isEmailValid"`
	IsOnline     bool      `gorm:"default:false" json:"isOnline"`
	LastSeen     time.Time `json:"lastSeen"`
	RegisteredAt time.Time `gorm:"autoCreateTime" json:"registeredAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
