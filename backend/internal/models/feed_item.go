package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type FeedItem struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SourceID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"source_id"`
	ExternalID   string         `gorm:"not null" json:"external_id"`
	GUID         *string        `json:"guid"`
	Title        string         `gorm:"not null" json:"title"`
	URL          string         `gorm:"not null" json:"url"`
	CanonicalURL *string        `json:"canonical_url"`
	Excerpt      string         `gorm:"not null" json:"excerpt"`
	ContentHTML  string         `gorm:"not null" json:"content_html"`
	ImageURL     string         `gorm:"not null" json:"image_url"`
	Author       string         `gorm:"not null" json:"author"`
	PublishedAt  time.Time      `gorm:"not null;index" json:"published_at"`
	FetchedAt    time.Time      `gorm:"not null" json:"fetched_at"`
	ContentHash  string         `gorm:"not null" json:"content_hash"`
	RawData      datatypes.JSON `gorm:"type:jsonb;not null" json:"raw_data"`
	CreatedAt    time.Time      `json:"created_at"`
	Source       Source         `gorm:"foreignKey:SourceID" json:"source,omitempty"`
	Tags         []Tag          `gorm:"many2many:feed_item_tags;joinForeignKey:ItemID;joinReferences:TagID" json:"tags,omitempty"`
	Categories   []Category     `gorm:"many2many:feed_item_categories;joinForeignKey:ItemID;joinReferences:CategoryID" json:"categories,omitempty"`
}
