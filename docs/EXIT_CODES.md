# Exit Codes — Ginger CLI

Tabela única de códigos de saída (GIN-020, Fase 4). Automatizações devem confiar nestes valores.

| Código | Significado | Exemplos |
|---|---|---|
| `0` | Sucesso | comando concluído |
| `1` | Erro de execução | projeto não encontrado, plano com erros, falha de apply, falha de `go get` |
| `2` | Erro de uso (argumentos/flags) | flag desconhecida, argumento extra, `--root` sem valor |
| `130` | Interrompido por SIGINT (128+2) | Ctrl-C durante `ginger run` (GIN-029) |
| `143` | Interrompido por SIGTERM (128+15) | `kill <pid>` durante `ginger run` |

## Flags globais

| Flag | Comandos | Efeito |
|---|---|---|
| `--root <dir>` | add, generate, init, inspect, docs, doctor, run, build | opera no projeto em `<dir>` a partir de qualquer diretório (GIN-027) |
| `--plan` | add, generate | renderiza o plano sem aplicar |
| `--json` | add, generate | saída máquina-legável do plano (CI) — combina com `--plan` |
| `--quiet` | add, generate | silencia a saída normal (só erros) |
| `--force` | add, generate, init | permite sobrescrever alvos gerenciados |
