package sources

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAccessible(ctx context.Context, userID uuid.UUID) ([]models.Source, error) {
	var sources []models.Source
	err := r.db.WithContext(ctx).
		Where("is_public = true OR created_by = ?", userID).
		Order("created_at DESC").
		Find(&sources).Error
	return sources, err
}

func (r *Repository) GetAccessible(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (*models.Source, error) {
	var source models.Source
	err := r.db.WithContext(ctx).
		Where("id = ? AND (is_public = true OR created_by = ?)", sourceID, userID).
		First(&source).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *Repository) Create(ctx context.Context, source *models.Source) error {
	return r.db.WithContext(ctx).Create(source).Error
}

func (r *Repository) Update(ctx context.Context, source *models.Source) error {
	return r.db.WithContext(ctx).Save(source).Error
}

func (r *Repository) Delete(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND created_by = ?", sourceID, userID).
		Delete(&models.Source{}).Error
}
