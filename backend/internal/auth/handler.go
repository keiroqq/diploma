package auth

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register godoc
// @Summary Зарегистрировать пользователя
// @Description Создает пользователя и возвращает JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body RegisterRequest true "Данные регистрации"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, resp)
}

// Login godoc
// @Summary Войти в приложение
// @Description Проверяет email и пароль, возвращает JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body LoginRequest true "Данные входа"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Me godoc
// @Summary Получить текущего пользователя
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.Me(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	switch {
	case errors.As(err, &validationErrors):
		httpx.RespondJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "validation failed", Details: validationErrors.Error()})
	case errors.Is(err, ErrEmailAlreadyExists):
		httpx.RespondError(w, http.StatusConflict, "email already exists")
	case errors.Is(err, ErrInvalidCredentials):
		httpx.RespondError(w, http.StatusUnauthorized, "invalid credentials")
	default:
		httpx.RespondError(w, http.StatusInternalServerError, "internal server error")
	}
}
