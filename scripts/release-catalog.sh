#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:-help}"
CATALOG_PATH="${2:-}"

usage() {
  cat <<'EOF'
usage:
  ./scripts/release-catalog.sh sync-manifest <catalog-path> <environment> <manifest-path> [manifest-sha256]
  ./scripts/release-catalog.sh record-operation <catalog-path> <environment> <selector> <operation> <report-dir>
  ./scripts/release-catalog.sh resolve <catalog-path> <selector>
  ./scripts/release-catalog.sh list <catalog-path>
  ./scripts/release-catalog.sh export-ledger <catalog-path> <selector> <out-file>

selectors:
  latest
  previous
  <release-id>
  release:<release-id>
  sha:<release-source-sha>
  manifest-sha:<sha256>
  /absolute/path/to/release-manifest.env
EOF
}

require_catalog_path() {
  if [[ -z "$CATALOG_PATH" ]]; then
    usage >&2
    exit 1
  fi
}

init_catalog_if_needed() {
  local environment="$1"
  mkdir -p "$(dirname "$CATALOG_PATH")"
  if [[ -f "$CATALOG_PATH" ]]; then
    return 0
  fi

  python3 - "$CATALOG_PATH" "$environment" <<'PY'
import json
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
environment = sys.argv[2]
payload = {
    "schemaVersion": "release-catalog-v1",
    "environment": environment,
    "generatedAt": "",
    "entries": [],
}
path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
PY
}

sync_manifest() {
  require_catalog_path
  local environment="${3:-}"
  local manifest_path="${4:-}"
  local manifest_sha256="${5:-}"

  if [[ -z "$environment" || -z "$manifest_path" ]]; then
    usage >&2
    exit 1
  fi
  if [[ ! -f "$manifest_path" ]]; then
    echo "manifest not found: $manifest_path" >&2
    exit 1
  fi

  init_catalog_if_needed "$environment"

  python3 - "$CATALOG_PATH" "$environment" "$manifest_path" "$manifest_sha256" <<'PY'
import hashlib
import json
import pathlib
import sys
from datetime import datetime, timezone

catalog_path = pathlib.Path(sys.argv[1])
environment = sys.argv[2]
manifest_path = pathlib.Path(sys.argv[3]).resolve()
manifest_sha256 = sys.argv[4].strip()

def now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")

def read_env(path: pathlib.Path) -> dict[str, str]:
    values: dict[str, str] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
      line = line.strip()
      if not line or line.startswith("#") or "=" not in line:
          continue
      key, value = line.split("=", 1)
      values[key] = value
    return values

payload = json.loads(catalog_path.read_text(encoding="utf-8"))
entries = payload.setdefault("entries", [])
manifest = read_env(manifest_path)
manifest_sha = manifest_sha256 or hashlib.sha256(manifest_path.read_bytes()).hexdigest()
release_id = manifest.get("RELEASE_ID", "")
source_sha = manifest.get("RELEASE_SOURCE_SHA", "")
source_run_id = manifest.get("PROMOTION_SOURCE_RUN_ID", "")
created_at = manifest.get("RELEASE_CREATED_AT", "")
images = {
    "apiGateway": manifest.get("API_GATEWAY_IMAGE", ""),
    "logger": manifest.get("LOGGER_IMAGE", ""),
    "calculator": manifest.get("CALCULATOR_IMAGE", ""),
    "frontend": manifest.get("FRONTEND_IMAGE", ""),
}

entry = None
for current in entries:
    if release_id and current.get("releaseId") == release_id:
        entry = current
        break
    if current.get("manifestSha256") == manifest_sha or current.get("manifestPath") == str(manifest_path):
        entry = current
        break

if entry is None:
    entry = {
        "releaseId": release_id,
        "environment": environment,
        "manifestPath": str(manifest_path),
        "manifestSha256": manifest_sha,
        "createdAt": created_at,
        "sourceSha": source_sha,
        "sourceRunId": source_run_id,
        "images": images,
        "operations": [],
    }
    entries.append(entry)
else:
    entry["environment"] = environment
    entry["manifestPath"] = str(manifest_path)
    entry["manifestSha256"] = manifest_sha
    entry["createdAt"] = created_at or entry.get("createdAt", "")
    entry["sourceSha"] = source_sha or entry.get("sourceSha", "")
    entry["sourceRunId"] = source_run_id or entry.get("sourceRunId", "")
    entry["images"] = images

entry["lastSyncedAt"] = now()
payload["environment"] = environment
payload["generatedAt"] = now()

catalog_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
print(entry.get("releaseId", "") or entry["manifestPath"])
PY
}

