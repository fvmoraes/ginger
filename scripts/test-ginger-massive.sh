#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"
CURRENT_BRANCH="$(git -C "$REPO_ROOT" branch --show-current 2>/dev/null || true)"

REPO_URL="${REPO_URL:-$REPO_ROOT}"
CHECKOUT_REF="${CHECKOUT_REF:-${CURRENT_BRANCH:-main}}"
BASE_WORKSPACE_DIR="${WORKSPACE_DIR:-$REPO_ROOT/my-local/workspace}"
SERVICE_PORT="${SERVICE_PORT:-18081}"
WORKER_PORT="${WORKER_PORT:-18082}"
VERBOSE="${VERBOSE:-1}"
COLOR_ENABLED="${COLOR_ENABLED:-1}"

if [[ -e "$BASE_WORKSPACE_DIR" ]]; then
	WORKSPACE_DIR="${BASE_WORKSPACE_DIR%/}-$(date +%Y%m%d-%H%M%S)"
else
	WORKSPACE_DIR="$BASE_WORKSPACE_DIR"
fi

BIN_DIR="$WORKSPACE_DIR/bin"
SRC_DIR="$WORKSPACE_DIR/ginger-src"
PLAYGROUND_DIR="$WORKSPACE_DIR/playground"
LOG_DIR="$WORKSPACE_DIR/logs"
ENV_FILE="$WORKSPACE_DIR/use-ginger-env.sh"

SERVICE_PROJECT="$PLAYGROUND_DIR/ginger-service"
WORKER_PROJECT="$PLAYGROUND_DIR/ginger-jobs"
CLI_PROJECT="$PLAYGROUND_DIR/ginger-cli"
GENERIC_PROJECT="$PLAYGROUND_DIR/ginger-generic"

SERVICE_PID=""
WORKER_PID=""
GENERIC_PID=""
SERVICE_TAIL_PID=""
WORKER_TAIL_PID=""
GENERIC_TAIL_PID=""
STEP_COUNTER=0
CURRENT_LOG=""
CURRENT_STEP=""
TOTAL_START_TS="$(date +%s)"
LAST_BG_PID=""
LAST_TAIL_PID=""
declare -a STEP_TITLES=()
declare -a STEP_RESULTS=()
declare -a STEP_DURATIONS=()
declare -a STEP_LOGS=()

if [[ "$COLOR_ENABLED" == "1" && -z "${NO_COLOR:-}" && ( -t 1 || "${FORCE_COLOR:-0}" == "1" ) && "${TERM:-}" != "dumb" ]]; then
	C_RESET=$'\033[0m'
	C_STEP=$'\033[1;38;5;213m'
	C_NOTE=$'\033[1;38;5;226m'
	C_CMD=$'\033[1;38;5;51m'
	C_LOG=$'\033[1;38;5;87m'
	C_OK=$'\033[1;38;5;118m'
	C_FAIL=$'\033[1;38;5;197m'
	C_TIME=$'\033[1;38;5;208m'
	C_DIM=$'\033[38;5;244m'
	C_SUM=$'\033[1;38;5;141m'
	C_WARN=$'\033[1;38;5;220m'
else
	C_RESET=""
	C_STEP=""
	C_NOTE=""
	C_CMD=""
	C_LOG=""
	C_OK=""
	C_FAIL=""
	C_TIME=""
	C_DIM=""
	C_SUM=""
	C_WARN=""
fi

INTEGRATIONS=(
	postgres
	mysql
	sqlite
	sqlserver
	couchbase
	mongodb
	clickhouse
	redis
	kafka
	rabbitmq
	nats
	pubsub
	grpc
	mcp
	sse
	websocket
	otel
	prometheus
	swagger
)

