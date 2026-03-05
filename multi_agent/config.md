# Multi-Agent Configuration (Enterprise Canonical Spec)

This file is the canonical source for allocation, routing, model selection, skill loading, memory, telemetry budgets, and governance defaults.

## Canonical Spec
```yaml
version: 5
parameter: xN
defaults:
  agent_count: 1
  max_agent_count: 14
  compact_brief_threshold: 8
  core_squad_agent_count: 10
  full_stack_squad_agent_count: 12
  max_unique_worker_agents: 14
telemetry:
  token_estimator: chars_div_4
  budgets:
    task_max_estimated_tokens: 500
    brief_max_estimated_tokens: 3600
    memo_max_estimated_tokens: 750
    summary_max_estimated_tokens: 2800
  budget_profiles:
    critical: 340
    execution: 280
    support: 180
    documentation: 160
models:
  gpt-5.4:
    tier: critical
    purpose: Architecture arbitration, irreversible decisions, single-agent critical mode
  gpt-5.3-codex:
    tier: critical_execution
    purpose: Orchestration, critical code, frontend/backend/API implementation, security hardening
  gpt-5.2-instant:
    tier: efficiency
    purpose: Research, QA analysis, UX review, governance, documentation
model_selection:
  default_agent_models:
    principal_architect: gpt-5.4
    team_lead: gpt-5.3-codex
    full_stack_staff_engineer: gpt-5.3-codex
    frontend_engineer: gpt-5.3-codex
    backend_engineer: gpt-5.3-codex
    api_integration_engineer: gpt-5.3-codex
    research_analyst: gpt-5.2-instant
    qa_guardian: gpt-5.2-instant
    security_reviewer: gpt-5.3-codex
    platform_reliability_engineer: gpt-5.2-instant
    product_strategy_analyst: gpt-5.2-instant
    visual_researcher: gpt-5.2-instant
    delivery_governor: gpt-5.2-instant
    documentation_analyst: gpt-5.2-instant
    team_lead_architect_combined: gpt-5.4
  conditional_escalations:
    - agent: platform_reliability_engineer
      escalation_model: gpt-5.3-codex
      when_keywords:
        - performance
        - latency
        - throughput
        - migration
        - observability
        - incident
        - outage
        - scalability
        - performans
        - gecis
        - gozlemlenebilirlik
    - agent: qa_guardian
      escalation_model: gpt-5.3-codex
      when_keywords:
        - release
        - rollback
        - prod
        - production
        - compliance
        - canary
        - deployment
        - yayin
    - agent: documentation_analyst
      escalation_model: gpt-5.3-codex
      when_keywords:
        - migration guide
        - runbook
        - compliance
        - onboarding
        - handbook
agents:
  principal_architect:
    model: gpt-5.4
    budget_profile: critical
    pod: architecture
    focus: Architecture sign-off, core boundaries, irreversible tradeoffs
    default_skills: [clean-code, code-architecture, code-review]
  team_lead:
    model: gpt-5.3-codex
    budget_profile: execution
    pod: integration
    focus: Orchestration, integration, conflict resolution, live reporting
    default_skills: [multi-agent-orchestrator, clean-code, implementation]
  full_stack_staff_engineer:
    model: gpt-5.3-codex
    budget_profile: execution
    pod: architecture
    focus: Critical implementation design across modules, migrations, and integration seams
    default_skills: [clean-code, implementation, code-review]
  frontend_engineer:
    model: gpt-5.3-codex
    budget_profile: execution
    pod: experience
    focus: Frontend architecture, UI implementation, state management, accessibility
    default_skills: [clean-code, frontend-development, testing-standards]
  backend_engineer:
    model: gpt-5.3-codex
    budget_profile: execution
    pod: architecture
    focus: Services, persistence, auth, domain boundaries, migrations
    default_skills: [clean-code, backend-development, testing-standards]
  api_integration_engineer:
    model: gpt-5.3-codex
    budget_profile: execution
    pod: delivery
    focus: API contracts, schema alignment, SDK boundaries, integration risk
    default_skills: [clean-code, api-integration, backend-development]
  research_analyst:
    model: gpt-5.2-instant
    budget_profile: support
    pod: architecture
    focus: Reading, constraints, dependency mapping, factual synthesis
    default_skills: [analysis]
  qa_guardian:
    model: gpt-5.2-instant
    escalation_model: gpt-5.3-codex
    budget_profile: support
    pod: risk
    focus: Regression prevention, test strategy, release confidence
    default_skills: [clean-code, testing-standards, code-review]
  security_reviewer:
    model: gpt-5.3-codex
    budget_profile: execution
    pod: risk
    focus: Auth, secrets, abuse-path, data exposure review
    default_skills: [clean-code, backend-security, code-review]
  platform_reliability_engineer:
    model: gpt-5.2-instant
    escalation_model: gpt-5.3-codex
    budget_profile: support
    pod: risk
    focus: Performance, observability, resilience, rollout safety
    default_skills: [clean-code, analysis, backend-development]
  product_strategy_analyst:
    model: gpt-5.2-instant
    budget_profile: support
    pod: experience
    focus: Outcome framing, acceptance criteria, scope quality
    default_skills: [analysis, documentation-operations]
  visual_researcher:
    model: gpt-5.2-instant
    budget_profile: support
    pod: experience
    focus: UX, accessibility, interface quality, visual consistency
    default_skills: [frontend-development, analysis]
  delivery_governor:
    model: gpt-5.2-instant
    budget_profile: support
    pod: delivery
    focus: Sequencing, ownership, release governance, dependency planning
    default_skills: [documentation-operations, analysis]
  documentation_analyst:
    model: gpt-5.2-instant
    escalation_model: gpt-5.3-codex
    budget_profile: documentation
    pod: delivery
    focus: README, ADR, runbook, changelog, onboarding and operator docs quality
    default_skills: [documentation-operations]
  team_lead_architect_combined:
    model: gpt-5.4
    budget_profile: critical
    pod: integration
    focus: Combined single-agent mode for critical but scoped work
    default_skills: [multi-agent-orchestrator, clean-code, code-architecture]
policies:
  catalog:
    worker_metadata_source_of_truth: multi_agent/agents/*.agent.md
    config_agent_map_role: routing_budget_overlay
    metadata_conflict_rule: agent_file_wins_until_config_is_synced
  engineering:
    mandatory_principles:
      - solid
      - dry
      - kiss
      - yagni
    mandatory_skill_for_coding_lanes: clean-code
    mandatory_instruction: multi_agent/instructions/clean-code-standards.md
  governance:
    release_oriented_keywords:
      - release
      - rollout
      - rollback
      - deployment
      - production
      - prod
      - canary
      - compliance
      - incident
      - outage
      - migration
      - launch
      - yayin
    blocker_authorities:
      - security_reviewer
      - qa_guardian
      - platform_reliability_engineer
    unresolved_guard_blocker_final_state: blocked
allocation:
  primary_agents:
    - principal_architect
    - team_lead
  named_squads:
    enterprise_x10_core:
      trigger_agent_count: 10
      strategy: full_coverage
      agents:
        - principal_architect
        - team_lead
        - full_stack_staff_engineer
        - research_analyst
        - qa_guardian
        - security_reviewer
        - platform_reliability_engineer
        - product_strategy_analyst
        - visual_researcher
        - delivery_governor
    enterprise_x12_full_stack:
      trigger_agent_count: 12
      strategy: full_coverage
      agents:
        - principal_architect
        - team_lead
        - full_stack_staff_engineer
        - frontend_engineer
        - backend_engineer
        - research_analyst
        - qa_guardian
        - security_reviewer
        - platform_reliability_engineer
        - product_strategy_analyst
        - visual_researcher
        - delivery_governor
  fallback_cycle:
    - full_stack_staff_engineer
    - research_analyst
    - qa_guardian
    - security_reviewer
    - platform_reliability_engineer
    - product_strategy_analyst
    - visual_researcher
    - delivery_governor
    - frontend_engineer
    - backend_engineer
    - documentation_analyst
    - api_integration_engineer
  expansion_cycle:
    - frontend_engineer
    - backend_engineer
    - api_integration_engineer
    - documentation_analyst
  optional_specialists:
    fallback_cycle:
      - api_integration_engineer
      - documentation_analyst
    routing_map:
      api_contracts: api_integration_engineer
      documentation_enablement: documentation_analyst
  pod_map:
    architecture:
      - principal_architect
      - full_stack_staff_engineer
      - backend_engineer
      - research_analyst
    experience:
      - frontend_engineer
      - visual_researcher
      - product_strategy_analyst
    risk:
      - security_reviewer
      - qa_guardian
      - platform_reliability_engineer
    delivery:
      - delivery_governor
      - documentation_analyst
      - api_integration_engineer
delegation:
  pod_context_limits:
    architecture:
      max_targeted_files: 8
      max_summary_estimated_tokens: 240
    experience:
      max_targeted_files: 7
      max_summary_estimated_tokens: 210
    risk:
      max_targeted_files: 7
      max_summary_estimated_tokens: 200
    delivery:
      max_targeted_files: 5
      max_summary_estimated_tokens: 170
  slot_contract:
    required_fields:
      - mission
      - scoped_context
      - constraints
      - evidence_anchors
      - output_budget
      - overlap_guard
      - skill_bundle
      - data_sensitivity
      - data_handling_guidance
      - release_relevance
  data_sensitivity:
    levels:
      - public
      - internal
      - restricted
    default: internal
    restricted_keywords:
      - customer data
      - pii
      - phi
      - payroll
      - ssn
      - tax id
      - credential
      - credentials
      - password
      - secret
      - token
      - private key
      - bearer
    restricted_rules:
      - pass only the minimum evidence anchors needed for the slot
      - never copy raw secrets, credentials, or full connection strings into briefs or handoffs
      - prefer summaries over raw excerpts whenever restricted material is nearby
  runtime_governance:
    canonical_sources:
      - multi_agent/config.md
      - multi_agent/agents
      - multi_agent/instructions
      - multi_agent/tools
      - skills
      - multi_agent/memory
      - multi_agent/metrics
      - multi_agent/todo
    generated_sources:
      - multi_agent/runtime/runs
      - multi_agent/runtime/validation-*.md
      - multi_agent/runtime/audit-*.md
    lazy_create_runtime_root: true
    pre_write_redaction_required: true
    redact_before_write_targets:
      - dispatch.json
      - brief.md
      - brief.json
      - handoffs/*
      - summary.md
      - summary.json
      - conflicts.md
      - conflicts.json
      - run.json
    artifact_persistence_policy: opt_in_for_debug_or_audit
    retention_mode: audit_minimized
    recovery_actions:
      - resume
      - replay
      - abort-clean
    stale_run_cleanup: redact_then_prune
reporting:
  live_status:
    commentary_required: true
    timestamp_format: iso8601
    cadence_guidance: on_start_then_on_state_change_or_every_20s
    default_sort: slot_asc
    default_format: markdown_table
    row_template: "| {timestamp} | {slot} | {agent} | {mission} | {skill_bundle} | {state} | {blockers} | {success_level} |"
    required_fields:
      - timestamp
      - slot
      - agent
      - mission
      - skill_bundle
      - state
      - blockers
      - success_level
    state_order:
      - queued
      - active
      - review
      - blocked
      - done
  final_report:
    required_sections:
      - execution_summary
      - live_run_digest
      - agent_scoreboard
      - blockers_and_resolutions
      - open_questions
      - score_delta
      - documentation_delta
      - metrics_delta
      - final_deep_analysis
  continuous_improvement:
    auto_run_end_of_cycle: true
    required_actions:
      - deep_analysis
      - documentation_sync
      - metrics_update
      - score_refresh
memory:
  session_directory: multi_agent/memory/sessions
  history_directory: multi_agent/memory/history
  active_plan: multi_agent/todo/active-plan.md
metrics:
  agent_performance: multi_agent/metrics/agent-performance.md
  token_usage: multi_agent/metrics/token-usage.md
routing:
  keyword_groups:
    - name: planning_analysis
      weight: 1.4
      keywords:
        - { term: analysis, weight: 1.4 }
        - { term: deep analysis, weight: 1.8 }
        - { term: analiz, weight: 1.4 }
        - { term: derinlemesine, weight: 1.6 }
        - { term: plan, weight: 1.5 }
        - { term: roadmap, weight: 1.8 }
        - { term: strategy, weight: 1.7 }
        - { term: optimize, weight: 1.4 }
        - { term: optimization, weight: 1.5 }
      priority_agents:
        - research_analyst
        - product_strategy_analyst
        - documentation_analyst
    - name: orchestration_ops
      weight: 1.3
      keywords:
        - { term: workflow, weight: 1.4 }
        - { term: orchestration, weight: 1.7 }
        - { term: handoff, weight: 1.3 }
        - { term: runtime, weight: 1.2 }
        - { term: automation, weight: 1.5 }
        - { term: governance, weight: 1.5 }
      priority_agents:
        - team_lead
        - delivery_governor
    - name: frontend_experience
      weight: 1.4
      keywords:
        - frontend
        - ui
        - ux
        - design
        - css
        - layout
        - react
        - accessibility
        - tasarim
        - gorsel
        - erisilebilirlik
      priority_agents:
        - frontend_engineer
        - visual_researcher
    - name: backend_platform
      weight: 1.4
      keywords:
        - backend
        - api
        - database
        - service
        - auth
        - migration
        - persistence
        - queue
        - worker
        - server
      priority_agents:
        - backend_engineer
        - full_stack_staff_engineer
    - name: api_contracts
      weight: 1.5
      keywords:
        - contract
        - schema
        - integration
        - sdk
        - webhook
        - openapi
        - grpc
        - event
        - api gateway
      priority_agents:
        - api_integration_engineer
        - backend_engineer
    - name: quality_release
      weight: 1.3
      keywords:
        - test
        - qa
        - regression
        - release
        - rollout
        - deployment
        - rollback
        - canary
      priority_agents:
        - qa_guardian
        - delivery_governor
    - name: security
      weight: 1.5
      keywords:
        - security
        - auth
        - jwt
        - oauth
        - encryption
        - threat
        - vulnerability
        - guvenlik
        - secret
      priority_agents:
        - security_reviewer
        - backend_engineer
    - name: architecture
      weight: 1.5
      keywords:
        - architecture
        - mimari
        - domain
        - clean architecture
        - solid
        - ddd
        - boundary
        - modular
        - contract
      priority_agents:
        - principal_architect
        - full_stack_staff_engineer
    - name: reliability_performance
      weight: 1.4
      keywords:
        - performance
        - latency
        - throughput
        - reliability
        - resiliency
        - observability
        - scaling
        - incident
        - outage
        - performans
      priority_agents:
        - platform_reliability_engineer
        - backend_engineer
    - name: enterprise_delivery
      weight: 1.3
      keywords:
        - acceptance
        - scope
        - rollout
        - ownership
        - dependency
        - stakeholder
        - adoption
        - onboarding
      priority_agents:
        - product_strategy_analyst
        - delivery_governor
    - name: documentation_enablement
      weight: 1.2
      keywords:
        - documentation
        - docs
        - readme
        - adr
        - runbook
        - changelog
        - handbook
        - guide
      priority_agents:
        - documentation_analyst
        - research_analyst
```

