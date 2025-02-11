package model

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"primaryKey; autoIncrement:false;" json:"userId"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	IsRead    bool      `gorm:"default:false" json:"isRead"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
