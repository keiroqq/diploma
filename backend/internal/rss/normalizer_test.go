package rss

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
)

func TestNormalizerExtractsCleanExcerptImageAndTags(t *testing.T) {
	normalizer := NewNormalizer(NewCleaner())
	sourceID := uuid.New()
	publishedAt := time.Date(2026, 5, 19, 10, 30, 0, 0, time.UTC)

	item, tags, err := normalizer.Normalize(sourceID, &gofeed.Item{
		GUID:            "habr-123",
		Title:           "  Go RSS pipeline  ",
		Link:            "https://example.com/posts/go-rss",
		Description:     `<p>Полезное описание &gt;&gt;Читать&gt;&gt;</p><img src="https://cdn.example.com/cover.png">`,
		Content:         "<p>Full content</p>",
		PublishedParsed: &publishedAt,
		Author:          &gofeed.Person{Name: "Alice"},
		Categories:      []string{"Go", "RSS", "go", " "},
	})
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}

	if item.SourceID != sourceID {
		t.Fatalf("SourceID = %s, want %s", item.SourceID, sourceID)
	}
	if item.Title != "Go RSS pipeline" {
		t.Fatalf("Title = %q", item.Title)
	}
	if item.GUID == nil || *item.GUID != "habr-123" {
		t.Fatalf("GUID = %v", item.GUID)
	}
	if item.CanonicalURL == nil || *item.CanonicalURL != "https://example.com/posts/go-rss" {
		t.Fatalf("CanonicalURL = %v", item.CanonicalURL)
	}
	if item.Excerpt != "Полезное описание" {
		t.Fatalf("Excerpt = %q", item.Excerpt)
	}
	if item.ImageURL != "https://cdn.example.com/cover.png" {
		t.Fatalf("ImageURL = %q", item.ImageURL)
	}
	if !item.PublishedAt.Equal(publishedAt) {
		t.Fatalf("PublishedAt = %s, want %s", item.PublishedAt, publishedAt)
	}
	if item.Author != "Alice" {
		t.Fatalf("Author = %q", item.Author)
	}
	if len(tags) != 2 || tags[0] != "Go" || tags[1] != "RSS" {
		t.Fatalf("tags = %#v", tags)
	}
	if len(item.RawData) == 0 {
		t.Fatal("RawData is empty")
	}
}

func TestNormalizerFallsBackToLinkAndUntitled(t *testing.T) {
	normalizer := NewNormalizer(NewCleaner())

	item, _, err := normalizer.Normalize(uuid.New(), &gofeed.Item{
		Link:        "https://example.com/posts/untitled",
		Description: "<p>Text only</p>",
	})
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}

	if item.Title != "Без заголовка" {
		t.Fatalf("Title = %q", item.Title)
	}
	if item.GUID == nil || *item.GUID != "https://example.com/posts/untitled" {
		t.Fatalf("GUID fallback = %v", item.GUID)
	}
	if item.ExternalID == "" {
		t.Fatal("ExternalID is empty")
	}
}
