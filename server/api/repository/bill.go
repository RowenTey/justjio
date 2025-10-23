package repository

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type BillRepository interface {
	WithTx(tx *gorm.DB) BillRepository

	Create(ctx context.Context, bill *model.Bill) error
	FindByID(ctx context.Context, billID uint) (*model.Bill, error)
	FindByRoom(ctx context.Context, roomID string) ([]model.Bill, error)
	FindByConsolidation(ctx context.Context, consolidationID uint) ([]model.Bill, error)
	DeleteByRoom(ctx context.Context, roomID string) error
	ConsolidateBills(ctx context.Context, roomID string) (*model.Consolidation, error)
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

func (r *billRepository) Create(ctx context.Context, bill *model.Bill) error {
	return r.db.WithContext(ctx).Create(bill).Error
}

func (r *billRepository) FindByID(ctx context.Context, billID uint) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.
		WithContext(ctx).
		Where("id = ?", billID).
		Preload("Owner", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Preload("Payers", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		First(&bill).Error
	return &bill, err
}

func (r *billRepository) FindByRoom(ctx context.Context, roomID string) ([]model.Bill, error) {
	var bills []model.Bill
	err := r.db.
		WithContext(ctx).
		Where("room_id = ?", roomID).
		Preload("Owner", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Preload("Payers", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Find(&bills).Error
	return bills, err
}

func (r *billRepository) DeleteByRoom(ctx context.Context, roomID string) error {
	return r.db.WithContext(ctx).Where("room_id = ?", roomID).Delete(&model.Bill{}).Error
}

func (r *billRepository) ConsolidateBills(ctx context.Context, roomID string) (*model.Consolidation, error) {
	// Create empty struct as fields will be auto populated by DB
	consolidation := model.Consolidation{}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.db.
			WithContext(ctx).
			Model(&model.Consolidation{}).
			Create(&consolidation).Error; err != nil {
			return err
		}

		if err := r.db.WithContext(ctx).Table("bills").
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

func (r *billRepository) FindByConsolidation(ctx context.Context, consolidationID uint) ([]model.Bill, error) {
	var bills []model.Bill
	err := r.db.
		WithContext(ctx).
		Model(&model.Bill{}).
		Preload("Payers", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Where("consolidation_id = ?", consolidationID).
		Find(&bills).Error
	return bills, err
}
