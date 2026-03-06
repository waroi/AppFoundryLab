#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="${1:-$ROOT_DIR/backend}"
shift || true

expected_go_version="$(awk '/^go / { print $2; exit }' "$ROOT_DIR/backend/go.mod")"

if [[ ! -x "$ROOT_DIR/.toolchain/go/bin/go" ]]; then
  echo "repo-local Go toolchain is missing; run ./scripts/bootstrap-go-toolchain.sh" >&2
  exit 1
fi

export PATH="$ROOT_DIR/.toolchain/go/bin:$PATH"
export GOCACHE="${GOCACHE:-$ROOT_DIR/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-$ROOT_DIR/.cache/go-mod}"

mkdir -p "$GOCACHE" "$GOMODCACHE"

current_go_version="$(go version | awk '{ sub(/^go/, "", $3); print $3 }')"
if [[ "$current_go_version" != "$expected_go_version" ]]; then
  echo "repo-local Go toolchain drift detected: have go$current_go_version, want go$expected_go_version" >&2
  echo "run ./scripts/bootstrap-go-toolchain.sh --force" >&2
  exit 1
fi

cd "$BACKEND_DIR"
go test ./... "$@"
