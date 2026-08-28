# BRIEFING — 2026-08-28T18:50:00Z

## Mission
Investigate the engine isometric math, world coordinate transforms, movement, camera, and map systems for the Project Zomboid Go recreation and assess requirements/strategies for upgrading tile size and math to 256x128 (4x texture resolution).

## 🔒 My Identity
- Archetype: explorer
- Roles: investigator, reporter
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2
- Original parent: f7a8f969-fc3f-4f72-a625-45c03a6444ae
- Milestone: codebase survey

## 🔒 Key Constraints
- Read-only investigation — do NOT implement
- Analyze math, transforms, movement, camera, map systems, collision, sorting, and 256x128 migration implications.

## Current Parent
- Conversation ID: f7a8f969-fc3f-4f72-a625-45c03a6444ae
- Updated: 2026-08-28T18:50:00Z

## Investigation State
- **Explored paths**: `internal/game/world/map.go`, `internal/game/game.go`, `internal/ecs/components.go`, `cmd/tools/genassets/main.go`, `internal/assets/assets.go`, all unit and stress tests.
- **Key findings**: Documented all constants, formulas, transforms, anchors, speeds, ranges, colliders, and full 4x upgrade requirements for 256x128 resolution.
- **Unexplored areas**: None for this survey mission.

## Key Decisions Made
- Produced comprehensive `survey_report.md` and 5-component `handoff.md`.

## Artifact Index
- `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey_report.md` — Comprehensive Survey Report
- `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/handoff.md` — 5-Component Handoff Report
