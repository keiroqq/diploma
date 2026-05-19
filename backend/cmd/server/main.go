package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/keiro/content-digest/backend/internal/app"
	"github.com/keiro/content-digest/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown app: %v", err)
		}
	case err := <-errCh:
		if err != nil {
			log.Fatalf("run app: %v", err)
		}
	}
}
