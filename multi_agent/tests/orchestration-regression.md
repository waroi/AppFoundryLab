# Orchestration Regression Matrix

This file mirrors the 10-unit regression cases in `multi_agent/tools/run-unit-tests.rule.md`.

| ID | Case | Expected |
| --- | --- | --- |
| 1 | `Refactor API x4` | `agent_count=4`, task trimmed |
| 2 | `Refactor API X5` | `agent_count=5` |
| 3 | `Refactor API x0` | normalized to `agent_count=1` |
| 4 | `debug dashboard` + `bug` in word mode | no match |
| 5 | `debug dashboard` + `bug` in substring mode | match |
| 6 | `deep analysis roadmap x4` | routing includes `planning_analysis` |
| 7 | `deep optimization improvement x4` | routing includes `planning_analysis` |
| 8 | `auth security oauth jwt test regression x4` | first routing hit is `security` |
| 9 | `basic task x1` | `team_lead_architect_combined` only |
| 10 | `backend hardening x8` | at least one `_2` assignment |
