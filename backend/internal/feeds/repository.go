package feeds

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Feed, error) {
	var feeds []models.Feed
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at ASC").
		Find(&feeds).Error
	return feeds, err
}

func (r *Repository) GetByIDForUser(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (*models.Feed, error) {
	var feed models.Feed
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", feedID, userID).
		First(&feed).Error
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *Repository) Create(ctx context.Context, feed *models.Feed) error {
	return r.db.WithContext(ctx).Create(feed).Error
}

func (r *Repository) Update(ctx context.Context, feed *models.Feed) error {
	return r.db.WithContext(ctx).Save(feed).Error
}

func (r *Repository) Delete(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", feedID, userID).
		Delete(&models.Feed{}).Error
}

func (r *Repository) AddSource(ctx context.Context, link *models.FeedSource) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "feed_id"}, {Name: "source_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"is_enabled": true,
				"priority":   link.Priority,
			}),
		}).
		Create(link).Error
}

func (r *Repository) SourceAccessible(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("sources").
		Where("id = ? AND (is_public = true OR created_by = ?)", sourceID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) RemoveSource(ctx context.Context, feedID uuid.UUID, sourceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("feed_id = ? AND source_id = ?", feedID, sourceID).
		Delete(&models.FeedSource{}).Error
}

func (r *Repository) ListEnabledSources(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) ([]models.FeedSource, error) {
	var links []models.FeedSource
	err := r.db.WithContext(ctx).
		Joins("JOIN feeds ON feeds.id = feed_sources.feed_id").
		Preload("Source").
		Where("feed_sources.feed_id = ? AND feeds.user_id = ? AND feed_sources.is_enabled = true", feedID, userID).
		Order("feed_sources.priority DESC, feed_sources.created_at ASC").
		Find(&links).Error
	return links, err
}
