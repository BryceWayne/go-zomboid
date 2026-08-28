# BRIEFING — 2026-08-28T19:24:45Z

## Mission
Investigate Requirement 1: Global Camera Zoom & DrawSystem rendering scale (50% scale matrix on world rendering, keeping HUD/UI 1:1, camera centering, screen dimensions).

## 🔒 My Identity
- Archetype: Explorer
- Roles: Investigation, Synthesis
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1
- Original parent: 9749292c-47da-41c9-80d9-536a89b92052
- Milestone: Requirement 1 Investigation (Global Camera Zoom & DrawSystem)

## 🔒 Key Constraints
- Read-only investigation — do NOT implement in source code
- Keep HUD/UI at 1:1 scale
- Game world elements scaled by 50% (0.5x zoom out)
- Screen dimensions: 1280x720
- Write handoff.md with 5 components and communicate to parent via send_message

## Current Parent
- Conversation ID: 9749292c-47da-41c9-80d9-536a89b92052
- Updated: 2026-08-28T19:24:45Z

## Investigation State
- **Explored paths**: `cmd/game/main.go`, `internal/game/game.go`, `internal/game/world/map.go`, test suites (`game_test.go`, `bezier_combat_test.go`, `game_stress_test.go`, `combat_empirical_challenger_m4_test.go`).
- **Key findings**:
  - Screen dimensions are 1280x720 (center is 640x360).
  - Current hardcoded camera offsets are 400x300 (legacy 800x600).
  - For 50% scale ($S=0.5$), unscaled isometric viewport is 2560x1440, camera offsets are $camX = playerIsoX - 1280, camY = playerIsoY - 720$.
  - Viewport sprite drawing translates to $0.5 \cdot (isoX + ax - camX)$ and scales by 0.5.
  - Bezier curves in `DrawAttackSwingArc` scale screen control points by 0.5 relative to $(camX, camY)$.
  - Inverted mouse click mapping is $mouseIsoX = 2.0 \cdot mx + camX, mouseIsoY = 2.0 \cdot my + camY$, followed by standard `IsoToWorld`.
  - `visionRadius` in `DrawSystem` must be expanded from 1000.0 to 2200.0 and FOV raycast radius to 22 tiles.
  - HUD/UI remains at 1:1 scale on `screen`.
- **Unexplored areas**: None for Requirement 1.

## Key Decisions Made
- Completed detailed mathematical formulation and architectural options in `handoff.md`.

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1/DISPATCH.md — Task dispatch
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1/BRIEFING.md — Working memory
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1/progress.md — Liveness heartbeat
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1/handoff.md — 5-component analysis report
