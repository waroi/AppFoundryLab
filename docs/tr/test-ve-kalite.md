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

Go testleri:

```bash
cd backend
/mnt/d/w/AppFoundryLab/.toolchain/go/bin/go test ./...
```

Rust worker testleri:

```bash
cd backend/core/calculator
cargo test
```

Host toolchain `backend/go.mod` ile ayni baseline'da degilse, Faz 1 toolchain hizalamasina kadar container build'leri ve odakli yerel kontroller gecici fallback olarak ele alinmalidir.

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

`ci-full` artik `ci-fast` kopyasi degil; tam release gate'i de calistirir.

## 4. Her katman neyi kanitlar

- `smoke`: static build isaretcileri ve opsiyonel API kontrat probe'lari
- `e2e`: selector, locale/theme davranisi ve mock-backed UI regresyonu
- `e2e:live`: kullanicinin tarayicidan birebir tekrar edebilecegi Docker-backed happy path
- `dev-up`: basari demeden once readiness ve bir authenticated admin smoke
- `release-gate.sh full`: repo ici static kontroller, Go testleri, Rust testleri ve frontend build/smoke

## 5. Halen acik bosluklar

- `SystemStatus.svelte` halen fazla buyuk ve parcali bakima ihtiyac duyuyor
- auth ve runtime hata dallari icin frontend coverage zayif
- repo-ici Go toolchain hizalamasi `PROGRESS.md` icinde halen acik
