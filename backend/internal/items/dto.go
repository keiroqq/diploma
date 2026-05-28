package items

import (
	"time"

	"github.com/google/uuid"
)

type ListMode string

const (
	ModeToday   ListMode = "today"
	ModeArchive ListMode = "archive"
	ModeAll     ListMode = "all"
)

type ListQuery struct {
	Mode       ListMode
	Cursor     *time.Time
	Limit      int
	Category   string
	Categories []string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type SearchQuery struct {
	Query  string
	FeedID *uuid.UUID
	Limit  int
}

type ItemResponse struct {
	ID          uuid.UUID `json:"id"`
	SourceID    uuid.UUID `json:"source_id"`
	SourceName  string    `json:"source_name"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Excerpt     string    `json:"excerpt"`
	ImageURL    string    `json:"image_url"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	Tags        []string  `json:"tags"`
	Categories  []string  `json:"categories"`
	Score       int       `json:"score"`
	IsSaved     bool      `json:"is_saved"`
}

type FeedItemsResponse struct {
	Items      []ItemResponse `json:"items"`
	Mode       ListMode       `json:"mode"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

type SavedItemsResponse struct {
	Items []ItemResponse `json:"items"`
}

type SearchItemsResponse struct {
	Items []ItemResponse `json:"items"`
	Query string         `json:"query"`
}
