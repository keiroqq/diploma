package feeds

import (
	"time"

	"github.com/google/uuid"
)

type CreateFeedRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=120"`
	Description string `json:"description" validate:"max=1000"`
	Icon        string `json:"icon" validate:"max=64"`
	ThemeColor  string `json:"theme_color" validate:"max=32"`
	LayoutType  string `json:"layout_type" validate:"max=32"`
	IsDefault   bool   `json:"is_default"`
}

type UpdateFeedRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=120"`
	Description string `json:"description" validate:"max=1000"`
	Icon        string `json:"icon" validate:"max=64"`
	ThemeColor  string `json:"theme_color" validate:"max=32"`
	LayoutType  string `json:"layout_type" validate:"max=32"`
	IsDefault   bool   `json:"is_default"`
}

type AddSourceRequest struct {
	SourceID uuid.UUID `json:"source_id" validate:"required"`
	Priority int       `json:"priority"`
}

type FeedResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	ThemeColor  string    `json:"theme_color"`
	LayoutType  string    `json:"layout_type"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FeedSourceResponse struct {
	ID        uuid.UUID `json:"id"`
	FeedID    uuid.UUID `json:"feed_id"`
	SourceID  uuid.UUID `json:"source_id"`
	IsEnabled bool      `json:"is_enabled"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}
