#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"

ENV_FILE="${1:-$ROOT_DIR/.env.single-host}"
BUNDLE_DIR="${2:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/restore-drill-single-host.sh [env-file] [bundle-dir]

If bundle-dir is omitted, the script seeds a deterministic restore fixture,
creates a fresh backup bundle, recreates the volumes, restores the bundle,
verifies the restored business fixture through the API, archives runtime
reports, and shuts the stack down again.
EOF
}

login_payload() {
  python3 - "$DEPLOY_ADMIN_USER" "$DEPLOY_ADMIN_PASSWORD" <<'PY'
import json
import sys

print(json.dumps({"username": sys.argv[1], "password": sys.argv[2]}))
PY
}

extract_token() {
  python3 -c '
import json
import sys

payload = json.load(sys.stdin)
print(payload["accessToken"])
'
}

build_restore_fixture() {
  local marker="$1"
  local fixture_file="$2"

  python3 - "$marker" "$fixture_file" <<'PY'
import json
import sys

marker = sys.argv[1]
fixture_file = sys.argv[2]
fixture = {
    "schemaVersion": "restore-drill-v2",
    "marker": marker,
    "users": [
        {
            "name": f"Restore Drill Alpha {marker}",
            "email": f"{marker}.alpha@example.com",
        },
        {
            "name": f"Restore Drill Beta {marker}",
            "email": f"{marker}.beta@example.com",
        },
        {
            "name": f"Restore Drill Gamma {marker}",
            "email": f"{marker}.gamma@example.com",
        },
    ],
    "requestLogs": [
        {
            "path": f"/restore-drill/{marker}/alpha",
            "method": "GET",
            "traceId": f"{marker}-trace-alpha",
            "durationMs": 12,
            "statusCode": 200,
        },
        {
            "path": f"/restore-drill/{marker}/beta",
            "method": "POST",
            "traceId": f"{marker}-trace-beta",
            "durationMs": 34,
            "statusCode": 202,
        },
    ],
}

with open(fixture_file, "w", encoding="utf-8") as handle:
    json.dump(fixture, handle, indent=2)
    handle.write("\n")
PY
}

build_legacy_fixture() {
  local marker="$1"
  local fixture_file="$2"

  python3 - "$marker" "$fixture_file" <<'PY'
import json
import sys

marker = sys.argv[1]
fixture_file = sys.argv[2]
fixture = {
    "schemaVersion": "restore-drill-v1-legacy",
    "marker": marker,
    "users": [
        {
            "name": f"Restore Drill {marker}",
            "email": f"{marker}@example.com",
        }
    ],
    "requestLogs": [
        {
            "path": f"/restore-drill/{marker}",
            "method": "GET",
            "traceId": marker,
            "durationMs": 12,
            "statusCode": 200,
        }
    ],
}

with open(fixture_file, "w", encoding="utf-8") as handle:
    json.dump(fixture, handle, indent=2)
    handle.write("\n")
PY
}

fixture_marker() {
  local fixture_file="$1"
  python3 - "$fixture_file" <<'PY'
import json
import sys

fixture = json.load(open(sys.argv[1], encoding="utf-8"))
print(fixture.get("marker", "restore-drill"))
PY
}

seed_restore_fixture() {
  local env_file="$1"
  local fixture_file="$2"
  local sql_file mongo_file mongo_db mongo_collection

  mongo_db="$(read_env_value "$env_file" "MONGO_DB")"
  mongo_collection="$(read_env_value "$env_file" "MONGO_COLLECTION")"
  sql_file="$(mktemp)"
  mongo_file="$(mktemp)"
  trap 'rm -f "$sql_file" "$mongo_file"' RETURN

  python3 - "$fixture_file" "$sql_file" "$mongo_file" "$mongo_db" "$mongo_collection" <<'PY'
import json
import sys

fixture = json.load(open(sys.argv[1], encoding="utf-8"))
sql_file = sys.argv[2]
mongo_file = sys.argv[3]
mongo_db = sys.argv[4]
mongo_collection = sys.argv[5]

def sql_escape(value: str) -> str:
    return value.replace("'", "''")

with open(sql_file, "w", encoding="utf-8") as handle:
    handle.write("INSERT INTO users (name, email) VALUES\n")
    rows = []
    for user in fixture.get("users", []):
        rows.append(
            f"('{sql_escape(user['name'])}', '{sql_escape(user['email'])}')"
        )
    handle.write(",\n".join(rows))
    handle.write("\nON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name;\n")

with open(mongo_file, "w", encoding="utf-8") as handle:
    handle.write(f"use {mongo_db}\n")
    for item in fixture.get("requestLogs", []):
        payload = {
            "path": item["path"],
            "method": item["method"],
            "ip": "127.0.0.1",
            "traceId": item["traceId"],
            "durationMs": item["durationMs"],
            "statusCode": item["statusCode"],
            "occurredAt": "2026-03-01T00:00:00Z",
        }
        payload_json = json.dumps(payload, separators=(",", ":"))
        handle.write(
            f'db.getCollection("{mongo_collection}").replaceOne('
            f'{{traceId: "{item["traceId"]}"}}, '
            f"{payload_json}, "
            "{upsert: true})\n"
        )
PY

  single_host_compose "$env_file" exec -T \
    -e PGPASSWORD="$(read_env_value "$env_file" "POSTGRES_PASSWORD")" \
    postgres psql -U "$(read_env_value "$env_file" "POSTGRES_USER")" -d "$(read_env_value "$env_file" "POSTGRES_DB")" \
    < "$sql_file"

  single_host_compose "$env_file" exec -T \
    mongo mongosh --quiet \
    --username "$(read_env_value "$env_file" "MONGO_INITDB_ROOT_USERNAME")" \
    --password "$(read_env_value "$env_file" "MONGO_INITDB_ROOT_PASSWORD")" \
    --authenticationDatabase admin \
    < "$mongo_file"
}

