# Dispatch Rule

Purpose:
- Produce deterministic assignment and routing evidence for a task prompt.

Inputs:
- `prompt` (string; may contain trailing `xN`/`XN`).
- `format` (`text` or `json`).

Dependencies:
- `multi_agent/config.md`
- `multi_agent/tools/lib/routing-engine.rule.md`

## Procedure
1. Parse prompt envelope.
2. Compute routing hits and ranked routing details.
3. Compute role assignments.
4. Emit dispatch artifact in selected format.

## Text Output Contract
- `Task: <task>`
- `Agents: <count>`
- `Routing: <group1, group2 ...>` or `Routing: default`
- `Assignments:` numbered rows `<slot>. <instance_name> [<source>] (<model>)`

## Json Output Contract
Use this exact schema:
```json
{
  "task": "analysis roadmap",
  "agent_count": 4,
  "routing_hits": ["planning_analysis"],
  "routing_details": [
    {
      "name": "planning_analysis",
      "match_count": 2,
      "matched_keywords": ["analysis", "roadmap"],
      "priority_roles": ["research_analyst", "qa_guardian"],
      "score": 4.16
    }
  ],
  "assignments": [
    {
      "slot": 1,
      "role_key": "principal_architect",
      "instance_name": "principal_architect",
      "model": "gpt-5.3-codex",
      "source": "primary"
    }
  ]
}
```

## Acceptance Checks
- `analysis roadmap x4` => `agent_count=4`, routing includes `planning_analysis`.
- `Refactor API X5` => `agent_count=5`.
- `Create debug dashboard x3` must not route `quality` via `bug`.
