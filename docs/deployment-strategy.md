# Deployment Strategy

This is the canonical deployment strategy document for the repository.

## 1. Repo strategy

Recommendation:

- keep the project as a single monorepo

Why:

- frontend, gateway, logger, worker, docs, and CI rules change together
- release gates, incident docs, deployment workflows, and restore drills already assume one repository
- single-host rollout stays understandable when one version boundary owns code, scripts, and docs

## 2. Deployment path hierarchy

Recommended order:

1. run the single-host package locally in `build` mode
2. publish one digest-pinned GHCR manifest
3. promote that same manifest to staging over SSH
4. exercise manifest-driven rollback and image-mode restore drill on staging
5. promote the same manifest to production
6. only scale out after traffic or compliance actually demands it

This keeps the first production shape understandable for a junior developer and avoids premature platform complexity.

## 3. Canonical single-host package

Core files:

- [docker-compose.single-host.yml](/mnt/d/w/AppFoundryLab/docker-compose.single-host.yml)
- [deploy/docker-compose.single-host.ghcr.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.single-host.ghcr.yml)
- [deploy/docker-compose.observability.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.yml)
- [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml)
- [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml)
- [deploy/observability/prometheus.yml](/mnt/d/w/AppFoundryLab/deploy/observability/prometheus.yml)
- [scripts/deploy-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/deploy-single-host.sh)
- [scripts/post-deploy-check.sh](/mnt/d/w/AppFoundryLab/scripts/post-deploy-check.sh)
- [scripts/archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh)
- [scripts/backup-postgres.sh](/mnt/d/w/AppFoundryLab/scripts/backup-postgres.sh)
- [scripts/backup-mongo.sh](/mnt/d/w/AppFoundryLab/scripts/backup-mongo.sh)
- [scripts/backup-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/backup-single-host.sh)
- [scripts/restore-postgres.sh](/mnt/d/w/AppFoundryLab/scripts/restore-postgres.sh)
- [scripts/restore-mongo.sh](/mnt/d/w/AppFoundryLab/scripts/restore-mongo.sh)
- [scripts/restore-drill-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/restore-drill-single-host.sh)
- [scripts/prune-incident-events.sh](/mnt/d/w/AppFoundryLab/scripts/prune-incident-events.sh)
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
- [deploy/caddy/Caddyfile.single-host.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.single-host.example)
- [deploy/caddy/Caddyfile.prometheus-operator.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.example)
- [deploy/caddy/Caddyfile.prometheus-operator.mtls.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.mtls.example)

GitHub Actions workflows:

- [deploy-single-host-staging.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-staging.yml)
- [deploy-single-host-production.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-production.yml)
- [single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml)
- [publish-ghcr-images.yml](/mnt/d/w/AppFoundryLab/.github/workflows/publish-ghcr-images.yml)
- [release-evidence-harvest.yml](/mnt/d/w/AppFoundryLab/.github/workflows/release-evidence-harvest.yml)
- [backup-lifecycle-drift.yml](/mnt/d/w/AppFoundryLab/.github/workflows/backup-lifecycle-drift.yml)
- [restore-drill-single-host.yml](/mnt/d/w/AppFoundryLab/.github/workflows/restore-drill-single-host.yml)

## 4. Local production-like validation

Use the single-host package on your own machine first:

```bash
cp .env.single-host.example .env.single-host
./scripts/deploy-single-host.sh up ./.env.single-host
```

If your machine exposes Docker Desktop as `docker.exe` instead of a Linux `docker` binary, export `DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"` before using the single-host scripts.

What this does:

1. uses the secure/internal-network compose layers
2. optionally enables a private Prometheus scrape stack when `ENABLE_OBSERVABILITY_STACK=true`
3. optionally layers a separate operator-facing Prometheus proxy when `ENABLE_OPERATOR_PROMETHEUS_ACCESS=true`
4. starts the backend stack with private data services
5. archives runtime diagnostics if admin credentials are provided
6. verifies frontend, API, logger JSON metrics, logger Prometheus metrics, incident endpoints, and request-log trace queries

Useful follow-ups:

```bash
./scripts/deploy-single-host.sh ps ./.env.single-host
./scripts/deploy-single-host.sh logs ./.env.single-host
./scripts/backup-single-host.sh ./.env.single-host
./scripts/restore-drill-single-host.sh ./.env.single-host
./scripts/prune-incident-events.sh ./.env.single-host
./scripts/rehearse-release-evidence-local.sh ./.env.single-host
```

## 5. VPS baseline

### Recommended host shape

For staging-like or demo production:

