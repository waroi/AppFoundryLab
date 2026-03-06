# Tools Overview

Core rule documents:
- `dispatch.rule.md`
- `generate-brief.rule.md`
- `summarize-handoffs.rule.md`
- `validate-setup.rule.md`
- `run-unit-tests.rule.md`

Library rules:
- `lib/routing-engine.rule.md`
- `lib/safety.rule.md`
- `lib/telemetry.rule.md`

Support rules still available:
- `detect-conflicts.rule.md`
- `orchestrate-run.rule.md`
- `cleanup-runtime.rule.md`

Rule design principles:
- canonical worker semantics come from `agents`
- skills, memory, and metrics are first-class governance inputs
- clean-code is mandatory for every technical lane
- release blockers and clean-code failures hard-stop finalization
- runtime artifacts are redacted before write and retained only as generated state
- every substantive cycle ends with deep analysis and doc refresh
