package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gobilling/internal/platform/cache"
)

type RateLimitMiddleware struct {
	cache            *cache.Cache
	readPerMinute    int
	writePerMinute   int
}

func NewRateLimitMiddleware(cache *cache.Cache, readPerMinute, writePerMinute int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cache:          cache,
		readPerMinute:  readPerMinute,
		writePerMinute: writePerMinute,
	}
}

func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		apiKey, ok := r.Context().Value(APIKeyContextKey).(string)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		limit := m.readPerMinute
		if r.Method == "POST" || r.Method == "PATCH" || r.Method == "DELETE" {
			limit = m.writePerMinute
		}

		key := fmt.Sprintf("ratelimit:%s:%d", apiKey, time.Now().Unix()/60)

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		count, err := m.cache.Get(ctx, key)
		if err == nil {
			if count >= fmt.Sprintf("%d", limit) {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
