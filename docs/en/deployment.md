# Deployment

Recommended starting point:

- keep the repository as a monorepo
- validate the production-like single-host package locally in `build` mode
- promote the same shape to a VPS
- then use `publish-ghcr-images.yml` to promote one digest-pinned manifest through staging rehearsal and production

## Main files

- [docker-compose.single-host.yml](/mnt/d/w/AppFoundryLab/docker-compose.single-host.yml)
- [deploy/docker-compose.single-host.ghcr.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.single-host.ghcr.yml)
- [deploy/docker-compose.observability.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.yml)
- [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml)
- [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml)
- [scripts/deploy-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/deploy-single-host.sh)
- [scripts/post-deploy-check.sh](/mnt/d/w/AppFoundryLab/scripts/post-deploy-check.sh)
- [scripts/archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh)
- [scripts/backup-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/backup-single-host.sh)
- [scripts/restore-drill-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/restore-drill-single-host.sh)
- [scripts/rollback-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/rollback-single-host.sh)
- [scripts/release-catalog.sh](/mnt/d/w/AppFoundryLab/scripts/release-catalog.sh)
- [scripts/collect-release-evidence.sh](/mnt/d/w/AppFoundryLab/scripts/collect-release-evidence.sh)
- [scripts/attest-release-ledger.sh](/mnt/d/w/AppFoundryLab/scripts/attest-release-ledger.sh)
- [scripts/verify-release-ledger-attestation.sh](/mnt/d/w/AppFoundryLab/scripts/verify-release-ledger-attestation.sh)
- [scripts/export-release-evidence.sh](/mnt/d/w/AppFoundryLab/scripts/export-release-evidence.sh)
- [scripts/rehearse-release-evidence-local.sh](/mnt/d/w/AppFoundryLab/scripts/rehearse-release-evidence-local.sh)
- [scripts/check-s3-lifecycle-policy.sh](/mnt/d/w/AppFoundryLab/scripts/check-s3-lifecycle-policy.sh)
- [scripts/generate-operator-mtls-certs.sh](/mnt/d/w/AppFoundryLab/scripts/generate-operator-mtls-certs.sh)
- [scripts/check-operator-mtls-readiness.sh](/mnt/d/w/AppFoundryLab/scripts/check-operator-mtls-readiness.sh)
- [scripts/bootstrap-playwright-linux.sh](/mnt/d/w/AppFoundryLab/scripts/bootstrap-playwright-linux.sh)
- [.env.single-host.example](/mnt/d/w/AppFoundryLab/.env.single-host.example)
- [deploy/backups/s3-lifecycle-policy.example.json](/mnt/d/w/AppFoundryLab/deploy/backups/s3-lifecycle-policy.example.json)
- [docs/operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)

## Local first

```bash
cp .env.single-host.example .env.single-host
./scripts/deploy-single-host.sh up ./.env.single-host
./scripts/backup-single-host.sh ./.env.single-host
./scripts/restore-drill-single-host.sh ./.env.single-host
./scripts/rehearse-release-evidence-local.sh ./.env.single-host
```

The restore drill now seeds a deterministic business fixture, stores `restore-drill-fixture.json` in newly created bundles, and writes canonical verification artifacts under `artifacts/restore-drill`. Backup bundles also maintain `backup-catalog.json` plus `latest-bundle.txt` for versioned local, SSH-copy, or S3/object-storage targets.

Runtime archive usage is now env/stdin-first:

```bash
export DEPLOY_ADMIN_PASSWORD='replace-me'
./scripts/archive-runtime-report.sh https://api.example.com admin ./artifacts/runtime-archive
printf '%s' "$DEPLOY_ADMIN_PASSWORD" | ./scripts/archive-runtime-report.sh --password-stdin https://api.example.com admin ./artifacts/runtime-archive
```

The exported request-log evidence is minimized by default so IPs and raw query strings do not leak into the archive.

If Docker Desktop is exposed as `docker.exe` instead of a Linux `docker` binary, export `DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"` before using the deploy scripts.

## VPS next

1. Provision Ubuntu LTS with Docker Engine and the Compose plugin.
2. Clone the repository on the host.
3. Copy `.env.single-host.example` to `.env.single-host`.
4. Replace every placeholder secret, backup target, and optional GHCR image refs.
5. Store pinned SSH `known_hosts` in GitHub environment variables.
6. Run either `build` mode or `image` mode deploy.

## CI/CD automation

Remote workflows now exist for:

- [deploy-single-host-staging.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-staging.yml)
- [deploy-single-host-production.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-production.yml)
- [single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml)
- [publish-ghcr-images.yml](/mnt/d/w/AppFoundryLab/.github/workflows/publish-ghcr-images.yml)
- [release-evidence-harvest.yml](/mnt/d/w/AppFoundryLab/.github/workflows/release-evidence-harvest.yml)
- [backup-lifecycle-drift.yml](/mnt/d/w/AppFoundryLab/.github/workflows/backup-lifecycle-drift.yml)
- [restore-drill-single-host.yml](/mnt/d/w/AppFoundryLab/.github/workflows/restore-drill-single-host.yml)

These workflows now cover pinned-host SSH, runtime archive evidence, release-evidence summaries, long-term audit export, signed ledger attestations, backup bundles, release catalog and ledger updates, restore drills, S3 lifecycle drift checks, automatic GHCR manifest promotion, runtime frontend API configuration, and optional operator-facing Prometheus proxying with basic auth or mTLS.

Treat signed attestation as required for the release-oriented image workflows. The practical inputs are:
- `RELEASE_EVIDENCE_AUDIT_TARGET`
- `RELEASE_LEDGER_ATTESTATION_KEY` or the script-level `LEDGER_ATTESTATION_SIGNING_KEY`
- optional `RELEASE_LEDGER_ATTESTATION_KEY_ID`

## Read next

- [deployment-strategy.md](/mnt/d/w/AppFoundryLab/docs/deployment-strategy.md)
- [operations.md](/mnt/d/w/AppFoundryLab/docs/en/operations.md)
- [incident-response.md](/mnt/d/w/AppFoundryLab/docs/en/incident-response.md)
- [operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)
