# Copy-Paste Ready — Código Pronto

[← Voltar ao Índice](./README.md)

Exemplos prontos para copiar e colar. Zero configuração, máxima velocidade.

---

## 🚀 Setup Completo em 1 Minuto

### main.go Básico

```go
package main

import (
    gingerapp "github.com/fvmoraes/ginger/pkg/app"
    "github.com/fvmoraes/ginger/pkg/config"
    "github.com/fvmoraes/ginger/pkg/middleware"
    "github.com/fvmoraes/ginger/pkg/router"
    "net/http"
)

func main() {
    cfg, _ := config.Load("configs/app.yaml")
    app := gingerapp.New(cfg)
    
    app.Router.Use(
        middleware.Logger(app.Logger),
        middleware.RequestID(),
        middleware.Recover(app.Logger),
        middleware.CORS(),
    )
    
    app.Router.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
        router.JSON(w, 200, map[string]string{"message": "pong"})
    })
    
    app.Run()
}
```

### configs/app.yaml

```yaml
app:
  name: foobar
  env: development
  version: 0.1.0

http:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30

log:
  level: info
  format: json
```

---

## 📦 Handler Completo

```go
package handlers

import (
    "net/http"
    "github.com/fvmoraes/ginger/pkg/router"
    "github.com/fvmoraes/ginger/pkg/response"
    apperrors "github.com/fvmoraes/ginger/pkg/errors"
)

type UserHandler struct {
    service UserService
}

func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{service: service}
}

func (h *UserHandler) Register(r *router.Router) {
    g := r.Group("/users")
    g.GET("/", h.List)
    g.GET("/{id}", h.Get)
    g.POST("/", h.Create)
    g.PUT("/{id}", h.Update)
    g.DELETE("/{id}", h.Delete)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    users, err := h.service.List(r.Context())
    if err != nil {
        router.Error(w, err)
        return
    }
    response.OK(w, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user, err := h.service.GetByID(r.Context(), id)
    if err != nil {
        router.Error(w, err)
        return
    }
    response.OK(w, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateUserInput
    if err := router.Decode(r, &input); err != nil {
        router.Error(w, err)
        return
    }
    
    user, err := h.service.Create(r.Context(), input)
    if err != nil {
        router.Error(w, err)
        return
    }
    
    response.Created(w, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    var input UpdateUserInput
    if err := router.Decode(r, &input); err != nil {
        router.Error(w, err)
        return
    }
    
    user, err := h.service.Update(r.Context(), id, input)
    if err != nil {
        router.Error(w, err)
        return
    }
    
    response.OK(w, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if err := h.service.Delete(r.Context(), id); err != nil {
        router.Error(w, err)
        return
    }
    response.NoContent(w)
}
```

---

## 🔧 Service Completo

```go
package services

import (
    "context"
    apperrors "github.com/fvmoraes/ginger/pkg/errors"
)

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context) ([]User, error) {
    return s.repo.FindAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, apperrors.Internal(err)
    }
    if user == nil {
        return nil, apperrors.NotFound("user not found")
    }
    return user, nil
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    if input.Email == "" {
        return nil, apperrors.BadRequest("email is required")
    }
    
    existing, _ := s.repo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, apperrors.Conflict("email already exists")
    }
    
    user := &User{
        Name:  input.Name,
        Email: input.Email,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, apperrors.Internal(err)
    }
    
    return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
    user, err := s.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    if input.Name != "" {
        user.Name = input.Name
    }
    if input.Email != "" {
        user.Email = input.Email
    }
    
    if err := s.repo.Update(ctx, user); err != nil {
        return nil, apperrors.Internal(err)
    }
    
    return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
    if _, err := s.GetByID(ctx, id); err != nil {
        return err
    }
    
    if err := s.repo.Delete(ctx, id); err != nil {
        return apperrors.Internal(err)
    }
    
    return nil
}
```

---

## 💾 Repository Completo

```go
package repositories

import (
    "context"
    "database/sql"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindAll(ctx context.Context) ([]User, error) {
    query := `SELECT id, name, email, created_at FROM users ORDER BY created_at DESC`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    
    return users, rows.Err()
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
    var u User
    err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &u, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    query := `SELECT id, name, email, created_at FROM users WHERE email = $1`
    var u User
    err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &u, err
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at`
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
    query := `UPDATE users SET name = $1, email = $2 WHERE id = $3`
    _, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
    return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM users WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}
```

---

## 🗄️ Model Completo

```go
package models

import "time"

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateUserInput struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UpdateUserInput struct {
    Name  string `json:"name,omitempty"`
    Email string `json:"email,omitempty"`
}
```

---

## 🔐 Middleware de Autenticação

```go
package middlewares

import (
    "context"
    "net/http"
    "strings"
    "github.com/fvmoraes/ginger/pkg/middleware"
    "github.com/fvmoraes/ginger/pkg/router"
    apperrors "github.com/fvmoraes/ginger/pkg/errors"
)

type contextKey string

const userIDKey contextKey = "user_id"

