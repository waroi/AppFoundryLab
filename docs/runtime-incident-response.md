# Runtime Incident Response

This document explains how runtime diagnostics now turns into incident response.

## 1. Main endpoints

- `GET /api/v1/admin/runtime-config`
- `GET /api/v1/admin/runtime-metrics`
- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`
- `GET /api/v1/admin/request-logs`

## 2. What changed

- `runtime-report` includes incident metadata and mapped runbooks
- the gateway emits incident events on alert state changes
- the logger service persists incident events and trace-correlated request logs in MongoDB
- webhook fan-out now supports HMAC signing and host allowlists
- the admin API can query recent request logs by `traceId`

## 3. Severity rules

- `sev-1`: at least one active critical alert
- `sev-2`: active alerts exist but highest severity is warning
- `sev-3`: no active alerts, but recent breaches exist
- `sev-4`: no active alerts and no recent breaches

## 4. Persistent incident journal

The gateway incident monitor evaluates current alerts on a timer. When an alert opens, changes materially after the dedupe window, or resolves, it emits an incident event.

Supported sink modes:

- `disabled`
- `logger`
- `stdout`
- `webhook`
- `logger+stdout`
- `logger+webhook`
- `stdout+webhook`
- `logger+stdout+webhook`

Key environment variables:

- `INCIDENT_EVENT_SINK`
- `INCIDENT_EVENT_INTERVAL_MS`
- `INCIDENT_EVENT_DEDUPE_WINDOW_SECONDS`
- `INCIDENT_EVENT_WEBHOOK_URL`
- `INCIDENT_EVENT_WEBHOOK_HMAC_SECRET`
- `INCIDENT_EVENT_WEBHOOK_ALLOWED_HOSTS`
- `INCIDENT_EVENT_RETENTION_DAYS`
- `MONGO_INCIDENT_COLLECTION`

## 5. Operator flow

1. Login as `admin`.
2. Check active alerts and recommended severity.
3. Open the mapped runbook.
4. Review persistent incident events.
5. Query `GET /api/v1/admin/request-logs?traceId=<id>` when you need request-level correlation.
6. After deploys, archive the runtime report with [archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh).
7. Back up the logger database as part of [backup-single-host.sh](/mnt/d/w/AppFoundryLab/scripts/backup-single-host.sh).
8. Prune old incident records with [prune-incident-events.sh](/mnt/d/w/AppFoundryLab/scripts/prune-incident-events.sh).

## 6. Current state

- the journal persists to Mongo via the logger service
- webhook fan-out is still available, but now expects signed and allowlisted destinations
- request-log trace correlation is available through the logger backend and admin API
- restore drills and host-level incident recovery are now scripted, but real-host evidence still needs to be captured outside this workspace
