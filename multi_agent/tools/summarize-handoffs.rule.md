# Summarize Handoffs Rule

Purpose:
- Summarize handoff memos into snapshot, decisions, risks, dependencies, open questions, action queue, safety, telemetry, documentation delta, and metrics delta.

Required summary sections:
1. `# Handoff Summary`
2. `## Snapshot`
3. `## Agent Scoreboard`
4. `## Decision Ledger`
5. `## Risk Register`
6. `## Dependency Watchlist`
7. `## Open Questions`
8. `## Action Queue`
9. `## Documentation Delta`
10. `## Metrics Delta`
11. `## Safety`
12. `## Telemetry`

Rules:
- Parse `Agent`, `Skill bundle`, `Status`, `Success score`, `Risks`, and `Suggested next actions` from each handoff.
- If a required field is missing, normalize it to `Unknown` or `None.` instead of aborting the summary.
- Redact sensitive content before writing.
- Carry unresolved blockers into the final scoreboard.
- If unresolved guard blockers remain for a release-oriented run, surface final run state as `blocked`.
