---
last-updated: 2026-03-06
tracking-version: 3
---

# Agent Performance Metrics

Tracks agent success rates, review efficiency, and blocker readiness.

## Agent Task Metrics
| Agent | Tasks Assigned | Completed | Blocked | Avg Success Score |
| --- | --- | --- | --- | --- |
| team_lead_architect_combined | 0 | 0 | 0 | - |
| principal_architect | 2 | 2 | 0 | 100 |
| team_lead | 2 | 2 | 0 | 100 |
| full_stack_staff_engineer | 2 | 2 | 0 | 100 |
| frontend_engineer | 2 | 2 | 0 | 100 |
| backend_engineer | 2 | 2 | 0 | 100 |
| api_integration_engineer | 0 | 0 | 0 | - |
| research_analyst | 2 | 2 | 0 | 100 |
| qa_guardian | 2 | 2 | 0 | 100 |
| security_reviewer | 1 | 1 | 0 | 100 |
| platform_reliability_engineer | 1 | 1 | 0 | 100 |
| product_strategy_analyst | 1 | 1 | 0 | 100 |
| visual_researcher | 1 | 1 | 0 | 100 |
| delivery_governor | 1 | 1 | 0 | 100 |
| documentation_analyst | 1 | 1 | 0 | 100 |

## Governance Readiness
| Check | Status | Evidence |
| --- | --- | --- |
| `x1` combined ownership | pass | `team-lead-architect-combined.agent.md` exists |
| `x10` core squad | pass | `multi_agent/config.md` named squad |
| `x12` full-stack squad | pass | `multi_agent/config.md` named squad |
| `x13` optional specialist routing | pass | `multi_agent/config.md` + routing rules |
| `x14` dual specialist order | pass | `multi_agent/config.md` fallback cycle |
| clean-code coverage on technical lanes | pass | technical agent bundles include `clean-code` |
| guard blockers hard-stop release runs | pass | review/orchestration rules define `blocked` finalization |
| live reporting timestamp contract | pass | `live-reporting.md` canonical table exists |
| artifact-wide redaction | pass | safety scope covers all generated artifacts |

## Notes
- This file is markdown governance, not runtime telemetry.
- Update it when a development cycle materially changes the agent catalog, review chain, or reporting model.
- The 2026-03-06 phase-closure cycle gathered direct memos from architecture, research, QA, frontend, backend, and full-stack lanes; remaining x14 coverage was integrated by the Team Lead because the live session subagent cap was lower than the canonical roster size.
