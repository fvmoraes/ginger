# Compatibilidade e Ciclo de Manutenção

## Política de versões Go (GIN-028, decisão P3 executada)

- **Framework**: suporta as **duas últimas versões estáveis** do Go (matriz CI: 1.24/1.25/1.26).
- **Módulo raiz** declara `go 1.22` (mínimo de compilação; não necessariamente testado em toda release).
- **Scaffolds** gravam mínimo determinístico por tipo: `service`/`worker` → **1.25** (runtime Ginger + otel), `generic`/`cli` → **1.22**.
- **Submódulo `pkg/telemetry`**: `go 1.25` (exigência do OpenTelemetry), isolado da raiz.

## Compatibilidade semântica

| Área | Compromisso |
|---|---|
| `pkg/` (API pública) | Sem breaking changes em minors/patches; novos campos são aditivos (`omitempty`); deprecação com aviso por **uma release mínima** antes da remoção |
| `ginger.yaml` / `.ginger/manifest.yaml` | Campos novos são opcionais; manifests antigos parseiam (KnownFields rejeita apenas desconhecidos no YAML); **downgrade** entre versões do Ginger pode rejeitar campos novos — documentado por release |
| CLI | Exit codes estáveis (`docs/EXIT_CODES.md`); novos códigos são aditivos |
| Templates gerados | O Ginger não reescreve projetos existentes sem plano revisável; mudanças de template afetam apenas scaffolds novos |
| Compose gerenciado | Merge direto apenas para conteúdo intacto desde a geração (hash de proveniência); modificado → patch revisável |

## Ciclo de manutenção

| Cadência | Atividade |
|---|---|
| Contínua | Dependabot (gomod raiz + telemetry + actions); CodeQL semanal; Scorecard semanal |
| Por release | govulncheck limpo (raiz + telemetry); E2E 15/15; suite de caracterização; goldens CLI |
| Mensal | Renovação da tabela central de imagens Docker (`internal/integrations/docker_services.go`) — revisão manual das minor/patações upstream |
| Trimestral | Revisão da política de versões Go; atualização da matriz CI |

## Deprecação

1. Aviso no output do comando e no CHANGELOG (seção "Deprecated").
2. Uma release mínima com aviso ativo.
3. Remoção apenas em major (ou minor, se funcionalidade nunca documentada).
