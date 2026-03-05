# Integration Smoke

Bu klasor, template tabanli servislerde minimum davranis dogrulama icin baseline test paketi sunar.

Secim yardimi:
- Hangi runtime/test varyantini sececegini netlestirmek icin once `docs/variant-selection-guide.md` dosyasindaki karar tablosuna bak.

## HTTP Smoke
- Script: `tests/integration/smoke-http.sh`
- Hedef: canli servis endpointlerinin temel health/readiness davranisini dogrulamak.
- CI ornegi: `tests/integration/ci-github-actions-snippet.md`
- Process-mode CI ornegi: `tests/integration/ci-github-actions-process-mode-snippet.md`
- Minimal compose ornegi: `docker-compose.minimal.yml`
- Security compose override: `docker-compose.security.yml`
- Process-mode local runner: `scripts/run-local.sh`
- Opsiyonel load shedding smoke: `tests/integration/load-shed-smoke.sh`
- Process-mode smoke wrapper: `tests/integration/process-mode-smoke.sh`

Kullanim:
```bash
BASE_URL=http://127.0.0.1:8080 \
HEALTH_PATH=/health \
READY_PATH=/health/ready \
./tests/integration/smoke-http.sh
```

Notlar:
- `READY_PATH` varsayilaninda `200` beklenir.
- `EXPECT_READY_503=true` ise `503` degrade durumu da kabul edilir.
- Servis kontratina gore path/env degerlerini degistir.
- Compose orneginde servis adini/path/env alanlarini kendi projenin kontratina gore guncelle.
- Daha sert baseline gerekiyorsa `docker compose -f docker-compose.minimal.yml -f docker-compose.security.yml up` kullan.
- Docker kullanmiyorsan servisi `APP_CMD="..." ./scripts/run-local.sh` ile ayaga kaldirip ayni smoke scriptlerini kullan.
- Daha tekrarlanabilir process-mode akisi icin `APP_CMD="..." ./tests/integration/process-mode-smoke.sh` kullan.

## Optional Load Shedding Smoke
- Script: `tests/integration/load-shed-smoke.sh`
- Hedef: servisinizde opt-in load shedding varsa overload endpointinin `503`, health endpointinin `200` dondugunu dogrulamak.

Kullanim:
```bash
BASE_URL=http://127.0.0.1:8080 \
OVERLOAD_PATH=/internal/test/overload \
HEALTH_PATH=/health \
./tests/integration/load-shed-smoke.sh
```

## Process-mode Smoke
- Script: `tests/integration/process-mode-smoke.sh`
- Hedef: servisi process-mode olarak kaldirip health/readiness smoke'unu tek komutla calistirmak.

Kullanim:
```bash
APP_CMD="go run ./cmd/service" \
BASE_URL=http://127.0.0.1:8080 \
HEALTH_PATH=/health \
READY_PATH=/health/ready \
./tests/integration/process-mode-smoke.sh
```

Opsiyonel load shedding smoke:
```bash
APP_CMD="go run ./cmd/service" \
RUN_LOAD_SHED_SMOKE=true \
OVERLOAD_PATH=/internal/test/overload \
./tests/integration/process-mode-smoke.sh
```
