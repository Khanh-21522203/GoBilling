package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gobilling/internal/platform/cache"
	"gobilling/internal/platform/database"
)

type IdempotencyMiddleware struct {
	cache *cache.Cache
	db    *database.DB
	ttl   time.Duration
}

func NewIdempotencyMiddleware(cache *cache.Cache, db *database.DB) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		cache: cache,
		db:    db,
		ttl:   24 * time.Hour,
	}
}

func (m *IdempotencyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" && r.Method != "PATCH" && r.Method != "DELETE" {
			next.ServeHTTP(w, r)
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		requestHash := hashRequest(r.Method, r.URL.Path, body)
		cacheKey := fmt.Sprintf("idempotency:%s:%s", idempotencyKey, requestHash)

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		cached, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var response IdempotentResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				if response.Status == "processing" {
					http.Error(w, "request already in progress", http.StatusConflict)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(response.StatusCode)
				w.Write([]byte(response.Body))
				return
			}
		}

		processingResponse := IdempotentResponse{
			Status:     "processing",
			StatusCode: 0,
			Body:       "",
		}
		processingJSON, _ := json.Marshal(processingResponse)
		_ = m.cache.Set(ctx, cacheKey, string(processingJSON), m.ttl)

		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(rec, r)

		completedResponse := IdempotentResponse{
			Status:     "completed",
			StatusCode: rec.statusCode,
			Body:       rec.body.String(),
		}
		completedJSON, _ := json.Marshal(completedResponse)
		_ = m.cache.Set(ctx, cacheKey, string(completedJSON), m.ttl)

		if rec.statusCode >= 500 {
			_ = m.cache.Del(ctx, cacheKey)
		}
	})
}

type IdempotentResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func hashRequest(method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte(path))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
