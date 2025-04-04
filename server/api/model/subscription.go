package model

import (
	"github.com/RowenTey/JustJio/server/api/utils"
	"gorm.io/gorm"
)

type Subscription struct {
	ID       string `gorm:"primaryKey; type:uuid" json:"id"`
	UserID   uint   `gorm:"not null" json:"userId"`
	Endpoint string `gorm:"unique; not null" json:"endpoint"`
	Auth     string `gorm:"not null" json:"auth"`
	P256dh   string `gorm:"not null" json:"p256dh"`
}

func (sub *Subscription) BeforeCreate(tx *gorm.DB) error {
	// Generate ULID first
	ulid := utils.CreateULID()

	// Convert to UUID for storage
	uuid := utils.ULIDToUUID(ulid)
	sub.ID = uuid.String()
	return nil
}
