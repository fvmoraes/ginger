// Package health provides a health check endpoint and checker registry.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Checker is implemented by any component that can report its health.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Status represents the health status of a component.
type Status struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}

// Response is the full health check response.
type Response struct {
	Healthy  bool     `json:"healthy"`
	Checks   []Status `json:"checks"`
	Duration string   `json:"duration"`
}

// Handler holds registered checkers and serves the health endpoint.
type Handler struct {
	mu       sync.RWMutex
	checkers []Checker
}

func New() *Handler {
	return &Handler{}
}

// Register adds a checker to the handler.
func (h *Handler) Register(c Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, c)
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.mu.RLock()
	checkers := h.checkers
	h.mu.RUnlock()

	resp := Response{Healthy: true}
	ctx := r.Context()

	for _, c := range checkers {
		s := Status{Name: c.Name(), Healthy: true}
		if err := c.Check(ctx); err != nil {
			s.Healthy = false
			s.Error = err.Error()
			resp.Healthy = false
		}
		resp.Checks = append(resp.Checks, s)
	}

	resp.Duration = time.Since(start).String()

	status := http.StatusOK
	if !resp.Healthy {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}
