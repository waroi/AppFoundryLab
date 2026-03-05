# Operator Playbook

Purpose:
- Provide practical operating rules for reliable `xN` orchestration in this workspace.

## xN Selection Guide
- `x1`: small, local changes, quick fixes, low risk.
- `x2-x4`: standard implementation tasks with review needs.
- `x5-x8`: cross-cutting tasks (architecture + QA + security + research lanes).
- `x9-x12`: deep audits, migration plans, or high-uncertainty multi-stream work.

## Prompt Quality Rules
- Include outcome + scope + constraints in one prompt.
- Include affected layers or modules when known.
- Include explicit risk focus when relevant (`security`, `regression`, `performance`, `ux`).
- Keep prompt concrete; avoid broad intent-only prompts.

## Default Workflow
1. Run dispatch rule: `multi_agent/tools/dispatch.rule.md`.
2. Run brief rule: `multi_agent/tools/generate-brief.rule.md`.
3. Run orchestration lifecycle rule: `multi_agent/tools/orchestrate-run.rule.md`.
4. Collect agent memos in run `handoffs/` using `multi_agent/instructions/handoff-format.md`.
5. Run summary rule: `multi_agent/tools/summarize-handoffs.rule.md`.
6. Run conflict rule: `multi_agent/tools/detect-conflicts.rule.md`.
7. Run validation checklist: `multi_agent/tools/validate-setup.rule.md`.
8. Generate text and json outputs in markdown as required by rule files.

## Escalation Criteria
- Escalate to `x5+` when:
  - requirements are ambiguous or contradictory
  - multiple risk classes apply (security + reliability + UX)
  - impact crosses architecture boundaries
- Security escalation required when:
  - auth/authz logic changes
  - token/session/secret handling changes
  - externally reachable sensitive paths are modified

## Review Thresholds
- Blocking:
  - unresolved security high-risk finding
  - unresolved regression-critical gap (QA Guardian)
  - conflicting architecture decision without Team Lead resolution
- Non-blocking but tracked:
  - medium risks with clear mitigation and owner
  - optional improvements without release impact

## Telemetry and Budget Guardrails
- Check `## Telemetry` in brief and summary artifacts.
- Brief generator uses compact mode automatically on high `xN` to reduce delegation-pack overhead.
- Check run-level telemetry in run metadata (`run.json -> telemetry.scorecard`).
- Track trend in `run.json -> telemetry.trend` to detect token growth across runs.
- If `status=over_budget`:
  - reduce prompt/context size
  - split task into smaller scoped runs
  - decrease N when duplicate findings are observed

## Phase 3 Reliability Notes
- High `xN` assignments now use adaptive balancing with per-role caps and overflow logic.
- Every orchestration run now includes `conflicts.md/conflicts.json` for contradiction triage.
- CI baseline is available at `.github/workflows/multi-agent-validation.yml`.

## Runtime Hygiene
- Keep run artifacts under `multi_agent/runtime/runs/`.
- Apply retention cleanup rules in `multi_agent/tools/cleanup-runtime.rule.md`.
- Record cleanup reports as markdown in `multi_agent/runtime/cleanup-report-<timestamp>.md`.

## Scriptless Governance
- PowerShell scripts are intentionally removed from orchestration operations.
- Behavior compatibility is preserved through rule contracts and templates in `multi_agent/tools/`.
- If workflow changes are needed, update rule docs first, then roles/prompts/AGENTS references.
