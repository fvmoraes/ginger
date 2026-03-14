<div align="center">
  <img src="../GINGER_LOGO.png" alt="Ginger Logo" width="180"/>
  <h1>Ginger</h1>
  <p><strong>Agilize e padronize projetos Go</strong></p>

  ![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat&logo=go)
  ![License](https://img.shields.io/badge/license-MIT-green?style=flat)
  ![Build](https://img.shields.io/badge/build-passing-brightgreen?style=flat)
</div>

> **Requires Go 1.25+** — Ginger depends on `go.opentelemetry.io/otel v1.42` which sets the minimum Go version to 1.25. All projects scaffolded by `ginger new` also target `go 1.25`.
>
> **Requer Go 1.25+** — O Ginger depende de `go.opentelemetry.io/otel v1.42`, que exige Go 1.25 como versão mínima. Todos os projetos gerados por `ginger new` também usam `go 1.25`.

---

## English

- [What is Ginger?](#what-is-ginger)
- [Core Principles](#core-principles)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [CLI Reference](#cli-reference)
- [Core Packages](#core-packages)
- [Example App](#example-app)
- [Configuration](#configuration)
- [Observability](#observability)
- [Docker & Kubernetes](#docker--kubernetes)

## Português

- [O que é o Ginger?](#o-que-é-o-ginger)
- [Princípios](#princípios)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Começando](#começando)
- [Referência da CLI](#referência-da-cli)
- [Pacotes Principais](#pacotes-principais)
- [App de Exemplo](#app-de-exemplo)
- [Configuração](#configuração)
- [Observabilidade](#observabilidade)
- [Docker e Kubernetes](#docker-e-kubernetes)

---

# 🇺🇸 English

## What is Ginger?

Ginger is a CLI tool and set of packages that accelerates and standardizes Go projects across teams. It is not a replacement for the standard library — it is a thin layer on top of it that enforces conventions, eliminates boilerplate, and ships with a CLI to scaffold new projects and generate code.

**Ginger does not hide Go from you. It organizes it.**

## Core Principles

- **Minimal dependencies** — only what is strictly necessary
- **Fast compilation** — no magic, no reflection-heavy DI
- **Idiomatic Go** — standard interfaces, standard patterns
- **Simple CLI** — scaffold, generate, run, build
- **Clear project structure** — every team member knows where things live
- **Developer productivity** — less setup, more shipping

## Project Structure

Every project created with `ginger new` follows this layout:

```
my-api/
├── cmd/
│   └── app/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP handlers
│   │   ├── services/            # Business logic
│   │   ├── repositories/        # Data access layer
│   │   └── middlewares/         # App-specific middlewares
│   ├── models/                  # Domain models
│   └── config/                  # Config loader wrapper
├── pkg/                         # Reusable internal packages
├── platform/                    # External integrations (DB, cache, messaging)
├── configs/
│   └── app.yaml                 # Application configuration
├── scripts/                     # Dev and CI scripts
├── tests/                       # Integration tests
├── docs/                        # Documentation
├── Dockerfile
├── Makefile
└── .env.example
```

## Getting Started

### Install the CLI

> Requires **Go 1.25+**. Check your version with `go version`.

```bash
git clone https://github.com/ginger-framework/ginger
cd ginger
go build -o /usr/local/bin/ginger ./cmd/ginger
```

### Create a new project

```bash
ginger new my-api
cd my-api
go mod tidy
ginger run
```

Your API is now running at `http://localhost:8080`.

Built-in endpoints available immediately:

| Method | Path      | Description        |
|--------|-----------|--------------------|
| GET    | /health   | Health check       |
| GET    | /api/v1/ping | Ping endpoint   |

## CLI Reference

```
ginger new <project-name>          Scaffold a new project
ginger run                         Run the app (go run ./cmd/app)
ginger build [output]              Build the binary
ginger generate handler <name>     Generate an HTTP handler
ginger generate service <name>     Generate a service
ginger generate repository <name>  Generate a repository
ginger generate crud <name>        Generate full CRUD (model+handler+service+repo+test)
ginger add <integration>           Add an integration to the project
ginger doctor                      Run project health diagnostics
ginger version                     Print Ginger version
ginger help                        Show help
```

### Integrations (`ginger add`)

| Category    | Command                    | Package                              |
|-------------|----------------------------|--------------------------------------|
| Databases   | `ginger add postgres`      | `github.com/lib/pq`                  |
|             | `ginger add mysql`         | `github.com/go-sql-driver/mysql`     |
|             | `ginger add sqlite`        | `github.com/mattn/go-sqlite3`        |
|             | `ginger add sqlserver`     | `github.com/microsoft/go-mssqldb`    |
| NoSQL       | `ginger add couchbase`     | `github.com/couchbase/gocb/v2`       |
|             | `ginger add mongodb`       | `go.mongodb.org/mongo-driver`        |
| Analytical  | `ginger add clickhouse`    | `github.com/ClickHouse/clickhouse-go/v2` |
|             | `ginger add rabbitmq`      | `github.com/rabbitmq/amqp091-go`     |
|             | `ginger add nats`          | `github.com/nats-io/nats.go`         |
|             | `ginger add pubsub`        | `cloud.google.com/go/pubsub`         |
| Protocols   | `ginger add grpc`          | `google.golang.org/grpc`             |
|             | `ginger add mcp`           | stdlib only                          |
| Observ.     | `ginger add otel`          | `go.opentelemetry.io/otel`           |
|             | `ginger add prometheus`    | `github.com/prometheus/client_golang`|

### Code generation example

```bash
ginger generate handler  product
ginger generate service  product
ginger generate repository product
```

This creates:

```
internal/api/handlers/product_handler.go
internal/api/services/product_service.go
internal/api/repositories/product_repository.go
```

Each file is fully wired with the correct interface, constructor, and method stubs — ready to fill in.

## Core Packages

### `pkg/app` — Application bootstrap

```go
cfg, _ := config.Load("configs/app.yaml")
app := gingerapp.New(cfg)

app.Router.Use(middleware.CORS())
app.OnStop(func(ctx context.Context) error {
    return db.Close()
})

app.Run() // blocks, handles SIGINT/SIGTERM
```

### `pkg/router` — HTTP routing

Wraps `net/http` ServeMux with method helpers, route groups, and JSON utilities.

```go
v1 := app.Router.Group("/api/v1")
v1.GET("/users", listUsers)
v1.POST("/users", createUser)

// JSON response
router.JSON(w, http.StatusOK, payload)

// Standardized error response
router.Error(w, apperrors.NotFound("user not found"))

// Decode request body
router.Decode(r, &input)
```

### `pkg/errors` — Typed errors

```go
apperrors.NotFound("user not found")       // 404
apperrors.BadRequest("invalid input")      // 400
apperrors.Unauthorized("token expired")    // 401
apperrors.Forbidden("access denied")       // 403
apperrors.Conflict("email already exists") // 409
apperrors.Internal(err)                    // 500
```

All errors serialize to a consistent JSON shape:

```json
{
  "code": "NOT_FOUND",
  "message": "user not found"
}
```

### `pkg/middleware` — Built-in middlewares

```go
middleware.Logger(log)    // structured request logging
middleware.Recover(log)   // panic recovery → 500
middleware.RequestID()    // injects X-Request-ID
middleware.CORS("*")      // CORS headers

// Compose
middleware.Chain(mw1, mw2, mw3)
```

### `pkg/health` — Health checks

```go
h := health.New()
h.Register(database.NewChecker(db)) // plug in any Checker

// Automatically mounted at GET /health
// Returns 200 if all checks pass, 503 otherwise
```

```json
{
  "healthy": true,
  "checks": [{ "name": "database", "healthy": true }],
  "duration": "1.2ms"
}
```

### `pkg/config` — Configuration

Loads from YAML file first, then overrides with environment variables.

```go
cfg, err := config.Load("configs/app.yaml")
// cfg.App.Name, cfg.HTTP.Port, cfg.Database.DSN, etc.
```

### `pkg/logger` — Structured logging

Built on `log/slog`. JSON by default, text for local dev.

```go
log := logger.New("info", "json")
log.Info("user created", "id", user.ID)
log.Error("db error", "error", err)

// Context-aware
ctx = logger.WithContext(ctx, log)
logger.FromContext(ctx).Info("handled")
```

### `pkg/telemetry` — OpenTelemetry

```go
provider, err := telemetry.Setup(ctx, telemetry.Config{
    ServiceName:    cfg.App.Name,
    ServiceVersion: cfg.App.Version,
    Exporter:       "stdout", // swap for "otlp" in production
})
defer provider.Shutdown(ctx)

tracer := telemetry.Tracer("my-api")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()
```

### `pkg/testhelper` — Test utilities

```go
rec := testhelper.NewRequest(t, handler, http.MethodGet, "/users").Do()
testhelper.AssertStatus(t, rec, http.StatusOK)

var result []User
testhelper.DecodeJSON(t, rec, &result)
```

## Example App

The `example/` directory contains a complete User CRUD API demonstrating the full Ginger stack:

```
example/
├── cmd/app/main.go                          # wires everything together
├── internal/
│   ├── models/user.go                       # User, CreateUserInput, UpdateUserInput
│   └── api/
│       ├── handlers/user_handler.go         # HTTP layer
│       ├── services/user_service.go         # Business logic
│       └── repositories/user_repository.go  # Data access
└── configs/app.yaml
```

```bash
cd example
go mod tidy
go run ./cmd/app
```

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# List users
curl http://localhost:8080/api/v1/users

# Health check
curl http://localhost:8080/health
```

## Configuration

`configs/app.yaml`:

```yaml
app:
  name: my-api
  env: development
  version: 0.1.0

http:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30  # seconds

database:
  driver: postgres
  dsn: postgres://user:pass@localhost:5432/mydb?sslmode=disable
  max_open: 25
  max_idle: 5

log:
  level: info    # debug | info | warn | error
  format: json   # json | text
```

All fields can be overridden with environment variables:

| Env var           | Config field              |
|-------------------|---------------------------|
| `APP_NAME`        | `app.name`                |
| `APP_ENV`         | `app.env`                 |
| `HTTP_PORT`       | `http.port`               |
| `DATABASE_DSN`    | `database.dsn`            |
| `LOG_LEVEL`       | `log.level`               |
| `LOG_FORMAT`      | `log.format`              |

## Observability

Ginger ships with OpenTelemetry integration out of the box. The default exporter writes traces to stdout. Swap it for OTLP to send to Jaeger, Tempo, or any OTel-compatible backend.

```go
provider, _ := telemetry.Setup(ctx, telemetry.Config{
    ServiceName: "my-api",
    Exporter:    "otlp", // configure OTEL_EXPORTER_OTLP_ENDPOINT env var
})
```

## Docker & Kubernetes

A `Dockerfile` is generated with every new project using a multi-stage build:

```bash
docker build -t my-api:latest .
docker run -p 8080:8080 my-api:latest
```

A Kubernetes `Deployment` + `Service` template is available at `templates/k8s/deployment.yaml`. It includes readiness and liveness probes pointed at `/health`, resource limits, and a `ClusterIP` service.

---

# 🇧🇷 Português

## O que é o Ginger?

Ginger é uma CLI e um conjunto de pacotes que agiliza e padroniza projetos Go entre equipes. Ele não substitui a biblioteca padrão — é uma camada fina sobre ela que impõe convenções, elimina boilerplate e vem com uma CLI para criar projetos e gerar código.

**O Ginger não esconde o Go de você. Ele o organiza.**

## Princípios

- **Dependências mínimas** — apenas o estritamente necessário
- **Compilação rápida** — sem mágica, sem DI pesada em reflection
- **Go idiomático** — interfaces padrão, padrões padrão
- **CLI simples** — scaffold, generate, run, build
- **Estrutura de projeto clara** — todo membro da equipe sabe onde as coisas ficam
- **Produtividade do desenvolvedor** — menos setup, mais entrega

## Estrutura do Projeto

Todo projeto criado com `ginger new` segue este layout:

```
my-api/
├── cmd/
│   └── app/
│       └── main.go              # Ponto de entrada da aplicação
├── internal/
│   ├── api/
│   │   ├── handlers/            # Handlers HTTP
│   │   ├── services/            # Lógica de negócio
│   │   ├── repositories/        # Camada de acesso a dados
│   │   └── middlewares/         # Middlewares específicos da aplicação
│   ├── models/                  # Modelos de domínio
│   └── config/                  # Wrapper do carregador de configuração
├── pkg/                         # Pacotes internos reutilizáveis
├── platform/                    # Integrações externas (DB, cache, mensageria)
├── configs/
│   └── app.yaml                 # Configuração da aplicação
├── scripts/                     # Scripts de dev e CI
├── tests/                       # Testes de integração
├── docs/                        # Documentação
├── Dockerfile
├── Makefile
└── .env.example
```

## Começando

### Instalar a CLI

> Requer **Go 1.25+**. Verifique com `go version`.

```bash
git clone https://github.com/ginger-framework/ginger
cd ginger
go build -o /usr/local/bin/ginger ./cmd/ginger
```

### Criar um novo projeto

```bash
ginger new minha-api
cd minha-api
go mod tidy
ginger run
```

Sua API estará rodando em `http://localhost:8080`.

Endpoints disponíveis imediatamente:

| Método | Caminho      | Descrição          |
|--------|--------------|--------------------|
| GET    | /health      | Health check       |
| GET    | /api/v1/ping | Endpoint de ping   |

## Referência da CLI

```
ginger new <nome-do-projeto>       Cria um novo projeto com scaffold completo
ginger run                         Executa a aplicação (go run ./cmd/app)
ginger build [saída]               Compila o binário
ginger generate handler <nome>     Gera um handler HTTP
ginger generate service <nome>     Gera um service
ginger generate repository <nome>  Gera um repository
ginger generate crud <nome>        Gera CRUD completo (model+handler+service+repo+test)
ginger add <integração>            Adiciona uma integração ao projeto
ginger doctor                      Diagnóstico de saúde do projeto
ginger version                     Exibe a versão do Ginger
ginger help                        Exibe a ajuda
```

### Integrações (`ginger add`)

| Categoria   | Comando                    | Pacote                               |
|-------------|----------------------------|--------------------------------------|
| Bancos      | `ginger add postgres`      | `github.com/lib/pq`                  |
|             | `ginger add mysql`         | `github.com/go-sql-driver/mysql`     |
|             | `ginger add sqlite`        | `github.com/mattn/go-sqlite3`        |
|             | `ginger add sqlserver`     | `github.com/microsoft/go-mssqldb`    |
| NoSQL       | `ginger add couchbase`     | `github.com/couchbase/gocb/v2`       |
|             | `ginger add mongodb`       | `go.mongodb.org/mongo-driver`        |
| Analítico   | `ginger add clickhouse`    | `github.com/ClickHouse/clickhouse-go/v2` |
| Cache       | `ginger add redis`         | `github.com/redis/go-redis/v9`       |
| Mensageria  | `ginger add kafka`         | `github.com/segmentio/kafka-go`      |
|             | `ginger add rabbitmq`      | `github.com/rabbitmq/amqp091-go`     |
|             | `ginger add nats`          | `github.com/nats-io/nats.go`         |
|             | `ginger add pubsub`        | `cloud.google.com/go/pubsub`         |
| Protocolos  | `ginger add grpc`          | `google.golang.org/grpc`             |
|             | `ginger add mcp`           | stdlib only                          |
| Observ.     | `ginger add otel`          | `go.opentelemetry.io/otel`           |
|             | `ginger add prometheus`    | `github.com/prometheus/client_golang`|

### Exemplo de geração de código

```bash
ginger generate handler  produto
ginger generate service  produto
ginger generate repository produto
```

Isso cria:

```
internal/api/handlers/produto_handler.go
internal/api/services/produto_service.go
internal/api/repositories/produto_repository.go
```

Cada arquivo já vem com a interface correta, construtor e stubs de métodos — pronto para preencher.

## Pacotes Principais

### `pkg/app` — Bootstrap da aplicação

```go
cfg, _ := config.Load("configs/app.yaml")
app := gingerapp.New(cfg)

app.Router.Use(middleware.CORS())
app.OnStop(func(ctx context.Context) error {
    return db.Close()
})

app.Run() // bloqueia, trata SIGINT/SIGTERM
```

### `pkg/router` — Roteamento HTTP

Encapsula o `net/http` ServeMux com helpers de método, grupos de rotas e utilitários JSON.

```go
v1 := app.Router.Group("/api/v1")
v1.GET("/usuarios", listarUsuarios)
v1.POST("/usuarios", criarUsuario)

// Resposta JSON
router.JSON(w, http.StatusOK, payload)

// Resposta de erro padronizada
router.Error(w, apperrors.NotFound("usuário não encontrado"))

// Decodificar body da requisição
router.Decode(r, &input)
```

### `pkg/errors` — Erros tipados

```go
apperrors.NotFound("usuário não encontrado")    // 404
apperrors.BadRequest("entrada inválida")        // 400
apperrors.Unauthorized("token expirado")        // 401
apperrors.Forbidden("acesso negado")            // 403
apperrors.Conflict("email já cadastrado")       // 409
apperrors.Internal(err)                         // 500
```

Todos os erros serializam para um formato JSON consistente:

```json
{
  "code": "NOT_FOUND",
  "message": "usuário não encontrado"
}
```

### `pkg/middleware` — Middlewares embutidos

```go
middleware.Logger(log)    // log estruturado de requisições
middleware.Recover(log)   // recuperação de panic → 500
middleware.RequestID()    // injeta X-Request-ID
middleware.CORS("*")      // headers CORS

// Composição
middleware.Chain(mw1, mw2, mw3)
```

### `pkg/health` — Health checks

```go
h := health.New()
h.Register(database.NewChecker(db)) // implemente a interface Checker

// Montado automaticamente em GET /health
// Retorna 200 se todos os checks passam, 503 caso contrário
```

```json
{
  "healthy": true,
  "checks": [{ "name": "database", "healthy": true }],
  "duration": "1.2ms"
}
```

### `pkg/config` — Configuração

Carrega do arquivo YAML primeiro, depois sobrescreve com variáveis de ambiente.

```go
cfg, err := config.Load("configs/app.yaml")
// cfg.App.Name, cfg.HTTP.Port, cfg.Database.DSN, etc.
```

### `pkg/logger` — Log estruturado

Construído sobre `log/slog`. JSON por padrão, texto para dev local.

```go
log := logger.New("info", "json")
log.Info("usuário criado", "id", usuario.ID)
log.Error("erro no banco", "error", err)

// Com contexto
ctx = logger.WithContext(ctx, log)
logger.FromContext(ctx).Info("processado")
```

### `pkg/telemetry` — OpenTelemetry

```go
provider, err := telemetry.Setup(ctx, telemetry.Config{
    ServiceName:    cfg.App.Name,
    ServiceVersion: cfg.App.Version,
    Exporter:       "stdout", // troque por "otlp" em produção
})
defer provider.Shutdown(ctx)

tracer := telemetry.Tracer("minha-api")
ctx, span := tracer.Start(ctx, "nome-da-operacao")
defer span.End()
```

### `pkg/testhelper` — Utilitários de teste

```go
rec := testhelper.NewRequest(t, handler, http.MethodGet, "/usuarios").Do()
testhelper.AssertStatus(t, rec, http.StatusOK)

var resultado []Usuario
testhelper.DecodeJSON(t, rec, &resultado)
```

## App de Exemplo

O diretório `example/` contém uma API CRUD completa de usuários demonstrando toda a stack do Ginger:

```
example/
├── cmd/app/main.go                               # conecta tudo
├── internal/
│   ├── models/user.go                            # User, CreateUserInput, UpdateUserInput
│   └── api/
│       ├── handlers/user_handler.go              # camada HTTP
│       ├── services/user_service.go              # lógica de negócio
│       └── repositories/user_repository.go       # acesso a dados
└── configs/app.yaml
```

```bash
cd example
go mod tidy
go run ./cmd/app
```

```bash
# Criar um usuário
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@exemplo.com"}'

# Listar usuários
curl http://localhost:8080/api/v1/users

# Health check
curl http://localhost:8080/health
```

## Configuração

`configs/app.yaml`:

```yaml
app:
  name: minha-api
  env: development
  version: 0.1.0

http:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30  # segundos

database:
  driver: postgres
  dsn: postgres://user:senha@localhost:5432/meubanco?sslmode=disable
  max_open: 25
  max_idle: 5

log:
  level: info    # debug | info | warn | error
  format: json   # json | text
```

Todos os campos podem ser sobrescritos por variáveis de ambiente:

| Variável de ambiente | Campo de configuração     |
|----------------------|---------------------------|
| `APP_NAME`           | `app.name`                |
| `APP_ENV`            | `app.env`                 |
| `HTTP_PORT`          | `http.port`               |
| `DATABASE_DSN`       | `database.dsn`            |
| `LOG_LEVEL`          | `log.level`               |
| `LOG_FORMAT`         | `log.format`              |

## Observabilidade

O Ginger vem com integração OpenTelemetry pronta para uso. O exportador padrão escreve traces no stdout. Troque por OTLP para enviar ao Jaeger, Tempo ou qualquer backend compatível com OTel.

```go
provider, _ := telemetry.Setup(ctx, telemetry.Config{
    ServiceName: "minha-api",
    Exporter:    "otlp", // configure a env OTEL_EXPORTER_OTLP_ENDPOINT
})
```

## Docker e Kubernetes

Um `Dockerfile` é gerado com cada novo projeto usando build multi-stage:

```bash
docker build -t minha-api:latest .
docker run -p 8080:8080 minha-api:latest
```

Um template de `Deployment` + `Service` Kubernetes está disponível em `templates/k8s/deployment.yaml`. Ele inclui probes de readiness e liveness apontando para `/health`, limites de recursos e um serviço `ClusterIP`.

---

<div align="center">
  <p>Built with ❤️ and idiomatic Go</p>
  <p>Feito com ❤️ e Go idiomático</p>
</div>
