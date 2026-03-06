#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

fail() {
  echo "test failure: $*" >&2
  exit 1
}

assert_file_exists() {
  local path="$1"
  [[ -f "$path" ]] || fail "expected file to exist: $path"
}

assert_contains() {
  local path="$1"
  local needle="$2"
  grep -F -- "$needle" "$path" >/dev/null 2>&1 || fail "expected $path to contain: $needle"
}

assert_not_contains() {
  local path="$1"
  local needle="$2"
  if grep -F -- "$needle" "$path" >/dev/null 2>&1; then
    fail "expected $path to not contain: $needle"
  fi
}

env_value() {
  local path="$1"
  local key="$2"
  awk -F= -v key="$key" '$1 == key { print substr($0, index($0, "=") + 1); exit }' "$path"
}

new_fixture() {
  local dir
  dir="$(mktemp -d)"
  mkdir -p "$dir/scripts" "$dir/presets" "$dir/backend/infrastructure/certs/dev" "$dir/deploy/observability" "$dir/deploy/backups"
  cp "$ROOT_DIR/scripts/bootstrap.sh" "$dir/scripts/bootstrap.sh"
  cp "$ROOT_DIR/scripts/single-host-common.sh" "$dir/scripts/single-host-common.sh"
  cp "$ROOT_DIR/scripts/certs-dev.sh" "$dir/scripts/certs-dev.sh"
  cp "$ROOT_DIR/scripts/dev-doctor.sh" "$dir/scripts/dev-doctor.sh"
  cp "$ROOT_DIR/scripts/dev-up.sh" "$dir/scripts/dev-up.sh"
  cp "$ROOT_DIR/scripts/dev-down.sh" "$dir/scripts/dev-down.sh"
  cp "$ROOT_DIR/scripts/deploy-single-host.sh" "$dir/scripts/deploy-single-host.sh"
  cp "$ROOT_DIR/scripts/rollback-single-host.sh" "$dir/scripts/rollback-single-host.sh"
  cp "$ROOT_DIR/scripts/release-catalog.sh" "$dir/scripts/release-catalog.sh"
  cp "$ROOT_DIR/scripts/collect-release-evidence.sh" "$dir/scripts/collect-release-evidence.sh"
  cp "$ROOT_DIR/scripts/export-release-evidence.sh" "$dir/scripts/export-release-evidence.sh"
  cp "$ROOT_DIR/scripts/attest-release-ledger.sh" "$dir/scripts/attest-release-ledger.sh"
  cp "$ROOT_DIR/scripts/verify-release-ledger-attestation.sh" "$dir/scripts/verify-release-ledger-attestation.sh"
  cp "$ROOT_DIR/scripts/bootstrap-playwright-linux.sh" "$dir/scripts/bootstrap-playwright-linux.sh"
  cp "$ROOT_DIR/scripts/check-s3-lifecycle-policy.sh" "$dir/scripts/check-s3-lifecycle-policy.sh"
  cp "$ROOT_DIR/scripts/generate-operator-mtls-certs.sh" "$dir/scripts/generate-operator-mtls-certs.sh"
  cp "$ROOT_DIR/scripts/check-operator-mtls-readiness.sh" "$dir/scripts/check-operator-mtls-readiness.sh"
  cp "$ROOT_DIR/scripts/rehearse-release-evidence-local.sh" "$dir/scripts/rehearse-release-evidence-local.sh"
  cp "$ROOT_DIR/scripts/post-deploy-check.sh" "$dir/scripts/post-deploy-check.sh"
  cp "$ROOT_DIR/scripts/archive-runtime-report.sh" "$dir/scripts/archive-runtime-report.sh"
  cp "$ROOT_DIR/scripts/backup-single-host.sh" "$dir/scripts/backup-single-host.sh"
  cp "$ROOT_DIR/scripts/backup-postgres.sh" "$dir/scripts/backup-postgres.sh"
  cp "$ROOT_DIR/scripts/backup-mongo.sh" "$dir/scripts/backup-mongo.sh"
  cp "$ROOT_DIR/scripts/restore-drill-single-host.sh" "$dir/scripts/restore-drill-single-host.sh"
  cp "$ROOT_DIR/scripts/restore-postgres.sh" "$dir/scripts/restore-postgres.sh"
  cp "$ROOT_DIR/scripts/restore-mongo.sh" "$dir/scripts/restore-mongo.sh"
  cp "$ROOT_DIR/.env.example" "$dir/.env.example"
  cp "$ROOT_DIR/.env.docker" "$dir/.env.docker"
  cp "$ROOT_DIR/.env.single-host.example" "$dir/.env.single-host.example"
  cp "$ROOT_DIR/presets/"*.env "$dir/presets/"
  cp "$ROOT_DIR/deploy/docker-compose.single-host.ghcr.yml" "$dir/deploy/docker-compose.single-host.ghcr.yml"
  cp "$ROOT_DIR/deploy/docker-compose.observability.yml" "$dir/deploy/docker-compose.observability.yml"
  cp "$ROOT_DIR/deploy/docker-compose.observability.operator.mtls.yml" "$dir/deploy/docker-compose.observability.operator.mtls.yml"
  cp "$ROOT_DIR/deploy/observability/prometheus.yml" "$dir/deploy/observability/prometheus.yml"
  cp "$ROOT_DIR/deploy/observability/Caddyfile.prometheus-operator.mtls" "$dir/deploy/observability/Caddyfile.prometheus-operator.mtls"
  cp "$ROOT_DIR/deploy/backups/s3-lifecycle-policy.example.json" "$dir/deploy/backups/s3-lifecycle-policy.example.json"
  touch "$dir/docker-compose.yml" "$dir/docker-compose.security.yml" "$dir/docker-compose.single-host.yml"
  chmod +x "$dir/scripts/"*.sh
  echo "$dir"
}

new_fakebin() {
  local dir
  dir="$(mktemp -d)"
  echo "$dir"
}

test_bootstrap_standard_generates_credentials() {
  local fixture
  fixture="$(new_fixture)"
  (
    cd "$fixture"
    ./scripts/bootstrap.sh standard --force >/tmp/test-dev-scripts-bootstrap-standard.out 2>&1
  )

  assert_file_exists "$fixture/.env"
  assert_file_exists "$fixture/.env.docker.local"
  assert_file_exists "$fixture/backend/infrastructure/certs/dev/ca.crt"
  assert_contains "$fixture/.env" "LOCAL_AUTH_MODE=generated"
  assert_contains "$fixture/.env.docker.local" "LOCAL_AUTH_MODE=generated"

  local admin_password user_password docker_admin docker_user
  admin_password="$(env_value "$fixture/.env" "BOOTSTRAP_ADMIN_PASSWORD")"
  user_password="$(env_value "$fixture/.env" "BOOTSTRAP_USER_PASSWORD")"
  docker_admin="$(env_value "$fixture/.env.docker.local" "BOOTSTRAP_ADMIN_PASSWORD")"
  docker_user="$(env_value "$fixture/.env.docker.local" "BOOTSTRAP_USER_PASSWORD")"

  [[ "$admin_password" != "admin_dev_password" ]] || fail "standard profile should not keep default admin password"
  [[ "$user_password" != "developer_dev_password" ]] || fail "standard profile should not keep default user password"
  [[ "$admin_password" == "$docker_admin" ]] || fail "admin password should sync across env files"
  [[ "$user_password" == "$docker_user" ]] || fail "user password should sync across env files"
}

