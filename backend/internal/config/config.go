package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	HTTPPort           string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiresIn       time.Duration
	LogLevel           string
	RSSRefreshCooldown time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtTTL, err := durationEnv("JWT_EXPIRES_IN", 24*time.Hour)
	if err != nil {
		return nil, err
	}

	refreshCooldown, err := durationEnv("RSS_REFRESH_COOLDOWN", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return &Config{
		AppEnv:             stringEnv("APP_ENV", "development"),
		HTTPPort:           stringEnv("HTTP_PORT", "8080"),
		DatabaseURL:        stringEnv("DATABASE_URL", "postgres://app:app@localhost:5432/content_digest?sslmode=disable"),
		JWTSecret:          stringEnv("JWT_SECRET", "change_me"),
		JWTExpiresIn:       jwtTTL,
		LogLevel:           stringEnv("LOG_LEVEL", "info"),
		RSSRefreshCooldown: refreshCooldown,
	}, nil
}

func stringEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	return time.ParseDuration(value)
}
