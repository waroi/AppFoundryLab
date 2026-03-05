#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CERT_DIR="${1:-$ROOT_DIR/deploy/observability/operator-certs}"
ENV_FILE="${2:-$ROOT_DIR/.env.single-host.example}"
MIN_VALID_DAYS="${MIN_VALID_DAYS:-14}"

usage() {
  cat <<'EOF'
usage: ./scripts/check-operator-mtls-readiness.sh [cert-dir] [env-file]
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ ! -d "$CERT_DIR" ]]; then
  echo "cert dir not found: $CERT_DIR" >&2
  exit 1
fi
if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE" >&2
  exit 1
fi

required_files=(
  "$CERT_DIR/client-ca.crt"
  "$CERT_DIR/server.crt"
  "$CERT_DIR/server.key"
  "$CERT_DIR/client.crt"
  "$CERT_DIR/client.key"
)

for path in "${required_files[@]}"; do
  [[ -f "$path" ]] || {
    echo "missing required operator mTLS file: $path" >&2
    exit 1
  }
done

check_expiry() {
  local cert_path="$1"
  local min_valid_seconds=$((MIN_VALID_DAYS * 24 * 60 * 60))
  if ! openssl x509 -checkend "$min_valid_seconds" -noout -in "$cert_path" >/dev/null 2>&1; then
    echo "certificate expires too soon: $cert_path" >&2
    exit 1
  fi
}

check_expiry "$CERT_DIR/client-ca.crt"
check_expiry "$CERT_DIR/server.crt"
check_expiry "$CERT_DIR/client.crt"

read_env_value() {
  awk -F= -v key="$2" '$1 == key { print substr($0, index($0, "=") + 1); exit }' "$1"
}

mode="$(read_env_value "$ENV_FILE" "PROMETHEUS_OPERATOR_ACCESS_MODE")"
tls_cert="$(read_env_value "$ENV_FILE" "PROMETHEUS_OPERATOR_TLS_CERT_FILE")"
tls_key="$(read_env_value "$ENV_FILE" "PROMETHEUS_OPERATOR_TLS_KEY_FILE")"
client_ca="$(read_env_value "$ENV_FILE" "PROMETHEUS_OPERATOR_CLIENT_CA_FILE")"

if [[ "$mode" != "mtls" ]]; then
  echo "PROMETHEUS_OPERATOR_ACCESS_MODE must be mtls in $ENV_FILE" >&2
  exit 1
fi
for value_name in tls_cert tls_key client_ca; do
  value="${!value_name:-}"
  [[ -n "$value" ]] || {
    echo "missing mTLS env reference: $value_name" >&2
    exit 1
  }
done

echo "operator mTLS readiness check passed"