test_bootstrap_minimal_keeps_demo_credentials() {
  local fixture
  fixture="$(new_fixture)"
  (
    cd "$fixture"
    ./scripts/bootstrap.sh minimal --force >/tmp/test-dev-scripts-bootstrap-minimal.out 2>&1
  )

  assert_contains "$fixture/.env" "LOCAL_AUTH_MODE=demo"
  assert_contains "$fixture/.env" "BOOTSTRAP_ADMIN_PASSWORD=admin_dev_password"
  assert_contains "$fixture/.env" "BOOTSTRAP_USER_PASSWORD=developer_dev_password"
}

test_dev_doctor_detects_missing_required_tools() {
  local fixture fakebin output
  fixture="$(new_fixture)"
  fakebin="$(new_fakebin)"
  ln -s "$(command -v bash)" "$fakebin/bash"
  ln -s "$(command -v curl)" "$fakebin/curl"
  ln -s "$(command -v openssl)" "$fakebin/openssl"
  output="$(mktemp)"

  if (cd "$fixture" && PATH="$fakebin" DOCKER_BIN="__missing_docker__" ./scripts/dev-doctor.sh >"$output" 2>&1); then
    fail "dev-doctor should fail when docker is missing"
  fi

  assert_contains "$output" "docker                   missing"
  assert_contains "$output" "doctor result: failed"
}

test_dev_up_and_down_compose_arguments() {
  local fixture fakebin log_file curl_log output_file
  fixture="$(new_fixture)"
  (
    cd "$fixture"
    ./scripts/bootstrap.sh minimal --force >/tmp/test-dev-scripts-bootstrap-dev-up.out 2>&1
  )
  cat >> "$fixture/.env.docker.local" <<'EOF'
FRONTEND_HOST_PORT=4400
API_GATEWAY_HOST_PORT=18081
LOGGER_HOST_PORT=18091
EOF

  fakebin="$(new_fakebin)"
  log_file="$fixture/docker-compose.log"
  curl_log="$fixture/dev-up-curl.log"
  output_file="$fixture/dev-up.out"
  cat > "$fakebin/docker" <<EOF
#!/usr/bin/env bash
echo "\$*" >> "$log_file"
exit 0
EOF
  cat > "$fakebin/curl" <<EOF
#!/usr/bin/env bash
url="\${@: -1}"
echo "\$url" >> "$curl_log"
if [[ "\$url" == *"/api/v1/auth/token" ]]; then
  printf '{"accessToken":"test-token"}'
fi
exit 0
EOF
  chmod +x "$fakebin/docker" "$fakebin/curl"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" ./scripts/dev-up.sh minimal security >"$output_file"
    PATH="$fakebin:$PATH" ./scripts/dev-down.sh security --volumes >/tmp/test-dev-scripts-dev-down.out
  )

  assert_contains "$log_file" "compose --env-file $fixture/.env.docker.local -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml up --build -d"
  assert_contains "$log_file" "compose --env-file $fixture/.env.docker.local -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml down --volumes"
  assert_contains "$curl_log" "http://127.0.0.1:18081/health/ready"
  assert_contains "$curl_log" "http://127.0.0.1:18091/health"
  assert_contains "$curl_log" "http://127.0.0.1:18091/metrics"
  assert_contains "$curl_log" "http://127.0.0.1:18081/api/v1/auth/token"
  assert_contains "$curl_log" "http://127.0.0.1:18081/api/v1/admin/runtime-report"
  assert_contains "$curl_log" "http://127.0.0.1:4400/healthz"
  assert_contains "$curl_log" "http://127.0.0.1:4400"
  assert_contains "$output_file" "frontend: http://127.0.0.1:4400"
  assert_contains "$output_file" "api: http://127.0.0.1:18081"
  assert_contains "$output_file" "logger metrics: http://127.0.0.1:18091/metrics"
}

