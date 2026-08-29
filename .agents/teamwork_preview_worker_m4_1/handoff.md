# Milestone 4 Handoff Report: Requirement R4 (Environmental Destruction & Resource Drops)

**Worker 3 (Milestone 4)**  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m4_1`  
**Date/Timestamp**: `2026-08-29T17:08:30Z`  
**Parent Agent ID**: `8fd0f6a8-cb46-4ae5-8024-c99358e741e1`

---

## 1. Observation

1. **Environmental Obstacle Infrastructure in `internal/game/world/map.go`**:
   - The map contains solid and destructible environmental obstacles: `TileFence` (residential yards, warehouse enclosures), `TileWall` (building partition and exterior walls), `TileTree`, `TileStump`, and `TileBench`.
   - Prior to Milestone 4, map cells did not track health or durability; obstacles were permanent and indestructible.
   - Perimeter border coordinates (`x == 0 || x == Width-1 || y == 0 || y == Height-1`) act as world boundaries to prevent entities from venturing out-of-bounds.

2. **Melee Combat and Weapon Attack Geometry in `internal/game/game.go`**:
   - Attacks trigger on `KeySpace`, `KeyX`, or `MouseButtonRight` with a 30-tick cooldown.
   - Fire Axe attacks cleave with reach `128.0px` and radius `128.0px`.
   - Club / standard weapon attacks reach `96.0px` with radius `96.0px`.
   - Shotgun blasts cover a spread cone up to `640.0px` range and point-blank radius `96.0px`.
   - Unarmed shove provides defensive pushback with 0 barrier damage.
   - Ground items are collected into the first empty slot of `player.Inventory` when within `64.0px` of the player (`processItems`).
   - `assets.WoodImage` is loaded from `images/wood.png` in `internal/assets/assets.go`.

3. **Verification Command & Output**:
   - Executing `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...` passed 100% of test suites with exit code 0.
   - Executing `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game` compiled successfully with zero warnings/errors.

---

## 2. Logic Chain

1. **Durability & Health Modeling (`world.Map`)**:
   - Added `TileDurability map[Point]int` to `world.Map` initialized in `NewMap(width, height)`.
   - Implemented `GetTileMaxDurability(t TileType) int`:
     - `TileFence`: 2 HP
     - `TileTree`: 3 HP
     - `TileStump`, `TileBench`: 2 HP
     - `TileWall`: 3 HP
     - Non-destructibles: 0 HP
   - Implemented `IsDestructible(x, y int) bool`:
     - Returns `false` for boundary perimeter coordinates (`x <= 0 || x >= m.Width-1 || y <= 0 || y >= m.Height-1`) and out-of-bounds cells.
     - Returns `true` for interior `TileFence`, `TileWall`, `TileTree`, `TileStump`, and `TileBench`.
   - Implemented `GetTileDurability(x, y int) int`:
     - Returns remaining durability from `TileDurability[Point{x, y}]` or defaults to `GetTileMaxDurability(m.GetTile(x, y))`.
   - Implemented `DamageTile(x, y int, amount int) (destroyed bool, dropType string)`:
     - Deducts durability by `amount`.
     - When durability reaches `<= 0`, deletes tracking entry, replaces `TileWall` with walkable `TileWoodFloor` or other destructible tiles with `TileGrass`, and returns `true, "wood"`.
     - Because `TileGrass` and `TileWoodFloor` have `IsSolid() == false` and `BlocksVision() == false`, physical collision and vision occlusions are cleared immediately upon destruction.

2. **Melee Barrier Chopping Integration (`internal/game/game.go`)**:
   - In `processInputAndCombat`:
     - **Axe Attack**: Cleave reach `128.0px`, radius `128.0px`, damage `2`. Checks all destructible tiles within the cleave sweep, applies damage, plays `HitSound`, decrements `player.WeaponDurability`, and upon destruction instantiates `ecs.Item{Type: "wood"}` at the center of the destroyed tile `(float64(tx)*TileSize + 64.0, float64(ty)*TileSize + 64.0)`.
     - **Club/Weapon Attack**: Reach `96.0px`, radius `96.0px`, damage `1`. Checks destructible tiles within attack radius, applies damage (taking 2 swings for a fence), plays `HitSound`, decrements durability, and drops wood on destruction.
     - **Shotgun Attack**: Firing shotgun with ammo deals `2` damage to destructible tiles within the spread cone / point-blank radius, dropping wood on destruction.
     - **Unarmed Shove**: Deals `0` damage and leaves barrier durability intact.
     - **Weapon Durability & Breaking**: When weapon durability drops to 0 after chopping barriers, weapon unequips to fists (`WeaponEquipped = false`, `WeaponType = ""`, `WeaponDurability = 0`).

3. **Resource Collection & Rendering**:
   - `processItems()` detects dropped `"wood"` items when player moves within `64.0px`, inserting `"wood"` into the first available inventory slot and removing the entity from `world`.
   - `DrawSystem.Draw` maps `case "wood": img = assets.WoodImage`, seamlessly rendering the dropped wood item on the ground.
   - HUD inventory slot displays `"wood"` and tooltips render correctly.

---

## 3. Caveats

- **Perimeter Boundaries**: Perimeter boundary walls are strictly indestructible (`IsDestructible` returns `false`), ensuring players and zombies cannot escape the map grid into unrendered memory space.
- **Weapon Breaking**: Chopping barriers consumes weapon durability on each successful swing that strikes a barrier, accurately reflecting tool wear.
- **Autotiling Dynamics**: Because autotiling queries live map tiles per frame, destroying a fence or wall automatically causes adjacent wall/fence segments to redraw with proper terminating endcaps without requiring manual cache invalidation.

---

## 4. Conclusion

Requirement R4 (Environmental Destruction & Resource Drops) is fully implemented with 100% genuine logic. Players can chop down wooden fences, building walls, trees, stumps, and benches using axes, clubs, or shotguns. Destroyed barriers immediately clear collision and line-of-sight blocking, instantiate `"wood"` resource drops at the tile center, and allow players to collect wood into their inventory upon stepping within 64px.

All unit tests and integration tests pass cleanly, and the project binary compiles without errors.

---

## 5. Verification Method

1. **Run Full Test Suite**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...
   ```
   *Result*: 100% tests pass (including `internal/game/world/destruction_test.go` and `internal/game/destruction_combat_test.go`).

2. **Build Game Binary**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
   ```
   *Result*: Binary builds successfully at `bin/game`.

3. **Inspect Modified Source Files**:
   - `internal/game/world/map.go`: Durability tracking, `IsDestructible`, `GetTileMaxDurability`, `GetTileDurability`, `DamageTile`.
   - `internal/game/game.go`: Melee chopping in `processInputAndCombat`, `assets.WoodImage` item rendering.
   - `internal/game/world/destruction_test.go`: 9 comprehensive unit tests for map durability and destruction.
   - `internal/game/destruction_combat_test.go`: 7 comprehensive combat and breach traversal tests.
