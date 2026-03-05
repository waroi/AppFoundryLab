#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${ROOT_DIR:-}" ]]; then
  ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

require_command() {
  local command_name="$1"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "required command not found: $command_name" >&2
    exit 1
  fi
}

is_truthy() {
  local raw="${1:-}"
  raw="${raw,,}"
  [[ "$raw" == "1" || "$raw" == "true" || "$raw" == "yes" || "$raw" == "on" ]]
}

normalized_operator_access_mode() {
  local mode="${PROMETHEUS_OPERATOR_ACCESS_MODE:-basic-auth}"
  mode="${mode,,}"
  case "$mode" in
    basic-auth|mtls)
      printf '%s\n' "$mode"
      ;;
    *)
      printf '%s\n' "basic-auth"
      ;;
  esac
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

set_env_value() {
  local file="$1"
  local key="$2"
  local value="$3"
  local tmp
  tmp="$(mktemp)"
  awk -F= -v key="$key" '
    $1 != key {
      print
    }
  ' "$file" >"$tmp"
  printf '%s=%s\n' "$key" "$value" >>"$tmp"
  mv "$tmp" "$file"
}

append_runtime_override_keys() {
  local file="$1"
  local key
  for key in \
    DOCKER_HOST_BIND_ADDRESS \
    POSTGRES_HOST_PORT \
    REDIS_HOST_PORT \
    MONGO_HOST_PORT \
    LOGGER_HOST_PORT \
    CALCULATOR_HOST_PORT \
    API_GATEWAY_HOST_PORT \
    FRONTEND_HOST_PORT \
    ALLOWED_ORIGINS \
    PUBLIC_API_BASE_URL; do
    if [[ ${!key+x} ]]; then
      set_env_value "$file" "$key" "${!key}"
    fi
  done
}

docker_bin_path() {
  printf '%s\n' "${DOCKER_BIN:-docker}"
}

is_windows_docker_bin() {
  local docker_bin="${1:-}"
  docker_bin="${docker_bin,,}"
  [[ "$docker_bin" == *.exe ]]
}

docker_compose_path() {
  local raw_path="$1"
  local docker_bin
  docker_bin="$(docker_bin_path)"

  if is_windows_docker_bin "$docker_bin" && command -v wslpath >/dev/null 2>&1; then
    wslpath -w "$raw_path"
    return
  fi

  printf '%s\n' "$raw_path"
}

docker_compose_args() {
  local translated=()
  local expect_path="false"
  local arg key value

  for arg in "$@"; do
    if [[ "$expect_path" == "true" ]]; then
      translated+=("$(docker_compose_path "$arg")")
      expect_path="false"
      continue
    fi

    case "$arg" in
      -f|--file|--env-file|--project-directory)
        translated+=("$arg")
        expect_path="true"
        ;;
      --file=*|--env-file=*|--project-directory=*)
        key="${arg%%=*}"
        value="${arg#*=}"
        translated+=("$key=$(docker_compose_path "$value")")
        ;;
      *)
        translated+=("$arg")
        ;;
    esac
  done

  printf '%s\0' "${translated[@]}"
}