test_windows_docker_bin_translates_compose_paths() {
  local fixture fakebin log_file env_log_file expected_compose expected_env
  fixture="$(new_fixture)"
  cp "$fixture/.env.docker" "$fixture/.env.docker.local"
  fakebin="$(new_fakebin)"
  log_file="$fixture/docker-compose-windows.log"
  env_log_file="$fixture/docker-compose-windows-env.log"
  if command -v wslpath >/dev/null 2>&1; then
    expected_compose="$(wslpath -w "$fixture/docker-compose.yml")"
    expected_env="$(wslpath -w "$fixture/.env.docker.local")"
  else
    expected_compose="$fixture/docker-compose.yml"
    expected_env="$fixture/.env.docker.local"
  fi

  cat > "$fakebin/docker.exe" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'ARGS=%s\n' "\$*" >> "$log_file"
args=("\$@")
for ((i = 0; i < \${#args[@]}; i++)); do
  if [[ "\${args[\$i]}" == "--env-file" ]]; then
    env_file_path="\${args[\$((i + 1))]}"
    if [[ "\$env_file_path" == *\\\\* ]]; then
      if command -v wslpath >/dev/null 2>&1; then
        env_file_path="\$(wslpath -u "\$env_file_path")"
      fi
    fi
    cat "\$env_file_path" > "$env_log_file"
    break
  fi
done
exit 0
EOF
  cat > "$fakebin/curl" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  cat > "$fakebin/wget" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$fakebin/docker.exe" "$fakebin/curl" "$fakebin/wget"

  (
    cd "$fixture"
    DOCKER_BIN="$fakebin/docker.exe" DEPLOY_MODE=image ./scripts/deploy-single-host.sh pull "$fixture/.env.docker.local" >/tmp/test-dev-scripts-docker-bin-windows.out
  )

  assert_contains "$log_file" "ARGS=compose -f $expected_compose"
  assert_contains "$env_log_file" "APP_ENV_FILE=$expected_env"
}

test_dev_up_windows_docker_bin_uses_runtime_env_file_and_windows_curl() {
  local fixture fakebin log_file env_log_file curl_exe_log expected_compose expected_security expected_env output_file
  fixture="$(new_fixture)"
  (
    cd "$fixture"
    ./scripts/bootstrap.sh minimal --force >/tmp/test-dev-scripts-bootstrap-dev-up-windows.out 2>&1
  )

  fakebin="$(new_fakebin)"
  log_file="$fixture/docker-compose-windows-dev-up.log"
  env_log_file="$fixture/docker-compose-windows-dev-up.env"
  curl_exe_log="$fixture/dev-up-windows-curl.log"
  output_file="$fixture/dev-up-windows.out"
  if command -v wslpath >/dev/null 2>&1; then
    expected_compose="$(wslpath -w "$fixture/docker-compose.yml")"
    expected_security="$(wslpath -w "$fixture/docker-compose.security.yml")"
    expected_env="$(wslpath -w "$fixture/.env.docker.local")"
  else
    expected_compose="$fixture/docker-compose.yml"
    expected_security="$fixture/docker-compose.security.yml"
    expected_env="$fixture/.env.docker.local"
  fi

  cat > "$fakebin/docker.exe" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'ARGS=%s\n' "\$*" >> "$log_file"
args=("\$@")
for ((i = 0; i < \${#args[@]}; i++)); do
  if [[ "\${args[\$i]}" == "--env-file" ]]; then
    env_file_path="\${args[\$((i + 1))]}"
    if [[ "\$env_file_path" == *\\\\* ]]; then
      if command -v wslpath >/dev/null 2>&1; then
        env_file_path="\$(wslpath -u "\$env_file_path")"
      fi
    fi
    cat "\$env_file_path" > "$env_log_file"
    break
  fi
done
exit 0
EOF
  cat > "$fakebin/curl" <<'EOF'
#!/usr/bin/env bash
exit 1
EOF
  cat > "$fakebin/curl.exe" <<EOF
#!/usr/bin/env bash
url="\${@: -1}"
echo "\$url" >> "$curl_exe_log"
if [[ "\$url" == *"/api/v1/auth/token" ]]; then
  printf '{"accessToken":"test-token"}'
fi
exit 0
EOF
  chmod +x "$fakebin/docker.exe" "$fakebin/curl" "$fakebin/curl.exe"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" \
      DOCKER_BIN="$fakebin/docker.exe" \
      CURL_EXE_BIN="$fakebin/curl.exe" \
      FRONTEND_HOST_PORT=5500 \
      API_GATEWAY_HOST_PORT=18082 \
      LOGGER_HOST_PORT=18092 \
      ./scripts/dev-up.sh minimal security >"$output_file"
  )

  assert_contains "$log_file" "ARGS=compose --env-file"
  assert_contains "$log_file" "$expected_compose"
  assert_contains "$log_file" "$expected_security"
  assert_contains "$env_log_file" "APP_ENV_FILE=$expected_env"
  assert_contains "$env_log_file" "FRONTEND_HOST_PORT=5500"
  assert_contains "$env_log_file" "API_GATEWAY_HOST_PORT=18082"
  assert_contains "$env_log_file" "LOGGER_HOST_PORT=18092"
  assert_contains "$curl_exe_log" "http://127.0.0.1:18082/health/ready"
  assert_contains "$curl_exe_log" "http://127.0.0.1:18092/health"
  assert_contains "$curl_exe_log" "http://127.0.0.1:18092/metrics"
  assert_contains "$curl_exe_log" "http://127.0.0.1:18082/api/v1/auth/token"
  assert_contains "$curl_exe_log" "http://127.0.0.1:18082/api/v1/admin/runtime-report"
  assert_contains "$curl_exe_log" "http://127.0.0.1:5500/healthz"
  assert_contains "$curl_exe_log" "http://127.0.0.1:5500"
  assert_contains "$output_file" "frontend: http://127.0.0.1:5500"
}

test_deploy_single_host_image_mode_uses_ghcr_overlay() {
  local fixture fakebin log_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.docker" "$fixture/.env.docker.local"

  fakebin="$(new_fakebin)"
  log_file="$fixture/docker-compose-image.log"
  cat > "$fakebin/docker" <<EOF
#!/usr/bin/env bash
echo "\$*" >> "$log_file"
exit 0
EOF
  cat > "$fakebin/curl" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  cat > "$fakebin/wget" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$fakebin/docker" "$fakebin/curl" "$fakebin/wget"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" DEPLOY_MODE=image ENABLE_OBSERVABILITY_STACK=true ./scripts/deploy-single-host.sh up ./.env.docker.local >/tmp/test-dev-scripts-deploy-image.out
  )

  assert_contains "$log_file" "compose -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml -f $fixture/docker-compose.single-host.yml -f $fixture/deploy/docker-compose.single-host.ghcr.yml -f $fixture/deploy/docker-compose.observability.yml --env-file ./.env.docker.local pull"
  assert_contains "$log_file" "compose -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml -f $fixture/docker-compose.single-host.yml -f $fixture/deploy/docker-compose.single-host.ghcr.yml -f $fixture/deploy/docker-compose.observability.yml --env-file ./.env.docker.local up -d"
  assert_not_contains "$log_file" "--build"
}

test_deploy_single_host_operator_observability_overlay_uses_proxy_compose_file() {
  local fixture fakebin log_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.docker" "$fixture/.env.docker.local"
  cp "$ROOT_DIR/deploy/docker-compose.observability.operator.yml" "$fixture/deploy/docker-compose.observability.operator.yml"
  cp "$ROOT_DIR/deploy/observability/Caddyfile.prometheus-operator" "$fixture/deploy/observability/Caddyfile.prometheus-operator"

  fakebin="$(new_fakebin)"
  log_file="$fixture/docker-compose-operator.log"
  cat > "$fakebin/docker" <<EOF
#!/usr/bin/env bash
echo "\$*" >> "$log_file"
exit 0
EOF
  cat > "$fakebin/curl" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  cat > "$fakebin/wget" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$fakebin/docker" "$fakebin/curl" "$fakebin/wget"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" \
      DEPLOY_MODE=image \
      ENABLE_OPERATOR_PROMETHEUS_ACCESS=true \
      ./scripts/deploy-single-host.sh pull ./.env.docker.local >/tmp/test-dev-scripts-deploy-operator-observability.out
  )

  assert_contains "$log_file" "compose -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml -f $fixture/docker-compose.single-host.yml -f $fixture/deploy/docker-compose.single-host.ghcr.yml -f $fixture/deploy/docker-compose.observability.yml -f $fixture/deploy/docker-compose.observability.operator.yml --env-file ./.env.docker.local pull"
}

test_deploy_single_host_operator_mtls_overlay_uses_proxy_compose_file() {
  local fixture fakebin log_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.docker" "$fixture/.env.docker.local"
  cp "$ROOT_DIR/deploy/docker-compose.observability.operator.yml" "$fixture/deploy/docker-compose.observability.operator.yml"

  fakebin="$(new_fakebin)"
  log_file="$fixture/docker-compose-operator-mtls.log"
  cat > "$fakebin/docker" <<EOF
#!/usr/bin/env bash
echo "\$*" >> "$log_file"
exit 0
EOF
  cat > "$fakebin/curl" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  cat > "$fakebin/wget" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$fakebin/docker" "$fakebin/curl" "$fakebin/wget"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" \
      DEPLOY_MODE=image \
      ENABLE_OPERATOR_PROMETHEUS_ACCESS=true \
      PROMETHEUS_OPERATOR_ACCESS_MODE=mtls \
      ./scripts/deploy-single-host.sh pull ./.env.docker.local >/tmp/test-dev-scripts-deploy-operator-mtls.out
  )

  assert_contains "$log_file" "compose -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml -f $fixture/docker-compose.single-host.yml -f $fixture/deploy/docker-compose.single-host.ghcr.yml -f $fixture/deploy/docker-compose.observability.yml -f $fixture/deploy/docker-compose.observability.operator.mtls.yml --env-file ./.env.docker.local pull"
}

test_backup_single_host_creates_bundle_manifest() {
  local fixture
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"

  cat > "$fixture/scripts/backup-postgres.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$2"
printf 'postgres backup\n' > "$2/postgres-20260301T000000Z.sql"
EOF
  cat > "$fixture/scripts/backup-mongo.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$2"
printf 'mongo backup\n' > "$2/mongo-20260301T000000Z.archive.gz"
EOF
  chmod +x "$fixture/scripts/backup-postgres.sh" "$fixture/scripts/backup-mongo.sh"

  (
    cd "$fixture"
    ./scripts/backup-single-host.sh ./.env.single-host >/tmp/test-dev-scripts-backup-bundle.out
  )

  bundle_dir="$(find "$fixture/artifacts/backups/bundles" -mindepth 1 -maxdepth 1 -type d | head -n 1)"
  [[ -n "$bundle_dir" ]] || fail "expected backup bundle directory to be created"
  assert_file_exists "$bundle_dir/manifest.env"
  assert_contains "$bundle_dir/manifest.env" "POSTGRES_BACKUP_FILE=postgres-20260301T000000Z.sql"
  assert_contains "$bundle_dir/manifest.env" "MONGO_BACKUP_FILE=mongo-20260301T000000Z.archive.gz"
  assert_file_exists "$fixture/artifacts/backups/bundles/backup-catalog.json"
  assert_contains "$fixture/artifacts/backups/bundles/backup-catalog.json" "\"bundleName\": \"$(basename "$bundle_dir")\""
  assert_contains "$fixture/artifacts/backups/bundles/latest-bundle.txt" "$(basename "$bundle_dir")"
}

test_backup_single_host_versioned_sync_writes_remote_catalog() {
  local fixture target_dir
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"
  target_dir="$fixture/remote-target"

  cat > "$fixture/scripts/backup-postgres.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$2"
printf 'postgres backup\n' > "$2/postgres-20260301T000000Z.sql"
EOF
  cat > "$fixture/scripts/backup-mongo.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$2"
printf 'mongo backup\n' > "$2/mongo-20260301T000000Z.archive.gz"
EOF
  chmod +x "$fixture/scripts/backup-postgres.sh" "$fixture/scripts/backup-mongo.sh"

  (
    cd "$fixture"
    BACKUP_SYNC_TARGET="$target_dir" \
    BACKUP_SYNC_PROFILE=versioned \
    BACKUP_ENCRYPTION_PASSPHRASE=test-passphrase \
    ./scripts/backup-single-host.sh ./.env.single-host >/tmp/test-dev-scripts-backup-versioned-sync.out
  )

  assert_file_exists "$target_dir/backup-catalog.json"
  assert_contains "$target_dir/backup-catalog.json" "\"bundleName\": \"single-host-"
  assert_file_exists "$target_dir/latest-bundle.txt"
}

test_backup_single_host_s3_sync_writes_object_storage_catalog() {
  local fixture fakebin aws_root log_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"
  fakebin="$(new_fakebin)"
  aws_root="$fixture/mock-s3"
  log_file="$fixture/aws.log"

  cat > "$fixture/scripts/backup-postgres.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$2"
printf 'postgres backup\n' > "$2/postgres-20260301T000000Z.sql"
EOF
  cat > "$fixture/scripts/backup-mongo.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "$2"
printf 'mongo backup\n' > "$2/mongo-20260301T000000Z.archive.gz"
EOF
  chmod +x "$fixture/scripts/backup-postgres.sh" "$fixture/scripts/backup-mongo.sh"

  cat > "$fakebin/aws" <<EOF
#!/usr/bin/env bash
set -euo pipefail
root="$aws_root"
log_file="$log_file"
mkdir -p "\$root"
printf '%s\n' "\$*" >> "\$log_file"
args=("\$@")
while [[ "\${#args[@]}" -gt 0 ]]; do
  case "\${args[0]}" in
    --region|--endpoint-url)
      args=("\${args[@]:2}")
      ;;
    *)
      break
      ;;
  esac
done
set -- "\${args[@]}"
if [[ "\$1" == "s3" && "\$2" == "cp" ]]; then
  shift 2
  recursive=false
  if [[ "\${1:-}" == "--recursive" ]]; then
    recursive=true
    shift
  fi
  src="\$1"
  dst="\$2"
  map_path() {
    local uri="\$1"
    local without_scheme="\${uri#s3://}"
    printf '%s/%s\n' "\$root" "\$without_scheme"
  }
  if [[ "\$src" == s3://* ]]; then
    src_path="\$(map_path "\$src")"
    mkdir -p "\$(dirname "\$dst")"
    cp "\$src_path" "\$dst"
  elif [[ "\$dst" == s3://* ]]; then
    dst_path="\$(map_path "\$dst")"
    mkdir -p "\$(dirname "\$dst_path")"
    if [[ "\$recursive" == true ]]; then
      cp -R "\$src/." "\$dst_path/"
    else
      cp "\$src" "\$dst_path"
    fi
  else
    exit 1
  fi
elif [[ "\$1" == "s3api" && "\$2" == "list-objects-v2" ]]; then
  shift 2
  bucket=""
  prefix=""
  while [[ "\$#" -gt 0 ]]; do
    case "\$1" in
      --bucket) bucket="\$2"; shift 2 ;;
      --prefix) prefix="\$2"; shift 2 ;;
      --output) shift 2 ;;
      *) shift ;;
    esac
  done
  target_root="\$root/\$bucket"
  export TARGET_ROOT="\$target_root"
  export PREFIX_FILTER="\$prefix"
  python3 - <<'PY'
import json
import os
from datetime import datetime, timezone
from pathlib import Path

root = Path(os.environ["TARGET_ROOT"])
prefix = os.environ["PREFIX_FILTER"]
contents = []
if root.exists():
    for path in sorted(root.rglob("*")):
        if not path.is_file():
            continue
        key = str(path.relative_to(root)).replace("\\\\", "/")
        if not key.startswith(prefix):
            continue
        contents.append(
            {
                "Key": key,
                "LastModified": datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
            }
        )
print(json.dumps({"Contents": contents}))
PY
elif [[ "\$1" == "s3" && "\$2" == "rm" ]]; then
  shift 2
  uri="\$1"
  target_path="\$root/\${uri#s3://}"
  rm -rf "\$target_path"
else
  exit 1
fi
EOF
  chmod +x "$fakebin/aws"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" \
      BACKUP_SYNC_TARGET="s3://test-bucket/archive" \
      BACKUP_SYNC_PROFILE=s3 \
      BACKUP_ENCRYPTION_PASSPHRASE=test-passphrase \
      ./scripts/backup-single-host.sh ./.env.single-host >/tmp/test-dev-scripts-backup-s3-sync.out
  )

  assert_file_exists "$aws_root/test-bucket/archive/backup-catalog.json"
  assert_file_exists "$aws_root/test-bucket/archive/latest-bundle.txt"
  assert_contains "$aws_root/test-bucket/archive/backup-catalog.json" "\"bundlePath\": \"s3://test-bucket/archive/bundles/"
}

