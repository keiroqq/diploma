package categories

import (
	"context"

	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&categories).Error
	return categories, err
}
