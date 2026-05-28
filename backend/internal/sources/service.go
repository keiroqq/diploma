package sources

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/fetch"
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

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]SourceResponse, error) {
	records, err := s.repo.ListAccessible(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]SourceResponse, 0, len(records))
	for _, record := range records {
		resp = append(resp, sourceResponse(record))
	}
	return resp, nil
}

func (s *Service) Get(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (*SourceResponse, error) {
	source, err := s.repo.GetAccessible(ctx, sourceID, userID)
	if err != nil {
		return nil, mapGormNotFound(err)
	}
	resp := sourceResponse(*source)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateSourceRequest) (*SourceResponse, error) {
	normalizeCreate(&req)
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	if err := validateSourceURLs(req.URL, req.FeedURL); err != nil {
		return nil, err
	}

	source := &models.Source{
		ID:          uuid.New(),
		CreatedBy:   &userID,
		Name:        req.Name,
		Type:        req.Type,
		URL:         req.URL,
		FeedURL:     req.FeedURL,
		Description: req.Description,
		Language:    req.Language,
		IsPublic:    req.IsPublic,
		StorageMode: req.StorageMode,
		Status:      models.SourceStatusActive,
	}
	if err := s.repo.Create(ctx, source); err != nil {
		return nil, err
	}

	resp := sourceResponse(*source)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID, req UpdateSourceRequest) (*SourceResponse, error) {
	source, err := s.repo.GetAccessible(ctx, sourceID, userID)
	if err != nil {
		return nil, mapGormNotFound(err)
	}
	if source.CreatedBy == nil || *source.CreatedBy != userID {
		return nil, httpx.ErrForbidden
	}

	normalizeUpdate(&req)
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	if err := validateSourceURLs(req.URL, req.FeedURL); err != nil {
		return nil, err
	}

	source.Name = req.Name
	source.Type = req.Type
	source.URL = req.URL
	source.FeedURL = req.FeedURL
	source.Description = req.Description
	source.Language = req.Language
	source.IsPublic = req.IsPublic
	source.StorageMode = req.StorageMode
	source.Status = req.Status

	if err := s.repo.Update(ctx, source); err != nil {
		return nil, err
	}
	resp := sourceResponse(*source)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) error {
	source, err := s.repo.GetAccessible(ctx, sourceID, userID)
	if err != nil {
		return mapGormNotFound(err)
	}
	if source.CreatedBy == nil || *source.CreatedBy != userID {
		return httpx.ErrForbidden
	}
	return s.repo.Delete(ctx, sourceID, userID)
}

func normalizeCreate(req *CreateSourceRequest) {
	req.Name = strings.TrimSpace(req.Name)
	req.Type = defaultString(strings.TrimSpace(req.Type), models.SourceTypeRSS)
	req.URL = strings.TrimSpace(req.URL)
	req.FeedURL = strings.TrimSpace(req.FeedURL)
	req.Description = strings.TrimSpace(req.Description)
	req.Language = defaultString(strings.TrimSpace(req.Language), "ru")
	req.StorageMode = defaultString(strings.TrimSpace(req.StorageMode), models.SourceStorageServer)
}

func normalizeUpdate(req *UpdateSourceRequest) {
	req.Name = strings.TrimSpace(req.Name)
	req.Type = defaultString(strings.TrimSpace(req.Type), models.SourceTypeRSS)
	req.URL = strings.TrimSpace(req.URL)
	req.FeedURL = strings.TrimSpace(req.FeedURL)
	req.Description = strings.TrimSpace(req.Description)
	req.Language = defaultString(strings.TrimSpace(req.Language), "ru")
	req.StorageMode = defaultString(strings.TrimSpace(req.StorageMode), models.SourceStorageServer)
	req.Status = defaultString(strings.TrimSpace(req.Status), models.SourceStatusActive)
}

func sourceResponse(source models.Source) SourceResponse {
	return SourceResponse{
		ID:            source.ID,
		CreatedBy:     source.CreatedBy,
		Name:          source.Name,
		Type:          source.Type,
		URL:           source.URL,
		FeedURL:       source.FeedURL,
		Description:   source.Description,
		Language:      source.Language,
		IsPublic:      source.IsPublic,
		StorageMode:   source.StorageMode,
		Status:        source.Status,
		LastFetchedAt: source.LastFetchedAt,
		CreatedAt:     source.CreatedAt,
		UpdatedAt:     source.UpdatedAt,
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func validateSourceURLs(pageURL string, feedURL string) error {
	if strings.TrimSpace(pageURL) != "" {
		if err := fetch.ValidateURL(pageURL); err != nil {
			return httpx.ErrInvalidInput
		}
	}
	if err := fetch.ValidateURL(feedURL); err != nil {
		return httpx.ErrInvalidInput
	}
	return nil
}

func mapGormNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return httpx.ErrNotFound
	}
	return err
}