test_export_release_evidence_writes_versioned_audit_bundle() {
  local fixture source_root target_root
  fixture="$(new_fixture)"
  source_root="$fixture/artifacts"
  target_root="$fixture/audit-target"
  mkdir -p "$source_root/release-catalog/staging" "$source_root/release-ledgers/staging" "$source_root/release-evidence/staging"

  cat > "$source_root/release-catalog/staging/catalog.json" <<'EOF'
{"schemaVersion":"release-catalog-v1","entries":[]}
EOF
  cat > "$source_root/release-ledgers/staging/release-ledger-sample.json" <<'EOF'
{"schemaVersion":"release-ledger-v1","entry":{"releaseId":"sample"}}
EOF
  cat > "$source_root/release-evidence/staging/release-evidence-summary.json" <<'EOF'
{"schemaVersion":"release-evidence-summary-v1","environment":"staging"}
EOF

  (
    cd "$fixture"
    ./scripts/export-release-evidence.sh staging \
      "$source_root/release-catalog/staging" \
      "$source_root/release-ledgers/staging" \
      "$source_root/release-evidence/staging" \
      "$target_root" >/tmp/test-dev-scripts-export-release-evidence.out
  )

  assert_file_exists "$target_root/release-evidence-audit-catalog.json"
  assert_file_exists "$target_root/latest-export.txt"
  export_name="$(cat "$target_root/latest-export.txt")"
  assert_file_exists "$target_root/exports/$export_name/export-manifest.txt"
}

