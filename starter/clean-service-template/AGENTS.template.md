# AGENTS Template (Project Local)

Purpose:
- Reuse the multi-agent method in new projects with minimal edits.

Rules:
- Prompt format supports trailing `xN`.
- `x1`: combined Team Lead + Principal Architect.
- `x2+`: primary roles first, then routing-priority roles, then fallback cycle.
- Team Lead must integrate into a single coherent response.

Project adaptation:
- Keep role focus unchanged.
- Update keyword routing in `multi_agent/config.json` for project vocabulary.
- Keep handoff contract stable unless the team agrees to a revision.
