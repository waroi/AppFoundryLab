---
last-updated: 2026-03-06
tracking-version: 4
---

# Token Usage Metrics

Tracks model-tier usage and token-governance assumptions.

## Measured Skill Weights
Measured with the repository's `chars_div_4` estimator from `multi_agent/config.md`.

| Skill | Approx Tokens |
| --- | --- |
| clean-code | 835 |
| frontend-development | 510 |
| backend-development | 444 |
| backend-security | 392 |
| api-integration | 354 |
| code-architecture | 337 |
| testing-standards | 319 |
| documentation-operations | 316 |
| implementation | 312 |
| code-review | 286 |
| multi-agent-orchestrator | 270 |
| analysis | 204 |

## Squad Cost View
| Scenario | Approx Tokens | Notes |
| --- | --- | --- |
| Naive `x12`: every slot loads every coding skill | 45468 | anti-pattern; high redundancy |
| Current optimized `x12` bundle sum | 13968 | lane-aware bundles with selective `clean-code` usage |
| Routed `x13` with API specialist | 14322 | `x12` + one optional specialist |
| Routed `x13` with docs specialist | 14284 | `x12` + one cheaper docs lane |
| Fixed `x14` full specialist expansion | 14638 | both optional specialists active |
| Shared-context theoretical floor | 4578 | possible only with a future executor that deduplicates shared context |

## Observations
- The current lane-aware `x12` bundle model reduces estimated skill-load cost by about `69%` versus naive all-skills-per-slot loading.
- `x13` lets the system pay for only one optional specialist when the prompt does not justify both lanes.
- `clean-code` remains the heaviest single skill, so it stays limited to technical lanes.
- Restricted-context handling also controls token growth because sensitive evidence is summarized instead of copied.
- A repo-wide phase-closure run can justify `x14`, but the highest-signal evidence still clusters around architecture, QA, frontend, backend, and documentation-heavy integration work.

## Guidance
- Use `x10` when frontend and backend dedicated coding lanes are not both needed.
- Use `x13` when only one optional specialist is strongly signaled.
- Use `x14` when code, QA, governance, and documentation all need concurrent ownership, as in repo-wide phase closure or release-readiness sync work.
- Split the work if a slot would require more than three skills or a large raw context dump.
