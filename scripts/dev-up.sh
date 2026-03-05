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

  wait_for_url "$API_BASE_URL/health/live" "api-gateway"
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
