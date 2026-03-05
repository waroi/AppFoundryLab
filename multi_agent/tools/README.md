# Scriptless Tool Rules

This directory no longer contains executable scripts.

Each former tool is represented by a rule document:
- `dispatch.rule.md`
- `generate-brief.rule.md`
- `summarize-handoffs.rule.md`
- `detect-conflicts.rule.md`
- `orchestrate-run.rule.md`
- `cleanup-runtime.rule.md`
- `run-unit-tests.rule.md`
- `validate-setup.rule.md`

Shared logic lives under `multi_agent/tools/lib/`:
- `routing-engine.rule.md`
- `safety.rule.md`
- `telemetry.rule.md`

Execution model:
- A Team Lead applies the rule steps deterministically.
- Outputs are written as markdown artifacts under `multi_agent/runtime/`.
- When json output is required, include a fenced `json` block in markdown.
