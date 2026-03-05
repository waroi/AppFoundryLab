# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Derin teknik analiz seti ve kanonik master plan (`docs/gelistirmePlanı.md`).
- Detayli mimari dokuman (`docs/mimari-dokumani.md`).
- Detayli gelistirme rehberi (`docs/gelistirme-rehberi.md`).
- Detayli kullanici dokumani (`docs/kullanici-dokumani.md`).
- Release gate JSON raporu icin parser scripti (`scripts/parse-release-gate-json.py`).

### Changed
- Dokuman seti kanonik plan referansina gore yeniden duzenlendi.
- Frontend `SystemStatus` API cagrilari `/api/v1` canonical path'e tasindi.
- API gateway legacy rota defaultu profile-aware hale getirildi (`RUNTIME_PROFILE=secure` -> legacy kapali).
- `presets/*`, `.env.example` ve `.env.docker` profile-default davranisini yansitacak sekilde guncellendi.
- Redis distributed rate-limit hata davranisi profile-aware hale getirildi (`RATE_LIMIT_REDIS_FAILURE_MODE`, secure default `closed`).
- Gateway async request logging modeli bounded queue + worker pipeline'a tasindi.
- Logger queue metrics'ine drop-ratio threshold policy/alarm alanlari eklendi (`LOGGER_DROP_ALERT_THRESHOLD_PCT`).
- Frontend smoke scriptine opsiyonel API contract smoke adimi eklendi (`SMOKE_API_BASE_URL`).
- CI perf benchmark job'ina PR-base trend diff adimi eklendi (`scripts/perf/compare_k6_summary.py`).
- Release checklist scripted/manual gate modeline tasindi (`scripts/release-gate.sh`).
- Doc drift gate kanonik dokuman setine sabitlendi (`README.md`, teknik analiz, `gelistirmePlanı.md`).
- Runtime profile behavior/policy matrix dokumani genisletildi (`docs/compatibility-matrix.md`).
- Starter template'e minimal integration smoke paketi eklendi (`starter/clean-service-template/tests/integration`).
- CI workflow'da `Release Gate (fast)` job'i quality job'lar icin zorunlu prerequisite yapildi.
- CI `Release Gate (fast)` job'unda JSON artifact upload + machine-parse summary adimlari eklendi.
- `Release Gate Full Nightly` workflow'u eklendi (`.github/workflows/release-gate-full-nightly.yml`).
- `release-gate.sh` komutuna `--json` rapor modu eklendi.
- Gateway async logger retry/backoff davranisi env ile yonetilebilir hale getirildi (`LOGGER_RETRY_BACKOFF_BASE_MS`, `LOGGER_RETRY_BACKOFF_MAX_MS`).
- Logger `/metrics` threshold alarm exportu Prometheus-uyumlu alan isimleriyle genisletildi (`logger_queue_drop_alert_threshold_*`).
- Frontend API contract smoke adimi CI PR matrix profillerine yayildi (`frontend-api-contract-matrix`: `minimal`, `standard`, `secure`).
- Starter template'e minimal docker compose ornegi eklendi (`starter/clean-service-template/docker-compose.minimal.yml`).
- Release policy checklist maddeleri kanonik `docs/release-checklist.json` dosyasina tasindi.
- `release-gate.sh` JSON raporu `docs/release-checklist.json` kaynagindan uretilir hale getirildi.
- `check-doc-drift.sh` icin explicit `--mode advisory|strict` eklendi; CI `strict` mod kullanacak sekilde guncellendi.
- ADR indeks dosyasi eklendi (`docs/adr/README.md`).
- ADR karar etki matrisi eklendi (`docs/adr/decision-impact-matrix.md`).
- CI sonunda markdown/json kalite skor raporu ureten `boilerplate-quality-report` job'u eklendi.
- Kalite raporu uretimi icin `scripts/generate-quality-report.py` eklendi.
- Perf trend diff markdown raporu PR step summary'ye yazdirilir hale getirildi.
- Starter integration smoke icin GitHub Actions snippet dokumani eklendi.
- `check-release-policy-drift.sh` ile release policy/workflow drift kontrolu eklendi.
- Drift kontrolu `Release Gate (full nightly)` referansini da dogrulayacak sekilde genisletildi.
- Release/doc gate'leri `docs/gelistirmePlanı.md` kanonik plan yoluna guncellendi.
- Profile presetleri ve env ornekleri logger retry backoff parametreleriyle guncellendi.
- API gateway'e in-flight load shedding middleware'i eklendi (`MAX_INFLIGHT_REQUESTS`, `LOAD_SHED_EXEMPT_PREFIXES`).
- API gateway `/metrics` ciktisi load shedding ve in-flight serileriyle genisletildi (`api_gateway_load_shed_total`, `api_gateway_inflight_requests`, `api_gateway_inflight_requests_peak`).
- Profile bazli kapasite dogrulamasi icin script eklendi (`scripts/check-profile-capacity.sh`).
- PR CI'a `profile-capacity-matrix` job'i eklendi (`minimal`, `standard`, `secure`).
- Load shedding alarm esikleri icin kanonik policy dosyasi eklendi (`docs/load-shedding-policy.json`).
- Load shedding operasyon akisi icin runbook eklendi (`docs/load-shedding-runbook.md`).
- Starter template icin optional load shedding middleware/snippet eklendi (`starter/clean-service-template/src/interfaces/http/load_shedding.go.example`).
- Starter template icin optional load shedding smoke scripti eklendi (`starter/clean-service-template/tests/integration/load-shed-smoke.sh`).
- Advanced perf icin spike ve soak senaryolari eklendi (`scripts/perf/k6-spike.js`, `scripts/perf/k6-soak.js`).
- Nightly extended perf workflow'u ve ortak scenario wrapper eklendi (`.github/workflows/perf-extended-nightly.yml`, `scripts/run-k6-scenario.sh`).
- Extended perf artifact parser'i eklendi (`scripts/parse-k6-summary.py`).
- Nightly extended perf kalite raporu eklendi (`perf-extended-quality-report`).
- Autoscaling/capacity karar cercevesi icin playbook eklendi (`docs/autoscaling-capacity-playbook.md`).
- Starter template icin security compose override eklendi (`starter/clean-service-template/docker-compose.security.yml`).
- Starter template icin process-mode local runner eklendi (`starter/clean-service-template/scripts/run-local.sh`).
- Starter template icin process-mode smoke wrapper ve CI snippet eklendi (`tests/integration/process-mode-smoke.sh`, `ci-github-actions-process-mode-snippet.md`).
- Nightly workflow governance matrisi eklendi (`docs/nightly-workflow-governance-matrix.md`).
- Branch protection / required checks mapping dokumani eklendi (`docs/branch-protection-required-checks.md`).
- Delivery workflow governance matrisi eklendi (`docs/delivery-workflow-governance-matrix.md`).
- `check-release-policy-drift.sh` checklist coverage capraz dogrulamasi yapacak sekilde genisletildi.
- Starter template icin varyant secim rehberi eklendi (`starter/clean-service-template/docs/variant-selection-guide.md`).
- Kalite raporu governance coverage sinyali uretecek sekilde genisletildi (`scripts/generate-quality-report.py`).

### Fixed
- API versioning sonrasi frontend tarafindaki legacy endpoint tutarsizligi giderildi.
- `POST /health/ready/invalidate` endpoint'i admin JWT gerektirecek sekilde korunmaya alindi.
- API gateway async logger sender, logger tarafindan donen non-2xx cevaplari artik basarisiz sayar (retry/hata akisi).

### Security
- Planlama iterasyonunda legacy policy ve internal endpoint koruma maddeleri onceliklendirildi.
- Ready cache invalidation endpoint'i public erisim yuzeyinden cikarildi.

## [0.1.0] - 2026-02-26

### Added
- Initial AppFoundryLab baseline with polyglot microservices architecture.
