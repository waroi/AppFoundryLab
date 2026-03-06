# Dependency Degradation Runbook

## Trigger

- `health.degraded`
- runtime diagnostics warning: `strict dependencies are disabled; dependency-backed routes degrade per endpoint`

## First checks

1. Identify which dependency is down in readiness checks
2. Confirm whether the issue is PostgreSQL, Redis, or worker
3. Check network reachability and credential changes
4. Review recent maintenance or restarts

## Endpoint matrix

| Route | Dependency | `STRICT_DEPENDENCIES=true` | `STRICT_DEPENDENCIES=false` | Runtime behavior |
| --- | --- | --- | --- | --- |
| `GET /health/ready` | Postgres, Redis, worker | Gateway fails fast if a required dependency cannot initialize. | Gateway can boot with degraded readiness. | Returns `503` with per-dependency checks until recovery. |
| `GET /api/v1/users` | Postgres | Gateway fails fast on Postgres init, ping, or migration failure. | Gateway keeps serving. | Returns `200` demo users only when `DEMO_FALLBACK_USERS=true`; otherwise `503 users_unavailable`. |
| `POST /api/v1/compute/fibonacci` and `POST /api/v1/compute/hash` | Worker | Gateway fails fast if the worker gRPC client cannot initialize or pass health. | Gateway keeps serving without a worker client. | Returns `503 worker_unavailable` until worker recovery; in-flight worker RPC failures return `502 worker_call_failed`. |
| Auth and API distributed rate limiting | Redis | Gateway fails fast if Redis init or ping fails. | Gateway keeps serving without a healthy Redis client. | `RATE_LIMIT_REDIS_FAILURE_MODE=open` keeps traffic flowing; `closed` returns `503 rate_limiter_unavailable`. |
| `GET /api/v1/admin/request-logs` | Logger | No startup change; logger is optional. | No startup change; logger is optional. | Returns `200` with an empty list when `LOGGER_ENDPOINT` is unset; otherwise logger failures return `503 logger_unavailable`. |
| `GET /api/v1/admin/runtime-metrics` and `GET /api/v1/admin/runtime-report` | Logger | No startup change; logger is optional. | No startup change; logger is optional. | Returns `200` and surfaces logger reachability, degraded health, and warnings in the payload. |

This same matrix is also exposed through `GET /api/v1/admin/runtime-config` so operators can confirm the active degradation contract from the running system.

## Immediate actions

- restore the failing dependency
- keep the runtime incident report with the ticket
- monitor for follow-up alerts after recovery

## Escalation notes

- If Redis is failing and rate limiting is configured `closed`, treat the incident as user-facing even when the API process is still live.
- If Postgres is failing under `DEMO_FALLBACK_USERS=true`, treat the system as degraded rather than healthy; the fallback only exists to keep the starter usable.
- If the logger is failing, capture the runtime report before restart because the report still contains the degraded health context.
