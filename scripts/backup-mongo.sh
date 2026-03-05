#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
OUT_DIR="${2:-$ROOT_DIR/artifacts/backups/mongo}"

usage() {
  cat <<'EOF'
usage: ./scripts/backup-mongo.sh [env-file] [out-dir]
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

  local user password db timestamp archive
  user="$(read_env_value "$ENV_FILE" "MONGO_INITDB_ROOT_USERNAME")"
  password="$(read_env_value "$ENV_FILE" "MONGO_INITDB_ROOT_PASSWORD")"
  db="$(read_env_value "$ENV_FILE" "MONGO_DB")"
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  archive="$OUT_DIR/mongo-$timestamp.archive.gz"

  single_host_compose "$ENV_FILE" exec -T mongo \
    mongodump \
      --gzip \
      --archive \
      --host localhost \
      --port 27017 \
      --username "$user" \
      --password "$password" \
      --authenticationDatabase admin \
      --db "$db" > "$archive"

  echo "mongo backup created: $archive"
}

main "$@"