usage() {
	cat <<EOF
Massive Ginger validation script.

This script:
1. Clones Ginger into an isolated workspace
2. Installs the CLI into a private bin directory
3. Exports PATH inside the script run
4. Validates version/help
5. Exercises ginger new for generic/service/worker/cli
6. Exercises every generator
7. Exercises every integration via ginger add
8. Builds, tests, doctors, and runs the generated projects
9. Prints a final full report with OK and FAIL per step

Defaults:
  REPO_URL       current repository root ($REPO_ROOT)
  CHECKOUT_REF   current branch (${CURRENT_BRANCH:-main})
  WORKSPACE_DIR  $REPO_ROOT/my-local/workspace

Behavior:
  - Local REPO_URL paths are copied as a working tree snapshot, including
    uncommitted changes.
  - Remote REPO_URL values are cloned with git using CHECKOUT_REF.

Environment variables:
  REPO_URL       Git repository URL or local path
  CHECKOUT_REF   Branch, tag, or ref to clone
  WORKSPACE_DIR  Output workspace directory
  SERVICE_PORT   Port used for service run validation (default: 18081)
  WORKER_PORT    Port used for worker run validation (default: 18082)
  VERBOSE        1 shows live output and saves logs, 0 keeps terminal quieter
  COLOR_ENABLED  1 enables colorful terminal output, 0 disables colors

Examples:
  ./scripts/test-ginger-massive.sh
  REPO_URL=https://github.com/fvmoraes/ginger.git CHECKOUT_REF=main ./scripts/test-ginger-massive.sh
EOF
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
	usage
	exit 0
fi

cleanup() {
	stop_process "$SERVICE_PID" "service project" "$SERVICE_TAIL_PID"
	stop_process "$WORKER_PID" "worker project" "$WORKER_TAIL_PID"
	stop_process "$GENERIC_PID" "generic project" "$GENERIC_TAIL_PID"
}

trap cleanup EXIT

require_commands() {
	local missing=0
	for cmd in git go curl mktemp pgrep tail; do
		if ! command -v "$cmd" >/dev/null 2>&1; then
			echo "Missing required command: $cmd" >&2
			missing=1
		fi
	done

	if [[ "$missing" -ne 0 ]]; then
		exit 1
	fi
}

safe_name() {
	echo "$1" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-'
}

timestamp() {
	date '+%H:%M:%S'
}

format_duration() {
	local seconds="$1"
	printf '%02dm%02ds' "$((seconds / 60))" "$((seconds % 60))"
}

port_is_open() {
	local port="$1"
	(: >/dev/tcp/127.0.0.1/"$port") >/dev/null 2>&1
}

choose_available_port() {
	local port="$1"

	while port_is_open "$port"; do
		printf '%s[%s]%s %sPort %s is already in use; trying %s%s\n' \
			"$C_TIME" "$(timestamp)" "$C_RESET" "$C_WARN" "$port" "$((port + 1))" "$C_RESET" >&2
		port=$((port + 1))
	done

	printf '%s\n' "$port"
}

note() {
	printf '%s[%s]%s %s%s%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_NOTE" "$*" "$C_RESET"
}

command_string() {
	printf '%q ' "$@"
}

run_cmd() {
	printf '%s[%s]%s %s$ %s%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_CMD" "$(command_string "$@")" "$C_RESET"
	"$@"
}

run_cmd_discard_output() {
	printf '%s[%s]%s %s$ %s%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_CMD" "$(command_string "$@")" "$C_RESET"
	"$@" >/dev/null
}

start_background() {
	local label="$1"
	local log_file="$2"
	shift 2

	: >"$log_file"
	LAST_TAIL_PID=""
	printf '%s[%s]%s %sBackground log:%s %s%s%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_LOG" "$C_RESET" "$C_DIM" "$log_file" "$C_RESET"
	printf '%s[%s]%s %s$ %s%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_CMD" "$(command_string "$@")" "$C_RESET"
	if [[ "$VERBOSE" == "1" ]]; then
		tail -n 0 -f "$log_file" &
		LAST_TAIL_PID="$!"
		note "Streaming live output for $label (tail PID $LAST_TAIL_PID)"
	fi
	"$@" >>"$log_file" 2>&1 &
	LAST_BG_PID="$!"
	note "Started $label with PID $LAST_BG_PID"
}

list_descendants() {
	local pid="$1"
	local child

	while IFS= read -r child; do
		[[ -n "$child" ]] || continue
		list_descendants "$child"
		printf '%s\n' "$child"
	done < <(pgrep -P "$pid" 2>/dev/null || true)
}

signal_process_tree() {
	local pid="$1"
	local signal_name="$2"
	local target
	local -a targets=()

	while IFS= read -r target; do
		[[ -n "$target" ]] || continue
		targets+=("$target")
	done < <(list_descendants "$pid")
	targets+=("$pid")

	for target in "${targets[@]}"; do
		if kill -0 "$target" >/dev/null 2>&1; then
			kill "-$signal_name" "$target" >/dev/null 2>&1 || true
		fi
	done
}

wait_for_exit() {
	local pid="$1"
	local timeout_seconds="$2"
	local second

	for second in $(seq 1 "$timeout_seconds"); do
		if ! kill -0 "$pid" >/dev/null 2>&1; then
			wait "$pid" >/dev/null 2>&1 || true
			return 0
		fi
		sleep 1
	done

	return 1
}

stop_tail() {
	local tail_pid="${1:-}"

	if [[ -n "$tail_pid" ]] && kill -0 "$tail_pid" >/dev/null 2>&1; then
		kill "$tail_pid" >/dev/null 2>&1 || true
		wait "$tail_pid" >/dev/null 2>&1 || true
	fi
}

stop_process() {
	local pid="$1"
	local label="$2"
	local tail_pid="${3:-}"

	if [[ -z "$pid" ]]; then
		stop_tail "$tail_pid"
		return 0
	fi

	if ! kill -0 "$pid" >/dev/null 2>&1; then
		stop_tail "$tail_pid"
		return 0
	fi

	note "Stopping $label (PID $pid)"
	signal_process_tree "$pid" INT
	if wait_for_exit "$pid" 5; then
		note "$label stopped gracefully"
		stop_tail "$tail_pid"
		return 0
	fi

	printf '%s[%s]%s %sGraceful stop timed out for %s; sending TERM%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_WARN" "$label" "$C_RESET"
	signal_process_tree "$pid" TERM
	if wait_for_exit "$pid" 5; then
		note "$label stopped after TERM"
		stop_tail "$tail_pid"
		return 0
	fi

	printf '%s[%s]%s %sForce killing %s after timeout%s\n' "$C_TIME" "$(timestamp)" "$C_RESET" "$C_FAIL" "$label" "$C_RESET"
	signal_process_tree "$pid" KILL
	wait "$pid" >/dev/null 2>&1 || true
	stop_tail "$tail_pid"
}

run_step() {
	local title="$1"
	shift
	local start_ts
	local end_ts
	local elapsed
	local status
	local step_exit=0

	STEP_COUNTER=$((STEP_COUNTER + 1))
	mkdir -p "$LOG_DIR"
	CURRENT_STEP="$title"
	CURRENT_LOG="$LOG_DIR/$(printf '%02d' "$STEP_COUNTER")-$(safe_name "$title").log"
	start_ts="$(date +%s)"

	printf '\n%s[%02d]%s %s%s%s\n' "$C_STEP" "$STEP_COUNTER" "$C_RESET" "$C_STEP" "$title" "$C_RESET"
	printf '%sLog:%s %s%s%s\n' "$C_LOG" "$C_RESET" "$C_DIM" "$CURRENT_LOG" "$C_RESET"
	set +e
	if [[ "$VERBOSE" == "1" ]]; then
		( set -Eeuo pipefail; "$@" ) > >(tee "$CURRENT_LOG") 2>&1
		step_exit="$?"
	else
		( set -Eeuo pipefail; "$@" ) >"$CURRENT_LOG" 2>&1
		step_exit="$?"
	fi
	set -e
	if [[ "$step_exit" -eq 0 ]]; then
		status="OK"
	else
		status="FAIL"
	fi

	end_ts="$(date +%s)"
	elapsed="$((end_ts - start_ts))"
	STEP_TITLES+=("$title")
	STEP_RESULTS+=("$status")
	STEP_DURATIONS+=("$(format_duration "$elapsed")")
	STEP_LOGS+=("$CURRENT_LOG")

	printf '%sDuration:%s %s%s%s\n' "$C_SUM" "$C_RESET" "$C_DIM" "$(format_duration "$elapsed")" "$C_RESET"
	if [[ "$status" == "OK" ]]; then
		printf '%sResult:%s %sOK%s\n' "$C_SUM" "$C_RESET" "$C_OK" "$C_RESET"
	else
		printf '%sResult:%s %sFAIL%s\n' "$C_SUM" "$C_RESET" "$C_FAIL" "$C_RESET"
	fi
	if [[ "$status" == "FAIL" ]]; then
		printf '%sLog:%s %s%s%s\n' "$C_LOG" "$C_RESET" "$C_DIM" "$CURRENT_LOG" "$C_RESET"
		tail -n 120 "$CURRENT_LOG" || true
		print_summary
		exit 1
	fi
}

assert_file() {
	local path="$1"
	note "Checking file exists: $path"
	[[ -f "$path" ]] || {
		echo "Expected file not found: $path" >&2
		return 1
	}
}

link_local_ginger() {
	local project_dir="$1"
	note "Linking local Ginger source into: $project_dir"
	(
		cd "$project_dir"
		run_cmd go mod edit -replace github.com/fvmoraes/ginger="$SRC_DIR"
	)
}

wait_for_http() {
	local url="$1"
	local retries="${2:-30}"
	local sleep_seconds="${3:-1}"
	local attempt

	note "Waiting for endpoint: $url"
	for attempt in $(seq 1 "$retries"); do
		if curl -fsS "$url" >/dev/null 2>&1; then
			note "Endpoint ready: $url"
			return 0
		fi
		if (( attempt == 1 || attempt % 5 == 0 )); then
			note "Still waiting ($attempt/$retries): $url"
		fi
		sleep "$sleep_seconds"
	done

	echo "Timed out waiting for $url" >&2
	return 1
}

assert_http_status() {
	local method="$1"
	local url="$2"
	local expected="$3"
	shift 3

	local body_file
	local status

	body_file="$(mktemp)"
	printf '%s[%s]%s %s$ curl -sS -X %s %s %s%s\n' \
		"$C_TIME" "$(timestamp)" "$C_RESET" "$C_CMD" "$method" "$url" "$(command_string "$@")" "$C_RESET"
	status="$(curl -sS -o "$body_file" -w '%{http_code}' -X "$method" "$url" "$@")"
	if [[ "$status" != "$expected" ]]; then
		echo "Unexpected HTTP status for $method $url: got $status, want $expected" >&2
		if [[ -s "$body_file" ]]; then
			echo "--- response body ---" >&2
			cat "$body_file" >&2
			echo >&2
		fi
		rm -f "$body_file"
		return 1
	fi
	rm -f "$body_file"
}

prepare_workspace() {
	run_cmd mkdir -p "$BIN_DIR" "$PLAYGROUND_DIR" "$LOG_DIR"
	note "Writing environment helper: $ENV_FILE"
	cat >"$ENV_FILE" <<EOF
#!/usr/bin/env bash
export PATH="$BIN_DIR:\$PATH"
EOF
	run_cmd chmod +x "$ENV_FILE"
}

clone_repo() {
	if [[ -d "$REPO_URL" ]]; then
		note "Local repository detected; copying current working tree snapshot."
		note "CHECKOUT_REF is ignored for local snapshots."
		run_cmd mkdir -p "$SRC_DIR"
		if command -v rsync >/dev/null 2>&1; then
			run_cmd rsync -a --delete \
				--exclude .git \
				--exclude my-local \
				--exclude bin \
				--exclude dist \
				"$REPO_URL"/ "$SRC_DIR"/
			return
		fi

		note "rsync not found; falling back to tar-based copy."
		(
			cd "$REPO_URL"
			tar --exclude .git --exclude my-local --exclude bin --exclude dist -cf - .
		) | (
			cd "$SRC_DIR"
			tar -xf -
		)
		return
	fi

	run_cmd git clone --depth 1 --branch "$CHECKOUT_REF" "$REPO_URL" "$SRC_DIR"
}

install_ginger() {
	note "Installing ginger CLI into: $BIN_DIR"
	(
		cd "$SRC_DIR"
		run_cmd env GOBIN="$BIN_DIR" go install ./cmd/ginger
	)
	export PATH="$BIN_DIR:$PATH"
	note "PATH updated to include: $BIN_DIR"
	run_cmd hash -r
	run_cmd command -v ginger
}

validate_cli_basics() {
	export PATH="$BIN_DIR:$PATH"
	run_cmd hash -r
	run_cmd ginger version
	run_cmd_discard_output ginger help
	run_cmd_discard_output ginger new --help
}

create_generic_project() {
	export PATH="$BIN_DIR:$PATH"
	run_cmd mkdir -p "$PLAYGROUND_DIR"
	(
		cd "$PLAYGROUND_DIR"
		run_cmd ginger new "$(basename "$GENERIC_PROJECT")"
	)
	link_local_ginger "$GENERIC_PROJECT"
	(
		cd "$GENERIC_PROJECT"
		run_cmd go mod tidy
		run_cmd go test -v ./...
		run_cmd go build ./...
		run_cmd ginger build
	)
	assert_file "$GENERIC_PROJECT/bin/$(basename "$GENERIC_PROJECT")"
}

run_generic_project() {
	export PATH="$BIN_DIR:$PATH"
	local prev_dir="$PWD"
	local run_log="$LOG_DIR/generic-run.out"
	note "Starting generic project runtime validation"
	cd "$GENERIC_PROJECT"
	start_background "generic project" "$run_log" ginger run
	GENERIC_PID="$LAST_BG_PID"
	GENERIC_TAIL_PID="$LAST_TAIL_PID"
	run_cmd sleep 2
	stop_process "$GENERIC_PID" "generic project" "$GENERIC_TAIL_PID"
	GENERIC_PID=""
	GENERIC_TAIL_PID=""
	cd "$prev_dir"
}

create_service_project() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$PLAYGROUND_DIR"
		run_cmd ginger new "$(basename "$SERVICE_PROJECT")" --service
	)
	link_local_ginger "$SERVICE_PROJECT"
	(
		cd "$SERVICE_PROJECT"
		run_cmd go mod tidy
	)
}

