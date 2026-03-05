#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
ACTION="${1:-up}"
ENV_FILE="${2:-$ROOT_DIR/.env.docker.local}"
DEPLOY_MODE="${DEPLOY_MODE:-build}"

usage() {
  cat <<'EOF'
usage: ./scripts/deploy-single-host.sh [up|down|ps|logs|pull] [env-file]

examples:
  ./scripts/deploy-single-host.sh up
  ./scripts/deploy-single-host.sh up ./.env.single-host
  ./scripts/deploy-single-host.sh logs ./.env.single-host
EOF
}

maybe_archive_runtime_report() {
  local out_dir
  if [[ -z "${DEPLOY_API_BASE_URL:-}" ]]; then
    return 0
  fi
  if [[ -z "${DEPLOY_ADMIN_USER:-}" || -z "${DEPLOY_ADMIN_PASSWORD:-}" ]]; then
    echo "DEPLOY_API_BASE_URL is set but DEPLOY_ADMIN_USER or DEPLOY_ADMIN_PASSWORD is missing" >&2
    return 1
  fi

  out_dir="${DEPLOY_REPORT_DIR:-$ROOT_DIR/artifacts/deploy-reports}"
  "$ROOT_DIR/scripts/archive-runtime-report.sh" \
    "$DEPLOY_API_BASE_URL" \
    "$DEPLOY_ADMIN_USER" \
    "$DEPLOY_ADMIN_PASSWORD" \
    "$out_dir"
}

write_deploy_metadata() {
  local out_dir timestamp manifest compose_snapshot images_snapshot current_ref
  out_dir="${DEPLOY_REPORT_DIR:-$ROOT_DIR/artifacts/deploy-reports}"
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  mkdir -p "$out_dir"

  manifest="$out_dir/deploy-manifest-$timestamp.txt"
  compose_snapshot="$out_dir/compose-ps-$timestamp.txt"
  images_snapshot="$out_dir/compose-images-$timestamp.txt"
  current_ref="unknown"
  if [[ -d "$ROOT_DIR/.git" ]]; then
    current_ref="$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || true)"
  fi

  single_host_compose "$ENV_FILE" ps > "$compose_snapshot"
  single_host_compose "$ENV_FILE" images > "$images_snapshot"

  cat > "$manifest" <<EOF
deployed_at=$timestamp
deploy_mode=$DEPLOY_MODE
env_file=$(basename "$ENV_FILE")
git_ref=$current_ref
observability_stack=$(if is_truthy "${ENABLE_OBSERVABILITY_STACK:-false}"; then echo true; else echo false; fi)
compose_snapshot=$(basename "$compose_snapshot")
images_snapshot=$(basename "$images_snapshot")
api_gateway_image=${API_GATEWAY_IMAGE:-}
logger_image=${LOGGER_IMAGE:-}
calculator_image=${CALCULATOR_IMAGE:-}
frontend_image=${FRONTEND_IMAGE:-}
release_id=${RELEASE_ID:-${DEPLOY_RELEASE_ID:-}}
release_source_sha=${RELEASE_SOURCE_SHA:-${DEPLOY_RELEASE_SOURCE_SHA:-}}
release_manifest_path=${RELEASE_MANIFEST_PATH:-}
release_manifest_sha256=${RELEASE_MANIFEST_SHA256:-}
promotion_run_id=${DEPLOY_PROMOTION_RUN_ID:-}
EOF
}

wait_for_url() {
  local url="$1"
  local label="$2"

  for _ in $(seq 1 90); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "$label is ready: $url"
      return 0
    fi
    sleep 2
  done

  echo "$label did not become ready in time: $url" >&2
  return 1
}

main() {
  case "$ACTION" in
    up|down|ps|logs|pull) ;;
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

  case "$ACTION" in
    up)
      if [[ "$DEPLOY_MODE" == "image" ]]; then
        single_host_compose "$ENV_FILE" pull
        single_host_compose "$ENV_FILE" up -d
      else
        single_host_compose "$ENV_FILE" up --build -d
      fi
      wait_for_url "http://127.0.0.1:8080/health/live" "api-gateway"
      wait_for_url "http://127.0.0.1:4321/healthz" "frontend"
      if is_truthy "${ENABLE_OBSERVABILITY_STACK:-false}"; then
        wait_for_url "http://127.0.0.1:9090/-/ready" "prometheus"
      fi
      maybe_archive_runtime_report
      "$ROOT_DIR/scripts/post-deploy-check.sh" "http://127.0.0.1:4321" "http://127.0.0.1:8080" "$ENV_FILE"
      write_deploy_metadata
      ;;
    down)
      single_host_compose "$ENV_FILE" down
      ;;
    ps)
      single_host_compose "$ENV_FILE" ps
      ;;
    logs)
      single_host_compose "$ENV_FILE" logs --tail=200
      ;;
    pull)
      single_host_compose "$ENV_FILE" pull
      ;;
  esac
}

main "$@"
