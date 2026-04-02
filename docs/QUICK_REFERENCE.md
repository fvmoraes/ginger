# Referência Rápida Ginger

[← Voltar ao Índice](./README.md)

Guia de consulta rápida com os comandos e padrões mais usados.

---

## 🚀 Comandos CLI

```bash
# Criar novo projeto
ginger new foobar            # genérico  -> cmd/foobar
ginger new foobar --api         # api       -> cmd/foobar-api
ginger new foobar --service         # service   -> cmd/foobar-service
ginger new foobar --worker         # worker    -> cmd/foobar-worker
ginger new foobar --cli         # cli       -> cmd/foobar-cli

# Executar
ginger run

# Build
ginger build
ginger build ./bin/foobar

# Gerar código
ginger generate handler foobar
ginger generate service foobar
ginger generate repository foobar
ginger generate crud foobar
ginger generate test foobar
ginger generate test app
ginger generate swagger
ginger generate swagger foobar

# Adicionar integrações
ginger add postgres
ginger add mongodb
ginger add redis
ginger add kafka
ginger add grpc
ginger add swagger

# Diagnosticar
ginger doctor

# Ajuda
ginger help
ginger version
# output: ginger x.y.z
```

---

## 📦 Imports Comuns

```go
import (
    // Core
    "github.com/fvmoraes/ginger/pkg/app"
    "github.com/fvmoraes/ginger/pkg/router"
    "github.com/fvmoraes/ginger/pkg/middleware"
    "github.com/fvmoraes/ginger/pkg/config"
    "github.com/fvmoraes/ginger/pkg/logger"
    
    // Errors & Response
    apperrors "github.com/fvmoraes/ginger/pkg/errors"
    "github.com/fvmoraes/ginger/pkg/response"
    
    // Real-time
    "github.com/fvmoraes/ginger/pkg/sse"
    "github.com/fvmoraes/ginger/pkg/ws"
    
    // Infra
    "github.com/fvmoraes/ginger/pkg/database"
    "github.com/fvmoraes/ginger/pkg/health"
    "github.com/fvmoraes/ginger/pkg/telemetry"
)
```

---

## 🏗️ Estrutura Básica

### main.go

```go
package main

import (
    "context"
    gingerapp "github.com/fvmoraes/ginger/pkg/app"
    "github.com/fvmoraes/ginger/pkg/config"
    "github.com/fvmoraes/ginger/pkg/middleware"
)

func main() {
    cfg, _ := config.Load("configs/app.yaml")
    app := gingerapp.New(cfg)
    
    // Middlewares
    app.Router.Use(
        middleware.Logger(app.Logger),
        middleware.RequestID(),
        middleware.Recover(app.Logger),
        middleware.CORS(),
    )
    
    // Rotas
    v1 := app.Router.Group("/api/v1")
    v1.GET("/users", listUsers)
    v1.POST("/users", createUser)
    
    // Cleanup
    app.OnStop(func(ctx context.Context) error {
        return db.Close()
    })
    
    app.Run()
}
```

### Handler

```go
type UserHandler struct {
    // svc UserService
}

func NewUserHandler() *UserHandler {
    return &UserHandler{}
}

func (h *UserHandler) Register(r *router.Router) {
    g := r.Group("/users")
    g.GET("/", h.list)
    g.POST("/", h.create)
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
    var body map[string]any
    if err := router.Decode(r, &body); err != nil {
        router.Error(w, err)
        return
    }

    router.JSON(w, http.StatusCreated, body)
}
```

### Service

```go
type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    if input.Email == "" {
        return nil, apperrors.BadRequest("email is required")
    }
    
    existing, _ := s.repo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, apperrors.Conflict("email already exists")
    }
    
    user := &User{Name: input.Name, Email: input.Email}
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, apperrors.Internal(err)
    }
    
    return user, nil
}
```

### Repository

```go
type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    query := `SELECT id, name, email FROM users WHERE email = $1`
    var user User
    err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &user, err
}
```

---

## 🔧 Configuração

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

database:
  driver: postgres
  dsn: postgres://<user>:<password>@localhost:5432/foobar?sslmode=disable
  max_open: 25
  max_idle: 5

log:
  level: info
  format: json
```

### Variáveis de Ambiente

```bash
# App
APP_NAME=foobar
APP_ENV=production
APP_VERSION=1.0.0

# HTTP
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Database
DATABASE_DRIVER=postgres
DATABASE_DSN=postgres://<user>:<password>@host:5432/db

# Log
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 🎯 Padrões Comuns

### Erro Handling

```go
// Service layer
if user == nil {
    return apperrors.NotFound("user not found")
}
if !user.Active {
    return apperrors.Forbidden("user is inactive")
}
if err := repo.Create(ctx, user); err != nil {
    return apperrors.Internal(err)
}

// Handler layer
if err != nil {
    router.Error(w, err)  // converte automaticamente
    return
}
```

### Response Patterns

```go
// Single resource
response.OK(w, user)
response.Created(w, user)

// List
response.Paginated(w, users, page, perPage, total)

// No content
response.NoContent(w)

// Custom JSON
router.JSON(w, http.StatusOK, map[string]string{"message": "success"})
```

### Middleware Chain

```go
app.Router.Use(middleware.Chain(
    middleware.Logger(log),
    middleware.RequestID(),
    middleware.Recover(log),
    middleware.CORS(),
))
```