exercise_service_integrations() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$SERVICE_PROJECT"
		for integration in "${INTEGRATIONS[@]}"; do
			run_cmd ginger add "$integration"
		done
		run_cmd go mod tidy
	)

	assert_file "$SERVICE_PROJECT/platform/database/postgres.go"
	assert_file "$SERVICE_PROJECT/platform/database/sqlite.go"
	assert_file "$SERVICE_PROJECT/platform/nosql/mongo.go"
	assert_file "$SERVICE_PROJECT/internal/api/swagger.go"
}

exercise_service_generators() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$SERVICE_PROJECT"
		run_cmd ginger generate crud user
		run_cmd ginger generate test user
		run_cmd ginger generate smoke-test
		run_cmd ginger generate swagger user
		run_cmd go mod tidy
		run_cmd go test -v ./...
		run_cmd go build ./...
		run_cmd ginger build
		run_cmd ginger doctor
	)

	assert_file "$SERVICE_PROJECT/internal/models/user.go"
	assert_file "$SERVICE_PROJECT/internal/api/handlers/user_handler.go"
	assert_file "$SERVICE_PROJECT/internal/api/handlers/user_handler_test.go"
	assert_file "$SERVICE_PROJECT/tests/integration/app_smoke_test.go"
	assert_file "$SERVICE_PROJECT/docs/openapi.json"
	assert_file "$SERVICE_PROJECT/bin/$(basename "$SERVICE_PROJECT")"
}

