#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"

BUCKET_NAME="${1:-}"
EXPECTED_POLICY_FILE="${2:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/check-s3-lifecycle-policy.sh <bucket-name> <expected-policy-file>

Optional env vars:
  BACKUP_AWS_REGION
  BACKUP_AWS_ENDPOINT_URL
EOF
}

if [[ -z "$BUCKET_NAME" || -z "$EXPECTED_POLICY_FILE" ]]; then
  usage >&2
  exit 1
fi

if [[ ! -f "$EXPECTED_POLICY_FILE" ]]; then
  echo "expected policy file not found: $EXPECTED_POLICY_FILE" >&2
  exit 1
fi

actual_file="$(mktemp)"
cleanup() {
  rm -f "$actual_file"
}
trap cleanup EXIT

aws_cli s3api get-bucket-lifecycle-configuration --bucket "$BUCKET_NAME" --output json >"$actual_file"

python3 - "$EXPECTED_POLICY_FILE" "$actual_file" <<'PY'
import json
import pathlib
import sys

expected = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
actual = json.loads(pathlib.Path(sys.argv[2]).read_text(encoding="utf-8"))


def normalize_rule(rule: dict) -> dict:
    normalized = {}
    for key in (
        "ID",
        "Status",
        "Filter",
        "Expiration",
        "Transitions",
        "NoncurrentVersionTransitions",
        "NoncurrentVersionExpiration",
        "AbortIncompleteMultipartUpload",
    ):
        if key in rule:
            normalized[key] = rule[key]
    return normalized


expected_rules = {
    normalize_rule(rule).get("ID", f"rule-{index}"): normalize_rule(rule)
    for index, rule in enumerate(expected.get("Rules", []))
}
actual_rules = {
    normalize_rule(rule).get("ID", f"rule-{index}"): normalize_rule(rule)
    for index, rule in enumerate(actual.get("Rules", []))
}

missing = []
different = []
for rule_id, expected_rule in expected_rules.items():
    actual_rule = actual_rules.get(rule_id)
    if actual_rule is None:
        missing.append(rule_id)
        continue
    if actual_rule != expected_rule:
        different.append({"id": rule_id, "expected": expected_rule, "actual": actual_rule})

if missing or different:
    payload = {"missing": missing, "different": different}
    print(json.dumps(payload, indent=2))
    raise SystemExit(1)

print("lifecycle policy matches expected rules")
PY
