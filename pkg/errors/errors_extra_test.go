package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestConstructorCodesAndStatuses(t *testing.T) {
	cases := []struct {
		err    *AppError
		code   Code
		status int
	}{
		{NotFound("nope"), CodeNotFound, http.StatusNotFound},
		{BadRequest("bad"), CodeBadRequest, http.StatusBadRequest},
		{Unauthorized("un"), CodeUnauthorized, http.StatusUnauthorized},
		{Forbidden("forb"), CodeForbidden, http.StatusForbidden},
		{Conflict("conf"), CodeConflict, http.StatusConflict},
		{Unprocessable("unp"), CodeUnprocessable, http.StatusUnprocessableEntity},
		{TooManyRequests("tmr"), CodeTooManyRequests, http.StatusTooManyRequests},
		{Internal(errors.New("x")), CodeInternal, http.StatusInternalServerError},
	}
	for _, c := range cases {
		if c.err.HTTPStatus() != c.status {
			t.Errorf("%v: status %d, want %d", c.err.Code, c.err.HTTPStatus(), c.status)
		}
		if c.err.Error() == "" {
			t.Errorf("%v: empty message", c.err.Code)
		}
	}
}

func TestWrapPreservesCause(t *testing.T) {
	cause := errors.New("root cause")
	wrapped := Wrap(CodeInternal, "wrapped", cause)
	if !errors.Is(wrapped, cause) {
		t.Fatal("Unwrap chain broken")
	}
	if !IsCode(wrapped, CodeInternal) {
		t.Fatal("IsCode failed on wrapped error")
	}
}

func TestAsExtractsAppError(t *testing.T) {
	appErr := NotFound("missing")
	target, ok := As(appErr)
	if !ok {
		t.Fatal("As must extract *AppError")
	}
	if target.Code != CodeNotFound {
		t.Fatalf("code = %v", target.Code)
	}
	plain := errors.New("plain")
	if _, ok := As(plain); ok {
		t.Fatal("As must fail for plain errors")
	}
	if IsCode(plain, CodeNotFound) {
		t.Fatal("IsCode must fail for plain errors")
	}
	if !IsCode(NotFound("x"), CodeNotFound) {
		t.Fatal("IsCode must match code")
	}
}
