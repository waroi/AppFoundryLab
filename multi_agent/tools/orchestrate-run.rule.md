# Orchestrate Run Rule

Purpose:
- Provide one deterministic lifecycle for dispatch -> brief -> handoff scaffold -> summary -> conflicts -> run metadata.

Inputs:
- `prompt`
- `run_id` (optional)
- `runtime_root` (default `multi_agent/runtime/runs`)
- `format` (`text` or `json`)

Dependencies:
- `dispatch.rule.md`
- `generate-brief.rule.md`
- `summarize-handoffs.rule.md`
- `detect-conflicts.rule.md`
- `multi_agent/tools/lib/telemetry.rule.md`

## Prompt Fallback
- If prompt is empty, use `Untitled task x1`.

## Run ID Rule
- If not provided: `<yyyyMMdd-HHmmss>-<8-char-guid>`.
- Run directory must be unique.

## Required Run Artifacts
Under `<run_dir>/`:
- `dispatch.json`
- `brief.md`
- `brief.json`
- `handoffs/`
- `summary.md`
- `summary.json`
- `conflicts.md`
- `conflicts.json`
- `run.json`

## Handoff Scaffold Rule
- Create one memo file per assigned slot:
  - `<slot-2-digit>-<instance_name>.md`
- Each file includes `handoff-format` fields with placeholders.

## Metadata Rule (`run.json`)
Must include:
- run_id, created_at, prompt, task, agent_count, routing_hits
- `paths.*` references
- embedded assignment/brief/summary/conflicts payloads
- telemetry scorecard and trend

Scorecard:
- task_estimated_tokens
- brief_estimated_tokens
- summary_estimated_tokens
- total_estimated_tokens
- conflict_count
- budget_status (`ok|over_budget|no_budget`)

Trend:
- previous_run_id
- previous_total_estimated_tokens
- token_delta
- token_trend (`up|down|flat`)

## Json Return Contract
```json
{
  "status": "ok",
  "run_id": "20260224-180114-45e79732",
  "run_dir": "D:/w/CodexA2A/multi_agent/runtime/runs/20260224-180114-45e79732",
  "handoffs_dir": "D:/w/CodexA2A/multi_agent/runtime/runs/20260224-180114-45e79732/handoffs",
  "conflicts_json": "D:/w/CodexA2A/multi_agent/runtime/runs/20260224-180114-45e79732/conflicts.json",
  "metadata_json": "D:/w/CodexA2A/multi_agent/runtime/runs/20260224-180114-45e79732/run.json"
}
```
