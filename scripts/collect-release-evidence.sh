#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENVIRONMENT="${1:-}"
CATALOG_PATH="${2:-}"
LEDGER_DIR="${3:-}"
OUT_DIR="${4:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/collect-release-evidence.sh <environment> <catalog-path> <ledger-dir> <out-dir>
EOF
}

if [[ -z "$ENVIRONMENT" || -z "$CATALOG_PATH" || -z "$LEDGER_DIR" || -z "$OUT_DIR" ]]; then
  usage >&2
  exit 1
fi

if [[ ! -f "$CATALOG_PATH" ]]; then
  echo "release catalog not found: $CATALOG_PATH" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"
"$ROOT_DIR/scripts/release-catalog.sh" export-ledger "$CATALOG_PATH" latest "$OUT_DIR/release-ledger-latest.json" >/dev/null

require_attestation=false
require_signed_attestation=false
case "$ENVIRONMENT" in
  staging|production)
    require_attestation=true
    require_signed_attestation=true
    ;;
esac

if [[ -n "${RELEASE_EVIDENCE_REQUIRE_ATTESTATION:-}" ]]; then
  require_attestation="${RELEASE_EVIDENCE_REQUIRE_ATTESTATION}"
fi
if [[ -n "${RELEASE_EVIDENCE_REQUIRE_SIGNED_ATTESTATION:-}" ]]; then
  require_signed_attestation="${RELEASE_EVIDENCE_REQUIRE_SIGNED_ATTESTATION}"
fi

status=0
if python3 - "$CATALOG_PATH" <<'PY' >/dev/null
import json
import pathlib
import sys

catalog = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
entries = catalog.get("entries", [])
if len(entries) > 1:
    raise SystemExit(10)
PY
then
  status=0
else
  status="$?"
fi
if [[ "$status" == "10" ]]; then
  "$ROOT_DIR/scripts/release-catalog.sh" export-ledger "$CATALOG_PATH" previous "$OUT_DIR/release-ledger-previous.json" >/dev/null
elif [[ "$status" != "0" ]]; then
  exit "$status"
fi

verify_attestations() {
  local require_any="$1"
  local require_signed="$2"
  local ledger

  shopt -s nullglob
  for ledger in "$LEDGER_DIR"/release-ledger-*.json; do
    case "$ledger" in
      *.attestation.json) continue ;;
    esac

    local attestation="${ledger%.json}.attestation.json"
    if [[ "$require_any" == "true" && ! -f "$attestation" ]]; then
      echo "attestation missing for ledger: $ledger" >&2
      return 1
    fi
    if [[ ! -f "$attestation" ]]; then
      continue
    fi

    "$ROOT_DIR/scripts/verify-release-ledger-attestation.sh" "$ledger" "$attestation" >/dev/null
    if [[ "$require_signed" == "true" ]]; then
      python3 - "$attestation" <<'PY'
import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
mode = payload.get("verification", {}).get("mode", "")
if mode != "signing-key":
    raise SystemExit(f"signed attestation required for {sys.argv[1]}")
PY
    fi
  done
}

verify_attestations "$require_attestation" "$require_signed_attestation"

python3 - "$ENVIRONMENT" "$CATALOG_PATH" "$LEDGER_DIR" "$OUT_DIR" "$require_attestation" "$require_signed_attestation" <<'PY'
import json
import pathlib
import sys
from collections import Counter
from datetime import datetime, timezone

environment = sys.argv[1]
catalog_path = pathlib.Path(sys.argv[2]).resolve()
ledger_dir = pathlib.Path(sys.argv[3]).resolve()
out_dir = pathlib.Path(sys.argv[4]).resolve()
require_attestation = sys.argv[5] == "true"
require_signed_attestation = sys.argv[6] == "true"
catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
entries = sorted(
    catalog.get("entries", []),
    key=lambda item: (
        item.get("createdAt", ""),
        item.get("lastSyncedAt", ""),
        item.get("lastRecordedAt", ""),
    ),
    reverse=True,
)


