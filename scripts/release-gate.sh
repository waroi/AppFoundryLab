#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:-fast}"
JSON_OUT=""
CHECKLIST_JSON="$ROOT_DIR/docs/release-checklist.json"

if [[ "$#" -ge 2 ]]; then
  if [[ "$2" == "--json" && -n "${3:-}" ]]; then
    JSON_OUT="$3"
  else
    echo "usage: ./scripts/release-gate.sh [fast|full] [--json <path>]"
    exit 1
  fi
fi

if [[ "$MODE" != "fast" && "$MODE" != "full" ]]; then
  echo "usage: ./scripts/release-gate.sh [fast|full] [--json <path>]"
  exit 1
fi

log() {
  echo "[release-gate] $*"
}

require_file() {
  local file="$1"
  if [[ ! -f "$ROOT_DIR/$file" ]]; then
    echo "missing required file: $file" >&2
    exit 1
  fi
}

check_static_policy() {
  log "checking static policy files"
  require_file "README.md"
  require_file "CHANGELOG.md"
  require_file "docs/gelistirmePlanı.md"
  require_file "docs/appfoundrylab-teknik-analiz.md"
  require_file "docs/release-checklist.json"
  require_file "scripts/perf/compare_k6_summary.py"
  require_file "scripts/perf/baseline/k6-summary.json"

  if ! grep -q "## \\[Unreleased\\]" "$ROOT_DIR/CHANGELOG.md"; then
    echo "CHANGELOG.md missing [Unreleased] section" >&2
    exit 1
  fi

  if ! grep -q "Sonraki aktif hedef:" "$ROOT_DIR/docs/gelistirmePlanı.md"; then
    echo "docs/gelistirmePlanı.md missing 'Sonraki aktif hedef' line" >&2
    exit 1
  fi
}

json_array_from_checklist() {
  local key="$1"
  python3 - "$CHECKLIST_JSON" "$key" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
key = sys.argv[2]
data = json.loads(path.read_text(encoding="utf-8"))
rg = data.get("releaseGate", {})
arr = rg.get(key)
if not isinstance(arr, list) or not arr:
    raise SystemExit(f"invalid release checklist json: releaseGate.{key}")
print(json.dumps(arr))
PY
}

check_api_contract_guard() {
  log "checking api versioning guard"
  if rg -n '"/api/auth/token"|"/api/users"|"/api/compute/fibonacci"|"/api/compute/hash"' \
    "$ROOT_DIR/frontend/src/components/Interactive/SystemStatus.svelte" >/dev/null; then
    echo "legacy /api path detected in frontend SystemStatus" >&2
    exit 1
  fi
}

run_fast_checks() {
  log "running toolchain governance"
  "$ROOT_DIR/scripts/check-toolchain.sh"
  log "running release policy drift check"
  "$ROOT_DIR/scripts/check-release-policy-drift.sh"
  check_static_policy
  check_api_contract_guard
}

run_full_checks() {
  run_fast_checks

  log "running go tests"
  "$ROOT_DIR/scripts/go-test.sh"

  log "running rust tests"
  (
    cd "$ROOT_DIR/backend/core/calculator"
    cargo test
  )

  log "running frontend build+smoke"
  (
    cd "$ROOT_DIR/frontend"
    bun run build
    bun run smoke
  )
}

if [[ "$MODE" == "full" ]]; then
  run_full_checks
else
  run_fast_checks
fi

if [[ -n "$JSON_OUT" ]]; then
  automated_checks_json="$(json_array_from_checklist automatedChecks)"
  manual_checks_json="$(json_array_from_checklist manualChecks)"
  cat >"$JSON_OUT" <<EOF
{
  "status": "passed",
  "mode": "$MODE",
  "automatedChecks": $automated_checks_json,
  "manualChecks": $manual_checks_json
}
EOF
fi

cat <<'EOF'
[release-gate] automated checks passed
[release-gate] manual checks still required:
  - CI required jobs green
  - Trivy/Gitleaks findings review
  - Dependabot queue review
  - Perf benchmark trend diff review
  - Release evidence audit review
  - Backup lifecycle drift review
EOF
