#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ZIG_WRAPPER="/tmp/appfoundrylab-zig-cc"

if [[ ! -x "$ROOT_DIR/.toolchain/zig/zig" ]]; then
  echo "missing zig toolchain: $ROOT_DIR/.toolchain/zig/zig" >&2
  exit 1
fi

cat > "$ZIG_WRAPPER" <<EOF
#!/usr/bin/env bash
args=()
for arg in "$@"; do
  if [[ "$arg" == "--target=x86_64-unknown-linux-gnu" ]]; then
    args+=("--target=x86_64-linux-gnu")
  else
    args+=("$arg")
  fi
done
exec "$ROOT_DIR/.toolchain/zig/zig" cc "\${args[@]}"
EOF
chmod +x "$ZIG_WRAPPER"

export RUSTUP_HOME="$ROOT_DIR/.toolchain/rust/rustup"
export CARGO_HOME="$ROOT_DIR/.toolchain/rust/cargo"
export PATH="$CARGO_HOME/bin:$PATH"
export CC="$ZIG_WRAPPER"
export CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER="$ZIG_WRAPPER"
export RUSTFLAGS="-C linker=$ZIG_WRAPPER"

cd "$ROOT_DIR/backend/core/calculator"
cargo test
