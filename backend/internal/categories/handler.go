package categories

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	httpx "github.com/keiro/content-digest/backend/internal/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/categories", h.List)
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
