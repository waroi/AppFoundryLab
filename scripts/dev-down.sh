#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
MODE="${1:-standard}"
ENV_FILE="$ROOT_DIR/.env.docker.local"

usage() {
  cat <<'EOF'
usage: ./scripts/dev-down.sh [standard|security]
EOF
}

main() {
  case "$MODE" in
    standard|security) ;;
    -h|--help|help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac

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

  docker_compose_with_app_env_file "$ENV_FILE" "${compose_args[@]}" down
}

main "$@"
