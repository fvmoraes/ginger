package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeChecker struct {
	name string
	err  error
}

func (f *fakeChecker) Name() string                    { return f.name }
func (f *fakeChecker) Check(ctx context.Context) error { return f.err }

func TestAllHealthyReturns200(t *testing.T) {
	h := New()
	h.Register(&fakeChecker{name: "db", err: nil})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Healthy {
		t.Fatal("expected healthy=true")
	}
	if resp.Checks[0].Error != "" || resp.Checks[0].ErrorCode != "" {
		t.Fatal("healthy check must not carry error fields")
	}
}

func TestFailingCheckerReturns503WithoutLeakingDSN(t *testing.T) {
	h := New()
	h.Register(&fakeChecker{name: "db", err: errors.New("dial tcp user=secret password=hunter2@db.internal:5432: refused")})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 503 {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	body := w.Body.String()
	if resp.Healthy {
		t.Fatal("expected healthy=false")
	}
	// GIN-022: err.Error() must NOT reach the payload.
	if containsAny(body, []string{"hunter2", "db.internal", "refused"}) {
		t.Fatalf("sensitive error leaked into /health: %s", body)
	}
	if resp.Checks[0].ErrorCode != "CHECK_FAILED" {
		t.Fatalf("ErrorCode = %q, want CHECK_FAILED", resp.Checks[0].ErrorCode)
	}
	if resp.Checks[0].Error != "CHECK_FAILED" {
		t.Fatalf("legacy Error field must carry the stable code, got %q", resp.Checks[0].Error)
	}
}

func TestTimeoutErrorCode(t *testing.T) {
	h := New()
	h.Register(&fakeChecker{name: "slow", err: context.DeadlineExceeded})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Checks[0].ErrorCode != "TIMEOUT" {
		t.Fatalf("ErrorCode = %q, want TIMEOUT", resp.Checks[0].ErrorCode)
	}
}

func TestEmptyRegistryHealthy(t *testing.T) {
	w := httptest.NewRecorder()
	New().ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	if !containsAny(w.Body.String(), []string{`"checks":[]`}) {
		t.Fatal("checks must serialize as [] not null")
	}
}

func TestChecksRunConcurrently(t *testing.T) {
	h := New()
	slow := func() Checker {
		return &slowChecker{}
	}
	h.Register(slow())
	h.Register(&fakeChecker{name: "fast"})
	start := time.Now()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	elapsed := time.Since(start)
	if elapsed > 40*time.Millisecond {
		t.Fatalf("checks ran sequentially (%v) — expected concurrent", elapsed)
	}
	if w.Code != 503 {
		t.Fatalf("slow checker must fail the endpoint, got %d", w.Code)
	}
}

type slowChecker struct{}

func (s *slowChecker) Name() string { return "slow" }
func (s *slowChecker) Check(ctx context.Context) error {
	time.Sleep(30 * time.Millisecond)
	return errors.New("slow failure")
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && index(s, sub) >= 0
}

func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
