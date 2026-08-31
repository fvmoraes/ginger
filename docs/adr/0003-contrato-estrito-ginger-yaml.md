# ADR-0003 — Contrato estrito ginger.yaml (KnownFields)

**Status**: Aceito · **Data**: 2026-06-28 (DWYT), ratificado 2026-08-31

## Contexto
Typos em configuração devem falhar cedo, não virar comportamento misterioso.

## Decisão
`ginger.yaml` e `.ginger/manifest.yaml` são parseados com `KnownFields(true)`: campos desconhecidos são rejeitados. Campos novos adicionados pelo Ginger são opcionais (`omitempty`) — manifests antigos parseiam.

## Consequências
+ Erros de configuração detectados na hora.
− **Downgrade** de versão do Ginger pode rejeitar manifests novos (documentado por release).
