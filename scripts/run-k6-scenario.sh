#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SCENARIO="${1:-}"
OUT_DIR="${2:-$ROOT_DIR/.artifacts/perf}"

if ! command -v k6 >/dev/null 2>&1; then
  echo "k6 is required. install: https://grafana.com/docs/k6/latest/set-up/install-k6/" >&2
  exit 1
fi

case "$SCENARIO" in
  smoke)
    SCRIPT="$ROOT_DIR/scripts/perf/k6-smoke.js"
    SUMMARY_FILE="$OUT_DIR/k6-smoke-summary.json"
    ;;
  spike)
    SCRIPT="$ROOT_DIR/scripts/perf/k6-spike.js"
    SUMMARY_FILE="$OUT_DIR/k6-spike-summary.json"
    ;;
  soak)
    SCRIPT="$ROOT_DIR/scripts/perf/k6-soak.js"
    SUMMARY_FILE="$OUT_DIR/k6-soak-summary.json"
    ;;
  *)
    echo "usage: ./scripts/run-k6-scenario.sh <smoke|spike|soak> [output_dir]" >&2
    exit 1
    ;;
esac

mkdir -p "$OUT_DIR"

K6_BASE_URL="${K6_BASE_URL:-http://127.0.0.1:8080}" \
K6_USERNAME="${K6_USERNAME:-developer}" \
K6_PASSWORD="${K6_PASSWORD:-developer_dev_password}" \
K6_VUS="${K6_VUS:-8}" \
K6_DURATION="${K6_DURATION:-30s}" \
K6_SOAK_VUS="${K6_SOAK_VUS:-12}" \
K6_SOAK_DURATION="${K6_SOAK_DURATION:-2m}" \
K6_SPIKE_PEAK_VUS="${K6_SPIKE_PEAK_VUS:-40}" \
K6_SPIKE_RAMP_UP="${K6_SPIKE_RAMP_UP:-20s}" \
K6_SPIKE_HOLD="${K6_SPIKE_HOLD:-20s}" \
K6_SPIKE_RAMP_DOWN="${K6_SPIKE_RAMP_DOWN:-20s}" \
  k6 run "$SCRIPT" --summary-export "$SUMMARY_FILE"

echo "k6 ${SCENARIO} summary: $SUMMARY_FILE"
