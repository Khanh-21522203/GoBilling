package middleware

import (
	"net/http"
	"strings"
)

const (
	MaxRequestBodySize = 1 * 1024 * 1024
)

type ValidationMiddleware struct{}

func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

func (m *ValidationMiddleware) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PATCH" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		if r.ContentLength > MaxRequestBodySize {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

		next.ServeHTTP(w, r)
	})
}
