#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"

ENV_FILE="${1:-$ROOT_DIR/.env.docker}"
OUT_ROOT="${2:-$ROOT_DIR/artifacts/local-release-evidence}"

usage() {
  cat <<'EOF'
usage: ./scripts/rehearse-release-evidence-local.sh [env-file] [out-root]

Optional env vars:
  LOCAL_RELEASE_EVIDENCE_AUDIT_TARGET
  LOCAL_RELEASE_EVIDENCE_KEEP_STACK
EOF
}

if [[ "${ENV_FILE:-}" == "-h" || "${ENV_FILE:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE" >&2
  exit 1
fi

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
source_sha="local"
if [[ -d "$ROOT_DIR/.git" ]]; then
  source_sha="$(git -C "$ROOT_DIR" rev-parse HEAD 2>/dev/null || echo local)"
fi

admin_user="$(read_env_value "$ENV_FILE" "BOOTSTRAP_ADMIN_USER")"
admin_password="$(read_env_value "$ENV_FILE" "BOOTSTRAP_ADMIN_PASSWORD")"
admin_user="${admin_user:-admin}"
admin_password="${admin_password:-admin_dev_password}"

base_report_dir="$OUT_ROOT/deploy-reports/base"
mkdir -p "$base_report_dir"

DEPLOY_MODE=build \
DEPLOY_API_BASE_URL="http://127.0.0.1:8080" \
DEPLOY_ADMIN_USER="$admin_user" \
DEPLOY_ADMIN_PASSWORD="$admin_password" \
DEPLOY_REPORT_DIR="$base_report_dir" \
"$ROOT_DIR/scripts/deploy-single-host.sh" up "$ENV_FILE"

environments=(staging-local production-local)
for environment in "${environments[@]}"; do
  manifest_dir="$OUT_ROOT/deploy-manifests/$environment"
  report_dir="$OUT_ROOT/deploy-reports/$environment"
  catalog_dir="$OUT_ROOT/release-catalog/$environment"
  ledger_dir="$OUT_ROOT/release-ledgers/$environment"
  evidence_dir="$OUT_ROOT/release-evidence/$environment"
  mkdir -p "$manifest_dir" "$report_dir" "$catalog_dir" "$ledger_dir" "$evidence_dir"

  cp -R "$base_report_dir/." "$report_dir/"
  manifest_path="$manifest_dir/release-manifest.env"
  release_id="local-${environment}-${timestamp}"
  cat >"$manifest_path" <<EOF
RELEASE_ID=$release_id
RELEASE_CREATED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
RELEASE_SOURCE_SHA=$source_sha
PROMOTION_SOURCE_RUN_ID=local-$timestamp
LOCAL_REHEARSAL=true
EOF

  "$ROOT_DIR/scripts/release-catalog.sh" sync-manifest "$catalog_dir/catalog.json" "$environment" "$manifest_path" >/dev/null
  "$ROOT_DIR/scripts/release-catalog.sh" record-operation "$catalog_dir/catalog.json" "$environment" "$release_id" deploy "$report_dir" >/dev/null
  "$ROOT_DIR/scripts/release-catalog.sh" export-ledger "$catalog_dir/catalog.json" "$release_id" "$ledger_dir/release-ledger-$release_id.json" >/dev/null
  "$ROOT_DIR/scripts/attest-release-ledger.sh" "$ledger_dir/release-ledger-$release_id.json" "$ledger_dir/release-ledger-$release_id.attestation.json" >/dev/null
  "$ROOT_DIR/scripts/verify-release-ledger-attestation.sh" "$ledger_dir/release-ledger-$release_id.json" "$ledger_dir/release-ledger-$release_id.attestation.json" >/dev/null
  "$ROOT_DIR/scripts/collect-release-evidence.sh" "$environment" "$catalog_dir/catalog.json" "$ledger_dir" "$evidence_dir" >/dev/null

  if [[ -n "${LOCAL_RELEASE_EVIDENCE_AUDIT_TARGET:-}" ]]; then
    RELEASE_EVIDENCE_AUDIT_PROFILE="${RELEASE_EVIDENCE_AUDIT_PROFILE:-versioned}" \
    "$ROOT_DIR/scripts/export-release-evidence.sh" \
      "$environment" \
      "$catalog_dir" \
      "$ledger_dir" \
      "$evidence_dir" \
      "$LOCAL_RELEASE_EVIDENCE_AUDIT_TARGET" >/dev/null
  fi
done

if [[ "${LOCAL_RELEASE_EVIDENCE_KEEP_STACK:-false}" != "true" ]]; then
  DEPLOY_MODE=build "$ROOT_DIR/scripts/deploy-single-host.sh" down "$ENV_FILE"
fi

echo "$OUT_ROOT"