- Ubuntu 24.04 LTS
- Docker Engine
- Docker Compose plugin
- Caddy or Nginx for public TLS

Hardware guidance:

- local dev: 4 vCPU, 8 GB RAM, 40 GB SSD
- small staging VPS: 4 vCPU, 8 GB RAM, 80 GB SSD
- small production VPS: 8 vCPU, 16 GB RAM, 160 GB SSD

### Step by step

1. Provision the VPS.
2. Install Docker Engine and the Compose plugin.
3. Clone the repository on the VPS at a stable path such as `/opt/appfoundrylab`.
4. Copy [.env.single-host.example](/mnt/d/w/AppFoundryLab/.env.single-host.example) to `.env.single-host`.
5. Replace every placeholder secret and domain.
6. Add pinned SSH host keys into GitHub environment vars instead of relying on `ssh-keyscan`.
7. Configure a reverse proxy with [Caddyfile.single-host.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.single-host.example) as a starting point. Keep Prometheus off public Caddy vhosts unless you add separate auth and IP allowlisting.
8. If you need operator-facing metrics behind VPN or private ingress, use [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml) and [deploy/caddy/Caddyfile.prometheus-operator.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.example) for basic auth, or [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml) and [deploy/caddy/Caddyfile.prometheus-operator.mtls.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.mtls.example) for client-certificate access instead of exposing `127.0.0.1:9090`.
9. Run:

```bash
./scripts/deploy-single-host.sh up ./.env.single-host
```

10. Confirm:

- frontend is reachable
- `/health/live` and `/health/ready` are green
- admin runtime report is reachable after login
- `GET /api/v1/admin/request-logs` returns trace-correlated request entries
- Prometheus is reachable only on `127.0.0.1:9090` when enabled

11. Archive the first runtime snapshot:

```bash
./scripts/archive-runtime-report.sh https://api.example.com admin strong_password
```

## 6. Immutable image promotion path

`publish-ghcr-images.yml` is now the canonical immutable promotion entrypoint. It builds once, validates once, writes a global release ledger plus attestation, then reuses the same release manifest through staging deploy, staging rollback-path rehearsal, staging restore drill, and production deploy.

### Publish

```bash
gh workflow run publish-ghcr-images.yml
```

This workflow now:

- builds and pushes first-party images to GHCR
- records digest-pinned image references and release metadata
- validates `DEPLOY_MODE=image` against the single-host compose stack
- uploads `release-manifest.env` and `release-manifest.json`
- exports `artifacts/release-catalog/global/catalog.json` and `artifacts/release-ledgers/global/release-ledger-<release-id>.json`
- generates `artifacts/release-evidence/global/release-evidence-summary.*` and per-ledger attestation files
- can export evidence bundles to a long-term audit target with `scripts/export-release-evidence.sh`
- promotes the same manifest to staging
- exercises manifest-driven rollback and restore-drill workflows on staging
- promotes the same manifest to production

### Deploy

Image mode expects a manifest with:

```bash
API_GATEWAY_IMAGE=ghcr.io/<owner>/<repo>/api-gateway@sha256:...
LOGGER_IMAGE=ghcr.io/<owner>/<repo>/logger@sha256:...
CALCULATOR_IMAGE=ghcr.io/<owner>/<repo>/calculator@sha256:...
FRONTEND_IMAGE=ghcr.io/<owner>/<repo>/frontend@sha256:...
```

The frontend image no longer bakes an environment-specific API base URL. `PUBLIC_API_BASE_URL` is rendered at container start into `/runtime-config.js`, so the same frontend digest can move between staging and production unchanged.

You can still source the manifest on the host and run `DEPLOY_MODE=image ./scripts/deploy-single-host.sh up ./.env.single-host`, but the canonical path is the workflow promotion chain.

Rollback supports both:

- a git ref for `build` mode
- a manifest file for `image` mode
- a release selector such as `latest`, `previous`, `release:<id>`, or `sha:<source-sha>` when `RELEASE_CATALOG_PATH` is available

## 7. Backup, restore, rollback, and retention

Expected operational baseline:

- PostgreSQL daily backup
- MongoDB daily backup
- off-host encrypted backup sync
- restore drill at least weekly on a disposable environment
- incident retention prune at least daily
- rollback must always point to a known-good git ref or digest manifest

Commands:

```bash
./scripts/backup-single-host.sh ./.env.single-host
./scripts/restore-drill-single-host.sh ./.env.single-host
./scripts/rollback-single-host.sh v0.1.0 ./.env.single-host
./scripts/rollback-single-host.sh ./artifacts/ghcr/release-manifest.env ./.env.single-host
RELEASE_CATALOG_PATH=./artifacts/release-catalog/staging/catalog.json ./scripts/rollback-single-host.sh previous ./.env.single-host
```

