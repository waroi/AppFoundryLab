# Operasyonlar ve CI/CD

## 1. Onemli environment degiskenleri

- `INCIDENT_EVENT_WEBHOOK_URL`
- `INCIDENT_EVENT_WEBHOOK_HMAC_SECRET`
- `INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS`
- `BACKUP_SYNC_TARGET`
- `BACKUP_SYNC_PROFILE`
- `BACKUP_AWS_REGION`
- `BACKUP_AWS_ENDPOINT_URL`
- `BACKUP_RETENTION_DAYS`
- `BACKUP_ENCRYPTION_PASSPHRASE`
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
- `DOCKER_BIN`
- `API_GATEWAY_IMAGE`
- `LOGGER_IMAGE`
- `CALCULATOR_IMAGE`
- `FRONTEND_IMAGE`

## 2. Temel operasyonel endpoint'ler

- `GET /health/live`
- `GET /health/ready`
- `GET /metrics`
- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`
- `GET /api/v1/admin/request-logs`

## 3. Ana workflow'lar

- `appfoundrylab-ci.yml`
- `deploy-single-host-staging.yml`
- `deploy-single-host-production.yml`
- `single-host-ops.yml`
- `publish-ghcr-images.yml`
- `release-evidence-harvest.yml`
- `backup-lifecycle-drift.yml`
- `restore-drill-single-host.yml`

## 4. Tek sunucu komutlari

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

- Prometheus gateway ve logger metrics endpoint'lerini private single-host network uzerinden scrape eder
- request log kayitlari logger backend uzerinden Mongo'ya yazilir, `GET /api/v1/admin/request-logs` ile sorgulanir ve admin trace lookup UI'da gosterilir
- operator erisimi gerekiyorsa `ENABLE_OPERATOR_PROMETHEUS_ACCESS=true` ile basic-auth veya mTLS proxy katmani acilabilir
- webhook fan-out korunur, fakat artik HMAC signing ve allowlist ister
- release ledger ciktilari `attest-release-ledger.sh` ve `verify-release-ledger-attestation.sh` ile dogrulanabilir hale getirildi
- operator mTLS rollout'u icin [operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md), `generate-operator-mtls-certs.sh` ve `check-operator-mtls-readiness.sh` kanonik yoldur

## 6. Sonraki okumalar

- [deployment.md](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)
- [deployment-strategy.md](/mnt/d/w/AppFoundryLab/docs/deployment-strategy.md)
- [incident-yonetimi.md](/mnt/d/w/AppFoundryLab/docs/tr/incident-yonetimi.md)
- [operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)
