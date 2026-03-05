# Summarize Handoffs Rule

Purpose:
- Summarize handoff memos into snapshot, risks, open questions, action queue, safety, and telemetry.

Inputs:
- `input_path` (defaults to `multi_agent/runtime/handoffs`)
- `output_path` (defaults to `multi_agent/runtime/summary.md`)
- `format` (`text` or `json`)

Dependencies:
- `multi_agent/instructions/handoff-format.md`
- `multi_agent/tools/lib/safety.rule.md`
- `multi_agent/tools/lib/telemetry.rule.md`

## Field Parsing
- Parse top-level fields using `- Field:` anchors.
- Support multiline values until next top-level field.
- Required extraction targets:
  - Task
  - Decisions
  - Risks
  - Open questions
  - Suggested next actions
  - Confidence

## Normalization
- Collapse empty lines and repeated whitespace.
- Convert multiline field bodies into compact single-line summaries for table cells.

## Risk Severity Rules
- `[high]` => high
- `[medium]` => medium
- `[low]` => low
- else => unspecified
- Sort risk register by severity rank then filename.

## Required Summary Sections
1. `# Handoff Summary`
2. `## Snapshot` (markdown table)
3. `## Risk Register`
4. `## Open Questions`
5. `## Action Queue`
6. `## Safety`
7. `## Telemetry`

## Safety and Telemetry
- Redact sensitive content before writing.
- Emit summary telemetry and memo over-budget list.

## Json Output Contract
```json
{
  "status": "ok",
  "generated_at": "2026-02-24 18:00:00",
  "input_path": "multi_agent/runtime/handoffs",
  "output_path": "multi_agent/runtime/summary.md",
  "files": [],
  "risks": [],
  "open_questions": [],
  "action_queue": [],
  "safety": {
    "redaction_count": 0,
    "pattern_hits": []
  },
  "telemetry": {
    "summary": {
      "char_count": 0,
      "word_count": 0,
      "estimated_tokens": 0,
      "budget": 2600,
      "status": "ok"
    },
    "memo_budget": 700,
    "over_budget_memos": []
  }
}
```

Status values:
- `ok`
- `no_handoff_directory`
- `no_handoff_files`
