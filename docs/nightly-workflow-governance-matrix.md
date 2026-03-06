# Nightly Workflow Governance Matrix

This matrix maps night-only checks to the repository nightly workflows.

## Checklist coverage

- `release-gate-full-optional`
- `live-stack-browser-smoke-review`
- `backup-lifecycle-drift-review`

## CI reference coverage

- `release-gate-full-nightly`
- `perf-extended-nightly`
- `backup-lifecycle-drift`

Notes:
- `release-gate-full-nightly` is the canonical workflow that runs `e2e:live` via `RUN_LIVE_STACK_BROWSER_SMOKE=true`.
- The live-stack browser smoke is nightly or manually dispatched release evidence, not a merge-blocking branch protection check.
- This stays in nightly governance because it exercises the full Docker-backed stack; the workflow budgets 60 minutes while the Playwright live spec itself uses a 45 second test timeout.
