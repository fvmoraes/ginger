// Package health provides a health check endpoint and checker registry.
// Checks run concurrently so a slow dependency does not delay the others.
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

// Status represents the health status of a single component.
type Status struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}

// Response is the full health check payload.
type Response struct {
	Healthy  bool     `json:"healthy"`
	Checks   []Status `json:"checks"`
	Duration string   `json:"duration"`
}

// Handler holds registered checkers and serves GET /health.
type Handler struct {
	mu       sync.RWMutex
	checkers []Checker
}

// New returns a Handler with no registered checkers.
func New() *Handler { return &Handler{} }

// Register adds a Checker. Safe to call concurrently.
func (h *Handler) Register(c Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, c)
}

// ServeHTTP implements http.Handler.
// All checkers run concurrently; the response is 200 if all pass, 503 otherwise.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	h.mu.RLock()
	checkers := h.checkers
	h.mu.RUnlock()

	// Run all checks concurrently — Effective Go §Concurrency.
	statuses := make([]Status, len(checkers))
	var wg sync.WaitGroup
	wg.Add(len(checkers))
	for i, c := range checkers {
		i, c := i, c // capture loop vars (pre-Go 1.22 safety)
		go func() {
			defer wg.Done()
			s := Status{Name: c.Name(), Healthy: true}
			if err := c.Check(r.Context()); err != nil {
				s.Healthy = false
				s.Error = err.Error()
			}
			statuses[i] = s
		}()
	}
	wg.Wait()

	resp := Response{
		Healthy:  true,
		Checks:   statuses, // never nil — serialises as [] not null
		Duration: time.Since(start).String(),
	}
	for _, s := range statuses {
		if !s.Healthy {
			resp.Healthy = false
			break
		}
	}

	code := http.StatusOK
	if !resp.Healthy {
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}
