# Ginger Evolution Plan

> Plano de evolução para tornar o Ginger o melhor framework Go da história.
> Última atualização: 2026-06-29

---

## Visão Geral

O Ginger é um framework web Go minimalista + CLI de scaffolding que hoje entrega:

- **Framework (pkg/)**: 13 pacotes — router, middleware, config, logger, errors, health, database, telemetry, SSE, WebSocket, test helpers
- **CLI (internal/)**: 12 pacotes — project, plan, region, capability, scaffold, generator, integrations, doctor, docsgen, manifest, buildinfo, cli

## Estrutura do Plano

| Documento | Conteúdo |
|-----------|----------|
| [roadmap.md](roadmap.md) | Timeline e milestones |
| [phase-1-core-framework.md](phase-1-core-framework.md) | Fase 1: 11 novos pacotes core do framework |
| [phase-2-developer-experience.md](phase-2-developer-experience.md) | Fase 2: 5 features de CLI e tooling |
| [phase-3-advanced-features.md](phase-3-advanced-features.md) | Fase 3: 5 pacotes enterprise |
| [CRITICAL_REVIEW.md](CRITICAL_REVIEW.md) | Revisão crítica: 15 problemas encontrados e corrigidos |
| [doc-validation.md](doc-validation.md) | Validação documental (35 issues) |
| [doc-revalidation.md](doc-revalidation.md) | Plano de revalidação, consolidação e limpeza |

### Decisões-chave da revisão crítica (CRITICAL_REVIEW.md)
1. Backend KV unificado em `pkg/store.KV`
2. Brotli removido (viola stdlib-first)
3. Prometheus removido (redundante com OTel)
4. `ScopeDB` removido (magia perigosa)
5. GraphQL Generator → template de integração
6. Nenhum middleware novo ativo por padrão (anti-regressão)

## Fases em Resumo

### Fase 1 — Fundação do Framework Core (11 pacotes novos)
`store` → `validate` → `negotiation` → `auth` → `ratelimit` → `cache` → `compress` → `pagination` → `metrics` → `resilience` → `idempotency`

### Fase 2 — Developer Experience (5 features)
Hot reload → Migration tooling → CRUD Generator v2 → Test Generator v2 → Custom templates

### Fase 3 — Funcionalidades Avançadas (5 pacotes)
Multi-tenancy → Audit logging → Feature flags → Event bus → Scheduler

## Princípios Mantidos

1. **Stdlib-first**: zero dependências externas sempre que possível
2. **Tipado e explícito**: sem magia, sem reflection desnecessária
3. **Graceful degradation**: tudo funciona com defaults sensatos
4. **Twelve-factor**: config via env-vars, stateless, disposable
5. **Observável por padrão**: logs, traces e métricas integrados
6. **Bilíngue**: documentação e templates PT-BR + EN
