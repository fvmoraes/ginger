# Arquitetura do Ginger Framework

[← Voltar ao README](../README.md)

## Índice

- [Visão Geral](#visão-geral)
- [Filosofia de Design](#filosofia-de-design)
- [Estrutura de Diretórios](#estrutura-de-diretórios)
- [Fluxo de Requisição](#fluxo-de-requisição)
- [Camadas da Aplicação](#camadas-da-aplicação)
- [Padrões de Código](#padrões-de-código)

---

## Visão Geral

O Ginger é construído sobre três pilares fundamentais:

1. **Stdlib-first** — Usa `net/http`, `log/slog`, `database/sql` como base
2. **Opinativo mas flexível** — Estrutura clara, mas você pode substituir qualquer parte
3. **Zero mágica** — Sem reflection pesada, sem DI automática, sem code generation em runtime

### Diagrama de Componentes

```
┌─────────────────────────────────────────────────────────────┐
│                      Ginger Application                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Router     │───▶│  Middleware  │───▶│   Handler    │  │
│  │  (pkg/router)│    │(pkg/middleware)   │ (internal/api)│  │
│  └──────────────┘    └──────────────┘    └──────┬───────┘  │
│                                                   │           │
│                                                   ▼           │
│                                          ┌──────────────┐    │
│                                          │   Service    │    │
│                                          │(internal/api)│    │
│                                          └──────┬───────┘    │
│                                                 │             │
│                                                 ▼             │
│                                        ┌──────────────┐      │
│                                        │  Repository  │      │
│                                        │(internal/api)│      │
│                                        └──────┬───────┘      │
│                                               │               │
│                                               ▼               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Database   │    │    Cache     │    │  Messaging   │  │
│  │  (platform/) │    │  (platform/) │    │  (platform/) │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Filosofia de Design

### 1. Separação de Responsabilidades

Cada camada tem uma responsabilidade clara:

- **Handler** — HTTP I/O (parse request, write response)
- **Service** — Lógica de negócio (validação, orquestração)
- **Repository** — Acesso a dados (queries, transactions)
- **Model** — Estruturas de domínio (sem lógica)

### 2. Dependency Injection Manual

Não usamos frameworks de DI. Construtores explícitos:

```go
// Ruim (mágica)
@Inject
var userService UserService

// Bom (explícito)
userRepo := repositories.NewUserRepository(db)
userService := services.NewUserService(userRepo)
userHandler := handlers.NewUserHandler(userService)
```

### 3. Interfaces no Consumidor

Interfaces são definidas onde são usadas, não onde são implementadas:

```go
// internal/api/services/user_service.go
type UserRepository interface {
    FindByID(ctx context.Context, id int) (*models.User, error)
    Create(ctx context.Context, user *models.User) error
}

type UserService struct {
    repo UserRepository  // interface, não struct concreta
}
```

### 4. Erros Tipados

Todos os erros de domínio são `*apperrors.AppError`:

```go
if user == nil {
    return apperrors.NotFound("user not found")
}
if !user.Active {
    return apperrors.Forbidden("user is inactive")
}
```

O router converte automaticamente para JSON:

```json
{
  "code": "NOT_FOUND",
  "message": "user not found"
}
```

---

## Estrutura de Diretórios

### Layout Completo

```
foobar/
├── cmd/
│   └── foobar/
│       └── main.go              # Entrypoint — wiring de dependências
│
├── internal/                    # Código privado da aplicação
│   ├── api/
│   │   ├── handlers/            # HTTP handlers (thin layer)
│   │   │   ├── user_handler.go
│   │   │   └── health.go
│   │   ├── services/            # Lógica de negócio
│   │   │   └── user_service.go
│   │   ├── repositories/        # Data access
│   │   │   └── user_repository.go
│   │   └── middlewares/         # App-specific middlewares
│   │       └── auth.go
│   ├── models/                  # Domain models
│   │   └── user.go
│   └── config/                  # Config loader wrapper
│       └── config.go
│
├── pkg/                         # Código reutilizável interno
│   └── validator/               # Exemplo: validação customizada
│       └── validator.go
│
├── platform/                    # Integrações externas
│   ├── database/
│   │   └── postgres.go          # ginger add postgres
│   ├── cache/
│   │   └── redis.go             # ginger add redis
│   └── messaging/
│       └── kafka.go             # ginger add kafka
│
├── configs/
│   └── app.yaml                 # Configuração principal
│
├── scripts/                     # Scripts de dev/CI
│   ├── migrate.sh
│   └── seed.sh
│
├── tests/                       # Testes de integração
│   └── api_test.go
│
├── docs/                        # Documentação
│   ├── API.md
│   └── DEPLOYMENT.md
│
├── Dockerfile                   # Multi-stage build
├── docker-compose.yml           # Dev environment
├── Makefile                     # Comandos comuns
├── .env.example                 # Template de env vars
├── go.mod
└── README.md
```

### Convenções de Nomenclatura

| Tipo | Padrão | Exemplo |
|------|--------|---------|
| Handler | `<resource>_handler.go` | `user_handler.go` |
| Service | `<resource>_service.go` | `user_service.go` |
| Repository | `<resource>_repository.go` | `user_repository.go` |
| Model | `<resource>.go` | `user.go` |
| Test | `<file>_test.go` | `user_handler_test.go` |
| Interface | `<Noun>er` ou `<Noun>Repository` | `UserRepository` |
| Struct | `<Noun>` | `UserService` |
| Constructor | `New<Type>` | `NewUserService` |

---

## Fluxo de Requisição

### Ciclo de Vida Completo

```
1. HTTP Request
   │
   ▼
2. Router (pkg/router)
   │ - Match route pattern
   │ - Extract path params
   │
   ▼
3. Middleware Chain (pkg/middleware)
   │ - Logger       → log request
   │ - RequestID    → inject X-Request-ID
   │ - Recover      → catch panics
   │ - CORS         → add CORS headers
   │ - Auth         → validate token (app-specific)
   │
   ▼
4. Handler (internal/api/handlers)
   │ - Parse request body
   │ - Validate input
   │ - Call service
   │ - Write response
   │
   ▼
5. Service (internal/api/services)
   │ - Business logic
   │ - Validation rules
   │ - Orchestrate repositories
   │ - Return domain errors
   │
   ▼
6. Repository (internal/api/repositories)
   │ - SQL queries
   │ - Transaction management
   │ - Map rows → models
   │
   ▼
7. Database (platform/)
   │ - Execute query
   │ - Return results
   │
   ▼
8. Response
   │ - Service returns model
   │ - Handler serializes to JSON
   │ - Middleware logs response
   │
   ▼
9. HTTP Response
```

### Exemplo Concreto: POST /api/v1/users

```go
// 1. Router registra a rota
v1.POST("/users", userHandler.Create)

// 2. Middleware chain executa
middleware.Chain(
    middleware.Logger(log),
    middleware.RequestID(),
    middleware.Recover(log),
    middleware.CORS(),
)

// 3. Handler recebe a requisição
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

// 4. Service executa lógica de negócio
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

// 5. Repository persiste no banco
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
}
```

---

## Camadas da Aplicação

### Handler Layer

**Responsabilidade:** HTTP I/O apenas

```go
type UserHandler struct {
    service UserService
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Parse input
    var input CreateUserInput
    if err := router.Decode(r, &input); err != nil {
        router.Error(w, err)
        return
    }
    
    // 2. Call service
    user, err := h.service.Create(r.Context(), input)
    if err != nil {
        router.Error(w, err)
        return
    }
    
    // 3. Write response
    response.Created(w, user)
}
```

**Regras:**
- ✅ Parse request, write response
- ✅ Chamar service
- ❌ Lógica de negócio
- ❌ Acesso direto ao banco

### Service Layer

**Responsabilidade:** Lógica de negócio

```go
type UserService struct {
    repo UserRepository
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    // 1. Validação de negócio
    if input.Email == "" {
        return nil, apperrors.BadRequest("email is required")
    }
    
    // 2. Regras de domínio
    existing, _ := s.repo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, apperrors.Conflict("email already exists")
    }
    
    // 3. Orquestração
    user := &User{Name: input.Name, Email: input.Email}
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, apperrors.Internal(err)
    }
    
    // 4. Retornar resultado
    return user, nil
}
```

**Regras:**
- ✅ Validação de negócio
- ✅ Orquestração de repositories
- ✅ Retornar erros tipados
- ❌ HTTP awareness
- ❌ SQL direto

### Repository Layer

**Responsabilidade:** Acesso a dados

```go
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*User, error) {
    query := `SELECT id, name, email FROM users WHERE id = $1`
    var user User
    err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
}
```

**Regras:**
- ✅ SQL queries
- ✅ Transaction management
- ✅ Map rows → structs
- ❌ Lógica de negócio
- ❌ HTTP awareness

---

## Padrões de Código

### 1. Constructor Pattern

Sempre use construtores explícitos:

```go
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### 2. Interface Segregation

Interfaces pequenas e focadas:

```go
// Ruim — interface grande
type UserRepository interface {
    FindByID(ctx context.Context, id int) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindAll(ctx context.Context) ([]*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int) error
    Count(ctx context.Context) (int, error)
}

// Bom — interfaces focadas
type UserFinder interface {
    FindByID(ctx context.Context, id int) (*User, error)
}

type UserCreator interface {
    Create(ctx context.Context, user *User) error
}

// Service usa apenas o que precisa
type UserService struct {
    finder  UserFinder
    creator UserCreator
}
```

### 3. Error Wrapping

Use `fmt.Errorf` com `%w` para preservar a cadeia:

```go
if err := repo.Create(ctx, user); err != nil {
    return fmt.Errorf("create user: %w", err)
}
```

### 4. Context Propagation

Sempre passe `context.Context` como primeiro parâmetro:

```go
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    // ...
}
```

### 5. Table-Driven Tests

Use tabelas para testes parametrizados:

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        wantErr bool
    }{
        {"valid user", CreateUserInput{Name: "Alice", Email: "alice@example.com"}, false},
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

## Próximos Passos

- [📦 Guia de Pacotes](./PACKAGES.md) — Documentação detalhada de cada pacote
- [🔌 Integrações](./INTEGRATIONS.md) — Como adicionar bancos, cache, mensageria
- [🧪 Testes](./TESTING.md) — Estratégias de teste e mocks
- [🚀 Deploy](./DEPLOYMENT.md) — Docker, Kubernetes, CI/CD

[← Voltar ao README](../README.md)
