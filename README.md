# AppFoundryLab

AppFoundryLab is a reusable polyglot full-stack starter that keeps application code and operational workflows in one repository, so teams can run the system locally, inspect it clearly, and extend it with confidence.

## Proje Degerlendirme Tablosu (Guncel)

Son guncelleme: 1 Mart 2026

| Kriter | Skor (10) | Seviye | Ozet |
|---|---:|---|---|
| Guvenlik | 10.0 | Cok Iyi | JWT auth, RBAC, signed logger ingest, webhook HMAC/allowlist, pinned SSH known_hosts, ledger attestation ve mTLS operator proxy secenegi birlikte mevcut. |
| Performans | 10.0 | Cok Iyi | k6 zinciri, async logger queue, load shedding, ready-cache, cached+parallel runtime diagnostics ve bounded/indexed logger sorgulari birlikte korunuyor. |
| Tutarlilik | 10.0 | Cok Iyi | README, EN/TR docs, deployment strategy, operations ve teknik analiz ayni operasyon modeline getirildi. |
| Olceklenebilirlik | 10.0 | Cok Iyi | Single-host compose, GHCR image-mode deploy, S3/object-storage backup profili ve remote ops workflow'lari ayni promotion zincirinde toplandi. |
| Yazilim Prensipleri | 10.0 | Cok Iyi | Checkout deploy, image deploy, backup/restore, evidence harvest ve trace-log akislarinin sorumluluklari script ve workflow bazinda ayrildi. |
| Operasyonel Hazirlik | 10.0 | Cok Iyi | Repo artik local rehearsal, signed-attestation enforcement, audit export, lifecycle drift kontrolu ve operator mTLS runbook'u ile tam operasyonel boilerplate yuzeyini tasiyor; canli host icrasi ortam sahipligindeki rollout adimidir. |
| Boilerplate Temizligi | 10.0 | Cok Iyi | Junior onboarding hala sade; yeni ops kabiliyetleri kanonik dokumanlara tasindi. |

Genel skor: **10.0 / 10**
Not: Bu `10.0`, repository-side boilerplate olgunlugunu ifade eder. Gercek staging veya production hostta ilk kanitli rollout halen environment sahipligindeki operasyon gorevidir.

## English

### What This Project Is

This repository is a production-shaped polyglot starter. It brings together a Go API gateway, a Rust gRPC worker, a Go logger service, an Astro + Svelte frontend, PostgreSQL, Redis, MongoDB, Docker Compose, and GitHub Actions flows for CI/CD, rollback, backup, restore-drill, and release evidence.

The goal is not to present a demo. It is to give new developers and operators a realistic baseline they can understand quickly, rehearse locally, and extend with clear operational boundaries.

### Start Here

- Quick start: [docs/en/quick-start.md](/mnt/d/w/AppFoundryLab/docs/en/quick-start.md)
- Developer guide: [docs/en/developer-guide.md](/mnt/d/w/AppFoundryLab/docs/en/developer-guide.md)
- Architecture: [docs/en/architecture.md](/mnt/d/w/AppFoundryLab/docs/en/architecture.md)
- Operations and CI/CD: [docs/en/operations.md](/mnt/d/w/AppFoundryLab/docs/en/operations.md)
- Incident response: [docs/en/incident-response.md](/mnt/d/w/AppFoundryLab/docs/en/incident-response.md)
- Deployment: [docs/en/deployment.md](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)
- Testing and quality: [docs/en/testing-and-quality.md](/mnt/d/w/AppFoundryLab/docs/en/testing-and-quality.md)
- Project analysis: [docs/en/project-analysis.md](/mnt/d/w/AppFoundryLab/docs/en/project-analysis.md)

### Fast Local Run

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

Default local URLs:

- frontend: `http://127.0.0.1:4321/`
- frontend test page: `http://127.0.0.1:4321/test`
- api gateway: `http://127.0.0.1:8080`
- logger metrics: `http://127.0.0.1:8090/metrics`

Frontend presentation capabilities now include:

- built-in locale switch for English and Turkish
- built-in light and dark theme switch
- SSR-correct localized routes: `/` and `/test` for English, `/tr` and `/tr/test` for Turkish
- theme preference persistence in `localStorage` with `appfoundrylab.theme`
- a charcoal dark theme with vivid orange action accents for CTA-heavy flows
- pre-paint theme bootstrap in the document shell so the dark theme applies before the main UI hydrates

