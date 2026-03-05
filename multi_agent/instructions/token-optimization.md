# Token Optimization Guidelines

- Share only the minimum context each agent needs to be effective.
- Prefer file paths and short excerpts over large dumps.
- Use concise bullets and avoid restating the user prompt.
- Deduplicate work: Team Lead assigns non-overlapping tasks.
- Summaries should target 5 to 10 bullets unless more is essential.
- Keep delegation prompts small: task, scoped context, constraints, output contract.
- Enforce per-role output budgets (word/bullet caps) in generated briefs.
- Use routing evidence to justify why a role is delegated; avoid unnecessary agents.
- Keep handoff memos compact and avoid repeating the same point across sections.
- Reuse generated artifacts under `multi_agent/runtime/` (`brief.md`, `summary.md`) instead of rebuilding context.
- Check `## Telemetry` sections in brief/summary outputs and treat `status=over_budget` as a prompt to split scope.
- Check `run.json -> telemetry.scorecard/trend` for run-level budget drift before increasing `xN`.
- Prefer `multi_agent/tools/orchestrate-run.rule.md` for consistent scoped runs and cleaner token budgeting across iterations.
- Use `multi_agent/tools/validate-setup.rule.md` after structural changes to catch token-cost regressions early.
