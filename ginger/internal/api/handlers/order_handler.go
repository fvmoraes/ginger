package handlers

import (
	"net/http"

	"github.com/ginger-framework/ginger/pkg/router"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	// svc OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler( /* svc OrderService */ ) *OrderHandler {
	return &OrderHandler{}
}

// Register mounts the Order routes on the given router group.
func (h *OrderHandler) Register(r *router.Router) {
	g := r.Group("/orders")
	g.GET("/", h.list)
	g.GET("/{id}", h.get)
	g.POST("/", h.create)
	g.PUT("/{id}", h.update)
	g.DELETE("/{id}", h.delete)
}

func (h *OrderHandler) list(w http.ResponseWriter, r *http.Request) {
	router.JSON(w, http.StatusOK, []any{})
}

func (h *OrderHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	router.JSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *OrderHandler) create(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := router.Decode(r, &body); err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusCreated, body)
}

func (h *OrderHandler) update(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := router.Decode(r, &body); err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusOK, body)
}

func (h *OrderHandler) delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
