# Deployment

Onerilen baslangic yolu:

- repoyu monorepo olarak koru
- production-benzeri tek sunucu paketini once localde `build` modunda dogrula
- ayni sekli bir VPS'e tasi
- sonra `publish-ghcr-images.yml` ile tek bir digest-pinned manifesti staging tatbikati ve production'a terfi ettir

## Temel dosyalar

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

## Once local

```bash
cp .env.single-host.example .env.single-host
./scripts/deploy-single-host.sh up ./.env.single-host
./scripts/backup-single-host.sh ./.env.single-host
./scripts/restore-drill-single-host.sh ./.env.single-host
./scripts/rehearse-release-evidence-local.sh ./.env.single-host
```

Restore drill artik deterministik bir business fixture seed eder, yeni olusturulan bundle'lara `restore-drill-fixture.json` ekler ve `artifacts/restore-drill` altinda kanonik dogrulama artifact'lari uretir. Backup bundle'lari da versioned local, SSH-copy veya S3/object-storage hedefleri icin `backup-catalog.json` ve `latest-bundle.txt` uretir.

Docker Desktop yalnizca `docker.exe` sagliyorsa, deploy scriptlerini calistirmadan once `DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"` export edin.

## Sonra VPS

1. Ubuntu LTS, Docker Engine ve Compose plugin kur.
2. Repoyu sunucuya klonla.
3. `.env.single-host.example` dosyasini `.env.single-host` olarak kopyala.
4. Tum placeholder secret, backup hedefi ve opsiyonel GHCR image ref degerlerini doldur.
5. GitHub ortam degiskenlerinde pinned SSH `known_hosts` tut.
6. `build` veya `image` modunda deploy calistir.

## CI/CD otomasyonu

Artik su workflow'lar var:

- [deploy-single-host-staging.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-staging.yml)
- [deploy-single-host-production.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-production.yml)
- [single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml)
- [publish-ghcr-images.yml](/mnt/d/w/AppFoundryLab/.github/workflows/publish-ghcr-images.yml)
- [release-evidence-harvest.yml](/mnt/d/w/AppFoundryLab/.github/workflows/release-evidence-harvest.yml)
- [backup-lifecycle-drift.yml](/mnt/d/w/AppFoundryLab/.github/workflows/backup-lifecycle-drift.yml)
- [restore-drill-single-host.yml](/mnt/d/w/AppFoundryLab/.github/workflows/restore-drill-single-host.yml)

Bu workflow'lar pinned-host SSH, runtime archive kaniti, release-evidence ozetleri, uzun omurlu audit export, opsiyonel signed enforcement ile ledger attestation, backup bundle, release katalogu ve ledger guncellemeleri, restore drill, S3 lifecycle drift check, otomatik GHCR manifest promotion, runtime frontend API konfigurasyonu ve basic-auth veya mTLS operator Prometheus proxy katmanini kapsar.

## Devaminda oku

- [deployment-strategy.md](/mnt/d/w/AppFoundryLab/docs/deployment-strategy.md)
- [operasyonlar.md](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
- [incident-yonetimi.md](/mnt/d/w/AppFoundryLab/docs/tr/incident-yonetimi.md)
- [operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)
