# Developer Guide

## 1. Project map

- `frontend/`: Astro app with Svelte islands
- `backend/services/api-gateway/`: HTTP API, JWT auth, RBAC, health, metrics, incident monitor
- `backend/services/logger/`: request log ingest and persistent incident journal
- `backend/core/calculator/`: Rust gRPC worker
- `scripts/`: local automation and quality gates
- `.github/workflows/`: CI/CD

## 2. Recommended local loop

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

## 3. Frontend change map

Add frontend presentation work here:

- document shell and pre-paint preference bootstrap: `frontend/src/layouts/BaseLayout.astro`
- localized route mapping: `frontend/src/lib/ui/routes.ts`
- locale/theme store and normalization: `frontend/src/lib/ui/preferences.ts`
- shared copy dictionary and formatters: `frontend/src/lib/ui/copy.ts`
- shared page shells for default and localized routes: `frontend/src/components/Page/`
- shared controls: `frontend/src/components/Layout/`
- locale-reactive shell copy: `frontend/src/components/Static/`
- diagnostics and restore-drill surfaces: `frontend/src/components/Interactive/`
- semantic tokens and theme classes: `frontend/src/styles/global.scss`

Rules:

- add every new visible string to both locales in `frontend/src/lib/ui/copy.ts`
- keep locale routing canonical through `frontend/src/lib/ui/routes.ts` instead of hand-building `href` values
- do not add new light-only utility colors when a semantic shared class can express the same intent
- if a UI state matters to smoke/e2e, add stable `data-testid` or `data-*` hooks instead of relying on visible translated text

## 4. Backend change map

Add new work here:

- handlers: `backend/services/api-gateway/internal/handlers/`
- middleware: `backend/services/api-gateway/internal/middleware/`
- runtime config: `backend/services/api-gateway/internal/runtimecfg/`
- incident monitor: `backend/services/api-gateway/internal/incidents/`
- logger persistence: `backend/services/logger/internal/`

## 5. Admin diagnostics endpoints

- `GET /api/v1/admin/runtime-config`
- `GET /api/v1/admin/runtime-metrics`
- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`

## 6. Practical rule

Whenever you add a new operational behavior:

1. add code
2. add tests
3. update docs in the same change set
4. update the project analysis and development plan if the architecture meaningfully changed

## 7. Read next

- [quick-start.md](/mnt/d/w/AppFoundryLab/docs/en/quick-start.md)
- [operations.md](/mnt/d/w/AppFoundryLab/docs/en/operations.md)
- [incident-response.md](/mnt/d/w/AppFoundryLab/docs/en/incident-response.md)
- [deployment.md](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)
