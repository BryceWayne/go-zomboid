# Handoff Report: World Map & Tile System Technical Survey (R3)

**Agent:** `teamwork_preview_explorer_survey_2`  
**Working Directory:** `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2`  
**Task:** In-depth technical survey of world map, tile systems, asset mapping, and rendering depth sorting for Requirement R3.

---

## 1. Observation

1. **World Map Structure (`internal/game/world/map.go:160-173`):**
   ```go
   type Map struct {
       Width, Height int
       Tiles         []TileType
       Visible       []bool
       Explored      []bool
       PlayerSpawn   FloatPoint
       Buildings     []Building
       LootSpawns    []LootSpawn
       ZombieSpawns  []FloatPoint
   }
   ```
   - Standard map dimension is 100x100 tiles (`TileSize = 128`).
   - Flat 1D slices indexed via `idx = y * Width + x`.
   - Bounding queries (`GetTile`, `IsColliding`) guard against out-of-bounds access.

2. **Existing `TileType` Enum & Methods (`internal/game/world/map.go:8-95`):**
   - Currently defines 16 constants: `TileGrass` (0) through `TileSign` (15).
   - `IsSolid()` returns `true` for `TileWall`, `TileTree`, `TileFence`, `TileDebris`, `TileTent`, `TileElevationBlock`, `TileStump`, `TileSign`.
   - `BlocksVision()` returns `true` exclusively for `TileWall` (line 44).
   - `IsFloor()` returns `true` for flat terrain: `TileGrass`, `TileDirt`, `TileWoodFloor`, `TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileRamp` (line 50).

3. **External Assets in `context/` (`find context -name "*.png"`):**
   - `Small Forest/Bench and chest/Bench.png` (52x37 RGBA)
   - `Small Forest/Bench and chest/Chest.png` (22x21 RGBA)
   - `Small Forest/Sculptures/Sculpture-1.png` (23x31 RGBA), `Sculture-2.png` (29x32 RGBA)
   - `Small Forest/Bushes/Bush-1.png` .. `Bush-4.png` (19x15 to 28x19 RGBA), `Stump.png` (29x19 RGBA)
   - `Small Forest/Flowers/Flower-1.png` .. `Flower-3.png` (24x22 to 26x25 RGBA)
   - `Small Forest/Stones/Stone-1.png` (28x19 RGBA), `Stone-2.png` (29x25 RGBA)
   - `Small Forest/Ground tileset/`, `Trees/`, `Fences/`
   - `Lab/Inside_C.png` (768x768)
   - `Zombie Apocalypse Tileset/Organized separated sprites/` (Urban Assets, Modular Barns, Roads, Pickable Items, Character/Zombie sheets)

4. **Rendering & Depth-Sorting Pipeline (`internal/game/game.go:879-1168`):**
   - Pass 1 (lines 879-927): Draws flat floor diamonds using `assets.GrassImage`, `assets.DirtImage`, `assets.WoodImage`, etc.
   - Pass 2 (lines 937-1168): Gathers all vertical obstacle tiles, world items, and character/zombie entities into a `sprites []Renderable` slice with `Depth = worldX + worldY`, then sorts via `sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].Depth < sprites[j].Depth })`.

5. **Existing World & Map Tests (`internal/game/world/map_test.go`, `world_empirical_stress_test.go`):**
   - `map_test.go` verifies `solidTiles`, `nonSolidTiles`, `floorTiles`, `BlocksVision`, `String()`, `TestNewMapProceduralTown`, `TestPlayerSafeSpawn`, `TestContextualLootSpawns`, `TestZombieSpawnsNoTrapping`, `TestCollisionDetection`, `TestFOVAndOcclusion`.
   - `world_empirical_stress_test.go` tests that the first 10 core tile types are generated in `NewMap(100, 100)` with count $> 0$.
   - `game_stress_test.go:245-384` stress-tests rendering across 24h lighting cycles, fog of war, and dead player states.

---

## 2. Logic Chain

1. **From Observation 1 & 2:** The game world uses an integer `TileType` abstraction where each tile governs physics (`IsSolid`), raycast occlusion (`BlocksVision`), and render staging (`IsFloor`).
2. **From Observation 3 & Requirement R3:** The user request explicitly names Benches, Chests, and Sculptures as new assets that must be mapped into the game world as new `TileType` constants. Expanding this with Bushes, Flowers, and Stones provides complete coverage of the `Small Forest` asset pack.
3. **From Observation 2 & 4:** To render these objects seamlessly without visual artifacts or holes in the terrain mesh:
   - In Pass 1 (Ground diamond pass), the cell underneath each prop must be drawn as `assets.GrassImage` (or `ConcreteImage`).
   - In Pass 2 (Sprite pass), each new tile type must be added to the sprite collector with `Depth = worldX + worldY` and mapped to `assets.<Name>Image`.
4. **From Observation 1 & 5:** Placing these props in `internal/game/world/map.go` within `placeEnvironmentalProps` allows town parks, sidewalks, yards, and warehouses to feature sculptures, benches, chests, bushes, flowers, and stones without breaking building footprints, road corridors, or safe player/zombie spawn distances ($> 1400$px).
5. **From Observation 5:** Adding new `TileType` constants with IDs 16..21 maintains 100% backward compatibility with all existing test suites while enabling full visual and physical simulation.

---

## 3. Caveats

- **Asset Test Dependency on `genassets`:** In `internal/assets/empirical_challenger_test.go:336`, there is an existing test executing `go run ./cmd/tools/genassets`. When `cmd/tools/genassets` is retired per Requirement R1, this test will fail unless updated to test native image loading instead of invoking `genassets`.
- **Tileset Sheet Slicing vs Single PNGs:** While `Bench.png`, `Chest.png`, `Sculpture-1.png`, etc. are discrete PNGs, `Lab/Inside_C.png` is a 768x768 multi-tile sheet. For R3, discrete prop PNGs should be imported directly into `internal/assets/images/`.
- No assumptions made outside the codebase.

---

## 4. Conclusion

Requirement R3 can be satisfied cleanly by:
1. Adding 6 new `TileType` constants to `internal/game/world/map.go`: `TileBench` (16), `TileChest` (17), `TileSculpture` (18), `TileBush` (19), `TileFlower` (20), `TileStone` (21).
2. Implementing the property methods: `TileBench`, `TileChest`, `TileSculpture`, `TileStone` as solid (`IsSolid() == true`); `TileBush`, `TileFlower` as walkable (`IsSolid() == false`); all transparent to vision (`BlocksVision() == false`) and non-floor (`IsFloor() == false`).
3. Updating `placeEnvironmentalProps` in `map.go` to generate sculptures in town plazas, benches on sidewalks/yards, chests in warehouses/campsites, and natural bushes/flowers/stones across parklands.
4. Updating `internal/assets/assets.go` to load `BenchImage`, `ChestImage`, `SculptureImage`, `BushImage`, `FlowerImage`, `StoneImage`.
5. Updating `DrawSystem.Draw` in `internal/game/game.go` to render underlying ground in Pass 1 and depth-sort props in Pass 2.

The comprehensive survey report has been written to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey.md`.

---

## 5. Verification Method

1. **Survey Artifact Inspection:**
   - Inspect `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey.md` for full technical details, property tables, coordinate transformations, and implementation plans.
2. **Current Test Suite Baseline Execution:**
   - Run: `CC=gcc go test -v -count=1 ./internal/game/world/... ./internal/game/...`
   - Invalidation condition: Any failure in `map_test.go` or `world_empirical_stress_test.go`.
