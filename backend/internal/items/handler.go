package items

import (
	"errors"
	"net/http"
	"strconv"

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
	r.Get("/feeds/{id}/items", h.ListFeedItems)
	r.Get("/items/search", h.SearchItems)
	r.Post("/items/{id}/save", h.SaveItem)
	r.Delete("/items/{id}/save", h.UnsaveItem)
	r.Get("/saved", h.ListSaved)
}

// ListFeedItems godoc
// @Summary Получить материалы ленты
// @Description mode=today возвращает сегодняшние материалы, mode=archive возвращает архивные материалы с cursor-пагинацией, mode=all возвращает материалы без ограничения по дате.
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param mode query string false "today, archive или all"
// @Param cursor query string false "RFC3339 cursor для archive"
// @Param limit query int false "Лимит, максимум 100"
// @Param category query string false "Slug категории, например ai или backend"
// @Param categories query string false "Slugs категорий через запятую, например ai,backend"
// @Param date_from query string false "Дата начала в формате YYYY-MM-DD или RFC3339"
// @Param date_to query string false "Дата конца в формате YYYY-MM-DD или RFC3339"
// @Success 200 {object} FeedItemsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/items [get]
func (h *Handler) ListFeedItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	query, err := ParseListQuery(r.URL.Query())
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid query parameters")
		return
	}

	resp, err := h.service.ListFeedItems(r.Context(), feedID, userID, query)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// SearchItems godoc
// @Summary Поиск материалов
// @Description Ищет по материалам из доступных пользователю потоков. Если передан feed_id, область поиска сужается до этого потока.
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param q query string true "Поисковый запрос"
// @Param feed_id query string false "Feed ID для поиска внутри одного потока"
// @Param limit query int false "Лимит, максимум 200"
// @Success 200 {object} SearchItemsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/search [get]
func (h *Handler) SearchItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}

	query, err := ParseSearchQuery(r.URL.Query())
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid query parameters")
		return
	}

	resp, err := h.service.SearchItems(r.Context(), userID, query)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// SaveItem godoc
// @Summary Сохранить материал в избранное
// @Tags saved
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/{id}/save [post]
func (h *Handler) SaveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	itemID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.SaveItem(r.Context(), userID, itemID); err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}

// UnsaveItem godoc
// @Summary Удалить материал из избранного
// @Tags saved
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/items/{id}/save [delete]
func (h *Handler) UnsaveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	itemID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.UnsaveItem(r.Context(), userID, itemID); err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}

// ListSaved godoc
// @Summary Получить избранные материалы
// @Tags saved
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Лимит, максимум 200"
// @Success 200 {object} SavedItemsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/saved [get]
func (h *Handler) ListSaved(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpx.RespondError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsed
	}

	resp, err := h.service.ListSaved(r.Context(), userID, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, httpx.ErrInvalidInput):
		httpx.RespondError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, httpx.ErrNotFound):
		httpx.RespondError(w, http.StatusNotFound, "item or feed not found")
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
