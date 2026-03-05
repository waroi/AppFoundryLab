# Generate Brief Rule

Purpose:
- Build delegation briefs with routing evidence, pod-aware context, skill bundles, quality gates, model rationale, safety, release gates, and telemetry.

Required Brief Sections:
1. `# Agent Brief`
2. `## Routing Evidence`
3. `## Team Topology`
4. `## Assignment Matrix`
5. `## Live Reporting Matrix`
6. `## Delegation Packs`
7. `## Release Gates`
8. `## Token Guardrails`
9. `## Safety`
10. `## Handoff Contract`
11. `## Telemetry`

Delegation metadata per slot:
- mission
- pod
- scoped context
- evidence anchors
- routing intent
- output budget
- quality gate
- model rationale
- overlap guard
- skill bundle
- engineering constraints
- reporting target
- data_sensitivity
- data_handling_guidance
- release_relevance

Rules:
- `brief_mode=compact` when `agent_count >= 8`.
- Group packs by pod for `x10+`.
- `x12` must show dedicated frontend and backend lanes.
- `x13-x14` must call out optional specialist selection reasoning explicitly.
- Technical slots must list `clean-code` and the applicable principle constraints in the pack.
- Project context must include AGENTS-derived platform rules in compact form when they affect execution.
- Restricted content should be summarized, not copied, whenever possible.
