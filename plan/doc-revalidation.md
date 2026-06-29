# Documentation Revalidation Plan

> Plano de revalidação completa, consolidação e limpeza de toda a documentação.
> Data: 2026-06-29 — Versão alvo: v1.4.1

---

## 1. Inventário Atual

### 1.1 Todos os arquivos `.md` (50 arquivos, excluindo `.git/`)

| Diretório | Arquivos | Linhas totais | Propósito |
|-----------|----------|---------------|-----------|
| Raiz | `README.md`, `CHANGELOG.md`, `AGENTS.md` | ~1.750 | Face pública, changelog, agent config |
| `docs/` | 13 arquivos | ~7.088 | Documentação do framework |
| `plan/` | 9 arquivos | ~62.034 | Planejamento de evolução |
| `releases/` | 23 arquivos | — | Notas de release históricas |
| `examples/` | 1 arquivo | — | README do exemplo |
| `pkg/` | 1 arquivo | — | README de submódulo |
| `scripts/` | 1 arquivo | — | Documentação de scripts |

### 1.2 Arquivos Duplicados ou Redundantes

| Arquivo | Problema | Ação |
|---------|----------|------|
| `docs/CHANGELOG.md` (203 linhas) | Changelog separado da documentação, preso em v1.0.0. Duplica propósito do `CHANGELOG.md` raiz | **DELETAR** — consolidar entradas relevantes no `CHANGELOG.md` raiz |
| `docs/SUMMARY.md` (272 linhas) | Index redundante — `docs/README.md` já serve como índice completo | **DELETAR** — conteúdo já coberto pelo `docs/README.md` |
| `docs/SAFE_EVOLUTION_PLAN.md` (67 linhas) | Plano tático de implementação, sobrepõe com `plan/README.md` | **MOVER** para `plan/impl-roadmap.md` ou **DELETAR** se já implementado |
| `docs/COPY_PASTE.md` (786 linhas) | Exemplos de código que em grande parte duplicam GETTING_STARTED + QUICK_REFERENCE + PACKAGES | **ANALISAR** — extrair exemplos únicos para os docs relevantes, depois deletar |
| `plan/CRITICAL_REVIEW.md` | Revisão já executada, agora é histórico | **MANTER** como registro histórico, mas marcar como `[RESOLVIDO]` |
| `plan/EXECUTIVE_SUMMARY.md` | Resumo duplica `plan/README.md` | **DELETAR** — consolidar no `plan/README.md` |
| `plan/architecture-gap-analysis.md` | Análise inicial, conteúdo já distribuído nas fases | **MOVER** para apêndice ou **DELETAR** — manter como referência histórica |

### 1.3 Arquivos que Precisam Atualização

| Arquivo | Status | Issues pendentes |
|---------|--------|------------------|
| `CHANGELOG.md` (raiz) | Parcial | Ordem dos links, entry v1.1.1 menciona Go 1.25 |
| `docs/TESTING.md` | Parcial | Não menciona `generate tests --scan`, CI usa Go 1.25 |
| `docs/DEPLOYMENT.md` | Parcial | Tags Docker/actions desatualizadas |
| `docs/COPY_PASTE.md` | Parcial | Sem exemplos de `ginger.yaml`, `--plan`, managed regions |
| `docs/GETTING_STARTED.md` | ✅ Corrigido | — |
| `docs/QUICK_REFERENCE.md` | ✅ Corrigido | — |
| `docs/INTEGRATIONS.md` | ✅ Corrigido | — |
| `docs/ARCHITECTURE.md` | ✅ Corrigido | — |
| `docs/PROJECT_STRUCTURE.md` | ✅ Corrigido | — |
| `docs/PACKAGES.md` | ✅ Corrigido | — |
| `docs/CHANGELOG.md` | ✅ Corrigido | — |
| `README.md` | ✅ Corrigido | — |

---

## 2. Plano de Limpeza

### Fase A — Deletar arquivos redundantes (5 ações)