test_collect_release_evidence_exports_latest_and_previous_ledgers() {
  local fixture catalog_dir manifest_a manifest_b out_dir ledger_dir key_file
  fixture="$(new_fixture)"
  catalog_dir="$fixture/artifacts/release-catalog/staging"
  ledger_dir="$fixture/artifacts/release-ledgers/staging"
  out_dir="$fixture/artifacts/release-evidence/staging"
  key_file="$fixture/attestation-key.pem"
  manifest_a="$fixture/release-a.env"
  manifest_b="$fixture/release-b.env"
  mkdir -p "$ledger_dir"
  openssl genrsa -out "$key_file" 2048 >/dev/null 2>&1

  cat > "$manifest_a" <<'EOF'
RELEASE_ID=release-a
RELEASE_CREATED_AT=2026-03-01T00:00:00Z
RELEASE_SOURCE_SHA=sha-a
PROMOTION_SOURCE_RUN_ID=100
API_GATEWAY_IMAGE=ghcr.io/example/api-gateway@sha256:a
LOGGER_IMAGE=ghcr.io/example/logger@sha256:a
CALCULATOR_IMAGE=ghcr.io/example/calculator@sha256:a
FRONTEND_IMAGE=ghcr.io/example/frontend@sha256:a
EOF
  cat > "$manifest_b" <<'EOF'
RELEASE_ID=release-b
RELEASE_CREATED_AT=2026-03-01T01:00:00Z
RELEASE_SOURCE_SHA=sha-b
PROMOTION_SOURCE_RUN_ID=101
API_GATEWAY_IMAGE=ghcr.io/example/api-gateway@sha256:b
LOGGER_IMAGE=ghcr.io/example/logger@sha256:b
CALCULATOR_IMAGE=ghcr.io/example/calculator@sha256:b
FRONTEND_IMAGE=ghcr.io/example/frontend@sha256:b
EOF

  (
    cd "$fixture"
    ./scripts/release-catalog.sh sync-manifest "$catalog_dir/catalog.json" staging "$manifest_a" >/tmp/test-dev-scripts-collect-evidence-a.out
    ./scripts/release-catalog.sh sync-manifest "$catalog_dir/catalog.json" staging "$manifest_b" >/tmp/test-dev-scripts-collect-evidence-b.out
    ./scripts/release-catalog.sh record-operation "$catalog_dir/catalog.json" staging release-a deploy "$fixture/artifacts/deploy-reports/staging" >/tmp/test-dev-scripts-collect-evidence-record-a.out
    ./scripts/release-catalog.sh record-operation "$catalog_dir/catalog.json" staging release-b restore-drill "$fixture/artifacts/restore-drill" >/tmp/test-dev-scripts-collect-evidence-record-b.out
    ./scripts/release-catalog.sh export-ledger "$catalog_dir/catalog.json" release-b "$ledger_dir/release-ledger-release-b.json" >/tmp/test-dev-scripts-collect-evidence-ledger.out
    LEDGER_ATTESTATION_SIGNING_KEY_FILE="$key_file" ./scripts/attest-release-ledger.sh "$ledger_dir/release-ledger-release-b.json" "$ledger_dir/release-ledger-release-b.attestation.json" >/tmp/test-dev-scripts-collect-evidence-attestation.out
    ./scripts/collect-release-evidence.sh staging "$catalog_dir/catalog.json" "$ledger_dir" "$out_dir" >/tmp/test-dev-scripts-collect-evidence.out
  )

  assert_file_exists "$out_dir/release-ledger-latest.json"
  assert_file_exists "$out_dir/release-ledger-previous.json"
  assert_file_exists "$out_dir/release-evidence-summary.json"
  assert_contains "$out_dir/release-evidence-summary.md" "Release ID: release-b"
  assert_contains "$out_dir/release-evidence-summary.md" "Attestation path: $ledger_dir/release-ledger-release-b.attestation.json"
  assert_contains "$out_dir/release-evidence-summary.md" "Attestation mode: signing-key"
}

test_check_s3_lifecycle_policy_matches_expected_rules() {
  local fixture fakebin policy_file
  fixture="$(new_fixture)"
  fakebin="$(new_fakebin)"
  policy_file="$fixture/deploy/backups/s3-lifecycle-policy.example.json"

  cat > "$fakebin/aws" <<EOF
#!/usr/bin/env bash
set -euo pipefail
if [[ "\$1" == "s3api" && "\$2" == "get-bucket-lifecycle-configuration" ]]; then
  cat "$policy_file"
  exit 0
fi
exit 1
EOF
  chmod +x "$fakebin/aws"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" ./scripts/check-s3-lifecycle-policy.sh test-bucket "$policy_file" >/tmp/test-dev-scripts-check-s3-lifecycle.out
  )
}

test_release_ledger_attestation_sign_and_verify() {
  local fixture ledger_dir ledger_file key_file
  fixture="$(new_fixture)"
  ledger_dir="$fixture/artifacts/release-ledgers/global"
  ledger_file="$ledger_dir/release-ledger-release-a.json"
  key_file="$fixture/attestation-key.pem"
  mkdir -p "$ledger_dir"

  cat > "$ledger_file" <<'EOF'
{
  "schemaVersion": "release-ledger-v1",
  "catalogEnvironment": "global",
  "entry": {
    "releaseId": "release-a"
  }
}
EOF
  openssl genrsa -out "$key_file" 2048 >/dev/null 2>&1

  (
    cd "$fixture"
    LEDGER_ATTESTATION_SIGNING_KEY_FILE="$key_file" ./scripts/attest-release-ledger.sh "$ledger_file" "$ledger_dir/release-ledger-release-a.attestation.json" >/tmp/test-dev-scripts-attest-ledger.out
    ./scripts/verify-release-ledger-attestation.sh "$ledger_file" "$ledger_dir/release-ledger-release-a.attestation.json" >/tmp/test-dev-scripts-verify-ledger.out
  )

  assert_contains "$ledger_dir/release-ledger-release-a.attestation.json" "\"mode\": \"signing-key\""
}

test_operator_mtls_cert_generation_and_readiness_check() {
  local fixture cert_dir env_file
  fixture="$(new_fixture)"
  cert_dir="$fixture/deploy/observability/operator-certs"
  env_file="$fixture/.env.operator-mtls"

  cat > "$env_file" <<EOF
PROMETHEUS_OPERATOR_ACCESS_MODE=mtls
PROMETHEUS_OPERATOR_TLS_CERT_FILE=$cert_dir/server.crt
PROMETHEUS_OPERATOR_TLS_KEY_FILE=$cert_dir/server.key
PROMETHEUS_OPERATOR_CLIENT_CA_FILE=$cert_dir/client-ca.crt
EOF

  (
    cd "$fixture"
    ./scripts/generate-operator-mtls-certs.sh "$cert_dir" >/tmp/test-dev-scripts-generate-operator-mtls.out
    ./scripts/check-operator-mtls-readiness.sh "$cert_dir" "$env_file" >/tmp/test-dev-scripts-check-operator-mtls.out
  )

  assert_file_exists "$cert_dir/server.crt"
  assert_file_exists "$cert_dir/client.crt"
  assert_file_exists "$cert_dir/manifest.env"
}

