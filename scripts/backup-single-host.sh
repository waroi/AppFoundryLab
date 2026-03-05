#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"

ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
OUT_ROOT="${2:-$ROOT_DIR/artifacts/backups}"

usage() {
  cat <<'EOF'
usage: ./scripts/backup-single-host.sh [env-file] [out-root]

Optional env vars:
  BACKUP_SYNC_TARGET
  BACKUP_SYNC_PROFILE
  BACKUP_RETENTION_DAYS
  BACKUP_ENCRYPTION_PASSPHRASE
EOF
}

update_backup_catalog() {
  local catalog_path="$1"
  local bundle_dir="$2"
  local sync_target="${3:-}"
  local catalog_bundle_path="${4:-}"

  python3 - "$catalog_path" "$bundle_dir" "$sync_target" "$catalog_bundle_path" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

catalog_path = pathlib.Path(sys.argv[1])
bundle_dir = pathlib.Path(sys.argv[2]).resolve()
sync_target = sys.argv[3]
catalog_bundle_path = sys.argv[4]
manifest_path = bundle_dir / "manifest.env"

def now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")

def read_env(path: pathlib.Path) -> dict[str, str]:
    payload: dict[str, str] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        payload[key] = value
    return payload

if catalog_path.exists() and catalog_path.read_text(encoding="utf-8").strip():
    catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
else:
    catalog = {
        "schemaVersion": "backup-catalog-v1",
        "generatedAt": "",
        "entries": [],
    }

manifest = read_env(manifest_path)
entry = {
    "bundleName": manifest.get("BACKUP_BUNDLE_NAME", bundle_dir.name),
    "createdAt": manifest.get("BACKUP_CREATED_AT", ""),
    "bundlePath": catalog_bundle_path or str(bundle_dir),
    "manifestPath": str(manifest_path),
    "encrypted": manifest.get("BACKUP_ENCRYPTED", "false") == "true",
    "syncTarget": sync_target,
    "postgresBackupFile": manifest.get("POSTGRES_BACKUP_FILE", ""),
    "mongoBackupFile": manifest.get("MONGO_BACKUP_FILE", ""),
    "postgresBackupSha256": manifest.get("POSTGRES_BACKUP_SHA256", ""),
    "mongoBackupSha256": manifest.get("MONGO_BACKUP_SHA256", ""),
}

entries = catalog.setdefault("entries", [])
entries = [item for item in entries if item.get("bundleName") != entry["bundleName"]]
entries.append(entry)
entries.sort(key=lambda item: item.get("createdAt", ""), reverse=True)
catalog["entries"] = entries
catalog["generatedAt"] = now()

catalog_path.write_text(json.dumps(catalog, indent=2) + "\n", encoding="utf-8")
PY
}

