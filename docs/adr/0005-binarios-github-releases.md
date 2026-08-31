# ADR-0005 — Binários distribuídos só via GitHub Releases (D-09)

**Status**: Aceito · **Data**: 2026-08-31 (decisão do mantenedor)

## Contexto
161 blobs binários (381 MB) rastreados no git: clones pesados, duplicação com o GitHub Releases, risco de binário desatualizado.

## Decisão
Binários de release são distribuídos **exclusivamente** como assets do GitHub Release. `git rm --cached`; `releases/` e `dist/` no `.gitignore`; guard no CI falha se `git ls-files bin/ releases/` retornar algo; workflow compila em `dist/` e nunca commita binários; `scripts/release.sh` aposentado.

## Consequências
+ Clones leves; fonte única de distribuição; checksums + SBOM como assets.
− Quem baixava binário do repositório migra para o GitHub Releases (avisado no CHANGELOG).
