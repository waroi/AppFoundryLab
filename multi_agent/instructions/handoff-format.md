# Handoff Memo Format

Contract:
- Use these top-level fields exactly with `- Field:` syntax.
- Single-line and multi-line values are both valid.
- For multi-line values, continue on following indented lines.
- Keep the memo concise and scoped to delegated work.

Required fields:
- Task: one sentence describing the delegated task.
- Findings: key facts, observations, or results.
- Decisions: recommendations or choices made, with brief rationale.
- Risks: technical risks, edge cases, or uncertainties. Optional severity tags: `[high]`, `[medium]`, `[low]`.
- Open questions: clarifications needed from Team Lead or user. Use `None.` if empty.
- Suggested next actions: concrete follow-ups or patches.

Optional field:
- Confidence: `high`, `medium`, or `low`.

Token guardrails:
- Prefer bullets or short paragraphs; avoid long prose.
- Avoid repeating the same point across Findings, Decisions, and Risks.
- Include only evidence relevant to the delegated scope.
