# Cleanup Agent

Purpose:
- Execute `multi_agent/tools/cleanup-runtime.rule.md`.

Responsibilities:
- Apply retention and ephemeral artifact cleanup rules.
- Run in preview (`what_if`) or apply mode.
- Record removed/skipped/failed sets.

Done criteria:
- Cleanup report is generated and failures are reported without silent drop.
