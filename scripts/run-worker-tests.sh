#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ZIG_WRAPPER="/tmp/appfoundrylab-zig-cc"

if [[ -x "$ROOT_DIR/.toolchain/zig/zig" ]]; then
  ZIG_BIN="$ROOT_DIR/.toolchain/zig/zig"
elif command -v zig >/dev/null 2>&1; then
  ZIG_BIN="$(command -v zig)"
else
  ZIG_BIN=""
fi

if [[ -n "$ZIG_BIN" ]]; then
  cat > "$ZIG_WRAPPER" <<EOF
#!/usr/bin/env bash
args=()
for arg in "\$@"; do
  if [[ "\$arg" == "--target=x86_64-unknown-linux-gnu" ]]; then
    args+=("--target=x86_64-linux-gnu")
  else
    args+=("\$arg")
  fi
done
exec "$ZIG_BIN" cc "\${args[@]}"
EOF
  chmod +x "$ZIG_WRAPPER"
  export CC="$ZIG_WRAPPER"
  export CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER="$ZIG_WRAPPER"
  export RUSTFLAGS="-C linker=$ZIG_WRAPPER"
fi

cd "$ROOT_DIR/backend/core/calculator"

if [[ -d "$ROOT_DIR/.toolchain/rust/cargo/bin" ]]; then
  export RUSTUP_HOME="$ROOT_DIR/.toolchain/rust/rustup"
  export CARGO_HOME="$ROOT_DIR/.toolchain/rust/cargo"
  export PATH="$CARGO_HOME/bin:$PATH"
fi

cargo test
