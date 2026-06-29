# Fase 3 — Funcionalidades Avançadas

> 5 pacotes + 1 integração para posicionar o Ginger como escolha enterprise e sistemas complexos.

---

## 1. `pkg/tenant` — Multi-Tenancy

**Status**: Inexistente | **Prioridade**: BAIXA | **Dependências**: `pkg/auth`

### Motivação
SaaS modernos precisam isolar dados por tenant. O middleware extrai o tenant ID e o disponibiliza no context. O desenvolvedor escreve queries com `WHERE tenant_id = $1` explicitamente — sem magia.

### Design

```go
// Middleware extrai tenant do header/subdomain/JWT
tenant := tenant.New(tenant.Config{
    Strategy: tenant.HeaderStrategy("X-Tenant-ID"),
    // ou: tenant.SubdomainStrategy(),
    // ou: tenant.JWTClaimStrategy("tenant_id"),
})
r.Use(tenant.Middleware)

// No handler
tenantID := tenant.FromContext(r.Context())
db.Query("SELECT * FROM users WHERE tenant_id = $1", tenantID)
```

### API planejada

```go
package tenant

type Strategy func(*http.Request) (string, error)
func HeaderStrategy(header string) Strategy
func SubdomainStrategy() Strategy
func JWTClaimStrategy(claim string) Strategy

type Config struct {
    Strategy Strategy
    Optional bool // permite requests sem tenant
}

// Middleware
func New(cfg Config) *Tenant
func (t *Tenant) Middleware(next http.Handler) http.Handler
func FromContext(ctx context.Context) string
```

### Nota pós-revisão
`ScopeDB` (wrapper que injetava `WHERE tenant_id` automaticamente) foi removido. Violava o princípio "explícito, sem magia" e exigiria parsing de SQL. O desenvolvedor escreve o `WHERE` manualmente usando `tenant.FromContext()`.

---

## 2. `pkg/audit` — Audit Logging

**Status**: Inexistente | **Prioridade**: BAIXA | **Dependências**: `pkg/auth`, `pkg/logger`

### Motivação
Sistemas enterprise precisam registrar quem fez o quê e quando para compliance (SOX, GDPR, LGPD, PCI).

### Design

```go
// Middleware que audita automaticamente mutations
audit := audit.New(audit.Config{
    Store:  audit.NewSQLStore(db),
    Methods: []string{"POST", "PUT", "PATCH", "DELETE"},
    BodyCapture: true, // captura o body da requisição
})
r.Use(audit.Middleware)

// Audit manual no handler
audit.Log(r.Context(), audit.Entry{
    Action: "user.login",
    Resource: "user/123",
    Metadata: map[string]string{"ip": r.RemoteAddr},
})
```

### API planejada

```go
package audit

type Entry struct {
    ID        string
    Timestamp time.Time
    Actor     string // user ID do context
    Action    string // "user.create", "payment.refund"
    Resource  string // "user/123"
    Method    string // HTTP method
    Path      string
    Status    int
    Body      string // opcional, request body
    Metadata  map[string]string
    Duration  time.Duration
}

type Store interface {
    Save(ctx context.Context, entry Entry) error
    Query(ctx context.Context, filter Filter) ([]Entry, error)
}

type Config struct {
    Store       Store
    Methods     []string // quais métodos auditar
    BodyCapture bool     // capturar request body
    MaxBodySize int      // default: 64KB
}

// Stores built-in
func NewSQLStore(db *sql.DB) Store
func NewLogStore(log *logger.Logger) Store // audit via logger

// Middleware
func New(cfg Config) *Audit
func (a *Audit) Middleware(next http.Handler) http.Handler

// Manual logging
func Log(ctx context.Context, entry Entry)
```

---

## 3. `pkg/featureflags` — Feature Flags

**Status**: Inexistente | **Prioridade**: BAIXA | **Dependências**: Nenhuma

### Motivação
Deploy contínuo sem feature flags é arriscado. Equipes precisam ativar/desativar funcionalidades sem redeploy.

### Design

```go
// Inicialização
ff := featureflags.New(featureflags.Config{
    Backend: featureflags.NewEnvBackend(), // flags via env vars
})

// No handler
if ff.IsEnabled(r.Context(), "new-checkout-flow") {
    newCheckout(w, r)
} else {
    oldCheckout(w, r)
}

// Middleware que bloqueia rotas por flag
r.Group("/api/v2", featureflags.RequireFlag("api-v2")).GET("/users", listUsers)
// Se flag desativada: 503 Service Unavailable
// Body: {"code":"FEATURE_DISABLED","message":"feature 'api-v2' is not enabled"}
```

### API planejada

```go
package featureflags

type Config struct {
    Backend Backend
    TTL     time.Duration // cache local de flags
}

type Backend interface {
    IsEnabled(ctx context.Context, flag string) bool
    All(ctx context.Context) map[string]bool
}

// Backends built-in
func NewEnvBackend() Backend           // FEATURE_<NAME>=true
func NewStaticBackend(flags map[string]bool) Backend
func NewRedisBackend(client *redis.Client) Backend

// Operações
func New(cfg Config) *FeatureFlags
func (ff *FeatureFlags) IsEnabled(ctx context.Context, flag string) bool
func (ff *FeatureFlags) Middleware(flag string) middleware.Func // 404 se flag desativada
```

