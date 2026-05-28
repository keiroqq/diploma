package rss

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/fetch"
	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	db              *gorm.DB
	parser          *Parser
	normalizer      *Normalizer
	refreshCooldown time.Duration
	logger          *slog.Logger
}

type RefreshResult struct {
	Sources      []SourceRefreshResult `json:"sources"`
	ItemsFound   int                   `json:"items_found"`
	ItemsCreated int                   `json:"items_created"`
	ItemsSkipped int                   `json:"items_skipped"`
	Errors       []string              `json:"errors,omitempty"`
}

type SourceRefreshResult struct {
	SourceID      uuid.UUID  `json:"source_id"`
	FeedURL       string     `json:"feed_url"`
	ItemsFound    int        `json:"items_found"`
	ItemsCreated  int        `json:"items_created"`
	ItemsSkipped  int        `json:"items_skipped"`
	Skipped       bool       `json:"skipped"`
	Reason        string     `json:"reason,omitempty"`
	Error         string     `json:"error,omitempty"`
	LastFetchedAt *time.Time `json:"last_fetched_at,omitempty"`
}

type PreviewItemsResponse struct {
	SourceID uuid.UUID             `json:"source_id"`
	FeedURL  string                `json:"feed_url"`
	Items    []PreviewItemResponse `json:"items"`
}

