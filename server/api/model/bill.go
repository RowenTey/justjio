package model

import (
	"time"

	"gorm.io/gorm"
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

type Transaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ConsolidationID uint      `gorm:"not null" json:"consolidationId"`
	PayerID         uint      `gorm:"not null" json:"payerId"`
	PayeeID         uint      `gorm:"not null" json:"payeeId"`
	Amount          float32   `gorm:"not null" json:"amount"`
	IsPaid          bool      `gorm:"default:false" json:"isPaid"`
	PaidOn          time.Time `gorm:"default:null" json:"paidOn"`

	// Associations
	Payer         User          `gorm:"not null; foreignKey:payer_id" json:"payer"`
	Payee         User          `gorm:"not null; foreignKey:payee_id" json:"payee"`
	Consolidation Consolidation `gorm:"not null" json:"consolidation"`
}

func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	if t.IsPaid {
		t.PaidOn = time.Now()
	}
	return nil
}
