#!/usr/bin/env bash
set -euo pipefail

# Optional smoke test for template-based services that implement load shedding.
# Example:
#   BASE_URL=http://127.0.0.1:8080 \
#   OVERLOAD_PATH=/internal/test/overload \
#   HEALTH_PATH=/health \
#   EXPECT_OVERLOAD_503=true \
#   ./tests/integration/load-shed-smoke.sh

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
OVERLOAD_PATH="${OVERLOAD_PATH:-/internal/test/overload}"
HEALTH_PATH="${HEALTH_PATH:-/health}"
EXPECT_OVERLOAD_503="${EXPECT_OVERLOAD_503:-true}"

overload_code="$(curl -sS -o /tmp/starter-loadshed.out -w "%{http_code}" "${BASE_URL}${OVERLOAD_PATH}")"
health_code="$(curl -sS -o /tmp/starter-loadshed-health.out -w "%{http_code}" "${BASE_URL}${HEALTH_PATH}")"

if [ "$EXPECT_OVERLOAD_503" = "true" ] && [ "$overload_code" != "503" ]; then
  echo "expected overload endpoint to return 503, got ${overload_code}" >&2
  exit 1
fi

if [ "$health_code" != "200" ]; then
  echo "expected health endpoint to remain 200 during overload test, got ${health_code}" >&2
  exit 1
fi

echo "optional load shedding smoke passed"
