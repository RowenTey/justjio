package database

import (
	"database/sql"

	"gorm.io/gorm"
)

// RunInTransaction executes the provided function within a database transaction.
func RunInTransaction(db *gorm.DB, isolationLevel sql.IsolationLevel, fn func(tx *gorm.DB) error) error {
	tx := db.Begin(&sql.TxOptions{
		Isolation: isolationLevel,
	})
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
