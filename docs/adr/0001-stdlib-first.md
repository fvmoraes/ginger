# ADR-0001 — Stdlib-first (Go HTTP + yaml.v3)

**Status**: Aceito · **Data**: 2026-06-28 (DWYT), ratificado 2026-08-31

## Contexto
Framework distribuído via `go get`; dependências transitivas custam confiança, auditoria e compatibilidade.

## Decisão
O módulo raiz depende apenas de `gopkg.in/yaml.v3`. `pkg/ws`, `pkg/sse` e o router são stdlib puro. Alternativas (Chi/Gin/Echo, gorilla/nhooyr websocket) rejeitadas.

## Consequências
+ Portabilidade, auditoria trivial, superfície de ataque mínima.
− Mais código próprio para manter (ex.: decoder WebSocket RFC 6455 — endurecido na Fase 1).
