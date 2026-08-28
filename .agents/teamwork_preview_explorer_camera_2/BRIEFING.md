# BRIEFING — 2026-08-28T19:24:59Z

## Mission
Investigate Requirement 1 (IsoToWorld Mouse-Click Inversion & Input Math with 0.5 rendering scale) and Requirement 2 (Smooth Camera Centering / Lerping) for go-zomboid.

## 🔒 My Identity
- Archetype: explorer
- Roles: investigator, analyzer, synthesizer
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_2
- Original parent: 9749292c-47da-41c9-80d9-536a89b92052
- Milestone: camera_and_mouse_input_analysis

## 🔒 Key Constraints
- Read-only investigation — do NOT implement in source code
- Produce 5-component handoff report at handoff.md
- Verify all file paths and line numbers

## Current Parent
- Conversation ID: 9749292c-47da-41c9-80d9-536a89b92052
- Updated: 2026-08-28T19:24:59Z

## Investigation State
- **Explored paths**:
  - `internal/game/game.go` (`UpdateSystem`, `DrawSystem`, `WorldToIso`, `IsoToWorld`, `Layout`, `DrawAttackSwingArc`)
  - `internal/game/world/map.go` (`Map`, `TileSize`, `CalculateFOV`, `IsColliding`)
  - `internal/game/*_test.go` and `internal/game/world/*_test.go` (verified 100% test pass and constructor backward compatibility)
- **Key findings**:
  - Derived exact bijective mouse inversion formula: `mouseIsoX = camX + (mx - 640)/0.5`, `mouseIsoY = camY + (my - 360)/0.5`, `IsoToWorld(mouseIsoX, mouseIsoY)`
  - Formulated persistent `Camera` struct with exponential per-frame lerping (`camX += (targetCamX - camX) * 0.10`) and spawn snapping
  - Formulated `GeoM` scaling matrix for 50% world rendering centered at `(640, 360)`
  - Expanded `visionRadius` to 2200px and `CalculateFOV` radius to 25 tiles
- **Unexplored areas**: None for this milestone.

## Key Decisions Made
- Handoff report written to `handoff.md` with complete 5-component structure and ready for implementation.

## Artifact Index
- handoff.md — Complete 5-component handoff report
- progress.md — Liveness and progress tracking
- DISPATCH.md — Initial task dispatch
