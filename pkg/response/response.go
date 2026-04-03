// Package response provides standardised JSON envelope types for REST APIs
// consumed by frontend clients (React, Vue, Angular, etc.).
//
// All responses share a consistent shape so the frontend can handle them
// generically without per-endpoint parsing logic.
package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard success wrapper for single-resource responses.
//
//	{ "data": <T>, "meta": { ... } }
type Envelope[T any] struct {
	Data T     `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// Page is the standard wrapper for paginated list responses.
//
//	{ "data": [...], "pagination": { "page": 1, "per_page": 20, "total": 100, "total_pages": 5 } }
type Page[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds cursor/offset pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Meta holds optional response metadata (e.g. request ID, version).
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Version   string `json:"version,omitempty"`
}

// OK writes a 200 JSON envelope around data.
func OK[T any](w http.ResponseWriter, data T) {
	write(w, http.StatusOK, Envelope[T]{Data: data})
}

// Created writes a 201 JSON envelope around data.
func Created[T any](w http.ResponseWriter, data T) {
	write(w, http.StatusCreated, Envelope[T]{Data: data})
}

// Paginated writes a 200 paginated list response.
func Paginated[T any](w http.ResponseWriter, data []T, page, perPage, total int) {
	if data == nil {
		data = []T{} // never serialize null
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = total / perPage
		if total%perPage != 0 {
			totalPages++
		}
	}
	write(w, http.StatusOK, Page[T]{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// NoContent writes a 204 with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
