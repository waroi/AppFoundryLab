# Operations and CI/CD

## 1. Important environment variables

- `INCIDENT_EVENT_WEBHOOK_URL`
- `INCIDENT_EVENT_WEBHOOK_HMAC_SECRET`
- `INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS`
- `BACKUP_SYNC_TARGET`
- `BACKUP_SYNC_PROFILE`
- `BACKUP_AWS_REGION`
- `BACKUP_AWS_ENDPOINT_URL`
- `BACKUP_RETENTION_DAYS`
- `BACKUP_ENCRYPTION_PASSPHRASE`
- `REQUEST_LOG_TRUSTED_PROXY_CIDRS`
- `LOGGER_HEALTH_TIMEOUT_MS`
- `LOGGER_INGEST_TIMESTAMP_MAX_AGE_SECONDS`
- `LOGGER_INGEST_TIMESTAMP_MAX_FUTURE_SKEW_SECONDS`
- `ENABLE_OBSERVABILITY_STACK`
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
- `RELEASE_EVIDENCE_AUDIT_TARGET`
- `RELEASE_LEDGER_ATTESTATION_KEY`
- `RELEASE_LEDGER_ATTESTATION_KEY_ID`
- `DOCKER_BIN`
- `API_GATEWAY_IMAGE`
- `LOGGER_IMAGE`
- `CALCULATOR_IMAGE`
- `FRONTEND_IMAGE`

## 2. Main operational endpoints

- `GET /health/live`
- `GET /health/ready`
- `GET /metrics`
- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`
- `GET /api/v1/admin/request-logs`

## 3. Main workflows

- `appfoundrylab-ci.yml`
- `deploy-single-host-staging.yml`
- `deploy-single-host-production.yml`
- `single-host-ops.yml`
- `publish-ghcr-images.yml`
- `release-evidence-harvest.yml`
- `backup-lifecycle-drift.yml`
- `restore-drill-single-host.yml`

## 4. Single-host commands

```bash
./scripts/deploy-single-host.sh up ./.env.single-host
./scripts/backup-single-host.sh ./.env.single-host
./scripts/restore-drill-single-host.sh ./.env.single-host
./scripts/rollback-single-host.sh ./artifacts/ghcr/release-manifest.env ./.env.single-host
RELEASE_CATALOG_PATH=./artifacts/release-catalog/staging/catalog.json ./scripts/rollback-single-host.sh previous ./.env.single-host
./scripts/collect-release-evidence.sh staging ./artifacts/release-catalog/staging/catalog.json ./artifacts/release-ledgers/staging ./artifacts/release-evidence/staging
./scripts/export-release-evidence.sh staging ./artifacts/release-catalog/staging/catalog.json ./artifacts/release-ledgers/staging ./artifacts/release-evidence/staging ./artifacts/release-audit
./scripts/rehearse-release-evidence-local.sh ./.env.single-host
./scripts/check-s3-lifecycle-policy.sh "$BUCKET_NAME" ./deploy/backups/s3-lifecycle-policy.example.json
```

## 5. Observability baseline

- Prometheus scrapes the gateway and logger metrics endpoints on the private single-host network
- request logs are stored in Mongo through the logger backend, queried through `GET /api/v1/admin/request-logs`, and surfaced in the admin trace lookup UI
- `GET /api/v1/admin/runtime-config` and the admin diagnostics panel now surface dependency policies, trusted proxy CIDRs, and logger timing knobs from the running system
- optional operator access can be proxied separately with `ENABLE_OPERATOR_PROMETHEUS_ACCESS=true` and either `PROMETHEUS_OPERATOR_ACCESS_MODE=basic-auth` or `PROMETHEUS_OPERATOR_ACCESS_MODE=mtls`
- webhook fan-out is still available, but now requires HMAC signing and allowlisted destinations
- release ledgers are expected to be attested and verified with `attest-release-ledger.sh` plus `verify-release-ledger-attestation.sh`
- operator mTLS rollout should use [operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md), `generate-operator-mtls-certs.sh`, and `check-operator-mtls-readiness.sh`
- runtime archive exports are redacted/minimized by default and should use `DEPLOY_ADMIN_PASSWORD` or `--password-stdin` instead of positional secrets

## 6. Quality automation notes

- `release-gate-full-nightly.yml` enables `RUN_LIVE_STACK_BROWSER_SMOKE=true` and exercises `./scripts/quality-gate.sh ci-full`
- `check-doc-drift.sh --mode strict` now checks semantic truth for archive usage, signed evidence requirements, and the `e2e` versus `e2e:live` split
- Mock-backed Playwright owns keyboard/focus and degraded-state regression coverage; `e2e:live` stays nightly or on-demand because the full Docker-backed stack has a higher cost and flake surface than regular merge gates

## 7. Read next

- [deployment.md](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)
- [deployment-strategy.md](/mnt/d/w/AppFoundryLab/docs/deployment-strategy.md)
- [incident-response.md](/mnt/d/w/AppFoundryLab/docs/en/incident-response.md)
- [operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)
