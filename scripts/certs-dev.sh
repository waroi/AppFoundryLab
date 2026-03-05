#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CERT_DIR="$ROOT_DIR/backend/infrastructure/certs/dev"
VALID_DAYS="${CERT_VALID_DAYS:-3650}"

mkdir -p "$CERT_DIR"

CA_KEY="$CERT_DIR/ca.key"
CA_CERT="$CERT_DIR/ca.crt"
SERVER_KEY="$CERT_DIR/server.key"
SERVER_CSR="$CERT_DIR/server.csr"
SERVER_CERT="$CERT_DIR/server.crt"
SERVER_EXT="$CERT_DIR/server.ext"
CLIENT_KEY="$CERT_DIR/client.key"
CLIENT_CSR="$CERT_DIR/client.csr"
CLIENT_CERT="$CERT_DIR/client.crt"
CLIENT_EXT="$CERT_DIR/client.ext"

openssl req -x509 -newkey rsa:4096 -sha256 -days "$VALID_DAYS" -nodes \
  -keyout "$CA_KEY" -out "$CA_CERT" \
  -subj "/CN=appfoundrylab-dev-ca"

openssl req -new -newkey rsa:2048 -nodes \
  -keyout "$SERVER_KEY" -out "$SERVER_CSR" \
  -subj "/CN=calculator"

cat > "$SERVER_EXT" <<'EOF'
subjectAltName = DNS:calculator,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$SERVER_CERT" -days "$VALID_DAYS" -sha256 -extfile "$SERVER_EXT"

openssl req -new -newkey rsa:2048 -nodes \
  -keyout "$CLIENT_KEY" -out "$CLIENT_CSR" \
  -subj "/CN=api-gateway"

cat > "$CLIENT_EXT" <<'EOF'
extendedKeyUsage = clientAuth
EOF

openssl x509 -req -in "$CLIENT_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$CLIENT_CERT" -days "$VALID_DAYS" -sha256 -extfile "$CLIENT_EXT"

rm -f "$SERVER_CSR" "$CLIENT_CSR" "$SERVER_EXT" "$CLIENT_EXT" "$CERT_DIR/ca.srl"

echo "dev certs generated in: $CERT_DIR"
