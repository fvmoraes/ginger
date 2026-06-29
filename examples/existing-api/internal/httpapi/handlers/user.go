package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/existing-api/internal/core"
)

// GetUser is existing application code; Ginger must never replace this file.
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(core.FindUser(r.PathValue("id")))
}