func RequireAuth() middleware.Func {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                router.Error(w, apperrors.Unauthorized("token required"))
                return
            }
            
            token = strings.TrimPrefix(token, "Bearer ")
            
            // Validar token aqui (JWT, session, etc.)
            userID := validateToken(token)
            if userID == "" {
                router.Error(w, apperrors.Unauthorized("invalid token"))
                return
            }
            
            ctx := context.WithValue(r.Context(), userIDKey, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func UserIDFromContext(ctx context.Context) string {
    if id, ok := ctx.Value(userIDKey).(string); ok {
        return id
    }
    return ""
}

func validateToken(token string) string {
    // TODO: implementar validação real
    return "user-123"
}
```

---

## 📄 Paginação

```go
package handlers

import (
    "net/http"
    "strconv"
    "github.com/fvmoraes/ginger/pkg/response"
)

func (h *UserHandler) ListPaginated(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    
    perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
    if perPage < 1 || perPage > 100 {
        perPage = 20
    }
    
    users, total, err := h.service.ListPaginated(r.Context(), page, perPage)
    if err != nil {
        router.Error(w, err)
        return
    }
    
    response.Paginated(w, users, page, perPage, total)
}
```

```go
// No service
func (s *UserService) ListPaginated(ctx context.Context, page, perPage int) ([]User, int, error) {
    offset := (page - 1) * perPage
    users, err := s.repo.FindPaginated(ctx, offset, perPage)
    if err != nil {
        return nil, 0, apperrors.Internal(err)
    }
    
    total, err := s.repo.Count(ctx)
    if err != nil {
        return nil, 0, apperrors.Internal(err)
    }
    
    return users, total, nil
}
```

```go
// No repository
func (r *UserRepository) FindPaginated(ctx context.Context, offset, limit int) ([]User, error) {
    query := `SELECT id, name, email, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    
    return users, rows.Err()
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
    query := `SELECT COUNT(*) FROM users`
    var count int
    err := r.db.QueryRowContext(ctx, query).Scan(&count)
    return count, err
}
```

---

## 📤 Upload de Arquivo

```go
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
    // Limite de 10MB
    r.ParseMultipartForm(10 << 20)
    
    file, header, err := r.FormFile("file")
    if err != nil {
        router.Error(w, apperrors.BadRequest("file required"))
        return
    }
    defer file.Close()
    
    // Validar tipo
    if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
        router.Error(w, apperrors.BadRequest("only images allowed"))
        return
    }
    
    // Salvar arquivo
    dst, err := os.Create("./uploads/" + header.Filename)
    if err != nil {
        router.Error(w, apperrors.Internal(err))
        return
    }
    defer dst.Close()
    
    if _, err := io.Copy(dst, file); err != nil {
        router.Error(w, apperrors.Internal(err))
        return
    }
    
    response.OK(w, map[string]string{
        "filename": header.Filename,
        "size":     fmt.Sprintf("%d", header.Size),
    })
}
```

---

## 🔄 SSE (Server-Sent Events)

```go
package handlers

import (
    "net/http"
    "time"
    "github.com/fvmoraes/ginger/pkg/sse"
)

func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
    stream, err := sse.New(w)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-r.Context().Done():
            return
        case t := <-ticker.C:
            stream.Send(sse.Event{
                Type: "update",
                Data: map[string]string{
                    "time":    t.Format(time.RFC3339),
                    "message": "Hello from server",
                },
            })
        }
    }
}
```

**Frontend:**
```javascript
const eventSource = new EventSource('/api/v1/stream');

eventSource.addEventListener('update', (e) => {
  const data = JSON.parse(e.data);
  console.log('Update:', data);
});
```

---

## 🔌 WebSocket

```go
package handlers

import (
    "net/http"
    "github.com/fvmoraes/ginger/pkg/ws"
)

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
    ws.Handle(w, r, func(conn *ws.Conn) {
        defer conn.Close()
        
        for {
            var msg ws.Message
            if err := conn.Recv(&msg); err != nil {
                return
            }
            
            // Echo back
            conn.Send(ws.Message{
                Type: "echo",
                Data: msg.Data,
            })
        }
    })
}
```

**Frontend:**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({ type: 'chat', data: { message: 'Hello!' } }));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log('Received:', msg);
};
```

---

## 🐳 Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/foobar ./cmd/foobar

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/foobar .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

ENTRYPOINT ["./foobar"]
```

---

## 🐙 docker-compose.yml

```yaml
version: '3.8'

services:
  foobar:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_DSN=postgres://postgres:postgres@db:5432/foobar?sslmode=disable
      - REDIS_ADDR=redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=foobar
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres-data:
```

---

## ☸️ Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foobar
spec:
  replicas: 3
  selector:
    matchLabels:
      app: foobar
  template:
    metadata:
      labels:
        app: foobar
    spec:
      containers:
      - name: foobar
        image: foobar:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: foobar-secrets
              key: database-dsn
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: foobar
spec:
  selector:
    app: foobar
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

---

## 🧪 Teste Completo

```go
package handlers_test

import (
    "net/http"
    "testing"
    "github.com/fvmoraes/ginger/pkg/router"
    "github.com/fvmoraes/ginger/pkg/testhelper"
)

func TestUserHandler_Create(t *testing.T) {
    handler := NewUserHandler()
    r := router.New()
    handler.Register(r)

    rec := testhelper.NewRequest(t, r, http.MethodPost, "/users/").
        WithBody(map[string]string{
            "name": "Alice",
        }).
        Do()

    testhelper.AssertStatus(t, rec, http.StatusCreated)
}
```

---

<div align="center">
  <p><strong>Copie, cole e customize!</strong></p>
  <p><a href="./README.md">← Voltar ao Índice</a></p>
</div>
