package catalog

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

func (r *Repository) FeedExistsForUser(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("feeds").
		Where("id = ? AND user_id = ?", feedID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) FindPublicSourceByCatalogPage(ctx context.Context, pageURL string) (*models.Source, error) {
	var source models.Source
	result := r.db.WithContext(ctx).
		Where("is_public = true AND url = ?", pageURL).
		Limit(1).
		Find(&source)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &source, nil
}

func (r *Repository) CreateSource(ctx context.Context, source *models.Source) error {
	return r.db.WithContext(ctx).Create(source).Error
}

func (r *Repository) UpdateSourceFeedURL(ctx context.Context, sourceID uuid.UUID, feedURL string) error {
	return r.db.WithContext(ctx).
		Model(&models.Source{}).
		Where("id = ?", sourceID).
		Update("feed_url", feedURL).Error
}

func (r *Repository) UpsertFeedSource(ctx context.Context, link *models.FeedSource) (*models.FeedSource, error) {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "feed_id"}, {Name: "source_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"is_enabled": true,
				"priority":   link.Priority,
			}),
		}).
		Create(link).Error; err != nil {
		return nil, err
	}

	var stored models.FeedSource
	if err := r.db.WithContext(ctx).
		Where("feed_id = ? AND source_id = ?", link.FeedID, link.SourceID).
		First(&stored).Error; err != nil {
		return nil, err
	}
	return &stored, nil
}
