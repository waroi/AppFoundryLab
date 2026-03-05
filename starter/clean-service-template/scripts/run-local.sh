#!/usr/bin/env bash
set -euo pipefail

# Minimal local process-mode runner for services created from this template.
# Override APP_CMD to match your stack.
#
# Examples:
#   APP_CMD="./app" PORT=8080 ./scripts/run-local.sh
#   APP_CMD="go run ./cmd/service" PORT=8080 ./scripts/run-local.sh

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${PORT:-8080}"
APP_CMD="${APP_CMD:-./app}"
HOST="${HOST:-127.0.0.1}"

cd "$ROOT_DIR"

echo "starting local service on ${HOST}:${PORT}"
echo "command: ${APP_CMD}"

export PORT
export HOST

exec bash -lc "$APP_CMD"
