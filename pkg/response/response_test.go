package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaginatedHandlesZeroPerPage(t *testing.T) {
	rec := httptest.NewRecorder()

	Paginated(rec, []string{"a", "b"}, 1, 0, 2)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var page Page[string]
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if page.Pagination.TotalPages != 0 {
		t.Fatalf("expected total_pages 0, got %d", page.Pagination.TotalPages)
	}
}
