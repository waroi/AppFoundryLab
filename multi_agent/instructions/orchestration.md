# Orchestration Rules (Enterprise Scriptless)

Goal:
- Execute user tasks via deterministic multi-agent workflow controlled by trailing `xN` / `XN`.
- Keep the operating model scriptless while still enforcing release gates, safety, and token discipline.

Canonical sources:
- `multi_agent/config.md`
- `multi_agent/instructions/agent-catalog.md`
- `multi_agent/instructions/clean-code-standards.md`
- `multi_agent/instructions/enterprise-squad.md`
- `multi_agent/instructions/context-packing.md`
- `multi_agent/instructions/project-context-discovery.md`
- `multi_agent/instructions/review-chain.md`
- `multi_agent/instructions/skill-loading.md`
- `multi_agent/instructions/live-reporting.md`
- `multi_agent/instructions/model-selection.md`
- `multi_agent/instructions/continuous-improvement.md`
- `multi_agent/instructions/runtime-governance.md`
- `multi_agent/instructions/handoff-format.md`
- `multi_agent/tools/*.rule.md`

## Core Flow
1. Parse prompt envelope.
2. Load required project context, including `AGENTS.md` when workspace operating rules matter.
3. Compute routing hits.
4. Select assignment strategy.
5. Resolve `release_oriented` and `data_sensitivity` from `multi_agent/config.md`.
6. Resolve model assignments and escalations.
7. Update active plan.
8. Produce dispatch artifact.
9. Produce skill-aware brief artifact.
10. Emit live status reporting.
11. Collect handoff memos.
12. Run review chain, blocker checks, and conflict detection.
13. Produce summary, metrics, and session memory artifacts.
14. Run final deep analysis, doc sync, metrics update, and score refresh.
15. Validate the run with `multi_agent/tools/validate-setup.rule.md`.

## Assignment Rules
- If `N == 1`: assign `team_lead_architect_combined`.
- If `N == 10`: assign the exact `enterprise_x10_core` squad.
- If `N == 12`: assign the exact `enterprise_x12_full_stack` squad.
- If `2 <= N < 10`:
  - assign `allocation.primary_agents`
  - insert routing-priority agents in routing rank order
  - dedupe while preserving first appearance order
  - fill remaining slots from `allocation.fallback_cycle`
- If `N == 11`:
  - seed `enterprise_x10_core`
  - add the top-ranked agent from `allocation.expansion_cycle` that matches the prompt and is not already present
- If `N == 13`:
  - seed `enterprise_x12_full_stack`
  - add the optional specialist resolved from `allocation.optional_specialists.routing_map`
  - if no optional specialist matches, add the first missing specialist from `allocation.optional_specialists.fallback_cycle`
- If `N == 14`:
  - seed `enterprise_x12_full_stack`
  - add both optional specialists from `allocation.optional_specialists.fallback_cycle`

## Review Loop
- `release_oriented` detection must use `policies.governance.release_oriented_keywords` from `multi_agent/config.md`.
- Guard-lane findings from `policies.governance.blocker_authorities` are blocking for release-oriented tasks.
- Clean-code violations are blocking for coding and code-review lanes.
- Engineering outputs are reviewed by `team_lead` and, when architecture changes, by `principal_architect`.
- Documentation changes are reviewed by `team_lead` when `documentation_analyst` is active.
- Maximum 2 revision rounds.
- If an unresolved guard blocker remains after 2 rounds, the final run state is `blocked`.

## Reporting Loop
- Team Lead must expose `timestamp`, `slot`, `agent`, `mission`, `skill bundle`, `state`, `blockers`, and `success level` during execution.
- Live reporting must follow the canonical table format and cadence from `multi_agent/instructions/live-reporting.md`.
- Final response must include an agent scoreboard, blocker resolution section, and final run state.
- Every development cycle must end with deep analysis, documentation updates, metrics update, and score refresh unless the user explicitly opts out.
