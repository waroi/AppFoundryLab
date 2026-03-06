#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
PROFILE="${1:-standard}"
MODE="${2:-standard}"
ENV_FILE="$ROOT_DIR/.env.docker.local"
HTTP_CURL_BIN=(curl)
FRONTEND_BASE_URL="http://127.0.0.1:4321"
API_BASE_URL="http://127.0.0.1:8080"
LOGGER_BASE_URL="http://127.0.0.1:8090"
LOGGER_METRICS_URL="http://127.0.0.1:8090/metrics"

usage() {
  cat <<'EOF'
usage: ./scripts/dev-up.sh [minimal|standard|secure] [standard|security]

examples:
  ./scripts/dev-up.sh
  ./scripts/dev-up.sh standard
  ./scripts/dev-up.sh secure security
EOF
}

wait_for_url() {
  local url="$1"
  local label="$2"

  for _ in $(seq 1 90); do
    if "${HTTP_CURL_BIN[@]}" -fsS "$url" >/dev/null 2>&1; then
      echo "$label is ready: $url"
      return 0
    fi
    sleep 2
  done

  echo "$label did not become ready in time: $url" >&2
  return 1
}

json_value_from_file() {
  local file="$1"
  local key="$2"
  python3 - "$file" "$key" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
value = payload.get(sys.argv[2], "")
if isinstance(value, str):
    print(value)
PY
}

login_payload() {
  local username="$1"
  local password="$2"
  python3 - "$username" "$password" <<'PY'
import json
import sys

print(json.dumps({"username": sys.argv[1], "password": sys.argv[2]}))
PY
}

auth_value_or_default() {
  local key="$1"
  local fallback="$2"
  local value
  value="${!key:-$(read_env_value "$ENV_FILE" "$key")}"
  if [[ -n "$value" ]]; then
    printf '%s\n' "$value"
    return
  fi
  printf '%s\n' "$fallback"
}

fetch_admin_token() {
  local admin_user admin_password response_file token
  admin_user="$(auth_value_or_default "BOOTSTRAP_ADMIN_USER" "admin")"
  admin_password="$(auth_value_or_default "BOOTSTRAP_ADMIN_PASSWORD" "admin_dev_password")"

  for _ in $(seq 1 10); do
    response_file="$(mktemp)"
    if "${HTTP_CURL_BIN[@]}" -fsS \
      -H 'Content-Type: application/json' \
      -X POST \
      --data "$(login_payload "$admin_user" "$admin_password")" \
      "$API_BASE_URL/api/v1/auth/token" >"$response_file" 2>/dev/null; then
      token="$(json_value_from_file "$response_file" "accessToken")"
      rm -f "$response_file"
      if [[ -n "$token" ]]; then
        printf '%s\n' "$token"
        return 0
      fi
    else
      rm -f "$response_file"
    fi
    sleep 2
  done

  echo "failed to obtain admin token from $API_BASE_URL/api/v1/auth/token" >&2
  return 1
}

wait_for_authorized_url() {
  local url="$1"
  local label="$2"
  local token="$3"

  for _ in $(seq 1 30); do
    if "${HTTP_CURL_BIN[@]}" -fsS -H "Authorization: Bearer $token" "$url" >/dev/null 2>&1; then
      echo "$label is ready: $url"
      return 0
    fi
    sleep 2
  done

  echo "$label did not become ready in time: $url" >&2
  return 1
}

check_postgres_auth() {
  local compose_args=("$@")
  set +e
  docker_compose_with_app_env_file "$ENV_FILE" "${compose_args[@]}" exec -T postgres \
    sh -lc 'PGPASSWORD="${POSTGRES_PASSWORD}" psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT 1" >/dev/null' \
    >/dev/null 2>&1
  local status=$?
  set -e
  return "$status"
}

check_mongo_auth() {
  local compose_args=("$@")
  set +e
  docker_compose_with_app_env_file "$ENV_FILE" "${compose_args[@]}" exec -T mongo \
    sh -lc 'mongosh --quiet --username "${MONGO_INITDB_ROOT_USERNAME}" --password "${MONGO_INITDB_ROOT_PASSWORD}" --authenticationDatabase admin --eval "db.adminCommand({ ping: 1 }).ok" >/dev/null' \
    >/dev/null 2>&1
  local status=$?
  set -e
  return "$status"
}

wait_for_data_credential_checks() {
  local mode="$1"
  shift
  local compose_args=("$@")
  local postgres_ok="false"
  local mongo_ok="false"

  if is_truthy "${SKIP_DATA_CREDENTIAL_CHECK:-false}"; then
    echo "data credential check skipped: SKIP_DATA_CREDENTIAL_CHECK=true"
    return 0
  fi

  for _ in $(seq 1 60); do
    if check_postgres_auth "${compose_args[@]}"; then
      postgres_ok="true"
    fi
    if check_mongo_auth "${compose_args[@]}"; then
      mongo_ok="true"
    fi
    if [[ "$postgres_ok" == "true" && "$mongo_ok" == "true" ]]; then
      echo "data credential check passed"
      return 0
    fi
    sleep 2
  done

  echo "data credential check failed for persisted services (postgres_ok=$postgres_ok mongo_ok=$mongo_ok)" >&2
  cat >&2 <<EOF
possible cause: existing local database volumes were initialized with different credentials.
fix options:
  1) align POSTGRES_PASSWORD and MONGO_INITDB_ROOT_PASSWORD in .env.docker.local with existing volume credentials
  2) recreate local volumes if data can be reset:
     ./scripts/dev-down.sh $mode --volumes