## Interpretation Rules
- Default keyword `match_mode` is `word`.
- Optional keyword objects may set `match_mode` and `weight`.
- Group score = `sum(matched keyword weights) * group.weight`.
- Sort routing hits by `score desc`, `match_count desc`, then declaration order asc.
- Prompt suffix parsing is case-insensitive.
- Worker identity, default skills, and pod ownership are sourced from `multi_agent/agents/*.agent.md`; the `agents:` map above is a routing and budget overlay that must stay in sync.
- If `agent_count == defaults.core_squad_agent_count`, seed assignments from `allocation.named_squads.enterprise_x10_core`.
- If `agent_count == defaults.full_stack_squad_agent_count`, seed assignments from `allocation.named_squads.enterprise_x12_full_stack`.
- `x11` seeds `enterprise_x10_core` then adds the top-ranked expansion agent not already present.
- `x13` seeds `enterprise_x12_full_stack` then adds the highest-ranked optional specialist matched by routing; if no optional specialist matches, use `allocation.optional_specialists.fallback_cycle` order.
- `x14` seeds `enterprise_x12_full_stack` then adds both optional specialists in `allocation.optional_specialists.fallback_cycle` order.
- No duplicate worker agents are needed while `agent_count <= defaults.max_unique_worker_agents`.
- Context slicing must respect `delegation.pod_context_limits` and `delegation.data_sensitivity`.
- Release-oriented detection must use `policies.governance.release_oriented_keywords`.
- Unresolved blocker findings from `policies.governance.blocker_authorities` force the final run state to `blocked`.
- Live reporting and end-of-cycle improvement must respect `reporting.*`.
