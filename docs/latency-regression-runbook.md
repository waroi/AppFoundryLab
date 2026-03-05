# Latency Regression Runbook

## Trigger

- `gateway.latency`

## First checks

1. Compare current latency trend with recent history
2. Check worker and logger backpressure
3. Inspect readiness cache age and dependency health
4. Review recent capacity changes

## Immediate actions

- reduce expensive traffic
- review retry amplification
- scale gateway or downstream dependencies if saturation is confirmed
