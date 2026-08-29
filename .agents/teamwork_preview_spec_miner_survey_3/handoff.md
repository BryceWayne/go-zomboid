# Handoff Report — Specification Mining & Survey of Rendering, Depth-Sorting, Lifecycle & Tests

**Agent:** `teamwork_preview_spec_miner_survey_3`  
**Working Directory:** `/home/bryce/code/go-zomboid/.agents/teamwork_preview_spec_miner_survey_3`  
**Parent Agent:** `2341cac8-3fc5-4c81-8832-e3f9a5a870ba`  
**Date:** 2026-08-29  

---

## 1. Observation
- **User Request (`ORIGINAL_REQUEST.md`)**:
  - R1: Retire Procedural Generation by completely deleting `cmd/tools/genassets` directory and its contents.
  - R2: Ingest external PNG files from `context/` into `internal/assets/images/` and update `internal/assets/assets.go` to load them natively.
  - R3: Infer and implement new world logic for imported assets (Benches, Chests, Sculptures), create new `TileType` constants in `internal/game/world/map.go`, and update `DrawSystem` in `internal/game/game.go` for proper rendering and depth sorting.
  - Acceptance Criteria: `cmd/tools/genassets` deleted; native PNG loading in `assets.go`; `CC=gcc go test ./...` passes all map and loading tests; `CC=gcc go run ./cmd/game` launches and visibly renders new world objects.
- **Rendering Architecture (`internal/game/game.go:803-1264`)**:
  - `WorldToIso(wx, wy)` converts Cartesian coordinates to isometric coordinates via $isoX = wx - wy, isoY = (wx + wy)/2$.
  - `DrawSystem.Draw` runs a 6-pass rendering pipeline:
    1. Pass 1: Background clear (`color.RGBA{15, 15, 15, 255}`).
    2. Pass 2: Base ground tile pass (Grass, Dirt, WoodFloor, Asphalt, Concrete, TileFloor) with FOV/explored memory tint (`ColorScale(0.2, 0.2, 0.3, 1)`).
    3. Pass 3: Depth-sorted sprite pass collecting obstacles/props, items, entities (player & zombies), and aim indicators, assigning `Depth = worldX + worldY`, sorted with `sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].Depth < sprites[j].Depth })`.
    4. Pass 4: Vector combat trails (quadratic Bezier curves and shotgun radial rays).
    5. Pass 5: Dynamic day/night ambient lighting overlay ($\alpha = 0.45 + 0.45\cos(2\pi \cdot timeOfDay / 24)$).
    6. Pass 6: Screen-space HUD / UI overlay (Health, Hunger, Thirst, Armor, Weapon/Ammo, Hotbar).
- **External Assets Survey (`context/`)**:
  - `context/Small Forest/Bench and chest/Bench.png` ($52\times37$ px)
  - `context/Small Forest/Bench and chest/Chest.png` ($22\times21$ px)
  - `context/Small Forest/Sculptures/Sculpture-1.png` ($23\times31$ px)
  - `context/Small Forest/Sculptures/Sculture-2.png` ($29\times32$ px)
  - `context/Small Forest/Bushes/` ($19\times15$ to $28\times19$ px)
  - `context/Small Forest/Stones/` ($28\times19$ to $29\times25$ px)
  - `context/Small Forest/Flowers/`, `context/Small Forest/Trees/`, `context/Small Forest/Fences/`
  - `context/Lab/Inside_C.png` ($768\times768$ px)
  - `context/Zombie Apocalypse Tileset/`
- **Game Lifecycle (`cmd/game/main.go`)**:
  - Entry point sets window size to $1280\times720$, window title "Go Zomboid", executes `assets.Load()`, instantiates `game.NewGame()` (which calls `assets.InitAudio()` and `g.Reset()`), and invokes `ebiten.RunGame(g)`.