run_service_project() {
	export PATH="$BIN_DIR:$PATH"
	local prev_dir="$PWD"
	local payload_file="$WORKSPACE_DIR/user-create.json"
	local run_log="$LOG_DIR/service-run.out"
	cd "$SERVICE_PROJECT"
	note "Writing request payload: $payload_file"
	printf '{"name":"massive-test","email":"massive@example.com"}\n' >"$payload_file"
	start_background "service project" "$run_log" env HTTP_PORT="$SERVICE_PORT" ginger run
	SERVICE_PID="$LAST_BG_PID"
	SERVICE_TAIL_PID="$LAST_TAIL_PID"
	wait_for_http "http://127.0.0.1:$SERVICE_PORT/health" 30 1
	assert_http_status GET "http://127.0.0.1:$SERVICE_PORT/health" 200
	assert_http_status GET "http://127.0.0.1:$SERVICE_PORT/api/v1/ping" 200
	assert_http_status GET "http://127.0.0.1:$SERVICE_PORT/swagger" 200
	assert_http_status GET "http://127.0.0.1:$SERVICE_PORT/swagger/openapi.json" 200
	assert_http_status GET "http://127.0.0.1:$SERVICE_PORT/api/v1/users/123" 404
	assert_http_status POST "http://127.0.0.1:$SERVICE_PORT/api/v1/users" 201 \
		-H "Content-Type: application/json" \
		--data @"$payload_file"
	assert_http_status GET "http://127.0.0.1:$SERVICE_PORT/api/v1/users/massive-test" 200
	stop_process "$SERVICE_PID" "service project" "$SERVICE_TAIL_PID"
	SERVICE_PID=""
	SERVICE_TAIL_PID=""
	cd "$prev_dir"
}

