package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type Service struct {
	repo      *Repository
	validate  *validator.Validate
	jwtSecret string
	jwtTTL    time.Duration
	logger    *slog.Logger
}

func NewService(repo *Repository, jwtSecret string, jwtTTL time.Duration, logger *slog.Logger) *Service {
	return &Service{
		repo:      repo,
		validate:  validator.New(),
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
		logger:    logger,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}

	s.logger.Info("user registered", "user_id", user.ID.String(), "email", user.Email)
	return s.authResponse(user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	s.logger.Info("user logged in", "user_id", user.ID.String(), "email", user.Email)
	return s.authResponse(user)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := userResponse(user)
	return &resp, nil
}

func (s *Service) authResponse(user *models.User) (*AuthResponse, error) {
	token, err := GenerateToken(user.ID, s.jwtSecret, s.jwtTTL)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: userResponse(user)}, nil
}

func userResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
