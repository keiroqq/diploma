package db

import (
	"context"

	"gorm.io/gorm"
)

func WithTransaction(ctx context.Context, database *gorm.DB, fn func(tx *gorm.DB) error) error {
	return database.WithContext(ctx).Transaction(fn)
}
