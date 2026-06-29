# Fase 1 — Fundação do Framework Core

> 11 novos pacotes para preencher os gaps críticos do framework.
> Revisado após critical review — 3 interfaces Backend unificadas, Brotli e Prometheus removidos, GraphQL movido para template de integração.

---

## 0. `pkg/store` — Interface de Storage Unificada

**Status**: Inexistente | **Prioridade**: FUNDAÇÃO | **Dependências**: Nenhuma

### Motivação
`cache`, `ratelimit` e `idempotency` compartilham o mesmo padrão: armazenar valor com TTL, recuperar, invalidar. Três interfaces `Backend` separadas violam o princípio de simplicidade.

### Design

```go
package store

// KV é a interface única de armazenamento usada por cache, ratelimit e idempotency.
type KV interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

// Backends built-in
func NewMemoryStore(maxSize int) KV      // LRU in-memory
```

### Integração
- `pkg/cache.Backend` → `store.KV`
- `pkg/ratelimit.Backend` → `store.KV` (com `Allow` wrapper)
- `pkg/idempotency.Backend` → `store.KV` (com serialize/desserialize de `CachedResponse`)
- Backends Redis, etc. ficam nos templates de integração (`ginger add redis` gera adapter `store.KV`)

---

## 0.5 `pkg/negotiation` — Content-Type Negotiation

**Status**: Inexistente | **Prioridade**: ALTA | **Dependências**: Nenhuma

### Motivação
APIs que suportam JSON e Protobuf (gRPC-gateway) ou XML precisam negociar o formato de resposta baseado no header `Accept`. Hoje o Ginger sempre retorna JSON.

### Design

```go
// No handler
accept := negotiation.Negotiate(r, []string{"application/json", "application/xml"})
switch accept {
case "application/json":
    response.OK(w, data)
case "application/xml":
    xml.NewEncoder(w).Encode(data)
}
```

### API planejada

```go
package negotiation

// Negotiate retorna o melhor content-type suportado baseado no header Accept.
// Se nenhum match, retorna o primeiro da lista de ofertas.
func Negotiate(r *http.Request, offers []string) string

// Match verifica se o request aceita um content-type específico.
func Match(r *http.Request, contentType string) bool

// ParseAccept extrai e ordena os tipos aceitos (com quality factor).
func ParseAccept(r *http.Request) []AcceptType

type AcceptType struct {
    Type    string
    Quality float64 // 0.0 a 1.0
}
```

---

## 1. `pkg/validate` — Validação de Entrada

**Status**: Inexistente | **Prioridade**: CRÍTICA | **Dependências**: Nenhuma

### Motivação
Hoje toda validação é manual nos handlers, duplicada entre CRUDs gerados. Não existe um contrato único de validação.

### Design

```go
// Uso
type CreateUserInput struct {
    Name  string `validate:"required,min=2,max=100"`
    Email string `validate:"required,email"`
    Age   int    `validate:"min=18,max=150"`
    Role  string `validate:"oneof=admin user guest"`
}

err := validate.Struct(input)
// err => validate.Errors{"email": "must be a valid email", "age": "must be at least 18"}
```

### API planejada

```go
package validate

// Struct valida campos usando struct tags
func Struct(v any) error

// Field valida um valor contra uma regra
func Field(value any, rule string) error

// Errors é um mapa de campo -> mensagem de erro
type Errors map[string]string
func (e Errors) Error() string

// Tags suportadas (stdlib-only, sem regex para manter zero-deps)
// required, email (net/mail), min, max, len, url (net/url), uuid, oneof, alphanum
```

### Integração com o framework
- `router.DecodeAndValidate(r, v) error` — novo método que chama `Decode` + `validate.Struct(v)`
- `router.Decode()` **permanece inalterado** — sem breaking change
- Templates de CRUD geram structs com tags `validate`
- `response.Errors()` helper para retornar erros de validação como 422

---

## 2. `pkg/auth` — Autenticação e Autorização

**Status**: Inexistente | **Prioridade**: CRÍTICA | **Dependências**: stdlib (`crypto/hmac`, `crypto/sha256`)

### Motivação
Hoje o Ginger não fornece nenhum mecanismo de auth. Todo projeto precisa implementar JWT ou API Key do zero ou buscar libs externas.

### Design

