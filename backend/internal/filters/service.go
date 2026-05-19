package filters

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	repo     *Repository
	validate *validator.Validate
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo, validate: validator.New()}
}

func (s *Service) List(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) ([]RuleResponse, error) {
	exists, err := s.repo.FeedExistsForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httpx.ErrNotFound
	}

	rules, err := s.repo.ListByFeedForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]RuleResponse, 0, len(rules))
	for _, rule := range rules {
		resp = append(resp, ruleResponse(rule))
	}
	return resp, nil
}

func (s *Service) Create(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, req CreateRuleRequest) (*RuleResponse, error) {
	normalizeCreate(&req)
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	exists, err := s.repo.FeedExistsForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httpx.ErrNotFound
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := &models.FilterRule{
		ID:         uuid.New(),
		FeedID:     feedID,
		RuleType:   req.RuleType,
		TargetType: req.TargetType,
		Value:      req.Value,
		Weight:     req.Weight,
		IsActive:   isActive,
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}
	resp := ruleResponse(*rule)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, ruleID uuid.UUID, userID uuid.UUID, req UpdateRuleRequest) (*RuleResponse, error) {
	req.RuleType = strings.TrimSpace(req.RuleType)
	req.TargetType = strings.TrimSpace(req.TargetType)
	req.Value = strings.TrimSpace(req.Value)
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	rule, err := s.repo.GetByIDForUser(ctx, ruleID, userID)
	if err != nil {
		return nil, mapGormNotFound(err)
	}

	rule.RuleType = req.RuleType
	rule.TargetType = req.TargetType
	rule.Value = req.Value
	rule.Weight = req.Weight
	rule.IsActive = req.IsActive

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, err
	}
	resp := ruleResponse(*rule)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, ruleID uuid.UUID, userID uuid.UUID) error {
	if _, err := s.repo.GetByIDForUser(ctx, ruleID, userID); err != nil {
		return mapGormNotFound(err)
	}
	return s.repo.Delete(ctx, ruleID, userID)
}

func normalizeCreate(req *CreateRuleRequest) {
	req.RuleType = strings.TrimSpace(req.RuleType)
	req.TargetType = strings.TrimSpace(req.TargetType)
	req.Value = strings.TrimSpace(req.Value)
}

func ruleResponse(rule models.FilterRule) RuleResponse {
	return RuleResponse{
		ID:         rule.ID,
		FeedID:     rule.FeedID,
		RuleType:   rule.RuleType,
		TargetType: rule.TargetType,
		Value:      rule.Value,
		Weight:     rule.Weight,
		IsActive:   rule.IsActive,
		CreatedAt:  rule.CreatedAt,
		UpdatedAt:  rule.UpdatedAt,
	}
}

func mapGormNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return httpx.ErrNotFound
	}
	return err
}