docker_compose_runtime_env_file() {
  local env_file="$1"
  local docker_bin
  docker_bin="$(docker_bin_path)"

  if ! is_windows_docker_bin "$docker_bin"; then
    printf '%s\n' "$env_file"
    return
  fi

  local temp_env_file
  temp_env_file="$(mktemp)"
  cat "$env_file" >"$temp_env_file"
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ "$line" == *=* ]] || continue
    local key="${line%%=*}"
    [[ -n "$key" && "$key" != \#* ]] || continue
    if [[ ${!key+x} ]]; then
      set_env_value "$temp_env_file" "$key" "${!key}"
    fi
  done <"$env_file"
  append_runtime_override_keys "$temp_env_file"
  set_env_value "$temp_env_file" "APP_ENV_FILE" "$(docker_compose_path "$env_file")"
  printf '%s\n' "$temp_env_file"
}

docker_cli() {
  local docker_bin
  docker_bin="$(docker_bin_path)"
  if [[ "${1:-}" == "compose" ]] && is_windows_docker_bin "$docker_bin"; then
    shift
    local compose_args=()
    while IFS= read -r -d '' value; do
      compose_args+=("$value")
    done < <(docker_compose_args "$@")
    "$docker_bin" compose "${compose_args[@]}"
    return
  fi

  "$docker_bin" "$@"
}

docker_compose_with_app_env_file() {
  local env_file="$1"
  local docker_bin compose_env_file app_env_file status previous_app_env_file had_app_env
  shift
  docker_bin="$(docker_bin_path)"
  compose_env_file="$env_file"
  app_env_file="$(docker_compose_path "$env_file")"

  if is_windows_docker_bin "$docker_bin"; then
    compose_env_file="$(docker_compose_runtime_env_file "$env_file")"
  fi

  if is_windows_docker_bin "$docker_bin"; then
    set +e
    docker_cli compose --env-file "$compose_env_file" "$@"
    status=$?
    set -e
    if [[ "$compose_env_file" != "$env_file" ]]; then
      rm -f "$compose_env_file"
    fi
    return "$status"
  fi

  previous_app_env_file="${APP_ENV_FILE-}"
  had_app_env="false"
  if [[ ${APP_ENV_FILE+x} ]]; then
    had_app_env="true"
  fi
  export APP_ENV_FILE="$app_env_file"
  set +e
  docker_cli compose --env-file "$compose_env_file" "$@"
  status=$?
  set -e
  if [[ "$had_app_env" == "true" ]]; then
    export APP_ENV_FILE="$previous_app_env_file"
  else
    unset APP_ENV_FILE
  fi
  return "$status"
}

single_host_compose_args() {
  local env_file="$1"
  local deploy_mode="${2:-${DEPLOY_MODE:-build}}"
  COMPOSE_ARGS=(
    -f "$ROOT_DIR/docker-compose.yml"
    -f "$ROOT_DIR/docker-compose.security.yml"
    -f "$ROOT_DIR/docker-compose.single-host.yml"
  )

  if [[ "$deploy_mode" == "image" ]]; then
    COMPOSE_ARGS+=(-f "$ROOT_DIR/deploy/docker-compose.single-host.ghcr.yml")
  fi
  if is_truthy "${ENABLE_OBSERVABILITY_STACK:-false}" || is_truthy "${ENABLE_OPERATOR_PROMETHEUS_ACCESS:-false}"; then
    COMPOSE_ARGS+=(-f "$ROOT_DIR/deploy/docker-compose.observability.yml")
  fi
  if is_truthy "${ENABLE_OPERATOR_PROMETHEUS_ACCESS:-false}"; then
    if [[ "$(normalized_operator_access_mode)" == "mtls" ]]; then
      COMPOSE_ARGS+=(-f "$ROOT_DIR/deploy/docker-compose.observability.operator.mtls.yml")
    else
      COMPOSE_ARGS+=(-f "$ROOT_DIR/deploy/docker-compose.observability.operator.yml")
    fi
  fi

  COMPOSE_ARGS+=(--env-file "$env_file")
}

single_host_compose_with_mode() {
  local env_file="$1"
  local deploy_mode="$2"
  local docker_bin compose_env_file app_env_file status previous_app_env_file had_app_env
  shift 2
  docker_bin="$(docker_bin_path)"
  compose_env_file="$env_file"
  app_env_file="$(docker_compose_path "$env_file")"

  if is_windows_docker_bin "$docker_bin"; then
    compose_env_file="$(docker_compose_runtime_env_file "$env_file")"
  fi

  single_host_compose_args "$compose_env_file" "$deploy_mode"

  if is_windows_docker_bin "$docker_bin"; then
    set +e
    docker_cli compose "${COMPOSE_ARGS[@]}" "$@"
    status=$?
    set -e
    if [[ "$compose_env_file" != "$env_file" ]]; then
      rm -f "$compose_env_file"
    fi
    return "$status"
  fi

  previous_app_env_file="${APP_ENV_FILE-}"
  had_app_env="false"
  if [[ ${APP_ENV_FILE+x} ]]; then
    had_app_env="true"
  fi
  export APP_ENV_FILE="$app_env_file"
  set +e
  docker_cli compose "${COMPOSE_ARGS[@]}" "$@"
  status=$?
  set -e
  if [[ "$had_app_env" == "true" ]]; then
    export APP_ENV_FILE="$previous_app_env_file"
  else
    unset APP_ENV_FILE
  fi
  return "$status"
}

single_host_compose() {
  local env_file="$1"
  shift
  single_host_compose_with_mode "$env_file" "${DEPLOY_MODE:-build}" "$@"
}

single_host_exec() {
  local env_file="$1"
  local service="$2"
  shift 2
  single_host_compose "$env_file" exec -T "$service" "$@"
}

sha256_file() {
  sha256sum "$1" | awk '{print $1}'
}

latest_matching_file() {
  local dir="$1"
  local pattern="$2"
  find "$dir" -maxdepth 1 -type f -name "$pattern" -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n1 | cut -d' ' -f2-
}

latest_matching_dir() {
  local dir="$1"
  local pattern="$2"
  find "$dir" -maxdepth 1 -type d -name "$pattern" -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n1 | cut -d' ' -f2-
}

is_remote_target() {
  local target="$1"
  if [[ "$target" == s3://* ]]; then
    return 1
  fi
  [[ "$target" == *:* && "$target" != /* ]]
}

is_s3_target() {
  local target="$1"
  [[ "$target" == s3://* ]]
}

s3_target_bucket() {
  local target="$1"
  local without_scheme="${target#s3://}"
  printf '%s\n' "${without_scheme%%/*}"
}

s3_target_prefix() {
  local target="$1"
  local without_scheme="${target#s3://}"
  if [[ "$without_scheme" == */* ]]; then
    printf '%s\n' "${without_scheme#*/}"
    return
  fi
  printf '\n'
}

s3_uri_join() {
  local base="${1%/}"
  local suffix="${2#/}"
  if [[ -z "$suffix" ]]; then
    printf '%s\n' "$base"
    return
  fi
  printf '%s/%s\n' "$base" "$suffix"
}

aws_cli() {
  require_command aws
  local args=()
  if [[ -n "${BACKUP_AWS_REGION:-}" ]]; then
    args+=(--region "$BACKUP_AWS_REGION")
  fi
  if [[ -n "${BACKUP_AWS_ENDPOINT_URL:-}" ]]; then
    args+=(--endpoint-url "$BACKUP_AWS_ENDPOINT_URL")
  fi
  aws "${args[@]}" "$@"
}

copy_s3_object_to_local() {
  local target_file="$1"
  local local_file="$2"
  aws_cli s3 cp "$target_file" "$local_file"
}

copy_local_file_to_s3() {
  local local_file="$1"
  local target_file="$2"
  aws_cli s3 cp "$local_file" "$target_file"
}

sync_dir_to_s3() {
  local source_dir="$1"
  local target_dir="$2"
  aws_cli s3 cp --recursive "$source_dir" "$target_dir"
}

prune_s3_retention() {
  local target="$1"
  local retention_days="$2"
  local name_pattern="$3"
  local bucket prefix list_json prefixes_file

  if [[ -z "$retention_days" || "$retention_days" -le 0 ]]; then
    return 0
  fi

  bucket="$(s3_target_bucket "$target")"
  prefix="$(s3_target_prefix "$target")"
  prefix="${prefix%/}"
  list_json="$(mktemp)"
  prefixes_file="$(mktemp)"

  aws_cli s3api list-objects-v2 \
    --bucket "$bucket" \
    --prefix "${prefix:+$prefix/}bundles/" \
    --output json >"$list_json"

  python3 - "$list_json" "$retention_days" "$name_pattern" "$prefixes_file" <<'PY'
import datetime as dt
import fnmatch
import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
retention_days = int(sys.argv[2])
name_pattern = sys.argv[3]
out_path = pathlib.Path(sys.argv[4])
cutoff = dt.datetime.now(dt.timezone.utc) - dt.timedelta(days=retention_days)

grouped: dict[str, dt.datetime] = {}
for item in payload.get("Contents", []):
    key = item.get("Key", "")
    parts = key.split("/")
    if len(parts) < 2:
        continue
    if "bundles" not in parts:
        continue
    bundle_name = parts[parts.index("bundles") + 1]
    if not fnmatch.fnmatch(bundle_name, name_pattern):
        continue
    last_modified_raw = item.get("LastModified", "")
    if not last_modified_raw:
        continue
    last_modified = dt.datetime.fromisoformat(last_modified_raw.replace("Z", "+00:00"))
    previous = grouped.get(bundle_name)
    if previous is None or last_modified > previous:
        grouped[bundle_name] = last_modified

expired = sorted(name for name, updated_at in grouped.items() if updated_at < cutoff)
out_path.write_text("\n".join(expired) + ("\n" if expired else ""), encoding="utf-8")
PY

  while IFS= read -r bundle_name; do
    [[ -n "$bundle_name" ]] || continue
    aws_cli s3 rm "$(s3_uri_join "$target" "bundles/$bundle_name")" --recursive
  done <"$prefixes_file"

  rm -f "$list_json" "$prefixes_file"
}

remote_target_host() {
  local target="$1"
  printf '%s\n' "${target%%:*}"
}

remote_target_path() {
  local target="$1"
  printf '%s\n' "${target#*:}"
}

ensure_target_dir() {
  local target="$1"
  if is_remote_target "$target"; then
    local remote_host remote_path
    remote_host="$(remote_target_host "$target")"
    remote_path="$(remote_target_path "$target")"
    ssh "$remote_host" "mkdir -p '$remote_path'"
  else
    mkdir -p "$target"
  fi
}

copy_target_file_to_local() {
  local target_file="$1"
  local local_file="$2"
  if is_remote_target "$target_file"; then
    local remote_host remote_path
    remote_host="$(remote_target_host "$target_file")"
    remote_path="$(remote_target_path "$target_file")"
    scp "$remote_host:$remote_path" "$local_file"
    return
  fi

  cp "$target_file" "$local_file"
}

copy_local_file_to_target() {
  local local_file="$1"
  local target_file="$2"
  if is_remote_target "$target_file"; then
    local remote_host remote_path remote_dir
    remote_host="$(remote_target_host "$target_file")"
    remote_path="$(remote_target_path "$target_file")"
    remote_dir="$(dirname "$remote_path")"
    ssh "$remote_host" "mkdir -p '$remote_dir'"
    scp "$local_file" "$remote_host:$remote_path"
    return
  fi

  mkdir -p "$(dirname "$target_file")"
  cp "$local_file" "$target_file"
}

sync_path_to_target() {
  local source_path="$1"
  local target="$2"
  ensure_target_dir "$target"
  if is_remote_target "$target"; then
    local remote_host remote_path
    remote_host="$(remote_target_host "$target")"
    remote_path="$(remote_target_path "$target")"
    if [[ -d "$source_path" ]]; then
      scp -r "$source_path" "$remote_host:$remote_path/"
    else
      scp "$source_path" "$remote_host:$remote_path/"
    fi
    return
  fi

  if [[ -d "$source_path" ]]; then
    cp -R "$source_path" "$target/"
  else
    cp "$source_path" "$target/"
  fi
}

prune_target_retention() {
  local target="$1"
  local retention_days="$2"
  local name_pattern="$3"
  if [[ -z "$retention_days" || "$retention_days" -le 0 ]]; then
    return 0
  fi

  if is_remote_target "$target"; then
    local remote_host remote_path
    remote_host="$(remote_target_host "$target")"
    remote_path="$(remote_target_path "$target")"
    ssh "$remote_host" "find '$remote_path' -mindepth 1 -maxdepth 1 -name '$name_pattern' -mtime +$retention_days -exec rm -rf {} +"
    return
  fi

  find "$target" -mindepth 1 -maxdepth 1 -name "$name_pattern" -mtime +"$retention_days" -exec rm -rf {} +
}

encrypt_file_if_needed() {
  local source_file="$1"
  local output_file="$2"
  local passphrase="${BACKUP_ENCRYPTION_PASSPHRASE:-}"

  if [[ -z "$passphrase" ]]; then
    cp "$source_file" "$output_file"
    return 0
  fi

  require_command openssl
  BACKUP_ENCRYPTION_PASSPHRASE="$passphrase" \
    openssl enc -aes-256-cbc -pbkdf2 -salt -in "$source_file" -out "$output_file" -pass env:BACKUP_ENCRYPTION_PASSPHRASE >/dev/null 2>&1
}

decrypt_file_if_needed() {
  local source_file="$1"
  local output_file="$2"

  if [[ "$source_file" == *.enc ]]; then
    if [[ -z "${BACKUP_ENCRYPTION_PASSPHRASE:-}" ]]; then
      echo "BACKUP_ENCRYPTION_PASSPHRASE is required to decrypt $source_file" >&2
      exit 1
    fi
    require_command openssl
    BACKUP_ENCRYPTION_PASSPHRASE="$BACKUP_ENCRYPTION_PASSPHRASE" \
      openssl enc -d -aes-256-cbc -pbkdf2 -in "$source_file" -out "$output_file" -pass env:BACKUP_ENCRYPTION_PASSPHRASE >/dev/null 2>&1
    return 0
  fi

  cp "$source_file" "$output_file"
}
