## 2026-08-29T15:56:46Z
You are the Technical Director Worker implementing Milestone 1: 2D Orthogonal Engine Overhaul (R1).
Working directory: /home/bryce/code/go-zomboid/.agents/worker_m1
Original request path: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project plan path: /home/bryce/code/go-zomboid/PROJECT.md
Survey reports:
- /home/bryce/code/go-zomboid/.agents/explorer_survey_1/handoff.md
- /home/bryce/code/go-zomboid/.agents/explorer_survey_3/handoff.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. An auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Implementation Details:
1. Coordinate Transformations (`internal/game/game.go`):
   - Rewrite `WorldToIso`, `IsoToWorld`, `ScreenToIso`, `ScreenToWorld`, and add `WorldToScreen` so they use strict 2D Orthogonal (top-down) grid Cartesian math.
   - Screen/world projection:
     screenX = (wx - camX) * zoom + 640.0
     screenY = (wy - camY) * zoom + 360.0
     wx = camX + (screenX - 640.0) / zoom
     wy = camY + (screenY - 360.0) / zoom
2. Camera Controller (`internal/game/game.go`):
   - Update `Camera` struct, `Camera.Snap()`, and `Camera.Update()` to operate on Cartesian world coordinates $(wx, wy)$.
   - In `Game.Reset()`, snap camera directly to player Cartesian world spawn coordinates `(gameMap.PlayerSpawn.X, gameMap.PlayerSpawn.Y)`.
   - In `UpdateSystem.Update()`, update camera to follow player's Cartesian position `(pPos.X, pPos.Y)`.
   - Update mouse click / aim unprojection in `processInputAndCombat()` to use orthogonal `ScreenToWorld`.
3. DrawSystem (`internal/game/game.go`):
   - Ground Pass: Remove 2:1 dimetric isometric diamond translations (`Translate(-128, 0)`). Draw rectangular tiles at top-left origin `(tx * TileSize, ty * TileSize)` scaled to fill the cell completely with ZERO black gaps.
   - Props / Obstacles Pass: Render props with top-down anchors and depth key `worldY + TileSize`.
   - Entities & Items Pass: Render entities centered at `(pos.X, pos.Y)` with depth key `pos.Y`.
   - Depth Sorting: Sort renderables by Y coordinate (`Depth`) for natural vertical top-down occlusion.
   - Bezier Combat Swoosh: In `DrawAttackSwingArc()`, convert control points directly from world Cartesian to screen coordinates without isometric skewing.
4. Asset Pipeline (`internal/assets/assets.go`):
   - Ensure all image pointers in `internal/assets/` load non-nil textures and are scaled/tiled seamlessly.
5. Verification:
   - Run `CC=gcc go build ./...`
   - Run `CC=gcc go test -v ./internal/ecs ./internal/game/world`
   - Document all changes and verification command outputs in `/home/bryce/code/go-zomboid/.agents/worker_m1/handoff.md`.
   - Send completion message to parent.