def now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def attestation_for_release(release_id: str) -> str:
    if not release_id:
        return ""
    candidate = ledger_dir / f"release-ledger-{release_id}.attestation.json"
    if candidate.is_file():
        return str(candidate)
    return ""


def attestation_mode(attestation_path: str) -> str:
    if not attestation_path:
        return ""
    payload = json.loads(pathlib.Path(attestation_path).read_text(encoding="utf-8"))
    return payload.get("verification", {}).get("mode", "")


def summary_entry(entry: dict) -> dict:
    operations = entry.get("operations", [])
    latest_operation = operations[-1] if operations else {}
    attestation_path = attestation_for_release(entry.get("releaseId", ""))
    return {
        "releaseId": entry.get("releaseId", ""),
        "createdAt": entry.get("createdAt", ""),
        "sourceSha": entry.get("sourceSha", ""),
        "manifestSha256": entry.get("manifestSha256", ""),
        "manifestPath": entry.get("manifestPath", ""),
        "operations": len(operations),
        "latestOperation": latest_operation.get("operation", ""),
        "latestOperationAt": latest_operation.get("recordedAt", ""),
        "attestationPath": attestation_path,
        "attestationMode": attestation_mode(attestation_path),
    }


operation_counts: Counter[str] = Counter()
for entry in entries:
    for operation in entry.get("operations", []):
        name = operation.get("operation", "")
        if name:
            operation_counts[name] += 1

summary = {
    "schemaVersion": "release-evidence-summary-v1",
    "generatedAt": now(),
    "environment": environment,
    "catalogPath": str(catalog_path),
    "ledgerDir": str(ledger_dir),
    "totalEntries": len(entries),
    "requirements": {
        "attestationRequired": require_attestation,
        "signedAttestationRequired": require_signed_attestation,
    },
    "operationCounts": dict(sorted(operation_counts.items())),
    "latest": summary_entry(entries[0]) if entries else {},
    "previous": summary_entry(entries[1]) if len(entries) > 1 else {},
    "entries": [summary_entry(entry) for entry in entries[:10]],
}

(out_dir / "release-evidence-summary.json").write_text(
    json.dumps(summary, indent=2) + "\n",
    encoding="utf-8",
)

lines = [
    f"# Release Evidence Summary ({environment})",
    "",
    f"- Generated at: {summary['generatedAt']}",
    f"- Catalog path: {summary['catalogPath']}",
    f"- Ledger dir: {summary['ledgerDir']}",
    f"- Total releases in catalog: {summary['totalEntries']}",
    f"- Attestation required: {summary['requirements']['attestationRequired']}",
    f"- Signed attestation required: {summary['requirements']['signedAttestationRequired']}",
    "",
    "## Operation Counts",
]
if summary["operationCounts"]:
    for name, count in summary["operationCounts"].items():
        lines.append(f"- {name}: {count}")
else:
    lines.append("- none")

for label in ("latest", "previous"):
    payload = summary.get(label, {})
    lines.extend(
        [
            "",
            f"## {label.title()} Release",
            f"- Release ID: {payload.get('releaseId', '') or 'n/a'}",
            f"- Created at: {payload.get('createdAt', '') or 'n/a'}",
            f"- Source SHA: {payload.get('sourceSha', '') or 'n/a'}",
            f"- Manifest SHA256: {payload.get('manifestSha256', '') or 'n/a'}",
            f"- Latest operation: {payload.get('latestOperation', '') or 'n/a'}",
            f"- Latest operation at: {payload.get('latestOperationAt', '') or 'n/a'}",
            f"- Attestation path: {payload.get('attestationPath', '') or 'n/a'}",
            f"- Attestation mode: {payload.get('attestationMode', '') or 'n/a'}",
        ]
    )

(out_dir / "release-evidence-summary.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
PY