create_worker_project() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$PLAYGROUND_DIR"
		run_cmd ginger new "$(basename "$WORKER_PROJECT")" -w
	)
	link_local_ginger "$WORKER_PROJECT"
	(
		cd "$WORKER_PROJECT"
		run_cmd go mod tidy
	)
}

exercise_worker_generators() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$WORKER_PROJECT"
		run_cmd ginger generate handler email
		run_cmd ginger generate service mailer
		run_cmd go mod tidy
		run_cmd go test -v ./...
		run_cmd go build ./...
		run_cmd ginger build
		run_cmd ginger doctor
	)

	assert_file "$WORKER_PROJECT/internal/worker/email_handler.go"
	assert_file "$WORKER_PROJECT/internal/services/mailer.go"
	assert_file "$WORKER_PROJECT/bin/$(basename "$WORKER_PROJECT")-worker"
}

run_worker_project() {
	export PATH="$BIN_DIR:$PATH"
	local prev_dir="$PWD"
	local run_log="$LOG_DIR/worker-run.out"
	cd "$WORKER_PROJECT"
	start_background "worker project" "$run_log" env HTTP_PORT="$WORKER_PORT" ginger run
	WORKER_PID="$LAST_BG_PID"
	WORKER_TAIL_PID="$LAST_TAIL_PID"
	wait_for_http "http://127.0.0.1:$WORKER_PORT/health" 30 1
	run_cmd_discard_output curl -fsS "http://127.0.0.1:$WORKER_PORT/health"
	stop_process "$WORKER_PID" "worker project" "$WORKER_TAIL_PID"
	WORKER_PID=""
	WORKER_TAIL_PID=""
	cd "$prev_dir"
}

