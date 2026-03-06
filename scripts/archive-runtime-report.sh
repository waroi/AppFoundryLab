#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
API_BASE_URL="${DEPLOY_API_BASE_URL:-}"
ADMIN_USER="${DEPLOY_ADMIN_USER:-}"
ADMIN_PASSWORD="${DEPLOY_ADMIN_PASSWORD:-}"
OUT_DIR="${DEPLOY_REPORT_DIR:-$ROOT_DIR/artifacts/deploy-reports}"
PASSWORD_STDIN=false
POSITIONAL_ARGS=()

usage() {
  cat <<'EOF'
usage: ./scripts/archive-runtime-report.sh [--password-stdin] <api-base-url> <admin-user> [out-dir]

examples:
  DEPLOY_ADMIN_PASSWORD=strong_password ./scripts/archive-runtime-report.sh http://127.0.0.1:8080 admin
  printf '%s' "$ADMIN_PASSWORD" | ./scripts/archive-runtime-report.sh --password-stdin https://api.example.com admin ./artifacts/deploy-reports/prod

request log and incident event archives are redacted by default.
EOF
}

parse_args() {
  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      --password-stdin)
        PASSWORD_STDIN=true
        shift
        ;;
      -h|--help|help)
        usage
        exit 0
        ;;
      *)
        POSITIONAL_ARGS+=("$1")
        shift
        ;;
    esac
  done

  case "${#POSITIONAL_ARGS[@]}" in
    2)
      API_BASE_URL="${POSITIONAL_ARGS[0]}"
      ADMIN_USER="${POSITIONAL_ARGS[1]}"
      ;;
    3)
      API_BASE_URL="${POSITIONAL_ARGS[0]}"
      ADMIN_USER="${POSITIONAL_ARGS[1]}"
      OUT_DIR="${POSITIONAL_ARGS[2]}"
      ;;
    0)
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
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

read_admin_password_from_stdin() {
  if [[ "$PASSWORD_STDIN" != true ]]; then
    return 0
  fi

  IFS= read -r ADMIN_PASSWORD || true
}

sanitize_request_logs() {
  local input_file="$1"
  local output_file="$2"
  local redaction_mode="${ARCHIVE_REQUEST_LOG_MODE:-minimized}"

  python3 - "$input_file" "$output_file" "$redaction_mode" <<'PY'
import hashlib
import json
import pathlib
import sys
from urllib.parse import urlsplit

input_path = pathlib.Path(sys.argv[1])
output_path = pathlib.Path(sys.argv[2])
mode = sys.argv[3].strip() or "minimized"

payload = json.loads(input_path.read_text(encoding="utf-8"))
items = payload.get("items", [])

if mode == "raw":
    output_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    raise SystemExit(0)

sanitized = []
for item in items:
    raw_path = str(item.get("path", "") or "")
    split_path = urlsplit(raw_path)
    trace_id = str(item.get("traceId", "") or "")
    sanitized.append(
        {
            "path": split_path.path or raw_path,
            "method": item.get("method", ""),
            "durationMs": item.get("durationMs", 0),
            "statusCode": item.get("statusCode", 0),
            "occurredAt": item.get("occurredAt", ""),
            "traceIdPresent": bool(trace_id),
            "traceIdHash": hashlib.sha256(trace_id.encode("utf-8")).hexdigest()[:16] if trace_id else "",
        }
    )

output = {
    "schemaVersion": "runtime-request-log-archive-v1",
    "redaction": {
        "mode": "minimized",
        "removedFields": ["ip", "queryString", "traceId"],
        "sourceItemCount": len(items),
        "storedItemCount": len(sanitized),
    },
    "items": sanitized,
}
output_path.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")
PY
}

sanitize_incident_events() {
  local input_file="$1"
  local output_file="$2"
  local redaction_mode="${ARCHIVE_INCIDENT_EVENT_MODE:-minimized}"

  python3 - "$input_file" "$output_file" "$redaction_mode" <<'PY'
import hashlib
import json
import pathlib
import sys

input_path = pathlib.Path(sys.argv[1])
output_path = pathlib.Path(sys.argv[2])
mode = sys.argv[3].strip() or "minimized"

payload = json.loads(input_path.read_text(encoding="utf-8"))
items = payload.get("items", [])

if mode == "raw":
    output_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    raise SystemExit(0)

sanitized = []
for item in items:
    trace_id = str(item.get("traceId", "") or "")
    sanitized_item = dict(item)
    sanitized_item["traceIdPresent"] = bool(trace_id)
    sanitized_item["traceIdHash"] = hashlib.sha256(trace_id.encode("utf-8")).hexdigest()[:16] if trace_id else ""
    sanitized_item.pop("traceId", None)
    sanitized.append(sanitized_item)

output = {
    "schemaVersion": "runtime-incident-event-archive-v1",
    "redaction": {
        "mode": "minimized",
        "removedFields": ["traceId"],
        "sourceItemCount": len(items),
        "storedItemCount": len(sanitized),
    },
    "items": sanitized,
}
output_path.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")
PY
}

main() {
  parse_args "$@"
  read_admin_password_from_stdin

  if [[ -z "$API_BASE_URL" || -z "$ADMIN_USER" || -z "$ADMIN_PASSWORD" ]]; then
    usage >&2
    exit 1
  fi

  require_command curl
  require_command python3

  API_BASE_URL="${API_BASE_URL%/}"
  mkdir -p "$OUT_DIR"

  local timestamp login_payload token runtime_report incident_report incident_events incident_events_raw incident_event_mode request_logs request_logs_raw manifest request_log_limit request_log_mode
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  request_log_limit="${ARCHIVE_REQUEST_LOG_LIMIT:-5}"
  request_log_mode="${ARCHIVE_REQUEST_LOG_MODE:-minimized}"
  incident_event_mode="${ARCHIVE_INCIDENT_EVENT_MODE:-minimized}"
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
  incident_events_raw="$(mktemp)"
  request_logs_raw="$(mktemp)"
  manifest="$OUT_DIR/archive-manifest-$timestamp.txt"
  trap 'rm -f "$incident_events_raw" "$request_logs_raw"' RETURN

  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/runtime-report" > "$runtime_report"
  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/runtime-incident-report" > "$incident_report"
  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/incident-events" > "$incident_events_raw"
  curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL/api/v1/admin/request-logs?limit=$request_log_limit" > "$request_logs_raw"
  sanitize_incident_events "$incident_events_raw" "$incident_events"
  sanitize_request_logs "$request_logs_raw" "$request_logs"

  cat > "$manifest" <<EOF
archived_at=$timestamp
api_base_url=$API_BASE_URL
runtime_report=$(basename "$runtime_report")
runtime_incident_report=$(basename "$incident_report")
incident_events=$(basename "$incident_events")
request_logs=$(basename "$request_logs")
incident_events_mode=$incident_event_mode
request_logs_mode=$request_log_mode
request_logs_limit=$request_log_limit
runtime_report_sha256=$(sha256_file "$runtime_report")
runtime_incident_report_sha256=$(sha256_file "$incident_report")
incident_events_sha256=$(sha256_file "$incident_events")
request_logs_sha256=$(sha256_file "$request_logs")
EOF

  echo "runtime artifacts archived under: $OUT_DIR"
}

main "$@"
