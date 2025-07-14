package repository

import (
	"errors"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type Status int

const (
	NO_BILLS Status = iota
	UNCONSOLIDATED
	CONSOLIDATED
)

type BillRepository interface {
	WithTx(tx *gorm.DB) BillRepository

	Create(bill *model.Bill) error
	FindByID(billID uint) (*model.Bill, error)
	FindByRoom(roomID string) (*[]model.Bill, error)
	DeleteByRoom(roomID string) error
	GetRoomBillConsolidationStatus(roomID string) (Status, error)
	FindByConsolidation(consolidationID uint) (*[]model.Bill, error)
	ConsolidateBills(roomID string) (*model.Consolidation, error)
}

type billRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepository{db: db}
}

// WithTx returns a new BillRepository with the provided transaction
func (r *billRepository) WithTx(tx *gorm.DB) BillRepository {
	if tx == nil {
		return r
	}
	return &billRepository{db: tx}
}

func (r *billRepository) Create(bill *model.Bill) error {
	return r.db.Omit("Room", "Owner").Create(bill).Error
}

func (r *billRepository) FindByID(billID uint) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.Where("id = ?", billID).First(&bill).Error
	return &bill, err
}

func (r *billRepository) FindByRoom(roomID string) (*[]model.Bill, error) {
	var bills []model.Bill
	err := r.db.
		Where("room_id = ?", roomID).
		Preload("Owner").
		Preload("Payers").
		Find(&bills).Error
	return &bills, err
}

func (r *billRepository) DeleteByRoom(roomID string) error {
	return r.db.Where("room_id = ?", roomID).Delete(&model.Bill{}).Error
}

func (r *billRepository) GetRoomBillConsolidationStatus(roomID string) (Status, error) {
	var bill *model.Bill
	err := r.db.Where("room_id = ?", roomID).First(&bill).Error
	// bill found -> check if consolidation ID is set
	if err == nil && bill.ConsolidationID != 0 {
		return CONSOLIDATED, nil
	} else if err == nil && bill.ConsolidationID == 0 {
		return UNCONSOLIDATED, nil
	}

	// record not found -> no bills in room
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NO_BILLS, nil
	}

	return NO_BILLS, err
}

func (r *billRepository) ConsolidateBills(roomID string) (*model.Consolidation, error) {
	// Create empty struct as fields will be auto populated by DB
	consolidation := model.Consolidation{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.db.
			Model(&model.Consolidation{}).
			Create(&consolidation).Error; err != nil {
			return err
		}

		if err := r.db.Table("bills").
			Where("room_id = ?", roomID).
			Update("consolidation_id", consolidation.ID).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &consolidation, nil
}

func (r *billRepository) FindByConsolidation(consolidationID uint) (*[]model.Bill, error) {
	var bills []model.Bill
	err := r.db.
		Model(&model.Bill{}).
		Where("consolidation_id = ?", consolidationID).
		Preload("Payers").
		Find(&bills).Error
	return &bills, err
}
