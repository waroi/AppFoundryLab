# Testing and Quality

## 1. Frontend validation layers

Fast utility and component validation:

```bash
cd frontend
../.toolchain/bun/bin/bun run lint
../.toolchain/bun/bin/bun run check
../.toolchain/bun/bin/bun run build
../.toolchain/bun/bin/bun run smoke
../.toolchain/bun/bin/bun run test
```

Mock-backed browser regression:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e
```

Real local-stack browser smoke:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e:live
```

Use `../.toolchain/bun/bin/bun` when `bun` is not on your PATH.

## 2. Backend and worker validation

Go tests:

```bash
cd backend
/mnt/d/w/AppFoundryLab/.toolchain/go/bin/go test ./...
```

Rust worker tests:

```bash
cd backend/core/calculator
cargo test
```

If the host toolchain does not satisfy `backend/go.mod`, treat container builds plus targeted local checks as the temporary fallback until Phase 1 toolchain alignment is complete.

## 3. Script and release gates

```bash
./scripts/test-dev-scripts.sh
./scripts/local-ci-smoke.sh
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
./scripts/quality-gate.sh ci-full
./scripts/check-doc-drift.sh --mode strict
./scripts/check-release-policy-drift.sh
./scripts/release-gate.sh fast
./scripts/release-gate.sh full
```

`ci-full` now includes the full release gate instead of mirroring `ci-fast`.

## 4. What each layer proves

- `smoke`: static build markers and optional API contract probes
- `e2e`: mock-backed UI regression for selectors and locale/theme behavior
- `e2e:live`: the Docker-backed happy path that the user can reproduce in a browser
- `dev-up`: readiness plus one authenticated admin smoke before reporting success
- `release-gate.sh full`: repo-side static checks, Go tests, Rust tests, and frontend build/smoke

## 5. Current open gaps

- `SystemStatus.svelte` is still too large and needs decomposition
- frontend component coverage for auth and runtime-error branches is still thin
- repo-local Go toolchain alignment remains open in `PROGRESS.md`
