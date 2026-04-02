# Guia de Início Rápido

[← Voltar ao Índice](./README.md)

Comece a usar o Ginger em 5 minutos.

---

## 1. Instalar o Ginger

> Requer **Go 1.25+**.

**Opção recomendada (Go install):**
```bash
go install github.com/fvmoraes/ginger/cmd/ginger@latest
```

**Alternativa (script de instalação):**
```bash
curl -fsSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
```

Verifique a instalação:
```bash
ginger version
```

---

## 2. Criar Seu Primeiro Projeto

```bash
ginger new loja --api
cd loja
go mod tidy
```

Estrutura criada:
```
loja/
├── cmd/loja-api/main.go      # Ponto de entrada
├── internal/api/            # Seu código
├── configs/app.yaml         # Configuração
└── Dockerfile               # Deploy
```

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
ginger generate crud produto
```

Isso cria:
- `internal/models/produto.go` — Modelo
- `internal/api/handlers/produto_handler.go` — HTTP
- `internal/api/services/produto_service.go` — Lógica
- `internal/api/repositories/produto_repository.go` — Dados
- `internal/api/handlers/produto_handler_test.go` — Testes

### Registrar no Router

Edite `cmd/loja-api/main.go`:

```go
package main

import (
    gingerapp "github.com/fvmoraes/ginger/pkg/app"
    "github.com/fvmoraes/ginger/pkg/config"
    "github.com/fvmoraes/ginger/pkg/middleware"
    
    "loja/internal/api/handlers"
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
    
    produtoHandler := handlers.NewProdutoHandler(nil) // TODO: injetar service
    v1.GET("/produtos", produtoHandler.List)
    v1.POST("/produtos", produtoHandler.Create)
    v1.GET("/produtos/{id}", produtoHandler.Get)
    v1.PUT("/produtos/{id}", produtoHandler.Update)
    v1.DELETE("/produtos/{id}", produtoHandler.Delete)
    
    app.Run()
}
```

### Testar

```bash
# Criar produto
curl -X POST http://localhost:8080/api/v1/produtos \
  -H "Content-Type: application/json" \
  -d '{"nome":"Notebook","preco":2500}'

# Listar produtos
curl http://localhost:8080/api/v1/produtos
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
  dsn: postgres://<user>:<password>@localhost:5432/foobar-api?sslmode=disable
  max_open: 25
  max_idle: 5
```

### Conectar

Edite `cmd/loja-api/main.go`:

```go
import (
    "loja/platform/database"
)

func main() {
    cfg, _ := config.Load("configs/app.yaml")
    
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
    
    // Injetar dependências
    produtoRepo := repositories.NewProdutoRepository(db)
    produtoService := services.NewProdutoService(produtoRepo)
    produtoHandler := handlers.NewProdutoHandler(produtoService)
    
    // Rotas...
    
    app.Run()
}
```

---

## 6. Executar com Docker

```bash
# Build
docker build -t loja .

# Run
docker run -p 8080:8080 \
  -e DATABASE_DSN="postgres://<user>:<password>@host/db" \
  loja
```

Ou use Docker Compose:

```bash
docker compose up -d
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
kubectl apply -f kubernetes/

# Helm
helm install loja ./helm
```

---

## Comandos Úteis

```bash
# Criar projeto
ginger new <nome>            # genérico -> cmd/<nome>
ginger new <nome> --api         # API      -> cmd/<nome>-api
ginger new <nome> --service         # service  -> cmd/<nome>-service
ginger new <nome> --worker         # worker   -> cmd/<nome>-worker
ginger new <nome> --cli         # CLI      -> cmd/<nome>-cli

# Gerar código
ginger generate crud <recurso>
ginger generate handler <nome>
ginger generate service <nome>

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
type ProdutoHandler struct {
    service ProdutoService
}

func NewProdutoHandler(service ProdutoService) *ProdutoHandler {
    return &ProdutoHandler{service: service}
}

func (h *ProdutoHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Parse input
    var input CreateProdutoInput
    if err := router.Decode(r, &input); err != nil {
        router.Error(w, err)
        return
    }
    
    // 2. Call service
    produto, err := h.service.Create(r.Context(), input)
    if err != nil {
        router.Error(w, err)
        return
    }
    
    // 3. Return response
    response.Created(w, produto)
}
```

---

## Estrutura Básica de um Service

```go
type ProdutoService struct {
    repo ProdutoRepository
}

func NewProdutoService(repo ProdutoRepository) *ProdutoService {
    return &ProdutoService{repo: repo}
}

func (s *ProdutoService) Create(ctx context.Context, input CreateProdutoInput) (*Produto, error) {
    // 1. Validar
    if input.Nome == "" {
        return nil, apperrors.BadRequest("nome é obrigatório")
    }
    
    // 2. Criar
    produto := &Produto{
        Nome:  input.Nome,
        Preco: input.Preco,
    }
    
    // 3. Persistir
    if err := s.repo.Create(ctx, produto); err != nil {
        return nil, apperrors.Internal(err)
    }
    
    return produto, nil
}
```

---

## Estrutura Básica de um Repository

```go
type ProdutoRepository struct {
    db *sql.DB
}

func NewProdutoRepository(db *sql.DB) *ProdutoRepository {
    return &ProdutoRepository{db: db}
}

func (r *ProdutoRepository) Create(ctx context.Context, produto *Produto) error {
    query := `INSERT INTO produtos (nome, preco) VALUES ($1, $2) RETURNING id`
    return r.db.QueryRowContext(ctx, query, produto.Nome, produto.Preco).Scan(&produto.ID)
}

func (r *ProdutoRepository) FindByID(ctx context.Context, id int) (*Produto, error) {
    query := `SELECT id, nome, preco FROM produtos WHERE id = $1`
    var p Produto
    err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Nome, &p.Preco)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &p, err
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