test_bootstrap_playwright_linux_user_mode_creates_env_file() {
  local fixture fakebin frontend_dir log_file
  fixture="$(new_fixture)"
  fakebin="$(new_fakebin)"
  frontend_dir="$fixture/frontend"
  log_file="$fixture/playwright-bootstrap.log"
  mkdir -p "$frontend_dir"

  cat > "$fakebin/apt" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
package="${@: -1}"
touch "${package}_1.0_all.deb"
EOF
  cat > "$fakebin/dpkg-deb" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
out_dir="${@: -1}"
mkdir -p "$out_dir/usr/lib/x86_64-linux-gnu"
touch "$out_dir/usr/lib/x86_64-linux-gnu/libplaywright-test.so"
EOF
  cat > "$fakebin/bun" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "\$*" >> "$log_file"
EOF
  chmod +x "$fakebin/apt" "$fakebin/dpkg-deb" "$fakebin/bun"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" ./scripts/bootstrap-playwright-linux.sh --mode user --frontend-dir "$frontend_dir" --cache-dir "$fixture/.toolchain/playwright-libs" >/tmp/test-dev-scripts-bootstrap-playwright.out
  )

  assert_file_exists "$frontend_dir/.playwright-linux.env"
  assert_contains "$frontend_dir/.playwright-linux.env" "LD_LIBRARY_PATH="
  assert_contains "$log_file" "x playwright install chromium"
}

test_bootstrap_playwright_linux_user_mode_falls_back_to_known_package_version() {
  local fixture fakebin frontend_dir log_file apt_log
  fixture="$(new_fixture)"
  fakebin="$(new_fakebin)"
  frontend_dir="$fixture/frontend"
  log_file="$fixture/playwright-bootstrap-fallback.log"
  apt_log="$fixture/playwright-bootstrap-fallback-apt.log"
  mkdir -p "$frontend_dir"

  cat > "$fakebin/apt" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "\$*" >> "$apt_log"
package="\${@: -1}"
if [[ "\$package" == "libasound2t64" ]]; then
  exit 1
fi
touch "\${package//=/__}_1.0_all.deb"
EOF
  cat > "$fakebin/apt-cache" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == "madison" && "$2" == "libasound2t64" ]]; then
  printf '%s\n' ' libasound2t64 | 1.2.11-1build2 | http://archive.ubuntu.com/ubuntu noble/main amd64 Packages'
fi
EOF
  cat > "$fakebin/dpkg-deb" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
out_dir="${@: -1}"
mkdir -p "$out_dir/usr/lib/x86_64-linux-gnu"
touch "$out_dir/usr/lib/x86_64-linux-gnu/libplaywright-test.so"
EOF
  cat > "$fakebin/bun" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "\$*" >> "$log_file"
EOF
  chmod +x "$fakebin/apt" "$fakebin/apt-cache" "$fakebin/dpkg-deb" "$fakebin/bun"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" ./scripts/bootstrap-playwright-linux.sh --mode user --frontend-dir "$frontend_dir" --cache-dir "$fixture/.toolchain/playwright-libs" >/tmp/test-dev-scripts-bootstrap-playwright-fallback.out
  )

  assert_file_exists "$frontend_dir/.playwright-linux.env"
  assert_contains "$frontend_dir/.playwright-linux.env" "LD_LIBRARY_PATH="
  assert_contains "$apt_log" "download libasound2t64"
  assert_contains "$apt_log" "download libasound2t64=1.2.11-1build2"
  assert_contains "$log_file" "x playwright install chromium"
}

test_rehearse_release_evidence_local_creates_environment_artifacts() {
  local fixture log_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"
  log_file="$fixture/rehearse-local.log"

  cat > "$fixture/scripts/deploy-single-host.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "\${DEPLOY_REPORT_DIR:-$fixture/artifacts/deploy-reports/base}"
if [[ "\$1" == "up" ]]; then
  printf 'archived_at=20260301T000000Z\nrequest_logs=request-logs.json\n' > "\${DEPLOY_REPORT_DIR}/archive-manifest-20260301T000000Z.txt"
  printf 'deploy_mode=build\n' > "\${DEPLOY_REPORT_DIR}/deploy-manifest-20260301T000000Z.txt"
fi
printf '%s %s\n' "\$1" "\$2" >> "$log_file"
EOF
  chmod +x "$fixture/scripts/deploy-single-host.sh"

  (
    cd "$fixture"
    ./scripts/rehearse-release-evidence-local.sh ./.env.single-host "$fixture/artifacts/local-release-evidence" >/tmp/test-dev-scripts-rehearse-local.out
  )

  assert_file_exists "$fixture/artifacts/local-release-evidence/release-evidence/staging-local/release-evidence-summary.json"
  assert_file_exists "$fixture/artifacts/local-release-evidence/release-evidence/production-local/release-evidence-summary.json"
}

test_archive_runtime_report_writes_request_log_artifacts() {
  local fixture fakebin out_dir curl_count_file
  fixture="$(new_fixture)"
  fakebin="$(new_fakebin)"
  out_dir="$fixture/artifacts/archive"
  curl_count_file="$fixture/curl-count.txt"

  cat > "$fakebin/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
count_file="$curl_count_file"
count=0
if [[ -f "\$count_file" ]]; then
  count="\$(cat "\$count_file")"
fi
count=\$((count + 1))
printf '%s\n' "\$count" > "\$count_file"
case "\$count" in
  1) printf '%s\n' '{"accessToken":"token"}' ;;
  2) printf '%s\n' '{"runtime":"report"}' ;;
  3) printf '%s\n' '{"incident":"report"}' ;;
  4) printf '%s\n' '{"items":[]}' ;;
  5) printf '%s\n' '{"items":[{"path":"/api/v1/admin/request-logs?traceId=trace-a","method":"GET","ip":"127.0.0.1","traceId":"trace-a","durationMs":12,"statusCode":200,"occurredAt":"2026-03-01T00:00:00Z"}]}' ;;
  *) exit 1 ;;
esac
EOF
  chmod +x "$fakebin/curl"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" DEPLOY_ADMIN_PASSWORD=password ./scripts/archive-runtime-report.sh http://127.0.0.1:8080 admin "$out_dir" >/tmp/test-dev-scripts-archive-runtime-report.out
  )

  manifest_file="$(find "$out_dir" -name 'archive-manifest-*.txt' -print -quit)"
  [[ -n "$manifest_file" ]] || fail "expected archive manifest to be created"
  assert_contains "$manifest_file" "request_logs="
  assert_contains "$manifest_file" "request_logs_mode=minimized"
  assert_contains "$manifest_file" "request_logs_sha256="
  request_logs_file="$(find "$out_dir" -name 'request-logs-*.json' -print -quit)"
  [[ -n "$request_logs_file" ]] || fail "expected request log archive"
  assert_contains "$request_logs_file" "\"removedFields\": ["
  assert_not_contains "$request_logs_file" "\"ip\":"
  assert_not_contains "$request_logs_file" "\"traceId\":"
  assert_contains "$request_logs_file" "\"traceIdHash\":"
  assert_contains "$request_logs_file" "\"path\": \"/api/v1/admin/request-logs\""
}

