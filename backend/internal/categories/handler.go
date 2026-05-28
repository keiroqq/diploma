package categories

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/categories", h.List)
	r.Get("/feeds/{id}/categories", h.ListByFeed)
}

// List godoc
// @Summary Получить нормализованные категории
// @Description Возвращает список категорий приложения, которые используются для нормализации тегов и фильтрации материалов.
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} CategoryResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.List(r.Context())
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// ListByFeed godoc
// @Summary Получить категории ленты
// @Description Возвращает только категории, которые соответствуют подключенным источникам выбранной ленты.
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Success 200 {array} CategoryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/categories [get]
func (h *Handler) ListByFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	resp, err := h.service.ListByFeed(r.Context(), feedID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
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
