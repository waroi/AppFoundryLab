# Tests

Keep tests focused on behavior and regression-critical paths.

Recommended baseline:
- `integration/smoke-http.sh`: minimal HTTP health/readiness smoke
- `integration/README.md`: env/path configuration guidance
- `integration/load-shed-smoke.sh`: optional overload/load-shedding smoke
- `integration/process-mode-smoke.sh`: process-mode runner + smoke wrapper
