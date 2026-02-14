package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type contextKey string

const (
	APIKeyContextKey contextKey = "api_key"
)

type APIKeyMiddleware struct {
}

func NewAPIKeyMiddleware() *APIKeyMiddleware {
	return &APIKeyMiddleware{}
}

func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		apiKey := parts[1]

		if !strings.HasPrefix(apiKey, "sk_live_") && !strings.HasPrefix(apiKey, "sk_test_") {
			http.Error(w, "invalid api key format", http.StatusUnauthorized)
			return
		}

		hash := sha256.Sum256([]byte(apiKey))
		keyHash := hex.EncodeToString(hash[:])

		ctx := context.WithValue(r.Context(), APIKeyContextKey, keyHash)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