---

## 4. `pkg/events` — Event Bus

**Status**: Inexistente | **Prioridade**: BAIXA | **Dependências**: Nenhuma

### Motivação
Desacoplar serviços via eventos é um padrão arquitetural fundamental. O Ginger deve oferecer um event bus local simples + suporte a outbox pattern para garantia de entrega.

### Design

```go
// Inicialização
bus := events.NewBus()

// Publicar
bus.Publish(ctx, events.Event{
    Type: "user.created",
    Data: UserCreated{ID: "123", Email: "a@b.com"},
})

// Assinar
bus.Subscribe("user.created", func(ctx context.Context, evt events.Event) error {
    // enviar email de boas-vindas
    return nil
})

// Outbox pattern (garantia de entrega via banco)
outbox := events.NewOutbox(db, events.OutboxConfig{
    PollInterval: time.Second,
    Publisher:    bus,
})
app.OnStop(outbox.Shutdown)
```

### API planejada

```go
package events

type Event struct {
    ID        string
    Type      string
    Data      any
    Timestamp time.Time
    Metadata  map[string]string
}

type Handler func(ctx context.Context, evt Event) error

type Bus struct { ... }
func NewBus() *Bus
func (b *Bus) Publish(ctx context.Context, evt Event) error
func (b *Bus) Subscribe(eventType string, handler Handler)
func (b *Bus) Unsubscribe(eventType string, handler Handler)

// Outbox (garantia de entrega)
type OutboxConfig struct {
    PollInterval time.Duration
    Publisher    *Bus
    MaxRetries   int
}

type Outbox struct { ... }
func NewOutbox(db *sql.DB, cfg OutboxConfig) *Outbox
func (o *Outbox) Append(ctx context.Context, evt Event) error // salva no banco
func (o *Outbox) Shutdown(ctx context.Context) error
```

### Limitação conhecida
`NewOutbox` é acoplado a `*sql.DB`. Para outros bancos (MongoDB, etc.), o usuário deve implementar seu próprio outbox. Na v2, introduzir interface `OutboxStore`.

---

## 5. `pkg/scheduler` — Job Scheduler

**Status**: Inexistente | **Prioridade**: BAIXA | **Dependências**: Nenhuma

### Motivação**
Tarefas agendadas (limpeza de tokens expirados, relatórios diários, sincronização) são comuns em todo sistema.

### Design

```go
sched := scheduler.New()

// Cron-style
sched.Add("cleanup-expired-tokens", "0 3 * * *", func(ctx context.Context) error {
    return tokenService.CleanupExpired(ctx)
})

// Intervalo fixo
sched.Add("health-check", "@every 30s", func(ctx context.Context) error {
    return healthChecker.CheckAll(ctx)
})

// Integração com ciclo de vida
app.OnStop(sched.Shutdown)
```

### API planejada

```go
package scheduler

type Job struct {
    Name     string
    Schedule string // cron expression ou "@every X"
    Handler  func(ctx context.Context) error
    Timeout  time.Duration // default: 30s
}

type Scheduler struct { ... }
func New() *Scheduler
func (s *Scheduler) Add(name, schedule string, handler func(context.Context) error) error
func (s *Scheduler) Remove(name string)
func (s *Scheduler) List() []Job
func (s *Scheduler) Run(name string) error // executa job manualmente
func (s *Scheduler) Shutdown(ctx context.Context) error
```

### Cron expression support (stdlib-only parser)
- `* * * * *` — minuto, hora, dia do mês, mês, dia da semana
- `@every 30s` / `@every 5m` / `@every 1h`
- `@daily`, `@hourly`

---

## 6. Integração GraphQL — `ginger add graphql`

**Status**: Inexistente | **Prioridade**: BAIXA | **Dependências**: `internal/integrations`

### Motivação
O plano original incluía um gerador de código GraphQL completo. Após revisão, concluiu-se que isso exigiria um mini-gqlgen. Em vez disso, fornecemos um template de integração que configura o gqlgen com padrões Ginger.

### Design

```bash
ginger add graphql

# Gera:
# internal/graphql/schema/schema.graphql     — schema vazio com scaffold
# internal/graphql/gqlgen.yml                — config do gqlgen
# internal/graphql/resolvers/resolver.go     — resolver stub no padrão Ginger
# internal/graphql/server.go                 — handler HTTP integrado ao router
```

### Escopo
- Template gera arquivos base para [gqlgen](https://github.com/99designs/gqlgen)
- Resolver stub segue o padrão Ginger (port → service → adapter)
- Handler HTTP usa o router Ginger
- Playground GraphQL na rota `/graphql`
- Após `ginger add graphql`, usuário roda `go run github.com/99designs/gqlgen generate`
