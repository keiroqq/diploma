package items

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"

	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	repo *Repository
}

type scoredItem struct {
	item  models.FeedItem
	score int
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListFeedItems(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, query ListQuery) (*FeedItemsResponse, error) {
	if query.Mode == "" {
		query.Mode = ModeToday
	}
	if query.Mode != ModeToday && query.Mode != ModeArchive {
		return nil, httpx.ErrInvalidInput
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	exists, err := s.repo.FeedExistsForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httpx.ErrNotFound
	}

	fetchLimit := query.Limit * 10
	if fetchLimit < 100 {
		fetchLimit = 100
	}
	if fetchLimit > 500 {
		fetchLimit = 500
	}

	candidates, err := s.repo.ListCandidates(ctx, feedID, userID, query, fetchLimit)
	if err != nil {
		return nil, err
	}

	rules, err := s.repo.ListActiveRules(ctx, feedID)
	if err != nil {
		return nil, err
	}

	scored := applyRules(candidates, rules)
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].item.PublishedAt.After(scored[j].item.PublishedAt)
	})

	hasMore := len(scored) > query.Limit
	if hasMore {
		scored = scored[:query.Limit]
	}

	itemIDs := make([]uuid.UUID, 0, len(scored))
	for _, record := range scored {
		itemIDs = append(itemIDs, record.item.ID)
	}
	saved, err := s.repo.SavedMap(ctx, userID, itemIDs)
	if err != nil {
		return nil, err
	}

	resp := &FeedItemsResponse{
		Items: make([]ItemResponse, 0, len(scored)),
		Mode:  query.Mode,
	}
	for _, record := range scored {
		resp.Items = append(resp.Items, itemResponse(record.item, record.score, saved[record.item.ID]))
	}
	if query.Mode == ModeArchive && hasMore && len(scored) > 0 {
		resp.NextCursor = scored[len(scored)-1].item.PublishedAt.Format(time.RFC3339)
	}

	return resp, nil
}

func (s *Service) SaveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	exists, err := s.repo.ItemExists(ctx, itemID)
	if err != nil {
		return err
	}
	if !exists {
		return httpx.ErrNotFound
	}

	return s.repo.Save(ctx, models.SavedItem{
		ID:     uuid.New(),
		UserID: userID,
		ItemID: itemID,
	})
}

func (s *Service) UnsaveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	return s.repo.Unsave(ctx, userID, itemID)
}

func (s *Service) ListSaved(ctx context.Context, userID uuid.UUID, limit int) (*SavedItemsResponse, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}

	records, err := s.repo.ListSaved(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	resp := &SavedItemsResponse{Items: make([]ItemResponse, 0, len(records))}
	for _, record := range records {
		resp.Items = append(resp.Items, itemResponse(record.Item, 0, true))
	}
	return resp, nil
}

func applyRules(items []models.FeedItem, rules []models.FilterRule) []scoredItem {
	result := make([]scoredItem, 0, len(items))
	includeRules := make([]models.FilterRule, 0)
	for _, rule := range rules {
		if rule.RuleType == models.RuleInclude {
			includeRules = append(includeRules, rule)
		}
	}

	for _, item := range items {
		score := 0
		excluded := false
		includeMatched := len(includeRules) == 0

		for _, rule := range rules {
			matched := matchesRule(item, rule)
			if !matched {
				continue
			}

			switch rule.RuleType {
			case models.RuleExclude:
				excluded = true
			case models.RuleInclude:
				includeMatched = true
				score += rule.Weight
			case models.RuleBoost:
				if rule.Weight == 0 {
					score++
				} else {
					score += rule.Weight
				}
			case models.RuleDownrank:
				if rule.Weight > 0 {
					score -= rule.Weight
				} else {
					score += rule.Weight
				}
			}
		}

		if excluded || !includeMatched {
			continue
		}
		result = append(result, scoredItem{item: item, score: score})
	}

	return result
}

func matchesRule(item models.FeedItem, rule models.FilterRule) bool {
	value := strings.ToLower(strings.TrimSpace(rule.Value))
	if value == "" {
		return false
	}

	switch rule.TargetType {
	case models.TargetKeyword:
		text := strings.ToLower(item.Title + " " + item.Excerpt + " " + item.ContentHTML)
		return strings.Contains(text, value)
	case models.TargetTag:
		ruleSlug := slug.Make(value)
		for _, tag := range item.Tags {
			tagName := strings.ToLower(tag.Name)
			if tagName == value || strings.Contains(tagName, value) || slug.Make(tag.Name) == ruleSlug {
				return true
			}
		}
		return false
	case models.TargetSource:
		if strings.EqualFold(item.SourceID.String(), value) {
			return true
		}
		return strings.Contains(strings.ToLower(item.Source.Name), value)
	case models.TargetAuthor:
		return strings.Contains(strings.ToLower(item.Author), value)
	default:
		return false
	}
}

func itemResponse(item models.FeedItem, score int, isSaved bool) ItemResponse {
	tags := make([]string, 0, len(item.Tags))
	for _, tag := range item.Tags {
		tags = append(tags, tag.Name)
	}

	return ItemResponse{
		ID:          item.ID,
		SourceID:    item.SourceID,
		SourceName:  item.Source.Name,
		Title:       item.Title,
		URL:         item.URL,
		Excerpt:     item.Excerpt,
		ImageURL:    item.ImageURL,
		Author:      item.Author,
		PublishedAt: item.PublishedAt,
		Tags:        tags,
		Score:       score,
		IsSaved:     isSaved,
	}
}

func ParseListQuery(values map[string][]string) (ListQuery, error) {
	query := ListQuery{Mode: ModeToday, Limit: 20}
	if rawMode := firstQuery(values, "mode"); rawMode != "" {
		query.Mode = ListMode(rawMode)
	}
	if rawLimit := firstQuery(values, "limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return ListQuery{}, err
		}
		query.Limit = limit
	}
	if rawCursor := firstQuery(values, "cursor"); rawCursor != "" {
		cursor, err := time.Parse(time.RFC3339, rawCursor)
		if err != nil {
			return ListQuery{}, err
		}
		query.Cursor = &cursor
	}
	return query, nil
}

func firstQuery(values map[string][]string, key string) string {
	if len(values[key]) == 0 {
		return ""
	}
	return values[key][0]
}
