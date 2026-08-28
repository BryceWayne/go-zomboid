## 2026-08-28T19:25:20Z
You are a Worker agent (teamwork_preview_worker_camera_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_camera_1
Project root: /home/bryce/code/go-zomboid

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

MANDATORY INSTRUCTIONS & INPUT FILES:
1. Read the original request at /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and the milestone scope at /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_4/SCOPE.md.
2. Read the detailed technical handoff reports from the 3 Explorers:
   - /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1/handoff.md
   - /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_2/handoff.md
   - /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_3/handoff.md

3. Scope & File Ownership:
   You exclusively own:
   - `internal/game/game.go`
   - `internal/game/camera_test.go` (create/modify unit tests for camera lerp, screen-to-world inversion, and culling)
   - Any necessary unit test updates in `internal/game/`

4. Requirements to Implement:
   - R1 (Global Camera Zoom): In `internal/game/game.go`, update `DrawSystem.Draw` to apply a global 50% scale matrix (`0.5, 0.5`) to all game world rendering (ground tiles, props/walls/obstacles, items, entities, facing reticle, Bezier combat swooshes, and shotgun blast lines) with screen centering offset `(640, 360)`. Keep HUD/UI at 1:1 scale on `screen`.
   - R1 (Mouse Inversion Math): Implement `ScreenToIso` and `ScreenToWorld` in `internal/game/game.go` (and use in `UpdateSystem.processInputAndCombat` / mouse input):
     $$mouseIsoX = camX + (mx - 640.0) / 0.5$$
     $$mouseIsoY = camY + (my - 360.0) / 0.5$$
     $$mouseWorldX, mouseWorldY = IsoToWorld(mouseIsoX, mouseIsoY)$$
     Ensure clicking on a tile moves the player accurately to that tile.
   - R2 (Smooth Camera Centering & Lerp): Create a `Camera` struct in `internal/game/game.go` with smooth lerp tracking (`LerpFactor = 0.10` or per-frame exponential smoothing), `Snap` method for spawn/reset, and wire a shared `*Camera` instance to both `UpdateSystem` and `DrawSystem` in `Game.Reset()`.
   - R3 (Vision Radius & FOV Culling Expansion): Expand `visionRadius` in `DrawSystem.Draw` to `2200.0` (from 1000.0) and `CalculateFOV` radius in `UpdateSystem.Update` to `22` (or `25`) tiles (from 15). Keep zombie AI sight `visionRadius := 600.0` in `processZombies` unchanged.
   - Update Day/Night lighting rectangle in `DrawSystem.Draw` to 1280x720.

5. Testing & Verification:
   - Write comprehensive unit tests in `internal/game/camera_test.go` testing:
     a. `ScreenToWorld` and `ScreenToIso` mathematical invertibility / bijectivity (forward -> invert -> roundtrip error < 1e-9).
     b. `Camera.Update` smooth exponential convergence towards target.
     c. `Camera.Snap` instant position snapping.
     d. Player ground anchor centering on (640, 360) screen coordinates.
   - Run tests: `CC=gcc go test ./...` and ensure 100% pass across all packages.
   - Test build: `CC=gcc go build -o /tmp/game_test ./cmd/game`.

6. Output & Handoff:
   - Write your complete implementation report and test results to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_camera_1/handoff.md`.
   - Send a message to the orchestrator when finished.
