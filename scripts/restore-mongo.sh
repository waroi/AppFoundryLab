#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
ARCHIVE_FILE="${2:-}"
DROP_FLAG="${3:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/restore-mongo.sh [env-file] <archive-file> [--drop]
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
  if [[ -z "$ARCHIVE_FILE" || ! -f "$ARCHIVE_FILE" ]]; then
    echo "archive file not found: $ARCHIVE_FILE" >&2
    exit 1
  fi

  local user password db drop_arg
  user="$(read_env_value "$ENV_FILE" "MONGO_INITDB_ROOT_USERNAME")"
  password="$(read_env_value "$ENV_FILE" "MONGO_INITDB_ROOT_PASSWORD")"
  db="$(read_env_value "$ENV_FILE" "MONGO_DB")"
  drop_arg=""
  if [[ "$DROP_FLAG" == "--drop" ]]; then
    drop_arg="--drop"
  fi

  single_host_compose "$ENV_FILE" exec -T mongo \
    mongorestore \
      --gzip \
      --archive \
      ${drop_arg:+$drop_arg} \
      --host localhost \
      --port 27017 \
      --username "$user" \
      --password "$password" \
      --authenticationDatabase admin \
      --nsInclude "$db.*" < "$ARCHIVE_FILE"

  echo "mongo restore completed from: $ARCHIVE_FILE"
}

main "$@"
