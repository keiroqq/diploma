package items

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keiro/content-digest/backend/internal/models"
)

func TestApplyRulesFiltersAndScoresItems(t *testing.T) {
	sourceID := uuid.New()
	otherSourceID := uuid.New()
	items := []models.FeedItem{
		{
			ID:          uuid.New(),
			SourceID:    sourceID,
			Title:       "React hooks in Go service",
			Excerpt:     "Practical backend notes",
			Author:      "Alice",
			PublishedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
			Source:      models.Source{ID: sourceID, Name: "Habr"},
			Tags:        []models.Tag{{Name: "Go"}},
			Categories:  []models.Category{{Name: "Backend", Slug: "backend"}},
		},
		{
			ID:          uuid.New(),
			SourceID:    sourceID,
			Title:       "Go casino tricks",
			Excerpt:     "Should be excluded",
			Author:      "Bob",
			PublishedAt: time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC),
			Source:      models.Source{ID: sourceID, Name: "Habr"},
			Tags:        []models.Tag{{Name: "Go"}},
			Categories:  []models.Category{{Name: "Backend", Slug: "backend"}},
		},
		{
			ID:          uuid.New(),
			SourceID:    otherSourceID,
			Title:       "Plain frontend note",
			Excerpt:     "No matching include tag",
			Author:      "Carol",
			PublishedAt: time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
			Source:      models.Source{ID: otherSourceID, Name: "News"},
			Tags:        []models.Tag{{Name: "Frontend"}},
		},
	}
	rules := []models.FilterRule{
		{RuleType: models.RuleInclude, TargetType: models.TargetTag, Value: "go", Weight: 2},
		{RuleType: models.RuleExclude, TargetType: models.TargetKeyword, Value: "casino"},
		{RuleType: models.RuleBoost, TargetType: models.TargetKeyword, Value: "react", Weight: 5},
		{RuleType: models.RuleDownrank, TargetType: models.TargetSource, Value: "habr", Weight: 2},
	}

	scored := applyRules(items, rules)

	if len(scored) != 1 {
		t.Fatalf("len(scored) = %d, want 1: %#v", len(scored), scored)
	}
	if scored[0].item.ID != items[0].ID {
		t.Fatalf("matched item = %s, want %s", scored[0].item.ID, items[0].ID)
	}
	if scored[0].score != 5 {
		t.Fatalf("score = %d, want 5", scored[0].score)
	}
}

func TestApplyRulesMatchesNormalizedCategoryAsTag(t *testing.T) {
	itemID := uuid.New()
	items := []models.FeedItem{
		{
			ID:          itemID,
			Title:       "Нейросети в продуктах",
			PublishedAt: time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
			Categories:  []models.Category{{Name: "AI", Slug: "ai"}},
		},
	}
	rules := []models.FilterRule{
		{RuleType: models.RuleInclude, TargetType: models.TargetTag, Value: "ai", Weight: 7},
	}

	scored := applyRules(items, rules)

	if len(scored) != 1 {
		t.Fatalf("len(scored) = %d, want 1", len(scored))
	}
	if scored[0].item.ID != itemID {
		t.Fatalf("matched item = %s, want %s", scored[0].item.ID, itemID)
	}
	if scored[0].score != 7 {
		t.Fatalf("score = %d, want 7", scored[0].score)
	}
}

func TestParseListQuery(t *testing.T) {
	cursor := "2026-05-19T10:30:00Z"
	query, err := ParseListQuery(map[string][]string{
		"mode":   {"archive"},
		"limit":  {"50"},
		"cursor": {cursor},
	})
	if err != nil {
		t.Fatalf("ParseListQuery returned error: %v", err)
	}

	if query.Mode != ModeArchive {
		t.Fatalf("Mode = %q, want %q", query.Mode, ModeArchive)
	}
	if query.Limit != 50 {
		t.Fatalf("Limit = %d, want 50", query.Limit)
	}
	if query.Cursor == nil || query.Cursor.Format(time.RFC3339) != cursor {
		t.Fatalf("Cursor = %v, want %s", query.Cursor, cursor)
	}
}

func TestParseListQueryCategory(t *testing.T) {
	query, err := ParseListQuery(map[string][]string{
		"category": {"Artificial Intelligence"},
	})
	if err != nil {
		t.Fatalf("ParseListQuery returned error: %v", err)
	}
	if query.Category != "artificial-intelligence" {
		t.Fatalf("Category = %q, want artificial-intelligence", query.Category)
	}
}
