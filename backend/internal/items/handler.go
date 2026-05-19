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
	r.Post("/items/{id}/save", h.SaveItem)
	r.Delete("/items/{id}/save", h.UnsaveItem)
	r.Get("/saved", h.ListSaved)
}

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
