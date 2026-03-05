# Telemetry Rule

Purpose:
- Provide stable token/cost estimation and budget status signals.

## Text Metrics
For any text:
- `char_count = text.length`
- `word_count = count(split(normalize_whitespace(text), " "))`
- `estimated_tokens = ceil(char_count / 4)`

## Budget Resolution
Budget keys from `multi_agent/config.md`:
- `task_max_estimated_tokens` (default `500`)
- `brief_max_estimated_tokens` (default `3600`)
- `memo_max_estimated_tokens` (default `750`)
- `summary_max_estimated_tokens` (default `2800`)
- Budget profiles:
  - `critical` (default `340`)
  - `execution` (default `280`)
  - `support` (default `180`)
  - `documentation` (default `160`)

Rule:
- `multi_agent/config.md` is the canonical source for numeric telemetry values.
- Local fallback numbers must stay aligned with the current config.

## Budget Status
- `no_budget` when budget <= 0
- `ok` when used <= budget
- `over_budget` when used > budget

## Reporting
- Brief and summary must include `## Telemetry`.
- Briefs should expose role or pod budget-profile mapping when compact mode hides verbose detail.
- Run metadata must include:
  - `scorecard.task_estimated_tokens`
  - `scorecard.brief_estimated_tokens`
  - `scorecard.summary_estimated_tokens`
  - `scorecard.total_estimated_tokens`
  - `scorecard.budget_status`
  - `trend.previous_run_id`
  - `trend.previous_total_estimated_tokens`
  - `trend.token_delta`
  - `trend.token_trend` (`up|down|flat`)