test_post_deploy_check_retries_admin_token() {
  local fixture fakebin output count_file
  fixture="$(new_fixture)"
  fakebin="$(new_fakebin)"
  output="$fixture/post-deploy-check.out"
  count_file="$fixture/post-deploy-auth-count.txt"

  cat > "$fakebin/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
count_file="$count_file"
url="\${@: -1}"
case "\$url" in
  http://127.0.0.1:4321/healthz|http://127.0.0.1:4321|http://127.0.0.1:8080/health/live|http://127.0.0.1:8080/health/ready)
    exit 0
    ;;
  http://127.0.0.1:8080/api/v1/auth/token)
    count=0
    if [[ -f "\$count_file" ]]; then
      count="\$(cat "\$count_file")"
    fi
    count=\$((count + 1))
    printf '%s\n' "\$count" > "\$count_file"
    if [[ "\$count" -lt 2 ]]; then
      exit 22
    fi
    printf '%s\n' '{"accessToken":"retry-token"}'
    ;;
  http://127.0.0.1:8080/api/v1/admin/runtime-report|http://127.0.0.1:8080/api/v1/admin/runtime-incident-report|http://127.0.0.1:8080/api/v1/admin/incident-events|http://127.0.0.1:8080/api/v1/admin/request-logs?limit=5)
    printf '%s\n' '{}'
    ;;
  *)
    exit 1
    ;;
esac
EOF
  chmod +x "$fakebin/curl"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" \
      DEPLOY_ADMIN_USER=admin \
      DEPLOY_ADMIN_PASSWORD=password \
      POST_DEPLOY_AUTH_RETRIES=3 \
      POST_DEPLOY_AUTH_RETRY_DELAY_SECONDS=0 \
      ./scripts/post-deploy-check.sh http://127.0.0.1:4321 http://127.0.0.1:8080 ./missing.env >"$output"
  )

  assert_contains "$output" "post-deploy check ok: admin runtime report"
  assert_contains "$output" "post-deploy checks passed"
}

test_rollback_single_host_manifest_uses_image_mode() {
  local fixture log_file manifest_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"
  log_file="$fixture/rollback.log"
  manifest_file="$fixture/release-manifest.env"

  cat > "$fixture/scripts/deploy-single-host.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'mode=%s action=%s env=%s\n' "\${DEPLOY_MODE:-}" "\$1" "\$2" >> "$log_file"
EOF
  cat > "$manifest_file" <<'EOF'
API_GATEWAY_IMAGE=ghcr.io/example/api-gateway@sha256:test
LOGGER_IMAGE=ghcr.io/example/logger@sha256:test
CALCULATOR_IMAGE=ghcr.io/example/calculator@sha256:test
FRONTEND_IMAGE=ghcr.io/example/frontend@sha256:test
EOF
  chmod +x "$fixture/scripts/deploy-single-host.sh"

  (
    cd "$fixture"
    ./scripts/rollback-single-host.sh ./release-manifest.env ./.env.single-host >/tmp/test-dev-scripts-rollback-manifest.out
  )

  assert_contains "$log_file" "mode=image action=up env=./.env.single-host"
}

test_release_catalog_resolves_previous_for_rollback() {
  local fixture catalog_dir manifest_a manifest_b log_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"
  catalog_dir="$fixture/artifacts/release-catalog/staging"
  manifest_a="$fixture/release-a.env"
  manifest_b="$fixture/release-b.env"
  log_file="$fixture/rollback.log"

  cat > "$fixture/scripts/deploy-single-host.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'mode=%s action=%s env=%s api=%s\n' "\${DEPLOY_MODE:-}" "\$1" "\$2" "\${API_GATEWAY_IMAGE:-}" >> "$log_file"
EOF
  cat > "$manifest_a" <<'EOF'
RELEASE_ID=release-a
RELEASE_CREATED_AT=2026-03-01T00:00:00Z
RELEASE_SOURCE_SHA=sha-a
PROMOTION_SOURCE_RUN_ID=100
API_GATEWAY_IMAGE=ghcr.io/example/api-gateway@sha256:a
LOGGER_IMAGE=ghcr.io/example/logger@sha256:a
CALCULATOR_IMAGE=ghcr.io/example/calculator@sha256:a
FRONTEND_IMAGE=ghcr.io/example/frontend@sha256:a
EOF
  cat > "$manifest_b" <<'EOF'
RELEASE_ID=release-b
RELEASE_CREATED_AT=2026-03-01T01:00:00Z
RELEASE_SOURCE_SHA=sha-b
PROMOTION_SOURCE_RUN_ID=101
API_GATEWAY_IMAGE=ghcr.io/example/api-gateway@sha256:b
LOGGER_IMAGE=ghcr.io/example/logger@sha256:b
CALCULATOR_IMAGE=ghcr.io/example/calculator@sha256:b
FRONTEND_IMAGE=ghcr.io/example/frontend@sha256:b
EOF
  chmod +x "$fixture/scripts/deploy-single-host.sh"

  (
    cd "$fixture"
    ./scripts/release-catalog.sh sync-manifest "$catalog_dir/catalog.json" staging "$manifest_a" >/tmp/test-dev-scripts-release-catalog-a.out
    ./scripts/release-catalog.sh sync-manifest "$catalog_dir/catalog.json" staging "$manifest_b" >/tmp/test-dev-scripts-release-catalog-b.out
    RELEASE_CATALOG_PATH="$catalog_dir/catalog.json" ./scripts/rollback-single-host.sh previous ./.env.single-host >/tmp/test-dev-scripts-rollback-selector.out
  )

  assert_contains "$log_file" "mode=image action=up env=./.env.single-host api=ghcr.io/example/api-gateway@sha256:a"
}

