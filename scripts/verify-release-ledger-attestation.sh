#!/usr/bin/env bash
set -euo pipefail

LEDGER_PATH="${1:-}"
ATTESTATION_PATH="${2:-}"
PUBLIC_KEY_FILE="${3:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/verify-release-ledger-attestation.sh <ledger-path> <attestation-path> [public-key-file]
EOF
}

if [[ -z "$LEDGER_PATH" || -z "$ATTESTATION_PATH" ]]; then
  usage >&2
  exit 1
fi

if [[ ! -f "$LEDGER_PATH" || ! -f "$ATTESTATION_PATH" ]]; then
  echo "ledger or attestation file not found" >&2
  exit 1
fi

command -v python3 >/dev/null 2>&1 || {
  echo "python3 is required" >&2
  exit 1
}

current_sha256="$(sha256sum "$LEDGER_PATH" | awk '{print $1}')"
expected_sha256="$(python3 - "$ATTESTATION_PATH" <<'PY'
import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
print(payload.get("ledger", {}).get("sha256", ""))
PY
)"

if [[ "$current_sha256" != "$expected_sha256" ]]; then
  echo "ledger sha256 does not match attestation" >&2
  exit 1
fi

verification_mode="$(python3 - "$ATTESTATION_PATH" <<'PY'
import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
print(payload.get("verification", {}).get("mode", ""))
PY
)"

if [[ "$verification_mode" != "signing-key" ]]; then
  echo "attestation verified in digest-only mode"
  exit 0
fi

command -v openssl >/dev/null 2>&1 || {
  echo "openssl is required for signed attestation verification" >&2
  exit 1
}

signature_b64="$(python3 - "$ATTESTATION_PATH" <<'PY'
import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
print(payload.get("signature", ""))
PY
)"

if [[ -z "$signature_b64" ]]; then
  echo "signed attestation is missing signature" >&2
  exit 1
fi

temp_sig_file="$(mktemp)"
temp_pub_file="$(mktemp)"
cleanup() {
  rm -f "$temp_sig_file" "$temp_pub_file"
}
trap cleanup EXIT

printf '%s' "$signature_b64" | base64 -d >"$temp_sig_file"

if [[ -n "$PUBLIC_KEY_FILE" ]]; then
  cp "$PUBLIC_KEY_FILE" "$temp_pub_file"
else
  python3 - "$ATTESTATION_PATH" "$temp_pub_file" <<'PY'
import json
import pathlib
import sys

payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
public_key = payload.get("embeddedPublicKeyPem", "")
if not public_key:
    raise SystemExit("embedded public key missing from attestation")
pathlib.Path(sys.argv[2]).write_text(public_key, encoding="utf-8")
PY
fi

openssl dgst -sha256 -verify "$temp_pub_file" -signature "$temp_sig_file" "$LEDGER_PATH" >/dev/null
echo "attestation signature verified"
