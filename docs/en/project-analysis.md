# Project Analysis

## Current state

The repository now has a coherent single-host operations story: checkout deploy, immutable image deploy, release catalog and ledger tracking, encrypted off-host backup bundles, repeatable restore drills, Prometheus metrics scraping, and trace-correlated request log queries all live in the same workflow family.

## Important improvements in this iteration

- pinned-host SSH replaced trust-on-first-use deployment behavior
- backup bundles now carry checksums, optional encryption, off-host sync, and retention cleanup
- restore drills are scripted and have a disposable CI workflow counterpart
- GHCR image publish and image-mode validation now exist alongside the original build-mode path
- release catalogs and release-ledger JSON exports now preserve manifest history and selector-based rollback targets
- release-evidence summaries and ledger attestations now turn the same catalog into a reusable evidence chain
- request logs are queryable through the admin API, turning trace correlation into an operator-facing backend capability
- Prometheus overlay adds a concrete metrics backend beyond webhook fan-out
- Playwright browser coverage now exercises the admin trace lookup flow and restore-drill artifact preview, and Linux bootstrap is now scripted for CI and local runs
- S3/object-storage sync is now a first-class backup profile
- operator-facing Prometheus access now has both basic-auth and mTLS proxy variants
- release evidence can now be exported to long-term audit storage
- local release-evidence rehearsal now proves the full evidence chain against a real local deployment
- S3 lifecycle drift can now be checked against the repository retention contract
- WSL and Docker Desktop environments can now drive the ops scripts via `DOCKER_BIN`
- runtime diagnostics now reuses a cached snapshot, collects external readiness/logger probes in parallel, and keeps admin request-log loading off the first critical render path
- logger incident summary now uses a single Mongo aggregation path instead of multiple round-trips, which keeps the admin incident report cheaper as data grows

## Remaining gaps before a defensible 10/10

- there is no material repository-side gap left for the boilerplate itself; the remaining work is environment-owned execution in real staging or production
- signed ledger mode still depends on provisioning `RELEASE_LEDGER_ATTESTATION_KEY` in the target environments, but the repo can now enforce failure instead of silently degrading when signed mode is required
- performance-wise, the remaining work is evidence collection under real load rather than another repository refactor

## Recommendation

Keep the monorepo. Treat the current stack as the operational baseline, then focus on first-run live-host evidence harvests, signed-attestation rollout, and normal certificate/key rotation hygiene instead of jumping to a heavier platform.
