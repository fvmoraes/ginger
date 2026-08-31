package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fvmoraes/ginger/pkg/logger"
	"github.com/fvmoraes/ginger/pkg/sse"
)

func newTestLogger() *logger.Logger { return logger.New("error", "json") }

func TestChainOrder(t *testing.T) {
	var order []string
	mk := func(name string) Func {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(mk("first"), mk("second"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	want := "first,second,handler"
	if got := strings.Join(order, ","); got != want {
		t.Fatalf("chain order = %q, want %q", got, want)
	}
}

func TestRecoverReturnsStructured500(t *testing.T) {
	h := Recover(newTestLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not json: %v", err)
	}
	if body["code"] != "INTERNAL" || body["message"] != "internal server error" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestRequestIDGeneratedAndEchoed(t *testing.T) {
	var captured string
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = RequestIDFromContext(r.Context())
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if captured == "" {
		t.Fatal("request id not injected")
	}
	if w.Header().Get("X-Request-ID") != captured {
		t.Fatal("request id not echoed in header")
	}
}

func TestCORSPreflight(t *testing.T) {
	h := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight must short-circuit the handler")
	}))
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://app.example.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("missing Allow-Origin header")
	}
}

// GIN-013: SSE (Flusher) behind Logger/Recover works via Unwrap +
// http.ResponseController.
func TestStreamingBehindDefaultMiddleware(t *testing.T) {
	handler := Chain(RequestID(), Recover(newTestLogger()), Logger(newTestLogger()))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := sse.New(w)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			_ = stream.Send(sse.Event{Data: "hello"})
		}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := make([]byte, 128)
	n, _ := resp.Body.Read(body)
	if !strings.Contains(string(body[:n]), "data: hello") {
		t.Fatalf("streamed content missing (read %d bytes): %q", n, string(body[:n]))
	}
}
