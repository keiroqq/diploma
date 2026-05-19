package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Slug        string    `gorm:"not null;uniqueIndex" json:"slug"`
	Description string    `gorm:"not null" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type TagAlias struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Provider   string    `gorm:"not null;uniqueIndex:idx_tag_alias_provider_slug" json:"provider"`
	RawTag     string    `gorm:"not null" json:"raw_tag"`
	RawTagSlug string    `gorm:"not null;index;uniqueIndex:idx_tag_alias_provider_slug" json:"raw_tag_slug"`
	CreatedAt  time.Time `json:"created_at"`
	Category   Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}
