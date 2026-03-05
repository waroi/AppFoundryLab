# Run Unit Tests Rule (Markdown Regression Set)

Purpose:
- Replace executable unit tests with deterministic scenario checks.

## Required Cases (30)
1. Parse lowercase suffix: `Refactor API x4` -> `agent_count=4`
2. Parse uppercase suffix: `Refactor API X5` -> `agent_count=5`
3. Normalize invalid counts: `Refactor API x0` -> `agent_count=1`
4. Word boundary protection: `debug dashboard` + `bug` in word mode -> no match
5. Planning routing: `deep analysis roadmap x4` -> includes `planning_analysis`
6. Frontend routing: `frontend accessibility redesign x6` -> includes `frontend_engineer`
7. Backend routing: `backend auth migration x6` -> includes `backend_engineer`
8. API routing: `schema contract integration x6` -> includes `api_integration_engineer`
9. Documentation routing: `runbook handbook refresh x6` -> includes `documentation_analyst`
10. x1 combined role: `basic task x1` -> `team_lead_architect_combined`
11. x10 core squad: `enterprise modernization x10` -> exactly 10 unique assignments
12. x11 routed expansion: `frontend redesign x11` -> `frontend_engineer` added to core squad
13. x12 full-stack squad: `enterprise modernization x12` -> includes `frontend_engineer` and `backend_engineer`
14. x13 API specialist: `schema contract review x13` -> includes `api_integration_engineer`
15. x13 documentation specialist: `runbook onboarding refresh x13` -> includes `documentation_analyst`
16. x14 specialist expansion: `api contract onboarding x14` -> includes `api_integration_engineer` and `documentation_analyst`
17. Model tier mapping: `auth hardening x12` -> `principal_architect -> gpt-5.4`, `security_reviewer -> gpt-5.3-codex`
18. Conditional escalation: `production latency rollback plan x12` -> reliability or QA escalates to `gpt-5.3-codex`
19. Context packing contract: `enterprise onboarding redesign x12` -> pod limits reference architecture, experience, risk, delivery
20. Skill loading contract: `frontend onboarding redesign x12` -> frontend lane lists `clean-code` and `frontend-development`
21. Clean-code governance: canonical docs reference `clean-code-standards.md` and `skills/clean-code/SKILL.md`
22. Coding lane enforcement: every coding or code-review agent lists `clean-code` in its skill bundle
23. Memory and metrics contract: docs review -> active plan, memory, and metrics files exist and are referenced
24. Legacy mirror removal: removed worker mirror directories do not exist and are not referenced by canonical docs
25. Live reporting contract: `enterprise delivery review x12` -> live status fields include timestamp, agent, mission, skill bundle, blockers, success level
26. Release-oriented detection: `production rollback review x12` -> `release_oriented=true`
27. Hard blocker finalization: unresolved guard blocker on a release-oriented run -> final state `blocked`
28. Data sensitivity contract: `customer data contract review x12` -> dispatch includes `data_sensitivity=restricted`
29. Artifact-wide safety: generated artifacts document redaction across dispatch, brief, handoff, summary, conflicts, and run metadata
30. Continuous improvement contract: any development prompt -> final cycle includes deep analysis, doc sync, metrics, score refresh
