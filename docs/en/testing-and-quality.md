# Testing and Quality

## 1. Backend tests

Run all Go tests:

```bash
cd backend && go test ./...
```

Run a focused integration test:

```bash
cd backend && go test ./services/api-gateway/cmd/api-gateway -run TestIntegrationAuthProtectedWorkerLoggerMetrics
```

## 2. Worker tests

```bash
cd backend/core/calculator && cargo test
```

If the environment does not provide a system `cc`, use the repository helper:

```bash
./scripts/run-worker-tests.sh
```

## 3. Frontend checks

```bash
cd frontend && bun run lint
cd frontend && ./node_modules/.bin/astro check
cd frontend && ./node_modules/.bin/astro build
cd frontend && node ./scripts/smoke.mjs
cd frontend && bun run e2e:bootstrap
cd frontend && ./scripts/run-playwright.sh
```

Optional API-backed smoke:

```bash
cd frontend && SMOKE_API_BASE_URL=http://127.0.0.1:8080 node ./scripts/smoke.mjs
```

## 4. Governance checks

```bash
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
./scripts/test-dev-scripts.sh
./scripts/local-ci-smoke.sh
./scripts/check-toolchain.sh
./scripts/check-doc-drift.sh --mode strict
./scripts/check-release-policy-drift.sh
./scripts/release-gate.sh fast
```

Notes:

- `./scripts/quality-gate.sh sandbox-safe` is the default for permission-limited sandboxes; it allows worker validation to degrade to explicit skip mode
- `./scripts/quality-gate.sh host-strict` is the recommended developer-machine gate before opening a PR
- CI uses `./scripts/quality-gate.sh ci-fast`, while nightly coverage uses `./scripts/quality-gate.sh ci-full`
- admin runtime diagnostics now exposes alert-oriented summaries, breach counts, and recommended actions in the same JSON used by the frontend diagnostics panel
- the runtime diagnostics path now reuses a cached snapshot, fans external probes out in parallel, and loads request logs after the core admin report is already visible
- incident report and persistent incident journal handlers now have focused backend tests as part of the gateway handler suite
- `node ./scripts/smoke.mjs` now checks SSR-stable frontend markers instead of locale-sensitive page copy
- locale/theme verification should cover `/`, `/test`, `/tr`, and `/tr/test`, plus the top-right toolbar, URL transitions, theme reload persistence, and `html[lang]` plus `html[data-theme]`
- frontend e2e selectors should prefer `data-testid` or `data-*` attributes over visible translated text
- `./scripts/test-dev-scripts.sh` validates `bootstrap`, `dev-doctor`, `dev-up`, and `dev-down` in temp fixtures without touching the real workspace
- `./scripts/test-dev-scripts.sh` also validates S3 backup sync, release-evidence summary export, audit export, ledger attestation, operator mTLS cert generation/readiness, local evidence rehearsal, and Playwright bootstrap behavior, including package-version fallback for local Linux runtime libs
- `./scripts/local-ci-smoke.sh` chains dev script tests, release policy drift, and worker helper validation
- `RUN_WORKER_TESTS=auto` is the default for `local-ci-smoke`; permission-limited sandboxes are skipped explicitly, while `RUN_WORKER_TESTS=true` keeps it strict
- `./scripts/rehearse-release-evidence-local.sh` is the canonical repo-side proof that catalog, ledger, attestation, summary, and audit-export flows still work together against a real local deploy

## 5. Performance checks

```bash
./scripts/run-k6-smoke.sh
./scripts/run-k6-scenario.sh spike
./scripts/run-k6-scenario.sh soak
```

## 6. Writing new tests

Rules:

- Add positive and negative cases
- Test authorization failures
- Test contract shape changes
- Test operational edge cases when behavior depends on env vars
- For frontend presentation changes, test locale switching, localized route navigation, theme switching, and theme reload persistence
- Assert `html[lang]` and `html[data-theme]` when locale/theme behavior changes
- Avoid assertions that depend on one translated visible string when a stable selector or raw `data-*` value can express the same behavior

## 7. When quality work is complete

You are usually in a safe state when:

- local targeted tests are green
- CI-relevant commands pass
- docs are updated
- new env vars or endpoints are documented