```go
// JWT Middleware
auth := auth.NewJWT(auth.JWTConfig{
    Secret:     []byte(os.Getenv("JWT_SECRET")),
    HeaderName: "Authorization",
    Scheme:     "Bearer",
})
r.Use(auth.Middleware)

// No handler, extrair claims
claims := auth.ClaimsFromContext(r.Context())
userID := claims.Subject

// API Key Middleware
apikey := auth.NewAPIKey(auth.APIKeyConfig{
    Keys:    map[string]string{"key1": "admin", "key2": "readonly"},
    Header:  "X-API-Key",
})
r.Use(apikey.Middleware)

// RBAC Helper
if !auth.HasRole(ctx, "admin") {
    return router.Error(w, errors.Forbidden("admin role required"))
}
```

### API planejada

```go
package auth

// JWT
type JWTConfig struct {
    Secret     []byte        // HMAC-SHA256 secret
    HeaderName string        // default: "Authorization"
    Scheme     string        // default: "Bearer"
    ClaimsFunc func(*Claims) error // validação extra (exp, iss, aud)
}

// Middleware Required: falha com 401 se token ausente/inválido
func (j *JWT) Required() middleware.Func

// Middleware Optional: sempre passa, ClaimsFromContext pode retornar claims vazias
func (j *JWT) Optional() middleware.Func

// IsAuthenticated verifica se há claims no contexto (útil com Optional)
func IsAuthenticated(ctx context.Context) bool

type Claims struct {
    Subject  string
    Roles    []string
    Metadata map[string]string
}

func NewJWT(cfg JWTConfig) *JWT
func (j *JWT) Middleware(next http.Handler) http.Handler
func ClaimsFromContext(ctx context.Context) *Claims

// API Key
type APIKeyConfig struct {
    Keys   map[string]string // key → role
    Header string            // default: "X-API-Key"
}

func NewAPIKey(cfg APIKeyConfig) *APIKey
func (a *APIKey) Middleware(next http.Handler) http.Handler

// RBAC helpers
func HasRole(ctx context.Context, role string) bool
func RequireRole(role string) middleware.Func
```

### Considerações
- JWT usa apenas HMAC-SHA256 (stdlib-only, sem dependência de libs JWT)
- Sem suporte a RSA/ECDSA (manteria zero-deps; usuários que precisam podem usar lib externa)
- Claims são armazenadas no context via `context.WithValue`

---

## 3. `pkg/ratelimit` — Rate Limiting

**Status**: Inexistente | **Prioridade**: CRÍTICA | **Dependências**: Nenhuma (stdlib-only)

### Motivação
APIs sem rate limiting são vulneráveis a abuso, DDoS e consumo excessivo de recursos. Deveria ser ativo por padrão.

### Design

```go
// Rate limit global: 100 req/s por IP
rl := ratelimit.New(ratelimit.Config{
    Rate:  100,
    Per:   time.Second,
    KeyFunc: ratelimit.IPKey, // extrai chave do request (IP, user, endpoint...)
})
r.Use(rl.Middleware)

// Rate limit por rota: 5 req/s para POST
r.Group("/api", rl.Middleware).POST("/users", createUser)

// Com Redis (backend pluggable)
rl := ratelimit.New(ratelimit.Config{
    Backend: ratelimit.NewRedisBackend(redisClient),
    Rate:    1000,
    Per:     time.Minute,
})
```

### API planejada

```go
package ratelimit

type Config struct {
    Rate    int           // número de requisições permitidas
    Per     time.Duration // janela de tempo
    Store   store.KV      // storage (in-memory default) — usa pkg/store
    KeyFunc func(*http.Request) string // como identificar o cliente
}

// Backend esperado: store.KV. O RateLimiter usa Get/Set com TTL.
// Backends concretos: store.NewMemoryStore() built-in.
// Backend Redis fica no template de integração (ginger add redis).

// Key extractors
func IPKey(r *http.Request) string     // por IP
func UserKey(r *http.Request) string   // por user ID (requer auth antes)
func EndpointKey(r *http.Request) string // por método+path

// Middleware
func New(cfg Config) *RateLimiter
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler
// Retorna 429 com headers: Retry-After, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
```

### Algoritmo
Token bucket puro em `sync.Mutex` sem goroutines de background — recalcula tokens sob demanda usando o timestamp da última requisição.

---

## 4. `pkg/cache` — Cache de Resposta HTTP

**Status**: Inexistente | **Prioridade**: ALTA | **Dependências**: Nenhuma

### Motivação
Endpoints de leitura (GET) frequentemente retornam dados que não mudam com frequência. Cache reduz carga no banco e latência.

