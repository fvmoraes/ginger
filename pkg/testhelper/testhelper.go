// Package testhelper provides utilities for testing Ginger HTTP handlers.
package testhelper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Request builds and executes an HTTP test request against a handler.
type Request struct {
	t       *testing.T
	handler http.Handler
	method  string
	path    string
	body    any
	headers map[string]string
}

// NewRequest creates a test request builder.
func NewRequest(t *testing.T, handler http.Handler, method, path string) *Request {
	t.Helper()
	return &Request{
		t:       t,
		handler: handler,
		method:  method,
		path:    path,
		headers: make(map[string]string),
	}
}

// WithBody sets a JSON body.
func (r *Request) WithBody(v any) *Request {
	r.body = v
	return r
}

// WithHeader adds a request header.
func (r *Request) WithHeader(key, value string) *Request {
	r.headers[key] = value
	return r
}

// Do executes the request and returns the response recorder.
func (r *Request) Do() *httptest.ResponseRecorder {
	r.t.Helper()

	var buf bytes.Buffer
	if r.body != nil {
		if err := json.NewEncoder(&buf).Encode(r.body); err != nil {
			r.t.Fatalf("testhelper: encode body: %v", err)
		}
	}

	req := httptest.NewRequest(r.method, r.path, &buf)
	if r.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	r.handler.ServeHTTP(rec, req)
	return rec
}

// DecodeJSON decodes the response body into v.
func DecodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("testhelper: decode response: %v", err)
	}
}

// AssertStatus fails the test if the response status doesn't match.
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Errorf("expected status %d, got %d\nbody: %s", want, rec.Code, rec.Body.String())
	}
}