### Route Groups

```go
v1 := app.Router.Group("/api/v1")
v1.GET("/users", listUsers)
v1.POST("/users", createUser)

admin := app.Router.Group("/admin", middleware.RequireAuth())
admin.GET("/stats", getStats)
```

---

## 🧪 Testes

### Handler Test

```go
func TestUserHandler_Create(t *testing.T) {
    handler := NewUserHandler()
    r := router.New()
    handler.Register(r)

    rec := testhelper.NewRequest(t, r, http.MethodPost, "/users/").
        WithBody(map[string]string{"name": "Alice"}).
        Do()

    testhelper.AssertStatus(t, rec, http.StatusCreated)
}
```

### Service Test

```go
func TestUserService_Create(t *testing.T) {
    mockRepo := &mockUserRepository{}
    service := NewUserService(mockRepo)
    
    user, err := service.Create(context.Background(), CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
}
```

### Table-Driven Test

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        wantErr bool
    }{
        {"valid", CreateUserInput{Name: "Alice", Email: "alice@example.com"}, false},
        {"missing email", CreateUserInput{Name: "Bob"}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

---

## 🐳 Docker

### Build

```bash
docker build -f devops/docker/Dockerfile -t foobar:latest .
```

### Run

```bash
docker run -p 8080:8080 \
  -e DATABASE_DSN="postgres://<user>:<password>@host/db" \
  -e LOG_LEVEL="info" \
  foobar:latest
```

### Docker Compose

```bash
docker compose -f devops/docker/docker-compose.yml up -d
docker compose logs -f foobar
docker compose down
```

---

## ☸️ Kubernetes

### Deploy

```bash
kubectl apply -f devops/kubernetes/
kubectl get pods -l app=foobar
kubectl logs -f deployment/foobar
```

### Scale

```bash
kubectl scale deployment foobar --replicas=5
```

### Rollout

```bash
kubectl set image deployment/foobar foobar=foobar:v2
kubectl rollout status deployment/foobar
kubectl rollout undo deployment/foobar
```

### Port Forward

```bash
kubectl port-forward svc/foobar 8080:80
```

---

## 📊 Observabilidade

### Health Check

```go
h := health.New()
h.Register(database.NewChecker(db))
h.Register(cache.NewChecker(redisClient))
app.Router.HandleRaw("/health", h)
```

### Prometheus

```go
m := metrics.NewDefaultMetrics("myapi")
app.Router.Use(metrics.Middleware(m))
app.Router.HandleRaw("/metrics", metrics.Handler())
```

### OpenTelemetry

```go
shutdown, _ := telemetry.Setup(ctx, "foobar", "1.0.0")
app.OnStop(shutdown)

tracer := telemetry.Tracer("foobar")
ctx, span := tracer.Start(ctx, "operation")
defer span.End()
```

---

## 🔌 Integrações Rápidas

### PostgreSQL

```bash
ginger add postgres
```

```go
db, _ := database.Connect(database.Config{
    DSN: "postgres://<user>:<password>@localhost:5432/foobar",
})
```

### Redis

```bash
ginger add redis
```

```go
client, _ := cache.Connect(cache.Config{
    Addr: "localhost:6379",
})
```

### Kafka

```bash
ginger add kafka
```

```go
writer := messaging.NewWriter(messaging.ProducerConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "events",
})
```

---

## 🎨 Real-time

### SSE

```go
func streamHandler(w http.ResponseWriter, r *http.Request) {
    stream, _ := sse.New(w)
    for {
        select {
        case <-r.Context().Done():
            return
        case event := <-eventCh:
            stream.Send(sse.Event{Type: "update", Data: event})
        }
    }
}
```

### WebSocket

```go
func chatHandler(w http.ResponseWriter, r *http.Request) {
    ws.Handle(w, r, func(conn *ws.Conn) {
        for {
            var msg ws.Message
            if err := conn.Recv(&msg); err != nil {
                return
            }
            conn.Send(ws.Message{Type: "echo", Data: msg.Data})
        }
    })
}
```

---

## 📝 Makefile Útil

```makefile
BIN=bin/foobar
CMD_DIR=cmd/foobar-api

.PHONY: run build test lint docker-build docker-run

run:
	go run ./$(CMD_DIR)

build:
	go build -o $(BIN) ./$(CMD_DIR)

test:
	go test -v -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

docker-build:
	docker build -f devops/docker/Dockerfile -t foobar:latest .

docker-run:
	docker run -p 8080:8080 foobar:latest

k8s-deploy:
	kubectl apply -f devops/kubernetes/

k8s-logs:
	kubectl logs -f deployment/foobar
```

---

## 🔍 Troubleshooting Rápido

### Logs

```bash
# Docker
docker logs -f foobar

# Kubernetes
kubectl logs -f deployment/foobar
kubectl logs -f deployment/foobar --previous
```

### Debug

```bash
# Exec into container
kubectl exec -it deployment/foobar -- sh

# Port forward
kubectl port-forward svc/foobar 8080:80
```

### Health Check

```bash
curl http://localhost:8080/health
```

---

<div align="center">
  <p><a href="./README.md">← Voltar ao Índice</a> | <a href="../README.md">README Principal</a></p>
</div>
