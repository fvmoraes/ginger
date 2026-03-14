// Package errors provides standardized, structured error types for Ginger
// applications. It wraps the standard library errors package and adds HTTP
// status mapping, typed error codes, and context-preserving wrapping via %w.
//
// Usage:
//
//	return apperrors.NotFound("user not found")
//	return apperrors.Internal(fmt.Errorf("db query: %w", err))
//
//	if appErr, ok := apperrors.As(err); ok {
//	    http.Error(w, appErr.Message, appErr.HTTPStatus())
//	}
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a machine-readable error classification string.
type Code string

const (
	// CodeNotFound indicates the requested resource does not exist.
	CodeNotFound Code = "NOT_FOUND"
	// CodeBadRequest indicates the client sent an invalid request.
	CodeBadRequest Code = "BAD_REQUEST"
	// CodeUnauthorized indicates missing or invalid authentication.
	CodeUnauthorized Code = "UNAUTHORIZED"
	// CodeForbidden indicates the caller lacks permission.
	CodeForbidden Code = "FORBIDDEN"
	// CodeConflict indicates a state conflict (e.g. duplicate resource).
	CodeConflict Code = "CONFLICT"
	// CodeInternal indicates an unexpected server-side failure.
	CodeInternal Code = "INTERNAL"
	// CodeUnprocessable indicates semantically invalid input.
	CodeUnprocessable Code = "UNPROCESSABLE"
	// CodeServiceUnavailable indicates a downstream dependency is down.
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
)

// AppError is the standard structured error for Ginger applications.
// It implements the error interface and is safe to serialize as JSON.
//
// The Err field holds the underlying cause and is excluded from JSON output.
// Use Unwrap to traverse the error chain with errors.As / errors.Is.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
// Format: "[CODE] message: cause"
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause, enabling errors.As and errors.Is
// to traverse the error chain as recommended by The Go Programming Language.
func (e *AppError) Unwrap() error { return e.Err }

// Is reports whether the target error has the same Code.
// This enables errors.Is(err, apperrors.ErrNotFound) style checks.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// HTTPStatus maps the error Code to the appropriate HTTP status code.
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeConflict:
		return http.StatusConflict
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// New creates an AppError with the given code, human-readable message, and
// optional underlying cause. The cause is wrapped so errors.Is / errors.As
// can inspect the full chain.
func New(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Wrap creates an AppError that wraps an existing error with additional context.
// The original error is preserved in the chain for errors.Is / errors.As.
func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Sentinel errors for use with errors.Is comparisons.
var (
	ErrNotFound      = New(CodeNotFound, "not found", nil)
	ErrBadRequest    = New(CodeBadRequest, "bad request", nil)
	ErrUnauthorized  = New(CodeUnauthorized, "unauthorized", nil)
	ErrForbidden     = New(CodeForbidden, "forbidden", nil)
	ErrConflict      = New(CodeConflict, "conflict", nil)
	ErrInternal      = New(CodeInternal, "internal server error", nil)
	ErrUnprocessable = New(CodeUnprocessable, "unprocessable entity", nil)
)

// NotFound returns a 404 AppError with the given message.
func NotFound(msg string) *AppError { return New(CodeNotFound, msg, nil) }

// BadRequest returns a 400 AppError with the given message.
func BadRequest(msg string) *AppError { return New(CodeBadRequest, msg, nil) }

// Unauthorized returns a 401 AppError with the given message.
func Unauthorized(msg string) *AppError { return New(CodeUnauthorized, msg, nil) }

// Forbidden returns a 403 AppError with the given message.
func Forbidden(msg string) *AppError { return New(CodeForbidden, msg, nil) }

// Conflict returns a 409 AppError with the given message.
func Conflict(msg string) *AppError { return New(CodeConflict, msg, nil) }

// Internal wraps a low-level error as a 500. The original error is preserved
// in the chain but not exposed to the client.
func Internal(err error) *AppError { return New(CodeInternal, "internal server error", err) }

// Unprocessable returns a 422 AppError with the given message.
func Unprocessable(msg string) *AppError { return New(CodeUnprocessable, msg, nil) }

// As unwraps err to *AppError using errors.As from the standard library.
// This is the idiomatic way to check and extract AppError from an error chain.
func As(err error) (*AppError, bool) {
	var e *AppError
	ok := errors.As(err, &e)
	return e, ok
}

// IsCode reports whether any error in the chain is an AppError with the given Code.
func IsCode(err error, code Code) bool {
	var e *AppError
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}
