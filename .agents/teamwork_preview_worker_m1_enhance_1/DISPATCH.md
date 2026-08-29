## 2026-08-29T16:52:25Z
You are Worker 1 implementing Milestone 1: Requirement R1 (Tile Rendering Upgrade & Autotiling).
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m1_enhance_1.
Read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md (especially section ## 2026-08-29T16:48:41Z) and /home/bryce/code/go-zomboid/PROJECT.md.
Also read the survey findings in /home/bryce/code/go-zomboid/.agents/explorer_survey_r1_1/handoff.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Tasks for Milestone 1:
1. Implement autotiling and terrain blending for the 2D orthogonal grid in `internal/game` / `internal/game/world` / `internal/assets`:
   - Implement bitmasking / neighbor evaluation (cardinal 4-bit and/or 4-quadrant sub-tile blob autotiling) for terrain tiles (`TileGrass`, `TileDirt`, `TileConcrete`, `TileAsphalt`, `TileWoodFloor`, `TileTileFloor`).
   - Eliminate harsh 90-degree square borders between different terrains (e.g. grass and dirt transitions, sidewalks and roads) by rendering seamless transition overlays or sub-tile autotiles.
   - Implement connected autotiling for `TileWall` and `TileFence` (connecting horizontal, vertical, corners NW/NE/SW/SE, T-junctions, and South-facing facade depth).
2. Integrate autotiling into `DrawSystem.Draw` in `internal/game/game.go` so ground and wall/fence rendering seamlessly uses autotile computations on the 2D orthogonal grid.
3. Verify that all existing and new tests pass using:
   `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v ./...`
   and verify compilation with:
   `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game`
4. Add unit tests for autotiling calculations and rendering invariants in `internal/game/world` and `internal/game`.
5. Write your handoff report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m1_enhance_1/handoff.md` and send a message back when complete.
