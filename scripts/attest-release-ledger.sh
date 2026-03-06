#!/usr/bin/env bash
set -euo pipefail

LEDGER_PATH="${1:-}"
OUT_FILE="${2:-}"

usage() {
  cat <<'EOF'
usage: ./scripts/attest-release-ledger.sh <ledger-path> <out-file>

Optional env vars:
  LEDGER_ATTESTATION_SIGNING_KEY
  LEDGER_ATTESTATION_SIGNING_KEY_FILE
  LEDGER_ATTESTATION_KEY_ID
EOF
}

if [[ -z "$LEDGER_PATH" || -z "$OUT_FILE" ]]; then
  usage >&2
  exit 1
fi

if [[ ! -f "$LEDGER_PATH" ]]; then
  echo "ledger not found: $LEDGER_PATH" >&2
  exit 1
fi

if [[ "${LEDGER_ATTESTATION_REQUIRE_SIGNED:-false}" == "true" ]] && [[ -z "${LEDGER_ATTESTATION_SIGNING_KEY:-}" && -z "${LEDGER_ATTESTATION_SIGNING_KEY_FILE:-}" ]]; then
  echo "signed attestation is required but no signing key was provided" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT_FILE")"

ledger_sha256="$(sha256sum "$LEDGER_PATH" | awk '{print $1}')"
attestation_mode="digest-only"
signature_b64=""
public_key_pem=""
key_id="${LEDGER_ATTESTATION_KEY_ID:-}"
temp_key_file=""
temp_sig_file=""
remove_temp_key=false

cleanup() {
  if [[ "$remove_temp_key" == true ]]; then
    rm -f "$temp_key_file"
  fi
  rm -f "$temp_sig_file"
}
trap cleanup EXIT

if [[ -n "${LEDGER_ATTESTATION_SIGNING_KEY:-}" || -n "${LEDGER_ATTESTATION_SIGNING_KEY_FILE:-}" ]]; then
  command -v openssl >/dev/null 2>&1 || {
    echo "openssl is required for signed ledger attestations" >&2
    exit 1
  }

  if [[ -n "${LEDGER_ATTESTATION_SIGNING_KEY_FILE:-}" ]]; then
    temp_key_file="$LEDGER_ATTESTATION_SIGNING_KEY_FILE"
  else
    temp_key_file="$(mktemp)"
    printf '%s\n' "$LEDGER_ATTESTATION_SIGNING_KEY" >"$temp_key_file"
    remove_temp_key=true
  fi

  temp_sig_file="$(mktemp)"
  openssl dgst -sha256 -sign "$temp_key_file" -out "$temp_sig_file" "$LEDGER_PATH"
  signature_b64="$(base64 -w 0 "$temp_sig_file")"
  public_key_pem="$(openssl pkey -in "$temp_key_file" -pubout 2>/dev/null)"
  if [[ -z "$key_id" ]]; then
    key_id="$(printf '%s' "$public_key_pem" | sha256sum | awk '{print substr($1, 1, 16)}')"
  fi
  attestation_mode="signing-key"
fi

python3 - "$LEDGER_PATH" "$OUT_FILE" "$ledger_sha256" "$attestation_mode" "$key_id" "$signature_b64" "$public_key_pem" <<'PY'
import json
import pathlib
import sys
from datetime import datetime, timezone

ledger_path = pathlib.Path(sys.argv[1]).resolve()
out_path = pathlib.Path(sys.argv[2]).resolve()
ledger_sha256 = sys.argv[3]
attestation_mode = sys.argv[4]
key_id = sys.argv[5]
signature_b64 = sys.argv[6]
public_key_pem = sys.argv[7]

ledger = json.loads(ledger_path.read_text(encoding="utf-8"))
entry = ledger.get("entry", {})
payload = {
    "schemaVersion": "release-ledger-attestation-v1",
    "attestedAt": datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
    "verification": {
        "mode": attestation_mode,
        "algorithm": "openssl-rsa-sha256" if attestation_mode == "signing-key" else "sha256",
        "keyId": key_id,
    },
    "ledger": {
        "path": str(ledger_path),
        "sha256": ledger_sha256,
        "sizeBytes": ledger_path.stat().st_size,
        "releaseId": entry.get("releaseId", ""),
        "environment": ledger.get("catalogEnvironment", ""),
    },
}
if signature_b64:
    payload["signature"] = signature_b64
if public_key_pem:
    payload["embeddedPublicKeyPem"] = public_key_pem

out_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
print(str(out_path))
PY
