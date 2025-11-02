package models

import (
	"time"
)

type Consolidation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

type Bill struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"not null" json:"name"`
	Amount          float32   `gorm:"not null" json:"amount"`
	Date            time.Time `gorm:"not null" json:"date"`
	IncludeOwner    bool      `gorm:"default:true" json:"includeOwner"`
	RoomID          string    `gorm:"not null; type:uuid" json:"roomId"`
	OwnerID         uint      `gorm:"not null" json:"ownerId"`
	ConsolidationID uint      `gorm:"default:null" json:"consolidationId"`

	// Associations
	Owner         User          `gorm:"not null; foreignKey:owner_id" json:"owner"`
	Room          Room          `gorm:"not null; foreignKey:room_id" json:"room"`
	Consolidation Consolidation `json:"consolidation"`
	Payers        []User        `gorm:"many2many:payers" json:"payers"`
}
