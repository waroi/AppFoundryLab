#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT_DIR/.artifacts/perf}"

"$ROOT_DIR/scripts/run-k6-scenario.sh" smoke "$OUT_DIR"

if [ -f "$OUT_DIR/k6-smoke-summary.json" ]; then
  cp "$OUT_DIR/k6-smoke-summary.json" "$OUT_DIR/k6-summary.json"
fi

echo "k6 smoke summary: $OUT_DIR/k6-summary.json"
