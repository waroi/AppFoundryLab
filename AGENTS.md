# Codex Multi-Agent Orchestration (Local)

Purpose:
- Provide a deterministic multi-agent workflow for every user task in this workspace.
- Use trailing `xN` in user prompts to allocate agent count.
- Keep token usage low with strict scoped-context handoffs.

Activation:
- Prompt ends with `xN` or `XN` (example: `Refactor API x4`) -> allocate exactly N agents.
- Missing `xN` -> default to one combined role (Team Lead + Principal Architect).
- Invalid or `<1` -> normalize to 1.

Core roles:
- Principal Architect (GPT-5.3 Codex): architecture and critical code.
- Team Lead (GPT-5.3 Codex): orchestration, review, integration.
- Research Analyst (GPT-5.2 Instant): reading, research, analysis.
- Visual Researcher (GPT-5.2 Instant): UI/UX and visual analysis.
- QA Guardian (GPT-5.2 Instant): test strategy and regression gates.
- Security Reviewer (GPT-5.2 Instant): threat-focused security checks.

Allocation rules:
- `N == 1`: assign `team_lead_architect_combined`.
- `N >= 2`: assign primary roles first (`principal_architect`, `team_lead`).
- Insert routing-priority roles based on boundary-aware keyword matching and routing scores in `multi_agent/config.md`.
- Fill remaining slots with `allocation.fallback_cycle`.
- When N exceeds unique roles, duplicate with numeric suffix (example: `research_analyst_2`).

Execution rules:
- Team Lead delegates scoped tasks and integrates one coherent final output.
- Principal Architect enforces SOLID, KISS, YAGNI and clean-code decisions.
- Analyst roles return concise memos in `multi_agent/instructions/handoff-format.md`.
- Team Lead requests revision on conflicts, ambiguity, or missing risk analysis.

Token optimization:
- Share only minimum relevant context per agent.
- Prefer summaries and precise file references over large raw dumps.
- Reuse generated brief + summary artifacts in `multi_agent/runtime/`.
- Track telemetry/budget status in generated artifacts and split tasks when over budget.

References:
- Configuration: `multi_agent/config.md`
- Orchestration guide: `multi_agent/instructions/orchestration.md`
- Operator playbook: `multi_agent/instructions/operator-playbook.md`
- Roles: `multi_agent/roles/`
- Prompts: `multi_agent/prompts/`
- Rule tools: `multi_agent/tools/*.rule.md`
