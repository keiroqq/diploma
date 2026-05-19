package feeds

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

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]FeedResponse, error) {
	records, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]FeedResponse, 0, len(records))
	for _, record := range records {
		resp = append(resp, feedResponse(record))
	}
	return resp, nil
}

func (s *Service) Get(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (*FeedResponse, error) {
	feed, err := s.repo.GetByIDForUser(ctx, feedID, userID)
	if err != nil {
		return nil, mapGormNotFound(err)
	}
	resp := feedResponse(*feed)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateFeedRequest) (*FeedResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Icon = defaultString(strings.TrimSpace(req.Icon), "newspaper")
	req.ThemeColor = defaultString(strings.TrimSpace(req.ThemeColor), "#2563eb")
	req.LayoutType = defaultString(strings.TrimSpace(req.LayoutType), "cards")

	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	feed := &models.Feed{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		ThemeColor:  req.ThemeColor,
		LayoutType:  req.LayoutType,
		IsDefault:   req.IsDefault,
	}
	if err := s.repo.Create(ctx, feed); err != nil {
		return nil, err
	}
	resp := feedResponse(*feed)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, req UpdateFeedRequest) (*FeedResponse, error) {
	feed, err := s.repo.GetByIDForUser(ctx, feedID, userID)
	if err != nil {
		return nil, mapGormNotFound(err)
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Icon = defaultString(strings.TrimSpace(req.Icon), "newspaper")
	req.ThemeColor = defaultString(strings.TrimSpace(req.ThemeColor), "#2563eb")
	req.LayoutType = defaultString(strings.TrimSpace(req.LayoutType), "cards")

	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	feed.Name = req.Name
	feed.Description = req.Description
	feed.Icon = req.Icon
	feed.ThemeColor = req.ThemeColor
	feed.LayoutType = req.LayoutType
	feed.IsDefault = req.IsDefault

	if err := s.repo.Update(ctx, feed); err != nil {
		return nil, err
	}
	resp := feedResponse(*feed)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) error {
	if _, err := s.repo.GetByIDForUser(ctx, feedID, userID); err != nil {
		return mapGormNotFound(err)
	}
	return s.repo.Delete(ctx, feedID, userID)
}

func (s *Service) AddSource(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, req AddSourceRequest) (*FeedSourceResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetByIDForUser(ctx, feedID, userID); err != nil {
		return nil, mapGormNotFound(err)
	}
	accessible, err := s.repo.SourceAccessible(ctx, req.SourceID, userID)
	if err != nil {
		return nil, err
	}
	if !accessible {
		return nil, httpx.ErrNotFound
	}

	link := &models.FeedSource{
		ID:        uuid.New(),
		FeedID:    feedID,
		SourceID:  req.SourceID,
		IsEnabled: true,
		Priority:  req.Priority,
	}
	if err := s.repo.AddSource(ctx, link); err != nil {
		return nil, err
	}

	resp := FeedSourceResponse{
		ID:        link.ID,
		FeedID:    link.FeedID,
		SourceID:  link.SourceID,
		IsEnabled: link.IsEnabled,
		Priority:  link.Priority,
		CreatedAt: link.CreatedAt,
	}
	return &resp, nil
}

func (s *Service) RemoveSource(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, sourceID uuid.UUID) error {
	if _, err := s.repo.GetByIDForUser(ctx, feedID, userID); err != nil {
		return mapGormNotFound(err)
	}
	return s.repo.RemoveSource(ctx, feedID, sourceID)
}

func feedResponse(feed models.Feed) FeedResponse {
	return FeedResponse{
		ID:          feed.ID,
		Name:        feed.Name,
		Description: feed.Description,
		Icon:        feed.Icon,
		ThemeColor:  feed.ThemeColor,
		LayoutType:  feed.LayoutType,
		IsDefault:   feed.IsDefault,
		CreatedAt:   feed.CreatedAt,
		UpdatedAt:   feed.UpdatedAt,
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func mapGormNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return httpx.ErrNotFound
	}
	return err
}