write_latest_bundle_pointer() {
  local pointer_path="$1"
  local bundle_name="$2"
  mkdir -p "$(dirname "$pointer_path")"
  printf '%s\n' "$bundle_name" > "$pointer_path"
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

  local timestamp bundle_dir staging_dir postgres_raw mongo_raw postgres_bundle mongo_bundle bundle_name sync_profile
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  bundle_name="single-host-$timestamp"
  bundle_dir="$OUT_ROOT/bundles/$bundle_name"
  staging_dir="$bundle_dir/.staging"
  mkdir -p "$staging_dir"

  "$ROOT_DIR/scripts/backup-postgres.sh" "$ENV_FILE" "$OUT_ROOT/postgres"
  "$ROOT_DIR/scripts/backup-mongo.sh" "$ENV_FILE" "$OUT_ROOT/mongo"

  postgres_raw="$(latest_matching_file "$OUT_ROOT/postgres" 'postgres-*.sql')"
  mongo_raw="$(latest_matching_file "$OUT_ROOT/mongo" 'mongo-*.archive.gz')"

  if [[ -z "$postgres_raw" || -z "$mongo_raw" ]]; then
    echo "failed to locate newly created backup files" >&2
    exit 1
  fi

  postgres_bundle="$bundle_dir/$(basename "$postgres_raw")"
  mongo_bundle="$bundle_dir/$(basename "$mongo_raw")"
  if [[ -n "${BACKUP_ENCRYPTION_PASSPHRASE:-}" ]]; then
    postgres_bundle="$postgres_bundle.enc"
    mongo_bundle="$mongo_bundle.enc"
  fi

  mkdir -p "$bundle_dir"
  encrypt_file_if_needed "$postgres_raw" "$postgres_bundle"
  encrypt_file_if_needed "$mongo_raw" "$mongo_bundle"

  printf '%s  %s\n' "$(sha256_file "$postgres_bundle")" "$(basename "$postgres_bundle")" > "$bundle_dir/postgres.sha256"
  printf '%s  %s\n' "$(sha256_file "$mongo_bundle")" "$(basename "$mongo_bundle")" > "$bundle_dir/mongo.sha256"

  cat > "$bundle_dir/manifest.env" <<EOF
BACKUP_BUNDLE_NAME=$bundle_name
BACKUP_CREATED_AT=$timestamp
BACKUP_ENV_FILE_BASENAME=$(basename "$ENV_FILE")
POSTGRES_BACKUP_FILE=$(basename "$postgres_bundle")
MONGO_BACKUP_FILE=$(basename "$mongo_bundle")
POSTGRES_BACKUP_SHA256=$(sha256_file "$postgres_bundle")
MONGO_BACKUP_SHA256=$(sha256_file "$mongo_bundle")
BACKUP_ENCRYPTED=$(if [[ -n "${BACKUP_ENCRYPTION_PASSPHRASE:-}" ]]; then echo true; else echo false; fi)
BACKUP_SYNC_TARGET=${BACKUP_SYNC_TARGET:-}
BACKUP_SYNC_PROFILE=${BACKUP_SYNC_PROFILE:-versioned}
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-}
EOF

  rm -rf "$staging_dir"

  update_backup_catalog "$OUT_ROOT/bundles/backup-catalog.json" "$bundle_dir" "${BACKUP_SYNC_TARGET:-}" "$bundle_dir"
  write_latest_bundle_pointer "$OUT_ROOT/bundles/latest-bundle.txt" "$bundle_name"

  if [[ -n "${BACKUP_SYNC_TARGET:-}" ]]; then
    if [[ -z "${BACKUP_ENCRYPTION_PASSPHRASE:-}" ]]; then
      echo "BACKUP_ENCRYPTION_PASSPHRASE is required when BACKUP_SYNC_TARGET is set" >&2
      exit 1
    fi
    sync_profile="${BACKUP_SYNC_PROFILE:-versioned}"
    if is_s3_target "$BACKUP_SYNC_TARGET"; then
      sync_profile="s3"
    fi
    case "$sync_profile" in
      versioned)
        local remote_catalog_temp remote_latest_temp target_catalog target_latest remote_bundle_path
        sync_path_to_target "$bundle_dir" "$BACKUP_SYNC_TARGET"
        remote_catalog_temp="$(mktemp)"
        remote_latest_temp="$(mktemp)"
        target_catalog="${BACKUP_SYNC_TARGET%/}/backup-catalog.json"
        target_latest="${BACKUP_SYNC_TARGET%/}/latest-bundle.txt"
        remote_bundle_path="${BACKUP_SYNC_TARGET%/}/$bundle_name"

        if ! copy_target_file_to_local "$target_catalog" "$remote_catalog_temp" >/dev/null 2>&1; then
          rm -f "$remote_catalog_temp"
          remote_catalog_temp="$(mktemp)"
        fi
        update_backup_catalog "$remote_catalog_temp" "$bundle_dir" "$BACKUP_SYNC_TARGET" "$remote_bundle_path"
        write_latest_bundle_pointer "$remote_latest_temp" "$bundle_name"
        copy_local_file_to_target "$remote_catalog_temp" "$target_catalog"
        copy_local_file_to_target "$remote_latest_temp" "$target_latest"
        rm -f "$remote_catalog_temp" "$remote_latest_temp"
        if [[ -n "${BACKUP_RETENTION_DAYS:-}" ]]; then
          prune_target_retention "$BACKUP_SYNC_TARGET" "$BACKUP_RETENTION_DAYS" 'single-host-*'
        fi
        ;;
      s3)
        local remote_catalog_temp remote_latest_temp target_catalog target_latest remote_bundle_path
        remote_catalog_temp="$(mktemp)"
        remote_latest_temp="$(mktemp)"
        remote_bundle_path="$(s3_uri_join "$BACKUP_SYNC_TARGET" "bundles/$bundle_name")"
        target_catalog="$(s3_uri_join "$BACKUP_SYNC_TARGET" "backup-catalog.json")"
        target_latest="$(s3_uri_join "$BACKUP_SYNC_TARGET" "latest-bundle.txt")"

        sync_dir_to_s3 "$bundle_dir" "$remote_bundle_path/"
        if ! copy_s3_object_to_local "$target_catalog" "$remote_catalog_temp" >/dev/null 2>&1; then
          rm -f "$remote_catalog_temp"
          remote_catalog_temp="$(mktemp)"
        fi
        update_backup_catalog "$remote_catalog_temp" "$bundle_dir" "$BACKUP_SYNC_TARGET" "$remote_bundle_path"
        write_latest_bundle_pointer "$remote_latest_temp" "$bundle_name"
        copy_local_file_to_s3 "$remote_catalog_temp" "$target_catalog"
        copy_local_file_to_s3 "$remote_latest_temp" "$target_latest"
        rm -f "$remote_catalog_temp" "$remote_latest_temp"
        if [[ -n "${BACKUP_RETENTION_DAYS:-}" ]]; then
          prune_s3_retention "$BACKUP_SYNC_TARGET" "$BACKUP_RETENTION_DAYS" 'single-host-*'
        fi
        ;;
      *)
        echo "unsupported BACKUP_SYNC_PROFILE: $sync_profile" >&2
        exit 1
        ;;
    esac
  fi

  if [[ -n "${BACKUP_RETENTION_DAYS:-}" ]]; then
    prune_target_retention "$OUT_ROOT/bundles" "$BACKUP_RETENTION_DAYS" 'single-host-*'
  fi

  echo "backup bundle created: $bundle_dir"
}

main "$@"
