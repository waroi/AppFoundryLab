# Incident Yonetimi

Runtime kaynakli problemleri incelemek icin ilk durak admin diagnostics paneli ve incident endpoint'leridir.

## Temel endpoint'ler

- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`

## Incident report neler icerir

- onerilen severity
- incident kategorisi
- baslik ve kisa ozet
- eslesen runbook'lar
- sonraki aksiyonlar
- health, alert ve logger durumundan toplanan kanitlar

## Kalici incident journal davranisi

- gateway, alert acildiginda, dedupe penceresinden sonra anlamli bicimde guncellendiginde veya cozuldugunde event uretir
- logger servisi bu event'leri MongoDB'de saklar
- monitor artik `logger`, `stdout`, `webhook` veya bunlarin kombinasyonlarina fan-out yapabilir
- admin UI en son event'leri gostererek problemin yeni mi tekrarli mi oldugunu anlamayi kolaylastirir

## Operasyonel devam adimlari

- deploy sonrasi diagnostics artifact'larini [archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh) ile `DEPLOY_ADMIN_PASSWORD` veya `--password-stdin` kullanarak arsivle
- eski incident kayitlarini [prune-incident-events.sh](/mnt/d/w/AppFoundryLab/scripts/prune-incident-events.sh) ile temizle
- uzaktaki prune veya rollback islemleri icin [single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml) kullan

## Ilgili dokumanlar

- Kanonik incident akisi: [runtime-incident-response.md](/mnt/d/w/AppFoundryLab/docs/runtime-incident-response.md)
- Deployment: [deployment.md](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)
