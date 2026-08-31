package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteMatches(t *testing.T) {
	r := New()
	hit := false
	r.GET("/ping", func(w http.ResponseWriter, req *http.Request) { hit = true })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ping", nil))
	if !hit || w.Code != 200 {
		t.Fatalf("route not matched: hit=%v code=%d", hit, w.Code)
	}
}

func TestMethodMismatchReturns405(t *testing.T) {
	r := New()
	r.GET("/only", func(w http.ResponseWriter, req *http.Request) {})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/only", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", w.Code)
	}
}

func TestUnknownRouteReturns404(t *testing.T) {
	r := New()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/nope", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestJSONAndDecodeHelpers(t *testing.T) {
	r := New()
	r.POST("/echo", func(w http.ResponseWriter, req *http.Request) {
		var in map[string]any
		if err := Decode(req, &in); err != nil {
			Error(w, err)
			return
		}
		JSON(w, 200, in)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/echo", strings.NewReader(`{"a":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"a":1`) {
		t.Fatalf("echo failed: %d %s", w.Code, w.Body.String())
	}
}