temporary bypass (not recommended): SKIP_DATA_CREDENTIAL_CHECK=true ./scripts/dev-up.sh <profile> <mode>
EOF
  return 1
}

env_value_or_default() {
  local key="$1"
  local fallback="$2"
  local value
  value="${!key:-}"
  if [[ -n "$value" ]]; then
    printf '%s\n' "$value"
    return
  fi
  value="$(read_env_value "$ENV_FILE" "$key")"
  if [[ -n "$value" ]]; then
    printf '%s\n' "$value"
    return
  fi
  printf '%s\n' "$fallback"
}

browser_host_for_bind_address() {
  local bind_address="$1"
  case "$bind_address" in
    ""|0.0.0.0|::|"[::]")
      printf '%s\n' "127.0.0.1"
      ;;
    *)
      printf '%s\n' "$bind_address"
      ;;
  esac
}

default_local_allowed_origins() {
  local frontend_port="$1"
  printf 'http://localhost:%s,http://127.0.0.1:%s\n' "$frontend_port" "$frontend_port"
}

resolve_public_api_base_url() {
  local api_base_url="$1"
  local configured_value
  configured_value="${PUBLIC_API_BASE_URL:-$(read_env_value "$ENV_FILE" "PUBLIC_API_BASE_URL")}"
  case "$configured_value" in
    ""|"http://localhost:8080"|"http://127.0.0.1:8080")
      printf '%s\n' "$api_base_url"
      ;;
    *)
      printf '%s\n' "$configured_value"
      ;;
  esac
}

resolve_allowed_origins() {
  local frontend_port="$1"
  local configured_value
  configured_value="${ALLOWED_ORIGINS:-$(read_env_value "$ENV_FILE" "ALLOWED_ORIGINS")}"
  case "$configured_value" in
    ""|"http://localhost:4321,http://127.0.0.1:4321")
      default_local_allowed_origins "$frontend_port"
      ;;
    *)
      printf '%s\n' "$configured_value"
      ;;
  esac
}

configure_runtime_endpoints() {
  local bind_address browser_host frontend_host_port api_host_port logger_host_port
  bind_address="$(env_value_or_default "DOCKER_HOST_BIND_ADDRESS" "127.0.0.1")"
  browser_host="$(browser_host_for_bind_address "$bind_address")"
  frontend_host_port="$(env_value_or_default "FRONTEND_HOST_PORT" "4321")"
  api_host_port="$(env_value_or_default "API_GATEWAY_HOST_PORT" "8080")"
  logger_host_port="$(env_value_or_default "LOGGER_HOST_PORT" "8090")"

  FRONTEND_BASE_URL="http://${browser_host}:${frontend_host_port}"
  API_BASE_URL="http://${browser_host}:${api_host_port}"
  LOGGER_BASE_URL="http://${browser_host}:${logger_host_port}"
  LOGGER_METRICS_URL="http://${browser_host}:${logger_host_port}/metrics"

  export PUBLIC_API_BASE_URL
  PUBLIC_API_BASE_URL="$(resolve_public_api_base_url "$API_BASE_URL")"
  export ALLOWED_ORIGINS
  ALLOWED_ORIGINS="$(resolve_allowed_origins "$frontend_host_port")"
}

configure_http_client() {
  local docker_bin curl_exe_bin
  docker_bin="$(docker_bin_path)"
  curl_exe_bin="${CURL_EXE_BIN:-/mnt/c/Windows/System32/curl.exe}"
  if is_windows_docker_bin "$docker_bin" && [[ -x "$curl_exe_bin" ]]; then
    HTTP_CURL_BIN=("$curl_exe_bin")
    return
  fi
  HTTP_CURL_BIN=(curl)
}

main() {
  case "$PROFILE" in
    minimal|standard|secure) ;;
    -h|--help|help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac

  case "$MODE" in
    standard|security) ;;
    *)
      usage >&2
      exit 1
      ;;
  esac

  if [[ ! -f "$ROOT_DIR/.env" || ! -f "$ENV_FILE" ]]; then
    "$ROOT_DIR/scripts/bootstrap.sh" "$PROFILE"
  fi

  if [[ ! -f "$ROOT_DIR/backend/infrastructure/certs/dev/ca.crt" ]]; then
    "$ROOT_DIR/scripts/certs-dev.sh"
  fi

  configure_runtime_endpoints
  configure_http_client

  local compose_args=(
    -f "$ROOT_DIR/docker-compose.yml"
  )

  if [[ "$MODE" == "security" ]]; then
    compose_args+=(-f "$ROOT_DIR/docker-compose.security.yml")
  fi

  docker_compose_with_app_env_file "$ENV_FILE" "${compose_args[@]}" up --build -d

  wait_for_data_credential_checks "$MODE" "${compose_args[@]}"
  wait_for_url "$API_BASE_URL/health/ready" "api-gateway"
  wait_for_url "$LOGGER_BASE_URL/health" "logger"
  wait_for_url "$LOGGER_METRICS_URL" "logger metrics"
  local admin_token
  admin_token="$(fetch_admin_token)"
  wait_for_authorized_url "$API_BASE_URL/api/v1/admin/runtime-report" "admin runtime report" "$admin_token"
  wait_for_url "$FRONTEND_BASE_URL/healthz" "frontend health"
  wait_for_url "$FRONTEND_BASE_URL" "frontend"

  cat <<EOF

stack started
profile: $PROFILE
mode: $MODE
frontend: $FRONTEND_BASE_URL
api: $API_BASE_URL
logger metrics: $LOGGER_METRICS_URL
stop with: ./scripts/dev-down.sh $MODE
EOF
}

main "$@"