| # | Arquivo | Ação | Justificativa |
|---|---------|------|---------------|
| A1 | `docs/CHANGELOG.md` | **DELETAR** | Duplica `CHANGELOG.md` raiz. Entradas consolidadas lá. |
| A2 | `docs/SUMMARY.md` | **DELETAR** | Index redundante. `docs/README.md` cobre tudo. |
| A3 | `docs/SAFE_EVOLUTION_PLAN.md` | **DELETAR** | Plano implementado na v1.4.0. Conteúdo migrado para `plan/`. |
| A4 | `plan/EXECUTIVE_SUMMARY.md` | **DELETAR** | Conteúdo consolidado no `plan/README.md`. |
| A5 | `plan/architecture-gap-analysis.md` | **DELETAR** | Análise inicial distribuída nos 3 documentos de fase. |

**Resultado: -5 arquivos, ~2.500 linhas removidas.**

### Fase B — Consolidar e mover (2 ações)

| # | De | Para | Ação |
|---|----|-----|------|
| B1 | `docs/COPY_PASTE.md` | Exemplos únicos → `docs/QUICK_REFERENCE.md` | Extrair exemplos não-duplicados, depois deletar |
| B2 | `plan/CRITICAL_REVIEW.md` | `plan/README.md` (resumo) | Consolidar decisões-chave no README, manter CRITICAL_REVIEW como histórico |

### Fase C — Atualizar pendências (4 ações)

| # | Arquivo | O que atualizar |
|---|---------|-----------------|
| C1 | `CHANGELOG.md` (raiz) | Reordenar links, adicionar nota "Go 1.22+ core" na entry v1.1.1 |
| C2 | `docs/TESTING.md` | Adicionar `gunger generate tests --scan`, ajustar CI para Go 1.22 |
| C3 | `docs/DEPLOYMENT.md` | Atualizar tags: `alpine:3.20`, `actions/checkout@v4`, `docker/setup-buildx-action@v3` |
| C4 | `docs/README.md` | Atualizar index removendo referências a arquivos deletados |

### Fase D — Revalidação cruzada EN ↔ PT (1 ação)

| # | O que validar | Critério |
|---|---------------|----------|
| D1 | `README.md` seções EN vs PT | Cada seção EN deve ter equivalente PT com mesmo conteúdo factual |

---

## 3. Checklist de Revalidação por Documento

### 3.1 `README.md` (raiz)

| # | Check | Critério |
|---|-------|----------|
| R1 | CLI Reference EN | Todos os comandos do `root.go` switch listados |
| R2 | CLI Reference PT | Idêntico ao EN em cobertura de comandos |
| R3 | Integration table EN | Todas as 22 entradas do registry listadas |
| R4 | Integration table PT | Idem, com paths corretos |
| R5 | Package list | 13 pacotes `pkg/` referenciados corretamente |
| R6 | Quick Start | Comandos válidos para o tipo de projeto no contexto |
| R7 | Version badge | Badge mostra `1.4.0` (ou versão corrente) |
| R8 | Go version | Afirma "core Go 1.22+, OTel Go 1.25+" (verificar `go.mod`) |
| R9 | Project structure | Diagramas EN+PT incluem `ginger.yaml` e `.ginger/` |
| R10 | Docker section PT | Não referencia modo `api` removido |
| R11 | Safe evolution flow | Comandos `init`, `inspect`, `add --plan`, `generate tests --scan` |
| R12 | Positioning table | Tabela comparativa com Gin/Echo/Fiber/Cobra presente |
| R13 | Links internos | Todos os `./docs/*.md` e `./examples/*` resolvem |

### 3.2 `docs/GETTING_STARTED.md`

| # | Check | Critério |
|---|-------|----------|
| G1 | Comando `--cli` | Caminho documentado: `cmd/<nome>`, NÃO `cmd/<nome>-cli` |
| G2 | Comandos úteis | Lista `init`, `inspect`, `docs`, `generate tests --scan` |
| G3 | Scaffold structure | Inclui `ginger.yaml` e `.ginger/` |
| G4 | Safe generation | Menciona `--plan` / `--force` |
| G5 | Go version | Afirma "core 1.22+" corretamente |

### 3.3 `docs/QUICK_REFERENCE.md`

| # | Check | Critério |
|---|-------|----------|
| Q1 | CLI commands | Todos os 11 comandos do `root.go` listados |
| Q2 | `--plan` / `--force` | Documentados em `add` e `generate` |
| Q3 | `generate tests --scan` | Listado |
| Q4 | Import paths | Todos os imports referenciam pacotes que existem |
| Q5 | `--cli` path | `cmd/foobar` (correto) |

