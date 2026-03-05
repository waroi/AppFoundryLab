# Orchestration Rules (Scriptless)

Goal:
- Execute user tasks via a deterministic multi-agent workflow controlled by trailing `xN` / `XN`.
- Preserve previous dispatch/brief/summary/conflict/telemetry behavior without executable scripts.

Canonical sources:
- `multi_agent/config.md`
- `multi_agent/instructions/handoff-format.md`
- `multi_agent/tools/*.rule.md`

## Core Flow
1. Parse prompt envelope.
2. Compute role assignments.
3. Produce dispatch artifact.
4. Produce brief artifact.
5. Collect handoff memos.
6. Produce summary artifact.
7. Produce conflict artifact.
8. Produce run metadata artifact.
9. Validate the run against `multi_agent/tools/validate-setup.rule.md`.

## Prompt Envelope Rules
- Read trailing `xN` or `XN` from the end of prompt.
- If suffix missing, use `defaults.agent_count`.
- Normalize `<1` to `1`.
- Cap by `defaults.max_agent_count`.
- If task text becomes empty, use `Untitled task`.

## Routing Rules
- Evaluate routing groups in `multi_agent/config.md`.
- Keyword matching is boundary-aware unless keyword explicitly sets substring mode.
- Compute group score using configured weights.
- Keep only groups with at least one matched keyword.
- Sort by:
  1. score (descending)
  2. match_count (descending)
  3. declaration order (ascending)

## Assignment Rules
- If `N == 1`: assign `team_lead_architect_combined`.
- If `N >= 2`: assign primary roles first (`principal_architect`, `team_lead`).
- Add routing-priority roles in routing rank order, deduplicated against already assigned role keys.
- Fill remaining slots from `allocation.fallback_cycle`.
- If role caps block completion, continue with `allocation.overflow_cycle`.
- If still blocked, lift role caps and continue cycle fill.
- Duplicate role instance naming:
  - first instance: `role_key`
  - second+ instance: `role_key_2`, `role_key_3`, ...

## Delegation Rules
- Team Lead delegates scoped, non-overlapping tasks.
- Each delegated slot must include:
  - mission
  - scoped context
  - output budget
  - quality gate
  - overlap guard (for duplicate instances)
- For high `xN` (`>= 8`), use compact brief mode.

## Safety and Telemetry Rules
- Apply redaction policy from `multi_agent/tools/lib/safety.rule.md`.
- Compute telemetry from `multi_agent/tools/lib/telemetry.rule.md`.
- Include budget status (`ok`, `over_budget`, `no_budget`) for task/brief/summary.

## Output Modes
- Every operational artifact supports two representations:
  - `text`: human-readable markdown sections
  - `json`: fenced `json` block embedded in markdown
- JSON field contracts are defined in each `*.rule.md` file under `multi_agent/tools/`.

## Review Loop
- If memos conflict, Team Lead requests revision or resolves with explicit rationale.
- QA Guardian and Security Reviewer findings are blocking for release-oriented tasks.
- If requirements are ambiguous, ask one concise user question.
