# Guia de Início Rápido

[← Voltar ao Índice](./README.md)

Comece a usar o Ginger em 5 minutos.

---

## 1. Instalar o Ginger

> Requer **Go 1.25+**.

**Opção recomendada (Go install):**
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@latest

# Se `ginger` não for encontrado, exporte o bin do Go no PATH
export PATH="$(go env GOPATH)/bin:$PATH"
```

**Alternativa (script de instalação):**
```bash
# instala a latest release por padrão
curl -fsSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

Verifique a instalação:
```bash
ginger version
# output: ginger x.y.z
```

---

## 2. Criar Seu Primeiro Projeto

```bash
ginger new foobar --service
# ou use a flag curta equivalente: -s
cd foobar
go mod tidy
```

Estrutura criada:
```
foobar/
├── cmd/foobar/main.go          # Ponto de entrada
├── internal/api/handlers/health.go
├── configs/app.yaml         # Configuração
└── devops/                  # Build, deploy e CI/CD
```

O restante da estrutura aparece conforme você usa `ginger generate` e `ginger add`.

---

## 3. Executar

```bash
ginger run
```

Acesse: http://localhost:8080/health

---

## 4. Criar Seu Primeiro Endpoint

### Gerar CRUD completo

```bash
ginger generate crud foobar
```

Isso cria:
- `internal/models/foobar.go` — Modelo
- `internal/api/handlers/foobar_handler.go` — HTTP
- `internal/services/foobar_service.go` — Lógica
- `internal/ports/foobar_repository.go` — Contrato
- `internal/adapters/foobar_memory_repository.go` — Adapter in-memory

Se quiser os testes depois, gere separadamente:

```bash
ginger generate test foobar
ginger generate smoke-test
```

### Registrar no Router

Edite `cmd/foobar/main.go`:

```go
package main

import (
    "log"

    "foobar/internal/api/handlers"
    "foobar/internal/config"
    gingerapp "github.com/fvmoraes/ginger/pkg/app"
)

func main() {
    cfg, _ := config.Load()
    app := gingerapp.New(cfg)

    v1 := app.Router.Group("/api/v1")
    foobarHandler := handlers.NewFoobarHandler()
    foobarHandler.Register(v1)

    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Testar

```bash
# Criar foobar
curl -X POST http://localhost:8080/api/v1/foobars \
  -H "Content-Type: application/json" \
  -d '{"name":"foobar"}'

# Listar foobars
curl http://localhost:8080/api/v1/foobars
```

---

## 5. Adicionar Banco de Dados

```bash
ginger add postgres
```

Isso cria `platform/database/postgres.go`.

### Configurar

Edite `configs/app.yaml`:

```yaml
database:
  driver: postgres
  dsn: postgres://<user>:<password>@localhost:5432/foobar?sslmode=disable
  max_open: 25
  max_idle: 5
```

### Conectar

Edite `cmd/foobar/main.go`:

```go
import (
    "context"
    "log"

    "foobar/internal/api/handlers"
    "foobar/internal/config"
    "foobar/platform/database"
    gingerapp "github.com/fvmoraes/ginger/pkg/app"
)

