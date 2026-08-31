# ADR-0004 — OpenTelemetry em submódulo opcional

**Status**: Aceito · **Data**: 2026-06-28 (DWYT), ratificado 2026-08-31

## Contexto
Telemetria é opcional para usuários; otel arrasta dependências pesadas (grpc etc.).

## Decisão
Raiz: go 1.22 + yaml apenas. `pkg/telemetry` é submódulo próprio (go 1.25) com ciclo de release/teste próprio (govulncheck blocking no CI desde a Fase 1). `pkg/logger` é provider-neutral (`WithTraceContext`); capability `otel` bloqueia projetos com go < 1.25.

## Consequências
+ Usuários sem telemetria não pagam as dependências nem as vulnerabilidades.
− Trade-off previsto e confirmado: o submódulo precisa de manutenção própria (17 vulns acumuladas antes da Fase 1 — corrigidas: otel 1.43.0 / grpc 1.82.1).
