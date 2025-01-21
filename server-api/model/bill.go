package model

import (
	"time"

	"gorm.io/gorm"
)

type Bill struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	Name            string        `gorm:"not null" json:"name"`
	Amount          float32       `gorm:"not null" json:"amount"`
	Date            time.Time     `gorm:"not null" json:"date"`
	IncludeOwner    bool          `gorm:"default:true" json:"includeOwner"`
	RoomID          string        `gorm:"not null; type:uuid" json:"roomId"`          // Foreign key to Room table
	Room            Room          `gorm:"not null" json:"room"`                       // gorm feature -> not actually stored in DB
	OwnerID         uint          `gorm:"not null" json:"ownerId"`                    // Foreign key to User table
	Owner           User          `gorm:"not null; foreignKey:owner_id" json:"owner"` // gorm feature -> not actually stored in DB
	ConsolidationID uint          `gorm:"default:null" json:"consolidationId"`        // Foreign key to Consolidation table
	Consolidation   Consolidation `json:"consolidation"`                              // gorm feature -> not actually stored in DB
	Payers          []User        `gorm:"many2many:payers" json:"payers"`
}

type Consolidation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

type Transaction struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	ConsolidationID uint          `gorm:"not null" json:"consolidationId"`            // Foreign key to Consolidation table
	Consolidation   Consolidation `gorm:"not null" json:"consolidation"`              // gorm feature -> not actually stored in DB
	PayerID         uint          `gorm:"not null" json:"payerId"`                    // Foreign key to User table
	Payer           User          `gorm:"not null; foreignKey:payer_id" json:"payer"` // gorm feature -> not actually stored in DB
	PayeeID         uint          `gorm:"not null" json:"payeeId"`                    // Foreign key to User table
	Payee           User          `gorm:"not null; foreignKey:payee_id" json:"payee"` // gorm feature -> not actually stored in DB
	Amount          float32       `gorm:"not null" json:"amount"`
	IsPaid          bool          `gorm:"default:false" json:"isPaid"`
	PaidOn          time.Time     `gorm:"default:null" json:"paidOn"`
}

func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	if t.IsPaid {
		t.PaidOn = time.Now()
	}
	return nil
}
