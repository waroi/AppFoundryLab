# API Degradation Runbook

## Trigger

- `gateway.error_rate`

## First checks

1. Read `runtime-incident-report`
2. Check `/health/ready`
3. Inspect recent 5xx trend and dependency health
4. Confirm PostgreSQL, Redis, and worker reachability
5. Review recent deploys

## Immediate actions

- reduce risky traffic if needed
- rollback the latest deploy if regressions started immediately after release
- keep an incident record with the downloaded incident report
