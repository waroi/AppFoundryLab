# Detect Conflicts Rule

Purpose:
- Flag contradictory decisions, risks, or next actions across handoff memos.

Inputs:
- `input_path` (defaults to `multi_agent/runtime/handoffs`)
- `output_path` (defaults to `multi_agent/runtime/conflicts.md`)
- `format` (`text` or `json`)

## Scope
Parse these sections when present:
- `Decisions`
- `Suggested next actions`
- `Risks`
- `Evidence`

## Statement Extraction
1. Split by lines, trim, and remove list markers.
2. Split each line by `.` and `;`.
3. Keep statements with length >= 4.
4. If a required section is missing, treat it as `None.` rather than failing the whole parse.

## Conflict Axes
1. `feature_toggle` (severity: medium)
2. `access_control` (severity: high)
3. `change_direction` (severity: medium)
4. `migration_strategy` (severity: medium)
5. `delivery_mode` (severity: medium)
6. `rollback_strategy` (severity: high)
7. `ownership_model` (severity: medium)
8. `security_posture` (severity: high)
9. `data_exposure` (severity: high)

Matching must be boundary-aware.

## Polarity Heuristic
Use explicit polarity buckets per axis.
- Positive examples: `enable`, `allow`, `migrate`, `ship`, `keep`, `expand`, `public`, `required`, `centralize`.
- Negative examples: `disable`, `deny`, `avoid`, `rollback`, `block`, `remove`, `private`, `optional`, `decentralize`.
- Axis-specific terms may refine these buckets when obvious from the section text.

## Conflict Decision Rule
A conflict exists for an axis only when:
- both positive and negative hits exist, and
- polarity differences appear across different files, and
- the statements concern the same axis rather than unrelated wording overlap.

## Text Output Contract
- Header with generated time and input path.
- `Conflict count: <n>`
- Per conflict:
  - `## <axis> (<severity>)`
  - `### Positive`
  - `### Negative`

## JSON Output Contract
```json
{
  "status": "ok",
  "generated_at": "2026-03-05 22:00:00",
  "input_path": "multi_agent/runtime/handoffs",
  "output_path": "multi_agent/runtime/conflicts.md",
  "conflict_count": 1,
  "conflicts": [
    {
      "axis": "rollback_strategy",
      "severity": "high",
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
