# Routing Engine Rule

Purpose:
- Preserve deterministic prompt parsing, routing, and assignment behavior without scripts.

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
   - if object: read `term` or `keyword`, optional `match_mode`/`mode`, optional `weight`
3. Match behavior:
   - `word`: boundary-aware unicode token matching.
   - `substring`: case-insensitive substring.
4. For each match:
   - append term to matched list
   - add keyword weight to raw score
5. Group score:
   - `weighted_score = raw_score * group.weight` (default group weight `1.0`)
6. Keep group only when at least one keyword matched.
7. Sort hits:
   1. score descending
   2. match_count descending
   3. group order ascending

## C. Allocate Roles
1. Initialize empty assignments and role occurrence map.
2. If `agent_count == 1`:
   - assign `team_lead_architect_combined` as slot 1, source `single_mode`.
3. If `agent_count >= 2`:
   - assign `allocation.primary_roles` in order, source `primary`.
   - insert priority roles from sorted routing hits in order, deduped by role key, source `routing`.
   - fill with `allocation.fallback_cycle`, source `fallback`.
   - if blocked by max instances, continue with `allocation.overflow_cycle`, source `overflow`.
   - if still blocked, disable role limits and continue filling cycles.
4. Instance naming:
   - first occurrence: `role_key`
   - second+ occurrence: `role_key_<n>`
5. Slot numbering starts at 1 and increments by append order.

## D. Role Limits
- Read `allocation.max_instances_per_role`.
- If role-specific limit absent, use `default`.
- If limit <=0: unlimited.

## E. Dispatch Artifact Contract
Text mode must include:
- Task
- Agents
- Routing summary
- Assignment list with slot/source/model

Json mode payload:
```json
{
  "task": "string",
  "agent_count": 4,
  "routing_hits": ["planning_analysis"],
  "routing_details": [
    {
      "name": "planning_analysis",
      "match_count": 2,
      "matched_keywords": ["analysis", "roadmap"],
      "priority_roles": ["research_analyst", "qa_guardian"],
      "score": 4.16
    }
  ],
  "assignments": [
    {
      "slot": 1,
      "role_key": "principal_architect",
      "instance_name": "principal_architect",
      "model": "gpt-5.3-codex",
      "source": "primary"
    }
  ]
}
```
