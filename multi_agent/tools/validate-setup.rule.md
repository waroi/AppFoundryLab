# Validate Setup Rule

Purpose:
- Run integration-grade checks for dispatch, brief, summary, conflicts, orchestration, cleanup, and unit regressions.

Outputs:
- `multi_agent/runtime/validation-report.md`
- Optional json block in same file.

## Required File Checklist
- `multi_agent/config.md`
- `multi_agent/instructions/orchestration.md`
- `multi_agent/instructions/handoff-format.md`
- `multi_agent/instructions/operator-playbook.md`
- `multi_agent/instructions/token-optimization.md`
- `multi_agent/tools/README.md`
- `multi_agent/tools/dispatch.rule.md`
- `multi_agent/tools/generate-brief.rule.md`
- `multi_agent/tools/summarize-handoffs.rule.md`
- `multi_agent/tools/detect-conflicts.rule.md`
- `multi_agent/tools/orchestrate-run.rule.md`
- `multi_agent/tools/cleanup-runtime.rule.md`
- `multi_agent/tools/run-unit-tests.rule.md`
- `multi_agent/tools/lib/routing-engine.rule.md`
- `multi_agent/tools/lib/safety.rule.md`
- `multi_agent/tools/lib/telemetry.rule.md`
- `multi_agent/tests/orchestration-regression.md`

## Integration Scenarios
1. Dispatch behavior:
   - `analysis roadmap x4` => `agent_count=4`, routing hit exists
   - stdin-equivalent prompt behavior must preserve task and count
2. Allocation behavior:
   - sample x4 yields 4 assignments
   - x1 maps to combined role
   - uppercase suffix parse works
   - debug must not trigger `bug`
   - optimization terms route planning
   - security-heavy prompt ranks security first and includes security reviewer
   - high N produces duplicate suffixes
3. Brief behavior:
   - required sections exist
   - includes telemetry and safety
   - x10 uses compact mode
4. Summary behavior:
   - multiline field extraction preserved
   - includes risk/open questions/action queue/safety/telemetry
5. Conflict behavior:
   - synthetic contradictory memos produce `conflict_count >= 1`
6. Orchestrate run behavior:
   - creates all required artifacts
   - handoff scaffold count matches `agent_count`
   - includes telemetry scorecard
7. Cleanup behavior:
   - supports what-if style reporting
   - reports `removed/skipped/failed`
8. Unit behavior:
   - run all 10 cases from `run-unit-tests.rule.md`

## Validation Status
- `passed`: no failed checks.
- `passed_with_warnings`: checks pass but cleanup/reporting warnings exist.
- `failed`: one or more blocking checks fail.

## Json Output Contract
```json
{
  "status": "passed_with_warnings",
  "checks": {
    "files": "pass",
    "dispatch": "pass",
    "brief": "pass",
    "summary": "pass",
    "conflicts": "pass",
    "orchestration": "pass",
    "cleanup": "pass",
    "unit": "pass"
  },
  "warnings": [],
  "errors": []
}
```
