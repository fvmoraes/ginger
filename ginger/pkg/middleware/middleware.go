// Package middleware provides common HTTP middlewares for Ginger applications.
package middleware

import (
	"net/http"
	"time"

	"github.com/ginger-framework/ginger/pkg/logger"
)

// Func is a standard middleware type.
type Func func(http.Handler) http.Handler

// Chain composes multiple middlewares into one.
func Chain(middlewares ...Func) Func {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Logger logs each request with method, path, status, and duration.
func Logger(log *logger.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// Recover catches panics and returns a 500.
func Recover(log *logger.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "error", rec, "path", r.URL.Path)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID injects a request ID into the context and response headers.
func RequestID() Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = generateID()
			}
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(
				withRequestID(r.Context(), id),
			))
		})
	}
}

// CORS adds basic CORS headers.
func CORS(allowedOrigins ...string) Func {
	origins := "*"
	if len(allowedOrigins) > 0 {
		origins = allowedOrigins[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
