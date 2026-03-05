#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RUN_WORKER="${RUN_WORKER_TESTS:-auto}"

log() {
  printf '[local-ci-smoke] %s\n' "$1"
}

main() {
  log "running dev script tests"
  "$ROOT_DIR/scripts/test-dev-scripts.sh"

  log "checking release policy drift"
  "$ROOT_DIR/scripts/check-release-policy-drift.sh"

  case "$RUN_WORKER" in
    true)
      log "running worker helper tests"
      "$ROOT_DIR/scripts/run-worker-tests.sh"
      ;;
    false)
      log "skipping worker helper tests because RUN_WORKER_TESTS=$RUN_WORKER"
      ;;
    auto)
      log "running worker helper tests"
      local worker_log
      worker_log="$(mktemp)"
      if "$ROOT_DIR/scripts/run-worker-tests.sh" >"$worker_log" 2>&1; then
        cat "$worker_log"
      elif grep -E "AccessDenied|operation not permitted|permission denied" "$worker_log" >/dev/null 2>&1; then
        log "skipping worker helper tests due to sandbox-style permission limits"
        tail -n 5 "$worker_log"
      else
        cat "$worker_log" >&2
        rm -f "$worker_log"
        exit 1
      fi
      rm -f "$worker_log"
      ;;
    *)
      echo "invalid RUN_WORKER_TESTS value: $RUN_WORKER" >&2
      exit 1
      ;;
  esac

  log "local ci smoke passed"
}

main "$@"