attach_fixture_to_bundle() {
  local fixture_file="$1"
  local bundle_dir="$2"
  local target_file temp_manifest

  target_file="$bundle_dir/restore-drill-fixture.json"
  cp "$fixture_file" "$target_file"

  temp_manifest="$(mktemp)"
  grep -v '^RESTORE_DRILL_FIXTURE_' "$bundle_dir/manifest.env" > "$temp_manifest" || true
  cat >> "$temp_manifest" <<EOF
RESTORE_DRILL_FIXTURE_FILE=$(basename "$target_file")
RESTORE_DRILL_FIXTURE_SHA256=$(sha256_file "$target_file")
EOF
  mv "$temp_manifest" "$bundle_dir/manifest.env"
}

resolve_fixture_file() {
  local bundle_dir="$1"
  local work_dir="$2"
  local legacy_marker="${3:-}"
  local manifest_fixture fixture_source fixture_target

  manifest_fixture="$(awk -F= '$1 == "RESTORE_DRILL_FIXTURE_FILE" { print $2 }' "$bundle_dir/manifest.env" 2>/dev/null || true)"
  fixture_source=""
  if [[ -n "$manifest_fixture" && -f "$bundle_dir/$manifest_fixture" ]]; then
    fixture_source="$bundle_dir/$manifest_fixture"
  elif [[ -f "$bundle_dir/restore-drill-fixture.json" ]]; then
    fixture_source="$bundle_dir/restore-drill-fixture.json"
  fi

  fixture_target="$work_dir/restore-drill-fixture.json"
  if [[ -n "$fixture_source" ]]; then
    cp "$fixture_source" "$fixture_target"
    printf '%s\n' "$fixture_target"
    return 0
  fi

  if [[ -n "$legacy_marker" ]]; then
    build_legacy_fixture "$legacy_marker" "$fixture_target"
    printf '%s\n' "$fixture_target"
    return 0
  fi

  echo "restore drill fixture metadata not found in bundle; provide RESTORE_DRILL_MARKER for legacy bundles" >&2
  exit 1
}

fetch_request_logs_for_fixture() {
  local api_base_url="$1"
  local token="$2"
  local fixture_file="$3"
  local logs_dir="$4"
  local trace_id index raw_file

  mkdir -p "$logs_dir"
  index=0
  while IFS= read -r trace_id; do
    if [[ -z "$trace_id" ]]; then
      continue
    fi
    raw_file="$(mktemp)"
    curl -fsS \
      -H "Authorization: Bearer $token" \
      "$api_base_url/api/v1/admin/request-logs?traceId=$trace_id&limit=10" \
      > "$raw_file"
    python3 - "$raw_file" "$logs_dir/request-logs-$index.json" <<'PY'
import hashlib
import json
import pathlib
import sys
from urllib.parse import urlsplit

input_path = pathlib.Path(sys.argv[1])
output_path = pathlib.Path(sys.argv[2])
payload = json.loads(input_path.read_text(encoding="utf-8"))
items = payload.get("items", [])
sanitized = []

for item in items:
    raw_path = str(item.get("path", "") or "")
    split_path = urlsplit(raw_path)
    trace_id = str(item.get("traceId", "") or "")
    sanitized.append(
        {
            "path": split_path.path or raw_path,
            "method": item.get("method", ""),
            "durationMs": item.get("durationMs", 0),
            "statusCode": item.get("statusCode", 0),
            "occurredAt": item.get("occurredAt", ""),
            "traceId": trace_id,
            "traceIdHash": hashlib.sha256(trace_id.encode("utf-8")).hexdigest()[:16] if trace_id else "",
        }
    )

output = {
    "schemaVersion": "restore-drill-request-log-archive-v1",
    "redaction": {
        "mode": "minimized",
        "removedFields": ["ip", "queryString"],
        "sourceItemCount": len(items),
        "storedItemCount": len(sanitized),
    },
    "items": sanitized,
}
output_path.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")
PY
    rm -f "$raw_file"
    index=$((index + 1))
  done < <(
    python3 - "$fixture_file" <<'PY'
import json
import sys

fixture = json.load(open(sys.argv[1], encoding="utf-8"))
seen = []
for item in fixture.get("requestLogs", []):
    trace_id = item.get("traceId", "")
    if trace_id and trace_id not in seen:
        seen.append(trace_id)
for trace_id in seen:
    print(trace_id)
PY
  )
}