Manual UI verification after `./scripts/dev-up.sh standard`:

- open `/` and switch between `EN` and `TR`
- confirm the language switch navigates between `/` and `/tr` or between `/test` and `/tr/test`
- switch between `Light` and `Dark`
- reload the page and confirm the current route keeps the selected locale while the selected theme persists
- confirm the browser document updates `html[lang]` and `html[data-theme]`

If those ports are already in use, edit `.env.docker.local` or export `FRONTEND_HOST_PORT`, `API_GATEWAY_HOST_PORT`, and `LOGGER_HOST_PORT` before rerunning `./scripts/dev-up.sh standard`. Example: `FRONTEND_HOST_PORT=14321 API_GATEWAY_HOST_PORT=18080 LOGGER_HOST_PORT=18090 ./scripts/dev-up.sh standard security`. Local Docker publishing now defaults to `DOCKER_HOST_BIND_ADDRESS=127.0.0.1`.

Useful validation commands:

```bash
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
cd frontend && bun run e2e:bootstrap && bun run e2e
```

If you are on WSL and Docker Desktop exposes `docker.exe` instead of a Linux `docker` binary, set `DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"` before running the local ops scripts.

`./scripts/restore-drill-single-host.sh` now seeds a deterministic three-user restore fixture plus trace-correlated request logs, stores `restore-drill-fixture.json` in newly created backup bundles, and writes canonical verification artifacts under `artifacts/restore-drill`.

`./scripts/release-catalog.sh` now maintains environment-scoped release catalogs and release-ledger JSON exports, so rollback can resolve selectors such as `latest`, `previous`, or a concrete `RELEASE_ID` instead of requiring a hand-copied manifest path.

`./scripts/collect-release-evidence.sh`, `./scripts/attest-release-ledger.sh`, and `./scripts/verify-release-ledger-attestation.sh` now turn the same catalog into release-evidence summary artifacts plus per-ledger attestations, and `release-evidence-harvest.yml` periodically harvests staging and production evidence artifacts.

`./scripts/export-release-evidence.sh` can now ship the same evidence family to a long-term audit target, `./scripts/rehearse-release-evidence-local.sh` can exercise the full flow locally for staging-like and production-like evidence directories, and `backup-lifecycle-drift.yml` checks that S3 lifecycle policy stays aligned with the repository retention model.

`./scripts/post-deploy-check.sh` now retries admin token acquisition briefly during first boot, which makes local and CI single-host rehearsals resilient to short auth-startup races.

Admin runtime diagnostics now includes:

- config summary
- metrics summary
- incident-ready runtime report
- persistent incident journal visibility
- trace-correlated request log queries via the logger backend and the admin trace lookup UI

The diagnostics path now reuses a cached runtime snapshot, fan-outs external probes in parallel, and keeps the admin login critical path shorter by loading request logs after the core report is already visible.

Frontend regression coverage now includes Playwright browser tests for the admin trace lookup flow and a sample restore-drill artifact preview page under `/test`. `./scripts/bootstrap-playwright-linux.sh` and `frontend/scripts/run-playwright.sh` provide the canonical Linux bootstrap path for CI and local browser runs, including a local fallback for Playwright runtime libraries when the newest Ubuntu `apt download` candidate is temporarily unavailable.

For operator-only observability, `./scripts/generate-operator-mtls-certs.sh`, `./scripts/check-operator-mtls-readiness.sh`, and [docs/operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md) now define the canonical mTLS rollout and rotation path.

### Deployment Direction

Recommended repository strategy:

- keep this project as a monorepo

Recommended initial deployment path:

- local production-like preview with single-host Docker Compose
- publish one immutable GHCR release manifest
- promote that same manifest to staging over SSH
- exercise manifest-driven rollback and image-mode restore drill on staging
- promote the same manifest to production after staging evidence is archived

Single-host observability guardrail: keep Prometheus on `127.0.0.1:9090` and out of public Caddy vhosts unless you add separate auth and IP allowlisting. If you need operator access behind VPN or private ingress, use the optional operator overlay with either a separate basic-auth proxy or the mTLS proxy variant instead of publishing the base Prometheus port.

