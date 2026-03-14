package generator

const handlerTmpl = `package handlers

import (
	"net/http"

	"github.com/ginger-framework/ginger/pkg/router"
)

// {{.NameTitle}}Handler handles HTTP requests for {{.NamePlural}}.
type {{.NameTitle}}Handler struct {
	// svc {{.NameTitle}}Service
}

// New{{.NameTitle}}Handler creates a new {{.NameTitle}}Handler.
func New{{.NameTitle}}Handler( /* svc {{.NameTitle}}Service */ ) *{{.NameTitle}}Handler {
	return &{{.NameTitle}}Handler{}
}

// Register mounts the {{.NameTitle}} routes on the given router group.
func (h *{{.NameTitle}}Handler) Register(r *router.Router) {
	g := r.Group("/{{.NamePlural}}")
	g.GET("/", h.list)
	g.GET("/{id}", h.get)
	g.POST("/", h.create)
	g.PUT("/{id}", h.update)
	g.DELETE("/{id}", h.delete)
}

func (h *{{.NameTitle}}Handler) list(w http.ResponseWriter, r *http.Request) {
	router.JSON(w, http.StatusOK, []any{})
}

func (h *{{.NameTitle}}Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	router.JSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *{{.NameTitle}}Handler) create(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := router.Decode(r, &body); err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusCreated, body)
}

func (h *{{.NameTitle}}Handler) update(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := router.Decode(r, &body); err != nil {
		router.Error(w, err)
		return
	}
	router.JSON(w, http.StatusOK, body)
}

func (h *{{.NameTitle}}Handler) delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
`

const serviceTmpl = `package services

import "context"

// {{.NameTitle}}Service defines the business logic for {{.NamePlural}}.
type {{.NameTitle}}Service interface {
	List(ctx context.Context) ([]any, error)
	Get(ctx context.Context, id string) (any, error)
	Create(ctx context.Context, input any) (any, error)
	Update(ctx context.Context, id string, input any) (any, error)
	Delete(ctx context.Context, id string) error
}

type {{.Name}}Service struct {
	// repo {{.NameTitle}}Repository
}

// New{{.NameTitle}}Service creates a new {{.Name}}Service.
func New{{.NameTitle}}Service( /* repo {{.NameTitle}}Repository */ ) {{.NameTitle}}Service {
	return &{{.Name}}Service{}
}

func (s *{{.Name}}Service) List(ctx context.Context) ([]any, error) {
	return []any{}, nil
}

func (s *{{.Name}}Service) Get(ctx context.Context, id string) (any, error) {
	return map[string]string{"id": id}, nil
}

func (s *{{.Name}}Service) Create(ctx context.Context, input any) (any, error) {
	return input, nil
}

func (s *{{.Name}}Service) Update(ctx context.Context, id string, input any) (any, error) {
	return input, nil
}

func (s *{{.Name}}Service) Delete(ctx context.Context, id string) error {
	return nil
}
`

const repositoryTmpl = `package repositories

import (
	"context"
	"database/sql"
)

// {{.NameTitle}}Repository defines data access for {{.NamePlural}}.
type {{.NameTitle}}Repository interface {
	FindAll(ctx context.Context) ([]any, error)
	FindByID(ctx context.Context, id string) (any, error)
	Save(ctx context.Context, entity any) (any, error)
	Update(ctx context.Context, id string, entity any) (any, error)
	Delete(ctx context.Context, id string) error
}

type {{.Name}}Repository struct {
	db *sql.DB
}

// New{{.NameTitle}}Repository creates a new {{.Name}}Repository.
func New{{.NameTitle}}Repository(db *sql.DB) {{.NameTitle}}Repository {
	return &{{.Name}}Repository{db: db}
}

func (r *{{.Name}}Repository) FindAll(ctx context.Context) ([]any, error) {
	return []any{}, nil
}

func (r *{{.Name}}Repository) FindByID(ctx context.Context, id string) (any, error) {
	return map[string]string{"id": id}, nil
}

func (r *{{.Name}}Repository) Save(ctx context.Context, entity any) (any, error) {
	return entity, nil
}

func (r *{{.Name}}Repository) Update(ctx context.Context, id string, entity any) (any, error) {
	return entity, nil
}

func (r *{{.Name}}Repository) Delete(ctx context.Context, id string) error {
	return nil
}
`

const modelTmpl = `package models

import "time"

// {{.NameTitle}} is the domain model for {{.NamePlural}}.
type {{.NameTitle}} struct {
	ID        string    ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// Create{{.NameTitle}}Input is the payload for creating a {{.Name}}.
type Create{{.NameTitle}}Input struct {
	// TODO: add fields
}

// Update{{.NameTitle}}Input is the payload for updating a {{.Name}}.
type Update{{.NameTitle}}Input struct {
	// TODO: add fields
}
`

const handlerTestTmpl = `package handlers_test

import (
	"net/http"
	"testing"

	"github.com/ginger-framework/ginger/pkg/testhelper"
)

// Test{{.NameTitle}}Handler uses table-driven tests as recommended by
// The Go Programming Language (Donovan & Kernighan, §11.2).
func Test{{.NameTitle}}Handler(t *testing.T) {
	// TODO: replace http.NotFoundHandler() with the real handler once wired:
	//   svc := &mock{{.NameTitle}}Service{}
	//   h := New{{.NameTitle}}Handler(svc)
	//   r := router.New()
	//   h.Register(r)

	tests := []struct {
		name       string
		method     string
		path       string
		body       any
		wantStatus int
	}{
		{
			name:       "list {{.NamePlural}}",
			method:     http.MethodGet,
			path:       "/{{.NamePlural}}/",
			wantStatus: http.StatusNotFound, // replace with http.StatusOK
		},
		{
			name:   "create {{.Name}}",
			method: http.MethodPost,
			path:   "/{{.NamePlural}}/",
			body: map[string]any{
				// TODO: add valid fields
			},
			wantStatus: http.StatusNotFound, // replace with http.StatusCreated
		},
		{
			name:       "get {{.Name}} by id",
			method:     http.MethodGet,
			path:       "/{{.NamePlural}}/123",
			wantStatus: http.StatusNotFound, // replace with http.StatusOK
		},
		{
			name:       "delete {{.Name}}",
			method:     http.MethodDelete,
			path:       "/{{.NamePlural}}/123",
			wantStatus: http.StatusNotFound, // replace with http.StatusNoContent
		},
	}

	handler := http.NotFoundHandler() // TODO: replace with real handler

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testhelper.NewRequest(t, handler, tt.method, tt.path)
			if tt.body != nil {
				req = req.WithBody(tt.body)
			}
			rec := req.Do()
			testhelper.AssertStatus(t, rec, tt.wantStatus)
		})
	}
}
`
