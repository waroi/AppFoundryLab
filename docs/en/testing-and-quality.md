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

Bootstrap the repo-local Go toolchain once:

```bash
./scripts/bootstrap-go-toolchain.sh
```

Go tests with isolated caches:

```bash
./scripts/go-test.sh
```

Rust worker tests:

```bash
cd backend/core/calculator
cargo test
```

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

`check-doc-drift.sh` now checks both required doc touch points and semantic truth around:
- safe `archive-runtime-report.sh` usage
- signed evidence requirements
- the mock-backed `e2e` versus real-stack `e2e:live` split

`ci-full` now includes the full release gate, and `release-gate-full-nightly.yml` enables the live-stack browser smoke with `RUN_LIVE_STACK_BROWSER_SMOKE=true`.

## 4. What each layer proves

- `smoke`: static build markers and optional API contract probes
- `e2e`: mock-backed UI regression for selectors, locale/theme screenshots, and unhappy-path UI states
- `e2e:live`: the Docker-backed admin login, runtime diagnostics, and trace lookup path that the user can reproduce in a browser
- `go-test.sh`: the backend test suite running against the repo-local Go baseline declared in `backend/go.mod`
- `dev-up`: readiness plus one authenticated admin smoke before reporting success
- `release-gate.sh full`: repo-side static checks, Go tests, Rust tests, and frontend build/smoke

## 5. Current posture

- The prior toolchain, `SystemStatus`, and `ci-full` drift items are closed.
- The dependency degradation contract is now documented in [dependency-degradation-runbook.md](/mnt/d/w/AppFoundryLab/docs/dependency-degradation-runbook.md) and exposed through `GET /api/v1/admin/runtime-config`.
- `PROGRESS.md` is the only canonical source for still-open repo backlog.
