#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT_DIR/.artifacts/sbom}"

if ! command -v syft >/dev/null 2>&1; then
  echo "syft is required. install: https://github.com/anchore/syft" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

syft "dir:$ROOT_DIR" -o cyclonedx-json="$OUT_DIR/sbom-cyclonedx.json"
syft "dir:$ROOT_DIR" -o spdx-json="$OUT_DIR/sbom-spdx.json"

echo "sbom generated:"
echo " - $OUT_DIR/sbom-cyclonedx.json"
echo " - $OUT_DIR/sbom-spdx.json"
