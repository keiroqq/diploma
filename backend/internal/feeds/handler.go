package feeds

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

type FeedRefresher interface {
	RefreshFeed(ctx context.Context, feedID uuid.UUID, userID uuid.UUID) (any, error)
}

type Handler struct {
	service   *Service
	refresher FeedRefresher
}

func NewHandler(service *Service, refresher FeedRefresher) *Handler {
	return &Handler{service: service, refresher: refresher}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/feeds", h.List)
	r.Post("/feeds", h.Create)
	r.Get("/feeds/{id}", h.Get)
	r.Put("/feeds/{id}", h.Update)
	r.Delete("/feeds/{id}", h.Delete)
	r.Post("/feeds/{id}/sources", h.AddSource)
	r.Delete("/feeds/{id}/sources/{sourceId}", h.RemoveSource)
	r.Post("/feeds/{id}/refresh", h.Refresh)
}

// List godoc
// @Summary Получить ленты пользователя
// @Tags feeds
// @Produce json
// @Security BearerAuth
// @Success 200 {array} FeedResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds [get]
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
// @Summary Получить ленту
// @Tags feeds
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Success 200 {object} FeedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	resp, err := h.service.Get(r.Context(), feedID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Create godoc
// @Summary Создать ленту
// @Tags feeds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreateFeedRequest true "Данные ленты"
// @Success 201 {object} FeedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req CreateFeedRequest
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
// @Summary Обновить ленту
// @Tags feeds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param payload body UpdateFeedRequest true "Данные ленты"
// @Success 200 {object} FeedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	var req UpdateFeedRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Update(r.Context(), feedID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Удалить ленту
// @Tags feeds
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), feedID, userID); err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}

// AddSource godoc
// @Summary Подключить source к ленте
// @Tags feeds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param payload body AddSourceRequest true "Source ID и priority"
// @Success 201 {object} FeedSourceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/sources [post]
func (h *Handler) AddSource(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	var req AddSourceRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.AddSource(r.Context(), feedID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, resp)
}

// RemoveSource godoc
// @Summary Отключить source от ленты
// @Tags feeds
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param sourceId path string true "Source ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/sources/{sourceId} [delete]
func (h *Handler) RemoveSource(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	sourceID, ok := uuidParam(w, r, "sourceId")
	if !ok {
		return
	}

	if err := h.service.RemoveSource(r.Context(), feedID, userID, sourceID); err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}

// Refresh godoc
// @Summary Обновить все источники ленты
// @Tags refresh
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Success 200 {object} rss.RefreshResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	if h.refresher == nil {
		httpx.RespondError(w, http.StatusNotImplemented, "feed refresh is not configured")
		return
	}

	resp, err := h.refresher.RefreshFeed(r.Context(), feedID, userID)
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
		httpx.RespondError(w, http.StatusNotFound, "feed not found")
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
