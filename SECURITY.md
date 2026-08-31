# Security Policy — Ginger

## Reporting a vulnerability

**NÃO abra uma issue pública para vulnerabilidades.**

1. Use o **GitHub Security Advisories** ("Report a vulnerability" na aba Security do repositório) — canal privado, com SLA de resposta.
2. Alternativa: e-mail para o mantenedor (ver perfil do repositório).

## SLA de resposta

| Severidade | Primeira resposta | Correção alvo |
|---|---|---|
| Crítica (RCE, perda de dados) | 48h | 7 dias |
| Alta (bypass de segurança, DoS) | 5 dias | 30 dias |
| Média/Baixa | 14 dias | Próxima release |

## Escopo

**Em escopo**: código-fonte do framework (`internal/`, `pkg/`), CLI (`cmd/ginger`), templates gerados, scripts de instalação (`install.sh`), workflows de CI/CD.

**Fora de escopo**: vulnerabilidades em projetos *gerados* pelo Ginger que decorram de código do usuário; dependências dos templates de integração devem ser reportadas aos projetos upstream (a tabela central em `internal/integrations/docker_services.go` é renovada periodicamente — ver `docs/COMPATIBILITY.md`).

## Práticas de segurança aplicadas

- Write boundary único (`internal/plan`): contenção de paths, verificação de symlinks, preflight de hash, criação exclusiva, escrita atômica.
- WebSocket/SSE endurecidos (limites de frame, origin check, deadlines, escaping — GIN-003/012).
- Zero segredos em logs (logger com redação).
- CI: CodeQL, govulncheck (raiz + submódulo), race detector, lint completo.
- Supply chain: actions pinadas por SHA, Dependabot, SBOM SPDX no release, proveniência SLSA L3.
- Instalador com verificação de checksum.

## Versões suportadas

Ver `docs/COMPATIBILITY.md` (política de compatibilidade e janela de suporte).
