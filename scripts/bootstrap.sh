#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PROFILE="standard"
FORCE="false"

usage() {
  cat <<'EOF'
usage: ./scripts/bootstrap.sh [minimal|standard|secure] [--force]

notes:
  - creates .env and .env.docker.local
  - generates random local-only service secrets
  - refuses to overwrite existing env files unless --force is passed
EOF
}

replace_env_value() {
  local file="$1"
  local key="$2"
  local value="$3"
  local tmp
  tmp="$(mktemp)"
  awk -F= -v key="$key" -v value="$value" '
    BEGIN { replaced = 0 }
    $1 == key {
      print key "=" value
      replaced = 1
      next
    }
    { print $0 }
    END {
      if (replaced == 0) {
        print key "=" value
      }
    }
  ' "$file" > "$tmp"
  mv "$tmp" "$file"
}

read_env_value() {
  local file="$1"
  local key="$2"
  awk -F= -v key="$key" '
    $1 == key {
      value = substr($0, index($0, "=") + 1)
    }
    END {
      print value
    }
  ' "$file"
}

generate_secret() {
  openssl rand -hex 24
}

apply_env_overrides_from_file() {
  local source_file="$1"
  local target_file="$2"
  local line key value

  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -n "$line" ]] || continue
    [[ "$line" != \#* ]] || continue
    [[ "$line" == *=* ]] || continue
    key="${line%%=*}"
    value="${line#*=}"
    replace_env_value "$target_file" "$key" "$value"
  done <"$source_file"
}

resolve_local_auth_mode() {
  case "$PROFILE" in
    minimal)
      echo "demo"
      ;;
    secure|standard)
      echo "generated"
      ;;
  esac
}

