package postgres

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// Transaction wraps database operations in a transaction with SERIALIZABLE isolation
// This prevents race conditions during concurrent tenant registration (FR-024)
func Transaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	// Begin transaction with SERIALIZABLE isolation level
	tx := db.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if tx.Error != nil {
		return tx.Error
	}

	// Execute the function
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}
