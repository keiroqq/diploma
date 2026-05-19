package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	httpx "github.com/keiro/content-digest/backend/internal/http"
)

type contextKey string

const userIDKey contextKey = "user_id"

func RequireAuth(jwtSecret string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.RespondError(w, http.StatusUnauthorized, "authorization header is required")
				return
			}

			tokenValue, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(tokenValue) == "" {
				httpx.RespondError(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			claims := jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenValue, &claims, func(token *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid || claims.Subject == "" {
				logger.Debug("jwt validation failed", "error", err)
				httpx.RespondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				httpx.RespondError(w, http.StatusUnauthorized, "invalid token subject")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
