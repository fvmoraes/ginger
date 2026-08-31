# ADR-0008 — Release com commit de versão + SLSA L3

**Status**: Aceito · **Data**: 2026-08-31 (Fases 1 e 5, GIN-001/017/018)

## Contexto
A tag apontava para commit sem os metadados da própria versão (v1.4.4 ↔ version.txt 1.4.3); sem proveniência, SBOM ou hardening.

## Decisão
Release em 3 jobs: **prepare** (computa versão, atualiza version files, commita com `[skip ci]` + actor guard, cria tag com proteção de duplicata), **build** (SLSA generic_slsa3 v2.0.0 — proveniência L3 dos binários), **release** (assets + body com changelog + SBOM SPDX). Actions pinadas por SHA; concurrency group; checksums por release; instalador valida checksum.

## Consequências
+ Tag sempre coerente; proveniência verificável; supply chain auditável (Scorecard).
− Release depende do reusable workflow SLSA (primeiro run valida em produção; rollback = revert do release.yml).
