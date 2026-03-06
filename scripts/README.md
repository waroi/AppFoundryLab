# Scripts Index

Bu klasor repo'nun yerel gelistirme, kalite, deployment ve evidence akisini tasir.

## Local Dev

- `dev-doctor.sh`: host gereksinimlerini ve Docker/Compose erisilebilirligini kontrol eder
- `bootstrap.sh`: `.env`, `.env.docker.local` ve local dev cert'lerini hazirlar
- `dev-up.sh`: stack'i build edip kaldirir; readiness ve authenticated smoke dogrular
- `dev-down.sh`: stack'i kapatir; `--volumes` ile local reset yapar

## Quality ve Validation

- `bootstrap-go-toolchain.sh`: `backend/go.mod` ile hizali repo-local Go toolchain kurar
- `go-test.sh`: repo-local Go baseline ve izole cache ile backend testlerini kosar
- `quality-gate.sh`: sandbox-safe, host-strict, ci-fast ve ci-full modlarini toplar
- `test-dev-scripts.sh`: dev script davranislarini fixture tabanli dogrular
- `local-ci-smoke.sh`: script + release policy + worker helper zinciri
- `release-gate.sh`: fast veya full repo release gate
- `check-doc-drift.sh`: dokuman touch-point ve semantik truthfulness kontrolu
- `check-release-policy-drift.sh`: release policy ile workflow/scripts drift kontrolu

## Frontend ve Browser

- `bootstrap-playwright-linux.sh`: Linux icin Playwright runtime kutuphanelerini hazirlar

## Deploy ve Ops

- `deploy-single-host.sh`: single-host deploy lifecycle
- `rollback-single-host.sh`: single-host rollback
- `post-deploy-check.sh`: deploy sonrasi health ve admin smoke kontrolu
- `archive-runtime-report.sh`: runtime report ve redacted request-log archive cikarir; env veya `--password-stdin` kullanir

## Backup ve Restore

- `backup-single-host.sh`: backup bundle uretir
- `restore-drill-single-host.sh`: restore drill ve verification artifact uretir
- `backup-postgres.sh`, `backup-mongo.sh`, `restore-postgres.sh`, `restore-mongo.sh`: datastore odakli yardimcilar

## Evidence ve Governance

- `release-catalog.sh`: release manifest katalogu uretir
- `collect-release-evidence.sh`: evidence summary seti toplar
- `export-release-evidence.sh`: evidence paketlerini hedefe tasir
- `attest-release-ledger.sh`: release ledger attestation uretir
- `verify-release-ledger-attestation.sh`: attestation dogrular
