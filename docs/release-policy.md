# Release Policy

This short policy keeps the release process understandable and compatible with the repository automation.

## Canonical sources

- Checklist JSON: `docs/release-checklist.json`
- Delivery governance: `docs/delivery-workflow-governance-matrix.md`
- Nightly governance: `docs/nightly-workflow-governance-matrix.md`
- Branch protection mapping: `docs/branch-protection-required-checks.md`
- Load shedding policy: `docs/load-shedding-policy.json`
- Load shedding runbook: `docs/load-shedding-runbook.md`

## Required workflow references

- `script-quality-gate-ci`: script quality gate with strict host/CI worker validation
- `release-gate-fast`: Release Gate (fast)
- `release-gate-full-nightly`: Release Gate (full nightly)
- `perf-extended-nightly`: nightly performance workflow
- `frontend-api-contract-matrix`: PR API contract coverage
- `profile-capacity-matrix`: profile capacity validation
- `boilerplate-quality-report`: governance coverage quality artifact
- `compare-k6-summary-vs-base-pr`: Perf benchmark smoke + trend diff
- `release-evidence-harvest`: scheduled or on-demand evidence harvest for staging and production
- `backup-lifecycle-drift`: scheduled or on-demand S3 lifecycle drift validation

## Operational workflow references

- `deploy-single-host-staging`: remote staging rollout over SSH
- `deploy-single-host-production`: remote production rollout over SSH
- `single-host-ops`: remote backup, rollback, and incident retention operations
- `release-evidence-harvest`: periodic evidence collection and optional restore-drill trigger
- `backup-lifecycle-drift`: S3 lifecycle policy drift check for backup and evidence retention

## Manual review items

- `dependabot-queue-review`
- `load-shed-policy-review`
- `release-evidence-audit-review`
- `backup-lifecycle-drift-review`

## Notes

- Use `docs/release-checklist.json` as the canonical structured checklist.
- Use `./scripts/quality-gate.sh sandbox-safe` inside permission-limited sandboxes, `host-strict` on real developer machines, and `ci-fast` / `ci-full` inside GitHub Actions.
- Keep `Release Gate (fast)` aligned with CI.
- Keep `Release Gate (full nightly)` aligned with nightly coverage.
- `Perf benchmark smoke + trend diff` must stay visible in both docs and automation.
- For single-host VPS rollouts, archive the runtime report after deploy and review the resulting artifact before calling the rollout complete.
- Review the latest release-evidence export and lifecycle drift report before treating the repository-side release proof as complete.
