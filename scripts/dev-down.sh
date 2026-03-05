#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
MODE="${1:-standard}"
ENV_FILE="$ROOT_DIR/.env.docker.local"
REMOVE_VOLUMES="false"

usage() {
  cat <<'EOF'
usage: ./scripts/dev-down.sh [standard|security] [--volumes]
EOF
}

parse_args() {
  for arg in "$@"; do
    case "$arg" in
      standard|security)
        MODE="$arg"
        ;;
      --volumes|-v)
        REMOVE_VOLUMES="true"
        ;;
      -h|--help|help)
        usage
        exit 0
        ;;
      *)
        usage >&2
        exit 1
        ;;
    esac
  done
}

main() {
  parse_args "$@"

  if [[ ! -f "$ENV_FILE" ]]; then
    echo "env file not found: $ENV_FILE" >&2
    exit 1
  fi

  local compose_args=(
    -f "$ROOT_DIR/docker-compose.yml"
  )

  if [[ "$MODE" == "security" ]]; then
    compose_args+=(-f "$ROOT_DIR/docker-compose.security.yml")
  fi

  local down_args=(down)
  if [[ "$REMOVE_VOLUMES" == "true" ]]; then
    down_args+=(--volumes)
  fi

  docker_compose_with_app_env_file "$ENV_FILE" "${compose_args[@]}" "${down_args[@]}"
}

main "$@"
