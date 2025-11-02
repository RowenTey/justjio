package models

import (
	"time"

	"gorm.io/gorm"
)

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