type PreviewItemResponse struct {
	ID            string    `json:"id"`
	SourceID      uuid.UUID `json:"source_id"`
	SourceName    string    `json:"source_name"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	Excerpt       string    `json:"excerpt"`
	ImageURL      string    `json:"image_url"`
	Author        string    `json:"author"`
	PublishedAt   time.Time `json:"published_at"`
	Tags          []string  `json:"tags"`
	Categories    []string  `json:"categories"`
	CategorySlugs []string  `json:"category_slugs"`
	SearchText    string    `json:"search_text"`
}

func NewService(db *gorm.DB, refreshCooldown time.Duration, logger *slog.Logger) *Service {
	cleaner := NewCleaner()
	return &Service{
		db:              db,
		parser:          NewParser(fetch.NewSafeHTTPClient(fetch.DefaultTimeout, fetch.DefaultMaxResponseBytes)),
		normalizer:      NewNormalizer(cleaner),
		refreshCooldown: refreshCooldown,
		logger:          logger,
	}
}

func (s *Service) RefreshSource(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (any, error) {
	source, err := s.getAccessibleSource(ctx, sourceID, userID)
	if err != nil {
		return nil, err
	}

	result, err := s.refreshSourceRecord(ctx, source)
	if err != nil {
		return nil, err
	}
	return aggregateSingle(result), nil
}

func (s *Service) PreviewSourceItems(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (any, error) {
	source, err := s.getAccessibleSource(ctx, sourceID, userID)
	if err != nil {
		return nil, err
	}
	if source.Type != models.SourceTypeRSS {
		return nil, httpx.ErrInvalidInput
	}
	if source.Status == models.SourceStatusDisabled {
		return nil, httpx.ErrInvalidInput
	}

	feed, err := s.parser.ParseURL(ctx, source.FeedURL)
	if err != nil {
		s.markSourceStatus(ctx, source.ID, models.SourceStatusError, nil)
		return nil, err
	}

	items := make([]PreviewItemResponse, 0, len(feed.Items))
	for _, item := range feed.Items {
		normalized, tags, err := s.normalizer.Normalize(source.ID, item)
		if err != nil {
			return nil, err
		}

		categorySlugs := categorySlugsForTags(tags)
		categoryNames, err := s.categoryNamesBySlug(ctx, categorySlugs)
		if err != nil {
			return nil, err
		}

		items = append(items, previewItemResponse(*source, normalized, tags, categorySlugs, categoryNames))
	}

	now := time.Now()
	s.markSourceStatus(ctx, source.ID, models.SourceStatusActive, &now)

	return &PreviewItemsResponse{
		SourceID: source.ID,
		FeedURL:  source.FeedURL,
		Items:    items,
	}, nil
}

func (s *Service) RefreshFeed(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (any, error) {
	exists, err := s.feedExistsForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httpx.ErrNotFound
	}

	var links []models.FeedSource
	if err := s.db.WithContext(ctx).
		Joins("JOIN feeds ON feeds.id = feed_sources.feed_id").
		Preload("Source").
		Where("feed_sources.feed_id = ? AND feeds.user_id = ? AND feed_sources.is_enabled = true", feedID, userID).
		Order("feed_sources.priority DESC, feed_sources.created_at ASC").
		Find(&links).Error; err != nil {
		return nil, err
	}

	result := &RefreshResult{Sources: make([]SourceRefreshResult, 0, len(links))}
	for _, link := range links {
		sourceResult, err := s.refreshSourceRecord(ctx, &link.Source)
		if err != nil {
			sourceResult = SourceRefreshResult{
				SourceID: link.SourceID,
				FeedURL:  link.Source.FeedURL,
				Error:    err.Error(),
			}
			result.Errors = append(result.Errors, err.Error())
		}

		result.Sources = append(result.Sources, sourceResult)
		result.ItemsFound += sourceResult.ItemsFound
		result.ItemsCreated += sourceResult.ItemsCreated
		result.ItemsSkipped += sourceResult.ItemsSkipped
	}

	return result, nil
}

func (s *Service) refreshSourceRecord(ctx context.Context, source *models.Source) (SourceRefreshResult, error) {
	result := SourceRefreshResult{
		SourceID:      source.ID,
		FeedURL:       source.FeedURL,
		LastFetchedAt: source.LastFetchedAt,
	}

	if source.Type != models.SourceTypeRSS {
		result.Skipped = true
		result.Reason = "only rss sources are supported in MVP"
		return result, nil
	}
	if source.StorageMode == models.SourceStorageLocal {
		result.Skipped = true
		result.Reason = "local storage source"
		return result, nil
	}
	if source.Status == models.SourceStatusDisabled {
		result.Skipped = true
		result.Reason = "source is disabled"
		return result, nil
	}
	if source.LastFetchedAt != nil && s.refreshCooldown > 0 && time.Since(*source.LastFetchedAt) < s.refreshCooldown {
		result.Skipped = true
		result.Reason = "source refreshed recently"
		return result, nil
	}

	s.logger.Info("rss refresh started", "source_id", source.ID.String(), "feed_url", source.FeedURL)

	feed, err := s.parser.ParseURL(ctx, source.FeedURL)
	if err != nil {
		result.Error = err.Error()
		s.markSourceStatus(ctx, source.ID, models.SourceStatusError, nil)
		s.logger.Error("rss refresh failed", "source_id", source.ID.String(), "feed_url", source.FeedURL, "error", err)
		return result, err
	}

	result.ItemsFound = len(feed.Items)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range feed.Items {
			normalized, tags, err := s.normalizer.Normalize(source.ID, item)
			if err != nil {
				return err
			}

			created, err := s.createItemIfNew(tx, &normalized, tags)
			if err != nil {
				return err
			}
			if created {
				result.ItemsCreated++
			} else {
				result.ItemsSkipped++
			}
		}

		now := time.Now()
		return tx.Model(&models.Source{}).
			Where("id = ?", source.ID).
			Updates(map[string]any{
				"last_fetched_at": now,
				"status":          models.SourceStatusActive,
			}).Error
	})
	if err != nil {
		result.Error = err.Error()
		s.markSourceStatus(ctx, source.ID, models.SourceStatusError, nil)
		return result, err
	}

	now := time.Now()
	result.LastFetchedAt = &now
	s.logger.Info(
		"rss refresh finished",
		"source_id", source.ID.String(),
		"feed_url", source.FeedURL,
		"items_found", result.ItemsFound,
		"items_created", result.ItemsCreated,
		"items_skipped", result.ItemsSkipped,
	)

	return result, nil
}

func (s *Service) createItemIfNew(tx *gorm.DB, item *models.FeedItem, tags []string) (bool, error) {
	var count int64
	query := tx.Model(&models.FeedItem{}).Where("source_id = ?", item.SourceID)

	conditions := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if item.GUID != nil && strings.TrimSpace(*item.GUID) != "" {
		conditions = append(conditions, "guid = ?")
		args = append(args, *item.GUID)
	}
	if item.CanonicalURL != nil && strings.TrimSpace(*item.CanonicalURL) != "" {
		conditions = append(conditions, "canonical_url = ?")
		args = append(args, *item.CanonicalURL)
	}
	if item.ContentHash != "" {
		conditions = append(conditions, "content_hash = ?")
		args = append(args, item.ContentHash)
	}
	if len(conditions) > 0 {
		if err := query.Where(strings.Join(conditions, " OR "), args...).Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return false, nil
		}
	}

	if err := tx.Create(item).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return false, nil
		}
		return false, err
	}

	for _, tagName := range tags {
		tag, err := findOrCreateTag(tx, tagName)
		if err != nil {
			return false, err
		}
		if err := tx.Exec(
			"INSERT INTO feed_item_tags (item_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			item.ID,
			tag.ID,
		).Error; err != nil {
			return false, err
		}
	}
	if err := linkCategoriesForTags(tx, item.ID, tags); err != nil {
		return false, err
	}

	return true, nil
}

func findOrCreateTag(tx *gorm.DB, name string) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	tagSlug := normalizeTagSlug(name)

	var tag models.Tag
	err := tx.
		Where("slug = ?", tagSlug).
		Attrs(models.Tag{
			ID:   uuid.New(),
			Name: name,
			Slug: tagSlug,
		}).
		FirstOrCreate(&tag).Error
	if err == nil {
		return &tag, nil
	}

	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return nil, err
	}

	if err := tx.Where("slug = ?", tagSlug).First(&tag).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}

func linkCategoriesForTags(tx *gorm.DB, itemID uuid.UUID, tags []string) error {
	slugs := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tagName := range tags {
		tagSlug := normalizeTagSlug(tagName)
		if tagSlug == "" {
			continue
		}
		if _, ok := seen[tagSlug]; ok {
			continue
		}
		seen[tagSlug] = struct{}{}
		slugs = append(slugs, tagSlug)
	}
	if len(slugs) == 0 {
		return nil
	}

	var aliases []models.TagAlias
	if err := tx.
		Where("raw_tag_slug IN ? AND provider IN ?", slugs, []string{"any", "habr"}).
		Find(&aliases).Error; err != nil {
		return err
	}
	categoryIDs := map[uuid.UUID]struct{}{}
	for _, alias := range aliases {
		categoryIDs[alias.CategoryID] = struct{}{}
	}

	categorySlugs := categorySlugsForTags(tags)
	if len(categorySlugs) > 0 {
		var categories []models.Category
		if err := tx.Where("slug IN ?", categorySlugs).Find(&categories).Error; err != nil {
			return err
		}
		for _, category := range categories {
			categoryIDs[category.ID] = struct{}{}
		}
	}

	for categoryID := range categoryIDs {
		if err := tx.Exec(
			"INSERT INTO feed_item_categories (item_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			itemID,
			categoryID,
		).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) categoryNamesBySlug(ctx context.Context, slugs []string) (map[string]string, error) {
	if len(slugs) == 0 {
		return map[string]string{}, nil
	}

	var categories []models.Category
	if err := s.db.WithContext(ctx).Where("slug IN ?", slugs).Find(&categories).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string, len(categories))
	for _, category := range categories {
		result[category.Slug] = category.Name
	}
	return result, nil
}

func previewItemResponse(source models.Source, item models.FeedItem, tags []string, categorySlugs []string, categoryNames map[string]string) PreviewItemResponse {
	categories := make([]string, 0, len(categorySlugs))
	for _, slug := range categorySlugs {
		if name := categoryNames[slug]; name != "" {
			categories = append(categories, name)
		}
	}

	return PreviewItemResponse{
		ID:            "local:" + item.ContentHash,
		SourceID:      item.SourceID,
		SourceName:    source.Name,
		Title:         item.Title,
		URL:           item.URL,
		Excerpt:       item.Excerpt,
		ImageURL:      item.ImageURL,
		Author:        item.Author,
		PublishedAt:   item.PublishedAt,
		Tags:          tags,
		Categories:    categories,
		CategorySlugs: categorySlugs,
		SearchText: strings.ToLower(strings.Join([]string{
			item.Title,
			item.Excerpt,
			item.Author,
			source.Name,
			strings.Join(tags, " "),
			strings.Join(categories, " "),
		}, " ")),
	}
}

func normalizeTagSlug(name string) string {
	tagSlug := slug.Make(strings.TrimSpace(name))
	if tagSlug == "" {
		tagSlug = slug.Make(strings.ToLower(strings.TrimSpace(name)))
	}
	return tagSlug
}

func (s *Service) getAccessibleSource(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (*models.Source, error) {
	var source models.Source
	err := s.db.WithContext(ctx).
		Where("id = ? AND (is_public = true OR created_by = ?)", sourceID, userID).
		First(&source).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (s *Service) feedExistsForUser(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Table("feeds").
		Where("id = ? AND user_id = ?", feedID, userID).
		Count(&count).Error
	return count > 0, err
}

func (s *Service) markSourceStatus(ctx context.Context, sourceID uuid.UUID, status string, fetchedAt *time.Time) {
	updates := map[string]any{"status": status}
	if fetchedAt != nil {
		updates["last_fetched_at"] = *fetchedAt
	}
	if err := s.db.WithContext(ctx).Model(&models.Source{}).Where("id = ?", sourceID).Updates(updates).Error; err != nil {
		s.logger.Error("failed to update source status", "source_id", sourceID.String(), "error", err)
	}
}

func aggregateSingle(sourceResult SourceRefreshResult) *RefreshResult {
	result := &RefreshResult{
		Sources:      []SourceRefreshResult{sourceResult},
		ItemsFound:   sourceResult.ItemsFound,
		ItemsCreated: sourceResult.ItemsCreated,
		ItemsSkipped: sourceResult.ItemsSkipped,
	}
	if sourceResult.Error != "" {
		result.Errors = []string{sourceResult.Error}
	}
	return result
}
