# ADR-0006 — Catálogo único de integrações com fail-closed

**Status**: Aceito · **Data**: 2026-08-31 (Fase 2, GIN-004)

## Contexto
Integrações (22) e capabilities (11) eram listas separadas; `checkCapabilityConstraints` retornava nil para não-catalogadas — bypass silencioso de validações.

## Decisão
`IntegrationSpec` declarativa em cada integração do registry (description/minGo/projectTypes); o capability registry **deriva** do catálogo (drift test). Constraints: fail-closed — nome não catalogado é recusado. Restrições existentes preservadas sem breaking (warnings-first fica para a matriz-alvo futura).

## Consequências
+ Impossível contornar validações por ausência do catálogo.
+ Help/docs podem derivar da mesma fonte.
− Adicionar integração exige preencher a Spec completa (consciente por construção).
