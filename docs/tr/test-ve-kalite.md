# Test ve Kalite

## 1. Backend testleri

Tum Go testlerini calistir:

```bash
cd backend && go test ./...
```

Odakli bir entegrasyon testi:

```bash
cd backend && go test ./services/api-gateway/cmd/api-gateway -run TestIntegrationAuthProtectedWorkerLoggerMetrics
```

## 2. Worker testleri

```bash
cd backend/core/calculator && cargo test
```

Ortamda sistem `cc` yoksa depo icindeki yardimci scripti kullanin:

```bash
./scripts/run-worker-tests.sh
```

## 3. Frontend kontrolleri

```bash
cd frontend && bun run lint
cd frontend && ./node_modules/.bin/astro check
cd frontend && ./node_modules/.bin/astro build
cd frontend && node ./scripts/smoke.mjs
cd frontend && bun run e2e:bootstrap
cd frontend && ./scripts/run-playwright.sh
```

Opsiyonel API baglantili smoke:

```bash
cd frontend && SMOKE_API_BASE_URL=http://127.0.0.1:8080 node ./scripts/smoke.mjs
```

## 4. Governance kontrolleri

```bash
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
./scripts/test-dev-scripts.sh
./scripts/local-ci-smoke.sh
./scripts/check-toolchain.sh
./scripts/check-doc-drift.sh --mode strict
./scripts/check-release-policy-drift.sh
./scripts/release-gate.sh fast
```

Notlar:

- `./scripts/quality-gate.sh sandbox-safe`, izin kisitli sandbox ortamlar icin varsayilan kapidir; worker dogrulamasinin acik skip moduna dusmesine izin verir
- `./scripts/quality-gate.sh host-strict`, PR acmadan once gelistirici makinesinde onerilen tam kapidir
- CI `./scripts/quality-gate.sh ci-fast`, nightly kapsama ise `./scripts/quality-gate.sh ci-full` kullanir
- admin runtime diagnostics artik frontend panelinin kullandigi ayni JSON icinde alert-odakli ozet, breach sayisi ve onerilen aksiyonlar da sunar
- runtime diagnostics yolu artik cache'lenmis snapshot'i tekrar kullanir, external probe'lari paralel toplar ve request loglarini cekirdek admin raporundan sonra yukler
- incident report ve kalici incident journal handler'lari artik gateway handler test paketinde odakli sekilde dogrulanir
- `node ./scripts/smoke.mjs` artik locale-sensitive sayfa metni yerine SSR-stable frontend isaretcilerini dogrular
- locale/theme dogrulamasi `/`, `/test`, `/tr` ve `/tr/test`, sag ust toolbar, URL gecisleri, theme kaliciligi ve `html[lang]` ile `html[data-theme]` uzerinden yapilmalidir
- frontend e2e selector'lari gorunur cevrilmis metin yerine `data-testid` veya `data-*` isaretcilerini tercih etmelidir
- `./scripts/test-dev-scripts.sh`, gercek workspace'i bozmadan temp fixture icinde `bootstrap`, `dev-doctor`, `dev-up` ve `dev-down` davranisini dogrular
- `./scripts/test-dev-scripts.sh`, buna ek olarak S3 backup sync, release-evidence summary export, audit export, ledger attestation, operator mTLS sertifika uretimi/hazirlik kontrolu, local evidence rehearsal ve Playwright bootstrap davranisini, Linux runtime kutuphaneleri icin package-version fallback dahil olacak sekilde dogrular
- `./scripts/local-ci-smoke.sh`, dev script testleri, release policy drift ve worker helper dogrulamasini tek akista toplar
- `local-ci-smoke` icin varsayilan `RUN_WORKER_TESTS=auto` modudur; izin kisitli sandbox ortamlarini acikca skip eder, `RUN_WORKER_TESTS=true` ise strict davranir
- `./scripts/rehearse-release-evidence-local.sh`, katalog, ledger, attestation, summary ve audit-export akislarinin gercek yerel deploy uzerinde birlikte calistigini kanitlayan kanonik repo ici dogrulamadir

## 5. Performans kontrolleri

```bash
./scripts/run-k6-smoke.sh
./scripts/run-k6-scenario.sh spike
./scripts/run-k6-scenario.sh soak
```

## 6. Yeni test yazarken

Kurallar:

- Pozitif ve negatif senaryo ekle
- Yetkilendirme hatalarini test et
- Contract sekli degisiyorsa onu test et
- Davranis env var'a bagliysa operasyonel kenar durumlari test et
- Frontend sunum degisikliklerinde locale degisimi, localized route navigation, theme degisimi ve theme kaliciligini test et
- Locale/theme davranisi degisiyorsa `html[lang]` ve `html[data-theme]` assertion'i ekle
- Stabil selector veya ham `data-*` degeri ayni davranisi ifade edebiliyorsa tek bir cevrilmis gorunur metne bagli assertion yazma

## 7. Kalite calismasi ne zaman tamam sayilir

Genelde su durumda guvenli kabul edilir:

- ilgili yerel testler yesilse
- CI ile uyumlu komutlar geciyorsa
- dokuman guncellendiyse
- yeni env var veya endpoint'ler dokumante edildiyse
