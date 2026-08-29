# BRIEFING — 2026-08-29T15:56:00Z

## Mission
Investigate go-zomboid testing infrastructure, game entry/loop in cmd/game, identify broken tests from Isometric to 2D Orthogonal migration, and design test strategy and acceptance criteria checklist.

## 🔒 My Identity
- Archetype: explorer
- Roles: investigation, test analysis, acceptance criteria design
- Working directory: /home/bryce/code/go-zomboid/.agents/explorer_survey_3
- Original parent: d24acf99-20c6-4e30-b7be-668df332bc88
- Milestone: survey

## 🔒 Key Constraints
- Read-only investigation — do NOT implement
- Produce 5-component handoff report in .agents/explorer_survey_3/handoff.md
- Communicate with parent via send_message

## Current Parent
- Conversation ID: d24acf99-20c6-4e30-b7be-668df332bc88
- Updated: 2026-08-29T15:56:00Z

## Investigation State
- **Explored paths**:
  - `cmd/game/main.go`
  - `internal/assets/*_test.go` and `assets.go`
  - `internal/game/*_test.go` and `game.go`
  - `internal/game/world/*_test.go` and `map.go`
  - `internal/ecs/*_test.go` and `components.go`
- **Key findings**:
  - `internal/game/world` and `internal/ecs` already operate in Cartesian / discrete tile coordinates and pass tests.
  - `internal/assets` tests fail due to hardcoded legacy 256x128 diamond dimensions while assets were switched to RPG Maker sprites.
  - `internal/game` tests fail on `WorldToIso` projection math, isometric vertical sprite anchors `(-128, -128)`, and dimetric depth sorting.
  - Formulated full test architecture and Acceptance Criteria Checklist for 2D Orthogonal math, seamless tiling, Dungeon Master wave spawning, dynamic loot drops, day/night aggression, and headless verification.
- **Unexplored areas**: None for this survey milestone.

## Key Decisions Made
- Mapped all 14 broken test cases and defined exact orthogonal replacements.
- Structured acceptance criteria around the 4 core verification requirements in `ORIGINAL_REQUEST.md`.

## Artifact Index
- `/home/bryce/code/go-zomboid/.agents/explorer_survey_3/handoff.md` — Complete 5-component handoff report
- `/home/bryce/code/go-zomboid/.agents/explorer_survey_3/progress.md` — Progress tracker
- `/home/bryce/code/go-zomboid/.agents/explorer_survey_3/DISPATCH.md` — Dispatch log
