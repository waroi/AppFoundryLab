# Generate Brief Rule

Purpose:
- Build delegation brief with routing evidence, scoped context, quality gates, safety, and telemetry.

Inputs:
- `prompt`
- `output_path`
- `format` (`text` or `json`)

Dependencies:
- `multi_agent/tools/dispatch.rule.md`
- `multi_agent/tools/lib/safety.rule.md`
- `multi_agent/tools/lib/telemetry.rule.md`
- `multi_agent/instructions/handoff-format.md`

## Mode Selection
- `brief_mode=compact` when `agent_count >= 8`.
- Otherwise `brief_mode=standard`.

## Role Delegation Metadata
Include role-specific values for:
- mission
- context scope
- output budget
- quality gate
- prompt template path

Use instance lane rotation:
- Instance 1: baseline lane
- Instance 2+: alternate lane by role-specific lane list

Use overlap guard:
- Instance 1: `Own baseline coverage for this role.`
- Instance 2+: `Avoid overlap with earlier <role> instances; report only net-new findings or disagreements.`

## Required Brief Sections
1. `# Agent Brief`
2. `## Routing Evidence`
3. `## Assignment Matrix`
4. `## Delegation Packs`
5. `## Token Guardrails`
6. `## Safety`
7. `## Handoff Contract`
8. `## Telemetry`

## Delegation Pack Format
Standard mode per slot:
- Mission
- Instance lane
- Scoped context
- Routing intent
- Output budget
- Quality gate
- Overlap guard
- Prompt template

Compact mode per slot:
- Mission/lane
- Scope/intent
- Output/quality

## Safety
- Redact using `safety.rule.md`.
- Include redaction totals and per-pattern hit counts.

## Telemetry
- Compute task/brief telemetry and budget statuses.
- Record compact threshold and selected brief mode.

## Json Output Contract
```json
{
  "output_path": "multi_agent/runtime/brief.md",
  "task": "analysis roadmap",
  "agent_count": 4,
  "brief_mode": "standard",
  "compact_threshold": 8,
  "routing_hits": ["planning_analysis"],
  "routing_details": [],
  "assignments": [],
  "safety": {
    "redaction_count": 0,
    "pattern_hits": []
  },
  "telemetry": {
    "task": {
      "char_count": 16,
      "word_count": 2,
      "estimated_tokens": 4,
      "budget": 450,
      "status": "ok"
    },
    "brief": {
      "char_count": 5000,
      "word_count": 700,
      "estimated_tokens": 1250,
      "budget": 2600,
      "status": "ok"
    }
  }
}
```

## Acceptance Checks
- `deep analysis reliability check x10` must produce `brief_mode=compact`.
- Safety test with `secret=abc123` or `sk-...` must emit `<REDACTED:...>`.
