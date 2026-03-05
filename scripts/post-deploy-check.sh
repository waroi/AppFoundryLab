#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"

FRONTEND_BASE_URL="${1:-http://127.0.0.1:4321}"
API_BASE_URL="${2:-http://127.0.0.1:8080}"
ENV_FILE="${3:-${DEPLOY_ENV_FILE:-$ROOT_DIR/.env.single-host}}"

login_payload() {
  python3 - "$DEPLOY_ADMIN_USER" "$DEPLOY_ADMIN_PASSWORD" <<'PY'
import json
import sys
print(json.dumps({"username": sys.argv[1], "password": sys.argv[2]}))
PY
}

extract_token() {
  local response_file="$1"
  python3 - "$response_file" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
token = payload.get("accessToken")
if not token:
    raise SystemExit(1)
print(token)
PY
}

fetch_admin_token() {
  local auth_url response_file token attempts delay
  auth_url="$API_BASE_URL/api/v1/auth/token"
  attempts="${POST_DEPLOY_AUTH_RETRIES:-10}"
  delay="${POST_DEPLOY_AUTH_RETRY_DELAY_SECONDS:-2}"

  for _ in $(seq 1 "$attempts"); do
    response_file="$(mktemp)"
    if curl -fsS \
      -H 'Content-Type: application/json' \
      -X POST \
      --data "$(login_payload)" \
      "$auth_url" >"$response_file" 2>/dev/null; then
      if token="$(extract_token "$response_file" 2>/dev/null)"; then
        rm -f "$response_file"
        printf '%s\n' "$token"
        return 0
      fi
    fi
    rm -f "$response_file"
    sleep "$delay"
  done

  echo "post-deploy check failed: unable to obtain admin token from $auth_url" >&2
  return 1
}

check_url() {
  local url="$1"
  local label="$2"
  if ! curl -fsS "$url" >/dev/null; then
    echo "post-deploy check failed: $label -> $url" >&2
    exit 1
  fi
  echo "post-deploy check ok: $label -> $url"
}

check_admin_endpoint() {
  local token="$1"
  local path="$2"
  local label="$3"
  if ! curl -fsS -H "Authorization: Bearer $token" "$API_BASE_URL$path" >/dev/null; then
    echo "post-deploy check failed: $label -> $path" >&2
    exit 1
  fi
  echo "post-deploy check ok: $label -> $path"
}

check_url "$FRONTEND_BASE_URL/healthz" "frontend health"
check_url "$FRONTEND_BASE_URL" "frontend root"
check_url "$API_BASE_URL/health/live" "api live"
check_url "$API_BASE_URL/health/ready" "api ready"

if [[ -f "$ENV_FILE" ]]; then
  single_host_exec "$ENV_FILE" logger wget -q -O - http://127.0.0.1:8090/health >/dev/null
  echo "post-deploy check ok: logger health"
  single_host_exec "$ENV_FILE" logger wget -q -O - http://127.0.0.1:8090/metrics >/dev/null
  echo "post-deploy check ok: logger json metrics"
  single_host_exec "$ENV_FILE" logger wget -q -O - http://127.0.0.1:8090/metrics/prometheus >/dev/null
  echo "post-deploy check ok: logger prometheus metrics"
fi

if [[ -n "${DEPLOY_ADMIN_USER:-}" && -n "${DEPLOY_ADMIN_PASSWORD:-}" ]]; then
  token="$(fetch_admin_token)"
  check_admin_endpoint "$token" "/api/v1/admin/runtime-report" "admin runtime report"
  check_admin_endpoint "$token" "/api/v1/admin/runtime-incident-report" "admin incident report"
  check_admin_endpoint "$token" "/api/v1/admin/incident-events" "admin incident events"
  check_admin_endpoint "$token" "/api/v1/admin/request-logs?limit=5" "admin request logs"
fi

if [[ -n "${DEPLOY_REPORT_DIR:-}" ]]; then
  latest_manifest="$(latest_matching_file "$DEPLOY_REPORT_DIR" 'archive-manifest-*.txt')"
  if [[ -z "$latest_manifest" ]]; then
    echo "post-deploy check failed: deploy report manifest missing in $DEPLOY_REPORT_DIR" >&2
    exit 1
  fi
  echo "post-deploy check ok: deploy report manifest -> $latest_manifest"
fi

echo "post-deploy checks passed"
