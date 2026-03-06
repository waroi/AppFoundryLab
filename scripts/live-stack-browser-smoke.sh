#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:-standard}"
BUN_BIN="${BUN_BIN:-$ROOT_DIR/.toolchain/bun/bin/bun}"
KEEP_STACK="${KEEP_LIVE_STACK_SMOKE_STACK:-false}"

cleanup() {
  if [[ "$KEEP_STACK" == "true" ]]; then
    return
  fi
  "$ROOT_DIR/scripts/dev-down.sh" "$MODE" --volumes >/dev/null 2>&1 || true
}

trap cleanup EXIT

if [[ ! -f "$ROOT_DIR/.env.docker.local" ]]; then
  "$ROOT_DIR/scripts/bootstrap.sh" "$MODE" --force >/dev/null
fi

"$ROOT_DIR/scripts/dev-up.sh" "$MODE"
(
  cd "$ROOT_DIR/frontend"
  "$BUN_BIN" run e2e:live
)