create_cli_project() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$PLAYGROUND_DIR"
		run_cmd ginger new "$(basename "$CLI_PROJECT")" --cli
	)
	link_local_ginger "$CLI_PROJECT"
	(
		cd "$CLI_PROJECT"
		run_cmd go mod tidy
	)
}

exercise_cli_generators() {
	export PATH="$BIN_DIR:$PATH"
	(
		cd "$CLI_PROJECT"
		run_cmd ginger generate command sync
		run_cmd ginger generate service deployer
		run_cmd go mod tidy
		run_cmd go test -v ./...
		run_cmd go build ./...
		run_cmd ginger build
		run_cmd ginger run version
		run_cmd ginger run sync
		run_cmd "./bin/$(basename "$CLI_PROJECT")" version
		run_cmd "./bin/$(basename "$CLI_PROJECT")" sync
	)

	assert_file "$CLI_PROJECT/internal/commands/sync.go"
	assert_file "$CLI_PROJECT/internal/services/deployer.go"
	assert_file "$CLI_PROJECT/bin/$(basename "$CLI_PROJECT")"
}

print_summary() {
	local ok_count=0
	local fail_count=0
	local total_count="${#STEP_TITLES[@]}"
	local total_elapsed="$(( $(date +%s) - TOTAL_START_TS ))"
	local i

	printf '\n%s==================== FINAL REPORT ====================%s\n' "$C_SUM" "$C_RESET"
	for i in "${!STEP_TITLES[@]}"; do
		if [[ "${STEP_RESULTS[$i]}" == "OK" ]]; then
			ok_count=$((ok_count + 1))
			printf '%s[%02d]%s %s%-4s%s %s %s\n' \
				"$C_STEP" "$((i + 1))" "$C_RESET" \
				"$C_OK" "${STEP_RESULTS[$i]}" "$C_RESET" \
				"${STEP_DURATIONS[$i]}" "${STEP_TITLES[$i]}"
		else
			fail_count=$((fail_count + 1))
			printf '%s[%02d]%s %s%-4s%s %s %s\n' \
				"$C_STEP" "$((i + 1))" "$C_RESET" \
				"$C_FAIL" "${STEP_RESULTS[$i]}" "$C_RESET" \
				"${STEP_DURATIONS[$i]}" "${STEP_TITLES[$i]}"
		fi
		printf '     %slog:%s %s%s%s\n' "$C_LOG" "$C_RESET" "$C_DIM" "${STEP_LOGS[$i]}" "$C_RESET"
	done
	printf '%s--------------------------------------------------%s\n' "$C_SUM" "$C_RESET"
	printf '%sTotals:%s %sOK=%d%s %sFAIL=%d%s TOTAL=%d\n' \
		"$C_SUM" "$C_RESET" "$C_OK" "$ok_count" "$C_RESET" "$C_FAIL" "$fail_count" "$C_RESET" "$total_count"
	if [[ "$fail_count" -eq 0 ]]; then
		printf '%sOverall:%s %sOK%s\n' "$C_SUM" "$C_RESET" "$C_OK" "$C_RESET"
	else
		printf '%sOverall:%s %sFAIL%s\n' "$C_SUM" "$C_RESET" "$C_FAIL" "$C_RESET"
	fi
	printf '%sElapsed:%s %s%s%s\n' "$C_SUM" "$C_RESET" "$C_DIM" "$(format_duration "$total_elapsed")" "$C_RESET"
	printf '%sWorkspace:%s %s%s%s\n' "$C_LOG" "$C_RESET" "$C_DIM" "$WORKSPACE_DIR" "$C_RESET"
	printf '%sLogs:%s %s%s%s\n' "$C_LOG" "$C_RESET" "$C_DIM" "$LOG_DIR" "$C_RESET"
	printf '%sEnv helper:%s %s%s%s\n' "$C_LOG" "$C_RESET" "$C_DIM" "$ENV_FILE" "$C_RESET"
}

