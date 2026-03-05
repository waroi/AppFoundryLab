#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CERT_DIR="${1:-$ROOT_DIR/deploy/observability/operator-certs}"
VALID_DAYS="${CERT_VALID_DAYS:-365}"

mkdir -p "$CERT_DIR"

CA_KEY="$CERT_DIR/ca.key"
CA_CERT="$CERT_DIR/client-ca.crt"
SERVER_KEY="$CERT_DIR/server.key"
SERVER_CSR="$CERT_DIR/server.csr"
SERVER_CERT="$CERT_DIR/server.crt"
SERVER_EXT="$CERT_DIR/server.ext"
CLIENT_KEY="$CERT_DIR/client.key"
CLIENT_CSR="$CERT_DIR/client.csr"
CLIENT_CERT="$CERT_DIR/client.crt"
CLIENT_EXT="$CERT_DIR/client.ext"
MANIFEST="$CERT_DIR/manifest.env"

openssl req -x509 -newkey rsa:4096 -sha256 -days "$VALID_DAYS" -nodes \
  -keyout "$CA_KEY" -out "$CA_CERT" \
  -subj "/CN=appfoundrylab-operator-ca"

openssl req -new -newkey rsa:2048 -nodes \
  -keyout "$SERVER_KEY" -out "$SERVER_CSR" \
  -subj "/CN=prometheus-operator-proxy"

cat >"$SERVER_EXT" <<'EOF'
subjectAltName = DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$SERVER_CERT" -days "$VALID_DAYS" -sha256 -extfile "$SERVER_EXT"

openssl req -new -newkey rsa:2048 -nodes \
  -keyout "$CLIENT_KEY" -out "$CLIENT_CSR" \
  -subj "/CN=prometheus-operator-client"

cat >"$CLIENT_EXT" <<'EOF'
extendedKeyUsage = clientAuth
EOF

openssl x509 -req -in "$CLIENT_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$CLIENT_CERT" -days "$VALID_DAYS" -sha256 -extfile "$CLIENT_EXT"

rm -f "$SERVER_CSR" "$CLIENT_CSR" "$SERVER_EXT" "$CLIENT_EXT" "$CERT_DIR/ca.srl"

cat >"$MANIFEST" <<EOF
OPERATOR_MTLS_CERT_DIR=$CERT_DIR
OPERATOR_MTLS_CA_CERT=$CA_CERT
OPERATOR_MTLS_SERVER_CERT=$SERVER_CERT
OPERATOR_MTLS_SERVER_KEY=$SERVER_KEY
OPERATOR_MTLS_CLIENT_CERT=$CLIENT_CERT
OPERATOR_MTLS_CLIENT_KEY=$CLIENT_KEY
OPERATOR_MTLS_GENERATED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
OPERATOR_MTLS_VALID_DAYS=$VALID_DAYS
EOF

echo "operator mTLS certs generated in: $CERT_DIR"
