package sources

import (
	"time"

	"github.com/google/uuid"
)

type CreateSourceRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=160"`
	Type        string `json:"type" validate:"omitempty,oneof=rss"`
	URL         string `json:"url" validate:"omitempty,url"`
	FeedURL     string `json:"feed_url" validate:"required,url"`
	Description string `json:"description" validate:"max=1000"`
	Language    string `json:"language" validate:"max=16"`
	IsPublic    bool   `json:"is_public"`
	StorageMode string `json:"storage_mode" validate:"omitempty,oneof=server local"`
}

type UpdateSourceRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=160"`
	Type        string `json:"type" validate:"omitempty,oneof=rss"`
	URL         string `json:"url" validate:"omitempty,url"`
	FeedURL     string `json:"feed_url" validate:"required,url"`
	Description string `json:"description" validate:"max=1000"`
	Language    string `json:"language" validate:"max=16"`
	IsPublic    bool   `json:"is_public"`
	StorageMode string `json:"storage_mode" validate:"omitempty,oneof=server local"`
	Status      string `json:"status" validate:"omitempty,oneof=active pending disabled error"`
}

type SourceResponse struct {
	ID            uuid.UUID  `json:"id"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	URL           string     `json:"url"`
	FeedURL       string     `json:"feed_url"`
	Description   string     `json:"description"`
	Language      string     `json:"language"`
	IsPublic      bool       `json:"is_public"`
	StorageMode   string     `json:"storage_mode"`
	Status        string     `json:"status"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