### Design

```go
// Cache de 1 minuto para GET /api/users
cache := cache.New(cache.Config{
    TTL:     1 * time.Minute,
    MaxSize: 1000, // máximo de entradas
})
r.Use(cache.Middleware)

// ETag para respostas individuais
r.GET("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    user := getUser(r.PathValue("id"))
    etag := cache.ETag(fmt.Sprintf("%d", user.UpdatedAt.Unix()))
    if cache.CheckETag(w, r, etag) {
        return // 304 Not Modified
    }
    response.OK(w, user)
})
```

### API planejada

```go
package cache

type Config struct {
    TTL     time.Duration
    MaxSize int           // default: 1000 (para MemoryStore)
    Store   store.KV      // storage (in-memory default) — usa pkg/store
    Methods []string      // default: ["GET", "HEAD"]
    KeyFunc func(*http.Request) string // default: method+path+query
}

// Backend: store.KV. Backends concretos em store.NewMemoryStore().
// Middleware
func New(cfg Config) *Cache
func (c *Cache) Middleware(next http.Handler) http.Handler

// ETag helpers (stdlib-only, sem middleware)
func ETag(value string) string           // gera valor de ETag
func CheckETag(w http.ResponseWriter, r *http.Request, etag string) bool // 304 se match
```

---

## 5. `pkg/compress` — Compressão HTTP

**Status**: Inexistente | **Prioridade**: ALTA | **Dependências**: `compress/gzip`

### Motivação
Compressão de resposta reduz largura de banda em 60-80% para JSON. Esperado em qualquer API pública.

### Design

```go
// Ativar compressão para todas as respostas > 1KB
r.Use(compress.Middleware(compress.Config{
    Level:   6,    // nível gzip (1-9)
    MinSize: 1024, // só comprime respostas > 1KB
}))
```

### API planejada

```go
package compress

type Config struct {
    Level   int // 1-9, default: 6
    MinSize int // mínimo de bytes para comprimir, default: 1024
}

func Middleware(cfg Config) middleware.Func
```

### Notas
- Gzip-only (stdlib `compress/gzip`). Sem dependências externas.
- Writer pool para reduzir alocações.
- Respeita header `Accept-Encoding: gzip`.

---

## 6. `pkg/pagination` — Paginação de Respostas

**Status**: Parcial (apenas envelope `response.Page`) | **Prioridade**: ALTA | **Dependências**: `pkg/response`

### Motivação
O CRUD gerado hoje retorna listas completas sem paginação. Para tabelas com milhões de registros, isso é inviável.

### Design

```go
// No handler
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    pg, err := pagination.FromRequest(r, pagination.Config{
        DefaultPerPage: 20,
        MaxPerPage:     100,
    })
    if err != nil {
        router.Error(w, err)
        return
    }
    users, err := h.service.List(r.Context(), pg)
    response.Paginated(w, users, pg.Page, pg.PerPage, users.Total)
}
```

### API planejada

```go
package pagination

type Params struct {
    Page    int // página atual (1-indexed)
    PerPage int // itens por página
    Cursor  string // cursor para cursor-based
    Sort    string // campo de ordenação
    Order   string // "asc" ou "desc"
}

type Result[T any] struct {
    Items      []T
    Total      int
    NextCursor string // para cursor-based
}

type Config struct {
    DefaultPerPage int // default: 20
    MaxPerPage     int // default: 100
    SortFields     []string // campos permitidos para ordenação
}

// Extrai parâmetros de paginação da query string
// ?page=1&per_page=20&sort=name&order=asc&cursor=eyJpZCI6MX0=
func FromRequest(r *http.Request, cfg Config) (*Params, error)

// Helpers para SQL
func (p *Params) Limit() int
func (p *Params) Offset() int
func (p *Params) OrderBy() string  // "name ASC" (sanitizado contra SortFields)
```

---

## 7. `pkg/metrics` — Métricas HTTP

**Status**: Inexistente | **Prioridade**: MÉDIA | **Dependências**: `pkg/logger`, OpenTelemetry

### Motivação
O Ginger já tem tracing e logging, mas falta a terceira perna da observabilidade: métricas. Com OTel metrics, podemos exportar para Prometheus automaticamente.

### Design

```go
// Em app.New(), ativado por padrão
r.Use(metrics.Middleware(metrics.Config{
    Namespace: cfg.App.Name,
    Subsystem: "http",
}))

// Métricas expostas automaticamente:
// GET /metrics → Prometheus text format
// Métricas coletadas: http_requests_total, http_request_duration_seconds,
//   http_requests_in_flight, http_response_size_bytes
```

