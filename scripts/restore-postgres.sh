#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
SQL_FILE="${2:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/restore-postgres.sh [env-file] <sql-file>
EOF
}

main() {
  if [[ "${ENV_FILE:-}" == "--help" || "${ENV_FILE:-}" == "-h" ]]; then
    usage
    exit 0
  fi
  if [[ ! -f "$ENV_FILE" ]]; then
    echo "env file not found: $ENV_FILE" >&2
    exit 1
  fi
  if [[ -z "$SQL_FILE" || ! -f "$SQL_FILE" ]]; then
    echo "sql file not found: $SQL_FILE" >&2
    exit 1
  fi

  local db user password
  db="$(read_env_value "$ENV_FILE" "POSTGRES_DB")"
  user="$(read_env_value "$ENV_FILE" "POSTGRES_USER")"
  password="$(read_env_value "$ENV_FILE" "POSTGRES_PASSWORD")"

  single_host_compose "$ENV_FILE" exec -T \
    -e PGPASSWORD="$password" \
    postgres psql -v ON_ERROR_STOP=1 -U "$user" -d "$db" < "$SQL_FILE"

  echo "postgres restore completed from: $SQL_FILE"
}

main "$@"
