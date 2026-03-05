#!/usr/bin/env bash
set -euo pipefail

# Process-mode integration smoke wrapper for services created from this template.
# Example:
#   APP_CMD="go run ./cmd/service" \
#   HEALTH_PATH=/health \
#   READY_PATH=/health/ready \
#   ./tests/integration/process-mode-smoke.sh

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
HEALTH_PATH="${HEALTH_PATH:-/health}"
READY_PATH="${READY_PATH:-/health/ready}"
STARTUP_TIMEOUT_SECONDS="${STARTUP_TIMEOUT_SECONDS:-30}"
APP_CMD="${APP_CMD:-./app}"
RUN_LOAD_SHED_SMOKE="${RUN_LOAD_SHED_SMOKE:-false}"

cleanup() {
  local exit_code=$?
  if [ -n "${RUNNER_PID:-}" ] && kill -0 "$RUNNER_PID" >/dev/null 2>&1; then
    kill "$RUNNER_PID" >/dev/null 2>&1 || true
    wait "$RUNNER_PID" >/dev/null 2>&1 || true
  fi
  exit $exit_code
}
trap cleanup EXIT

(
  cd "$ROOT_DIR"
  APP_CMD="$APP_CMD" ./scripts/run-local.sh
) &
RUNNER_PID=$!

for i in $(seq 1 "$STARTUP_TIMEOUT_SECONDS"); do
  if curl -fsS "${BASE_URL}${HEALTH_PATH}" >/dev/null; then
    break
  fi
  if [ "$i" -eq "$STARTUP_TIMEOUT_SECONDS" ]; then
    echo "service did not become healthy in time" >&2
    exit 1
  fi
  sleep 1
done

(
  cd "$ROOT_DIR"
  BASE_URL="$BASE_URL" \
  HEALTH_PATH="$HEALTH_PATH" \
  READY_PATH="$READY_PATH" \
  ./tests/integration/smoke-http.sh
)

if [ "$RUN_LOAD_SHED_SMOKE" = "true" ]; then
  (
    cd "$ROOT_DIR"
    BASE_URL="$BASE_URL" \
    HEALTH_PATH="$HEALTH_PATH" \
    OVERLOAD_PATH="${OVERLOAD_PATH:-/internal/test/overload}" \
    ./tests/integration/load-shed-smoke.sh
  )
fi

echo "process-mode integration smoke passed"