### API planejada

```go
package metrics

type Config struct {
    Namespace string
    Subsystem string
    Buckets   []float64 // default: prometheus.DefBuckets
}

func Middleware(cfg Config) middleware.Func

// Handler expõe /metrics em formato Prometheus
func Handler() http.Handler

// Métricas customizáveis expostas para o usuário
type Counter interface {
    Inc()
    Add(float64)
}

type Histogram interface {
    Observe(float64)
}

func NewCounter(name, help string) Counter
func NewHistogram(name, help string, buckets []float64) Histogram
```

### Notas
- Usa **apenas OpenTelemetry metrics** (já é dependência do framework via `pkg/telemetry`)
- Sem dependência adicional do `prometheus/client_golang`
- Se OTel Prometheus exporter estiver configurado, métricas aparecem em `/metrics` automaticamente
- Se OTel não estiver configurado, middleware opera em modo no-op (zero custo)
- Métricas exportadas: `http.server.request_count`, `http.server.duration`, `http.server.request_body_size`, `http.server.response_body_size`

---

## 8. `pkg/resilience` — Circuit Breaker + Retry

**Status**: Inexistente | **Prioridade**: MÉDIA | **Dependências**: Nenhuma

### Motivação
Chamadas a serviços externos (APIs, bancos, caches) precisam de proteção contra falhas em cascata.

### Design

```go
// Circuit breaker para chamadas HTTP externas
cb := resilience.NewCircuitBreaker("user-api", resilience.CBConfig{
    MaxFailures:  5,
    ResetTimeout: 30 * time.Second,
    HalfOpenMax:  2,
})

err := cb.Do(ctx, func() error {
    resp, err := http.Get("https://api.externa.com/users")
    // processa resposta
    return err
})

// Retry com backoff
err := resilience.Retry(ctx, resilience.RetryConfig{
    MaxAttempts: 3,
    Backoff:     resilience.ExponentialBackoff(time.Second, 30*time.Second),
}, func() error {
    return db.Ping()
})
```

### API planejada

```go
package resilience

// Circuit Breaker (3 estados: closed, open, half-open)
type CBConfig struct {
    MaxFailures  int
    ResetTimeout time.Duration // tempo até tentar half-open
    HalfOpenMax  int           // máx requisições em half-open
}

type CircuitBreaker struct { ... }
func NewCircuitBreaker(name string, cfg CBConfig) *CircuitBreaker
func (cb *CircuitBreaker) Do(ctx context.Context, fn func() error) error
func (cb *CircuitBreaker) State() string // "closed", "open", "half-open"

// Retry
type RetryConfig struct {
    MaxAttempts int
    Backoff     BackoffStrategy
    RetryOn     func(error) bool // quais erros disparam retry
}

type BackoffStrategy func(attempt int) time.Duration
func ConstantBackoff(d time.Duration) BackoffStrategy
func ExponentialBackoff(base, max time.Duration) BackoffStrategy
func LinearBackoff(d time.Duration) BackoffStrategy

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error
```

---

## 9. `pkg/idempotency` — Idempotência de Requisições

**Status**: Inexistente | **Prioridade**: MÉDIA | **Dependências**: Nenhuma

### Motivação
APIs de pagamento, criação de recursos e operações críticas precisam garantir que uma requisição repetida não cause efeito duplicado.

### Design

```go
idem := idempotency.New(idempotency.Config{
    HeaderName: "Idempotency-Key",
    TTL:        24 * time.Hour,
    Backend:    idempotency.NewMemoryBackend(),
})

// Aplica só em POST, PUT, PATCH
r.Group("/api/v1", idem.Middleware).POST("/payments", createPayment)
```

### API planejada

```go
package idempotency

type Config struct {
    HeaderName string        // default: "Idempotency-Key"
    TTL        time.Duration // quanto tempo guardar a resposta
    Store      store.KV      // storage (in-memory default) — usa pkg/store
}

// A serialização de CachedResponse para []byte é interna ao pacote.
// Backend concreto em store.NewMemoryStore().

type CachedResponse struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
}

func New(cfg Config) *Idempotency
func (i *Idempotency) Middleware(next http.Handler) http.Handler
// Se a chave já existe, retorna a resposta cached (mesmo status e body)
// Se a chave não existe, executa o handler e armazena a resposta
```
