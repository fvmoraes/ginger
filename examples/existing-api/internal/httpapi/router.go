package httpapi

import (
	"net/http"

	"example.com/existing-api/internal/httpapi/handlers"
)

func Register(mux *http.ServeMux) {
	// ginger:begin routes
	// ginger:route GET /users/{id}
	mux.HandleFunc("GET /users/{id}", handlers.GetUser)
	// ginger:end routes
}