test_restore_drill_existing_bundle_runs_restore_verify_and_archive() {
  local fixture fakebin bundle_dir log_file marker curl_count_file
  fixture="$(new_fixture)"
  cp "$fixture/.env.single-host.example" "$fixture/.env.single-host"
  fakebin="$(new_fakebin)"
  bundle_dir="$fixture/artifacts/backups/bundles/single-host-20260301T000000Z"
  log_file="$fixture/restore-drill.log"
  marker="restore-drill-test"
  curl_count_file="$fixture/restore-curl-count.txt"

  mkdir -p "$bundle_dir"
  cat > "$bundle_dir/manifest.env" <<'EOF'
POSTGRES_BACKUP_FILE=postgres.sql
MONGO_BACKUP_FILE=mongo.archive.gz
RESTORE_DRILL_FIXTURE_FILE=restore-drill-fixture.json
EOF
  printf 'postgres backup\n' > "$bundle_dir/postgres.sql"
  printf 'mongo backup\n' > "$bundle_dir/mongo.archive.gz"
  cat > "$bundle_dir/restore-drill-fixture.json" <<EOF
{
  "schemaVersion": "restore-drill-v2",
  "marker": "$marker",
  "users": [
    { "name": "Restore Drill Alpha $marker", "email": "$marker.alpha@example.com" },
    { "name": "Restore Drill Beta $marker", "email": "$marker.beta@example.com" },
    { "name": "Restore Drill Gamma $marker", "email": "$marker.gamma@example.com" }
  ],
  "requestLogs": [
    {
      "path": "/restore-drill/$marker/alpha",
      "method": "GET",
      "traceId": "$marker-trace-alpha",
      "durationMs": 12,
      "statusCode": 200
    },
    {
      "path": "/restore-drill/$marker/beta",
      "method": "POST",
      "traceId": "$marker-trace-beta",
      "durationMs": 34,
      "statusCode": 202
    }
  ]
}
EOF

  cat > "$fixture/scripts/deploy-single-host.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'deploy %s %s mode=%s\n' "\$1" "\$2" "\${DEPLOY_MODE:-build}" >> "$log_file"
EOF
  cat > "$fixture/scripts/restore-postgres.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'restore-postgres %s %s\n' "\$1" "\$2" >> "$log_file"
EOF
  cat > "$fixture/scripts/restore-mongo.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'restore-mongo %s %s %s\n' "\$1" "\$2" "\${3:-}" >> "$log_file"
EOF
  cat > "$fixture/scripts/archive-runtime-report.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
mkdir -p "\$3"
printf 'archived_at=20260301T000000Z\nrequest_logs=request-logs.json\nrequest_logs_sha256=test\n' > "\$3/archive-manifest-20260301T000000Z.txt"
printf 'archive %s %s %s\n' "\$1" "\$2" "\$3" >> "$log_file"
EOF
  cat > "$fixture/scripts/post-deploy-check.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'post-deploy %s %s %s\n' "\$1" "\$2" "\$3" >> "$log_file"
EOF
  chmod +x "$fixture/scripts/"*.sh

  cat > "$fakebin/docker" <<EOF
#!/usr/bin/env bash
printf 'docker %s\n' "\$*" >> "$log_file"
exit 0
EOF
  cat > "$fakebin/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
count_file="$curl_count_file"
count=0
if [[ -f "\$count_file" ]]; then
  count="\$(cat "\$count_file")"
fi
count=\$((count + 1))
printf '%s\n' "\$count" > "\$count_file"
url="\${@: -1}"
case "\$url" in
  */api/v1/auth/token)
    printf '%s\n' '{"accessToken":"token"}'
    ;;
  */api/v1/users)
    printf '%s\n' '{"data":[{"id":1,"name":"Restore Drill Alpha $marker","email":"$marker.alpha@example.com","createdAt":"2026-03-01T00:00:00Z"},{"id":2,"name":"Restore Drill Beta $marker","email":"$marker.beta@example.com","createdAt":"2026-03-01T00:00:00Z"},{"id":3,"name":"Restore Drill Gamma $marker","email":"$marker.gamma@example.com","createdAt":"2026-03-01T00:00:00Z"}]}'
    ;;
  *"traceId=$marker-trace-alpha"*)
    printf '%s\n' '{"items":[{"path":"/restore-drill/'"$marker"'/alpha","method":"GET","ip":"127.0.0.1","traceId":"'"$marker"'-trace-alpha","durationMs":12,"statusCode":200,"occurredAt":"2026-03-01T00:00:00Z"}]}'
    ;;
  *"traceId=$marker-trace-beta"*)
    printf '%s\n' '{"items":[{"path":"/restore-drill/'"$marker"'/beta","method":"POST","ip":"127.0.0.1","traceId":"'"$marker"'-trace-beta","durationMs":34,"statusCode":202,"occurredAt":"2026-03-01T00:00:00Z"}]}'
    ;;
  *)
    exit 1
    ;;
esac
EOF
  chmod +x "$fakebin/docker" "$fakebin/curl"

  (
    cd "$fixture"
    PATH="$fakebin:$PATH" \
      DEPLOY_ADMIN_USER=admin \
      DEPLOY_ADMIN_PASSWORD=password \
      RESTORE_DRILL_API_BASE_URL=http://127.0.0.1:8080 \
      RESTORE_DRILL_FRONTEND_BASE_URL=http://127.0.0.1:4321 \
      DEPLOY_REPORT_DIR=./artifacts/restore-drill \
      ./scripts/restore-drill-single-host.sh ./.env.single-host "$bundle_dir" >/tmp/test-dev-scripts-restore-drill-existing-bundle.out
  )

  assert_contains "$log_file" "restore-mongo ./.env.single-host"
  assert_contains "$log_file" "--drop"
  assert_contains "$log_file" "archive http://127.0.0.1:8080 admin ./artifacts/restore-drill"
  assert_contains "$log_file" "docker compose -f $fixture/docker-compose.yml -f $fixture/docker-compose.security.yml -f $fixture/docker-compose.single-host.yml --env-file ./.env.single-host down -v"
  assert_file_exists "$fixture/artifacts/restore-drill/fixture-verification-$marker.json"
  assert_file_exists "$fixture/artifacts/restore-drill/fixture-actual-$marker.json"
}

main() {
  echo "running: bootstrap standard"
  test_bootstrap_standard_generates_credentials
  echo "running: bootstrap minimal"
  test_bootstrap_minimal_keeps_demo_credentials
  echo "running: dev doctor"
  test_dev_doctor_detects_missing_required_tools
  echo "running: dev up/down"
  test_dev_up_and_down_compose_arguments
  echo "running: docker bin windows path translation"
  test_windows_docker_bin_translates_compose_paths
  echo "running: dev up windows curl"
  test_dev_up_windows_docker_bin_uses_runtime_env_file_and_windows_curl
  echo "running: deploy image mode"
  test_deploy_single_host_image_mode_uses_ghcr_overlay
  echo "running: operator observability overlay"
  test_deploy_single_host_operator_observability_overlay_uses_proxy_compose_file
  echo "running: operator observability mtls overlay"
  test_deploy_single_host_operator_mtls_overlay_uses_proxy_compose_file
  echo "running: backup bundle"
  test_backup_single_host_creates_bundle_manifest
  echo "running: backup versioned sync"
  test_backup_single_host_versioned_sync_writes_remote_catalog
  echo "running: backup s3 sync"
  test_backup_single_host_s3_sync_writes_object_storage_catalog
  echo "running: release evidence export"
  test_export_release_evidence_writes_versioned_audit_bundle
  echo "running: archive runtime report"
  test_archive_runtime_report_writes_request_log_artifacts
  echo "running: post deploy auth retry"
  test_post_deploy_check_retries_admin_token
  echo "running: rollback manifest"
  test_rollback_single_host_manifest_uses_image_mode
  echo "running: rollback selector"
  test_release_catalog_resolves_previous_for_rollback
  echo "running: collect release evidence"
  test_collect_release_evidence_exports_latest_and_previous_ledgers
  echo "running: s3 lifecycle drift"
  test_check_s3_lifecycle_policy_matches_expected_rules
  echo "running: ledger attestation"
  test_release_ledger_attestation_sign_and_verify
  echo "running: operator mtls certs"
  test_operator_mtls_cert_generation_and_readiness_check
  echo "running: playwright bootstrap"
  test_bootstrap_playwright_linux_user_mode_creates_env_file
  echo "running: playwright bootstrap fallback"
  test_bootstrap_playwright_linux_user_mode_falls_back_to_known_package_version
  echo "running: local release evidence rehearsal"
  test_rehearse_release_evidence_local_creates_environment_artifacts
  echo "running: restore drill existing bundle"
  test_restore_drill_existing_bundle_runs_restore_verify_and_archive
  echo "dev script tests passed"
}

main "$@"
