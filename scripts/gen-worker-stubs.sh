#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PROTO_FILE="$ROOT_DIR/backend/proto/worker.proto"
OUT_DIR="$ROOT_DIR/backend/services/api-gateway/internal/worker/workerpb"
PROTOC_BIN="${PROTOC_BIN:-}"

if [ -z "$PROTOC_BIN" ]; then
  if command -v protoc >/dev/null 2>&1; then
    PROTOC_BIN="$(command -v protoc)"
  else
    CANDIDATE="$(find "$ROOT_DIR/.toolchain/rust/cargo/registry/src" -type f -path '*protoc-bin-vendored-linux-x86_64*/bin/protoc' 2>/dev/null | head -n 1 || true)"
    if [ -n "$CANDIDATE" ]; then
      PROTOC_BIN="$CANDIDATE"
    fi
  fi
fi

if [ -z "$PROTOC_BIN" ]; then
  echo "error: protoc not found. install protobuf compiler first or set PROTOC_BIN."
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "error: protoc-gen-go not found. install with:"
  echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5"
  exit 1
fi

mkdir -p "$OUT_DIR"

if command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  "$PROTOC_BIN" \
    --proto_path="$ROOT_DIR/backend/proto" \
    --go_out="$OUT_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT_DIR" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_FILE"

  echo "worker stubs generated (modern mode):"
  echo "  $OUT_DIR/worker.pb.go"
  echo "  $OUT_DIR/worker_grpc.pb.go"
  exit 0
fi

echo "warning: protoc-gen-go-grpc not found, trying legacy plugins=grpc mode"
if ! "$PROTOC_BIN" \
  --proto_path="$ROOT_DIR/backend/proto" \
  --go_out=plugins=grpc,paths=source_relative:"$OUT_DIR" \
  "$PROTO_FILE"; then
  echo "error: legacy plugins=grpc mode failed."
  echo "install protoc-gen-go-grpc for modern mode:"
  echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1"
  exit 1
fi

rm -f "$OUT_DIR/worker_grpc.pb.go"
echo "worker stubs generated (legacy grpc-in-go mode):"
echo "  $OUT_DIR/worker.pb.go"
