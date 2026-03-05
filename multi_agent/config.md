# Multi-Agent Configuration (Markdown Canonical Spec)

This file replaces `config.json` and is now the canonical source for allocation, routing, and telemetry budgets.

## Canonical Spec
```yaml
version: 2
parameter: xN
defaults:
  agent_count: 1
  max_agent_count: 12
telemetry:
  token_estimator: chars_div_4
  budgets:
    task_max_estimated_tokens: 450
    brief_max_estimated_tokens: 2600
    memo_max_estimated_tokens: 700
    summary_max_estimated_tokens: 2600
models:
  gpt-5.3-codex:
    purpose: Architecture, critical code, review, orchestration
  gpt-5.2-instant:
    purpose: Reading, research, analysis, visuals, quality checks
roles:
  principal_architect:
    model: gpt-5.3-codex
    focus: Architecture and critical code decisions
  team_lead:
    model: gpt-5.3-codex
    focus: Orchestration, review, integration
  research_analyst:
    model: gpt-5.2-instant
    focus: Reading, research, analysis
  visual_researcher:
    model: gpt-5.2-instant
    focus: Visual research, UI references, quick analysis
  qa_guardian:
    model: gpt-5.2-instant
    focus: Test strategy, regression checks, release confidence
  security_reviewer:
    model: gpt-5.2-instant
    focus: Security posture, auth risks, threat-aware review
  team_lead_architect_combined:
    model: gpt-5.3-codex
    focus: Combined responsibilities of Team Lead and Principal Architect
allocation:
  primary_roles:
    - principal_architect
    - team_lead
  fallback_cycle:
    - research_analyst
    - visual_researcher
    - qa_guardian
    - security_reviewer
  overflow_cycle:
    - research_analyst
    - visual_researcher
    - qa_guardian
    - security_reviewer
    - principal_architect
    - team_lead
  max_instances_per_role:
    default: 2
    team_lead_architect_combined: 1
routing:
  keyword_groups:
    - name: planning_analysis
      weight: 1.3
      keywords:
        - { term: analysis, weight: 1.4 }
        - { term: deep analysis, weight: 1.7 }
        - { term: analiz, weight: 1.4 }
        - { term: derinlemesine, weight: 1.6 }
        - { term: plan, weight: 1.5 }
        - { term: roadmap, weight: 1.8 }
        - { term: strategy, weight: 1.6 }
        - { term: gelistirme, weight: 1.3 }
        - { term: iyilestirme, weight: 1.3 }
        - { term: optimizasyon, weight: 1.4 }
        - { term: optimize, weight: 1.4 }
        - { term: optimization, weight: 1.5 }
        - { term: improve, weight: 1.3 }
        - { term: improvement, weight: 1.4 }
        - { term: mukemmel, weight: 1.5 }
        - { term: perfect, weight: 1.4 }
        - { term: perfection, weight: 1.5 }
        - { term: feature, weight: 1.2 }
        - { term: ozellik, weight: 1.2 }
      priority_roles:
        - research_analyst
        - qa_guardian
    - name: orchestration_ops
      weight: 1.2
      keywords:
        - { term: workflow, weight: 1.4 }
        - { term: orchestration, weight: 1.6 }
        - { term: orchestrate, weight: 1.5 }
        - { term: handoff, weight: 1.3 }
        - { term: session, weight: 1.3 }
        - { term: runtime, weight: 1.2 }
        - { term: automation, weight: 1.5 }
        - { term: pipeline, weight: 1.4 }
        - { term: ci, weight: 1.2 }
        - { term: kaldigimiz yerden, weight: 1.3 }
      priority_roles:
        - research_analyst
        - qa_guardian
    - name: frontend
      keywords:
        - frontend
        - ui
        - ux
        - design
        - visual
        - css
        - layout
        - tasarim
        - gorsel
      priority_roles:
        - visual_researcher
        - qa_guardian
    - name: quality
      keywords:
        - test
        - qa
        - regression
        - bug
        - fix
        - refactor
        - hardening
      priority_roles:
        - qa_guardian
        - research_analyst
    - name: security
      keywords:
        - security
        - auth
        - jwt
        - oauth
        - encryption
        - threat
        - vulnerability
        - guvenlik
      priority_roles:
        - security_reviewer
        - research_analyst
    - name: architecture
      keywords:
        - architecture
        - mimari
        - domain
        - clean architecture
        - solid
        - ddd
      priority_roles:
        - research_analyst
```

## Interpretation Rules
- Default keyword `match_mode` is `word` (boundary-aware exact token/phrase match).
- Optional keyword objects may set `match_mode` (`word` or `substring`) and `weight`.
- Group score = `sum(matched keyword weights) * group.weight` (default group weight is `1.0`).
- Sort routing hits by `score desc`, `match_count desc`, then config order asc.
- Prompt suffix parsing is case-insensitive: both `xN` and `XN` are valid.
