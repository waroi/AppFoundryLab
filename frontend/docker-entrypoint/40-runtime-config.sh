#!/usr/bin/env sh
set -eu

template="/opt/frontend/runtime-config.js.template"
output="/usr/share/nginx/html/runtime-config.js"

PUBLIC_API_BASE_URL="${PUBLIC_API_BASE_URL:-}"
export PUBLIC_API_BASE_URL

envsubst '${PUBLIC_API_BASE_URL}' < "$template" > "$output"
