package catalog

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/fetch"
	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	repo       *Repository
	discoverer *Discoverer
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:       repo,
		discoverer: NewDiscoverer(fetch.NewSafeHTTPClient(fetch.DefaultTimeout, fetch.DefaultMaxResponseBytes)),
	}
}

func (s *Service) Topics(ctx context.Context) []Topic {
	return Topics()
}

func (s *Service) Discover(ctx context.Context, req DiscoverRequest) (*DiscoverResponse, error) {
	resp, err := s.discoverer.Discover(ctx, req.PageURL)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *Service) ConnectCatalogSources(ctx context.Context, feedID uuid.UUID, userID uuid.UUID, req ConnectCatalogSourcesRequest) (*ConnectCatalogSourcesResponse, error) {
	if len(req.SourceIDs) == 0 {
		return nil, httpx.ErrInvalidInput
	}

	exists, err := s.repo.FeedExistsForUser(ctx, feedID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httpx.ErrNotFound
	}

	response := &ConnectCatalogSourcesResponse{
		Connected: make([]ConnectedCatalogSource, 0, len(req.SourceIDs)),
	}
	for index, catalogSourceID := range req.SourceIDs {
		catalogSource, ok := FindCatalogSource(catalogSourceID)
		if !ok {
			return nil, httpx.ErrNotFound
		}

		discovered, err := s.discoverer.Discover(ctx, catalogSource.PageURL)
		if err != nil {
			return nil, err
		}

		source, err := s.findOrCreateSource(ctx, catalogSource, discovered.FeedURL)
		if err != nil {
			return nil, err
		}

		link := &models.FeedSource{
			ID:        uuid.New(),
			FeedID:    feedID,
			SourceID:  source.ID,
			IsEnabled: true,
			Priority:  len(req.SourceIDs) - index,
		}
		storedLink, err := s.repo.UpsertFeedSource(ctx, link)
		if err != nil {
			return nil, err
		}

		response.Connected = append(response.Connected, ConnectedCatalogSource{
			CatalogSourceID: catalogSource.ID,
			SourceID:        source.ID,
			FeedSourceID:    storedLink.ID,
			Title:           catalogSource.Title,
			PageURL:         catalogSource.PageURL,
			FeedURL:         source.FeedURL,
			CreatedAt:       storedLink.CreatedAt,
		})
	}

	return response, nil
}

func (s *Service) findOrCreateSource(ctx context.Context, catalogSource CatalogSource, feedURL string) (*models.Source, error) {
	source, err := s.repo.FindPublicSourceByCatalogPage(ctx, catalogSource.PageURL)
	if err == nil {
		if source.FeedURL != feedURL {
			if err := s.repo.UpdateSourceFeedURL(ctx, source.ID, feedURL); err != nil {
				return nil, err
			}
			source.FeedURL = feedURL
		}
		return source, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	source = &models.Source{
		ID:          uuid.New(),
		CreatedBy:   nil,
		Name:        "Habr: " + catalogSource.Title,
		Type:        models.SourceTypeRSS,
		URL:         catalogSource.PageURL,
		FeedURL:     feedURL,
		Description: catalogSource.Description,
		Language:    "ru",
		IsPublic:    true,
		StorageMode: models.SourceStorageServer,
		Status:      models.SourceStatusActive,
	}
	if err := s.repo.CreateSource(ctx, source); err != nil {
		existing, findErr := s.repo.FindPublicSourceByCatalogPage(ctx, catalogSource.PageURL)
		if findErr == nil {
			return existing, nil
		}
		return nil, err
	}
	return source, nil
}
