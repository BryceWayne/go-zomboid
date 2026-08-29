## 2026-08-29T15:27:21Z

You are teamwork_preview_worker_m3.
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m3.
Please read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md, /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_5/PROJECT.md, and the Explorer survey report at /home/bryce/code/go-zomboid/.agents/teamwork_preview_spec_miner_survey_3/survey.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Task (Milestone 3: DrawSystem Rendering & Depth-Sorting):
1. In `internal/game/game.go`:
   - Inspect `DrawSystem.Draw` and the multi-pass rendering pipeline:
     - Ground Pass (Pass 1): For new prop tiles (`TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`), ensure the underlying terrain diamond (e.g. `assets.GrassImage`) is drawn beneath them so no transparent/black holes appear in the ground grid.
     - Depth-Sorted Sprite Pass (Pass 2): Collect sprites for the new tile types into the `sprites []Renderable` slice:
       - `TileBench` -> `assets.BenchImage`
       - `TileChest` -> `assets.ChestImage`
       - `TileSculpture` -> `assets.SculptureImage` (or `assets.Sculpture1Image`)
       - `TileBush` -> `assets.BushImage` (or `assets.Bush1Image`)
       - `TileFlower` -> `assets.FlowerImage` (or `assets.Flower1Image`)
       - `TileStone` -> `assets.StoneImage` (or `assets.Stone1Image`)
       - Set `Depth = worldX + worldY` so they are correctly depth-sorted relative to walls, trees, items, player, and zombies.
       - Apply proper geometric anchor transform (e.g., centering X and aligning bottom of sprite with tile base) and FOV explored/visible color tinting.
2. In `internal/game/`:
   - Add/update tests in `internal/game/` (e.g. `game_test.go` or `game_stress_test.go`) to verify that the new tile types are collected and depth-sorted correctly during `DrawSystem.Draw()`.
3. Verification:
   - Run `CC=gcc go test -v ./internal/game/...`
   - Run `CC=gcc go test ./...`
   - Run `CC=gcc go build ./cmd/game`

Write your handoff report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m3/handoff.md`. Send a message when complete.
