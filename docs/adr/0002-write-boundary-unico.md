# ADR-0002 — Write boundary único via plan-apply

**Status**: Aceito · **Data**: 2026-06-28 (DWYT), concluído 2026-08-31 (Fase 3)

## Contexto
Ferramentas que escrevem em projetos do usuário precisam de garantia contra corrupção e perda de trabalho.

## Decisão
Toda escrita flui por `internal/plan`: contenção na raiz, verificação de symlink ancestors, preflight de hash SHA-256, `O_EXCL` em creates, temp+rename em modifies, ownership em `.ginger/manifest.yaml`. Caminhos legados de mutação direta foram **removidos** (Fase 3, GIN-006/030); conteúdo de proveniência (hash) habilita merges condicionais.

## Consequências
+ Perda de dados impossível por construção (merge condicional + patch revisável).
+ Zero mutação direta fora de plan/scaffold (verificado por grep + testes de caracterização).
− Complexidade do plan core (justificada por 2 auditorias sem regressão).
