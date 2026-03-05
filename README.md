# AppFoundryLab

AppFoundryLab, production-shaped bir polyglot full-stack starter'dir.
Amaci yeni bir gelistiricinin stack'i yerelde kaldirmasi, tarayicidan dogrulamasi ve extension noktalarini butun operator runbook'larini okumadan gorebilmesidir.

## Stack

- Frontend: Astro + Svelte (`frontend/`)
- API Gateway: Go (`backend/services/api-gateway`)
- Logger Service: Go (`backend/services/logger`)
- Worker: Rust gRPC (`backend/core/calculator`)
- Data: PostgreSQL, Redis, MongoDB
- Orchestration: Docker Compose
- Ops ve governance: `scripts/`, GitHub Actions, `docs/`

## Quick Start

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

Varsayilan yerel adresler:
- Frontend: `http://127.0.0.1:4321/`
- Frontend health: `http://127.0.0.1:4321/healthz`
- API live: `http://127.0.0.1:8080/health/live`
- API ready: `http://127.0.0.1:8080/health/ready`
- Logger metrics: `http://127.0.0.1:8090/metrics`

`dev-doctor` WSL icinde `docker compose unavailable` diyorsa Docker Desktop WSL integration'i acin veya `DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"` ile tekrar deneyin.

## First Browser Smoke

1. `http://127.0.0.1:4321/` adresini acin.
2. `admin` kullanicisini ve `./scripts/bootstrap.sh` cikisindaki veya `.env.docker.local` icindeki `BOOTSTRAP_ADMIN_PASSWORD` degerini kullanin.
3. Giristen sonra `runtime-metrics-summary`, `trace-lookup-panel` ve request log satirlarinin yüklendigini dogrulayin.

Gercek stack browser smoke:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e:live
```

Mock-backed hizli UI regresyonu:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e
```

## Validation Modes

- Local bring-up truth: `./scripts/dev-up.sh standard`
- Local teardown with reset: `./scripts/dev-down.sh standard --volumes`
- Mock-backed UI regression: `cd frontend && ../.toolchain/bun/bin/bun run e2e`
- Real stack browser smoke: `cd frontend && ../.toolchain/bun/bin/bun run e2e:live`
- Script + policy gate: `./scripts/quality-gate.sh sandbox-safe`
- Deeper repo gate: `./scripts/quality-gate.sh ci-full`

## Documentation Map

English:
- [Quick Start](/mnt/d/w/AppFoundryLab/docs/en/quick-start.md)
- [Developer Guide](/mnt/d/w/AppFoundryLab/docs/en/developer-guide.md)
- [Architecture](/mnt/d/w/AppFoundryLab/docs/en/architecture.md)
- [Operations](/mnt/d/w/AppFoundryLab/docs/en/operations.md)
- [Deployment](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)
- [Incident Response](/mnt/d/w/AppFoundryLab/docs/en/incident-response.md)
- [Testing and Quality](/mnt/d/w/AppFoundryLab/docs/en/testing-and-quality.md)
- [Project Analysis](/mnt/d/w/AppFoundryLab/docs/en/project-analysis.md)

Turkish:
- [Hizli Baslangic](/mnt/d/w/AppFoundryLab/docs/tr/hizli-baslangic.md)
- [Gelistirme Rehberi](/mnt/d/w/AppFoundryLab/docs/tr/gelistirme-rehberi.md)
- [Mimari](/mnt/d/w/AppFoundryLab/docs/tr/mimari.md)
- [Operasyonlar](/mnt/d/w/AppFoundryLab/docs/tr/operasyonlar.md)
- [Deployment](/mnt/d/w/AppFoundryLab/docs/tr/deployment.md)
- [Incident Yonetimi](/mnt/d/w/AppFoundryLab/docs/tr/incident-yonetimi.md)
- [Test ve Kalite](/mnt/d/w/AppFoundryLab/docs/tr/test-ve-kalite.md)
- [Proje Analizi](/mnt/d/w/AppFoundryLab/docs/tr/proje-analizi.md)

Core docs:
- [Teknik Analiz](/mnt/d/w/AppFoundryLab/docs/appfoundrylab-teknik-analiz.md)
- [Gelistirme Plani](/mnt/d/w/AppFoundryLab/docs/gelistirmePlanı.md)
- [Progress](/mnt/d/w/AppFoundryLab/PROGRESS.md)
- [Scripts Index](/mnt/d/w/AppFoundryLab/scripts/README.md)

## Notes

- `PROGRESS.md` repo backlog'unun tek kanonik kaynagidir.
- `docs/gelistirmePlanı.md` stratejik faz sirasini tutar; canli backlog tutmaz.
- Advanced ops yuzeyi (release evidence, attestation, observability overlays, single-host deploy) starter'in ustune gelen opsiyonel katmandir.
- `dev-up` artik yalnizca process liveness degil; readiness, logger erisimi ve bir authenticated admin smoke ile basari raporlar.
