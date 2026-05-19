package filters

import (
	"time"

	"github.com/google/uuid"
)

type CreateRuleRequest struct {
	RuleType   string `json:"rule_type" validate:"required,oneof=include exclude boost downrank"`
	TargetType string `json:"target_type" validate:"required,oneof=keyword tag source author"`
	Value      string `json:"value" validate:"required,max=255"`
	Weight     int    `json:"weight"`
	IsActive   *bool  `json:"is_active"`
}

type UpdateRuleRequest struct {
	RuleType   string `json:"rule_type" validate:"required,oneof=include exclude boost downrank"`
	TargetType string `json:"target_type" validate:"required,oneof=keyword tag source author"`
	Value      string `json:"value" validate:"required,max=255"`
	Weight     int    `json:"weight"`
	IsActive   bool   `json:"is_active"`
}

type RuleResponse struct {
	ID         uuid.UUID `json:"id"`
	FeedID     uuid.UUID `json:"feed_id"`
	RuleType   string    `json:"rule_type"`
	TargetType string    `json:"target_type"`
	Value      string    `json:"value"`
	Weight     int       `json:"weight"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
