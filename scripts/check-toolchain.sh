#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MANIFEST="$ROOT_DIR/toolchain.versions.json"
CI_FILE="$ROOT_DIR/.github/workflows/appfoundrylab-ci.yml"
NIGHTLY_CI_FILE="$ROOT_DIR/.github/workflows/release-gate-full-nightly.yml"
GO_MOD_FILE="$ROOT_DIR/backend/go.mod"
API_DOCKERFILE="$ROOT_DIR/backend/services/api-gateway/Dockerfile"
LOGGER_DOCKERFILE="$ROOT_DIR/backend/services/logger/Dockerfile"
WORKER_DOCKERFILE="$ROOT_DIR/backend/core/calculator/Dockerfile"
FRONTEND_DOCKERFILE="$ROOT_DIR/frontend/Dockerfile"
FRONTEND_PROD_DOCKERFILE="$ROOT_DIR/frontend/Dockerfile.prod"

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

extract_with_jq() {
  local key="$1"
  jq -r --arg key "$key" '.[$key] // empty' "$MANIFEST"
}

extract_with_python() {
  local key="$1"
  python3 - "$MANIFEST" "$key" <<'PY'
import json
import sys

manifest = sys.argv[1]
key = sys.argv[2]

try:
    with open(manifest, encoding="utf-8") as fh:
        data = json.load(fh)
except Exception:
    sys.exit(2)

value = data.get(key, "")
if isinstance(value, str):
    print(value)
PY
}

extract() {
  local key="$1"
  if has_cmd jq; then
    extract_with_jq "$key"
    return
  fi

  if has_cmd python3; then
    extract_with_python "$key"
    return
  fi

  echo "missing parser: install jq or python3" >&2
  exit 1
}

GO_CI="$(extract go_ci)"
RUST_CI="$(extract rust_ci)"
BUN_CI="$(extract bun_ci)"
GO_DOCKER="$(extract go_docker)"
RUST_DOCKER="$(extract rust_docker)"
BUN_DOCKER="$(extract bun_docker)"

[[ -n "$GO_CI" && -n "$RUST_CI" && -n "$BUN_CI" && -n "$GO_DOCKER" && -n "$RUST_DOCKER" && -n "$BUN_DOCKER" ]] || {
  echo "toolchain manifest parse failed (jq/python3)"
  exit 1
}

grep -q "go-version: \"${GO_CI}\"" "$CI_FILE" || { echo "go ci version drift"; exit 1; }
grep -q "go-version: \"${GO_CI}\"" "$NIGHTLY_CI_FILE" || { echo "go nightly ci version drift"; exit 1; }
grep -q "bun-version: \"${BUN_CI}\"" "$CI_FILE" || { echo "bun ci version drift"; exit 1; }
grep -q "rust-toolchain@${RUST_CI}" "$CI_FILE" || { echo "rust ci version drift"; exit 1; }
GO_MOD_PATTERN="${GO_CI/%\.x/(\\.[0-9]+)?}"
grep -Eq "^go ${GO_MOD_PATTERN}$" "$GO_MOD_FILE" || { echo "go.mod version drift"; exit 1; }

grep -q "FROM golang:${GO_DOCKER}" "$API_DOCKERFILE" || { echo "api docker go version drift"; exit 1; }
grep -q "FROM golang:${GO_DOCKER}" "$LOGGER_DOCKERFILE" || { echo "logger docker go version drift"; exit 1; }
grep -q "FROM rust:${RUST_DOCKER}" "$WORKER_DOCKERFILE" || { echo "worker docker rust version drift"; exit 1; }
grep -q "FROM oven/bun:${BUN_DOCKER}" "$FRONTEND_DOCKERFILE" || { echo "frontend docker bun version drift"; exit 1; }
grep -q "FROM oven/bun:${BUN_DOCKER}" "$FRONTEND_PROD_DOCKERFILE" || { echo "frontend prod docker bun version drift"; exit 1; }

echo "toolchain governance check passed"
