## 2026-08-29T15:36:09Z
You are teamwork_preview_explorer_remediation_1.
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_remediation_1.
Please read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md, /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_5/PROJECT.md, and the FULL Victory Audit report at /home/bryce/code/go-zomboid/.agents/victory_auditor_4/handoff.md.

Task:
1. Examine `internal/assets/assets.go` and `internal/assets/assets_test.go`.
   - Identify all 27 legacy `*ebiten.Image` pointer variables (`PlayerImage`, `ZombieImage`, `RunnerImage`, `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage`, `FoodImage`, `WaterImage`, `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `AntidoteImage`).
   - Identify their expected legacy image file paths in `images/` (e.g. `images/player.png`, `images/grass.png`, `images/wall.png`, `images/tree.png`, etc.) and their expected dimensions.
2. Examine the new external asset pointer variables (`BenchImage`, `ChestImage`, `SculptureImage`, `Sculpture1Image`, `Sculpture2Image`, `BushImage`, `Bush1Image`, `Bush2Image`, `Bush3Image`, `Bush4Image`, `FlowerImage`, `Flower1Image`, `Flower2Image`, `Flower3Image`, `StoneImage`, `Stone1Image`, `Stone2Image`, `ForestStumpImage`, `GrassTuft1Image`, `GrassTuft2Image`, `LabTilesetImage`, `ZombieTilesetImage`).
   - Verify their paths to the newly ingested PNG files from `context/` (e.g. `images/Small Forest/Bench and chest/Bench.png`, etc.).
3. Examine `internal/game/draw_depth_test.go` and `internal/game/game.go` to see why `TestDrawSystem_SpriteGeometricAnchors` failed on Wall and Tree anchors, and ensure the rendering/anchor logic handles both 256x256 legacy obstacles and the new prop sprites properly.
4. Formulate the exact code changes needed in `internal/assets/assets.go`, `internal/game/game.go`, and test files so that:
   - All 27 legacy pointers load from their proper `images/<name>.png` files.
   - All new external asset pointers load from their external PNG paths in `images/`.
   - `CC=gcc go test ./...` passes 100% across all packages.
   - `CC=gcc go build ./cmd/game` builds cleanly.

Write your report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_remediation_1/analysis.md` and handoff to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_remediation_1/handoff.md`. Send a message when complete.
