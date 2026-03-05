# AppFoundryLab — Comprehensive Wiki

> **Canonical reference for the AppFoundryLab repository.** This page consolidates all documentation into a single navigable source of truth. Individual topic pages provide additional depth; links to them are included throughout.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture](#2-architecture)
3. [Technology Stack](#3-technology-stack)
4. [Repository Structure](#4-repository-structure)
5. [Quick Start](#5-quick-start)
6. [Configuration Reference](#6-configuration-reference)
7. [API Reference](#7-api-reference)
8. [Developer Guide](#8-developer-guide)
9. [Testing & Quality Gates](#9-testing--quality-gates)
10. [Deployment](#10-deployment)
11. [Operations & Observability](#11-operations--observability)
12. [Incident Management](#12-incident-management)
13. [CI/CD Workflows](#13-cicd-workflows)
14. [Runtime Profiles](#14-runtime-profiles)
15. [Multi-Agent Orchestration](#15-multi-agent-orchestration)
16. [Troubleshooting](#16-troubleshooting)
17. [Glossary](#17-glossary)
18. [Documentation Map](#18-documentation-map)

---

## 1. Project Overview

**AppFoundryLab** is a production-shaped, polyglot full-stack starter template designed to help teams:

- **Run the entire application locally** with Docker Compose in minutes
- **Rehearse deployment and recovery flows** end-to-end before going live
- **Evolve systems** with clear architectural and operational boundaries
- **Learn DevOps patterns** by working with a reference-grade codebase

It is not a toy — it is a deliberately opinionated reference implementation that combines application code and operational workflows in a single monorepo. Every layer is real: real auth, real metrics, real incident events, real backup/restore drills, and real performance tests.

### Goals

| Goal | How it is achieved |
|------|-------------------|
| Local production parity | `docker-compose.single-host.yml` mirrors the VPS layout |
| Operational confidence | Backup, restore, rollback, and release-evidence scripts |
| Observability from day one | Prometheus metrics, Grafana overlays, incident journal |
| Polyglot best-of-breed | Go gateway + Rust gRPC worker + Astro/Svelte frontend |
| Quality assurance | Unit, integration, e2e, performance, doc-drift checks |
| Internationalization | EN/TR locales with URL-authoritative routing |

---

## 2. Architecture

### 2.1 High-Level Component Map

```mermaid
flowchart LR
    Browser[Browser]
    Frontend[Astro + Svelte\nFrontend :4321]
    Gateway[Go API Gateway\n:8080]
    Monitor[Incident Monitor\n(embedded in Gateway)]
    Worker[Rust gRPC Worker\n:7070]
    Logger[Go Logger Service\n:8090]
    Postgres[(PostgreSQL\n:5432)]
    Redis[(Redis\n:6379)]
    Mongo[(MongoDB\n:27017)]

    Browser --> Frontend
    Frontend --> Gateway
    Gateway --> Worker
    Gateway --> Logger
    Gateway --> Postgres
    Gateway --> Redis
    Gateway --> Monitor
    Monitor --> Logger
    Logger --> Mongo
```

### 2.2 Request Flow

```
Browser
  │  HTTP GET /
  ▼
Astro Frontend (SSR)
  │  fetch /api/v1/*
  ▼
Go API Gateway
  ├── JWT verification
  ├── Rate limit check (Redis or in-memory)
  ├── RBAC check
  ├── Route dispatch
  │     ├── gRPC → Rust Worker (compute)
  │     ├── HTTP → Go Logger (log ingest)
  │     └── Direct DB → PostgreSQL (auth)
  └── Async enqueue → Logger queue
```

### 2.3 Incident Flow

```
Gateway collects metrics every INCIDENT_EVENT_INTERVAL_MS
  │
  ▼
Incident Monitor evaluates alert thresholds
  │
  ├── Alert opened / updated / resolved?
  │       │
  │       ▼
  │   Emit incident event → Logger (sink: logger | stdout | webhook)
  │       │
  │       ▼
  │   Logger stores event in MongoDB (with deduplication)
  │
  └── No change → wait for next tick

Frontend polls /api/v1/admin/incident-events
  └── Renders incident history in the admin diagnostics panel
```

### 2.4 Layers

| Layer | Component | Responsibility |
|-------|-----------|----------------|
| Presentation | Astro + Svelte | SSR pages, locale routing, interactive dashboards |
| API | Go API Gateway | Auth, RBAC, rate limiting, routing, load shedding, metrics |
| Compute | Rust gRPC Worker | Fibonacci, hash computation via mTLS gRPC |
| Logging | Go Logger Service | Async request log ingest, incident event journal |
| Data | PostgreSQL | Auth tokens, user accounts, schema migrations |
| Data | MongoDB | Request logs, incident events, audit trail |
| Data | Redis | Rate limit state, distributed cache |
| Proxy | Caddy | TLS termination, routing, operator mTLS proxy |
| Observability | Prometheus + Grafana | Metrics collection, dashboards, alerting |

### 2.5 Network Isolation (Production)

In `docker-compose.single-host.yml`, services communicate over private Docker networks. No service ports are exposed to the host except through the Caddy reverse proxy. This means:

- `frontend` and `api-gateway` are reachable only via Caddy
- `postgres`, `redis`, `mongo` are not accessible from outside the Docker network
- `calculator` gRPC is only reachable from `api-gateway` over the private network

---

## 3. Technology Stack

### Frontend

| Technology | Version | Purpose |
|-----------|---------|---------|
| Astro | 5.2.x | SSR-first static site framework |
| Svelte | 5.20.x | Reactive UI components (islands) |
| Tailwind CSS | 4.2.x | Utility-first styling |
| SCSS | — | Global styles, semantic tokens |
| Playwright | — | End-to-end browser tests |
| Biome | — | Linting and formatting |
| Bun | — | Package manager and runtime |

**Key patterns:**
- Locale-authoritative URL routing (`/` = EN, `/tr` = TR)
- `data-testid` / `data-*` selectors for stable e2e assertions
- Theme persistence via `localStorage` (light/dark)
- Circuit breaker + retry logic in `frontend/src/lib/api/`

### Backend — API Gateway

| Technology | Version | Purpose |
|-----------|---------|---------|
| Go | 1.24.x | Primary language |
| Chi v5 | — | Lightweight HTTP router |
| JWT v5 | — | Authentication token signing/verification |
| pgx/v5 | — | PostgreSQL driver |
| go-redis/v9 | — | Redis client |
| mongo-driver | 1.17.x | MongoDB client |
| golang-migrate | — | Database schema migrations |
| gRPC Go | 1.74.x | gRPC client for Worker |
| protobuf | 1.36.x | Protocol Buffer codec |

### Backend — Worker

| Technology | Version | Purpose |
|-----------|---------|---------|
| Rust | stable | Primary language |
| Tonic | — | gRPC server framework |
| Protocol Buffers 3 | — | RPC contract (`backend/proto/worker.proto`) |
| mTLS | — | Mutual TLS authentication |

### Data

| Technology | Version | Purpose |
|-----------|---------|---------|
| PostgreSQL | 16 | Relational data, migrations |
| Redis | 7 | Rate limiting, cache |
| MongoDB | 7 | Logs, incident events |

### Infrastructure & DevOps

| Technology | Purpose |
|-----------|---------|
| Docker + Docker Compose | Container orchestration |
| Caddy | Reverse proxy, TLS termination |
| Prometheus | Metrics collection |
| Grafana | Dashboards |
| GitHub Actions | CI/CD pipelines |
| k6 | Performance testing |
| SSH | Remote deployment |

---

## 4. Repository Structure

```
AppFoundryLab/
├── backend/                          # Go + Rust backend services
│   ├── services/
│   │   ├── api-gateway/              # Go HTTP gateway (auth, RBAC, routing)
│   │   │   ├── cmd/api-gateway/      # Entrypoint (main.go)
│   │   │   └── internal/
│   │   │       ├── handlers/         # HTTP handlers
│   │   │       ├── middleware/        # Auth, rate-limit, load-shed middleware
│   │   │       ├── runtimecfg/       # Runtime configuration snapshot
│   │   │       └── incidents/        # Incident monitor logic
│   │   └── logger/                   # Go Logger service
│   │       ├── cmd/logger/           # Entrypoint (main.go)
│   │       └── internal/             # Queue, persistence, metrics
│   ├── core/
│   │   └── calculator/               # Rust gRPC worker
│   │       └── src/main.rs           # gRPC server + Fibonacci/Hash logic
│   ├── proto/
│   │   └── worker.proto              # gRPC service contract
│   ├── pkg/                          # Shared Go packages
│   ├── infrastructure/
│   │   ├── certs/                    # Dev TLS certificates
│   │   └── postgres/                 # SQL migration files
│   ├── go.mod / go.sum               # Go module definition
│   └── Dockerfile.*                  # Service Docker images
│
├── frontend/                         # Astro + Svelte UI
│   ├── src/
│   │   ├── components/
│   │   │   ├── Interactive/          # Svelte islands (diagnostics, admin panels)
│   │   │   ├── Layout/               # Shared layout controls
│   │   │   ├── Page/                 # Page shell components (EN + TR)
│   │   │   └── Static/               # Locale-reactive static elements
│   │   ├── pages/
│   │   │   ├── index.astro           # EN home (/)
│   │   │   ├── test.astro            # EN test (/test)
│   │   │   └── [locale]/             # TR route variants (/tr, /tr/test)
│   │   ├── layouts/
│   │   │   └── BaseLayout.astro      # Document shell, theme bootstrap
│   │   ├── lib/
│   │   │   ├── api/                  # HTTP client, circuit breaker, retry
│   │   │   └── ui/
│   │   │       ├── preferences.ts    # Locale/theme store
│   │   │       ├── routes.ts         # Locale-aware URL map
│   │   │       ├── copy.ts           # EN/TR copy dictionary
│   │   │       └── formatters.ts     # Date, percent, duration formatters
│   │   └── styles/
│   │       └── global.scss           # Semantic tokens, dark/light themes
│   ├── e2e/                          # Playwright test suites
│   ├── scripts/                      # Smoke test, Playwright bootstrap
│   ├── astro.config.mjs              # Astro configuration
│   ├── tailwind.config.cjs           # Tailwind CSS configuration
│   ├── package.json                  # Bun dependencies
│   └── Dockerfile / Dockerfile.prod  # Container images
│
├── scripts/                          # Operational automation
│   ├── dev-up.sh / dev-down.sh       # Local stack orchestration
│   ├── dev-doctor.sh                 # Toolchain validation
│   ├── bootstrap.sh                  # Environment initialization
│   ├── deploy-single-host.sh         # VPS deployment
│   ├── rollback-single-host.sh       # Rollback to previous release
│   ├── backup-single-host.sh         # Database backup
│   ├── restore-drill-single-host.sh  # Restore drill execution
│   ├── release-gate.sh               # Release checklist automation
│   ├── quality-gate.sh               # Test & quality validation
│   ├── release-catalog.sh            # Release catalog management
│   ├── collect-release-evidence.sh   # Evidence collection
│   ├── run-k6-smoke.sh               # Performance smoke test
│   └── run-k6-scenario.sh            # Performance scenario runner
│
├── deploy/                           # Production compose overlays
│   ├── docker-compose.observability.yml
│   ├── docker-compose.observability.operator.yml
│   ├── docker-compose.observability.operator.mtls.yml
│   ├── docker-compose.single-host.ghcr.yml
│   ├── caddy/                        # Reverse proxy configs
│   ├── observability/                # Prometheus config, alert rules
│   └── backups/                      # S3 lifecycle policies
│
├── docs/                             # Documentation
│   ├── en/                           # English documentation
│   ├── tr/                           # Turkish documentation
│   ├── adr/                          # Architecture Decision Records
│   └── *.md                          # Runbooks, governance docs
│
├── multi_agent/                      # Multi-agent orchestration
│   ├── agents/                       # Agent definitions
│   ├── prompts/                      # LLM prompts
│   ├── roles/                        # Role definitions
│   └── instructions/                 # Orchestration rules
│
├── presets/                          # Runtime profile configurations
│   ├── minimal.env
│   ├── standard.env
│   └── secure.env
│
├── skills/                           # Reusable skill modules
│   └── multi-agent-orchestrator/
│
├── starter/                          # New service template
│   └── clean-service-template/
│
├── docker-compose.yml                # Local development stack
├── docker-compose.single-host.yml    # Production-like single-host stack
├── docker-compose.security.yml       # Security overlays
├── .env.example                      # Development environment template
├── .env.single-host.example          # VPS environment template
├── toolchain.versions.json           # Required tool versions
├── CHANGELOG.md                      # Version history
└── README.md                         # Project overview
```

---

## 5. Quick Start

### 5.1 Prerequisites

| Tool | Minimum Version | Check |
|------|----------------|-------|
| Docker Engine | 24.x | `docker --version` |
| Docker Compose plugin | 2.x | `docker compose version` |
| Bash | 5.x | `bash --version` |

> **Tip:** Run `./scripts/dev-doctor.sh` to validate all prerequisites automatically.

### 5.2 First Run

```bash
# Step 1 — Validate toolchain
./scripts/dev-doctor.sh

# Step 2 — Bootstrap environment (creates .env.docker.local, runs DB migrations)
./scripts/bootstrap.sh standard --force

# Step 3 — Start all services
./scripts/dev-up.sh standard
```

### 5.3 Default Local URLs

| Service | URL |
|---------|-----|
| Frontend (EN) | `http://127.0.0.1:4321/` |
| Frontend (TR) | `http://127.0.0.1:4321/tr` |
| Frontend test page | `http://127.0.0.1:4321/test` |
| API Gateway | `http://127.0.0.1:8080` |
| Logger metrics | `http://127.0.0.1:8090/metrics` |

### 5.4 Stopping Services

```bash
./scripts/dev-down.sh standard
```

### 5.5 Port Conflicts

If default ports are occupied, override them before starting:

```bash
FRONTEND_HOST_PORT=14321 \
API_GATEWAY_HOST_PORT=18080 \
LOGGER_HOST_PORT=18090 \
./scripts/dev-up.sh standard
```

Or edit `.env.docker.local` persistently.

### 5.6 What to Explore After Starting

1. **Home page** — System status dashboard, uptime, request rate
2. **Test page** — `/test` — API interaction panel
3. **Locale switching** — Toggle EN/TR in the top-right toolbar
4. **Theme switching** — Toggle light/dark in the top-right toolbar
5. **Admin login** — Use bootstrap credentials (`BOOTSTRAP_ADMIN_USER` / `BOOTSTRAP_ADMIN_PASSWORD`)
6. **Admin diagnostics** — Runtime config, metrics, incident report

---

## 6. Configuration Reference

### 6.1 Environment Files

| File | Purpose |
|------|---------|
| `.env.example` | Development defaults template |
| `.env.docker` | Docker Compose environment defaults |
| `.env.docker.local` | Local overrides (generated, git-ignored) |
| `.env.single-host.example` | VPS/production environment template |
| `presets/minimal.env` | Minimal feature profile |
| `presets/standard.env` | Standard feature profile |
| `presets/secure.env` | Security-hardened feature profile |

### 6.2 Key Environment Variables

#### Service Ports

```env
FRONTEND_HOST_PORT=4321
API_GATEWAY_HOST_PORT=8080
LOGGER_HOST_PORT=8090
```

#### Database Connections

```env
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=appfoundry
POSTGRES_USER=appfoundry
POSTGRES_PASSWORD=<secret>

REDIS_HOST=redis
REDIS_PORT=6379

MONGO_HOST=mongo
MONGO_PORT=27017
MONGO_DB=appfoundry
MONGO_USER=appfoundry
MONGO_PASSWORD=<secret>
```

#### Authentication

```env
JWT_SECRET=<strong-secret>
JWT_ISSUER=appfoundrylab
JWT_TTL_SECONDS=3600
BOOTSTRAP_ADMIN_USER=admin
BOOTSTRAP_ADMIN_PASSWORD=<strong-password>
LOCAL_AUTH_MODE=demo        # use 'demo' for local only
```

#### Rate Limiting

```env
RATE_LIMIT_STORE=memory          # or 'redis'
AUTH_RATE_LIMIT_PER_MINUTE=30
API_RATE_LIMIT_PER_MINUTE=120
RATE_LIMIT_REDIS_FAILURE_MODE=open   # 'open' or 'closed'
```

#### Load Shedding

```env
MAX_INFLIGHT_REQUESTS=256
LOAD_SHED_EXEMPT_PREFIXES=/health,/metrics
```

#### Logger Service

```env
LOGGER_QUEUE_SIZE=2048
LOGGER_WORKERS=4
LOGGER_RETRY_MAX=1
LOGGER_RETRY_BACKOFF_BASE_MS=50
LOGGER_DROP_ALERT_THRESHOLD_PCT=10
```

#### Incident Monitoring

```env
INCIDENT_EVENT_SINK=logger              # logger | stdout | webhook | logger,webhook
INCIDENT_EVENT_INTERVAL_MS=10000
INCIDENT_EVENT_WEBHOOK_URL=             # optional
INCIDENT_EVENT_WEBHOOK_HMAC_SECRET=     # optional, for signed webhooks
INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS=   # allowlist
```

#### gRPC Worker

```env
WORKER_GRPC_ADDRESS=calculator:7070
WORKER_GRPC_MTLS_ENABLED=false          # true for mTLS
WORKER_GRPC_CA_CERT_FILE=
WORKER_GRPC_CLIENT_CERT_FILE=
WORKER_GRPC_CLIENT_KEY_FILE=
```

#### Frontend API Client

```env
PUBLIC_API_BASE_URL=http://localhost:8080
PUBLIC_API_RETRY_MAX_ATTEMPTS=2
PUBLIC_API_RETRY_BASE_DELAY_MS=200
PUBLIC_API_CIRCUIT_FAILURE_THRESHOLD=5
PUBLIC_API_CIRCUIT_COOLDOWN_MS=15000
```

#### Runtime Profile

```env
RUNTIME_PROFILE=standard         # minimal | standard | secure
```

#### Observability

```env
ENABLE_OBSERVABILITY_STACK=false
ENABLE_OPERATOR_PROMETHEUS_ACCESS=false
PROMETHEUS_OPERATOR_ACCESS_MODE=basic-auth    # basic-auth | mtls
```

#### Backup & Restore

```env
BACKUP_SYNC_TARGET=local          # local | ssh | s3
BACKUP_ENCRYPTION_PASSPHRASE=     # optional
BACKUP_RETENTION_DAYS=30
BACKUP_AWS_REGION=
BACKUP_AWS_ENDPOINT_URL=
```

---

## 7. API Reference

### 7.1 Health & Status

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health/live` | None | Liveness check |
| GET | `/health/ready` | None | Readiness check + diagnostics summary |
| POST | `/health/ready/invalidate` | Admin | Invalidate readiness cache |

**Response: `/health/ready`**
```json
{
  "status": "ok",
  "services": {
    "postgres": "ok",
    "redis": "ok",
    "mongo": "ok",
    "worker": "ok",
    "logger": "ok"
  }
}
```

### 7.2 Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/login` | None | Username + password → JWT |
| POST | `/auth/refresh` | JWT | Refresh access token |

**Request: `/auth/login`**
```json
{
  "username": "admin",
  "password": "strong_password"
}
```

**Response:**
```json
{
  "token": "eyJ...",
  "expires_at": "2024-01-01T12:00:00Z"
}
```

### 7.3 Compute Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/compute/fibonacci` | JWT | Compute Fibonacci(n) |
| POST | `/api/v1/compute/hash` | JWT | Compute hash of input |

**Request: Fibonacci**
```json
{ "n": 10 }
```
**Response:**
```json
{ "value": 55, "computed_at": "2024-01-01T12:00:00Z" }
```

**Request: Hash**
```json
{ "input": "hello world", "algorithm": "sha256" }
```
**Response:**
```json
{ "hash": "b94d27b9...", "algorithm": "sha256" }
```

### 7.4 Admin Endpoints

All admin endpoints require `Authorization: Bearer <jwt>` with admin role.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/runtime-config` | Runtime config snapshot |
| GET | `/api/v1/admin/runtime-metrics` | Prometheus-format metrics |
| GET | `/api/v1/admin/runtime-report` | Full diagnostics report |
| GET | `/api/v1/admin/runtime-incident-report` | Incident summary with recommended actions |
| GET | `/api/v1/admin/incident-events` | Recent incident event history |
| POST | `/api/v1/admin/incident-events/clear` | Purge old incident events |
| GET | `/api/v1/admin/request-logs` | Paginated request log query |
| GET | `/api/v1/logs/trace/{traceId}` | Single request trace lookup |

### 7.5 gRPC Interface (Internal)

Defined in `backend/proto/worker.proto`:

```protobuf
service WorkerService {
  rpc Health(HealthRequest) returns (HealthResponse);
  rpc ComputeFibonacci(ComputeFibonacciRequest) returns (ComputeFibonacciResponse);
  rpc ComputeHash(ComputeHashRequest) returns (ComputeHashResponse);
}
```

> The gRPC interface is internal — only the API Gateway calls the Worker. mTLS is supported via `WORKER_GRPC_MTLS_ENABLED=true`.

### 7.6 Webhooks (Optional)

When `INCIDENT_EVENT_SINK` includes `webhook`:
- **Method**: POST to `INCIDENT_EVENT_WEBHOOK_URL`
- **Signature**: HMAC-SHA256 over body, sent in `X-Hub-Signature-256` header
- **Payload**: Incident event JSON

---

## 8. Developer Guide

### 8.1 Recommended Local Development Loop

```bash
# Initial setup (once)
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard

# During development
# Make changes to code files
# Services auto-reload (frontend Astro dev server, Go with air if configured)

# Run targeted tests
cd backend && go test ./services/api-gateway/...
cd frontend && bun run lint

# Before committing
./scripts/quality-gate.sh host-strict

# Stop services
./scripts/dev-down.sh standard
```

### 8.2 Adding a New Frontend Feature

1. **Add copy strings** to both locales in `frontend/src/lib/ui/copy.ts`
2. **Add routes** in `frontend/src/lib/ui/routes.ts` if needed
3. **Create components** in the appropriate `frontend/src/components/` subdirectory
4. **Use semantic tokens** from `global.scss` for styling (avoid one-off utility colors)
5. **Add `data-testid` hooks** for any testable UI state
6. **Add e2e test** in `frontend/e2e/`
7. **Update docs** in the same changeset

### 8.3 Adding a New Backend Endpoint

1. **Add handler** in `backend/services/api-gateway/internal/handlers/`
2. **Register route** in the router (typically in `cmd/api-gateway/main.go`)
3. **Add middleware** as needed (auth, rate limit, etc.)
4. **Add tests** covering positive, negative, and auth failure cases
5. **Update docs** and API reference
6. **Update runtime config** snapshot if a new env var is introduced

### 8.4 Adding a New Environment Variable

1. Add to `.env.example` with a descriptive comment
2. Add to `.env.single-host.example`
3. Add to the relevant `presets/*.env` files
4. Document in the [Configuration Reference](#6-configuration-reference)
5. Handle gracefully when missing (fail-safe defaults or explicit fail-fast)

### 8.5 Frontend Architecture Rules

- **Locale routing is URL-authoritative**: `/` and `/test` are EN; `/tr` and `/tr/test` are TR
- **Language switching uses full-page navigation** to the localized route variant
- **Theme persists in `localStorage`** — locale does not
- **All new visible strings** must be added to both locales in `copy.ts`
- **Stable selectors** (`data-testid`, `data-role`, `data-mode`, `data-status`) over translated text
- **Dark palette**: charcoal-led; **CTA accent**: vivid orange

### 8.6 Backend Architecture Rules

- **Rate limiting**: in-memory (`RATE_LIMIT_STORE=memory`) or Redis-backed
- **Load shedding**: `MAX_INFLIGHT_REQUESTS` is a hard cap; exempt paths in `LOAD_SHED_EXEMPT_PREFIXES`
- **Auth**: all `/api/v1/*` endpoints require JWT; admin endpoints require admin RBAC role
- **Logging**: async queue — never block request path on log write
- **Metrics**: increment Prometheus counters for all observable events

### 8.7 Starter Template

For new Go services, use the clean service template:

```bash
cp -r starter/clean-service-template/ my-new-service/
```

The template includes:
- Go service skeleton
- Integration smoke tests
- Minimal and security Docker Compose variants
- Local run scripts
- Variant selection guide

---

## 9. Testing & Quality Gates

### 9.1 Backend Tests

```bash
# All Go tests
cd backend && go test ./...

# Focused integration test
cd backend && go test ./services/api-gateway/cmd/api-gateway \
  -run TestIntegrationAuthProtectedWorkerLoggerMetrics -v

# Worker tests (Rust)
cd backend/core/calculator && cargo test

# Or use the repo helper (handles system cc issues)
./scripts/run-worker-tests.sh
```

### 9.2 Frontend Tests

```bash
# Linting
cd frontend && bun run lint

# Type checking
cd frontend && ./node_modules/.bin/astro check

# Build validation
cd frontend && ./node_modules/.bin/astro build

# Smoke test (SSR marker checks)
cd frontend && node ./scripts/smoke.mjs

# Smoke with real API
cd frontend && SMOKE_API_BASE_URL=http://127.0.0.1:8080 node ./scripts/smoke.mjs

# Playwright e2e (bootstrap first)
cd frontend && bun run e2e:bootstrap
cd frontend && ./scripts/run-playwright.sh
```

### 9.3 Quality Gate Commands

| Command | When to Use |
|---------|-------------|
| `./scripts/quality-gate.sh sandbox-safe` | Permission-limited sandbox (CI, containers) |
| `./scripts/quality-gate.sh host-strict` | Before opening a PR (developer machine) |
| `./scripts/quality-gate.sh ci-fast` | Standard CI pipeline |
| `./scripts/quality-gate.sh ci-full` | Nightly full coverage |

### 9.4 Governance Checks

```bash
./scripts/check-doc-drift.sh --mode strict   # Docs consistency
./scripts/check-release-policy-drift.sh       # Release policy adherence
./scripts/check-toolchain.sh                  # Tool version validation
./scripts/release-gate.sh fast                # Fast release checklist
./scripts/release-gate.sh full                # Full nightly gate
./scripts/local-ci-smoke.sh                   # Local CI chain
./scripts/test-dev-scripts.sh                 # Dev script validation
```

### 9.5 Performance Tests

```bash
# Smoke (baseline)
./scripts/run-k6-smoke.sh

# Spike scenario
./scripts/run-k6-scenario.sh spike

# Soak scenario
./scripts/run-k6-scenario.sh soak
```

Performance scenarios are in `scripts/perf/` (k6 scripts for smoke, spike, soak).

### 9.6 What "Green" Looks Like

You are in a safe state when:
- Local targeted tests pass
- `./scripts/quality-gate.sh host-strict` passes
- Documentation is updated alongside code changes
- New env vars and endpoints are documented
- CI reports green on the PR

### 9.7 Test Writing Rules

- Add positive **and** negative cases
- Test authorization failures explicitly
- Test contract shape changes (API breaking changes)
- Test operational edge cases driven by env vars
- For frontend: test locale switching, theme switching, reload persistence
- Assert `html[lang]` and `html[data-theme]` for locale/theme tests
- Prefer `data-testid` / `data-*` selectors over translated text

---

## 10. Deployment

### 10.1 Local Single-Host Package (First)

Before deploying to a real VPS, validate the full stack locally:

```bash
cp .env.single-host.example .env.single-host
# Edit .env.single-host — replace all placeholder secrets

./scripts/deploy-single-host.sh up ./.env.single-host
./scripts/backup-single-host.sh ./.env.single-host
./scripts/restore-drill-single-host.sh ./.env.single-host
./scripts/rehearse-release-evidence-local.sh ./.env.single-host
```

### 10.2 VPS Deployment

1. Provision Ubuntu LTS with Docker Engine and Compose plugin
2. Clone the repository on the host
3. Copy `.env.single-host.example` → `.env.single-host`
4. Fill all secrets, backup targets, and optional GHCR image refs
5. Store pinned SSH `known_hosts` in GitHub environment secrets
6. Run via GitHub Actions workflow or directly:

```bash
# Build mode (build images from source on the host)
./scripts/deploy-single-host.sh up ./.env.single-host

# Image mode (pull pre-built GHCR images)
# Set API_GATEWAY_IMAGE, LOGGER_IMAGE, etc. in .env.single-host first
./scripts/deploy-single-host.sh up ./.env.single-host
```

### 10.3 Rollback

```bash
# Rollback to previous GHCR manifest
./scripts/rollback-single-host.sh ./artifacts/ghcr/release-manifest.env ./.env.single-host

# Rollback using release catalog
RELEASE_CATALOG_PATH=./artifacts/release-catalog/staging/catalog.json \
./scripts/rollback-single-host.sh previous ./.env.single-host
```

### 10.4 Reverse Proxy (Caddy)

Caddy configuration examples are in `deploy/caddy/`. For local development, ports are exposed directly. For production:
- Caddy terminates TLS
- Routes public traffic to `frontend` and `api-gateway`
- Optional mTLS proxy for Prometheus operator access

### 10.5 Observability Stack (Optional)

```bash
# Add Prometheus + Grafana to the stack
docker compose \
  -f docker-compose.single-host.yml \
  -f deploy/docker-compose.observability.yml \
  up -d

# With operator Prometheus access (basic-auth)
ENABLE_OPERATOR_PROMETHEUS_ACCESS=true \
PROMETHEUS_OPERATOR_ACCESS_MODE=basic-auth \
docker compose -f deploy/docker-compose.observability.operator.yml up -d

# With mTLS
PROMETHEUS_OPERATOR_ACCESS_MODE=mtls \
docker compose -f deploy/docker-compose.observability.operator.mtls.yml up -d
```

### 10.6 Release Evidence

```bash
# Collect evidence
./scripts/collect-release-evidence.sh staging \
  ./artifacts/release-catalog/staging/catalog.json \
  ./artifacts/release-ledgers/staging \
  ./artifacts/release-evidence/staging

# Attest ledger
./scripts/attest-release-ledger.sh \
  ./artifacts/release-ledgers/staging/ledger.json

# Export to long-term audit storage
./scripts/export-release-evidence.sh staging \
  ./artifacts/release-catalog/staging/catalog.json \
  ./artifacts/release-ledgers/staging \
  ./artifacts/release-evidence/staging \
  ./artifacts/release-audit
```

---

## 11. Operations & Observability

### 11.1 Health Checks

```bash
# Liveness
curl http://127.0.0.1:8080/health/live

# Readiness (full diagnostics)
curl http://127.0.0.1:8080/health/ready
```

### 11.2 Metrics

The API Gateway and Logger expose Prometheus-format metrics:

```bash
# Gateway metrics
curl http://127.0.0.1:8080/metrics

# Logger metrics
curl http://127.0.0.1:8090/metrics
```

Key metrics:
- `http_requests_total` — total request count by method, path, status
- `http_request_duration_seconds` — request latency histogram
- `logger_queue_depth` — current logger queue depth
- `logger_drop_total` — total log events dropped
- `incident_events_total` — total incident events emitted

### 11.3 Request Logs

```bash
# Query recent request logs (admin JWT required)
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/api/v1/admin/request-logs

# Trace a specific request
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/api/v1/logs/trace/{traceId}
```

### 11.4 Backup Operations

```bash
# Create backup bundle
./scripts/backup-single-host.sh ./.env.single-host

# Run restore drill (non-destructive)
./scripts/restore-drill-single-host.sh ./.env.single-host

# Prune old incident events
./scripts/prune-incident-events.sh ./.env.single-host
```

Backup bundles include:
- PostgreSQL dump
- MongoDB dump
- Checksum file
- Optional encryption (AES via `BACKUP_ENCRYPTION_PASSPHRASE`)
- `backup-catalog.json` and `latest-bundle.txt` for versioned tracking

### 11.5 S3 Lifecycle Drift Check

```bash
./scripts/check-s3-lifecycle-policy.sh \
  "$BUCKET_NAME" \
  ./deploy/backups/s3-lifecycle-policy.example.json
```

### 11.6 Operator mTLS

```bash
# Generate operator mTLS certificates
./scripts/generate-operator-mtls-certs.sh

# Check readiness
./scripts/check-operator-mtls-readiness.sh
```

---

## 12. Incident Management

### 12.1 Incident Flow Overview

1. Gateway collects request metrics and health data every `INCIDENT_EVENT_INTERVAL_MS` ms
2. Incident Monitor evaluates alert thresholds
3. When an alert opens, updates (after dedup window), or resolves → emit incident event
4. Events are stored in MongoDB via the Logger service
5. Frontend Admin panel polls `/api/v1/admin/incident-events` and shows history

### 12.2 Incident API

```bash
# Get current incident report (with recommended actions)
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/api/v1/admin/runtime-incident-report

# Get recent incident events
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/api/v1/admin/incident-events

# Clear old events
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/api/v1/admin/incident-events/clear
```

### 12.3 Incident Report Structure

```json
{
  "severity": "warning",
  "category": "latency",
  "title": "High P99 latency detected",
  "summary": "P99 response time exceeds threshold",
  "runbooks": ["latency-regression-runbook.md"],
  "next_actions": ["Check worker gRPC response times", "Review rate limiting config"],
  "evidence": {
    "health": { "status": "degraded" },
    "alerts": [{ "name": "high_latency", "state": "open" }],
    "logger_state": { "queue_depth": 512 }
  }
}
```

### 12.4 Alert Deduplication

- Default deduplication window: **5 minutes**
- Events for the same alert will not be re-emitted within the window
- On resolution, a separate `resolved` event is emitted immediately

### 12.5 Incident Sinks

| Sink | Config | Description |
|------|--------|-------------|
| `logger` | Default | Stored in MongoDB via Logger service |
| `stdout` | Log to console | Useful for debugging |
| `webhook` | `INCIDENT_EVENT_WEBHOOK_URL` | HMAC-signed POST to external URL |

Multiple sinks can be combined: `INCIDENT_EVENT_SINK=logger,webhook`

### 12.6 Related Runbooks

- [`docs/runtime-incident-response.md`](../runtime-incident-response.md)
- [`docs/latency-regression-runbook.md`](../latency-regression-runbook.md)
- [`docs/api-degradation-runbook.md`](../api-degradation-runbook.md)
- [`docs/dependency-degradation-runbook.md`](../dependency-degradation-runbook.md)
- [`docs/logger-pipeline-runbook.md`](../logger-pipeline-runbook.md)
- [`docs/load-shedding-runbook.md`](../load-shedding-runbook.md)

---

## 13. CI/CD Workflows

### 13.1 Workflow Overview

All workflows live in `.github/workflows/`.

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `appfoundrylab-ci.yml` | PR / merge | Main pipeline: tests, quality gates, perf benchmarks |
| `deploy-single-host-staging.yml` | Manual / merge | Deploy to staging VPS |
| `deploy-single-host-production.yml` | Manual | Deploy to production VPS |
| `single-host-ops.yml` | Manual | Remote operations (backup, prune, rollback) |
| `publish-ghcr-images.yml` | Tag / manual | Build and push images to GHCR |
| `release-evidence-harvest.yml` | Post-deploy | Collect and attest release evidence |
| `backup-lifecycle-drift.yml` | Schedule | Check S3 lifecycle drift |
| `restore-drill-single-host.yml` | Schedule | Automated restore drill |
| `release-gate-full-nightly.yml` | Nightly | Full release gate execution |
| `perf-extended-nightly.yml` | Nightly | Extended performance scenarios |

### 13.2 Main CI Pipeline (`appfoundrylab-ci.yml`)

1. **Backend tests** — Go unit + integration tests
2. **Worker tests** — Rust cargo tests
3. **Frontend checks** — Lint, type check, build, smoke
4. **Quality gate** — `quality-gate.sh ci-fast`
5. **Doc drift** — `check-doc-drift.sh`
6. **Release policy** — `check-release-policy-drift.sh`
7. **Performance benchmark** — k6 smoke test

### 13.3 Deployment Workflow Pattern

```
1. Publish GHCR images (publish-ghcr-images.yml)
        │
        ▼
2. Deploy to staging (deploy-single-host-staging.yml)
        │
        ▼
3. Run post-deploy checks (post-deploy-check.sh)
        │
        ▼
4. Collect release evidence (release-evidence-harvest.yml)
        │
        ▼
5. Attest ledger (attest-release-ledger.sh)
        │
        ▼
6. Promote to production (deploy-single-host-production.yml)
```

### 13.4 Required GitHub Secrets (for Deployment)

| Secret | Description |
|--------|-------------|
| `SSH_PRIVATE_KEY` | Deploy SSH key |
| `SSH_KNOWN_HOSTS` | Pinned host fingerprints |
| `DEPLOY_HOST` | VPS hostname or IP |
| `DEPLOY_USER` | SSH username |
| `RELEASE_LEDGER_ATTESTATION_KEY` | Signing key (if `LEDGER_ATTESTATION_REQUIRE_SIGNED=true`) |

---

## 14. Runtime Profiles

Three runtime profiles control feature activation and security posture:

### 14.1 Minimal

**File:** `presets/minimal.env`

- Bare services only
- No observability stack
- In-memory rate limiting
- No mTLS
- Suitable for: local development, CI smoke tests

### 14.2 Standard

**File:** `presets/standard.env`

- Full feature set
- All services enabled
- Redis-backed rate limiting available
- Incident monitoring enabled
- Suitable for: developer workstations, staging

### 14.3 Secure

**File:** `presets/secure.env`

- Full security hardening
- mTLS for gRPC
- Strict rate limits
- Signed ledger attestation required
- Webhook HMAC validation enforced
- Suitable for: production environments

### 14.4 Profile Activation

```bash
# Standard profile (recommended default)
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard

# Secure profile
./scripts/bootstrap.sh secure --force
./scripts/dev-up.sh secure security
```

---

## 15. Multi-Agent Orchestration

The `multi_agent/` directory contains an orchestration framework for AI agent workflows:

```
multi_agent/
├── agents/         # Agent definitions (brief, dispatch, summary, etc.)
├── prompts/        # LLM system and task prompts
├── roles/          # Agent role definitions (operator, developer, reviewer)
├── instructions/   # Orchestration rules and coordination logic
├── tools/          # Agent-callable tools
├── runtime/        # Runtime behavior definitions
├── config.md       # Framework configuration
└── tests/          # Agent workflow test scenarios
```

The `skills/multi-agent-orchestrator/SKILL.md` documents the reusable orchestrator skill interface.

> This is an advanced capability for teams exploring AI-assisted operations and development workflows.

---

## 16. Troubleshooting

### 16.1 Services Won't Start

```bash
# Check Docker is running
docker info

# Check if ports are in use
ss -tlnp | grep -E '4321|8080|8090|5432|6379|27017'

# Check logs for a specific service
docker compose logs api-gateway
docker compose logs frontend

# Re-run doctor
./scripts/dev-doctor.sh
```

### 16.2 Database Connection Errors

```bash
# Check if PostgreSQL is ready
docker compose exec postgres pg_isready

# Re-run migrations
cd backend && go run ./infrastructure/postgres/...

# Check env vars
grep POSTGRES .env.docker.local
```

### 16.3 Frontend Build Errors

```bash
# Clean and reinstall
cd frontend && rm -rf node_modules && bun install

# Check Astro config
cd frontend && ./node_modules/.bin/astro check

# Check TypeScript errors
cd frontend && bun run lint
```

### 16.4 Worker gRPC Errors

```bash
# Check worker is running
docker compose ps calculator

# Check gRPC connectivity (from gateway container)
docker compose exec api-gateway grpcurl \
  -plaintext calculator:7070 list

# Check mTLS cert paths if mTLS is enabled
ls backend/infrastructure/certs/
```

### 16.5 Rate Limit Errors

If you see `429 Too Many Requests`:
- Increase `API_RATE_LIMIT_PER_MINUTE` in `.env.docker.local`
- Or switch to `RATE_LIMIT_STORE=redis` for distributed state

### 16.6 WSL / Docker Desktop Issues

```bash
# If docker binary is docker.exe on WSL
export DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe"
./scripts/dev-up.sh standard
```

### 16.7 Logger Queue Drops

If `logger_drop_total` is increasing:
- Increase `LOGGER_QUEUE_SIZE`
- Increase `LOGGER_WORKERS`
- Check MongoDB connectivity and performance

---

## 17. Glossary

| Term | Definition |
|------|-----------|
| **API Gateway** | The Go service that handles all inbound HTTP traffic, authentication, routing, and metrics |
| **Worker** | The Rust gRPC service that performs compute-intensive operations |
| **Logger Service** | The Go service that persists request logs and incident events to MongoDB |
| **Incident Event** | A record of an alert state change (open, updated, resolved) |
| **Incident Monitor** | The component embedded in the API Gateway that evaluates alert thresholds |
| **Load Shedding** | Rejecting new requests when the inflight request count exceeds `MAX_INFLIGHT_REQUESTS` |
| **Rate Limiting** | Throttling requests per client per time window to prevent abuse |
| **mTLS** | Mutual TLS — both client and server authenticate with certificates |
| **gRPC** | Google Remote Procedure Call — the binary RPC protocol used between Gateway and Worker |
| **RBAC** | Role-Based Access Control — admin vs. developer roles |
| **Release Catalog** | A JSON file tracking released image versions and their metadata |
| **Release Ledger** | A signed record of what was deployed and when |
| **Release Evidence** | Collected artifacts (logs, reports, ledger) proving a release occurred |
| **Restore Drill** | A scripted test of the backup restore process (non-destructive) |
| **Quality Gate** | A script that runs all required checks before a release can proceed |
| **Runtime Profile** | A pre-configured set of env vars (`minimal`, `standard`, `secure`) |
| **SSR** | Server-Side Rendering — Astro renders HTML on the server per request |
| **Island** | A Svelte component that hydrates interactively in an otherwise static Astro page |
| **GHCR** | GitHub Container Registry — where Docker images are published |
| **ADR** | Architecture Decision Record — a document capturing a significant design decision |

---

## 18. Documentation Map

### English Documentation

| Document | Path | Description |
|----------|------|-------------|
| **Wiki (this page)** | `docs/en/wiki.md` | Comprehensive reference |
| Quick Start | `docs/en/quick-start.md` | First run guide |
| Architecture | `docs/en/architecture.md` | System design |
| Developer Guide | `docs/en/developer-guide.md` | Development workflows |
| Operations | `docs/en/operations.md` | Operational procedures |
| Deployment | `docs/en/deployment.md` | Deployment guide |
| Incident Response | `docs/en/incident-response.md` | Incident handling |
| Testing & Quality | `docs/en/testing-and-quality.md` | Test suite guide |
| Project Analysis | `docs/en/project-analysis.md` | Current state analysis |

### Turkish Documentation

| Document | Path | Description |
|----------|------|-------------|
| **Wiki (bu sayfa)** | `docs/tr/wiki.md` | Kapsamlı referans |
| Hızlı Başlangıç | `docs/tr/hizli-baslangic.md` | İlk çalıştırma |
| Mimari | `docs/tr/mimari.md` | Sistem tasarımı |
| Geliştirme Rehberi | `docs/tr/gelistirme-rehberi.md` | Geliştirme iş akışları |
| Operasyonlar | `docs/tr/operasyonlar.md` | Operasyonel prosedürler |
| Dağıtım | `docs/tr/deployment.md` | Dağıtım rehberi |
| Incident Yönetimi | `docs/tr/incident-yonetimi.md` | Incident yönetimi |
| Test ve Kalite | `docs/tr/test-ve-kalite.md` | Test suite rehberi |
| Proje Analizi | `docs/tr/proje-analizi.md` | Güncel durum analizi |

### Core Runbooks & Governance

| Document | Path |
|----------|------|
| Deployment Strategy | `docs/deployment-strategy.md` |
| Runtime Incident Response | `docs/runtime-incident-response.md` |
| Operator Observability Runbook | `docs/operator-observability-runbook.md` |
| Latency Regression Runbook | `docs/latency-regression-runbook.md` |
| API Degradation Runbook | `docs/api-degradation-runbook.md` |
| Dependency Degradation Runbook | `docs/dependency-degradation-runbook.md` |
| Logger Pipeline Runbook | `docs/logger-pipeline-runbook.md` |
| Load Shedding Policy | `docs/load-shedding-policy.json` |
| Load Shedding Runbook | `docs/load-shedding-runbook.md` |
| Release Policy | `docs/release-policy.md` |
| Release Checklist | `docs/release-checklist.json` |
| Branch Protection | `docs/branch-protection-required-checks.md` |
| Delivery Workflow Governance | `docs/delivery-workflow-governance-matrix.md` |
| Nightly Workflow Governance | `docs/nightly-workflow-governance-matrix.md` |
| Technical Analysis (TR) | `docs/appfoundrylab-teknik-analiz.md` |
| Development Plan (TR) | `docs/gelistirmePlanı.md` |

### Architecture Decision Records

Architecture decisions are documented in `docs/adr/`. Each ADR captures:
- Context and problem statement
- Considered options
- Decision outcome
- Consequences

---

*Last updated: 2026-03 | Language: English | Turkish version: [`docs/tr/wiki.md`](../tr/wiki.md)*
