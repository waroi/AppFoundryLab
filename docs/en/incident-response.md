# Incident Response

Use the runtime diagnostics panel and the admin incident endpoints as the first place to investigate runtime problems.

## Main endpoints

- `GET /api/v1/admin/runtime-report`
- `GET /api/v1/admin/runtime-incident-report`
- `GET /api/v1/admin/incident-events`

## What the incident report contains

- recommended severity
- incident category
- title and short summary
- mapped runbooks
- next actions
- evidence from health, alerts, and logger state

## Persistent journal behavior

- the gateway emits incident events when alerts open, update after the dedupe window, or resolve
- the logger service stores those events in MongoDB
- the monitor can now fan out to `logger`, `stdout`, `webhook`, or combinations of them
- the admin UI shows the most recent events so operators can see whether the issue is new or recurring

## Operational follow-ups

- archive deploy-time diagnostics with [archive-runtime-report.sh](/mnt/d/w/AppFoundryLab/scripts/archive-runtime-report.sh)
- prune old incident records with [prune-incident-events.sh](/mnt/d/w/AppFoundryLab/scripts/prune-incident-events.sh)
- use [single-host-ops.yml](/mnt/d/w/AppFoundryLab/.github/workflows/single-host-ops.yml) to run remote pruning or rollback

## Related docs

- Canonical incident flow: [runtime-incident-response.md](/mnt/d/w/AppFoundryLab/docs/runtime-incident-response.md)
- Deployment: [deployment.md](/mnt/d/w/AppFoundryLab/docs/en/deployment.md)
