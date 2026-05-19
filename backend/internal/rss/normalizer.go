package rss

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"gorm.io/datatypes"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Normalizer struct {
	cleaner *Cleaner
}

func NewNormalizer(cleaner *Cleaner) *Normalizer {
	return &Normalizer{cleaner: cleaner}
}

func (n *Normalizer) Normalize(sourceID uuid.UUID, item *gofeed.Item) (models.FeedItem, []string, error) {
	publishedAt := time.Now()
	if item.PublishedParsed != nil {
		publishedAt = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		publishedAt = *item.UpdatedParsed
	}

	guid := strings.TrimSpace(item.GUID)
	link := strings.TrimSpace(item.Link)
	if guid == "" {
		guid = link
	}

	description := strings.TrimSpace(item.Description)
	content := strings.TrimSpace(item.Content)
	excerptSource := description
	if excerptSource == "" {
		excerptSource = content
	}

	imageURL := n.imageURL(item, excerptSource)
	excerpt := n.cleaner.CleanText(excerptSource)
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = "Без заголовка"
	}

	author := ""
	if item.Author != nil {
		author = strings.TrimSpace(item.Author.Name)
	}

	raw, err := json.Marshal(map[string]any{
		"guid":        item.GUID,
		"title":       item.Title,
		"link":        item.Link,
		"published":   item.Published,
		"updated":     item.Updated,
		"author":      author,
		"categories":  item.Categories,
		"description": item.Description,
	})
	if err != nil {
		return models.FeedItem{}, nil, err
	}

	canonicalURL := link
	hash := contentHash(sourceID, guid, canonicalURL, title)

	feedItem := models.FeedItem{
		ID:           uuid.New(),
		SourceID:     sourceID,
		ExternalID:   firstNonEmpty(guid, canonicalURL, hash),
		GUID:         nilIfEmpty(guid),
		Title:        title,
		URL:          canonicalURL,
		CanonicalURL: nilIfEmpty(canonicalURL),
		Excerpt:      excerpt,
		ContentHTML:  content,
		ImageURL:     imageURL,
		Author:       author,
		PublishedAt:  publishedAt,
		FetchedAt:    time.Now(),
		ContentHash:  hash,
		RawData:      datatypes.JSON(raw),
	}

	return feedItem, normalizeTags(item.Categories), nil
}

func (n *Normalizer) imageURL(item *gofeed.Item, html string) string {
	if item.Image != nil && strings.TrimSpace(item.Image.URL) != "" {
		return strings.TrimSpace(item.Image.URL)
	}
	for _, enclosure := range item.Enclosures {
		if enclosure == nil {
			continue
		}
		if strings.HasPrefix(strings.ToLower(enclosure.Type), "image/") && strings.TrimSpace(enclosure.URL) != "" {
			return strings.TrimSpace(enclosure.URL)
		}
	}
	return n.cleaner.ExtractImage(html)
}

func contentHash(sourceID uuid.UUID, guid string, canonicalURL string, title string) string {
	hash := sha256.Sum256([]byte(sourceID.String() + "|" + guid + "|" + canonicalURL + "|" + title))
	return hex.EncodeToString(hash[:])
}

func normalizeTags(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func nilIfEmpty(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
