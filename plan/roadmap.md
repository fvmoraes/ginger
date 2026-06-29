# Roadmap

> Timeline e milestones do plano de evolução.

---

## Fase 1: Fundação do Framework Core

**Objetivo**: Preencher os gaps críticos de segurança e produtividade no framework.
**Duração estimada**: 6-8 semanas
**Target version**: v1.4.0 → v1.5.0

### Milestones

| Semana | Milestone | Entregáveis |
|--------|-----------|-------------|
| 1 | **M1: Store + Validate** | `pkg/store` (KV interface), `pkg/validate` (struct tags), `pkg/negotiation` (content-type) |
| 2-3 | **M2: Auth + Rate Limit** | `pkg/auth` (JWT + API Key + RBAC), `pkg/ratelimit` (token bucket) |
| 4 | **M3: Cache + Compress** | `pkg/cache` (response cache + ETag), `pkg/compress` (gzip-only) |
| 5-6 | **M4: Pagination + Metrics** | `pkg/pagination` (cursor + offset), `pkg/metrics` (OTel metrics) |
| 7-8 | **M5: Resilience + Idempotency** | `pkg/resilience` (circuit breaker + retry), `pkg/idempotency` |

### Ordem de implementação

```
store ──► validate ──► negotiation ──► auth ──► ratelimit
                                                    │
cache ◄──────────────────────────────────────────────┘
  │
  └──► compress ──► pagination ──► metrics ──► resilience ──► idempotency
```

> **Nota pós-revisão**: Backend KV unificado em `pkg/store`. `auth`, `ratelimit`, `cache`, `idempotency` compartilham a mesma interface de storage.

---

## Fase 2: Developer Experience

**Objetivo**: Acelerar o ciclo de desenvolvimento e tornar o Ginger auto-suficiente.
**Duração estimada**: 4 semanas
**Target version**: v1.6.0

### Milestones

| Semana | Milestone | Entregáveis |
|--------|-----------|-------------|
| 1-2 | **M6: Hot Reload + Migrations** | `ginger dev` (file watcher + rebuild), `ginger migrate` (create/up/down) |
| 2-3 | **M7: CRUD Generator v2** | Templates com validação, paginação, filtros, ordenação automática |
| 3-4 | **M8: Test Generator v2** | `ginger generate test --all`, benchmarks, fuzz tests |
| 4 | **M9: Templates customizáveis** | `.ginger.yaml` para configuração de templates por org/usuário |

---

## Fase 3: Funcionalidades Avançadas

**Objetivo**: Posicionar o Ginger como escolha para sistemas complexos e enterprise.
**Duração estimada**: 4-6 semanas
**Target version**: v2.0.0

### Milestones

| Semana | Milestone | Entregáveis |
|--------|-----------|-------------|
| 1-2 | **M10: Multi-tenancy + Audit** | `pkg/tenant`, `pkg/audit` |
| 2-3 | **M11: Feature Flags** | `pkg/featureflags` (env, Redis, LaunchDarkly-compat) |
| 3-4 | **M12: Event Bus + Scheduler** | `pkg/events` (local + outbox), `pkg/scheduler` (cron) |
| 4-6 | **M13: GraphQL Integration** | `ginger add graphql` (template de integração gqlgen, não gerador complexo) |

---

## Visão de Longo Prazo (v2.1+)

| Feature | Descrição |
|---------|-----------|
| **gRPC-first** | Suporte nativo a gRPC com codegen do protobuf |
| **Event Sourcing** | Templates para event sourcing + CQRS |
| **Plugin System** | Sistema de plugins para extender a CLI |
| **Admin Dashboard** | UI web para monitorar apps Ginger em dev |
| **WASM Edge** | Suporte a compilação para WASM/WASI |

---

## Métricas de Sucesso

| Métrica | Baseline (v1.3.6) | Target (v2.0) |
|---------|-------------------|---------------|
| Pacotes no framework | 13 | **29** |
| Tempo scaffold→deploy | ~5 min | ~3 min |
| Linhas de código do framework | ~1.345 | ~4.000 |
| Cobertura de testes do framework | TBD | >85% |
| Integrações disponíveis | 23 | **28** |
| Dependências externas | 1 (yaml) | 1 (yaml) |
| Middlewares built-in | 4 | **15** |
