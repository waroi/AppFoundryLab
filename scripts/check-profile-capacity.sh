#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/single-host-common.sh"
PROFILE="${1:-}"
OUT_MD="${2:-$ROOT_DIR/artifacts/capacity/profile-capacity-${PROFILE:-unknown}.md}"
DOCKER_BIN_PATH="$(docker_bin_path)"

if [ -z "$PROFILE" ]; then
  echo "usage: ./scripts/check-profile-capacity.sh <minimal|standard|secure> [out-md]" >&2
  exit 1
fi

case "$PROFILE" in
  minimal|standard|secure) ;;
  *)
    echo "invalid profile: $PROFILE (expected minimal|standard|secure)" >&2
    exit 1
    ;;
esac

if [[ "$DOCKER_BIN_PATH" == */* ]]; then
  if [[ ! -x "$DOCKER_BIN_PATH" ]]; then
    echo "docker is required" >&2
    exit 1
  fi
elif ! command -v "$DOCKER_BIN_PATH" >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
STATUS_FILE="$TMP_DIR/statuses.txt"
TOKEN_FILE="$TMP_DIR/token.json"
mkdir -p "$(dirname "$OUT_MD")"

restore_env() {
  if [ -f "$ROOT_DIR/.env.docker.base" ]; then
    mv "$ROOT_DIR/.env.docker.base" "$ROOT_DIR/.env.docker"
  fi
}

cleanup() {
  local exit_code=$?
  if [ $exit_code -ne 0 ]; then
    docker_cli compose --env-file "$ROOT_DIR/.env.docker" logs --tail=200 || true
  fi
  docker_cli compose --env-file "$ROOT_DIR/.env.docker" down -v >/dev/null 2>&1 || true
  restore_env
  rm -rf "$TMP_DIR"
  exit $exit_code
}
trap cleanup EXIT

cp "$ROOT_DIR/.env.docker" "$ROOT_DIR/.env.docker.base"
cat "$ROOT_DIR/presets/$PROFILE.env" >> "$ROOT_DIR/.env.docker"
cat >> "$ROOT_DIR/.env.docker" <<'EOF'
MAX_INFLIGHT_REQUESTS=8
LOAD_SHED_EXEMPT_PREFIXES=/health,/metrics
API_RATE_LIMIT_PER_MINUTE=2
AUTH_RATE_LIMIT_PER_MINUTE=200
EOF

"$ROOT_DIR/scripts/certs-dev.sh"
docker_cli compose --env-file "$ROOT_DIR/.env.docker" up --build -d

for i in $(seq 1 80); do
  if curl -fsS http://127.0.0.1:8080/health/live >/dev/null; then
    break
  fi
  if [ "$i" -eq 80 ]; then
    echo "api-gateway did not become live in time (profile=$PROFILE)" >&2
    exit 1
  fi
  sleep 2
done

curl -fsS -X POST http://127.0.0.1:8080/api/v1/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"username":"developer","password":"developer_dev_password"}' > "$TOKEN_FILE"

TOKEN="$(python3 - <<PY
import json
from pathlib import Path
obj = json.loads(Path("$TOKEN_FILE").read_text())
print(obj.get("accessToken", ""))
PY
)"

if [ -z "$TOKEN" ]; then
  echo "failed to retrieve access token for profile=$PROFILE" >&2
  exit 1
fi

export TOKEN
seq 1 120 | xargs -I{} -P 40 bash -lc '
  curl -sS -o /dev/null -w "%{http_code}\n" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -X POST http://127.0.0.1:8080/api/v1/compute/fibonacci \
    -d "{\"n\": 38}" || true
' >> "$STATUS_FILE"

sleep 1

users_1="$(curl -sS -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/users || true)"
users_2="$(curl -sS -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/users || true)"
users_3="$(curl -sS -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/users || true)"

live_status="$(curl -sS -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/health/live || true)"
ready_status="$(curl -sS -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/health/ready || true)"
load_shed_metric="$(curl -fsS http://127.0.0.1:8080/metrics | awk '/^api_gateway_load_shed_total / {print $2}' | tail -n1)"

shed_count="$(grep -c '^503$' "$STATUS_FILE" || true)"
ok_count="$(grep -c '^200$' "$STATUS_FILE" || true)"
rate_limited_count="$(grep -c '^429$' "$STATUS_FILE" || true)"

if [ "${shed_count:-0}" -lt 1 ]; then
  echo "expected at least one 503 from load shedding (profile=$PROFILE)" >&2
  exit 1
fi
if [ "${ok_count:-0}" -lt 1 ]; then
  echo "expected at least one successful 200 response during burst (profile=$PROFILE)" >&2
  exit 1
fi
if [ "$users_3" != "429" ]; then
  echo "expected 3rd /api/v1/users request to be 429, got $users_3 (profile=$PROFILE)" >&2
  exit 1
fi
if [ "$live_status" != "200" ]; then
  echo "expected /health/live to remain 200 under load, got $live_status (profile=$PROFILE)" >&2
  exit 1
fi
if [ "$ready_status" != "200" ] && [ "$ready_status" != "503" ]; then
  echo "expected /health/ready to be 200 or 503, got $ready_status (profile=$PROFILE)" >&2
  exit 1
fi
if [ -z "${load_shed_metric:-}" ]; then
  echo "expected api_gateway_load_shed_total metric to be exposed (profile=$PROFILE)" >&2
  exit 1
fi

cat > "$OUT_MD" <<EOF
# Profile Capacity Check

- profile: \`$PROFILE\`
- load shed (503) count: \`${shed_count:-0}\`
- successful (200) count: \`${ok_count:-0}\`
- burst 429 count: \`${rate_limited_count:-0}\`
- users statuses: \`$users_1,$users_2,$users_3\`
- health statuses: \`live=$live_status\`, \`ready=$ready_status\`
- metric \`api_gateway_load_shed_total\`: \`${load_shed_metric}\`

Result: PASS
EOF

echo "profile capacity check passed for profile=$PROFILE"
