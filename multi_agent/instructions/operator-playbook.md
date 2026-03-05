# Operator Playbook

Purpose:
- Help the orchestrator choose the right team shape with the least waste.

## Quick Selection
- `x1`: one owner, narrow critical task, low coordination need.
- `x2-x4`: architecture + implementation + research, limited blast radius.
- `x5-x9`: mixed work that needs risk and delivery coverage but not a fixed full squad.
- `x10`: strategy-heavy or governance-heavy work across architecture, risk, delivery, and integration.
- `x12`: true full-stack delivery where dedicated frontend and backend lanes are both required.
- `x13`: `x12` plus one routing-matched optional specialist.
- `x14`: `x12` plus both optional specialists.

## Intent-First Routing Hints
- API contract, schema, SDK, webhook, OpenAPI -> prefer `x13` or `x14` with `api_integration_engineer`.
- Documentation, onboarding, runbook, handbook, ADR -> prefer `x13` or `x14` with `documentation_analyst`.
- Release, rollback, canary, production, outage -> treat as release-oriented and expect guard-lane hard gates.

## Execution Guidance
1. Run project context discovery first.
2. Prefer routing-inserted specialists before increasing expensive model count.
3. Delegate by pods for `x10+`, otherwise by slot.
4. Make `clean-code` mandatory in every coding or code-review lane.
5. Respect `data_sensitivity` and pass restricted context only as redacted evidence anchors.
6. Keep live reporting visible throughout the run using the canonical table format.
7. End every cycle with deep analysis, doc sync, metrics update, and score refresh.