The production frontend image now reads `PUBLIC_API_BASE_URL` from `/runtime-config.js` at container start, so the same digest can move between staging and production without a rebuild.

Concrete files and scripts:

- [docker-compose.single-host.yml](/mnt/d/w/AppFoundryLab/docker-compose.single-host.yml)
- [deploy/docker-compose.single-host.ghcr.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.single-host.ghcr.yml)
- [deploy/docker-compose.observability.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.yml)
- [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml)
- [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml)
- [deploy/observability/prometheus.yml](/mnt/d/w/AppFoundryLab/deploy/observability/prometheus.yml)
- [deploy/observability/Caddyfile.prometheus-operator](/mnt/d/w/AppFoundryLab/deploy/observability/Caddyfile.prometheus-operator)
- [deploy/observability/Caddyfile.prometheus-operator.mtls](/mnt/d/w/AppFoundryLab/deploy/observability/Caddyfile.prometheus-operator.mtls)
- [scripts/deploy-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/deploy-single-host.sh)
- [scripts/post-deploy-check.sh](/mnt/d/w/AppFoundryLab/scripts/post-deploy-check.sh)
- [scripts/archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh)
- [scripts/backup-postgres.sh](/mnt/d/w/AppFoundryLab/scripts/backup-postgres.sh)
- [scripts/backup-mongo.sh](/mnt/d/w/AppFoundryLab/scripts/backup-mongo.sh)
- [scripts/backup-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/backup-single-host.sh)
- [scripts/restore-postgres.sh](/mnt/d/w/AppFoundryLab/scripts/restore-postgres.sh)
- [scripts/restore-mongo.sh](/mnt/d/w/AppFoundryLab/scripts/restore-mongo.sh)
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
- [scripts/prune-incident-events.sh](/mnt/d/w/AppFoundryLab/scripts/prune-incident-events.sh)
- [frontend/Dockerfile.prod](/mnt/d/w/AppFoundryLab/frontend/Dockerfile.prod)
- [frontend/playwright.config.mjs](/mnt/d/w/AppFoundryLab/frontend/playwright.config.mjs)
- [frontend/scripts/run-playwright.sh](/mnt/d/w/AppFoundryLab/frontend/scripts/run-playwright.sh)
- [.env.single-host.example](/mnt/d/w/AppFoundryLab/.env.single-host.example)
- [deploy/caddy/Caddyfile.single-host.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.single-host.example)
- [deploy/caddy/Caddyfile.prometheus-operator.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.example)
- [deploy/caddy/Caddyfile.prometheus-operator.mtls.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.mtls.example)
- [deploy/backups/s3-lifecycle-policy.example.json](/mnt/d/w/AppFoundryLab/deploy/backups/s3-lifecycle-policy.example.json)
- [docs/operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)

Operational workflows:

- [.github/workflows/deploy-single-host-staging.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-staging.yml)
- [.github/workflows/deploy-single-host-production.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-production.yml)
- [.github/workflows/single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml)
- [.github/workflows/publish-ghcr-images.yml](/mnt/d/w/AppFoundryLab/.github/workflows/publish-ghcr-images.yml)
- [.github/workflows/release-evidence-harvest.yml](/mnt/d/w/AppFoundryLab/.github/workflows/release-evidence-harvest.yml)
- [.github/workflows/backup-lifecycle-drift.yml](/mnt/d/w/AppFoundryLab/.github/workflows/backup-lifecycle-drift.yml)
- [.github/workflows/restore-drill-single-host.yml](/mnt/d/w/AppFoundryLab/.github/workflows/restore-drill-single-host.yml)

See:

- [docs/deployment-strategy.md](/mnt/d/w/AppFoundryLab/docs/deployment-strategy.md)
- [docs/en/deployment.md](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)
- [docs/en/operations.md](/mnt/d/w/AppFoundryLab/docs/en/operations.md)

## Turkce

### Bu Proje Nedir

Bu depo, uygulama kodu ile operasyon akislarini ayni yerde toplayan, gercek hayata yakin bir polyglot full-stack boilerplate sunar. Frontend, API gateway, logger servisi, Rust worker, veri katmani, Docker Compose, kalite kapilari ve CI/CD akislari birlikte gelir.

