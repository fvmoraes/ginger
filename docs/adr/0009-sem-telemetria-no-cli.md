# ADR-0009 — Sem telemetria no CLI

**Status**: Aceito · **Data**: 2026-08-31 (Fase 5)

## Contexto
Métricas de adoção ajudam priorização, mas o Ginger roda em máquinas de desenvolvedores e em CI corporativo.

## Decisão
O CLI **não coleta nem envia telemetria**. Métricas de adoção vêm de fontes públicas (downloads do GitHub Releases, clonagens, issues).

## Consequências
+ Privacidade por padrão; nada para auditar no binário; zero custo de conformidade (LGPD/GDPR).
− Visibilidade de adoção limitada a proxies públicos.
