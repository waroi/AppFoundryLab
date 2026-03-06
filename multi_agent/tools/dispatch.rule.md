# Dispatch Rule

Purpose:
- Produce deterministic assignment, model routing, skill bundle, routing evidence, and governance classification for a task prompt.

Inputs:
- `prompt`
- `format` (`text` or `json`)

Dependencies:
- `multi_agent/config.md`
- `multi_agent/agents/*.agent.md`
- `multi_agent/tools/lib/routing-engine.rule.md`
- `multi_agent/instructions/model-selection.md`
- `multi_agent/instructions/skill-loading.md`

## Procedure
1. Parse prompt envelope.
2. Compute routing hits and ranked routing details.
3. Compute agent assignments with dedupe preserved in first-hit order.
4. Resolve optional specialist behavior for `x13-x14` using `allocation.optional_specialists`.
5. Resolve model escalations where applicable.
6. Classify `release_oriented` and `data_sensitivity` signals.
7. Attach default skill bundles from the agent catalog.
8. Emit dispatch artifact in selected format.

## Artifact Requirements
- Include assignment source per slot.
- Include `release_oriented` classification.
- Include `data_sensitivity` classification.
- Include optional specialist selection reason when `x13-x14` is used.
- Include resolved skill bundle per slot.
- Include dedupe-safe unique slot list.

## Acceptance Checks
- `deep analysis roadmap x4` => routing includes `planning_analysis`.
- `Enterprise modernization x10` => 10 unique assignments.
- `Enterprise modernization x12` => includes `frontend_engineer` and `backend_engineer`.
- `API contract migration x13` => `api_integration_engineer` is added as the optional specialist.
- `Operator handbook refresh x13` => `documentation_analyst` is added as the optional specialist.
- `API contract migration x14` => includes `api_integration_engineer` and `documentation_analyst`.
- Coding and code-review assignments include `clean-code` in the resolved skill bundle.
- `Production rollback review x12` => `release_oriented=true`.
- `Customer data contract review x12` => `data_sensitivity=restricted`.
