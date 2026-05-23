package filters

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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
	r.Get("/feeds/{id}/rules", h.List)
	r.Post("/feeds/{id}/rules", h.Create)
	r.Put("/feeds/{id}/rules/{ruleId}", h.Update)
	r.Delete("/feeds/{id}/rules/{ruleId}", h.Delete)
}

// List godoc
// @Summary Получить правила фильтрации ленты
// @Tags filter-rules
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Success 200 {array} RuleResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/rules [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}
	resp, err := h.service.List(r.Context(), feedID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Create godoc
// @Summary Создать правило фильтрации
// @Tags filter-rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param payload body CreateRuleRequest true "Данные правила"
// @Success 201 {object} RuleResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/rules [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	feedID, ok := uuidParam(w, r, "id")
	if !ok {
		return
	}

	var req CreateRuleRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Create(r.Context(), feedID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, resp)
}

// Update godoc
// @Summary Обновить правило фильтрации
// @Tags filter-rules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param ruleId path string true "Rule ID"
// @Param payload body UpdateRuleRequest true "Данные правила"
// @Success 200 {object} RuleResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/rules/{ruleId} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	ruleID, ok := uuidParam(w, r, "ruleId")
	if !ok {
		return
	}

	var req UpdateRuleRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.service.Update(r.Context(), ruleID, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Удалить правило фильтрации
// @Tags filter-rules
// @Security BearerAuth
// @Param id path string true "Feed ID"
// @Param ruleId path string true "Rule ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/feeds/{id}/rules/{ruleId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUser(w, r)
	if !ok {
		return
	}
	ruleID, ok := uuidParam(w, r, "ruleId")
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), ruleID, userID); err != nil {
		h.handleError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	switch {
	case errors.As(err, &validationErrors):
		httpx.RespondJSON(w, http.StatusBadRequest, httpx.ErrorResponse{Error: "validation failed", Details: validationErrors.Error()})
	case errors.Is(err, httpx.ErrNotFound):
		httpx.RespondError(w, http.StatusNotFound, "filter rule not found")
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
