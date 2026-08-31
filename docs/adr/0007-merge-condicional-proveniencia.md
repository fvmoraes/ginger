# ADR-0007 — Merge condicional por hash de proveniência

**Status**: Aceito · **Data**: 2026-08-31 (Fase 1, GIN-002)

## Contexto
`add` reescrevia o docker-compose gerenciado destruíndo personalização (anchors, comentários, x- fields, networks). Trocar sempre por patch quebraria scaffolds novos (1º add deixaria de atualizar o compose).

## Decisão
O manifest registra `generated_hash` (SHA-256 do conteúdo gerado). Merge condicional: arquivo intacto desde a geração → merge direto (scaffold novo funciona); modificado pelo usuário ou hash ausente → **patch revisável** em `.ginger/patches/` (data-safe). `composeBuild` aceita forma abreviada (`build: .`).

## Consequências
+ Nenhum conteúdo do usuário é destruído; scaffold novo continua "só funciona".
− Manifests pré-Fase 1 (sem hash) recebem patch por segurança (UX levemente diferente até o próximo apply).
