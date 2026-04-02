// Package middleware provides common HTTP middlewares for Ginger applications.
// All middlewares follow the standard func(http.Handler) http.Handler signature
// and can be composed with Chain.
package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fvmoraes/ginger/pkg/logger"
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
			requestLog := log.With(
				"request_id", RequestIDFromContext(r.Context()),
				"http.method", r.Method,
				"http.path", r.URL.Path,
				"http.remote_addr", r.RemoteAddr,
				"http.user_agent", r.UserAgent(),
			)
			ctx := logger.WithContext(r.Context(), requestLog)

			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(ctx))
			logger.FromContext(ctx).Info("request_finished",
				"http.status", rw.status,
				"http.duration", time.Since(start).String(),
			)
		})
	}
}

// Recover catches panics, logs the stack, and returns a structured JSON 500.
// Using a structured response keeps error format consistent with router.Error.
func Recover(log *logger.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestLog := log.With(
						"request_id", RequestIDFromContext(r.Context()),
						"http.method", r.Method,
						"http.path", r.URL.Path,
						"http.remote_addr", r.RemoteAddr,
						"http.user_agent", r.UserAgent(),
					)
					requestLog.Error("panic_recovered", "error", rec, "http.path", r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
						"code":    "INTERNAL",
						"message": "internal server error",
					})
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
			next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
		})
	}
}

// CORSConfig holds the configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is the list of origins allowed to make cross-origin requests.
	// Use ["*"] to allow all origins. Defaults to ["*"].
	AllowedOrigins []string
	// AllowedHeaders is the list of headers allowed in cross-origin requests.
	// Defaults to ["Content-Type", "Authorization", "X-Request-ID"].
	AllowedHeaders []string
	// AllowedMethods is the list of HTTP methods allowed.
	// Defaults to ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"].
	AllowedMethods []string
	// AllowCredentials sets Access-Control-Allow-Credentials.
	// NOTE: cannot be used together with AllowedOrigins: ["*"].
	AllowCredentials bool
	// MaxAge sets Access-Control-Max-Age in seconds (preflight cache).
	// Defaults to 0 (no cache header sent).
	MaxAge int
}

// CORS adds CORS headers using the provided config.
// When called with no arguments it defaults to allow-all (origin: *).
//
//	middleware.CORS()                          // allow all
//	middleware.CORS(CORSConfig{...})           // custom config
func CORS(cfg ...CORSConfig) Func {
	c := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}
	if len(cfg) > 0 {
		c = cfg[0]
		if len(c.AllowedOrigins) == 0 {
			c.AllowedOrigins = []string{"*"}
		}
		if len(c.AllowedHeaders) == 0 {
			c.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
		}
		if len(c.AllowedMethods) == 0 {
			c.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
		}
	}

	methods := strings.Join(c.AllowedMethods, ", ")
	headers := strings.Join(c.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := resolveOrigin(origin, c.AllowedOrigins)

			h := w.Header()
			h.Set("Access-Control-Allow-Origin", allowed)
			h.Set("Access-Control-Allow-Methods", methods)
			h.Set("Access-Control-Allow-Headers", headers)
			if c.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if c.MaxAge > 0 {
				h.Set("Access-Control-Max-Age", strconv.Itoa(c.MaxAge))
			}
			// Vary header ensures caches don't serve wrong origin responses.
			h.Add("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// resolveOrigin returns the matching allowed origin or "*".
func resolveOrigin(requestOrigin string, allowed []string) string {
	if len(allowed) == 1 && allowed[0] == "*" {
		return "*"
	}
	for _, o := range allowed {
		if o == requestOrigin {
			return requestOrigin
		}
	}
	// Return first allowed origin as fallback (browser will block mismatches).
	if len(allowed) > 0 {
		return allowed[0]
	}
	return "*"
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
