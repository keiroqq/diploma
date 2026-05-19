package models

import (
	"time"

	"github.com/google/uuid"
)

type FeedSource struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FeedID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_feed_source" json:"feed_id"`
	SourceID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_feed_source" json:"source_id"`
	IsEnabled bool      `gorm:"not null" json:"is_enabled"`
	Priority  int       `gorm:"not null" json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	Source    Source    `gorm:"foreignKey:SourceID" json:"source,omitempty"`
}
