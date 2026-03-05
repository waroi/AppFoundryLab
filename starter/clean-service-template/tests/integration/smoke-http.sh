#!/usr/bin/env bash
set -euo pipefail

# Minimal integration smoke for HTTP services created from this template.
# Required:
#   BASE_URL=http://127.0.0.1:8080
# Optional:
#   HEALTH_PATH=/health
#   READY_PATH=/health/ready
#   EXPECT_READY_503=true

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
HEALTH_PATH="${HEALTH_PATH:-/health}"
READY_PATH="${READY_PATH:-/health/ready}"
EXPECT_READY_503="${EXPECT_READY_503:-true}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required for integration smoke" >&2
  exit 1
fi

health_code="$(curl -sS -o /tmp/starter-health.out -w "%{http_code}" "${BASE_URL}${HEALTH_PATH}")"
if [[ "$health_code" != "200" ]]; then
  echo "health check failed: ${BASE_URL}${HEALTH_PATH} -> ${health_code}" >&2
  exit 1
fi

ready_code="$(curl -sS -o /tmp/starter-ready.out -w "%{http_code}" "${BASE_URL}${READY_PATH}")"
if [[ "$ready_code" == "200" ]]; then
  echo "ready check passed: ${BASE_URL}${READY_PATH} -> 200"
  exit 0
fi

if [[ "$EXPECT_READY_503" == "true" && "$ready_code" == "503" ]]; then
  echo "ready check accepted degraded status: ${BASE_URL}${READY_PATH} -> 503"
  exit 0
fi

echo "ready check failed: ${BASE_URL}${READY_PATH} -> ${ready_code}" >&2
exit 1
