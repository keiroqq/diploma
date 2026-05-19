package models

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex" json:"name"`
	Slug      string    `gorm:"not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}
