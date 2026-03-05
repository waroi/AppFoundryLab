# Dispatch Agent

Purpose:
- Execute `multi_agent/tools/dispatch.rule.md` and publish assignment artifacts.

Responsibilities:
- Parse prompt envelope.
- Compute routing evidence and assignments.
- Produce text and json dispatch outputs.

Done criteria:
- Output contains deterministic `task`, `agent_count`, `routing_hits`, and `assignments`.
