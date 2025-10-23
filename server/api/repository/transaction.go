package repository

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	WithTx(tx *gorm.DB) TransactionRepository

	Create(ctx context.Context, transactions []model.Transaction) error
	FindByUser(ctx context.Context, isPaid bool, userID string) ([]model.Transaction, error)
	FindByID(ctx context.Context, transactionID string) (*model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// WithTx returns a new TransactionRepository with the provided transaction
func (r *transactionRepository) WithTx(tx *gorm.DB) TransactionRepository {
	if tx == nil {
		return r
	}
	return &transactionRepository{db: tx}
}

func (r *transactionRepository) Create(ctx context.Context, transactions []model.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// return r.db.Omit("Consolidation").Create(&transactions).Error
	return r.db.WithContext(ctx).Create(&transactions).Error
}

// TODO: Implement pagination
func (r *transactionRepository) FindByUser(ctx context.Context, isPaid bool, userID string) ([]model.Transaction, error) {
	var transactions []model.Transaction
	err := r.db.
		WithContext(ctx).
		Where("is_paid = ? AND (payee_id = ? OR payer_id = ?)", isPaid, userID, userID).
		Preload("Payee", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Preload("Payer", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) FindByID(ctx context.Context, transactionID string) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.WithContext(ctx).First(&transaction, transactionID).Error
	return &transaction, err
}

func (r *transactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}
