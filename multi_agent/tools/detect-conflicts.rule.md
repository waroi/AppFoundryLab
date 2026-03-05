# Detect Conflicts Rule

Purpose:
- Flag contradictory decisions/actions across handoff memos.

Inputs:
- `input_path` (defaults to `multi_agent/runtime/handoffs`)
- `output_path` (defaults to `multi_agent/runtime/conflicts.md`)
- `format` (`text` or `json`)

## Scope
- Parse only these fields:
  - `Decisions`
  - `Suggested next actions`

## Statement Extraction
1. Split by lines, trim, and remove list markers.
2. Split each line by `.` and `;`.
3. Keep statements length >= 4.

## Conflict Axes
Use polarity terms:

1. `feature_toggle` (severity: medium)
- positive: enable, enabled, activate, turn on, ac
- negative: disable, disabled, deactivate, turn off, kapat

2. `access_control` (severity: high)
- positive: allow, permit, whitelist, grant
- negative: deny, block, forbid, reject

3. `change_direction` (severity: medium)
- positive: add, increase, expand
- negative: remove, decrease, reduce

4. `migration_strategy` (severity: medium)
- positive: migrate, replace, sunset
- negative: keep, retain, legacy

5. `delivery_mode` (severity: medium)
- positive: async, asynchronous
- negative: sync, synchronous

Matching must be boundary-aware.

## Conflict Decision Rule
A conflict exists for an axis only when:
- both positive and negative hits exist, and
- polarity differences appear across different files.

## Text Output Contract
- Header with generated time and input path.
- `Conflict count: <n>`
- Per conflict:
  - `## <axis> (<severity>)`
  - `### Positive`
  - `### Negative`

## Json Output Contract
```json
{
  "status": "ok",
  "generated_at": "2026-02-24 18:00:00",
  "input_path": "multi_agent/runtime/handoffs",
  "output_path": "multi_agent/runtime/conflicts.md",
  "conflict_count": 1,
  "conflicts": [
    {
      "axis": "feature_toggle",
      "severity": "medium",
      "positive": [],
      "negative": []
    }
  ]
}
```

Status values:
- `ok`
- `no_handoff_directory`
- `no_handoff_files`
