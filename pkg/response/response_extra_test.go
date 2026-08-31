package response

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestOKCreatedNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, map[string]string{"x": "y"})
	if w.Code != 200 || !json.Valid(w.Body.Bytes()) {
		t.Fatalf("OK: %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	Created(w, map[string]string{"id": "1"})
	if w.Code != 201 {
		t.Fatalf("Created: %d", w.Code)
	}

	w = httptest.NewRecorder()
	NoContent(w)
	if w.Code != 204 {
		t.Fatalf("NoContent: %d", w.Code)
	}
}

func TestPaginatedEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []string{"a", "b"}, 1, 2, 5)
	var env struct {
		Data       []string `json:"data"`
		Pagination struct {
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
			Total   int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("envelope: %v", err)
	}
	if len(env.Data) != 2 || env.Pagination.Page != 1 || env.Pagination.Total != 5 {
		t.Fatalf("unexpected envelope: %+v", env)
	}
}

func TestPaginatedZeroPerPageClamps(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []string{}, 1, 0, 0)
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
}