Important env vars:

- `BACKUP_SYNC_TARGET`
- `BACKUP_SYNC_PROFILE`
- `BACKUP_AWS_REGION`
- `BACKUP_AWS_ENDPOINT_URL`
- `BACKUP_RETENTION_DAYS`
- `BACKUP_ENCRYPTION_PASSPHRASE`
- `DOCKER_BIN`
- `ENABLE_OPERATOR_PROMETHEUS_ACCESS`
- `PROMETHEUS_OPERATOR_ACCESS_MODE`
- `PROMETHEUS_OPERATOR_BIND_ADDRESS`
- `PROMETHEUS_OPERATOR_PORT`
- `PROMETHEUS_OPERATOR_USERNAME`
- `PROMETHEUS_OPERATOR_PASSWORD_HASH`
- `PROMETHEUS_OPERATOR_TLS_CERT_FILE`
- `PROMETHEUS_OPERATOR_TLS_KEY_FILE`
- `PROMETHEUS_OPERATOR_CLIENT_CA_FILE`
- `LEDGER_ATTESTATION_REQUIRE_SIGNED`
- `INCIDENT_EVENT_WEBHOOK_URL`
- `INCIDENT_EVENT_WEBHOOK_HMAC_SECRET`
- `INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS`

`backup-single-host.sh` creates a bundle directory with:

- encrypted or plaintext Postgres dump
- encrypted or plaintext Mongo archive
- per-file SHA-256 sidecars
- `manifest.env`
- `backup-catalog.json` and `latest-bundle.txt` for versioned local, SSH-copy, or S3/object-storage targets

The same retention model can now be checked against an S3 bucket policy with `check-s3-lifecycle-policy.sh` and [deploy/backups/s3-lifecycle-policy.example.json](/mnt/d/w/AppFoundryLab/deploy/backups/s3-lifecycle-policy.example.json).

Restore drills created through `restore-drill-single-host.sh` also add `restore-drill-fixture.json` to the bundle and emit canonical `fixture-expected-*`, `fixture-actual-*`, `fixture-verification-*`, and `fixture-manifest-*` artifacts under `artifacts/restore-drill`.

Image-mode deploy, rollback, and restore-drill workflows now also maintain environment-scoped `artifacts/release-catalog/<env>/catalog.json`, `artifacts/release-ledgers/<env>/release-ledger-*.json`, per-ledger attestation files, and `artifacts/release-evidence/<env>/release-evidence-summary.*` exports.

`rehearse-release-evidence-local.sh` provides the same release-catalog, ledger, attestation, summary, and optional audit-export chain against a local single-host deployment so the repo can prove the full flow even when a real staging host is not available.

`post-deploy-check.sh` now retries admin token acquisition during first boot so startup-order jitter does not create false negatives in local or CI rehearsals.

`single-host-ops.yml` can run:

- `backup-all`
- `backup-postgres`
- `backup-mongo`
- `prune-incidents`
- `rollback`
- `restore-drill`
- `list-release-catalog`

`release-evidence-harvest.yml` periodically harvests staging and production release catalogs and can also run a scheduled staging restore drill before collection.

`backup-lifecycle-drift.yml` checks that the live S3 lifecycle policy still matches the repository retention contract for backup bundles and exported release evidence.

## 8. Observability baseline

Current single-host observability baseline is intentionally pragmatic:

- Prometheus scrapes `api-gateway:8080/metrics`
- Prometheus scrapes `logger:8090/metrics/prometheus`
- Prometheus UI/API stays host-local on `127.0.0.1:9090` and is not part of the default public Caddy surface
- optional operator access can be proxied through the separate Caddy sidecar on `PROMETHEUS_OPERATOR_PORT` with either independent basic auth or mTLS
- gateway incident diagnostics remain available through admin endpoints
- trace-correlated request logs are stored in Mongo via the logger backend and exposed through `GET /api/v1/admin/request-logs`

Use [docs/operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md) plus `generate-operator-mtls-certs.sh` and `check-operator-mtls-readiness.sh` as the canonical mTLS rollout path.

This keeps the first observability step simple while still moving beyond webhook-only fan-out.

## 9. Future move options

When traffic or organizational complexity grows, the next step should be incremental:

1. keep the monorepo
2. treat first real staging and production evidence harvests as environment execution, not as missing repository functionality
3. rotate operator mTLS material and ledger signing keys as part of normal ops hygiene
4. add a collector / OTLP topology only when metrics and trace volume justify it
5. only consider Kubernetes after the team actually needs it
