#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
INSTALL_DIR="$ROOT_DIR/.toolchain/go"
FORCE=false

usage() {
  cat <<'EOF'
usage: ./scripts/bootstrap-go-toolchain.sh [--force]

notes:
  - installs the Go version declared in backend/go.mod under ./.toolchain/go
  - skips download when the repo-local toolchain already matches
EOF
}

parse_args() {
  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      --force)
        FORCE=true
        shift
        ;;
      -h|--help|help)
        usage
        exit 0
        ;;
      *)
        usage >&2
        exit 1
        ;;
    esac
  done
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "$1 is required" >&2
    exit 1
  }
}

target_go_version() {
  awk '/^go / { print $2; exit }' "$BACKEND_DIR/go.mod"
}

normalize_os() {
  case "$(uname -s)" in
    Linux) printf '%s\n' "linux" ;;
    Darwin) printf '%s\n' "darwin" ;;
    *)
      echo "unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

normalize_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf '%s\n' "amd64" ;;
    arm64|aarch64) printf '%s\n' "arm64" ;;
    *)
      echo "unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

installed_go_version() {
  if [[ ! -x "$INSTALL_DIR/bin/go" ]]; then
    return 1
  fi
  "$INSTALL_DIR/bin/go" version | awk '{ sub(/^go/, "", $3); print $3 }'
}

download_go_archive() {
  local version="$1"
  local platform_os="$2"
  local platform_arch="$3"
  local out_file="$4"
  local primary_url="https://go.dev/dl/go${version}.${platform_os}-${platform_arch}.tar.gz"
  local fallback_url="https://dl.google.com/go/go${version}.${platform_os}-${platform_arch}.tar.gz"

  if curl -fsSL "$primary_url" -o "$out_file"; then
    return 0
  fi
  curl -fsSL "$fallback_url" -o "$out_file"
}

main() {
  parse_args "$@"
  require_command curl
  require_command tar

  local version platform_os platform_arch current_version archive_file temp_dir
  version="$(target_go_version)"
  if [[ -z "$version" ]]; then
    echo "failed to resolve Go version from $BACKEND_DIR/go.mod" >&2
    exit 1
  fi

  if current_version="$(installed_go_version 2>/dev/null)" && [[ "$FORCE" != true ]] && [[ "$current_version" == "$version" ]]; then
    echo "repo-local Go toolchain already matches go$version"
    "$INSTALL_DIR/bin/go" version
    exit 0
  fi

  platform_os="$(normalize_os)"
  platform_arch="$(normalize_arch)"
  temp_dir="$(mktemp -d)"
  archive_file="$temp_dir/go.tar.gz"
  trap 'rm -rf "$temp_dir"' EXIT

  echo "installing Go ${version} into $INSTALL_DIR"
  download_go_archive "$version" "$platform_os" "$platform_arch" "$archive_file"

  rm -rf "$INSTALL_DIR"
  mkdir -p "$(dirname "$INSTALL_DIR")"
  tar -C "$temp_dir" -xzf "$archive_file"
  mv "$temp_dir/go" "$INSTALL_DIR"

  "$INSTALL_DIR/bin/go" version
}

main "$@"
