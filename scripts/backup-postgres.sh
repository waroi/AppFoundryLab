#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
OUT_DIR="${2:-$ROOT_DIR/artifacts/backups/postgres}"

usage() {
  cat <<'EOF'
usage: ./scripts/backup-postgres.sh [env-file] [out-dir]
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

  mkdir -p "$OUT_DIR"

  local db user password timestamp outfile
  db="$(read_env_value "$ENV_FILE" "POSTGRES_DB")"
  user="$(read_env_value "$ENV_FILE" "POSTGRES_USER")"
  password="$(read_env_value "$ENV_FILE" "POSTGRES_PASSWORD")"
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  outfile="$OUT_DIR/postgres-$timestamp.sql"

  single_host_compose "$ENV_FILE" exec -T \
    -e PGPASSWORD="$password" \
    postgres pg_dump -U "$user" -d "$db" > "$outfile"

  echo "postgres backup created: $outfile"
}

main "$@"
