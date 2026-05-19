package catalog

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
	r.Get("/catalog/topics", h.Topics)
	r.Post("/catalog/discover", h.Discover)
	r.Post("/feeds/{id}/catalog-sources", h.ConnectCatalogSources)
}

// Topics godoc
// @Summary Получить каталог тем
// @Description Возвращает curated-каталог тем и страниц источников, которые можно подключить к ленте.
// @Tags catalog
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Topic
// @Failure 401 {object} map[string]string
// @Router /api/catalog/topics [get]
func (h *Handler) Topics(w http.ResponseWriter, r *http.Request) {
	httpx.RespondJSON(w, http.StatusOK, h.service.Topics(r.Context()))
}

// Discover godoc
// @Summary Найти RSS на странице
// @Description Скачивает HTML-страницу и ищет link rel=alternate type=application/rss+xml.
// @Tags catalog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body DiscoverRequest true "URL страницы"
// @Success 200 {object} DiscoverResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/catalog/discover [post]
func (h *Handler) Discover(w http.ResponseWriter, r *http.Request) {
	var req DiscoverRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Discover(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// ConnectCatalogSources godoc
// @Summary Подключить источники каталога к ленте
// @Description Находит RSS для выбранных catalog source ids, создает sources и feed_sources.
// @Tags catalog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param payload body ConnectCatalogSourcesRequest true "Catalog source IDs"
// @Success 201 {object} ConnectCatalogSourcesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/catalog-sources [post]
func (h *Handler) ConnectCatalogSources(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	var req ConnectCatalogSourcesRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.ConnectCatalogSources(r.Context(), feedID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, httpx.ErrInvalidInput):
		httpx.RespondError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, httpx.ErrNotFound):
		httpx.RespondError(w, http.StatusNotFound, "catalog source or feed not found")
	case errors.Is(err, ErrRSSNotFound):
		httpx.RespondError(w, http.StatusUnprocessableEntity, "rss feed not found on page")
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
