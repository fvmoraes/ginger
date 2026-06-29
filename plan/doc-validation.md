# Documentation Validation Plan

> Validação completa de todos os `.md` do repositório contra a realidade do código v1.4.0.
> Data: 2026-06-29

---

## Metodologia

1. **Ground truth**: código fonte real em `internal/` e `pkg/`
2. **Cross-reference**: cada claim da documentação verificada contra o código
3. **Severidade**: CRITICAL (quebra UX, engana usuário), HIGH (informação desatualizada), MEDIUM (omissão), LOW (cosmético)

---

## Sumário de Arquivos Analisados

| Arquivo | Linhas | Issues | Status |
|---------|--------|--------|--------|
| `README.md` | 1357 | 7 | Precisa correção |
| `docs/GETTING_STARTED.md` | 444 | 4 | Precisa correção |
| `docs/ARCHITECTURE.md` | 541 | 3 | Precisa correção |
| `docs/PACKAGES.md` | 873 | 1 | Precisa correção |
| `docs/INTEGRATIONS.md` | 766 | 5 | Precisa correção |
| `docs/PROJECT_STRUCTURE.md` | 323 | 4 | Precisa correção |
| `docs/QUICK_REFERENCE.md` | 628 | 4 | Precisa correção |
| `docs/COPY_PASTE.md` | 786 | 1 | Baixa prioridade |
| `docs/TESTING.md` | 812 | 2 | Precisa correção |
| `docs/DEPLOYMENT.md` | 893 | 2 | Baixa prioridade |
| `docs/SUMMARY.md` | 272 | 2 | Precisa correção |
| `docs/SAFE_EVOLUTION_PLAN.md` | 67 | 0 | OK |
| `docs/CHANGELOG.md` | 182 | 1 | Grave |
| `AGENTS.md` | 84 | 0 | OK |
| `CHANGELOG.md` (root) | 306 | 2 | Baixa prioridade |
| `scripts/USAGE.md` | 84 | 0 | OK |
| `pkg/telemetry/README.md` | 10 | 0 | OK |

**Total: 18 arquivos, 35 issues encontrados, 0 falsos positivos.**

---

## Issues por Arquivo

### README.md (7 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 1 | **CRITICAL** | CLI Reference (EN+PT) omite comandos `init`, `inspect`, `docs`, `generate tests --scan` |
| 2 | **HIGH** | CLI Reference não documenta flags `--plan`/`--force` |
| 3 | **HIGH** | Integration tables (EN+PT) não listam ORMs `gorm`, `sqlx`, `bun` |
| 4 | **HIGH** | Versão PT do Docker section referencia `api` mode removido no v1.3.0 |
| 5 | **HIGH** | Quick Start mostra `generate service deployer` em projeto `--service` (inválido) |
| 6 | **MEDIUM** | PT CLI Reference não lista `generate command`, `handler`, `service` |
| 7 | **LOW** | PT Cheat Sheet não lista `generate smoke-test` |

### docs/GETTING_STARTED.md (4 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 8 | **CRITICAL** | `--cli` documentado como `cmd/<nome>-cli` (código real: `cmd/<nome>`) |
| 9 | **HIGH** | Entrada `--cli` duplicada com caminhos conflitantes |
| 10 | **HIGH** | Nenhuma menção a `ginger init`, `ginger inspect`, `--plan`, `--force` |
| 11 | **MEDIUM** | Estrutura de scaffold mostrada está desatualizada (sem `ginger.yaml`, `.ginger/`) |

### docs/ARCHITECTURE.md (3 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 12 | **HIGH** | Diagrama de diretórios não inclui `project/`, `plan/`, `region/`, `capability/`, `manifest/`, `docsgen/` |
| 13 | **MEDIUM** | Sem menção a project-safety (init, inspect, managed regions) |
| 14 | **LOW** | Step mismatch entre diagrama de 7 passos e texto (repository vs ports/adapters) |

### docs/PACKAGES.md (1 issue)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 35 | **CRITICAL** | `telemetry.Setup()` documentado com assinatura errada (args posicionais vs Config struct) |

### docs/INTEGRATIONS.md (5 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 18 | **HIGH** | Tabela de integrações não inclui ORMs (`gorm`, `sqlx`, `bun`) |
| 19 | **MEDIUM** | Sem documentação das flags `--plan`/`--force` |
| 20 | **MEDIUM** | MongoDB package path incorreto (falta `/v2`) |
| 21 | **MEDIUM** | PubSub package path incorreto (falta `/v2`) |
| 22 | **LOW** | Funções como `database.ConnectMySQL` referenciadas como se fossem do framework (são geradas) |

### docs/PROJECT_STRUCTURE.md (4 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 15 | **CRITICAL** | Lista de pacotes `internal/` omite 7 pacotes (buildinfo, capability, docsgen, manifest, plan, project, region) |
| 16 | **MEDIUM** | Omite diretório raiz `plan/` |
| 17 | **MEDIUM** | Lista de docs omite `SAFE_EVOLUTION_PLAN.md` |
| 18 | **LOW** | Lista de docs omite `PROJECT_STRUCTURE.md` (auto-referência) |

### docs/QUICK_REFERENCE.md (4 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 21 | **CRITICAL** | CLI commands omite `init`, `inspect`, `docs` |
| 22 | **MEDIUM** | Sem `--plan`/`--force` flags |
| 23 | **CRITICAL** | `--cli` documentado como `cmd/foobar` (correto) e `cmd/<nome>-cli` (errado na GETTING_STARTED) |
| 24 | **MEDIUM** | `generate tests --scan` não listado |

### docs/CHANGELOG.md (docs/) (1 issue)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 25 | **CRITICAL** | Preso na v1.0.0 (2024-03-14). Framework está na v1.4.0. 2+ anos de evolução não documentados. |

### docs/TESTING.md + DEPLOYMENT.md + SUMMARY.md (4 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 28 | **MEDIUM** | TESTING.md omite `ginger generate tests --scan` |
| 29 | **LOW** | TESTING.md CI examples usam Go 1.25 (core requer 1.22) |
| 30 | **LOW** | DEPLOYMENT.md Docker tags outdated (alpine 3.19, actions v3) |
| 31 | **MEDIUM** | SUMMARY.md herda omissões de todos os docs |

### CHANGELOG.md (root) (2 issues)

| # | Severidade | Descrição |
|---|-----------|-----------|
| 26 | **LOW** | v1.1.1 entry "Go 1.25+ required" (agora core é 1.22+) |
| 27 | **LOW** | Link references em ordem não-decrescente (v1.3.6 após v1.4.0) |

---

## Ordem de Correção

1. **README.md** — face pública do projeto, todas as issues
2. **docs/QUICK_REFERENCE.md** — referência rápida mais usada
3. **docs/GETTING_STARTED.md** — onboarding de novos usuários
4. **docs/INTEGRATIONS.md** — adicionar ORMs e corrigir paths
5. **docs/PROJECT_STRUCTURE.md** — atualizar lista de pacotes
6. **docs/ARCHITECTURE.md** — atualizar diagrama de diretórios
7. **docs/PACKAGES.md** — corrigir telemetry.Setup
8. **docs/CHANGELOG.md** (docs/) — adicionar entries faltantes
9. **docs/TESTING.md** — adicionar `--scan`
10. **docs/DEPLOYMENT.md** — tags e versões
11. **docs/SUMMARY.md** — atualizar com novas docs
12. **CHANGELOG.md** (root) — correções históricas
