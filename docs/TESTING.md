# Guia de Testes

[← Voltar ao README](../README.md)

## Índice

- [Filosofia de Testes](#filosofia-de-testes)
- [Estrutura de Testes](#estrutura-de-testes)
- [Testes Unitários](#testes-unitários)
- [Testes de Integração](#testes-de-integração)
- [Mocks e Stubs](#mocks-e-stubs)
- [Test Helpers](#test-helpers)
- [Coverage](#coverage)
- [CI/CD](#cicd)

---

## Filosofia de Testes

### Pirâmide de Testes

```
        ┌─────────┐
        │   E2E   │  ← Poucos, lentos, frágeis
        ├─────────┤
        │ Integr. │  ← Médio número, médio tempo
        ├─────────┤
        │ Unitár. │  ← Muitos, rápidos, isolados
        └─────────┘
```

### Princípios

1. **Testes rápidos** — Suite completa < 10s
2. **Testes isolados** — Sem dependências externas (use mocks)
3. **Testes determinísticos** — Mesmo input = mesmo output
4. **Testes legíveis** — Nome do teste descreve o comportamento
5. **Testes mantíveis** — Refatore testes junto com código

---

## Estrutura de Testes

### Geração de Testes com o CLI

```bash
# Gera testes unitários para handler, service e adapter in-memory
ginger generate test foobar

# Gera só o smoke test da aplicação
ginger generate smoke-test
```

### Convenções de Nomenclatura

```
internal/api/handlers/
├── user_handler.go
└── user_handler_test.go          ← Testes unitários

internal/services/
├── user_service.go
└── user_service_test.go

internal/adapters/
├── user_memory_repository.go
└── user_memory_repository_test.go

tests/integration/
├── user_test.go                  ← Fluxo CRUD do recurso
└── app_smoke_test.go             ← Smoke test da aplicação
```

### Padrão de Nome de Teste

```go
func Test<Type>_<Method>_<Scenario>(t *testing.T)

// Exemplos
func TestUserService_Create_ValidInput(t *testing.T)
func TestUserService_Create_DuplicateEmail(t *testing.T)
func TestUserHandler_Get_NotFound(t *testing.T)
```

---

## Testes Unitários

### Handler Tests

```go
package handlers_test

import (
    "net/http"
    "testing"
    
    "github.com/fvmoraes/ginger/pkg/testhelper"
    "yourmodule/internal/api/handlers"
    "github.com/fvmoraes/ginger/pkg/router"
)

func TestUserHandler_Get_Success(t *testing.T) {
    // Arrange
    handler := handlers.NewUserHandler()
    r := router.New()
    handler.Register(r)
    
    // Act
    rec := testhelper.NewRequest(t, r, http.MethodGet, "/users/1").Do()
    
    // Assert
    testhelper.AssertStatus(t, rec, http.StatusOK)
}

func TestUserHandler_Get_NotFound(t *testing.T) {
    handler := handlers.NewUserHandler()
    r := router.New()
    handler.Register(r)

    rec := testhelper.NewRequest(t, r, http.MethodGet, "/users/999").Do()

    testhelper.AssertStatus(t, rec, http.StatusOK)
}
```

### Service Tests

```go
package services_test

import (
    "context"
    "testing"
    
    "yourmodule/internal/services"
    "yourmodule/internal/models"
)

func TestUserService_Create_ValidInput(t *testing.T) {
    // Arrange
    mockRepo := &mockUserRepository{}
    service := services.NewUserService(mockRepo)
    
    input := services.CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    }
    
    // Act
    user, err := service.Create(context.Background(), input)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
    if !mockRepo.createCalled {
        t.Error("expected Create to be called")
    }
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
    mockRepo := &mockUserRepository{
        findByEmailResult: &models.User{ID: 1, Email: "alice@example.com"},
    }
    service := services.NewUserService(mockRepo)
    
    input := services.CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    }
    
    _, err := service.Create(context.Background(), input)
    
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    
    var appErr *apperrors.AppError
    if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeConflict {
        t.Errorf("expected Conflict error, got %v", err)
    }
}
```

### Table-Driven Tests

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   services.CreateUserInput
        setup   func(*mockUserRepository)
        wantErr bool
        errCode apperrors.Code
    }{
        {
            name:  "valid input",
            input: services.CreateUserInput{Name: "Alice", Email: "alice@example.com"},
            setup: func(m *mockUserRepository) {},
            wantErr: false,
        },
        {
            name:  "missing email",
            input: services.CreateUserInput{Name: "Bob"},
            setup: func(m *mockUserRepository) {},
            wantErr: true,
            errCode: apperrors.CodeBadRequest,
        },
        {
            name:  "duplicate email",
            input: services.CreateUserInput{Name: "Alice", Email: "alice@example.com"},
            setup: func(m *mockUserRepository) {
                m.findByEmailResult = &models.User{ID: 1}
            },
            wantErr: true,
            errCode: apperrors.CodeConflict,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &mockUserRepository{}
            tt.setup(mockRepo)
            service := services.NewUserService(mockRepo)
            
            _, err := service.Create(context.Background(), tt.input)
            
            if tt.wantErr {
                if err == nil {
                    t.Fatal("expected error, got nil")
                }
                var appErr *apperrors.AppError
                if errors.As(err, &appErr) && appErr.Code != tt.errCode {
                    t.Errorf("expected code %s, got %s", tt.errCode, appErr.Code)
                }
            } else {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
            }
        })
    }
}
```

---

## Testes de Integração

### Database Integration Tests

```go
package integration_test

import (
    "context"
    "database/sql"
    "testing"
    
    _ "github.com/lib/pq"
    "yourmodule/internal/adapters"
    "yourmodule/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    dsn := "postgres://test:test@localhost:5432/test_db?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("failed to connect to test db: %v", err)
    }
    
    // Run migrations
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL
        )
    `)
    if err != nil {
        t.Fatalf("failed to create table: %v", err)
    }
    
    // Cleanup
    t.Cleanup(func() {
        db.Exec("DROP TABLE users")
        db.Close()
    })
    
    return db
}

func TestUserRepository_Create_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    db := setupTestDB(t)
    repo := repositories.NewUserRepository(db)
    
    user := &models.User{
        Name:  "Alice",
        Email: "alice@example.com",
    }
    
    err := repo.Create(context.Background(), user)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if user.ID == 0 {
        t.Error("expected ID to be set")
    }
    
    // Verify in database
    found, err := repo.FindByID(context.Background(), user.ID)
    if err != nil {
        t.Fatalf("failed to find user: %v", err)
    }
    if found.Email != "alice@example.com" {
        t.Errorf("expected alice@example.com, got %s", found.Email)
    }
}
```

### API Integration Tests

```go
package integration_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    // Ajuste para o diretório real do seu comando em cmd/<nome> ou cmd/<nome>-worker
    _ "yourmodule/cmd/foobar"
)

func setupTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    
    cfg := &config.Config{
        HTTP: config.HTTPConfig{Port: 0},
        // ... outras configs
    }
    
    app := app.New(cfg)
    // Wire dependencies com mocks ou test database
    
    return httptest.NewServer(app.Router)
}

func TestAPI_CreateUser_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    srv := setupTestServer(t)
    defer srv.Close()
    
    // Create user
    payload := map[string]string{
        "name":  "Alice",
        "email": "alice@example.com",
    }
    body, _ := json.Marshal(payload)
    
    resp, err := http.Post(srv.URL+"/api/v1/users", "application/json", bytes.NewReader(body))
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        t.Errorf("expected 201, got %d", resp.StatusCode)
    }
    
    var envelope response.Envelope[models.User]
    json.NewDecoder(resp.Body).Decode(&envelope)
    
    if envelope.Data.Name != "Alice" {
        t.Errorf("expected Alice, got %s", envelope.Data.Name)
    }
}
```

---

## Mocks e Stubs

### Manual Mocks

```go
// mockUserRepository implementa a interface UserRepository
type mockUserRepository struct {
    // Controle de chamadas
    createCalled      bool
    findByIDCalled    bool
    findByEmailCalled bool
    
    // Dados de retorno
    findByIDResult    *models.User
    findByEmailResult *models.User
    
    // Erros de retorno
    createErr      error
    findByIDErr    error
    findByEmailErr error
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
    m.createCalled = true
    if m.createErr != nil {
        return m.createErr
    }
    user.ID = 1  // simula ID gerado
    return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
    m.findByIDCalled = true
    return m.findByIDResult, m.findByIDErr
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    m.findByEmailCalled = true
    return m.findByEmailResult, m.findByEmailErr
}
```

### Usando Testify/Mock (Opcional)

```bash
go get github.com/stretchr/testify/mock
```

```go
import "github.com/stretchr/testify/mock"

type mockUserRepository struct {
    mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

// Uso
func TestUserService_Create_WithTestify(t *testing.T) {
    mockRepo := new(mockUserRepository)
    mockRepo.On("FindByEmail", mock.Anything, "alice@example.com").Return(nil, nil)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    service := services.NewUserService(mockRepo)
    
    _, err := service.Create(context.Background(), services.CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    mockRepo.AssertExpectations(t)
}
```

---

## Test Helpers

### pkg/testhelper

```go
import "github.com/fvmoraes/ginger/pkg/testhelper"

// HTTP request helper
rec := testhelper.NewRequest(t, handler, http.MethodGet, "/users/1").Do()

// Com body
rec := testhelper.NewRequest(t, handler, http.MethodPost, "/users").
    WithBody(map[string]string{"name": "Alice"}).
    Do()

// Com headers
rec := testhelper.NewRequest(t, handler, http.MethodGet, "/users").
    WithHeader("Authorization", "Bearer token").
    Do()

// Assertions
testhelper.AssertStatus(t, rec, http.StatusOK)
testhelper.AssertHeader(t, rec, "Content-Type", "application/json")

// Decode JSON
var user models.User
testhelper.DecodeJSON(t, rec, &user)
```

### Custom Test Helpers

```go
// testhelpers/database.go
package testhelpers

func NewTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    
    t.Cleanup(func() { db.Close() })
    
    return db
}

func SeedUsers(t *testing.T, db *sql.DB, users []models.User) {
    t.Helper()
    
    for _, u := range users {
        _, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
        if err != nil {
            t.Fatalf("failed to seed user: %v", err)
        }
    }
}
```

---

## Coverage

### Executar com Coverage

```bash
# Gerar coverage
go test -coverprofile=coverage.out ./...

# Ver coverage no terminal
go tool cover -func=coverage.out

# Gerar HTML
go tool cover -html=coverage.out -o coverage.html
```

### Coverage por Pacote

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "^total:"
```

**Output:**
```
total:  (statements)    78.5%
```

### Makefile Targets

```makefile
.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: test-short
test-short:
	go test -v -short -timeout 10s ./...

.PHONY: test-integration
test-integration:
	go test -v -run Integration ./tests/...

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: coverage-ci
coverage-ci:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
```

---

## CI/CD

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test -v -short -race -coverprofile=coverage.out ./...
      
      - name: Run integration tests
        env:
          DATABASE_DSN: postgres://test:test@localhost:5432/test_db?sslmode=disable
        run: go test -v -run Integration ./tests/...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test
  - coverage

test:
  stage: test
  image: golang:1.25
  services:
    - postgres:15
  variables:
    POSTGRES_DB: test_db
    POSTGRES_USER: test
    POSTGRES_PASSWORD: test
    DATABASE_DSN: postgres://test:test@postgres:5432/test_db?sslmode=disable
  script:
    - go mod download
    - go test -v -short -race ./...
    - go test -v -run Integration ./tests/...

coverage:
  stage: coverage
  image: golang:1.25
  script:
    - go test -coverprofile=coverage.out -covermode=atomic ./...
    - go tool cover -func=coverage.out
  coverage: '/total:.*?(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

---

## Boas Práticas

### 1. Teste Comportamento, Não Implementação

```go
// ❌ Ruim — testa implementação
func TestUserService_Create_CallsRepositoryCreate(t *testing.T) {
    mockRepo := &mockUserRepository{}
    service := services.NewUserService(mockRepo)
    
    service.Create(ctx, input)
    
    if !mockRepo.createCalled {
        t.Error("expected Create to be called")
    }
}

// ✅ Bom — testa comportamento
func TestUserService_Create_ReturnsCreatedUser(t *testing.T) {
    mockRepo := &mockUserRepository{}
    service := services.NewUserService(mockRepo)
    
    user, err := service.Create(ctx, input)
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != input.Name {
        t.Errorf("expected %s, got %s", input.Name, user.Name)
    }
}
```

### 2. Use t.Helper()

```go
func assertUser(t *testing.T, got, want *models.User) {
    t.Helper()  // stack trace aponta para o caller
    
    if got.Name != want.Name {
        t.Errorf("name: expected %s, got %s", want.Name, got.Name)
    }
    if got.Email != want.Email {
        t.Errorf("email: expected %s, got %s", want.Email, got.Email)
    }
}
```

### 3. Cleanup com t.Cleanup()

```go
func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { os.Remove(f.Name()) })
    
    // test logic
}
```

### 4. Parallel Tests

```go
func TestUserService_Create(t *testing.T) {
    t.Parallel()  // roda em paralelo com outros testes
    
    // test logic
}
```

### 5. Skip Slow Tests

```go
func TestIntegration_SlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test")
    }
    
    // test logic
}
```

```bash
# Roda apenas testes rápidos
go test -short ./...
```

---

## Próximos Passos

- [🏗️ Arquitetura](./ARCHITECTURE.md) — Estrutura e padrões
- [📦 Pacotes](./PACKAGES.md) — Documentação de cada pacote
- [🔌 Integrações](./INTEGRATIONS.md) — Bancos, cache, mensageria
- [🚀 Deploy](./DEPLOYMENT.md) — Docker, Kubernetes, CI/CD

[← Voltar ao README](../README.md)
