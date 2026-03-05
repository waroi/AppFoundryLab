# Run Unit Tests Rule (Markdown Regression Set)

Purpose:
- Replace executable unit tests with deterministic scenario checks.

Execution:
- Validate each case manually or via another rule-driven agent.
- Report as `pass|fail|skip` in `multi_agent/runtime/unit-test-report.md`.

## Required Cases (10)
1. Parse lowercase suffix:
   - Input: `Refactor API x4`
   - Expect: `agent_count=4`, task `Refactor API`
2. Parse uppercase suffix:
   - Input: `Refactor API X5`
   - Expect: `agent_count=5`
3. Normalize invalid counts:
   - Input: `Refactor API x0`
   - Expect: `agent_count=1`
4. Word boundary protection:
   - Input task `debug dashboard`, keyword `bug`, mode `word`
   - Expect: no match
5. Substring mode behavior:
   - Input task `debug dashboard`, keyword `bug`, mode `substring`
   - Expect: match
6. Planning routing:
   - Input: `deep analysis roadmap x4`
   - Expect routing contains `planning_analysis`
7. Optimization routing:
   - Input: `deep optimization improvement x4`
   - Expect routing contains `planning_analysis`
8. Security ranking:
   - Input: `auth security oauth jwt test regression x4`
   - Expect first routing hit is `security`
9. x1 combined role:
   - Input: `basic task x1`
   - Expect single assignment `team_lead_architect_combined`
10. High-N duplicates:
   - Input: `backend hardening x8`
   - Expect at least one instance name suffixed `_2`

## Json Report Contract
```json
{
  "status": "passed",
  "total": 10,
  "passed": 10,
  "failed": 0,
  "skipped": 0,
  "cases": [
    { "id": 1, "name": "Parse lowercase suffix", "status": "pass" }
  ]
}
```
