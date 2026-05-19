package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	RuleInclude  = "include"
	RuleExclude  = "exclude"
	RuleBoost    = "boost"
	RuleDownrank = "downrank"

	TargetKeyword = "keyword"
	TargetTag     = "tag"
	TargetSource  = "source"
	TargetAuthor  = "author"
)

type FilterRule struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FeedID     uuid.UUID `gorm:"type:uuid;not null;index" json:"feed_id"`
	RuleType   string    `gorm:"not null" json:"rule_type"`
	TargetType string    `gorm:"not null" json:"target_type"`
	Value      string    `gorm:"not null" json:"value"`
	Weight     int       `gorm:"not null" json:"weight"`
	IsActive   bool      `gorm:"not null" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
