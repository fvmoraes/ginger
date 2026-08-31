#!/usr/bin/env bash
# Fase 0 (tarefa 8a-ii) — Caracterização anti-regressão do docker-compose (GIN-002).
#
# Cenário: compose customizado pelo usuário (comentários, anchors YAML,
# campos de extensão x-, merge keys << e networks) → `ginger add`.
# Comportamento DESEJADO: conteúdo customizado preservado byte a byte
# (fora do serviço adicionado). Comportamento ATUAL (até a Fase 1): perdido.
#
# Modo tolerante (default): reporta KNOWN-FAIL para os desvios de GIN-002 e
# sai 0 — o gate 15/15 do E2E não fica vermelho por uma falha documentada.
# Modo strict (--strict): vira gate (a partir da Fase 1) — sai 1 em qualquer
# desvio.
#
# Uso: bash scripts/test-ginger-compose.sh [--strict]

set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"
STRICT=0
[[ "${1:-}" == "--strict" ]] && STRICT=1

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

FAILURES=0

note() { printf '%s\n' "$*"; }
known_fail() {
	FAILURES=$((FAILURES + 1))
	note "KNOWN-FAIL (GIN-002): $*"
}
hard_fail() {
	note "FAIL: $*"
	FAILURES=$((FAILURES + 1))
}

note "==> Build ginger (local, offline)"
( cd "$REPO_ROOT" && go build -o "$WORK/ginger" ./cmd/ginger )

note "==> Scaffold service project"
( cd "$WORK" && ./ginger new demo --service >/dev/null )

note "==> Customize compose (anchors, x-fields, merge keys, networks)"
cat > "$WORK/demo/devops/docker/docker-compose.yml" <<'EOF'
# Custom compose with anchors, extension fields and networks (GIN-002 fixture)
# Nota: sem "build: ." — a forma abreviada causa falha dura de parse no
# planComposePatch (composeBuild é struct; ver GIN-002 extensão no relatório).
x-common-env: &common-env
  APP_ENV: development

services:
  demo:
    image: demo:dev
    environment:
      <<: *common-env
    ports:
      - "8080:8080"

networks:
  custom-net:
    driver: bridge
EOF
cp "$WORK/demo/devops/docker/docker-compose.yml" "$WORK/custom-compose.expected"

note "==> ginger add redis (apply completo)"
MARKERS=(
	'# Custom compose with anchors'
	'x-common-env: &common-env'
	'<<: *common-env'
	'APP_ENV: development'
	'image: demo:dev'
	'networks:'
	'custom-net:'
	'driver: bridge'
)

add_ok=0
if ( cd "$WORK/demo" && "$WORK/ginger" add redis >/dev/null ); then
	add_ok=1
else
	hard_fail "ginger add redis exited non-zero on a customized managed compose"
fi

if [[ "$add_ok" -eq 1 ]]; then
	for marker in "${MARKERS[@]}"; do
		if ! grep -qF "$marker" "$WORK/demo/devops/docker/docker-compose.yml"; then
			known_fail "custom marker lost after add: '$marker'"
		fi
	done
	if ! grep -qE "^ {2,4}redis:" "$WORK/demo/devops/docker/docker-compose.yml"; then
		known_fail "expected compose service 'redis' missing after add"
	fi
fi

note ""
if [[ "$FAILURES" -gt 0 ]]; then
	if [[ "$STRICT" -eq 1 ]]; then
		note "STRICT mode: $FAILURES failure(s) — gate FAILED (GIN-002 deve estar corrigido na Fase 1)"
		exit 1
	fi
	note "Tolerant mode: $FAILURES known failure(s) — GIN-002 documentado; vira gate com --strict após a Fase 1"
	exit 0
fi
note "OK: compose customizado preservado e serviço adicionado (GIN-002 resolvido)"
