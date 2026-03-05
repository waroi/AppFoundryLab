#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MODE="auto"
FRONTEND_DIR="$ROOT_DIR/frontend"
CACHE_DIR="$ROOT_DIR/.toolchain/playwright-libs"
SKIP_BROWSER_INSTALL=false
PRINT_ENV=false

install_chromium() {
  if command -v bun >/dev/null 2>&1; then
    (cd "$FRONTEND_DIR" && bun x playwright install chromium)
    return
  fi

  if [[ -x "$FRONTEND_DIR/node_modules/.bin/playwright" ]]; then
    (cd "$FRONTEND_DIR" && ./node_modules/.bin/playwright install chromium)
    return
  fi

  echo "playwright installer not found: bun and local playwright binary are both unavailable" >&2
  exit 1
}

usage() {
  cat <<'EOF'
usage: ./scripts/bootstrap-playwright-linux.sh [--mode auto|system|user] [--frontend-dir <path>] [--cache-dir <path>] [--skip-browser-install] [--print-env]
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="${2:-}"
      shift 2
      ;;
    --frontend-dir)
      FRONTEND_DIR="${2:-}"
      shift 2
      ;;
    --cache-dir)
      CACHE_DIR="${2:-}"
      shift 2
      ;;
    --skip-browser-install)
      SKIP_BROWSER_INSTALL=true
      shift
      ;;
    --print-env)
      PRINT_ENV=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ "$(uname -s)" != "Linux" ]]; then
  if [[ "$SKIP_BROWSER_INSTALL" == false ]]; then
    install_chromium
  fi
  exit 0
fi

asound_package="${PLAYWRIGHT_ASOUND_PACKAGE:-libasound2t64}"
if [[ "${PLAYWRIGHT_ALLOW_LEGACY_ASOUND:-false}" == "true" ]]; then
  asound_package="libasound2"
fi
packages=(libnspr4 libnss3 "$asound_package")
env_file="$FRONTEND_DIR/.playwright-linux.env"

system_has_runtime_dependency() {
  local package_name="${1:-}"
  case "$package_name" in
    libnspr4)
      find /usr/lib /lib -name "libnspr4.so" 2>/dev/null | head -n 1 | grep -q .
      ;;
    libnss3)
      find /usr/lib /lib -name "libnss3.so" 2>/dev/null | head -n 1 | grep -q .
      ;;
    libasound2|libasound2t64)
      find /usr/lib /lib -name "libasound.so.2" 2>/dev/null | head -n 1 | grep -q .
      ;;
    *)
      return 1
      ;;
  esac
}

run_system_install() {
  local installer=()
  if [[ "$(id -u)" == "0" ]]; then
    installer=(apt-get)
  elif command -v sudo >/dev/null 2>&1; then
    installer=(sudo apt-get)
  else
    return 1
  fi

  "${installer[@]}" update
  if ! "${installer[@]}" install -y "${packages[@]}"; then
    if [[ "$asound_package" == "libasound2t64" ]]; then
      packages=(libnspr4 libnss3 libasound2)
      "${installer[@]}" install -y "${packages[@]}"
    else
      return 1
    fi
  fi
  return 0
}

run_user_install() {
  local download_dir package_name package_file
  mkdir -p "$CACHE_DIR"
  download_dir="$(mktemp -d)"
  rm -f "$env_file"

  for package_name in "${packages[@]}"; do
    if ! download_package_archive "$download_dir" "$package_name"; then
      if [[ "$package_name" == "libasound2t64" ]]; then
        if download_package_archive "$download_dir" "libasound2"; then
          package_name="libasound2"
        elif system_has_runtime_dependency "libasound2"; then
          continue
        else
          echo "failed to download package: libasound2" >&2
          exit 1
        fi
      elif system_has_runtime_dependency "$package_name"; then
        continue
      else
        echo "failed to download package: $package_name" >&2
        exit 1
      fi
    fi
    package_file="$(find "$download_dir" -maxdepth 1 -type f -name "${package_name}_*.deb" | head -n 1)"
    [[ -n "$package_file" ]] || {
      echo "downloaded package archive not found for $package_name" >&2
      exit 1
    }
    dpkg-deb -x "$package_file" "$CACHE_DIR"
  done

  python3 - "$CACHE_DIR" "$env_file" <<'PY'
import pathlib
import sys

cache_dir = pathlib.Path(sys.argv[1]).resolve()
env_file = pathlib.Path(sys.argv[2]).resolve()
lib_dirs = sorted(
    {
        str(path.parent.resolve())
        for path in cache_dir.rglob("*.so*")
        if path.is_file()
    }
)
if not lib_dirs:
    raise SystemExit("playwright user-mode bootstrap did not extract any library directories")
env_file.write_text(
    "export LD_LIBRARY_PATH={}:${{LD_LIBRARY_PATH:-}}\n".format(":".join(lib_dirs)),
    encoding="utf-8",
)
PY
}

download_package_archive() {
  local download_dir="${1:-}"
  local package_name="${2:-}"
  local version

  if (cd "$download_dir" && apt download "$package_name" >/dev/null 2>&1); then
    return 0
  fi

  while read -r version; do
    [[ -n "$version" ]] || continue
    if (cd "$download_dir" && apt download "${package_name}=${version}" >/dev/null 2>&1); then
      return 0
    fi
  done < <(apt-cache madison "$package_name" | awk '{ print $3 }')

  return 1
}

case "$MODE" in
  auto)
    if ! run_system_install; then
      run_user_install
    fi
    ;;
  system)
    run_system_install
    ;;
  user)
    run_user_install
    ;;
  *)
    echo "unsupported mode: $MODE" >&2
    exit 1
    ;;
esac

if [[ "$SKIP_BROWSER_INSTALL" == false ]]; then
  install_chromium
fi

if [[ "$PRINT_ENV" == true && -f "$env_file" ]]; then
  cat "$env_file"
fi