verify_restore_fixture() {
  local api_base_url="$1"
  local fixture_file="$2"
  local token users_response logs_dir expected_file actual_file verification_file manifest_file marker

  marker="$(fixture_marker "$fixture_file")"
  mkdir -p "$ROOT_DIR/artifacts/restore-drill"

  expected_file="$ROOT_DIR/artifacts/restore-drill/fixture-expected-$marker.json"
  actual_file="$ROOT_DIR/artifacts/restore-drill/fixture-actual-$marker.json"
  verification_file="$ROOT_DIR/artifacts/restore-drill/fixture-verification-$marker.json"
  manifest_file="$ROOT_DIR/artifacts/restore-drill/fixture-manifest-$marker.txt"
  users_response="$ROOT_DIR/artifacts/restore-drill/users-$marker.json"
  logs_dir="$ROOT_DIR/artifacts/restore-drill/request-logs-$marker"

  cp "$fixture_file" "$expected_file"

  token="$(
    curl -fsS \
      -H 'Content-Type: application/json' \
      -X POST \
      --data "$(login_payload)" \
      "$api_base_url/api/v1/auth/token" | extract_token
  )"

  curl -fsS -H "Authorization: Bearer $token" "$api_base_url/api/v1/users" > "$users_response"
  fetch_request_logs_for_fixture "$api_base_url" "$token" "$fixture_file" "$logs_dir"

  python3 - "$fixture_file" "$users_response" "$logs_dir" "$actual_file" "$verification_file" <<'PY'
import json
import pathlib
import sys

fixture = json.load(open(sys.argv[1], encoding="utf-8"))
users_payload = json.load(open(sys.argv[2], encoding="utf-8")).get("data", [])
logs_dir = pathlib.Path(sys.argv[3])
actual_file = pathlib.Path(sys.argv[4])
verification_file = pathlib.Path(sys.argv[5])

available_users = {item.get("email"): item for item in users_payload}
expected_users = fixture.get("users", [])
matched_users = []
missing_users = []
for user in expected_users:
    current = available_users.get(user["email"])
    if current:
        matched_users.append(current)
    else:
        missing_users.append(user["email"])

all_logs = []
for path in sorted(logs_dir.glob("request-logs-*.json")):
    payload = json.load(open(path, encoding="utf-8"))
    all_logs.extend(payload.get("items", []))

matched_logs = []
missing_logs = []
for expected in fixture.get("requestLogs", []):
    current = next(
        (
            item
            for item in all_logs
            if item.get("traceId") == expected["traceId"]
            and item.get("path") == expected["path"]
            and item.get("method") == expected["method"]
            and item.get("statusCode") == expected["statusCode"]
        ),
        None,
    )
    if current:
        matched_logs.append(current)
    else:
        missing_logs.append(expected)

actual_payload = {
    "marker": fixture.get("marker", ""),
    "users": matched_users,
    "requestLogs": matched_logs,
}
verification_payload = {
    "schemaVersion": fixture.get("schemaVersion", "restore-drill"),
    "marker": fixture.get("marker", ""),
    "usersExpected": len(expected_users),
    "usersFound": len(matched_users),
    "requestLogsExpected": len(fixture.get("requestLogs", [])),
    "requestLogsFound": len(matched_logs),
    "missingUsers": missing_users,
    "missingRequestLogs": missing_logs,
    "status": "ok" if not missing_users and not missing_logs else "failed",
}

actual_file.write_text(json.dumps(actual_payload, indent=2) + "\n", encoding="utf-8")
verification_file.write_text(
    json.dumps(verification_payload, indent=2) + "\n",
    encoding="utf-8",
)

if missing_users or missing_logs:
    raise SystemExit("restore drill fixture verification failed")
PY

  cat > "$manifest_file" <<EOF
fixture_marker=$marker
expected_file=$(basename "$expected_file")
actual_file=$(basename "$actual_file")
verification_file=$(basename "$verification_file")
users_response=$(basename "$users_response")
request_logs_dir=$(basename "$logs_dir")
expected_sha256=$(sha256_file "$expected_file")
actual_sha256=$(sha256_file "$actual_file")
verification_sha256=$(sha256_file "$verification_file")
EOF
}

