package models

import (
	"time"

	"github.com/google/uuid"
)

type SavedItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_saved_user_item" json:"user_id"`
	ItemID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_saved_user_item" json:"item_id"`
	CreatedAt time.Time `json:"created_at"`
	Item      FeedItem  `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}
