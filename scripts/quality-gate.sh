#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:-sandbox-safe}"

log() {
  printf '[quality-gate] %s\n' "$1"
}

run_common_checks() {
  log "running toolchain governance"
  "$ROOT_DIR/scripts/check-toolchain.sh"

  log "running dev script regression suite"
  "$ROOT_DIR/scripts/test-dev-scripts.sh"

  log "running release policy drift check"
  "$ROOT_DIR/scripts/check-release-policy-drift.sh"
}

run_smoke_with_policy() {
  local worker_mode="$1"
  log "running local smoke chain with RUN_WORKER_TESTS=$worker_mode"
  RUN_WORKER_TESTS="$worker_mode" "$ROOT_DIR/scripts/local-ci-smoke.sh"
}

run_release_gate_fast() {
  log "running release gate (fast)"
  "$ROOT_DIR/scripts/release-gate.sh" fast
}

run_release_gate_full() {
  log "running release gate (full)"
  "$ROOT_DIR/scripts/release-gate.sh" full
}

main() {
  case "$MODE" in
    sandbox-safe)
      run_common_checks
      run_smoke_with_policy auto
      run_release_gate_fast
      ;;
    host-strict)
      run_common_checks
      run_smoke_with_policy true
      run_release_gate_fast
      ;;
    ci-fast)
      run_common_checks
      run_smoke_with_policy true
      ;;
    ci-full)
      run_common_checks
      run_smoke_with_policy true
      ;;
    *)
      echo "usage: ./scripts/quality-gate.sh [sandbox-safe|host-strict|ci-fast|ci-full]" >&2
      exit 1
      ;;
  esac

  log "mode '$MODE' passed"
}

main "$@"
