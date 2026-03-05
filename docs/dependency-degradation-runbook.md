# Dependency Degradation Runbook

## Trigger

- `health.degraded`

## First checks

1. Identify which dependency is down in readiness checks
2. Confirm whether the issue is PostgreSQL, Redis, or worker
3. Check network reachability and credential changes
4. Review recent maintenance or restarts

## Immediate actions

- restore the failing dependency
- keep the runtime incident report with the ticket
- monitor for follow-up alerts after recovery