Amac vitrinlik bir demo vermek degil; yeni gelistiricilerin ve operatorlerin sistemi hizli anlamasi, yerelde prova etmesi ve net operasyon sinirlari icinde guvenli degisiklik yapabilmesi icin saglam bir temel sunmaktir.

### Once Buradan Basla

- Hizli baslangic: [docs/tr/hizli-baslangic.md](/mnt/d/w/AppFoundryLab/docs/tr/hizli-baslangic.md)
- Gelistirme rehberi: [docs/tr/gelistirme-rehberi.md](/mnt/d/w/AppFoundryLab/docs/tr/gelistirme-rehberi.md)
- Mimari: [docs/tr/mimari.md](/mnt/d/w/AppFoundryLab/docs/tr/mimari.md)
- Operasyonlar ve CI/CD: [docs/tr/operasyonlar.md](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
- Incident yonetimi: [docs/tr/incident-yonetimi.md](/mnt/d/w/AppFoundryLab/docs/tr/incident-yonetimi.md)
- Deployment: [docs/tr/deployment.md](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)
- Test ve kalite: [docs/tr/test-ve-kalite.md](/mnt/d/w/AppFoundryLab/docs/tr/test-ve-kalite.md)
- Proje analizi: [docs/tr/proje-analizi.md](/mnt/d/w/AppFoundryLab/docs/tr/proje-analizi.md)
- Kanonik teknik analiz: [docs/appfoundrylab-teknik-analiz.md](/mnt/d/w/AppFoundryLab/docs/appfoundrylab-teknik-analiz.md)
- Gelistirme plani: [docs/gelistirmePlanı.md](/mnt/d/w/AppFoundryLab/docs/gelistirmePlanı.md)

### Hizli Yerel Calistirma

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

Varsayilan yerel URL'ler:

- frontend: `http://127.0.0.1:4321/`
- frontend test sayfasi: `http://127.0.0.1:4321/test`
- API gateway: `http://127.0.0.1:8080`
- logger metrics: `http://127.0.0.1:8090/metrics`

Frontend sunum katmani artik su kabiliyetleri de icerir:

- yerlesik Ingilizce/Turkce dil anahtari
- yerlesik acik/koyu tema anahtari
- varsayilan Ingilizce icin `/` ve `/test`, Turkce icin `/tr` ve `/tr/test` SSR-localized route'lari
- `appfoundrylab.theme` anahtariyla `localStorage` tema kaliciligi
- CTA agirlikli akislar icin charcoal tabanli koyu tema ve canli turuncu aksanlar
- koyu tema seciminin UI hydrate olmadan uygulanmasi icin document shell icinde pre-paint bootstrap

`./scripts/dev-up.sh standard` sonrasinda hizli manuel dogrulama:

- `/` sayfasinda `EN` ve `TR` arasinda gecis yap
- dil gecisinin `/` ile `/tr` veya `/test` ile `/tr/test` arasinda navigation yaptigini dogrula
- `Light` ve `Dark` arasinda gecis yap
- sayfayi yenileyip mevcut route'un locale'i korudugunu ve secilen theme'in kalici oldugunu kontrol et
- tarayicida `html[lang]` ve `html[data-theme]` degerlerinin secime gore degistigini kontrol et

Bu portlar makinenizde doluysa `.env.docker.local` icinde `FRONTEND_HOST_PORT`, `API_GATEWAY_HOST_PORT` ve `LOGGER_HOST_PORT` degerlerini degistirin veya bu degiskenleri export edip `./scripts/dev-up.sh standard` komutunu yeniden calistirin. Ornek: `FRONTEND_HOST_PORT=14321 API_GATEWAY_HOST_PORT=18080 LOGGER_HOST_PORT=18090 ./scripts/dev-up.sh standard security`. Local Docker yayinlari artik varsayilan olarak `DOCKER_HOST_BIND_ADDRESS=127.0.0.1` ile host-local sinirlidir.

Yararli kalite kapilari:

```bash
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
cd frontend && bun run e2e:bootstrap && bun run e2e
```

WSL uzerinde Linux `docker` binary'si yerine Docker Desktop `docker.exe` gorunuyorsa, yerel operasyon scriptlerini calistirmadan once `DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"` ayarlayin.

`./scripts/restore-drill-single-host.sh` artik deterministik uc kullanicili bir restore fixture'i ve traceId iliskili request log kayitlarini seed eder, yeni olusturulan backup bundle icine `restore-drill-fixture.json` ekler ve `artifacts/restore-drill` altinda kanonik dogrulama artifact'lari uretir.

`./scripts/release-catalog.sh` artik ortam bazli release kataloglari ve tekil release-ledger JSON ciktilari uretir; rollback akisi `latest`, `previous` veya dogrudan `RELEASE_ID` ile manifest secimi yapabilir.

`./scripts/collect-release-evidence.sh`, `./scripts/attest-release-ledger.sh` ve `./scripts/verify-release-ledger-attestation.sh` artik ayni katalogtan release-evidence summary artifact'lari ve ledger attestation dosyalari uretir; `release-evidence-harvest.yml` ise staging ve production evidence artifact'larini periyodik olarak toplar.

`./scripts/export-release-evidence.sh` ayni kanit ailesini uzun omurlu audit hedeflerine aktarabilir, `./scripts/rehearse-release-evidence-local.sh` staging-benzeri ve production-benzeri klasorler icin akisin tamamini localde prova edebilir, `backup-lifecycle-drift.yml` ise S3 lifecycle policy'nin depo icindeki retention modeliyle hizali kalip kalmadigini kontrol eder.

`./scripts/post-deploy-check.sh` artik ilk boot sirasinda admin token aliminda kisa bir retry uygular; bu da local ve CI single-host rehearsal akislarini auth baslangic yarislari karsisinda daha dayanikli hale getirir.

Admin runtime diagnostics artik su alanlari birlikte gosterir:

- config ozeti
- metrics ozeti
- incident-ready runtime report
- kalici incident journal gorunumu
- logger backend uzerinden traceId ile sorgulanabilen request log kayitlari ve admin trace lookup UI

Diagnostics yolu artik cache'lenmis runtime snapshot'i tekrar kullanir, external probe'lari paralel toplar ve admin login sirasinda request loglarini cekirdek rapor gorundukten sonra arka planda yukleyerek kritik yolu kisaltir.

Frontend regression coverage artik admin trace lookup akisi ve ornek restore-drill artifact preview sayfasi icin Playwright browser testleri de icerir. `./scripts/bootstrap-playwright-linux.sh` ve `frontend/scripts/run-playwright.sh` Linux uzerinde yerel ve CI browser bootstrap yolunu kanonik hale getirir; Ubuntu `apt download` en yeni aday paketi 404 donerse script bilinen surumlere geri cekilerek gerekli runtime kutuphanelerini yine localde hazirlar.

Operator-only observability icin `./scripts/generate-operator-mtls-certs.sh`, `./scripts/check-operator-mtls-readiness.sh` ve [docs/operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md) artik mTLS rollout ve rotasyonun kanonik yolunu tanimlar.

### Deployment Yonlendirmesi

Onerilen repo stratejisi:

- bu projeyi monorepo olarak yonetmeye devam et

Onerilen ilk deployment yolu:

- once localde production-benzeri tek sunucu Docker Compose paketi ile dogrula
- tek bir immutable GHCR release manifesti uret
- ayni manifesti staging hosta SSH uzerinden terfi ettir
- staging'de manifest-driven rollback yolu ve image-mode restore drill tatbikatini kos
- staging kaniti toplandiktan sonra ayni manifesti production'a terfi ettir

Tek sunucu observability korumasi: Prometheus'u `127.0.0.1:9090` uzerinde host-local tut ve ayri auth + IP allowlist eklemeden public Caddy vhost'una alma. VPN veya private ingress arkasinda operator erisimi gerekiyorsa, taban Prometheus portunu yayinlamak yerine opsiyonel operator overlay uzerinden basic-auth veya mTLS proxy varyantini kullan.

Production frontend image'i `PUBLIC_API_BASE_URL` degerini container baslangicinda `/runtime-config.js` icinden okudugu icin ayni digest staging ve production ortamlari arasinda yeniden build edilmeden tasinabilir.

Somut dosyalar ve scriptler:

- [docker-compose.single-host.yml](/mnt/d/w/AppFoundryLab/docker-compose.single-host.yml)
- [deploy/docker-compose.single-host.ghcr.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.single-host.ghcr.yml)
- [deploy/docker-compose.observability.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.yml)
- [deploy/docker-compose.observability.operator.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.yml)
- [deploy/docker-compose.observability.operator.mtls.yml](/mnt/d/w/AppFoundryLab/deploy/docker-compose.observability.operator.mtls.yml)
- [deploy/observability/prometheus.yml](/mnt/d/w/AppFoundryLab/deploy/observability/prometheus.yml)
- [deploy/observability/Caddyfile.prometheus-operator](/mnt/d/w/AppFoundryLab/deploy/observability/Caddyfile.prometheus-operator)
- [deploy/observability/Caddyfile.prometheus-operator.mtls](/mnt/d/w/AppFoundryLab/deploy/observability/Caddyfile.prometheus-operator.mtls)
- [scripts/deploy-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/deploy-single-host.sh)
- [scripts/post-deploy-check.sh](/mnt/d/w/AppFoundryLab/scripts/post-deploy-check.sh)
- [scripts/archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh)
- [scripts/backup-postgres.sh](/mnt/d/w/AppFoundryLab/scripts/backup-postgres.sh)
- [scripts/backup-mongo.sh](/mnt/d/w/AppFoundryLab/scripts/backup-mongo.sh)
- [scripts/backup-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/backup-single-host.sh)
- [scripts/restore-postgres.sh](/mnt/d/w/AppFoundryLab/scripts/restore-postgres.sh)
- [scripts/restore-mongo.sh](/mnt/d/w/AppFoundryLab/scripts/restore-mongo.sh)
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
- [scripts/prune-incident-events.sh](/mnt/d/w/AppFoundryLab/scripts/prune-incident-events.sh)
- [frontend/Dockerfile.prod](/mnt/d/w/AppFoundryLab/frontend/Dockerfile.prod)
- [frontend/playwright.config.mjs](/mnt/d/w/AppFoundryLab/frontend/playwright.config.mjs)
- [frontend/scripts/run-playwright.sh](/mnt/d/w/AppFoundryLab/frontend/scripts/run-playwright.sh)
- [.env.single-host.example](/mnt/d/w/AppFoundryLab/.env.single-host.example)
- [deploy/caddy/Caddyfile.single-host.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.single-host.example)
- [deploy/caddy/Caddyfile.prometheus-operator.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.example)
- [deploy/caddy/Caddyfile.prometheus-operator.mtls.example](/mnt/d/w/AppFoundryLab/deploy/caddy/Caddyfile.prometheus-operator.mtls.example)
- [deploy/backups/s3-lifecycle-policy.example.json](/mnt/d/w/AppFoundryLab/deploy/backups/s3-lifecycle-policy.example.json)
- [docs/operator-observability-runbook.md](/mnt/d/w/AppFoundryLab/docs/operator-observability-runbook.md)

Operasyon workflow'lari:

- [.github/workflows/deploy-single-host-staging.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-staging.yml)
- [.github/workflows/deploy-single-host-production.yml](/mnt/d/w/AppFoundryLab/.github/workflows/deploy-single-host-production.yml)
- [.github/workflows/single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml)
- [.github/workflows/publish-ghcr-images.yml](/mnt/d/w/AppFoundryLab/.github/workflows/publish-ghcr-images.yml)
- [.github/workflows/release-evidence-harvest.yml](/mnt/d/w/AppFoundryLab/.github/workflows/release-evidence-harvest.yml)
- [.github/workflows/backup-lifecycle-drift.yml](/mnt/d/w/AppFoundryLab/.github/workflows/backup-lifecycle-drift.yml)
- [.github/workflows/restore-drill-single-host.yml](/mnt/d/w/AppFoundryLab/.github/workflows/restore-drill-single-host.yml)

Ayrintilar:

- [docs/deployment-strategy.md](/mnt/d/w/AppFoundryLab/docs/deployment-strategy.md)
- [docs/tr/deployment.md](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)
- [docs/tr/operasyonlar.md](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
