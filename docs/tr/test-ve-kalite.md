# Test ve Kalite

## 1. Frontend dogrulama katmanlari

Hizli utility ve component dogrulamasi:

```bash
cd frontend
../.toolchain/bun/bin/bun run lint
../.toolchain/bun/bin/bun run check
../.toolchain/bun/bin/bun run build
../.toolchain/bun/bin/bun run smoke
../.toolchain/bun/bin/bun run test
```

Mock-backed browser regresyonu:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e
```

Gercek yerel stack browser smoke:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e:live
```

`bun` PATH'te yoksa `../.toolchain/bun/bin/bun` kullanin.

## 2. Backend ve worker dogrulamasi

Repo-local Go toolchain'i bir kez bootstrap edin:

```bash
./scripts/bootstrap-go-toolchain.sh
```

Izole cache ile Go testleri:

```bash
./scripts/go-test.sh
```

Rust worker testleri:

```bash
cd backend/core/calculator
cargo test
```

## 3. Script ve release kapilari

```bash
./scripts/test-dev-scripts.sh
./scripts/local-ci-smoke.sh
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
./scripts/quality-gate.sh ci-full
./scripts/check-doc-drift.sh --mode strict
./scripts/check-release-policy-drift.sh
./scripts/release-gate.sh fast
./scripts/release-gate.sh full
```

`check-doc-drift.sh` artik yalnizca degisen dosyalari degil; su dogruluk kurallarini da denetler:
- `archive-runtime-report.sh` icin guvenli env/stdin kullanimi
- signed evidence gereksinimleri
- mock-backed `e2e` ile gercek stack `e2e:live` ayrimi

`ci-full` artik tam release gate'i calistirir; `release-gate-full-nightly.yml` ise `RUN_LIVE_STACK_BROWSER_SMOKE=true` ile live-stack browser smoke'u da acik getirir.

## 4. Her katman neyi kanitlar

- `smoke`: static build isaretcileri ve opsiyonel API kontrat probe'lari
- `e2e`: selector, locale/theme screenshot'lari ve unhappy-path UI durumlari icin mock-backed regresyon
- `e2e:live`: kullanicinin tarayicidan birebir tekrar edebilecegi Docker-backed admin login, runtime diagnostics ve trace lookup akisi
- `go-test.sh`: `backend/go.mod` icindeki repo-local Go baseline'i ile calisan backend test suiti
- `dev-up`: basari demeden once readiness ve bir authenticated admin smoke
- `release-gate.sh full`: repo ici static kontroller, Go testleri, Rust testleri ve frontend build/smoke

## 5. Guncel durum

- Onceki toolchain, `SystemStatus` ve `ci-full` drift maddeleri kapatildi.
- Dependency degradation kontrati artik [dependency-degradation-runbook.md](/mnt/d/w/AppFoundryLab/docs/dependency-degradation-runbook.md) icinde belgelenir ve `GET /api/v1/admin/runtime-config` uzerinden yayinlanir.
- Halen acik repo backlog'un tek kanonik kaynagi `PROGRESS.md` dosyasidir.
