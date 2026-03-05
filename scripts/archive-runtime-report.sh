#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
API_BASE_URL="${1:-${DEPLOY_API_BASE_URL:-}}"
ADMIN_USER="${2:-${DEPLOY_ADMIN_USER:-}}"
ADMIN_PASSWORD="${3:-${DEPLOY_ADMIN_PASSWORD:-}}"
OUT_DIR="${4:-${DEPLOY_REPORT_DIR:-$ROOT_DIR/artifacts/deploy-reports}}"

usage() {
  cat <<'EOF'
usage: ./scripts/archive-runtime-report.sh <api-base-url> <admin-user> <admin-password> [out-dir]

examples:
  ./scripts/archive-runtime-report.sh http://127.0.0.1:8080 admin strong_password
  ./scripts/archive-runtime-report.sh https://api.example.com admin strong_password ./artifacts/deploy-reports/prod
EOF
}

build_login_payload() {
  python3 - "$ADMIN_USER" "$ADMIN_PASSWORD" <<'PY'
import json
import sys

print(json.dumps({"username": sys.argv[1], "password": sys.argv[2]}))
PY
}

extract_token() {
  python3 -c '
import json
import sys

payload = json.load(sys.stdin)
token = payload.get("accessToken", "")
if not token:
    raise SystemExit("accessToken missing in auth response")
print(token)
'
}

main() {
  if [[ -z "$API_BASE_URL" || -z "$ADMIN_USER" || -z "$ADMIN_PASSWORD" ]]; then
    usage >&2
    exit 1
  fi

  require_command curl
  require_command python3

  API_BASE_URL="${API_BASE_URL%/}"
  mkdir -p "$OUT_DIR"

  local timestamp login_payload token runtime_report incident_report incident_events request_logs manifest
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  login_payload="$(build_login_payload)"
  token="$(
    curl -fsS \
      -H 'Content-Type: application/json' \
      -X POST \
      --data "$login_payload" \
      "$API_BASE_URL/api/v1/auth/token" | extract_token
  )"

  runtime_report="$OUT_DIR/runtime-report-$timestamp.json"
  incident_report="$OUT_DIR/runtime-incident-report-$timestamp.json"
  incident_events="$OUT_DIR/incident-events-$timestamp.json"
  request_logs="$OUT_DIR/request-logs-$timestamp.json"
  manifest="$OUT_DIR/archive-manifest-$timestamp.txt"

  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/runtime-report" > "$runtime_report"
  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/runtime-incident-report" > "$incident_report"
  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/incident-events" > "$incident_events"
  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/request-logs?limit=20" > "$request_logs"

  cat > "$manifest" <<EOF
archived_at=$timestamp
api_base_url=$API_BASE_URL
runtime_report=$(basename "$runtime_report")
runtime_incident_report=$(basename "$incident_report")
incident_events=$(basename "$incident_events")
request_logs=$(basename "$request_logs")
runtime_report_sha256=$(sha256_file "$runtime_report")
runtime_incident_report_sha256=$(sha256_file "$incident_report")
incident_events_sha256=$(sha256_file "$incident_events")
request_logs_sha256=$(sha256_file "$request_logs")
EOF

  echo "runtime artifacts archived under: $OUT_DIR"
}

main "$@"
