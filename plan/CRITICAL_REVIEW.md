# Critical Review — Plano de Evolução Ginger

> Revisão de ambiguidades, riscos de regressão e features questionáveis.
> Data: 2026-06-28

---

## 🔴 Problemas Encontrados (15 issues)

### 1. Três interfaces `Backend` redundantes

`pkg/cache.Backend`, `pkg/idempotency.Backend` e `pkg/ratelimit.Backend` são essencialmente o mesmo padrão (KV store com TTL). Três interfaces diferentes para o mesmo conceito viola o princípio de simplicidade.

**Decisão**: Unificar em uma interface comum `pkg/store.KV` que cache, idempotency e ratelimit consumam. Backends concretos (memory, Redis) implementam uma interface só.

```go
// pkg/store/kv.go — usada por cache, idempotency, ratelimit
type KV interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

---

### 2. Brotli viola filosofia stdlib-first

`pkg/compress` propõe suporte a Brotli que depende de pacote externo. O runtime check "só ativa se disponível" é frágil e não idiomático.

**Decisão**: Remover Brotli do escopo. `pkg/compress` será **gzip-only** (stdlib `compress/gzip`). Se o usuário quiser Brotli, que use middleware externo.

---

### 3. Dupla dependência de métricas (Prometheus + OTel)

O plano propõe `prometheus/client_golang` como dependência + OTel metrics. Isso é redundante — OTel já exporta para Prometheus via OTLP.

**Decisão**: `pkg/metrics` usa **apenas OTel metrics** (já é dependência do framework). Sem dependência adicional. O handler `/metrics` expõe via OTel Prometheus exporter se disponível, caso contrário, stdout.

---

### 4. `pkg/tenant.ScopeDB` — magia perigosa

A proposta de `ScopeDB` que "automaticamente adiciona WHERE tenant_id = ?" requer parsing de SQL, é frágil com queries complexas (JOINs, subqueries, UNIONs) e viola o princípio "explícito, sem magia".

**Decisão**: `pkg/tenant` fornece apenas `FromContext(ctx)`. O desenvolvedor escreve `WHERE tenant_id = $1` explicitamente. Sem `ScopeDB`.

---

### 5. `pkg/ratelimit.NewRedisBackend` no pacote core

O constructor `NewRedisBackend(client *redis.Client)` no pacote `ratelimit` cria uma dependência implícita de Redis no core do framework.

**Decisão**: O pacote `ratelimit` expõe apenas a interface `Backend`. O constructor `NewRedisBackend` vai para o template de integração Redis (`ginger add redis` gera o adapter).

---

### 6. `pkg/auth` com `Optional: true` — semântica ambígua

Quando `Optional: true`, `ClaimsFromContext()` pode retornar `nil`. Isso força nil-checks em todo handler e é propenso a bugs.

**Decisão**: Remover `Optional`. Substituir por dois middlewares separados: `auth.Required()` (falha se token ausente) e `auth.Optional()` (sempre passa, mas claims podem ser vazias com `IsAuthenticated(ctx) bool`).

---

### 7. Integração `router.Decode()` + validação — breaking change

Se `router.Decode()` passar a validar após desserializar, código existente que faz validação manual após `Decode` vai quebrar ou validar duplamente.

**Decisão**: `router.Decode()` **não** valida. Adicionar método separado `router.DecodeAndValidate(r, v) error` que chama `Decode` + `validate.Struct(v)`. Zero breaking change.

---

### 8. Middlewares ativos por padrão — risco de regressão

Se `app.New()` passar a aplicar `validate`, `auth`, `ratelimit`, `compress` por padrão, upgrades quebram apps existentes que não têm esses middlewares configurados.

**Decisão**: Nenhum middleware novo é aplicado por padrão em `app.New()`. Todos são opt-in. A documentação e o scaffold de novos projetos os incluem como recomendados, mas o upgrade é seguro.

---

### 9. `response.Page` → `pagination.Result` — breaking change

A API existente `response.Page[T]` seria substituída por `pagination.Result[T]`.

**Decisão**: Manter `response.Page[T]` como está. `pagination.Result[T]` é um tipo separado usado pelo handler internamente. `response.Paginated()` aceita `Params` e `Result` e converte para `Page`. Zero breaking change.

---

### 10. GraphQL Generator é inviável no escopo atual

Gerar código GraphQL de qualidade requer um codegen complexo (gqlgen, graphql-go). Fazer um gerador "simples" resultaria em código inútil ou exigiria manter um mini-gqlgen.

**Decisão**: Remover GraphQL Generator do plano. Em seu lugar, criar template de integração `ginger add graphql` que gera arquivos base para gqlgen (schema, gqlgen.yml, resolver stub). Mais útil e mais simples.

---

### 11. `pkg/featureflags.RequireFlag` retorna 404

Feature flag desativada retornar 404 (Not Found) é semanticamente errado — o recurso existe, só está desativado.

**Decisão**: Retornar **503 Service Unavailable** com corpo explicativo `{"code":"FEATURE_DISABLED","message":"feature 'X' is not enabled"}`.

---

### 12. `.ginger.yaml` referencia features inexistentes

O config proposto tem `middleware_chain: ["auth", "ratelimit", "logger"]` — mas `auth` e `ratelimit` só existirão após a Fase 1. Se a Fase 2 for implementada antes da Fase 1, o config referencia features que não existem.

**Decisão**: `.ginger.yaml` só referencia features que já existem. A documentação de cada feature da Fase 1 inclui a chave de config correspondente, mas o parser ignora chaves desconhecidas (não quebra).

---

### 13. `pkg/events.Outbox` acoplado a `*sql.DB`

O constructor `NewOutbox(db *sql.DB, ...)` impede uso com MongoDB ou outros bancos.

**Decisão**: Manter o acoplamento com `*sql.DB` para v1. É o caso de uso mais comum. Documentar como limitação conhecida. Na v2, introduzir interface `OutboxStore`.

---

### 14. Ordem de middlewares não documentada

O plano menciona middlewares mas não especifica a ordem recomendada. A ordem importa: `RequestID → Logger → Recover → Compress → Cache → Ratelimit → Auth → Handler`.

**Decisão**: Documentar a ordem recomendada no `README.md` do pacote `middleware` e no guia de arquitetura.

---

### 15. Content-Type negotiation ausente no plano

A discussão original mencionou content negotiation (JSON/XML/Protobuf) mas não aparece em nenhuma fase.

**Decisão**: Adicionar `pkg/negotiation` na Fase 1. É um pacote pequeno: parsing de `Accept` header, helper `Negotiate(r, offers []string) string`. Zero dependências.

---

## ✅ Features validadas (sem problemas)

| Feature | Status |
|---------|--------|
| `pkg/validate` com tags | OK. Bem definido, stdlib-only |
| `pkg/auth` JWT HMAC | OK após correção do Optional |
| `pkg/ratelimit` token bucket | OK após remover Redis constructor do core |
| `pkg/cache` middleware | OK após unificar Backend |
| `pkg/compress` gzip | OK após remover Brotli |
| `pkg/pagination` cursor+offset | OK após não quebrar response.Page |
| `pkg/metrics` OTel | OK após remover dependência do Prometheus |
| `pkg/resilience` circuit breaker | OK. Bem isolado |
| `pkg/idempotency` | OK após unificar Backend |
| `ginger dev` | OK. Bem definido |
| `ginger migrate` | OK. SQL puro, simples |
| CRUD Generator v2 | OK. Depende de validate + pagination |
| Test Generator v2 | OK |
| `.ginger.yaml` | OK após correção de features referenciadas |
| `pkg/tenant` | OK após remover ScopeDB |
| `pkg/audit` | OK |
| `pkg/featureflags` | OK após corrigir HTTP status |
| `pkg/events` | OK após documentar limitação do Outbox |
| `pkg/scheduler` | OK |

---

## 🔄 Mudanças no plano original

| Mudança | Antes | Depois |
|---------|-------|--------|
| Backend KV | 3 interfaces separadas | 1 interface `pkg/store.KV` |
| Brotli | Suporte opcional | Removido |
| Métricas | Prometheus + OTel | Apenas OTel |
| ScopeDB | SQL mágico | Removido, só `FromContext()` |
| Redis Backend | No pacote core | No template de integração |
| Auth Optional | Flag booleana | Dois middlewares separados |
| Decode + Validate | Comportamento alterado | Novo método `DecodeAndValidate` |
| Middlewares default | Ativos por padrão | Opt-in |
| GraphQL | Generator complexo | Template de integração simples |
| Feature flag 404 | HTTP 404 | HTTP 503 |
| Content negotiation | Ausente | `pkg/negotiation` na Fase 1 |

---

## 📊 Novo total de pacotes

| Fase | Pacotes | Antes | Depois |
|------|---------|-------|--------|
| Fase 1 | `validate`, `auth`, `ratelimit`, `store`, `cache`, `compress`, `pagination`, `metrics`, `negotiation`, `resilience`, `idempotency` | 9 | **11** |
| Fase 3 | `tenant`, `audit`, `featureflags`, `events`, `scheduler` | 5 | **5** |
| **Total novos** | | 14 | **16** |
| **Total framework** | (13 existentes + novos) | 22 | **29** |

---

## 🛡️ Garantia anti-regressão

1. **Nenhum middleware novo é ativo por padrão** — `app.New()` permanece igual
2. **`router.Decode()` permanece inalterado** — novo método `DecodeAndValidate` separado
3. **`response.Page` não é removido** — `pagination.Result` é aditivo
4. **Todos os novos pacotes são opt-in** — não quebram código existente
5. **Pacotes existentes não são modificados** — apenas estendidos (ex: `router.DecodeAndValidate`)
6. **Test suite existente continua passando** — novos testes são aditivos
