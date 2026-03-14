// Package router provides a thin wrapper around net/http ServeMux with
// route grouping, middleware chaining, and JSON response helpers.
package router

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/ginger-framework/ginger/pkg/errors"
	"github.com/ginger-framework/ginger/pkg/middleware"
)

// Router wraps http.ServeMux with group and middleware support.
type Router struct {
	mux         *http.ServeMux
	middlewares []middleware.Func
	prefix      string
}

// New creates a new Router.
func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Use appends global middlewares.
func (r *Router) Use(mw ...middleware.Func) {
	r.middlewares = append(r.middlewares, mw...)
}

// Group creates a sub-router with a path prefix and optional extra middlewares.
func (r *Router) Group(prefix string, mw ...middleware.Func) *Router {
	return &Router{
		mux:         r.mux,
		middlewares: append(append([]middleware.Func{}, r.middlewares...), mw...),
		prefix:      r.prefix + prefix,
	}
}

// Handle registers a handler with the given method and pattern.
func (r *Router) Handle(method, pattern string, h http.HandlerFunc) {
	full := method + " " + r.prefix + pattern
	chain := middleware.Chain(r.middlewares...)(h)
	r.mux.Handle(full, chain)
}

func (r *Router) GET(pattern string, h http.HandlerFunc)    { r.Handle(http.MethodGet, pattern, h) }
func (r *Router) POST(pattern string, h http.HandlerFunc)   { r.Handle(http.MethodPost, pattern, h) }
func (r *Router) PUT(pattern string, h http.HandlerFunc)    { r.Handle(http.MethodPut, pattern, h) }
func (r *Router) PATCH(pattern string, h http.HandlerFunc)  { r.Handle(http.MethodPatch, pattern, h) }
func (r *Router) DELETE(pattern string, h http.HandlerFunc) { r.Handle(http.MethodDelete, pattern, h) }

// HandleRaw registers a handler directly on the mux without middleware.
func (r *Router) HandleRaw(pattern string, h http.Handler) {
	r.mux.Handle(pattern, h)
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// JSON writes a JSON response.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// Error writes a standardized JSON error response.
func Error(w http.ResponseWriter, err error) {
	if appErr, ok := apperrors.As(err); ok {
		JSON(w, appErr.HTTPStatus(), appErr)
		return
	}
	JSON(w, http.StatusInternalServerError, apperrors.Internal(err))
}

// Decode decodes a JSON request body into v.
func Decode(r *http.Request, v any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	return nil
}
