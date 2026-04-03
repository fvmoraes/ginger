package testhelper

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/fvmoraes/ginger/pkg/response"
)

func TestRequestHelpersSupportHeadersAndJSONResponses(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("expected Authorization header to be propagated, got %q", got)
		}
		response.OK(w, map[string]string{"name": "Alice"})
	})

	rec := NewRequest(t, handler, http.MethodGet, "/users").
		WithHeader("Authorization", "Bearer token").
		Do()

	AssertStatus(t, rec, http.StatusOK)
	AssertHeader(t, rec, "Content-Type", "application/json")

	var got struct {
		Data map[string]string `json:"data"`
	}
	DecodeJSON(t, rec, &got)
	if got.Data["name"] != "Alice" {
		t.Fatalf("expected decoded name Alice, got %q", got.Data["name"])
	}
}

func TestRequestHelperWithBodyEncodesJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		response.Created(w, input)
	})

	rec := NewRequest(t, handler, http.MethodPost, "/users").
		WithBody(map[string]string{"name": "Alice"}).
		Do()

	AssertStatus(t, rec, http.StatusCreated)
	AssertHeader(t, rec, "Content-Type", "application/json")
}
