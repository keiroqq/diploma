package users

import (
	"context"

	"github.com/google/uuid"

	"github.com/keiro/content-digest/backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(ctx, userID)
}
