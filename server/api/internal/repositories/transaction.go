package repositories

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	WithTx(tx *gorm.DB) TransactionRepository

	Create(ctx context.Context, transactions []models.Transaction) error
	FindByID(ctx context.Context, transactionID string) (*models.Transaction, error)
	FindByUser(ctx context.Context, isPaid bool, userID string) ([]models.Transaction, error)
	Update(ctx context.Context, transaction *models.Transaction) error
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

func (r *transactionRepository) Create(ctx context.Context, transactions []models.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&transactions).Error
}

// TODO: Implement pagination
func (r *transactionRepository) FindByUser(ctx context.Context, isPaid bool, userID string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.
		WithContext(ctx).
		Where("transactions.is_paid = ? AND (transactions.payee_id = ? OR transactions.payer_id = ?)", isPaid, userID, userID).
		Joins("Payee", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Joins("Payer", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "picture_url")
		}).
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *transactionRepository) FindByID(ctx context.Context, transactionID string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.WithContext(ctx).First(&transaction, transactionID).Error; err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}
