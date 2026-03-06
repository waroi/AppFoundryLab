# Active Plan

Last updated: 2026-03-06
Owner: team_lead

## Objective
Keep the repository in a no-open-phase state after closing the final `PROGRESS.md` phases and syncing code, docs, tests, and governance evidence.

## Phase 1: Runtime Knob Transparency
- Owner: backend_engineer
- Status: completed
- Goal: expose request logging trusted proxy CIDRs and logger timing knobs through runtime config, admin diagnostics, and operator docs
- Blockers: none
- Success checks:
  - `GET /api/v1/admin/runtime-config` publishes request logging and logger timing summaries
  - admin diagnostics renders the runtime knob panel
  - EN/TR docs describe the same runtime contract

## Phase 2: Browser Coverage Depth
- Owner: frontend_engineer
- Status: completed
- Goal: deepen operator-surface browser coverage across keyboard flow, degraded-state diagnostics, and runtime-knob visibility
- Blockers: none
- Success checks:
  - mock-backed Playwright covers keyboard/focus flow
  - degraded admin diagnostics and runtime knob fallbacks are asserted
  - live-stack smoke verifies runtime-knob visibility on the real stack

## Phase 3: Live Smoke Cost Governance
- Owner: delivery_governor
- Status: completed
- Goal: keep `e2e:live` as nightly or on-demand release evidence with a single policy story across docs and automation
- Blockers: none
- Success checks:
  - `RUN_LIVE_STACK_BROWSER_SMOKE` remains the explicit switch for the live lane
  - branch protection docs exclude live smoke from merge blockers
  - release policy, checklist, and workflow matrix all describe the same governance decision

## Phase 4: Documentation, Memory, and Metrics Sync
- Owner: qa_guardian
- Status: completed
- Goal: close the loop with canonical backlog repair, drift-proof docs, and end-of-cycle memory plus metrics updates
- Blockers: none
- Success checks:
  - `PROGRESS.md` exists once and matches `docs/gelistirmePlanı.md`
  - release and incident docs reflect the completed phases
  - session memory and metrics capture the closure state

## Now
- Preserve the closed-backlog state until fresh analysis creates a new repo-owned phase.

## Next
- Re-run the validated gates whenever runtime-config, admin diagnostics, or governance docs change.
- Open a new `PROGRESS.md` phase only with fresh analysis, test evidence, and synced documentation.

## Open Risks
- `e2e:live` still depends on host-owned Docker capacity and is intentionally outside branch protection.
- The worktree contains unrelated changes outside this cycle that need separate ownership before any release action.
