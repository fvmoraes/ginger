package handlers

import (
	"net/http"

	"github.com/fvmoraes/ginger/example/internal/api/services"
	"github.com/fvmoraes/ginger/example/internal/models"
	apperrors "github.com/fvmoraes/ginger/pkg/errors"
	"github.com/fvmoraes/ginger/pkg/router"
)

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	svc services.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register mounts user routes on the given router group.
func (h *UserHandler) Register(r *router.Router) {
	g := r.Group("/users")
	g.GET("/", h.list)
	g.GET("/{id}", h.get)
	g.POST("/", h.create)
	g.PUT("/{id}", h.update)
	g.DELETE("/{id}", h.delete)
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.svc.Get(r.Context(), id)
	if err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateUserInput
	if err := router.Decode(r, &input); err != nil {
		router.Error(w, apperrors.BadRequest("invalid request body"))
		return
	}
	user, err := h.svc.Create(r.Context(), &input)
	if err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusCreated, user)
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input models.UpdateUserInput
	if err := router.Decode(r, &input); err != nil {
		router.Error(w, apperrors.BadRequest("invalid request body"))
		return
	}
	user, err := h.svc.Update(r.Context(), id, &input)
	if err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		router.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
