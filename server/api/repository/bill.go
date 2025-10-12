package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type BillRepository interface {
	WithTx(tx *gorm.DB) BillRepository

	Create(bill *model.Bill) error
	FindByID(billID uint) (*model.Bill, error)
	FindByRoom(roomID string) ([]model.Bill, error)
	DeleteByRoom(roomID string) error
	FindByConsolidation(consolidationID uint) ([]model.Bill, error)
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
	return r.db.Create(bill).Error
}

func (r *billRepository) FindByID(billID uint) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.Where("id = ?", billID).First(&bill).Error
	return &bill, err
}

func (r *billRepository) FindByRoom(roomID string) ([]model.Bill, error) {
	var bills []model.Bill
	err := r.db.
		Where("room_id = ?", roomID).
		Preload("Owner").
		Preload("Payers").
		Find(&bills).Error
	return bills, err
}

func (r *billRepository) DeleteByRoom(roomID string) error {
	return r.db.Where("room_id = ?", roomID).Delete(&model.Bill{}).Error
}

func (r *billRepository) ConsolidateBills(roomID string) (*model.Consolidation, error) {
	// Create empty struct as fields will be auto populated by DB
	consolidation := model.Consolidation{}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
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
	}); err != nil {
		return nil, err
	}

	return &consolidation, nil
}

func (r *billRepository) FindByConsolidation(consolidationID uint) ([]model.Bill, error) {
	var bills []model.Bill
	err := r.db.
		Model(&model.Bill{}).
		Preload("Payers").
		Where("consolidation_id = ?", consolidationID).
		Find(&bills).Error
	return bills, err
}