record_operation() {
  require_catalog_path
  local environment="${3:-}"
  local selector="${4:-}"
  local operation="${5:-}"
  local report_dir="${6:-}"

  if [[ -z "$environment" || -z "$selector" || -z "$operation" || -z "$report_dir" ]]; then
    usage >&2
    exit 1
  fi

  init_catalog_if_needed "$environment"

  python3 - "$CATALOG_PATH" "$environment" "$selector" "$operation" "$report_dir" "${GITHUB_RUN_ID:-}" "${DEPLOY_PROMOTION_RUN_ID:-}" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

catalog_path = pathlib.Path(sys.argv[1])
environment = sys.argv[2]
selector = sys.argv[3]
operation = sys.argv[4]
report_dir = pathlib.Path(sys.argv[5]).resolve()
workflow_run_id = sys.argv[6]
promotion_run_id = sys.argv[7]

def now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")

def score(entry: dict) -> tuple[str, str]:
    return (
        entry.get("createdAt", ""),
        entry.get("lastSyncedAt", ""),
    )

def resolve(entries: list[dict], raw: str) -> dict:
    ordered = sorted(entries, key=score, reverse=True)
    if raw == "latest":
        if not ordered:
            raise SystemExit("release catalog is empty")
        return ordered[0]
    if raw == "previous":
        if len(ordered) < 2:
            raise SystemExit("release catalog does not contain a previous entry")
        return ordered[1]
    if raw.startswith("release:"):
        raw = raw.split(":", 1)[1]
    if raw.startswith("sha:"):
        value = raw.split(":", 1)[1]
        for entry in ordered:
            if entry.get("sourceSha") == value:
                return entry
    if raw.startswith("manifest-sha:"):
        value = raw.split(":", 1)[1]
        for entry in ordered:
            if entry.get("manifestSha256") == value:
                return entry
    for entry in ordered:
        if entry.get("releaseId") == raw:
            return entry
        if entry.get("manifestPath") == raw:
            return entry
        if entry.get("manifestSha256") == raw:
            return entry
        if entry.get("sourceSha") == raw:
            return entry
    raise SystemExit(f"release selector not found: {raw}")

def latest_matching(path: pathlib.Path, pattern: str) -> str:
    matches = sorted(path.glob(pattern), key=lambda item: item.stat().st_mtime, reverse=True)
    if not matches:
        return ""
    return str(matches[0].resolve())

payload = json.loads(catalog_path.read_text(encoding="utf-8"))
entries = payload.setdefault("entries", [])
entry = resolve(entries, selector)
operation_record = {
    "operation": operation,
    "recordedAt": now(),
    "workflowRunId": workflow_run_id,
    "promotionRunId": promotion_run_id,
    "reportDir": str(report_dir),
    "archiveManifestPath": latest_matching(report_dir, "archive-manifest-*.txt"),
    "deployManifestPath": latest_matching(report_dir, "deploy-manifest-*.txt"),
    "fixtureManifestPath": latest_matching(report_dir, "fixture-manifest-*.txt"),
    "fixtureVerificationPath": latest_matching(report_dir, "fixture-verification-*.json"),
}
entry.setdefault("operations", []).append(operation_record)
entry["lastRecordedAt"] = operation_record["recordedAt"]
payload["environment"] = environment
payload["generatedAt"] = now()

catalog_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
print(entry.get("releaseId", "") or entry.get("manifestPath", ""))
PY
}