### 3.4 `docs/ARCHITECTURE.md`

| # | Check | Critério |
|---|-------|----------|
| A1 | Internal packages | 12 pacotes listados no diagrama |
| A2 | Project-safety | Menciona `init`, `inspect`, plan/apply |
| A3 | Request flow | Consistente entre diagrama e texto |
| A4 | Directory diagram | Reflete estrutura real do scaffold |

### 3.5 `docs/PACKAGES.md`

| # | Check | Critério |
|---|-------|----------|
| P1 | 13 pacotes documentados | Um por pacote em `pkg/` |
| P2 | `telemetry.Setup` | Assinatura correta: `func Setup(ctx, cfg Config) (*Provider, error)` |
| P3 | APIs exportadas | Funções/tipos documentados batem com código |
| P4 | Exemplos de código | Compilam com a versão documentada |

### 3.6 `docs/INTEGRATIONS.md`

| # | Check | Critério |
|---|-------|----------|
| I1 | 22 integrações | Todas do registry listadas |
| I2 | ORMs | `gorm`, `sqlx`, `bun` presentes |
| I3 | Package paths | Batem exatamente com o registry |
| I4 | `--plan` / `--force` | Documentados |
| I5 | Compose updates | Lista de integrações que atualizam compose confere com código |

### 3.7 `docs/PROJECT_STRUCTURE.md`

| # | Check | Critério |
|---|-------|----------|
| S1 | Internal packages | 12 pacotes listados |
| S2 | pkg packages | 13 pacotes listados |
| S3 | Root dirs | Inclui `plan/` |
| S4 | Module info | `go 1.22.0` confere com `go.mod` |

### 3.8 `docs/TESTING.md`

| # | Check | Critério |
|---|-------|----------|
| T1 | `generate tests --scan` | Mencionado |
| T2 | CI Go version | Recomenda 1.22+ (não 1.25) |
| T3 | Test helpers | API documentada bate com `pkg/testhelper` |

### 3.9 `docs/DEPLOYMENT.md`

| # | Check | Critério |
|---|-------|----------|
| D1 | Docker image tags | `alpine:3.20`+, `golang:1.22`+ |
| D2 | GitHub Actions | `actions/checkout@v4`, `docker/setup-buildx-action@v3` |
| D3 | Compose integrations | Lista confere com `mergeIntegrationIntoCompose` |

### 3.10 `AGENTS.md`

| # | Check | Critério |
|---|-------|----------|
| AG1 | Bloco DWYT | Presente e válido |
| AG2 | Comandos referenciados | `rtk` disponível? |

---

## 4. Ordem de Execução

```
Fase A (deletar) → Fase B (consolidar) → Fase C (atualizar) → Fase D (EN↔PT)
                                                                    │
                                                    ┌───────────────┘
                                                    ▼
                                           Revalidação completa
                                           (checklist 3.1–3.10)
                                                    │
                                                    ▼
                                           go build ./...
                                           git diff --stat
                                                    │
                                                    ▼
                                           Commit + release v1.4.1
```

---

## 5. Métricas de Sucesso

| Métrica | Antes | Depois |
|---------|-------|--------|
| Arquivos `.md` totais | 50 | ~43 |
| `docs/` arquivos | 13 | 10 |
| `plan/` arquivos | 9 | 6 |
| Linhas totais docs/ | ~7.088 | ~6.000 |
| Duplicações eliminadas | — | 5 |
| Issues de validação abertas | — | 0 |
| Cobertura EN↔PT | Parcial | Completa |
| Build | ✅ | ✅ |

---

## 6. Notas

- **Release notes** (`releases/v*.*.*/RELEASE_NOTES.md`): 23 arquivos históricos. Não deletar — são registros imutáveis de cada release. O `gh release create` os referencia.
- **`examples/existing-api/README.md`**: Mantido — documenta o exemplo de uso do Ginger em projeto não-Ginger.
- **`pkg/telemetry/README.md`**: Mantido — documenta o submódulo opcional.
- **`scripts/USAGE.md`**: Mantido — documenta os scripts de release e teste massivo.
- **`AGENTS.md`**: Mantido — configuração de agentes AI (DWYT).
