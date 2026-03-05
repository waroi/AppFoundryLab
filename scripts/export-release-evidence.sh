#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"

ENVIRONMENT="${1:-}"
CATALOG_DIR="${2:-}"
LEDGER_DIR="${3:-}"
EVIDENCE_DIR="${4:-}"
AUDIT_TARGET="${5:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/export-release-evidence.sh <environment> <catalog-dir> <ledger-dir> <evidence-dir> <audit-target>

Optional env vars:
  RELEASE_EVIDENCE_AUDIT_PROFILE
  RELEASE_EVIDENCE_AWS_REGION
  RELEASE_EVIDENCE_AWS_ENDPOINT_URL
EOF
}

update_export_catalog() {
  local catalog_path="$1"
  local export_name="$2"
  local environment="$3"
  local export_path="$4"
  local source_catalog_dir="$5"
  local source_ledger_dir="$6"
  local source_evidence_dir="$7"
  local audit_target="$8"

  python3 - "$catalog_path" "$export_name" "$environment" "$export_path" "$source_catalog_dir" "$source_ledger_dir" "$source_evidence_dir" "$audit_target" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

catalog_path = pathlib.Path(sys.argv[1])
export_name = sys.argv[2]
environment = sys.argv[3]
export_path = sys.argv[4]
source_catalog_dir = pathlib.Path(sys.argv[5]).resolve()
source_ledger_dir = pathlib.Path(sys.argv[6]).resolve()
source_evidence_dir = pathlib.Path(sys.argv[7]).resolve()
audit_target = sys.argv[8]

if catalog_path.exists() and catalog_path.read_text(encoding="utf-8").strip():
    payload = json.loads(catalog_path.read_text(encoding="utf-8"))
else:
    payload = {
        "schemaVersion": "release-evidence-audit-catalog-v1",
        "generatedAt": "",
        "exports": [],
    }

now = datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
entry = {
    "exportName": export_name,
    "environment": environment,
    "createdAt": now,
    "exportPath": export_path,
    "auditTarget": audit_target,
    "sources": {
        "catalogDir": str(source_catalog_dir),
        "ledgerDir": str(source_ledger_dir),
        "evidenceDir": str(source_evidence_dir),
    },
}
exports = [item for item in payload.get("exports", []) if item.get("exportName") != export_name]
exports.append(entry)
exports.sort(key=lambda item: item.get("createdAt", ""), reverse=True)
payload["exports"] = exports
payload["generatedAt"] = now
catalog_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
PY
}

write_export_manifest() {
  local export_dir="$1"
  local environment="$2"
  local export_name="$3"
  local manifest_path="$export_dir/export-manifest.txt"

  {
    printf 'environment=%s\n' "$environment"
    printf 'export_name=%s\n' "$export_name"
    printf 'generated_at=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    find "$export_dir" -type f ! -name 'export-manifest.txt' -print0 | sort -z | while IFS= read -r -d '' file; do
      rel="${file#"$export_dir"/}"
      printf 'file=%s sha256=%s\n' "$rel" "$(sha256_file "$file")"
    done
  } >"$manifest_path"
}

if [[ -z "$ENVIRONMENT" || -z "$CATALOG_DIR" || -z "$LEDGER_DIR" || -z "$EVIDENCE_DIR" || -z "$AUDIT_TARGET" ]]; then
  usage >&2
  exit 1
fi

for required_dir in "$CATALOG_DIR" "$LEDGER_DIR" "$EVIDENCE_DIR"; do
  if [[ ! -d "$required_dir" ]]; then
    echo "required directory not found: $required_dir" >&2
    exit 1
  fi
done

export BACKUP_AWS_REGION="${RELEASE_EVIDENCE_AWS_REGION:-${BACKUP_AWS_REGION:-}}"
export BACKUP_AWS_ENDPOINT_URL="${RELEASE_EVIDENCE_AWS_ENDPOINT_URL:-${BACKUP_AWS_ENDPOINT_URL:-}}"

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
export_name="release-evidence-${ENVIRONMENT}-${timestamp}"
stage_dir="$(mktemp -d)"
export_dir="$stage_dir/$export_name"
mkdir -p "$export_dir/catalog" "$export_dir/ledgers" "$export_dir/evidence"

cp -R "$CATALOG_DIR/." "$export_dir/catalog/"
cp -R "$LEDGER_DIR/." "$export_dir/ledgers/"
cp -R "$EVIDENCE_DIR/." "$export_dir/evidence/"
write_export_manifest "$export_dir" "$ENVIRONMENT" "$export_name"

catalog_file="$stage_dir/release-evidence-audit-catalog.json"
pointer_file="$stage_dir/latest-export.txt"
target_catalog=""
target_pointer=""
target_export_path=""

case "${RELEASE_EVIDENCE_AUDIT_PROFILE:-versioned}" in
  versioned) ;;
  *)
    echo "unsupported RELEASE_EVIDENCE_AUDIT_PROFILE: ${RELEASE_EVIDENCE_AUDIT_PROFILE:-}" >&2
    rm -rf "$stage_dir"
    exit 1
    ;;
esac

if is_s3_target "$AUDIT_TARGET"; then
  target_export_path="$(s3_uri_join "$AUDIT_TARGET" "exports/$export_name")"
  target_catalog="$(s3_uri_join "$AUDIT_TARGET" "release-evidence-audit-catalog.json")"
  target_pointer="$(s3_uri_join "$AUDIT_TARGET" "latest-export.txt")"

  if ! copy_s3_object_to_local "$target_catalog" "$catalog_file" >/dev/null 2>&1; then
    : >"$catalog_file"
  fi
  update_export_catalog "$catalog_file" "$export_name" "$ENVIRONMENT" "$target_export_path" "$CATALOG_DIR" "$LEDGER_DIR" "$EVIDENCE_DIR" "$AUDIT_TARGET"
  printf '%s\n' "$export_name" >"$pointer_file"
  sync_dir_to_s3 "$export_dir" "$target_export_path/"
  copy_local_file_to_s3 "$catalog_file" "$target_catalog"
  copy_local_file_to_s3 "$pointer_file" "$target_pointer"
else
  target_export_path="${AUDIT_TARGET%/}/exports/$export_name"
  target_catalog="${AUDIT_TARGET%/}/release-evidence-audit-catalog.json"
  target_pointer="${AUDIT_TARGET%/}/latest-export.txt"

  if ! copy_target_file_to_local "$target_catalog" "$catalog_file" >/dev/null 2>&1; then
    : >"$catalog_file"
  fi
  update_export_catalog "$catalog_file" "$export_name" "$ENVIRONMENT" "$target_export_path" "$CATALOG_DIR" "$LEDGER_DIR" "$EVIDENCE_DIR" "$AUDIT_TARGET"
  printf '%s\n' "$export_name" >"$pointer_file"
  sync_path_to_target "$export_dir" "${AUDIT_TARGET%/}/exports"
  copy_local_file_to_target "$catalog_file" "$target_catalog"
  copy_local_file_to_target "$pointer_file" "$target_pointer"
fi

echo "$target_export_path"
rm -rf "$stage_dir"
