# Routing Engine Rule

Purpose:
- Preserve deterministic prompt parsing, routing, named squad activation, and assignment behavior without scripts.

Inputs:
- Prompt text.
- `multi_agent/config.md`.

## A. Parse Prompt Envelope
1. Trim prompt text.
2. Match suffix regex `(?i)(?:^|\s)x(\d+)\s*$`.
3. If matched:
   - `agent_count = int(capture)`
   - `task = prompt_without_suffix.trim()`
4. Else:
   - `agent_count = defaults.agent_count`
   - `task = trimmed_prompt`
5. Normalize `agent_count`:
   - if `<1`: set to `1`
   - if `>defaults.max_agent_count`: cap
6. If task empty: set task to `Untitled task`.

## B. Evaluate Routing Hits
1. Iterate `routing.keyword_groups` in declaration order.
2. For each keyword:
   - if string: `term=<keyword>`, `match_mode=word`, `weight=1.0`
   - if object: read `term`, optional `match_mode`, optional `weight`
3. Match behavior:
   - `word`: boundary-aware token matching
   - `substring`: case-insensitive substring
4. Group score:
   - `weighted_score = sum(matched keyword weights) * group.weight`
5. Sort hits by score descending, match_count descending, then group order ascending.

## C. Allocate Agents
1. Initialize empty assignments.
2. If `agent_count == 1`:
   - assign `team_lead_architect_combined`, source `single_mode`
3. If `agent_count == defaults.core_squad_agent_count`:
   - assign `allocation.named_squads.enterprise_x10_core.agents`, source `named_squad`
4. If `agent_count == defaults.full_stack_squad_agent_count`:
   - assign `allocation.named_squads.enterprise_x12_full_stack.agents`, source `named_squad`
5. If `2 <= agent_count < defaults.core_squad_agent_count`:
   - assign `allocation.primary_agents`, source `primary`
   - insert priority agents from ranked routing hits, source `routing`
   - dedupe while preserving first appearance order
   - fill with `allocation.fallback_cycle`, source `fallback`
6. If `agent_count == 11`:
   - seed `enterprise_x10_core`
   - add the top-ranked missing agent from `allocation.expansion_cycle`, source `expansion`
7. If `agent_count == 13`:
   - seed `enterprise_x12_full_stack`
   - add the highest-ranked missing optional specialist matched by `allocation.optional_specialists.routing_map`
   - if none match, use the first missing agent in `allocation.optional_specialists.fallback_cycle`
8. If `agent_count == 14`:
   - seed `enterprise_x12_full_stack`
   - add all missing agents from `allocation.optional_specialists.fallback_cycle` in order
9. Final assignments must be unique.

## D. Model Resolution
- Resolve base model from `model_selection.default_agent_models`.
- Apply conditional escalation per agent when prompt keywords match.
- Record both `base_model` and `resolved_model` when they differ.

## E. Governance Resolution
- Resolve `release_oriented` using `policies.governance.release_oriented_keywords`.
- Resolve `data_sensitivity=restricted` when prompt text matches any `delegation.data_sensitivity.restricted_keywords`.
- Otherwise resolve `data_sensitivity` to `delegation.data_sensitivity.default`.

## F. Dispatch Artifact Contract
Text mode must include:
- Task
- Agents
- Routing summary
- Assignment list with slot, source, pod, skill bundle, and resolved model
- `release_oriented`
- `data_sensitivity`
