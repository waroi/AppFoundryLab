# Delivery Workflow Governance Matrix

This matrix maps delivery-time governance items to the regular CI workflow.

## Checklist coverage

- `release-gate-fast`
- `ci-required-jobs-green`
- `trivy-gitleaks-review`
- `dependabot-queue-review`
- `perf-trend-diff-review`
- `load-shed-policy-review`
- `release-evidence-audit-review`

## CI reference coverage

- `release-gate-fast`
- `profile-capacity-matrix`
- `boilerplate-quality-report`
- `compare-k6-summary-vs-base-pr`
- `release-evidence-harvest`

Notes:
- Delivery-time merge gates intentionally exclude `live-stack-browser-smoke-review`; that evidence belongs to nightly or on-demand release confidence.
- Regular branch protection should stay focused on fast CI, contract coverage, and governance drift checks.
