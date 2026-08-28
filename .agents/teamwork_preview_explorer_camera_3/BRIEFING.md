# BRIEFING — 2026-08-28T19:24:50Z

## Mission
Investigate Requirement 3: Vision Radius & FOV Culling Expansion for 50% zoomed-out 1280x720 camera.

## 🔒 My Identity
- Archetype: Explorer
- Roles: Read-only investigator, codebase analyst, synthesis & handoff
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_3
- Original parent: 9749292c-47da-41c9-80d9-536a89b92052
- Milestone: Requirement 3 (Vision Radius & FOV Culling Expansion)

## 🔒 Key Constraints
- Read-only investigation — do NOT implement
- Analyze world/map.go (visionRadius, FOV raycasting/culling) and game.go (tile, entity, obstacle visibility checks and screen culling)
- Identify all culling thresholds and render cutoffs
- Determine required expansions for 50% zoom-out (0.5x zoom or 50% zoom out viewport)
- Check all unit/integration tests affected

## Current Parent
- Conversation ID: 9749292c-47da-41c9-80d9-536a89b92052
- Updated: 2026-08-28T19:24:50Z

## Investigation State
- **Explored paths**: `internal/game/world/map.go`, `internal/game/world/*_test.go`, `internal/game/game.go`, `internal/game/*_test.go`, `cmd/tools/genassets/*`, `internal/assets/*`
- **Key findings**:
  - Screen resolution 1280x720 with 0.5x zoom creates a 2560x1440 isometric viewport.
  - Viewport corners reach Cartesian distance 1362.35px from camera center.
  - With sprite anchors/dimensions (256px) and camera lerping lag (200px), minimum distance is 1818.35px.
  - Existing DrawSystem culling `visionRadius := 1000.0` is inadequate and causes aggressive pop-in in the outer 30% of the screen. Must be expanded to `2000.0` px.
  - Existing FOV calculation in UpdateSystem calls `CalculateFOV(px, py, 15)`. Must be expanded to `20` tiles (2560px).
  - Zombie AI perception `visionRadius := 600.0` in `processZombies` must NOT be changed to preserve combat balance.
  - All existing unit/integration tests remain 100% passing and unaffected.
- **Unexplored areas**: None for Requirement 3.

## Key Decisions Made
- Completed full mathematical and code analysis for Requirement 3.
- Produced 5-component handoff report in `handoff.md`.

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_3/DISPATCH.md — Dispatch instructions
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_3/BRIEFING.md — Working memory
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_3/progress.md — Liveness heartbeat
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_3/handoff.md — Complete investigation handoff report
