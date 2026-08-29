## 2026-08-29T15:24:10Z

You are teamwork_preview_worker_m2.
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2.
Please read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_5/PROJECT.md.
Also refer to the Explorer survey report at /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Task (Milestone 2: World TileType & Map Logic):
1. In `internal/game/world/map.go`:
   - Declare new `TileType` constants:
     - `TileBench TileType = 16`
     - `TileChest TileType = 17`
     - `TileSculpture TileType = 18`
     - `TileBush TileType = 19`
     - `TileFlower TileType = 20`
     - `TileStone TileType = 21`
   - Implement `TileType` methods:
     - `IsSolid()`: return true for `TileBench`, `TileChest`, `TileSculpture`, `TileStone`; return false for `TileBush`, `TileFlower`.
     - `BlocksVision()`: return false for all new props (only `TileWall` blocks vision).
     - `IsFloor()`: return false for all new props.
     - `String()`: return "bench", "chest", "sculpture", "bush", "flower", "stone".
   - Update world generation in `internal/game/world/map.go` (`NewMap` / `placeEnvironmentalProps`):
     - Procedurally place benches in town parks/sidewalks, chests in houses/warehouses, sculptures in town plazas/parks, bushes, flowers, and stones in open green spaces.
     - CRITICAL: Ensure all 10 legacy tile types (`TileGrass`, `TileWall`, `TileDirt`, `TileWoodFloor`, `TileTree`, `TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileFence`, `TileDebris`) continue to be generated with non-zero counts so that existing tests like `TestEmpirical_All10TileTypesGenerated` pass without regression.
2. In `internal/game/world/map_test.go` and/or new test files:
   - Add unit tests verifying properties for all new tile types.
   - Verify procedural generation counts and placement for new tile types.
3. Verification:
   - Run `CC=gcc go test -v ./internal/game/world/...`
   - Run `CC=gcc go test ./...`
   - Run `CC=gcc go build ./cmd/game`

Write your handoff report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2/handoff.md`. Send a message when complete.
