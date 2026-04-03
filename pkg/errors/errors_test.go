package errors

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTooManyRequestsMapsToHTTP429(t *testing.T) {
	err := TooManyRequests("slow down")

	if err.Code != CodeTooManyRequests {
		t.Fatalf("expected code %q, got %q", CodeTooManyRequests, err.Code)
	}
	if got := err.HTTPStatus(); got != http.StatusTooManyRequests {
		t.Fatalf("expected HTTP 429, got %d", got)
	}
}

func TestIsCodeMatchesWrappedTooManyRequests(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", TooManyRequests("slow down"))

	if !IsCode(err, CodeTooManyRequests) {
		t.Fatal("expected IsCode to match wrapped too many requests error")
	}
}
