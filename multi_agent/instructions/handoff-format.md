# Handoff Memo Format

Contract:
- Use these top-level fields exactly with `- Field:` syntax.
- Single-line and multi-line values are valid.
- Keep the memo concise, scoped, and evidence-backed.

Required fields:
- Agent: delegated worker agent name.
- Task: one sentence describing the delegated task.
- Skill bundle: loaded skills for the slot.
- Status: `complete`, `partial`, or `blocked`.
- Success score: integer `0-100`.
- Findings: the relevant facts, observations, or results.
- Evidence: direct references, files, conditions, or reasoning anchors behind the findings.
- Decisions: recommendations or choices made, with brief rationale.
- Risks: technical risks, edge cases, or uncertainties. Optional severity tags: `[high]`, `[medium]`, `[low]`.
- Open questions: clarifications needed from Team Lead or user. Use `None.` if empty.
- Suggested next actions: concrete follow-ups or patches.

Optional fields:
- Dependencies: upstream/downstream coupling, owners, or rollout dependencies.
- Confidence: `high`, `medium`, or `low`.
- Files touched: canonical files or code paths affected.

Token guardrails:
- Prefer bullets or short paragraphs.
- Do not repeat the same point across Findings, Evidence, and Risks.
- Include only evidence relevant to the delegated lane.
