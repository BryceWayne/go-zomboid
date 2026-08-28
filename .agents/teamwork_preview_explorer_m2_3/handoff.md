# Handoff Report: Milestone 2 — Tile System Expansion, Collision/FOV & Thematic Spawning Integration

## 1. Observation

### 1.1 Existing Assets and Image FS
From `internal/assets/assets.go:16-74`:
- Floor tile handles: `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage` (all 64x32 isometric floor diamonds).
- Vertical obstacle / prop handles: `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage` (all 64x64 isometric blocks).
- Item handles: `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `FoodImage`, `WaterImage` (all 16x16 pixel sprites).
- All image files exist in `internal/assets/images/*.png` and load successfully via `assets.Load()`.

### 1.2 Existing World Map Implementation
From `internal/game/world/map.go:10-184`:
- Tile types currently restricted to 5 constants: `TileGrass` (0), `TileWall` (1), `TileDirt` (2), `TileWoodFloor` (3), `TileTree` (4).
- `NewMap(width, height)` generates a static 3-tile wide crossroad of `TileDirt`, 7 hardcoded single-room houses with wood floors, and random tree scatter.
- `IsColliding` hardcodes checks against `TileWall` and `TileTree`. It lacks checks for `TileFence` or `TileDebris`.
- `CalculateFOV` checks `m.GetTile(tx, ty) == TileWall` to break raycasts.
- The `Map` struct lacks spawn metadata (`PlayerSpawn`, `Buildings`, `LootSpawns`, `ZombieSpawns`).

### 1.3 Existing Game Loop and ECS Spawning
From `internal/game/game.go:34-120` and `588-725`:
- Player spawn is hardcoded to center coordinates `(50.0 * 32, 50.0 * 32) = (1600.0, 1600.0)`.
- Loot items (20 items: weapons, food, water) and zombies (150 entities) are spawned at unvalidated uniform random positions `(100 + rand.Intn(3000), 100 + rand.Intn(3000))`, causing clipping into walls and trees.
- `DrawSystem.Draw` only renders ground diamonds for `TileGrass`, `TileDirt`, and `TileWoodFloor`, and only renders vertical obstacles for `TileWall` and `TileTree`. Item rendering only switches between `"food"`, `"water"`, and generic `"weapon"`.

---

## 2. Logic Chain

1. **Tile System Expansion** (`internal/game/world/map.go`):
   - Adding `TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileFence`, `TileDebris` (total 10 tile types) directly matches the assets generated in M1 (`asphalt.png`, `concrete.png`, `tile_floor.png`, `fence.png`, `debris.png`).
   - Defining `IsSolid() bool` returns `true` for `TileWall`, `TileTree`, `TileFence`, `TileDebris`, enabling universal collision detection in `IsColliding(rectX, rectY, rectW, rectH)`.
   - Defining `BlocksVision() bool` returns `true` only for `TileWall`, enabling player visibility through fences (`TileFence`) and over debris piles (`TileDebris`) while blocking raycasts through solid walls.
   - Defining `IsFloor() bool` distinguishes flat ground diamond tiles (`TileGrass`, `TileDirt`, `TileWoodFloor`, `TileAsphalt`, `TileConcrete`, `TileTileFloor`) from vertical props.

2. **Procedural Town Generation & Multi-Room Archetypes** (`internal/game/world/map.go`):
   - A multi-tier road network: 5-tile wide East-West and North-South main thoroughfares with asphalt centers and concrete sidewalks, plus secondary residential and industrial access streets and dirt park pathways.
   - Four distinct thematic quadrants with multi-room building archetypes:
     - **North-West (Residential)**: Suburban houses with Living Rooms (`TileWoodFloor`), Kitchens (`TileTileFloor`), internal partition walls, doorways, and fenced backyards (`TileFence`).
     - **North-East (Commercial)**: Grocery Store (store floor + back storage) and Pharmacy/Clinic with concrete plazas.
     - **South-West (Municipal/Defense)**: Police Station with Front Office and fortified Armory (`TileConcrete`), plus fenced motor pool.
     - **South-East (Industrial)**: Large Warehouse with concrete floor, internal debris/crates (`TileDebris`), and industrial perimeter fencing.
   - Outdoor props: Debris clusters in alleys/depots and tree clusters in neighborhood parks and yards.

3. **Contextual ECS Spawning Architecture**:
   - `Map` stores `PlayerSpawn FloatPoint`, `Buildings []Building`, `LootSpawns []LootSpawn`, and `ZombieSpawns []FloatPoint`.
   - **Safe Player Spawn**: Placed inside House 1 living room, guaranteed walkable (`!IsSolid()`) and $\ge 350\text{ px}$ from any zombie spawn.
   - **Thematic Loot Spawning**:
     - Residential Kitchens / Grocery: `"food"`, `"water"`.
     - Police Armory / Warehouse: `"weapon"`, `"axe"`, `"shotgun"`, `"ammo"`, `"armor"`.
     - Pharmacy: `"food"`, `"water"`, `"armor"`.
     - Guaranteed starting items placed directly in player starting house.
   - **Safe Zombie Spawns**: 140+ zombies distributed on validated walkable tiles (`!IsSolid()`), with Euclidean distance $>350\text{ px}$ from player spawn, eliminating wall clipping and unfair insta-death.

4. **Rendering Pipeline Integration** (`internal/game/game.go`):
   - Ground diamond pass handles all 6 floor types (`TileGrass`, `TileDirt`, `TileWoodFloor`, `TileAsphalt`, `TileConcrete`, `TileTileFloor`).
   - Depth-sorted vertical pass renders `WallImage`, `TreeImage`, `FenceImage`, and `DebrisImage` sorted by depth $wx + wy$.
   - Item pass maps item types `"food"`, `"water"`, `"weapon"`, `"axe"`, `"shotgun"`, `"ammo"`, `"armor"` to their corresponding sprite handles in `internal/assets`.

---

## 3. Caveats

1. The procedural generator currently uses a structured deterministic layout with pseudo-random prop and zombie distributions to guarantee building connectivity, room access, and safe spawns on a 100x100 grid. Small fallback generation is included for $<30\times 30$ maps.
2. Weapon combat mechanics for Axe, Shotgun, Ammo, and Armor damage mitigation are designed in M3/M4; in M2, their item entities and sprites are populated, rendered, and collectable in the 9-slot inventory.
3. No external dependencies or CGO requirements were introduced; all code is standard pure Go.

---

## 4. Conclusion

The design for Milestone 2 is complete, fully tested, and verified via `proposed_map.go`, `proposed_map_test.go`, and `proposed_game_patch.go`.
Applying these proposed changes will:
1. Provide full support for all 10 tile types with correct physical collision (`IsSolid`) and FOV occlusion (`BlocksVision`).
2. Generate an immersive town with distinct residential, commercial, police, and industrial districts featuring multi-room interiors, roads, sidewalks, fences, debris, and trees.
3. Establish safe, contextual ECS entity spawning (safe player house start, room-appropriate loot distribution, zero wall-clipped zombies).
4. Render all floor diamonds, vertical obstacles, and expanded item sprites with isometric projection and Y-depth sorting.

---

## 5. Verification Method

### 5.1 Unit and Integration Test Verification
Run the Go test suite across the whole project:
```sh
CC=gcc go test -v ./...
```
Expected output:
- `internal/game/world`: All tests (`TestTileTypeProperties`, `TestNewMapProceduralTown`, `TestPlayerSafeSpawn`, `TestContextualLootSpawns`, `TestZombieSpawnsNoTrapping`, `TestCollisionDetection`, `TestFOVAndOcclusion`) pass with 0 failures.
- `internal/game`: `TestWorldToIso`, `TestNewGameInitialization`, and `TestGameResetContextualSpawns` pass with 0 failures.

### 5.2 Standalone Map Verification
Run the standalone test suite in the agent folder:
```sh
go test -v /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3/proposed_map.go /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3/proposed_map_test.go
```
Expected output:
```
=== RUN   TestTileTypeProperties
--- PASS: TestTileTypeProperties (0.00s)
=== RUN   TestNewMapProceduralTown
--- PASS: TestNewMapProceduralTown (0.00s)
=== RUN   TestPlayerSafeSpawn
--- PASS: TestPlayerSafeSpawn (0.00s)
=== RUN   TestContextualLootSpawns
--- PASS: TestContextualLootSpawns (0.00s)
=== RUN   TestZombieSpawnsNoTrapping
--- PASS: TestZombieSpawnsNoTrapping (0.00s)
=== RUN   TestCollisionDetection
--- PASS: TestCollisionDetection (0.00s)
=== RUN   TestFOVAndOcclusion
--- PASS: TestFOVAndOcclusion (0.00s)
PASS
```

### 5.3 Interactive Game Loop Verification
```sh
CC=gcc go run ./cmd/game
```
Expected output:
- Window opens at 800x600.
- Player spawns safely inside the bedroom/living room of a residential house.
- Roads (asphalt + concrete sidewalks), multi-room stores, police station, warehouse, fences, trees, and debris render cleanly in isometric view.
- Items (food, water, axe, shotgun, ammo, armor, weapon) spawn in thematic rooms and render with accurate sprites.
- Zombies spawn outdoors and inside open facilities without trapping inside solid walls or fences.
