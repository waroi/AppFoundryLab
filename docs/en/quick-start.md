# Quick Start

## 1. Bring the local stack up

```bash
./scripts/dev-doctor.sh
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

`bootstrap.sh` generates `.env`, a gitignored `.env.docker.local`, and the local development certificates. `.env.docker` stays as the checked-in template.

If `dev-doctor` reports `docker compose unavailable` on WSL, enable Docker Desktop WSL integration or rerun with:

```bash
DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe" ./scripts/dev-doctor.sh
DOCKER_BIN="/mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe" ./scripts/dev-up.sh standard
```

## 2. Understand the local health contract

- `GET /health/live` means the gateway process is running
- `GET /health/ready` means the dependency-backed stack is usable
- `GET /healthz` on the frontend is the lightweight shell health endpoint
- `dev-up` now waits for readiness, logger reachability, and one authenticated admin endpoint before reporting success

Default URLs:
- Frontend: `http://127.0.0.1:4321/`
- Frontend health: `http://127.0.0.1:4321/healthz`
- API live: `http://127.0.0.1:8080/health/live`
- API ready: `http://127.0.0.1:8080/health/ready`
- Logger metrics: `http://127.0.0.1:8090/metrics`

## 3. Run the first browser smoke

- Open `http://127.0.0.1:4321/`
- Sign in as `admin`
- Use the password printed by `./scripts/bootstrap.sh` or stored in the generated local `.env.docker.local` under `BOOTSTRAP_ADMIN_PASSWORD`
- Confirm that the runtime summary, runtime knobs panel, trace lookup panel, and request-log list load successfully

Automated real-stack browser smoke:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e:live
```

Mock-backed UI regression:

```bash
cd frontend
../.toolchain/bun/bin/bun run e2e
```

If Playwright is missing Linux runtime libraries, bootstrap them once:

```bash
./scripts/bootstrap-playwright-linux.sh --frontend-dir frontend
```

After that, both `e2e` and `e2e:live` automatically load `frontend/.playwright-linux.env`.

## 4. Reset local state when credentials drift

If persisted Postgres or Mongo volumes reject the configured credentials:

```bash
./scripts/dev-down.sh standard --volumes
./scripts/bootstrap.sh standard --force
./scripts/dev-up.sh standard
```

## 5. Next docs

- [Developer Guide](/mnt/d/w/AppFoundryLab/docs/en/developer-guide.md)
- [Operations](/mnt/d/w/AppFoundryLab/docs/en/operations.md)
- [Testing and Quality](/mnt/d/w/AppFoundryLab/docs/en/testing-and-quality.md)
- [Project Analysis](/mnt/d/w/AppFoundryLab/docs/en/project-analysis.md)
