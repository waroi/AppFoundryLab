# Orchestration Regression Matrix

This file mirrors the 30-unit regression cases in `multi_agent/tools/run-unit-tests.rule.md`.

| ID | Case | Expected |
| --- | --- | --- |
| 1 | `Refactor API x4` | `agent_count=4`, task trimmed |
| 2 | `Refactor API X5` | `agent_count=5` |
| 3 | `Refactor API x0` | normalized to `agent_count=1` |
| 4 | `debug dashboard` + `bug` in word mode | no match |
| 5 | `deep analysis roadmap x4` | routing includes `planning_analysis` |
| 6 | `frontend accessibility redesign x6` | `frontend_engineer` is routeable |
| 7 | `backend auth migration x6` | `backend_engineer` is routeable |
| 8 | `schema contract integration x6` | `api_integration_engineer` is routeable |
| 9 | `runbook handbook refresh x6` | `documentation_analyst` is routeable |
| 10 | `basic task x1` | `team_lead_architect_combined` only |
| 11 | `enterprise modernization x10` | 10 unique assignments |
| 12 | `frontend redesign x11` | core squad + `frontend_engineer` |
| 13 | `enterprise modernization x12` | fixed full-stack squad |
| 14 | `schema contract review x13` | `api_integration_engineer` selected |
| 15 | `runbook onboarding refresh x13` | `documentation_analyst` selected |
| 16 | `api contract onboarding x14` | both optional specialists included |
| 17 | `auth hardening x12` | principal/security models map correctly |
| 18 | `production latency rollback plan x12` | QA or reliability escalates |
| 19 | `enterprise onboarding redesign x12` | context-packing limits cover all four pods |
| 20 | `frontend onboarding redesign x12` | frontend skill bundle includes `clean-code` |
| 21 | docs review | clean-code governance docs exist |
| 22 | agent catalog review | coding and review agents include `clean-code` |
| 23 | docs review | memory, metrics, and active plan exist |
| 24 | repo hygiene review | no legacy worker mirror directories remain |
| 25 | `enterprise delivery review x12` | live reporting fields include timestamp, agent, mission, skill bundle, blockers, success level |
| 26 | `production rollback review x12` | dispatch resolves `release_oriented=true` |
| 27 | guard blocker unresolved | final run state is `blocked` |
| 28 | `customer data contract review x12` | dispatch resolves `data_sensitivity=restricted` |
| 29 | runtime artifact review | safety scope covers all generated artifacts |
| 30 | any development prompt | final cycle includes analysis, doc sync, metrics, score refresh |
