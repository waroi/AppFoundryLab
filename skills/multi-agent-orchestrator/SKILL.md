---
name: multi-agent-orchestrator
description: Scriptless orchestration skill for xN-driven multi-agent execution, squad selection, routing, brief generation, live reporting, review chain, and end-of-cycle audit. Use when a prompt ends with `xN` or when multi-agent delegation is explicitly requested.
---

# Skill: multi-agent-orchestrator

Workflow:
1. Read `AGENTS.md` and `multi_agent/instructions/orchestration.md`.
2. Read `multi_agent/instructions/clean-code-standards.md` before dispatching any coding or code-review lane.
3. Run project context discovery.
4. Parse trailing `xN` with `multi_agent/tools/lib/routing-engine.rule.md`.
5. Allocate agents using `multi_agent/config.md`.
6. Generate dispatch and brief artifacts.
7. Emit live status reporting during execution.
8. Collect handoffs, apply review chain, and enforce clean-code blockers.
9. Update memory, metrics, docs, and score tables at the end of the cycle.

Notes:
- `x10` means the principal core squad.
- `x12` means the full-stack enterprise squad.
- Keep context and skill loading minimal per slot.