main() {
	require_commands

	SERVICE_PORT="$(choose_available_port "$SERVICE_PORT")"
	WORKER_PORT="$(choose_available_port "$WORKER_PORT")"

	note "Verbose mode: $VERBOSE"
	note "Color mode: $COLOR_ENABLED"
	note "Workspace base: $BASE_WORKSPACE_DIR"
	note "Repository: $REPO_URL"
	note "Checkout ref: $CHECKOUT_REF"
	note "Service port: $SERVICE_PORT"
	note "Worker port: $WORKER_PORT"
	note "The script will stream commands live and finish with a full report."

	run_step "Prepare isolated workspace" prepare_workspace
	run_step "Clone Ginger repository" clone_repo
	run_step "Install Ginger locally and export PATH" install_ginger
	run_step "Validate ginger version, help, and top-level commands" validate_cli_basics

	run_step "Create generic project and validate build/test" create_generic_project
	run_step "Run generic project with ginger run" run_generic_project

	run_step "Create service project" create_service_project
	run_step "Run every ginger add integration in the service project" exercise_service_integrations
	run_step "Run every service generator and validate doctor/build/test" exercise_service_generators
	run_step "Run service project and validate health, ping, and CRUD endpoints" run_service_project

	run_step "Create worker project" create_worker_project
	run_step "Run worker generators and validate doctor/build/test" exercise_worker_generators
	run_step "Run worker project and validate /health" run_worker_project

	run_step "Create CLI project" create_cli_project
	run_step "Run CLI generators and validate build/test/version/command" exercise_cli_generators

	print_summary
}

main "$@"