- **Test Suite Execution**:
  - Running `CC=gcc go test ./...` passed across all packages (`cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`).
  - Total 22 test files cataloged across all 5 packages.
  - Critical test coupling observed: `internal/assets/empirical_challenger_test.go:303-357` has `TestEmpiricalGenerationDeterminism` which runs `exec.Command("go", "run", "./cmd/tools/genassets")`. When `cmd/tools/genassets` is deleted, this test will fail if not retired/updated.
  - `internal/game/world/world_empirical_stress_test.go:8-45` has `TestEmpirical_All10TileTypesGenerated` asserting non-zero occurrence for all 10 original `TileType` values (`Grass`, `Wall`, `Dirt`, `WoodFloor`, `Tree`, `Asphalt`, `Concrete`, `TileFloor`, `Fence`, `Debris`).

---

## 2. Logic Chain
1. *From Observation of `ORIGINAL_REQUEST.md` and `context/`*: The objective is to retire procedural generation (`cmd/tools/genassets`), import external PNG assets (Benches, Chests, Sculptures, etc.), define new `TileType` constants, place them in `NewMap()`, and depth-sort and render them in `DrawSystem`.
2. *From Observation of `game.go:929-1168`*: The rendering engine handles vertical objects through the `sprites` slice where every object has `Depth = worldX + worldY`. Since isometric Y is $(worldX + worldY)/2$, sorting by `worldX + worldY` guarantees correct isometric Y-sorting between new objects (Benches, Chests, Sculptures) and dynamic character entities (Player, Zombies).
3. *From Observation of `map.go:8-95`*: New tile constants `TileBench`, `TileChest`, `TileSculpture` must be added with `IsSolid() = true`, `BlocksVision() = false` (players can see over benches/chests/statues), and `IsFloor() = false` (rendered in sprite pass, not flat diamond pass).
4. *From Observation of `map.go:174-352` & `world_empirical_stress_test.go:8-45`*: When placing new objects in `NewMap()` (e.g. Benches in parks and residential yards, Chests in bedrooms and armory, Sculptures in town courtyards/parks), we must preserve non-zero counts of the original 10 TileTypes to keep `TestEmpirical_All10TileTypesGenerated` passing.
5. *From Observation of `empirical_challenger_test.go` and `assets_test.go`*: Asset tests currently check 27 procedural assets and `TestEmpiricalGenerationDeterminism` runs `genassets`. When implementing R1 and R2, asset loading in `assets.go` must load the new external PNGs natively, and tests must be updated to test native asset loading and exclude retired `genassets` binary execution.

---

## 3. Caveats
- No changes to source code or tests were made in this task (this agent operates strictly in survey/spec-mining mode).
- Context PNGs in `context/Small Forest/Sculptures/` have a slight spelling discrepancy in source filename (`Sculture-2.png` instead of `Sculpture-2.png`); target filenames in `internal/assets/images/` should use clean, canonical names (e.g. `sculpture_1.png`, `sculpture_2.png` or `sculpture.png`).
- Anchor translation offsets for newly imported sprites should align with the sprite bounding box base: `Translate(-W/2, -H + 32)`.

---

## 4. Conclusion
The rendering system, depth-sorting mechanism, game lifecycle, and test suite have been completely surveyed and documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_spec_miner_survey_3/survey.md`.
The implementation path is straightforward and robust:
1. Delete `cmd/tools/genassets/`.
2. Copy external PNGs from `context/` to `internal/assets/images/`, load in `internal/assets/assets.go`.
3. Add `TileBench`, `TileChest`, `TileSculpture` to `internal/game/world/map.go` with solid collision, non-blocking vision, and placement in `NewMap()`.
4. Add new tile types to `DrawSystem.Draw` with `Depth = worldX + worldY` and appropriate sprite transforms.
5. Update asset tests to reflect native asset loading without `genassets` execution.

---

## 5. Verification Method
1. Inspect the survey report:
   ```bash
   cat /home/bryce/code/go-zomboid/.agents/teamwork_preview_spec_miner_survey_3/survey.md
   ```
2. Verify all existing tests across the repository:
   ```bash
   CC=gcc go test -v ./...
   ```
3. Invalidation Conditions:
   - If any `TileType` lacks a string name or physical properties (`IsSolid`, `BlocksVision`, `IsFloor`).
   - If new world objects fail to depth-sort against moving player or zombie entities.
   - If `cmd/tools/genassets` remains after task completion.
