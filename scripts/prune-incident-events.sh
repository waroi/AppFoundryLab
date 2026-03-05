#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
DAYS="${2:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/prune-incident-events.sh [env-file] [retention-days]
EOF
}

read_env() {
  local file="$1"
  local key="$2"
  awk -F= -v key="$key" '$1 == key { print substr($0, index($0, "=") + 1); exit }' "$file"
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

  local user password db collection retention_days cutoff
  user="$(read_env "$ENV_FILE" "MONGO_INITDB_ROOT_USERNAME")"
  password="$(read_env "$ENV_FILE" "MONGO_INITDB_ROOT_PASSWORD")"
  db="$(read_env "$ENV_FILE" "MONGO_DB")"
  collection="$(read_env "$ENV_FILE" "MONGO_INCIDENT_COLLECTION")"
  retention_days="${DAYS:-$(read_env "$ENV_FILE" "INCIDENT_EVENT_RETENTION_DAYS")}"
  cutoff="$(date -u -d "-${retention_days} days" +%Y-%m-%dT%H:%M:%SZ)"

  local compose_env_file
  compose_env_file="$(docker_compose_runtime_env_file "$ENV_FILE")"

  local status
  set +e
  docker_cli compose \
    -f "$ROOT_DIR/docker-compose.yml" \
    -f "$ROOT_DIR/docker-compose.security.yml" \
    -f "$ROOT_DIR/docker-compose.single-host.yml" \
    --env-file "$compose_env_file" \
    exec -T mongo sh -lc \
    "mongosh --quiet 'mongodb://$user:$password@localhost:27017/admin' --eval \"db.getSiblingDB('$db').getCollection('$collection').deleteMany({ lastSeenAt: { \\\$lt: '$cutoff' } })\""
  status=$?
  set -e

  if [[ "$compose_env_file" != "$ENV_FILE" ]]; then
    rm -f "$compose_env_file"
  fi
  if [[ "$status" -ne 0 ]]; then
    return "$status"
  fi

  echo "incident events older than $retention_days day(s) pruned"
}

main "$@"
