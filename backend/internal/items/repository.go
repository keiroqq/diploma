package items

import (
	"context"
	"strings"
	"time"

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

func (r *Repository) ListCandidates(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, query ListQuery, fetchLimit int) ([]models.FeedItem, error) {
	startOfToday := startOfDay(time.Now())
	db := r.db.WithContext(ctx).
		Model(&models.FeedItem{}).
		Distinct("feed_items.*").
		Joins("JOIN feed_sources ON feed_sources.source_id = feed_items.source_id").
		Joins("JOIN feeds ON feeds.id = feed_sources.feed_id").
		Preload("Source").
		Preload("Tags").
		Preload("Categories").
		Where("feed_sources.feed_id = ? AND feeds.user_id = ? AND feed_sources.is_enabled = true", feedID, userID)

	if query.DateFrom != nil || query.DateTo != nil {
		if query.DateFrom != nil {
			db = db.Where("feed_items.published_at >= ?", *query.DateFrom)
		}
		if query.DateTo != nil {
			db = db.Where("feed_items.published_at < ?", *query.DateTo)
		}
	} else if query.Mode == ModeToday {
		db = db.Where("feed_items.published_at >= ?", startOfToday)
	} else if query.Mode == ModeArchive {
		db = db.Where("feed_items.published_at < ?", startOfToday)
		if query.Cursor != nil {
			db = db.Where("feed_items.published_at < ?", *query.Cursor)
		}
	}
	if len(query.Categories) > 0 {
		db = db.
			Joins("JOIN feed_item_categories ON feed_item_categories.item_id = feed_items.id").
			Joins("JOIN categories ON categories.id = feed_item_categories.category_id").
			Where("categories.slug IN ?", query.Categories)
	}

	var records []models.FeedItem
	err := db.Order("feed_items.published_at DESC").
		Limit(fetchLimit).
		Find(&records).Error
	return records, err
}

func (r *Repository) ListActiveRules(ctx context.Context, feedID uuid.UUID) ([]models.FilterRule, error) {
	var rules []models.FilterRule
	err := r.db.WithContext(ctx).
		Where("feed_id = ? AND is_active = true", feedID).
		Order("created_at ASC").
		Find(&rules).Error
	return rules, err
}

func (r *Repository) SavedMap(ctx context.Context, userID uuid.UUID, itemIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool, len(itemIDs))
	if len(itemIDs) == 0 {
		return result, nil
	}

	var records []models.SavedItem
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND item_id IN ?", userID, itemIDs).
		Find(&records).Error; err != nil {
		return nil, err
	}
	for _, record := range records {
		result[record.ItemID] = true
	}
	return result, nil
}

func (r *Repository) ItemExists(ctx context.Context, itemID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.FeedItem{}).
		Where("id = ?", itemID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetAccessibleItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (models.FeedItem, error) {
	var record models.FeedItem
	err := r.db.WithContext(ctx).
		Preload("Source").
		Preload("Tags").
		Preload("Categories").
		Where("feed_items.id = ?", itemID).
		Where(`
			(
				EXISTS (
					SELECT 1
					FROM feed_sources
					JOIN feeds ON feeds.id = feed_sources.feed_id
					WHERE feed_sources.source_id = feed_items.source_id
						AND feed_sources.is_enabled = true
						AND feeds.user_id = ?
				)
				OR EXISTS (
					SELECT 1
					FROM saved_items
					WHERE saved_items.item_id = feed_items.id
						AND saved_items.user_id = ?
				)
			)
		`, userID, userID).
		First(&record).Error
	return record, err
}

func (r *Repository) UpdateContentHTML(ctx context.Context, itemID uuid.UUID, contentHTML string) error {
	return r.db.WithContext(ctx).
		Model(&models.FeedItem{}).
		Where("id = ?", itemID).
		Update("content_html", contentHTML).Error
}

func (r *Repository) Save(ctx context.Context, item models.SavedItem) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&item).Error
}

func (r *Repository) Unsave(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND item_id = ?", userID, itemID).
		Delete(&models.SavedItem{}).Error
}

func (r *Repository) ListSaved(ctx context.Context, userID uuid.UUID, limit int) ([]models.SavedItem, error) {
	var records []models.SavedItem
	err := r.db.WithContext(ctx).
		Preload("Item.Source").
		Preload("Item.Tags").
		Preload("Item.Categories").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *Repository) SearchAccessibleItems(ctx context.Context, userID uuid.UUID, query SearchQuery) ([]models.FeedItem, error) {
	searchPattern := "%" + strings.ToLower(query.Query) + "%"
	searchCondition := `(
		LOWER(feed_items.title) LIKE ? OR
		LOWER(feed_items.excerpt) LIKE ? OR
		LOWER(feed_items.content_html) LIKE ? OR
		LOWER(feed_items.author) LIKE ? OR
		LOWER(sources.name) LIKE ? OR
		LOWER(tags.name) LIKE ? OR
		LOWER(categories.name) LIKE ?
	)`
	args := []any{
		searchPattern,
		searchPattern,
		searchPattern,
		searchPattern,
		searchPattern,
		searchPattern,
		searchPattern,
	}

	db := r.db.WithContext(ctx).
		Model(&models.FeedItem{}).
		Distinct("feed_items.*").
		Joins("JOIN feed_sources ON feed_sources.source_id = feed_items.source_id").
		Joins("JOIN feeds ON feeds.id = feed_sources.feed_id").
		Joins("JOIN sources ON sources.id = feed_items.source_id").
		Joins("LEFT JOIN feed_item_tags ON feed_item_tags.item_id = feed_items.id").
		Joins("LEFT JOIN tags ON tags.id = feed_item_tags.tag_id").
		Joins("LEFT JOIN feed_item_categories ON feed_item_categories.item_id = feed_items.id").
		Joins("LEFT JOIN categories ON categories.id = feed_item_categories.category_id").
		Preload("Source").
		Preload("Tags").
		Preload("Categories").
		Where("feeds.user_id = ? AND feed_sources.is_enabled = true", userID).
		Where(searchCondition, args...)

	if query.FeedID != nil {
		db = db.Where("feed_sources.feed_id = ?", *query.FeedID)
	}

	var records []models.FeedItem
	err := db.Order("feed_items.published_at DESC").
		Limit(query.Limit).
		Find(&records).Error
	return records, err
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
