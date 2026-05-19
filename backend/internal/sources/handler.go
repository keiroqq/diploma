package sources

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/middleware"
)

type SourceRefresher interface {
	RefreshSource(ctx context.Context, sourceID uuid.UUID, userID uuid.UUID) (any, error)
}

type Handler struct {
	service   *Service
	refresher SourceRefresher
}

func NewHandler(service *Service, refresher SourceRefresher) *Handler {
	return &Handler{service: service, refresher: refresher}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/sources", h.List)
	r.Post("/sources", h.Create)
	r.Get("/sources/{id}", h.Get)
	r.Put("/sources/{id}", h.Update)
	r.Delete("/sources/{id}", h.Delete)
	r.Post("/sources/{id}/refresh", h.Refresh)
}

// List godoc
// @Summary Получить доступные источники
// @Tags sources
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SourceResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sources [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	resp, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Get godoc
// @Summary Получить источник
// @Tags sources
// @Produce json
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Success 200 {object} SourceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sources/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	sourceID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	resp, err := h.service.Get(r.Context(), sourceID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Create godoc
// @Summary Создать RSS-источник
// @Tags sources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreateSourceRequest true "Данные источника"
// @Success 201 {object} SourceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sources [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req CreateSourceRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, resp)
}

// Update godoc
// @Summary Обновить источник
// @Tags sources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Param payload body UpdateSourceRequest true "Данные источника"
// @Success 200 {object} SourceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sources/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	sourceID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	var req UpdateSourceRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Update(r.Context(), sourceID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Удалить источник
// @Tags sources
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sources/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	sourceID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), sourceID, userID); err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}

// Refresh godoc
// @Summary Обновить источник
// @Tags refresh
// @Produce json
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Success 200 {object} rss.RefreshResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sources/{id}/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	sourceID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	if h.refresher == nil {
		httpx.RespondError(w, http.StatusNotImplemented, "source refresh is not configured")
		return
	}

	resp, err := h.refresher.RefreshSource(r.Context(), sourceID, userID)
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
	case errors.Is(err, httpx.ErrNotFound):
		httpx.RespondError(w, http.StatusNotFound, "source not found")
	case errors.Is(err, httpx.ErrForbidden):
		httpx.RespondError(w, http.StatusForbidden, "source is not editable by this user")
	default:
		httpx.RespondError(w, http.StatusInternalServerError, "internal server error")
	}
}

func requireUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, false
	}
	return userID, true
}

func uuidParam(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value := chi.URLParam(r, key)
	parsed, err := uuid.Parse(value)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid "+key)
		return uuid.Nil, false
	}
	return parsed, true
}