func main() {
    cfg, _ := config.Load()
    
    // Conectar banco
    db, err := database.Connect(database.Config{
        DSN:     cfg.Database.DSN,
        MaxOpen: cfg.Database.MaxOpen,
        MaxIdle: cfg.Database.MaxIdle,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    app := gingerapp.New(cfg)
    
    // Cleanup
    app.OnStop(func(ctx context.Context) error {
        return db.Close()
    })

    foobarHandler := handlers.NewFoobarHandler()
    foobarHandler.Register(app.Router.Group("/api/v1"))

    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

---

## 6. Executar com Docker

```bash
# Build
docker build -f devops/docker/Dockerfile -t foobar .

# Run
docker run -p 8080:8080 \
  -e DATABASE_DSN="postgres://<user>:<password>@host/db" \
  foobar
```

Ou use Docker Compose:

```bash
docker compose -f devops/docker/docker-compose.yml up -d
```

---

## Próximos Passos

### Aprender Mais

- [📦 Pacotes](./PACKAGES.md) — API completa de cada pacote
- [🏗️ Arquitetura](./ARCHITECTURE.md) — Como o Ginger funciona
- [🔌 Integrações](./INTEGRATIONS.md) — Redis, Kafka, gRPC, etc.

### Adicionar Funcionalidades

```bash
# Cache
ginger add redis

# Mensageria
ginger add kafka

# gRPC
ginger add grpc

# Métricas
ginger add prometheus
```

### Testar

```bash
go test ./...
```

### Deploy

```bash
# Kubernetes
kubectl apply -f devops/kubernetes/

# Helm
helm install foobar ./devops/helm
```

---

## Comandos Úteis

```bash
# Criar projeto
ginger new <nome>            # genérico -> cmd/<nome>
ginger new <nome> --service | -s   # service  -> cmd/<nome>
ginger new <nome> --worker | -w    # worker   -> cmd/<nome>-worker
ginger new <nome> --cli | -c       # cli      -> cmd/<nome>
ginger new <nome> --cli | -c       # CLI      -> cmd/<nome>-cli

# Gerar código
ginger generate crud <recurso>
ginger generate test <recurso>
ginger generate swagger [recurso]

# Adicionar integração
ginger add <integração>

# Executar
ginger run

# Build
ginger build

# Diagnosticar
ginger doctor
```

---

## Estrutura Básica de um Handler

```go
type FoobarHandler struct {
    // svc FoobarService
}

func NewFoobarHandler() *FoobarHandler {
    return &FoobarHandler{}
}

func (h *FoobarHandler) Register(r *router.Router) {
    g := r.Group("/foobars")
    g.GET("/", h.list)
    g.POST("/", h.create)
}

func (h *FoobarHandler) create(w http.ResponseWriter, r *http.Request) {
    var body map[string]any
    if err := router.Decode(r, &body); err != nil {
        router.Error(w, err)
        return
    }
    router.JSON(w, http.StatusCreated, body)
}
```

---

## Estrutura Básica de um Service

```go
type FoobarService struct {
    repo FoobarRepository
}

func NewFoobarService(repo FoobarRepository) *FoobarService {
    return &FoobarService{repo: repo}
}

func (s *FoobarService) Create(ctx context.Context, input CreateFoobarInput) (*Foobar, error) {
    // 1. Validar
    if input.Name == "" {
        return nil, apperrors.BadRequest("name is required")
    }
    
    // 2. Criar
    foobar := &Foobar{
        Name: input.Name,
    }
    
    // 3. Persistir
    if err := s.repo.Create(ctx, foobar); err != nil {
        return nil, apperrors.Internal(err)
    }
    
    return foobar, nil
}
```

---

## Estrutura Básica de um Repository

```go
type FoobarRepository struct {
    db *sql.DB
}

func NewFoobarRepository(db *sql.DB) *FoobarRepository {
    return &FoobarRepository{db: db}
}

func (r *FoobarRepository) Create(ctx context.Context, foobar *Foobar) error {
    query := `INSERT INTO foobars (name) VALUES ($1) RETURNING id`
    return r.db.QueryRowContext(ctx, query, foobar.Name).Scan(&foobar.ID)
}

func (r *FoobarRepository) FindByID(ctx context.Context, id int) (*Foobar, error) {
    query := `SELECT id, name FROM foobars WHERE id = $1`
    var f Foobar
    err := r.db.QueryRowContext(ctx, query, id).Scan(&f.ID, &f.Name)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &f, err
}
```

---

## Dúvidas Comuns

### Como adicionar autenticação?

Crie um middleware customizado:

```go
func RequireAuth() middleware.Func {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                router.Error(w, apperrors.Unauthorized("token required"))
                return
            }
            // Validar token...
            next.ServeHTTP(w, r)
        })
    }
}
```

### Como fazer paginação?

Use `response.Paginated`:

```go
func (h *ProdutoHandler) List(w http.ResponseWriter, r *http.Request) {
    page := 1    // parse de query params
    perPage := 20
    
    produtos, total, err := h.service.List(r.Context(), page, perPage)
    if err != nil {
        router.Error(w, err)
        return
    }
    
    response.Paginated(w, produtos, page, perPage, total)
}
```

### Como fazer upload de arquivos?

```go
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
    file, header, err := r.FormFile("file")
    if err != nil {
        router.Error(w, apperrors.BadRequest("file required"))
        return
    }
    defer file.Close()
    
    // Processar arquivo...
    
    response.OK(w, map[string]string{"filename": header.Filename})
}
```

---

<div align="center">
  <p><strong>Pronto para começar!</strong></p>
  <p><a href="./README.md">← Voltar ao Índice</a> | <a href="./PACKAGES.md">Ver Pacotes →</a></p>
</div>
