#!/usr/bin/env bash
set -euo pipefail

FRONTEND_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$FRONTEND_DIR/.playwright-linux.env"

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$ENV_FILE"
fi

cd "$FRONTEND_DIR"
exec ./node_modules/.bin/playwright test "$@"
