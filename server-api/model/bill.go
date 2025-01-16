package model

import (
	"time"

	"gorm.io/gorm"
)

type Bill struct {
	ID              uint          `gorm:"primaryKey"`
	Name            string        `gorm:"not null" json:"name"`
	Amount          float32       `gorm:"not null" json:"amount"`
	Date            time.Time     `gorm:"not null" json:"date"`
	RoomID          string        `gorm:"not null; type:uuid" json:"room_id"`                         // Foreign key to Room table
	Room            Room          `gorm:"not null" json:"room"`                                       // gorm feature -> not actually stored in DB
	OwnerID         uint          `gorm:"not null" json:"owner_id"`                                   // Foreign key to User table
	Owner           User          `gorm:"not null; foreignKey:owner_id" json:"owner"`                 // gorm feature -> not actually stored in DB
	ConsolidationID uint          `gorm:"not null" json:"consolidation_id"`                           // Foreign key to Consolidation table
	Consolidation   Consolidation `gorm:"not null; foreignKey:consolidation_id" json:"consolidation"` // gorm feature -> not actually stored in DB
	Payers          []User        `gorm:"many2many:payers" json:"payers"`
}

type Consolidation struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Transaction struct {
	ID              uint          `gorm:"primaryKey"`
	ConsolidationID uint          `gorm:"not null" json:"consolidation_id"`           // Foreign key to Consolidation table
	Consolidation   Consolidation `gorm:"not null" json:"consolidation"`              // gorm feature -> not actually stored in DB
	PayerID         uint          `gorm:"not null" json:"payer_id"`                   // Foreign key to User table
	Payer           User          `gorm:"not null; foreignKey:payer_id" json:"payer"` // gorm feature -> not actually stored in DB
	PayeeID         uint          `gorm:"not null" json:"payee_id"`                   // Foreign key to User table
	Payee           User          `gorm:"not null; foreignKey:payee_id" json:"payee"` // gorm feature -> not actually stored in DB
	Amount          float32       `gorm:"not null" json:"amount"`
	IsPaid          bool          `gorm:"default:false" json:"is_paid"`
	PaidOn          time.Time     `gorm:"default:null" json:"paid_on"`
}

func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	if t.IsPaid == true {
		t.PaidOn = time.Now()
	}
	return nil
}
