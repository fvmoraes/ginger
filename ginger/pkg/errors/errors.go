// Package errors provides standardized error types for Ginger applications.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code represents an application error code.
type Code string

const (
	CodeNotFound           Code = "NOT_FOUND"
	CodeBadRequest         Code = "BAD_REQUEST"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeConflict           Code = "CONFLICT"
	CodeInternal           Code = "INTERNAL"
	CodeUnprocessable      Code = "UNPROCESSABLE"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
)

// AppError is the standard application error.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// HTTPStatus maps error codes to HTTP status codes.
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

func New(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func NotFound(msg string) *AppError      { return New(CodeNotFound, msg, nil) }
func BadRequest(msg string) *AppError    { return New(CodeBadRequest, msg, nil) }
func Unauthorized(msg string) *AppError  { return New(CodeUnauthorized, msg, nil) }
func Forbidden(msg string) *AppError     { return New(CodeForbidden, msg, nil) }
func Conflict(msg string) *AppError      { return New(CodeConflict, msg, nil) }
func Internal(err error) *AppError       { return New(CodeInternal, "internal server error", err) }
func Unprocessable(msg string) *AppError { return New(CodeUnprocessable, msg, nil) }

// As unwraps to *AppError.
func As(err error) (*AppError, bool) {
	var e *AppError
	ok := errors.As(err, &e)
	return e, ok
}
