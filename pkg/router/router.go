// Package router provides a thin wrapper around net/http ServeMux with
// route grouping, middleware chaining, and JSON response helpers.
package router

import (
	"encoding/json"
	"io"
	"net/http"

	apperrors "github.com/fvmoraes/ginger/pkg/errors"
	"github.com/fvmoraes/ginger/pkg/middleware"
)

// Router wraps http.ServeMux with group and middleware support.
type Router struct {
	mux         *http.ServeMux
	middlewares []middleware.Func
	prefix      string
}

// New creates a new Router backed by a fresh http.ServeMux.
func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Use appends global middlewares applied to every route on this router.
func (r *Router) Use(mw ...middleware.Func) {
	r.middlewares = append(r.middlewares, mw...)
}

// Group creates a sub-router sharing the same mux but with a path prefix
// and optional additional middlewares.
func (r *Router) Group(prefix string, mw ...middleware.Func) *Router {
	combined := make([]middleware.Func, len(r.middlewares), len(r.middlewares)+len(mw))
	copy(combined, r.middlewares)
	combined = append(combined, mw...)
	return &Router{
		mux:         r.mux,
		middlewares: combined,
		prefix:      r.prefix + prefix,
	}
}

// Handle registers handler h for the given HTTP method and pattern.
func (r *Router) Handle(method, pattern string, h http.HandlerFunc) {
	full := method + " " + r.prefix + pattern
	r.mux.Handle(full, middleware.Chain(r.middlewares...)(h))
}

// Convenience wrappers for common HTTP methods.
func (r *Router) GET(pattern string, h http.HandlerFunc)    { r.Handle(http.MethodGet, pattern, h) }
func (r *Router) POST(pattern string, h http.HandlerFunc)   { r.Handle(http.MethodPost, pattern, h) }
func (r *Router) PUT(pattern string, h http.HandlerFunc)    { r.Handle(http.MethodPut, pattern, h) }
func (r *Router) PATCH(pattern string, h http.HandlerFunc)  { r.Handle(http.MethodPatch, pattern, h) }
func (r *Router) DELETE(pattern string, h http.HandlerFunc) { r.Handle(http.MethodDelete, pattern, h) }

// HandleRaw registers a handler directly on the mux, bypassing middleware.
// Use for internal endpoints like /health that should not be logged or traced.
func (r *Router) HandleRaw(pattern string, h http.Handler) {
	r.mux.Handle(pattern, h)
}

// ServeHTTP implements http.Handler, delegating to the underlying mux.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// JSON writes v as a JSON response with the given HTTP status code.
// Content-Type is set to application/json automatically.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// Error writes a standardised JSON error response.
//   - *AppError values use their own Code and HTTPStatus.
//   - All other errors are wrapped as 500 Internal; the original cause is
//     not exposed to the client (Effective Go §Errors: don't leak internals).
func Error(w http.ResponseWriter, err error) {
	if appErr, ok := apperrors.As(err); ok {
		JSON(w, appErr.HTTPStatus(), appErr)
		return
	}
	JSON(w, http.StatusInternalServerError, apperrors.Internal(err))
}

// Decode decodes a JSON request body into v.
// The body is limited to 1 MiB to prevent resource exhaustion.
// Returns a BadRequest AppError on malformed JSON.
func Decode(r *http.Request, v any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	return nil
}
