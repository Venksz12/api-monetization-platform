package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/yourusername/api-monetization-platform/internal/repository"
)

type ContextKey string

const (
	APIKeyContextKey ContextKey = "api_key"
	OwnerContextKey  ContextKey = "owner_id"
)

type APIKeyMiddleware struct {
	Repo *repository.APIKeyRepository
}

func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if raw == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}
		rec, err := m.Repo.FindByHash(r.Context(), HashAPIKey(raw))
		if err != nil || rec.Status != "ACTIVE" {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), APIKeyContextKey, rec.ID)
		ctx = context.WithValue(ctx, OwnerContextKey, rec.OwnerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
