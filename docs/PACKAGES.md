# Guia Completo de Pacotes

[← Voltar ao README](../README.md)

## Índice

- [pkg/app](#pkgapp) — Application bootstrap
- [pkg/router](#pkgrouter) — HTTP routing
- [pkg/middleware](#pkgmiddleware) — HTTP middlewares
- [pkg/errors](#pkgerrors) — Typed errors
- [pkg/response](#pkgresponse) — JSON envelopes
- [pkg/config](#pkgconfig) — Configuration
- [pkg/logger](#pkglogger) — Structured logging
- [pkg/database](#pkgdatabase) — Database connection
- [pkg/health](#pkghealth) — Health checks
- [pkg/telemetry](#pkgtelemetry) — OpenTelemetry
- [pkg/sse](#pkgsse) — Server-Sent Events
- [pkg/ws](#pkgws) — WebSocket
- [pkg/testhelper](#pkgtesthelper) — Test utilities

---

## pkg/app

**Propósito:** Bootstrap da aplicação com lifecycle management

### API Completa

```go
type App struct {
    Router *router.Router
    Config *config.Config
    // campos privados: server, logger, stopFns
}

func New(cfg *config.Config) *App
func (a *App) OnStop(fn func(context.Context) error)
func (a *App) Run() error
```

### Uso Básico

```go
cfg, _ := config.Load("configs/app.yaml")
app := gingerapp.New(cfg)

// Registrar rotas
app.Router.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
    router.JSON(w, http.StatusOK, map[string]string{"message": "pong"})
})

// Registrar cleanup hooks
app.OnStop(func(ctx context.Context) error {
    return db.Close()
})

// Iniciar servidor (bloqueia até SIGINT/SIGTERM)
app.Run()
```

### Lifecycle Hooks

`OnStop` permite registrar funções de cleanup executadas no shutdown:

```go
// Fechar conexão de banco
app.OnStop(func(ctx context.Context) error {
    return db.Close()
})

// Flush telemetry
app.OnStop(func(ctx context.Context) error {
    return telemetryProvider.Shutdown(ctx)
})

// Fechar conexão Redis
app.OnStop(func(ctx context.Context) error {
    return redisClient.Close()
})
```

**Ordem de execução:** LIFO (last-in, first-out) — última registrada executa primeiro.

### Graceful Shutdown

O `App.Run()` captura `SIGINT` e `SIGTERM` automaticamente:

1. Sinal recebido
2. Para de aceitar novas conexões
3. Aguarda requisições ativas terminarem (timeout: `cfg.HTTP.ShutdownTimeout`)
4. Executa todos os `OnStop` hooks
5. Encerra

**Timeout padrão:** 30 segundos (configurável via `http.shutdown_timeout` no YAML)

### Configuração HTTP

```yaml
http:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30  # segundos
```

```go
// Acesso programático
app.Config.HTTP.Port           // 8080
app.Config.HTTP.ShutdownTimeout // 30
```

---

## pkg/router

**Propósito:** Wrapper sobre `net/http` ServeMux com helpers

### API Completa

```go
type Router struct { /* ... */ }

// Criação
func New() *Router

// Middlewares
func (r *Router) Use(mw ...middleware.Func)

// Grupos
func (r *Router) Group(prefix string, mw ...middleware.Func) *Router

// Rotas
func (r *Router) Handle(method, pattern string, h http.HandlerFunc)
func (r *Router) GET(pattern string, h http.HandlerFunc)
func (r *Router) POST(pattern string, h http.HandlerFunc)
func (r *Router) PUT(pattern string, h http.HandlerFunc)
func (r *Router) PATCH(pattern string, h http.HandlerFunc)
func (r *Router) DELETE(pattern string, h http.HandlerFunc)
func (r *Router) HandleRaw(pattern string, h http.Handler)

// Helpers
func JSON(w http.ResponseWriter, status int, v any)
func Error(w http.ResponseWriter, err error)
func Decode(r *http.Request, v any) error
```

### Registro de Rotas

```go
r := router.New()

// Rota simples
r.GET("/ping", pingHandler)

// Rota com path param (Go 1.22+)
r.GET("/users/{id}", getUserHandler)

// Múltiplos métodos
r.POST("/users", createUserHandler)
r.PUT("/users/{id}", updateUserHandler)
r.DELETE("/users/{id}", deleteUserHandler)
```

### Grupos de Rotas

```go
// API v1
v1 := r.Group("/api/v1")
v1.GET("/users", listUsers)
v1.POST("/users", createUser)

// API v2 com middleware adicional da aplicação
v2 := r.Group("/api/v2", rateLimit())
v2.GET("/users", listUsersV2)

// Admin com autenticação da aplicação
admin := r.Group("/admin", requireAuth())
admin.GET("/stats", getStats)
admin.POST("/users/{id}/ban", banUser)
```

### Path Parameters

Go 1.22+ suporta path params nativamente:

```go
r.GET("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
})

// Wildcard
r.GET("/files/{path...}", func(w http.ResponseWriter, r *http.Request) {
    path := r.PathValue("path")  // captura "a/b/c" de /files/a/b/c
    // ...
})
```

### JSON Helpers

#### router.JSON

Serializa qualquer valor para JSON:

```go
router.JSON(w, http.StatusOK, map[string]string{
    "message": "success",
})

router.JSON(w, http.StatusOK, user)

router.JSON(w, http.StatusOK, []User{user1, user2})
```

#### router.Error

Converte erros para JSON padronizado:

```go
// AppError → usa code e status próprios
err := apperrors.NotFound("user not found")
router.Error(w, err)
// → 404 {"code":"NOT_FOUND","message":"user not found"}

// Erro genérico → 500 Internal
err := errors.New("database connection failed")
router.Error(w, err)
// → 500 {"code":"INTERNAL","message":"internal server error"}
```

#### router.Decode

Decodifica JSON do body (limite: 1 MiB):

```go
var input CreateUserInput
if err := router.Decode(r, &input); err != nil {
    router.Error(w, err)
    return
}
```

### HandleRaw

Registra handler sem middlewares (útil para `/health`):

```go
healthHandler := health.New()
r.HandleRaw("/health", healthHandler)
```

---

## pkg/middleware

**Propósito:** Middlewares HTTP reutilizáveis

### API Completa

```go
type Func func(http.Handler) http.Handler

// Composição
func Chain(middlewares ...Func) Func

// Built-in middlewares
func Logger(log *logger.Logger) Func
func Recover(log *logger.Logger) Func
func RequestID() Func
func CORS(cfg ...CORSConfig) Func

// Context helpers
func RequestIDFromContext(ctx context.Context) string
```

### Logger

Loga cada requisição com método, path, status, duração:

```go
app.Router.Use(middleware.Logger(log))
```

**Output:**
```json
{
  "level": "info",
  "msg": "request",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "duration": "12.5ms",
  "remote_addr": "192.168.1.1:54321"
}
```

### Recover

Captura panics e retorna 500 JSON:

```go
app.Router.Use(middleware.Recover(log))
```

**Comportamento:**
- Panic capturado
- Stack trace logado
- Cliente recebe: `{"code":"INTERNAL","message":"internal server error"}`
- Servidor continua rodando

### RequestID

Injeta `X-Request-ID` no contexto e headers:

```go
app.Router.Use(middleware.RequestID())

// No handler
func myHandler(w http.ResponseWriter, r *http.Request) {
    reqID := middleware.RequestIDFromContext(r.Context())
    log.Info("processing", "request_id", reqID)
}
```

**Comportamento:**
- Se cliente envia `X-Request-ID` → usa o valor
- Caso contrário → gera ID aleatório (16 hex chars)
- Adiciona ao response header `X-Request-ID`

### CORS

Configuração flexível de CORS:

```go
// Allow-all (padrão)
app.Router.Use(middleware.CORS())

// Configuração customizada
app.Router.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
    AllowCredentials: true,
    MaxAge:           86400,  // 24 horas
}))
```

**Campos:**

| Campo | Tipo | Padrão | Descrição |
|-------|------|--------|-----------|
| `AllowedOrigins` | `[]string` | `["*"]` | Origens permitidas |
| `AllowedMethods` | `[]string` | `["GET","POST","PUT","PATCH","DELETE","OPTIONS"]` | Métodos HTTP |
| `AllowedHeaders` | `[]string` | `["Content-Type","Authorization","X-Request-ID"]` | Headers permitidos |
| `AllowCredentials` | `bool` | `false` | Permite cookies/auth |
| `MaxAge` | `int` | `0` | Cache de preflight (segundos) |

**Nota:** `AllowCredentials: true` não pode ser usado com `AllowedOrigins: ["*"]`

### Chain

Compõe múltiplos middlewares:

```go
app.Router.Use(middleware.Chain(
    middleware.Logger(log),
    middleware.RequestID(),
    middleware.Recover(log),
    middleware.CORS(),
))
```

**Ordem de execução:** esquerda → direita (Logger executa primeiro)

### Middleware Customizado

```go
func RateLimit() middleware.Func {
    limiter := rate.NewLimiter(10, 100)  // 10 req/s, burst 100
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                router.Error(w, apperrors.New(
                    apperrors.CodeTooManyRequests,
                    "rate limit exceeded",
                    nil,
                ))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## pkg/errors

**Propósito:** Erros tipados com códigos HTTP

### API Completa

```go
type Code string

const (
    CodeBadRequest     Code = "BAD_REQUEST"
    CodeUnauthorized   Code = "UNAUTHORIZED"
    CodeForbidden      Code = "FORBIDDEN"
    CodeNotFound       Code = "NOT_FOUND"
    CodeConflict       Code = "CONFLICT"
    CodeInternal       Code = "INTERNAL"
    CodeTooManyRequests Code = "TOO_MANY_REQUESTS"
)

type AppError struct {
    Code    Code   `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

// Construtores
func New(code Code, message string, err error) *AppError
func BadRequest(message string) *AppError
func Unauthorized(message string) *AppError
func Forbidden(message string) *AppError
func NotFound(message string) *AppError
func Conflict(message string) *AppError
func Internal(err error) *AppError
func TooManyRequests(message string) *AppError

// Métodos
func (e *AppError) Error() string
func (e *AppError) Unwrap() error
func (e *AppError) Is(target error) bool
func (e *AppError) HTTPStatus() int

// Helpers
func As(err error) (*AppError, bool)
```

### Uso Básico

```go
// Service layer
func (s *UserService) GetByID(ctx context.Context, id int) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, apperrors.Internal(err)
    }
    if user == nil {
        return nil, apperrors.NotFound("user not found")
    }
    return user, nil
}

// Handler layer
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user, err := h.service.GetByID(r.Context(), id)
    if err != nil {
        router.Error(w, err)  // converte automaticamente
        return
    }
    response.OK(w, user)
}
```

### Mapeamento Code → HTTP Status

| Code | HTTP Status | Uso |
|------|-------------|-----|
| `BAD_REQUEST` | 400 | Input inválido |
| `UNAUTHORIZED` | 401 | Não autenticado |
| `FORBIDDEN` | 403 | Sem permissão |
| `NOT_FOUND` | 404 | Recurso não existe |
| `CONFLICT` | 409 | Conflito (ex: email duplicado) |
| `TOO_MANY_REQUESTS` | 429 | Rate limit |
| `INTERNAL` | 500 | Erro interno |

### Error Wrapping

```go
// Preserva a cadeia de erros
if err := repo.Create(ctx, user); err != nil {
    return apperrors.Internal(err)
}

// Unwrap funciona
var appErr *apperrors.AppError
if errors.As(err, &appErr) {
    fmt.Println(appErr.Code)
}

// Is funciona
if errors.Is(err, sql.ErrNoRows) {
    // ...
}
```

### Erros Customizados

```go
const CodePaymentRequired Code = "PAYMENT_REQUIRED"

func PaymentRequired(message string) *AppError {
    return &AppError{
        Code:    CodePaymentRequired,
        Message: message,
    }
}

// Adicionar mapeamento HTTP
func (e *AppError) HTTPStatus() int {
    switch e.Code {
    case CodePaymentRequired:
        return 402
    // ... outros cases
    }
}
```

---

## pkg/response

**Propósito:** Envelopes JSON padronizados

### API Completa

```go
// Tipos
type Envelope[T any] struct {
    Data T     `json:"data"`
    Meta *Meta `json:"meta,omitempty"`
}

type Page[T any] struct {
    Data       []T        `json:"data"`
    Pagination Pagination `json:"pagination"`
}

type Pagination struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

// Funções
func OK[T any](w http.ResponseWriter, data T)
func Created[T any](w http.ResponseWriter, data T)
func Paginated[T any](w http.ResponseWriter, data []T, page, perPage, total int)
func NoContent(w http.ResponseWriter)
```

### OK — Resposta Simples

```go
user := &User{ID: 1, Name: "Alice"}
response.OK(w, user)
```

**Output:**
```json
{
  "data": {
    "id": 1,
    "name": "Alice"
  }
}
```

### Created — Recurso Criado

```go
user := &User{ID: 1, Name: "Alice"}
response.Created(w, user)
```

**Output:** 201 Created
```json
{
  "data": {
    "id": 1,
    "name": "Alice"
  }
}
```

### Paginated — Lista Paginada

```go
users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
response.Paginated(w, users, 1, 20, 42)
```

**Output:**
```json
{
  "data": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"}
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 42,
    "total_pages": 3
  }
}
```

### NoContent — Sem Body

```go
response.NoContent(w)
```

**Output:** 204 No Content (sem body)

### Frontend Integration

```typescript
// React/TypeScript
interface Envelope<T> {
  data: T;
  meta?: { request_id?: string; version?: string };
}

interface Page<T> {
  data: T[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

// Fetch user
const res = await fetch('/api/v1/users/1');
const envelope: Envelope<User> = await res.json();
const user = envelope.data;

// Fetch paginated list
const res = await fetch('/api/v1/users?page=1&per_page=20');
const page: Page<User> = await res.json();
const users = page.data;
const totalPages = page.pagination.total_pages;
```

---

## pkg/sse

**Propósito:** Server-Sent Events para streaming unidirecional

### API Completa

```go
type Event struct {
    ID    string  // opcional: permite reconnect
    Type  string  // opcional: default "message"
    Data  any     // payload (auto-JSON se não for string)
    Retry int     // opcional: ms para reconnect
}

type Stream struct { /* ... */ }

func New(w http.ResponseWriter) (*Stream, error)
func (s *Stream) Send(e Event) error
func (s *Stream) SendComment(comment string)
```

### Exemplo Completo

```go
func streamHandler(w http.ResponseWriter, r *http.Request) {
    stream, err := sse.New(w)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-r.Context().Done():
            return
        case t := <-ticker.C:
            stream.Send(sse.Event{
                Type: "tick",
                Data: map[string]string{"time": t.Format(time.RFC3339)},
            })
        }
    }
}
```

### Frontend (JavaScript)

```javascript
const eventSource = new EventSource('/api/v1/stream');

eventSource.addEventListener('tick', (e) => {
  const data = JSON.parse(e.data);
  console.log('Tick:', data.time);
});

eventSource.addEventListener('error', (e) => {
  console.error('SSE error:', e);
  eventSource.close();
});
```

### Casos de Uso

- **Live feeds** — notificações, chat, atividade
- **Progress updates** — upload, processamento batch
- **Real-time dashboards** — métricas, logs
- **Notifications** — alertas, mensagens

### Nginx Configuration

SSE requer buffering desabilitado:

```nginx
location /api/v1/stream {
    proxy_pass http://backend;
    proxy_buffering off;
    proxy_cache off;
    proxy_set_header Connection '';
    proxy_http_version 1.1;
    chunked_transfer_encoding off;
}
```

Ginger adiciona `X-Accel-Buffering: no` automaticamente.

---

## pkg/ws

**Propósito:** WebSocket para comunicação bidirecional

### API Completa

```go
type Message struct {
    Type string `json:"type"`
    Data any    `json:"data,omitempty"`
}

type Conn struct { /* ... */ }

func (c *Conn) Send(v any) error
func (c *Conn) Recv(v any) error
func (c *Conn) Close() error

type Handler func(conn *Conn)

func Handle(w http.ResponseWriter, r *http.Request, fn Handler)
```

### Exemplo Completo

```go
func chatHandler(w http.ResponseWriter, r *http.Request) {
    ws.Handle(w, r, func(conn *ws.Conn) {
        defer conn.Close()
        
        for {
            var msg ws.Message
            if err := conn.Recv(&msg); err != nil {
                return  // client disconnected
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

### Frontend (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'chat',
    data: { message: 'Hello!' }
  }));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log('Received:', msg.type, msg.data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket closed');
};
```

### Broadcast Pattern

```go
type Hub struct {
    clients map[*ws.Conn]bool
    mu      sync.RWMutex
}

func (h *Hub) Register(conn *ws.Conn) {
    h.mu.Lock()
    h.clients[conn] = true
    h.mu.Unlock()
}

func (h *Hub) Unregister(conn *ws.Conn) {
    h.mu.Lock()
    delete(h.clients, conn)
    h.mu.Unlock()
}

func (h *Hub) Broadcast(msg ws.Message) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    for conn := range h.clients {
        conn.Send(msg)  // ignora erros
    }
}

// Handler
func chatHandler(w http.ResponseWriter, r *http.Request) {
    ws.Handle(w, r, func(conn *ws.Conn) {
        hub.Register(conn)
        defer hub.Unregister(conn)
        
        for {
            var msg ws.Message
            if err := conn.Recv(&msg); err != nil {
                return
            }
            hub.Broadcast(msg)
        }
    })
}
```

---

## Próximos Passos

- [🏗️ Arquitetura](./ARCHITECTURE.md) — Estrutura e padrões
- [🔌 Integrações](./INTEGRATIONS.md) — Bancos, cache, mensageria
- [🧪 Testes](./TESTING.md) — Estratégias de teste
- [🚀 Deploy](./DEPLOYMENT.md) — Docker, Kubernetes, CI/CD

[← Voltar ao README](../README.md)
