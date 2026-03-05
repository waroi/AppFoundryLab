# Logger Pipeline Runbook

## Trigger

- `logger.unreachable`
- `logger.drop_threshold`
- `gateway.logger_drop`

## First checks

1. Check logger `/health`
2. Check logger `/metrics`
3. Review queue depth, drop ratio, and last error
4. Confirm shared secret and route configuration
5. Check MongoDB availability

## Immediate actions

- restore logger reachability first
- reduce queue pressure if drops are active
- keep stdout incident output enabled if logger persistence is unstable
