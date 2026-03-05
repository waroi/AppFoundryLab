# Quick Start

## 1. First local run

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

If the default local ports are already occupied, edit `.env.docker.local` or export `FRONTEND_HOST_PORT`, `API_GATEWAY_HOST_PORT`, and `LOGGER_HOST_PORT` before restarting the stack. Example: `FRONTEND_HOST_PORT=14321 API_GATEWAY_HOST_PORT=18080 LOGGER_HOST_PORT=18090 ./scripts/dev-up.sh standard security`. Local Docker publishing defaults to `DOCKER_HOST_BIND_ADDRESS=127.0.0.1`.

## 2. Open the stack

- Frontend: `http://127.0.0.1:<FRONTEND_HOST_PORT>/` (default: `http://127.0.0.1:4321/`)
- Frontend test page: `http://127.0.0.1:<FRONTEND_HOST_PORT>/test` (default: `http://127.0.0.1:4321/test`)
- Turkish home: `http://127.0.0.1:<FRONTEND_HOST_PORT>/tr`
- Turkish test page: `http://127.0.0.1:<FRONTEND_HOST_PORT>/tr/test`
- API gateway: `http://127.0.0.1:<API_GATEWAY_HOST_PORT>` (default: `http://127.0.0.1:8080`)
- Logger metrics: `http://127.0.0.1:<LOGGER_HOST_PORT>/metrics` (default: `http://127.0.0.1:8090/metrics`)

## 3. Verify locale and theme

- Use the top-right toolbar on `/` and `/test`
- Switch between `EN` and `TR`
- Switch between `Light` and `Dark`
- Confirm the language switch navigates between `/` and `/tr`, or between `/test` and `/tr/test`
- Reload the page and confirm the current URL keeps the selected locale while the selected theme persists
- Confirm the document updates `html[lang]` and `html[data-theme]`

## 4. What to try after admin login

- review runtime config
- review runtime metrics
- download the runtime report
- download the incident report
- inspect recent incident events

## 5. Next docs

- [developer-guide.md](/mnt/d/w/AppFoundryLab/docs/en/developer-guide.md)
- [operations.md](/mnt/d/w/AppFoundryLab/docs/en/operations.md)
- [incident-response.md](/mnt/d/w/AppFoundryLab/docs/en/incident-response.md)
- [deployment.md](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)

## 6. If you want the single-host deployment package locally

```bash
cp .env.single-host.example .env.single-host
./scripts/deploy-single-host.sh up ./.env.single-host
./scripts/archive-runtime-report.sh http://127.0.0.1:<API_GATEWAY_HOST_PORT> admin strong_password
```
