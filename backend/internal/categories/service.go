package categories

import (
	"context"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]CategoryResponse, error) {
	records, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]CategoryResponse, 0, len(records))
	for _, record := range records {
		resp = append(resp, categoryResponse(record))
	}
	return resp, nil
}

func categoryResponse(category models.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
	}
}
