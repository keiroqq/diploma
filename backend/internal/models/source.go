package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	SourceTypeRSS = "rss"

	SourceStatusActive   = "active"
	SourceStatusPending  = "pending"
	SourceStatusDisabled = "disabled"
	SourceStatusError    = "error"

	SourceStorageServer = "server"
	SourceStorageLocal  = "local"
)

type Source struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedBy     *uuid.UUID `gorm:"type:uuid;index" json:"created_by"`
	Name          string     `gorm:"not null" json:"name"`
	Type          string     `gorm:"not null" json:"type"`
	URL           string     `gorm:"not null" json:"url"`
	FeedURL       string     `gorm:"not null" json:"feed_url"`
	Description   string     `gorm:"not null" json:"description"`
	Language      string     `gorm:"not null" json:"language"`
	IsPublic      bool       `gorm:"not null;index" json:"is_public"`
	StorageMode   string     `gorm:"not null;default:server;index" json:"storage_mode"`
	Status        string     `gorm:"not null" json:"status"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
