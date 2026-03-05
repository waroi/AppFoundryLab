# Validate Setup Rule

Purpose:
- Run integration-grade checks for dispatch, brief, summary, orchestration, memory, metrics, and regression contracts.

## Required File Checklist
- `multi_agent/config.md`
- `multi_agent/instructions/agent-catalog.md`
- `multi_agent/instructions/clean-code-standards.md`
- `multi_agent/instructions/orchestration.md`
- `multi_agent/instructions/enterprise-squad.md`
- `multi_agent/instructions/context-packing.md`
- `multi_agent/instructions/project-context-discovery.md`
- `multi_agent/instructions/review-chain.md`
- `multi_agent/instructions/session-memory.md`
- `multi_agent/instructions/task-planning.md`
- `multi_agent/instructions/skill-loading.md`
- `multi_agent/instructions/live-reporting.md`
- `multi_agent/instructions/model-selection.md`
- `multi_agent/instructions/continuous-improvement.md`
- `multi_agent/instructions/runtime-governance.md`
- `multi_agent/instructions/quality-scorecard.md`
- `multi_agent/instructions/handoff-format.md`
- `multi_agent/instructions/operator-playbook.md`
- `multi_agent/instructions/token-optimization.md`
- `multi_agent/tools/dispatch.rule.md`
- `multi_agent/tools/generate-brief.rule.md`
- `multi_agent/tools/summarize-handoffs.rule.md`
- `multi_agent/tools/validate-setup.rule.md`
- `multi_agent/tools/run-unit-tests.rule.md`
- `multi_agent/tools/lib/routing-engine.rule.md`
- `multi_agent/tools/lib/telemetry.rule.md`
- `multi_agent/tools/lib/safety.rule.md`
- `multi_agent/tools/orchestrate-run.rule.md`
- `multi_agent/tools/detect-conflicts.rule.md`
- `multi_agent/tests/orchestration-regression.md`
- `multi_agent/metrics/agent-performance.md`
- `multi_agent/metrics/token-usage.md`
- `multi_agent/memory/sessions/_session-template.md`
- `multi_agent/todo/active-plan.md`
- `skills/clean-code/SKILL.md`

## Integration Scenarios
- `x10` core squad exists and is unique.
- `x12` full-stack squad exists and includes dedicated frontend/backend lanes.
- `x13` selects the routing-matched optional specialist.
- `x14` includes both optional specialists in fallback order.
- API and documentation specialists are routeable.
- skill bundles are documented per agent.
- clean-code is present in every coding and code-review skill bundle.
- live reporting fields include timestamp, agent, mission, skill bundle, blockers, and success level.
- release-oriented detection uses canonical governance keywords.
- unresolved guard blockers force final run state `blocked`.
- runtime governance documents artifact-wide redaction, retention, and recovery.
- session memory, metrics, and active plan artifacts exist.
- active plan includes blockers and success checks.
- legacy `roles/` and `prompts/` directories are absent and not referenced by canonical docs.
- end-of-cycle deep analysis, doc sync, metrics update, and score refresh are documented as default.
- run all 30 cases from `run-unit-tests.rule.md`.

## Result Contract
Return either text or json with:
- `status: pass|fail`
- `generated_at`
- `case_results[]`
- `failed_cases[]`
- `notes`
