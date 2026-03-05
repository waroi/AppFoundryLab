# Starter Variant Selection Guide

Purpose:
- Help new teams choose the smallest starter variant that still matches their delivery risk.
- Prevent copying every example into a new service without a clear reason.

## Decision Table

| Need / Constraint | Recommended Variant | Why | Extra Pieces to Add Only If Needed |
|---|---|---|---|
| I want the smallest Docker-based baseline for local bring-up | `docker-compose.minimal.yml` | Fastest path to first working service | `tests/integration/smoke-http.sh` |
| I want a stricter local baseline closer to hardened runtime defaults | `docker-compose.minimal.yml` + `docker-compose.security.yml` | Read-only fs, fewer privileges, localhost bind reduce accidental exposure | `tests/integration/ci-github-actions-snippet.md` |
| I do not want Docker in the inner loop | `scripts/run-local.sh` | Keeps feedback loop short for process-mode development | `tests/integration/process-mode-smoke.sh` |
| I want process-mode smoke in CI | `scripts/run-local.sh` + `tests/integration/process-mode-smoke.sh` + `tests/integration/ci-github-actions-process-mode-snippet.md` | Reuses the same local path in CI with minimal adaptation | Optional `RUN_LOAD_SHED_SMOKE=true` |
| My service has overload risk or explicit concurrency guard | Base variant + `src/interfaces/http/load_shedding.go.example` | Makes overload behavior explicit instead of implicit timeouts | `tests/integration/load-shed-smoke.sh` |
| I only need baseline health/readiness checks | Any base variant + `tests/integration/smoke-http.sh` | Avoids premature complexity | Do not copy load shedding or security override unless there is a concrete need |

## Selection Rules
1. Start from exactly one base runtime path: `docker-compose.minimal.yml` or `scripts/run-local.sh`.
2. Add `docker-compose.security.yml` only when you need a harder Docker baseline.
3. Add load shedding sample only if the service has a real concurrency, burst, or backpressure problem.
4. Prefer smoke scripts before adding heavier custom integration tooling.
5. If two variants seem necessary on day one, document the reason in an ADR before copying both flows.

## Anti-Patterns
- Copying `docker-compose.security.yml` into every new service without testing file-system or bind assumptions.
- Copying load shedding middleware before the service has measured overload risk.
- Running both compose-mode and process-mode in CI without a clear difference in confidence gained.
- Treating starter examples as mandatory framework pieces instead of opt-in references.