resolve_selector() {
  require_catalog_path
  local selector="${3:-}"

  if [[ -z "$selector" ]]; then
    usage >&2
    exit 1
  fi

  python3 - "$CATALOG_PATH" "$selector" <<'PY'
import json
import pathlib
import sys

catalog_path = pathlib.Path(sys.argv[1])
selector = sys.argv[2]

if pathlib.Path(selector).is_file():
    print(str(pathlib.Path(selector).resolve()))
    raise SystemExit(0)

payload = json.loads(catalog_path.read_text(encoding="utf-8"))
entries = payload.get("entries", [])
ordered = sorted(
    entries,
    key=lambda entry: (entry.get("createdAt", ""), entry.get("lastSyncedAt", ""), entry.get("lastRecordedAt", "")),
    reverse=True,
)

def find(raw: str) -> dict:
    if raw == "latest":
        if not ordered:
            raise SystemExit("release catalog is empty")
        return ordered[0]
    if raw == "previous":
        if len(ordered) < 2:
            raise SystemExit("release catalog does not contain a previous entry")
        return ordered[1]
    if raw.startswith("release:"):
        raw = raw.split(":", 1)[1]
    if raw.startswith("sha:"):
        value = raw.split(":", 1)[1]
        for entry in ordered:
            if entry.get("sourceSha") == value:
                return entry
    if raw.startswith("manifest-sha:"):
        value = raw.split(":", 1)[1]
        for entry in ordered:
            if entry.get("manifestSha256") == value:
                return entry
    for entry in ordered:
        if entry.get("releaseId") == raw:
            return entry
        if entry.get("manifestPath") == raw:
            return entry
        if entry.get("manifestSha256") == raw:
            return entry
        if entry.get("sourceSha") == raw:
            return entry
    raise SystemExit(f"release selector not found: {raw}")

print(find(selector).get("manifestPath", ""))
PY
}

list_catalog() {
  require_catalog_path
  if [[ ! -f "$CATALOG_PATH" ]]; then
    echo '{"schemaVersion":"release-catalog-v1","entries":[]}'
    return 0
  fi
  cat "$CATALOG_PATH"
}

export_ledger() {
  require_catalog_path
  local selector="${3:-}"
  local out_file="${4:-}"

  if [[ -z "$selector" || -z "$out_file" ]]; then
    usage >&2
    exit 1
  fi

  mkdir -p "$(dirname "$out_file")"
  python3 - "$CATALOG_PATH" "$selector" "$out_file" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

catalog_path = pathlib.Path(sys.argv[1])
selector = sys.argv[2]
out_file = pathlib.Path(sys.argv[3])
payload = json.loads(catalog_path.read_text(encoding="utf-8"))
entries = payload.get("entries", [])
ordered = sorted(
    entries,
    key=lambda entry: (entry.get("createdAt", ""), entry.get("lastSyncedAt", ""), entry.get("lastRecordedAt", "")),
    reverse=True,
)

def resolve(raw: str) -> dict:
    if raw == "latest":
        if not ordered:
            raise SystemExit("release catalog is empty")
        return ordered[0]
    if raw == "previous":
        if len(ordered) < 2:
            raise SystemExit("release catalog does not contain a previous entry")
        return ordered[1]
    if raw.startswith("release:"):
        raw = raw.split(":", 1)[1]
    if raw.startswith("sha:"):
        value = raw.split(":", 1)[1]
        for entry in ordered:
            if entry.get("sourceSha") == value:
                return entry
    if raw.startswith("manifest-sha:"):
        value = raw.split(":", 1)[1]
        for entry in ordered:
            if entry.get("manifestSha256") == value:
                return entry
    for entry in ordered:
        if entry.get("releaseId") == raw:
            return entry
        if entry.get("manifestPath") == raw:
            return entry
        if entry.get("manifestSha256") == raw:
            return entry
        if entry.get("sourceSha") == raw:
            return entry
    raise SystemExit(f"release selector not found: {raw}")

entry = resolve(selector)
ledger = {
    "schemaVersion": "release-ledger-v1",
    "exportedAt": datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
    "catalogEnvironment": payload.get("environment", ""),
    "entry": entry,
}
out_file.write_text(json.dumps(ledger, indent=2) + "\n", encoding="utf-8")
print(str(out_file.resolve()))
PY
}

case "$COMMAND" in
  sync-manifest)
    sync_manifest "$@"
    ;;
  record-operation)
    record_operation "$@"
    ;;
  resolve)
    resolve_selector "$@"
    ;;
  list)
    list_catalog
    ;;
  export-ledger)
    export_ledger "$@"
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
