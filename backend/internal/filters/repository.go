package filters

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

func (r *Repository) ListByFeedForUser(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) ([]models.FilterRule, error) {
	var rules []models.FilterRule
	err := r.db.WithContext(ctx).
		Joins("JOIN feeds ON feeds.id = filter_rules.feed_id").
		Where("filter_rules.feed_id = ? AND feeds.user_id = ?", feedID, userID).
		Order("filter_rules.created_at ASC").
		Find(&rules).Error
	return rules, err
}

func (r *Repository) FeedExistsForUser(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("feeds").
		Where("id = ? AND user_id = ?", feedID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetByIDForUser(ctx context.Context, ruleID uuid.UUID, userID uuid.UUID) (*models.FilterRule, error) {
	var rule models.FilterRule
	err := r.db.WithContext(ctx).
		Joins("JOIN feeds ON feeds.id = filter_rules.feed_id").
		Where("filter_rules.id = ? AND feeds.user_id = ?", ruleID, userID).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *Repository) Create(ctx context.Context, rule *models.FilterRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *Repository) Update(ctx context.Context, rule *models.FilterRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *Repository) Delete(ctx context.Context, ruleID uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND feed_id IN (SELECT id FROM feeds WHERE user_id = ?)", ruleID, userID).
		Delete(&models.FilterRule{}).Error
}
