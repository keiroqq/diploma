package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/keiro/content-digest/backend/internal/auth"
	"github.com/keiro/content-digest/backend/internal/config"
	"github.com/keiro/content-digest/backend/internal/db"
	"github.com/keiro/content-digest/backend/internal/feeds"
	"github.com/keiro/content-digest/backend/internal/filters"
	"github.com/keiro/content-digest/backend/internal/items"
	"github.com/keiro/content-digest/backend/internal/logger"
	"github.com/keiro/content-digest/backend/internal/rss"
	"github.com/keiro/content-digest/backend/internal/sources"
	"github.com/keiro/content-digest/backend/internal/users"
)

type App struct {
	server *http.Server
	db     *gorm.DB
	sqlDB  *sql.DB
	logger *slog.Logger
}

func New(cfg *config.Config) (*App, error) {
	log := logger.New(cfg.LogLevel, cfg.AppEnv)

	gormDB, sqlDB, err := db.Open(cfg.DatabaseURL, log)
	if err != nil {
		return nil, err
	}

	authRepo := auth.NewRepository(gormDB)
	authService := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTExpiresIn, log)
	authHandler := auth.NewHandler(authService)

	usersRepo := users.NewRepository(gormDB)
	_ = users.NewService(usersRepo)

	feedRepo := feeds.NewRepository(gormDB)
	feedService := feeds.NewService(feedRepo)

	sourceRepo := sources.NewRepository(gormDB)
	sourceService := sources.NewService(sourceRepo)

	filterRepo := filters.NewRepository(gormDB)
	filterService := filters.NewService(filterRepo)

	itemRepo := items.NewRepository(gormDB)
	itemService := items.NewService(itemRepo)

	rssService := rss.NewService(gormDB, cfg.RSSRefreshCooldown, log)

	handlers := routerHandlers{
		Auth:    authHandler,
		Feeds:   feeds.NewHandler(feedService, rssService),
		Sources: sources.NewHandler(sourceService, rssService),
		Items:   items.NewHandler(itemService),
		Filters: filters.NewHandler(filterService),
	}

	router := newRouter(log, handlers, cfg.JWTSecret)
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server: server,
		db:     gormDB,
		sqlDB:  sqlDB,
		logger: log,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("server started", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("server shutdown started")
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	if a.sqlDB != nil {
		if err := a.sqlDB.Close(); err != nil {
			return err
		}
	}
	a.logger.Info("server shutdown finished")
	return nil
}
