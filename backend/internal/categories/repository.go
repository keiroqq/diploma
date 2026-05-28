package categories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Repository struct {
	db *gorm.DB
}

type feedSourceRef struct {
	Name    string
	URL     string
	FeedURL string
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

func (r *Repository) ListBySlugs(ctx context.Context, slugs []string) ([]models.Category, error) {
	if len(slugs) == 0 {
		return []models.Category{}, nil
	}

	var categories []models.Category
	err := r.db.WithContext(ctx).
		Where("slug IN ?", slugs).
		Order("name ASC").
		Find(&categories).Error
	return categories, err
}

func (r *Repository) FeedExistsForUser(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("feeds").
		Where("id = ? AND user_id = ?", feedID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListFeedSources(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) ([]feedSourceRef, error) {
	var sources []feedSourceRef
	err := r.db.WithContext(ctx).
		Table("sources").
		Select("sources.name, sources.url, sources.feed_url").
		Joins("JOIN feed_sources ON feed_sources.source_id = sources.id").
		Joins("JOIN feeds ON feeds.id = feed_sources.feed_id").
		Where("feed_sources.feed_id = ? AND feeds.user_id = ? AND feed_sources.is_enabled = true", feedID, userID).
		Scan(&sources).Error
	return sources, err
}