restore_bundle() {
  local bundle_dir="$1"
  local work_dir="$2"
  local postgres_bundle mongo_bundle postgres_input mongo_input

  postgres_bundle="$(awk -F= '$1 == "POSTGRES_BACKUP_FILE" { print $2 }' "$bundle_dir/manifest.env")"
  mongo_bundle="$(awk -F= '$1 == "MONGO_BACKUP_FILE" { print $2 }' "$bundle_dir/manifest.env")"
  if [[ -z "$postgres_bundle" || -z "$mongo_bundle" ]]; then
    echo "manifest.env missing backup file references" >&2
    exit 1
  fi

  postgres_input="$work_dir/postgres.sql"
  mongo_input="$work_dir/mongo.archive.gz"
  decrypt_file_if_needed "$bundle_dir/$postgres_bundle" "$postgres_input"
  decrypt_file_if_needed "$bundle_dir/$mongo_bundle" "$mongo_input"

  "$ROOT_DIR/scripts/restore-postgres.sh" "$ENV_FILE" "$postgres_input"
  "$ROOT_DIR/scripts/restore-mongo.sh" "$ENV_FILE" "$mongo_input" --drop
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
  if [[ -z "${DEPLOY_ADMIN_USER:-}" || -z "${DEPLOY_ADMIN_PASSWORD:-}" ]]; then
    echo "DEPLOY_ADMIN_USER and DEPLOY_ADMIN_PASSWORD are required for restore drill verification" >&2
    exit 1
  fi

  local frontend_base_url api_base_url raw_marker marker work_dir fixture_file
  frontend_base_url="${RESTORE_DRILL_FRONTEND_BASE_URL:-http://127.0.0.1:4321}"
  api_base_url="${RESTORE_DRILL_API_BASE_URL:-http://127.0.0.1:8080}"
  raw_marker="${RESTORE_DRILL_MARKER:-}"
  marker="${raw_marker:-restore-drill-$(date -u +%Y%m%d%H%M%S)}"
  work_dir="$(mktemp -d)"
  trap "rm -rf \"$work_dir\"" EXIT

  "$ROOT_DIR/scripts/deploy-single-host.sh" up "$ENV_FILE"

  if [[ -z "$BUNDLE_DIR" ]]; then
    fixture_file="$work_dir/restore-drill-fixture.json"
    build_restore_fixture "$marker" "$fixture_file"
    seed_restore_fixture "$ENV_FILE" "$fixture_file"
    "$ROOT_DIR/scripts/backup-single-host.sh" "$ENV_FILE"
    BUNDLE_DIR="$(latest_matching_dir "$ROOT_DIR/artifacts/backups/bundles" 'single-host-*')"
    if [[ -n "$BUNDLE_DIR" && -d "$BUNDLE_DIR" ]]; then
      attach_fixture_to_bundle "$fixture_file" "$BUNDLE_DIR"
    fi
  fi
  if [[ -z "$BUNDLE_DIR" || ! -d "$BUNDLE_DIR" ]]; then
    echo "backup bundle not found: $BUNDLE_DIR" >&2
    exit 1
  fi

  fixture_file="$(resolve_fixture_file "$BUNDLE_DIR" "$work_dir" "$raw_marker")"

  single_host_compose "$ENV_FILE" down -v
  "$ROOT_DIR/scripts/deploy-single-host.sh" up "$ENV_FILE"
  restore_bundle "$BUNDLE_DIR" "$work_dir"

  DEPLOY_REPORT_DIR="${DEPLOY_REPORT_DIR:-$ROOT_DIR/artifacts/restore-drill}" \
    DEPLOY_API_BASE_URL="$api_base_url" \
    DEPLOY_ADMIN_PASSWORD="$DEPLOY_ADMIN_PASSWORD" \
    "$ROOT_DIR/scripts/archive-runtime-report.sh" \
    "$api_base_url" \
    "$DEPLOY_ADMIN_USER" \
    "${DEPLOY_REPORT_DIR:-$ROOT_DIR/artifacts/restore-drill}"

  DEPLOY_ADMIN_USER="$DEPLOY_ADMIN_USER" \
  DEPLOY_ADMIN_PASSWORD="$DEPLOY_ADMIN_PASSWORD" \
  "$ROOT_DIR/scripts/post-deploy-check.sh" "$frontend_base_url" "$api_base_url" "$ENV_FILE"

  verify_restore_fixture "$api_base_url" "$fixture_file"

  if ! is_truthy "${RESTORE_DRILL_KEEP_RUNNING:-false}"; then
    "$ROOT_DIR/scripts/deploy-single-host.sh" down "$ENV_FILE"
  fi

  echo "restore drill completed with bundle: $BUNDLE_DIR"
}

main "$@"
