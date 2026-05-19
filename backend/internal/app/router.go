package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/keiro/content-digest/backend/internal/auth"
	"github.com/keiro/content-digest/backend/internal/catalog"
	"github.com/keiro/content-digest/backend/internal/feeds"
	"github.com/keiro/content-digest/backend/internal/filters"
	httpx "github.com/keiro/content-digest/backend/internal/http"
	"github.com/keiro/content-digest/backend/internal/items"
	"github.com/keiro/content-digest/backend/internal/middleware"
	"github.com/keiro/content-digest/backend/internal/sources"
)

type routerHandlers struct {
	Auth    *auth.Handler
	Catalog *catalog.Handler
	Feeds   *feeds.Handler
	Sources *sources.Handler
	Items   *items.Handler
	Filters *filters.Handler
}

func newRouter(logger *slog.Logger, handlers routerHandlers, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger(logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	authMiddleware := middleware.RequireAuth(jwtSecret, logger)
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.With(authMiddleware).Get("/me", handlers.Auth.Me)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			handlers.Catalog.RegisterRoutes(r)
			handlers.Feeds.RegisterRoutes(r)
			handlers.Sources.RegisterRoutes(r)
			handlers.Items.RegisterRoutes(r)
			handlers.Filters.RegisterRoutes(r)
		})
	})

	return r
}
