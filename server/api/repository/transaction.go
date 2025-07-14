package repository

import (
	"github.com/RowenTey/JustJio/server/api/model"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	WithTx(tx *gorm.DB) TransactionRepository

	Create(transactions *[]model.Transaction) error
	FindByUser(isPaid bool, userID string) (*[]model.Transaction, error)
	FindByID(transactionID string) (*model.Transaction, error)
	Update(transaction *model.Transaction) error
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

func (r *transactionRepository) Create(transactions *[]model.Transaction) error {
	return r.db.Omit("Consolidation").Create(&transactions).Error
}

// TODO: Implement pagination
func (r *transactionRepository) FindByUser(isPaid bool, userID string) (*[]model.Transaction, error) {
	var transactions []model.Transaction
	err := r.db.
		Where("is_paid = ? AND (payee_id = ? OR payer_id = ?)", isPaid, userID, userID).
		Preload("Payee").
		Preload("Payer").
		Find(&transactions).Error
	return &transactions, err
}

func (r *transactionRepository) FindByID(transactionID string) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.First(&transaction, transactionID).Error
	return &transaction, err
}

func (r *transactionRepository) Update(transaction *model.Transaction) error {
	return r.db.Save(transaction).Error
}
