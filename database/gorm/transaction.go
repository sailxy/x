package gorm

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

type Tx struct {
	db *gorm.DB
}

func NewTx(db *gorm.DB) *Tx {
	return &Tx{db: db}
}

// Execute a transaction.
func (t *Tx) Exec(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Set the transaction to the context.
		ctx = context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

// Get the transaction from the context. If not found, return the original database.
func (t *Tx) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return t.db
}
