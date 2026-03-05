# Cleanup Runtime Rule

Purpose:
- Apply deterministic retention and hygiene on runtime artifacts.

Inputs:
- `runtime_root` (default `multi_agent/runtime`)
- `keep_run_count` (default `20`)
- `max_age_days` (default `14`)
- `what_if` (default `false`)
- `format` (`text` or `json`)

## Run Retention
For `runtime/runs/*` directories (excluding `.gitkeep`):
- Sort by `LastWriteTime desc`.
- Remove when:
  - index >= `keep_run_count`, or
  - older than `max_age_days`.
- Otherwise keep.

## Ephemeral Directory Patterns
Top-level runtime directories matching:
- `^tmp-`
- `test`
- `^validation-handoffs-`
- `^conflict-audit-`

## Ephemeral File Globs
Top-level runtime files matching:
- `validation-*.md`
- `*-test*.md`
- `brief-x*.md`
- `summary-x*.md`

Skip `.gitkeep`.

## Failure Handling
- Cleanup is best-effort.
- On delete failure, do not fail the run.
- Record failures under `failed[]` with reason and error message.

## Json Output Contract
```json
{
  "status": "ok",
  "runtime_root": "D:/w/CodexA2A/multi_agent/runtime",
  "keep_run_count": 20,
  "max_age_days": 14,
  "what_if": true,
  "removed": [],
  "skipped": [],
  "failed": []
}
```
