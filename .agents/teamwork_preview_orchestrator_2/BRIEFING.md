# BRIEFING — 2026-08-28T18:47:36Z

## Mission
Significantly improve the game's resolution to achieve extremely smooth, high-fidelity tiles (256x128 4x scale) matching Dribbble vector art, upgrade isometric math across the engine, and implement dynamic bezier curve combat attack swoosh dynamics in DrawSystem.

## 🔒 My Identity
- Archetype: teamwork_preview_orchestrator_2
- Roles: orchestrator, user_liaison, human_reporter, successor
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_2
- Original parent: parent
- Original parent conversation ID: 57babd7d-3cc2-4a0a-8df9-13b3238d25a0

## 🔒 My Workflow
- **Pattern**: Project Pattern
- **Scope document**: /home/bryce/code/go-zomboid/PROJECT.md
1. **Decompose**: Survey (3 explorers in parallel), decompose into modular milestones + parallel E2E testing track.
2. **Dispatch & Execute**:
   - For each milestone: 3 Explorers -> 1 Worker -> 2 Reviewers -> 2 Challengers -> 1 Forensic Auditor -> Gate check in GATE_STATUS.md.
   - Dual-track: Implementation milestones + E2E Testing track publishing TEST_READY.md.
   - Final milestone: Pass 100% E2E tests + Tier 5 adversarial hardening.
3. **On failure**:
   - Retry / Replace / Skip / Redistribute / Redesign.
   - Audit violation is a strict binary veto.
4. **Succession**: Threshold 16 spawns. Self-succeed after writing handoff.md.

- **Work items**:
  0. Survey & full codebase mapping [in-progress]
  1. Milestone 1: High-Fidelity Asset Generation 4x Scaling (256x128 floor tiles, proportional entities/props/walls, geometric vector overlays) [pending]
  2. Milestone 2: Isometric Engine Math & World Coordinates (TileSize, WorldToIso, IsoToWorld, camera/speed/chunk bounds) [pending]
  3. Milestone 3: Bezier Curve Combat Dynamics in DrawSystem (weapon swing arcs/swooshes, facing direction, mouse click interaction) [pending]
  4. Parallel Track: E2E Testing Suite (Tiers 1-4, runner, verification) [pending]
  5. Milestone 4: Integration, Full E2E Test Suite Validation, and Adversarial Coverage Hardening (Tier 5) [pending]
- **Current phase**: 0 (Survey)
- **Current focus**: Survey phase with 3 parallel explorers

## 🔒 Key Constraints
- NEVER write, modify, or create source code directly.
- NEVER run build/test commands directly.
- NEVER investigate codebase directly — dispatch Explorers.
- Audit is a binary veto.
- Always include path to ORIGINAL_REQUEST.md in every dispatch.
- Never reuse subagents after handoff.
- Set safety timers on dispatch.

## Current Parent
- Conversation ID: 57babd7d-3cc2-4a0a-8df9-13b3238d25a0
- Updated: 2026-08-28T18:47:36Z

## Key Decisions Made
- Initiated Milestone 2 Orchestration with Survey phase (3 Explorers).

## Team Roster
| Agent | Type | Work Item | Status | Conv ID |
|-------|------|-----------|--------|---------|
| survey_explorer_1 | teamwork_preview_explorer | Survey asset generation & 4x scale | completed | 0f8b90c1-42af-4aa8-871c-fc704f966a20 |
| survey_explorer_2 | teamwork_preview_explorer | Survey isometric math & engine coordinates | completed | 472505d3-189a-4a2e-9007-ce0c51d8e847 |
| survey_explorer_3 | teamwork_preview_explorer | Survey combat dynamics & bezier curve trails | completed | e56938bd-08e5-4e65-b5b6-cee06936b3f6 |
| m1_explorer_1 | teamwork_preview_explorer | M1 Floor tiles (256x128) & overlays analysis | completed | 097112d6-9c47-403a-adef-3e4f3b831cf9 |
| m1_explorer_2 | teamwork_preview_explorer | M1 Obstacles (256x256) & Entities (64x128) analysis | completed | 89697d71-90c4-4145-98d3-954526d1ae08 |
| m1_explorer_3 | teamwork_preview_explorer | M1 Items (64x64) & Asset tests analysis | completed | c8dc83d7-d0d3-4d3d-9d64-5c73cfbec812 |
| m1_worker_1 | teamwork_preview_worker | Implement M1 4x asset generator and test suites | completed | 298986e8-2d43-42a5-ae03-431714e77551 |
| m1_reviewer_1 | teamwork_preview_reviewer | Review M1 implementation and tests | in-progress | eb64b3b2-5bc6-4704-befb-75dd5a422def |
| m1_reviewer_2 | teamwork_preview_reviewer | Review M1 code quality and determinism | in-progress | 0e72f2e7-e447-4694-8ed8-89d4e796f7f7 |
| m1_challenger_1 | teamwork_preview_challenger | Empirical verification of M1 dimensions & bounds | in-progress | c3dee392-48da-42ba-b85f-7cfeba188044 |
| m1_challenger_2 | teamwork_preview_challenger | Empirical stress-test of M1 asset loader & colors | in-progress | b3f7257e-c61e-4048-aab9-4938b8a0f285 |
| m1_auditor_1 | teamwork_preview_auditor | Forensic integrity audit of M1 | completed | 0d0f71eb-2c8f-458f-bba6-7bb7d73180d3 |
| m1_explorer_fix_1 | teamwork_preview_explorer | Investigate dirt.png and assets.go fixes | completed | 37d116e3-06df-434c-bdc7-baf52a9ea4ca |
| m1_explorer_fix_2 | teamwork_preview_explorer | Audit all floor generators for alpha/bounds | completed | abf7baf4-d653-41ca-8f9c-68857c40fc67 |
| m1_explorer_fix_3 | teamwork_preview_explorer | Validate test suite assertions & race conditions | completed | 01a78dc3-a12c-4de8-ac27-fc20e78fd3e7 |
| m1_worker_2 | teamwork_preview_worker | Apply dirt.png and assets.Load sync.Once fixes | in-progress | e704e342-b9e9-4d04-be7e-6ebef9865dcf |

## Succession Status
- Succession required: yes (threshold 16 reached, all subagents completed)
- Spawn count: 16 / 16
- Pending subagents: none
- Predecessor: none
- Successor spawned: 67ce220b-8fe1-4a30-9c84-cbb0937a62cd
- Successor generation: gen3

## Active Timers
- Heartbeat cron: killing for succession
- Safety timer: none
- On succession: kill all timers before spawning successor
- On context truncation: run `manage_task(Action="list")` — re-create if missing

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md — User request specification
- /home/bryce/code/go-zomboid/PROJECT.md — Global project plan and milestones
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_2/progress.md — Orchestrator liveness and execution log
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_2/GATE_STATUS.md — Milestone gate verdicts
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1/survey_report.md — Survey report for assets
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey_report.md — Survey report for engine math
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3/survey_report.md — Survey report for combat & bezier dynamics
