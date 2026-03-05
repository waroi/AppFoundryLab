# AppFoundryLab

AppFoundryLab is a production-shaped polyglot full-stack starter.

It combines application code and operational workflows in one repository so teams can:
- run everything locally,
- rehearse deployment and recovery flows,
- evolve the system with clear architecture and operations boundaries.

## Stack Summary

- Frontend: Astro + Svelte (`frontend/`)
- API Gateway: Go (`backend/services/api-gateway`)
- Worker: Rust gRPC (`backend/core/calculator`)
- Logger Service: Go (`backend/services/logger`)
- Data: PostgreSQL, Redis, MongoDB
- Orchestration: Docker Compose
- CI/CD and Operations: GitHub Actions + `scripts/`

## Quick Start (Local)

Prerequisites:
- Docker / Docker Desktop
- Bash shell

Run:

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

Default local URLs:
- Frontend: `http://127.0.0.1:4321/`
- Frontend test page: `http://127.0.0.1:4321/test`
- API Gateway: `http://127.0.0.1:8080`
- Logger metrics: `http://127.0.0.1:8090/metrics`

Stop services:

```bash
./scripts/dev-down.sh standard
```

If ports are already in use, update `.env.docker.local` before `dev-up`.

## Documentation Map

English:
- [Quick Start](docs/en/quick-start.md)
- [Developer Guide](docs/en/developer-guide.md)
- [Architecture](docs/en/architecture.md)
- [Operations](docs/en/operations.md)
- [Deployment](docs/en/deployment.md)
- [Incident Response](docs/en/incident-response.md)
- [Testing and Quality](docs/en/testing-and-quality.md)
- [Project Analysis](docs/en/project-analysis.md)

Turkish:
- [Hizli Baslangic](docs/tr/hizli-baslangic.md)
- [Gelistirme Rehberi](docs/tr/gelistirme-rehberi.md)
- [Mimari](docs/tr/mimari.md)
- [Operasyonlar](docs/tr/operasyonlar.md)
- [Deployment](docs/tr/deployment.md)
- [Incident Yonetimi](docs/tr/incident-yonetimi.md)
- [Test ve Kalite](docs/tr/test-ve-kalite.md)
- [Proje Analizi](docs/tr/proje-analizi.md)

Core runbooks and governance:
- [Teknik Analiz](docs/appfoundrylab-teknik-analiz.md)
- [Deployment Strategy](docs/deployment-strategy.md)
- [Runtime Incident Response](docs/runtime-incident-response.md)
- [Operator Observability Runbook](docs/operator-observability-runbook.md)
- [Release Policy](docs/release-policy.md)
- [Gelistirme Plani](docs/gelistirmePlanı.md)

## Common Workflows

Quality and checks:

```bash
./scripts/quality-gate.sh sandbox-safe
./scripts/quality-gate.sh host-strict
./scripts/check-doc-drift.sh
```

Frontend browser tests:

```bash
cd frontend
bun run e2e:bootstrap
bun run e2e
```

Single-host deployment and lifecycle:
- Deploy and operations: [`scripts/deploy-single-host.sh`](scripts/deploy-single-host.sh)
- Rollback: [`scripts/rollback-single-host.sh`](scripts/rollback-single-host.sh)
- Backup and restore drill: [`scripts/backup-single-host.sh`](scripts/backup-single-host.sh), [`scripts/restore-drill-single-host.sh`](scripts/restore-drill-single-host.sh)
- Release catalog and evidence: [`scripts/release-catalog.sh`](scripts/release-catalog.sh), [`scripts/collect-release-evidence.sh`](scripts/collect-release-evidence.sh)

Related workflows:
- [CI Pipeline](.github/workflows/appfoundrylab-ci.yml)
- [Staging Deploy](.github/workflows/deploy-single-host-staging.yml)
- [Production Deploy](.github/workflows/deploy-single-host-production.yml)
- [Release Evidence Harvest](.github/workflows/release-evidence-harvest.yml)

## Repository Layout

```text
backend/        Go services, Rust worker, proto, infra
frontend/       Astro + Svelte application
scripts/        Local ops, CI helpers, release/backup tooling
deploy/         Compose overlays, caddy and observability configs
docs/           EN/TR documentation and runbooks
multi_agent/    Orchestration rules and prompts
```

## Notes

- `.env` is ignored by git; use examples such as `.env.example` and `.env.docker.local`.
- `.toolchain/` is local-only and intentionally ignored.
- On WSL, if Linux `docker` is not available, set `DOCKER_BIN` to Docker Desktop's `docker.exe` path before running scripts.
