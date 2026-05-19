package models

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"not null" json:"description"`
	Icon        string    `gorm:"not null" json:"icon"`
	ThemeColor  string    `gorm:"not null" json:"theme_color"`
	LayoutType  string    `gorm:"not null" json:"layout_type"`
	IsDefault   bool      `gorm:"not null" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
