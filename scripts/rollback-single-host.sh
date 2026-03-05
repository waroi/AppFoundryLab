#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
TARGET_REF="${1:-}"
ENV_FILE="${2:-$ROOT_DIR/.env.single-host}"

usage() {
  cat <<'EOF'
usage: ./scripts/rollback-single-host.sh <git-ref|manifest-path|release-selector> [env-file]

example:
  ./scripts/rollback-single-host.sh v0.1.0 ./.env.single-host
  ./scripts/rollback-single-host.sh ./artifacts/ghcr/release-manifest.env ./.env.single-host
  RELEASE_CATALOG_PATH=./artifacts/release-catalog/staging/catalog.json ./scripts/rollback-single-host.sh previous ./.env.single-host

Optional env vars:
  DEPLOY_API_BASE_URL
  DEPLOY_ADMIN_USER
  DEPLOY_ADMIN_PASSWORD
  DEPLOY_REPORT_DIR
  RELEASE_CATALOG_PATH
EOF
}

resolve_target_ref() {
  local raw_target="$1"
  if [[ -f "$raw_target" || -z "${RELEASE_CATALOG_PATH:-}" ]]; then
    printf '%s\n' "$raw_target"
    return 0
  fi
  if [[ ! -f "$RELEASE_CATALOG_PATH" ]]; then
    printf '%s\n' "$raw_target"
    return 0
  fi

  "$ROOT_DIR/scripts/release-catalog.sh" resolve "$RELEASE_CATALOG_PATH" "$raw_target"
}

main() {
  if [[ -z "$TARGET_REF" ]]; then
    usage >&2
    exit 1
  fi
  if [[ ! -f "$ENV_FILE" ]]; then
    echo "env file not found: $ENV_FILE" >&2
    exit 1
  fi

  TARGET_REF="$(resolve_target_ref "$TARGET_REF")"

  if [[ -f "$TARGET_REF" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "$TARGET_REF"
    set +a
    DEPLOY_MODE=image "$ROOT_DIR/scripts/deploy-single-host.sh" up "$ENV_FILE"
    cat <<EOF
rollback completed
mode: image
manifest: $TARGET_REF
EOF
    exit 0
  fi

  if [[ ! -d "$ROOT_DIR/.git" ]]; then
    echo "rollback requires either a git checkout or an image manifest file" >&2
    exit 1
  fi

  local previous_ref
  previous_ref="$(git -C "$ROOT_DIR" rev-parse --short HEAD)"
  trap 'git -C "$ROOT_DIR" checkout "$previous_ref" >/dev/null 2>&1 || true' ERR

  git -C "$ROOT_DIR" fetch --all --tags --prune
  git -C "$ROOT_DIR" checkout "$TARGET_REF"
  DEPLOY_MODE=build "$ROOT_DIR/scripts/deploy-single-host.sh" up "$ENV_FILE"
  trap - ERR

  cat <<EOF
rollback completed
mode: build
previous ref: $previous_ref
current ref: $(git -C "$ROOT_DIR" rev-parse --short HEAD)
EOF
}

main "$@"