parse_args() {
  for arg in "$@"; do
    case "$arg" in
      minimal|standard|secure)
        PROFILE="$arg"
        ;;
      --force)
        FORCE="true"
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
  local env_existed="false"
  local docker_env_existed="false"
  local created_any="false"

  if [[ -f "$ROOT_DIR/.env" ]]; then
    env_existed="true"
  fi
  if [[ -f "$ROOT_DIR/.env.docker.local" ]]; then
    docker_env_existed="true"
  fi

  if ! command -v openssl >/dev/null 2>&1; then
    echo "openssl is required for bootstrap because secrets and dev certs are generated" >&2
    exit 1
  fi

  if [[ -f "$ROOT_DIR/.env" && -f "$ROOT_DIR/.env.docker.local" ]] && [[ "$FORCE" != "true" ]]; then
    echo "bootstrap refused to overwrite existing .env or .env.docker.local" >&2
    echo "rerun with --force if you want to regenerate local env files" >&2
    exit 1
  fi

  if [[ ! -f "$ROOT_DIR/.env" || "$FORCE" == "true" ]]; then
    cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
    apply_env_overrides_from_file "$ROOT_DIR/presets/${PROFILE}.env" "$ROOT_DIR/.env"
    created_any="true"
  fi

  if [[ ! -f "$ROOT_DIR/.env.docker.local" || "$FORCE" == "true" ]]; then
    cp "$ROOT_DIR/.env.docker" "$ROOT_DIR/.env.docker.local"
    apply_env_overrides_from_file "$ROOT_DIR/presets/${PROFILE}.env" "$ROOT_DIR/.env.docker.local"
    created_any="true"
  fi

  if [[ "$FORCE" == "true" || ( "$env_existed" == "false" && "$docker_env_existed" == "false" ) ]]; then
    jwt_secret="$(generate_secret)"
    logger_secret="$(generate_secret)"
    postgres_password="$(generate_secret)"
    redis_password="$(generate_secret)"
    mongo_password="$(generate_secret)"
    bootstrap_admin_password="$(generate_secret)"
    bootstrap_user_password="$(generate_secret)"
  elif [[ "$env_existed" == "true" ]]; then
    jwt_secret="$(read_env_value "$ROOT_DIR/.env" "JWT_SECRET")"
    logger_secret="$(read_env_value "$ROOT_DIR/.env" "LOGGER_SHARED_SECRET")"
    postgres_password="$(read_env_value "$ROOT_DIR/.env" "POSTGRES_PASSWORD")"
    redis_password="$(read_env_value "$ROOT_DIR/.env" "REDIS_PASSWORD")"
    mongo_password="$(read_env_value "$ROOT_DIR/.env" "MONGO_INITDB_ROOT_PASSWORD")"
    bootstrap_admin_password="$(read_env_value "$ROOT_DIR/.env" "BOOTSTRAP_ADMIN_PASSWORD")"
    bootstrap_user_password="$(read_env_value "$ROOT_DIR/.env" "BOOTSTRAP_USER_PASSWORD")"
  else
    jwt_secret="$(read_env_value "$ROOT_DIR/.env.docker.local" "JWT_SECRET")"
    logger_secret="$(read_env_value "$ROOT_DIR/.env.docker.local" "LOGGER_SHARED_SECRET")"
    postgres_password="$(read_env_value "$ROOT_DIR/.env.docker.local" "POSTGRES_PASSWORD")"
    redis_password="$(read_env_value "$ROOT_DIR/.env.docker.local" "REDIS_PASSWORD")"
    mongo_password="$(read_env_value "$ROOT_DIR/.env.docker.local" "MONGO_INITDB_ROOT_PASSWORD")"
    bootstrap_admin_password="$(read_env_value "$ROOT_DIR/.env.docker.local" "BOOTSTRAP_ADMIN_PASSWORD")"
    bootstrap_user_password="$(read_env_value "$ROOT_DIR/.env.docker.local" "BOOTSTRAP_USER_PASSWORD")"
  fi

  local_auth_mode="$(resolve_local_auth_mode)"
  if [[ "$local_auth_mode" == "demo" ]]; then
    bootstrap_admin_password="admin_dev_password"
    bootstrap_user_password="developer_dev_password"
  fi

  if [[ -z "${bootstrap_admin_password:-}" ]]; then
    bootstrap_admin_password="$(generate_secret)"
  fi
  if [[ -z "${bootstrap_user_password:-}" ]]; then
    bootstrap_user_password="$(generate_secret)"
  fi

  for file in "$ROOT_DIR/.env" "$ROOT_DIR/.env.docker.local"; do
    replace_env_value "$file" "JWT_SECRET" "$jwt_secret"
    replace_env_value "$file" "LOGGER_SHARED_SECRET" "$logger_secret"
    replace_env_value "$file" "POSTGRES_PASSWORD" "$postgres_password"
    replace_env_value "$file" "REDIS_PASSWORD" "$redis_password"
    replace_env_value "$file" "MONGO_INITDB_ROOT_PASSWORD" "$mongo_password"
    replace_env_value "$file" "LOCAL_AUTH_MODE" "$local_auth_mode"
    replace_env_value "$file" "BOOTSTRAP_ADMIN_PASSWORD" "$bootstrap_admin_password"
    replace_env_value "$file" "BOOTSTRAP_USER_PASSWORD" "$bootstrap_user_password"
  done

  "$ROOT_DIR/scripts/certs-dev.sh"

  cat <<EOF
bootstrap completed with profile=$PROFILE
generated files:
  - .env
  - .env.docker.local
  - backend/infrastructure/certs/dev/*

local demo credentials:
  - mode: $local_auth_mode
  - developer / $bootstrap_user_password
  - admin / $bootstrap_admin_password

local-only secrets were $( [[ "$FORCE" == "true" || ( "$env_existed" == "false" && "$docker_env_existed" == "false" ) ]] && echo "regenerated" || echo "synced" ) for:
  - JWT_SECRET
  - LOGGER_SHARED_SECRET
  - POSTGRES_PASSWORD
  - REDIS_PASSWORD
  - MONGO_INITDB_ROOT_PASSWORD

next step:
  ./scripts/dev-up.sh $PROFILE
EOF
}

main "$@"
