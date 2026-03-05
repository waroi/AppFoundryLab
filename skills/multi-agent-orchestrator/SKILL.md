# Skill: multi-agent-orchestrator

Use this skill when a user prompt ends with `xN` to activate multi-agent orchestration.

Workflow:
1. Read `AGENTS.md` and `multi_agent/instructions/orchestration.md`.
2. Parse trailing `xN` with rules in `multi_agent/tools/lib/routing-engine.rule.md`.
3. Allocate roles using `multi_agent/config.md` (primary + routing + fallback + overflow).
4. Produce dispatch artifact using `multi_agent/tools/dispatch.rule.md`.
5. Produce/update brief using `multi_agent/tools/generate-brief.rule.md`.
6. Delegate tasks using templates in `multi_agent/prompts/`.
7. Collect handoff memos using `multi_agent/instructions/handoff-format.md`.
8. Summarize handoffs with `multi_agent/tools/summarize-handoffs.rule.md`.
9. Detect contradictions with `multi_agent/tools/detect-conflicts.rule.md`.
10. Run final checklist from `multi_agent/tools/validate-setup.rule.md`.
11. Team Lead integrates outputs and prepares the final response.

Notes:
- Keep context minimal for each agent.
- Ask the user only if a single clarification is required.
- Produce both markdown text and fenced json payloads when rule contracts require automation outputs.
