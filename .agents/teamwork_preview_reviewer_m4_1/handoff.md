# Milestone 4 Review Report: Requirement R4 (Environmental Destruction & Resource Drops)

**Reviewer 1**  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_reviewer_m4_1`  
**Date/Timestamp**: `2026-08-29T17:10:30Z`  
**Verdict**: **APPROVE**  
**Parent Agent ID**: `8fd0f6a8-cb46-4ae5-8024-c99358e741e1`

---

## 1. Observation

1. **Map Durability and Destructibility Infrastructure (`internal/game/world/map.go`)**:
   - `Map` struct contains `TileDurability map[Point]int` (line 192), initialized in `NewMap` (line 207).
   - `GetTileMaxDurability(t TileType) int` (lines 1170-1184):
     - `TileFence`: 2 HP
     - `TileTree`: 3 HP
     - `TileStump`: 2 HP
     - `TileBench`: 2 HP
     - `TileWall`: 3 HP
     - Other tile types: 0 HP
   - `IsDestructible(x, y int) bool` (lines 1154-1168):
     - Perimeter boundary protection: `if x <= 0 || x >= m.Width-1 || y <= 0 || y >= m.Height-1 { return false }`
     - Returns `true` for interior `TileFence`, `TileTree`, `TileStump`, `TileBench`, and `TileWall`.
     - Returns `false` for indestructible terrain (`TileGrass`, `TileDirt`, `TileConcrete`, `TileChest`, etc.).
   - `GetTileDurability(x, y int) int` (lines 1186-1199):
     - Checks `m.TileDurability[p]`; falls back to `GetTileMaxDurability(m.GetTile(x, y))`.
   - `DamageTile(x, y int, amount int) (destroyed bool, dropType string)` (lines 1201-1225):
     - Validates `!m.IsDestructible(x, y) || amount <= 0`.
     - Decrements durability by `amount`.
     - On `<= 0`, deletes the tracking point, converts `TileWall` to `TileWoodFloor` and other destructible tiles to `TileGrass`, and returns `true, "wood"`.
     - Because `TileGrass` and `TileWoodFloor` are non-solid and non-vision-blocking (`IsSolid() == false`, `BlocksVision() == false`), collision and line-of-sight are immediately unblocked.

2. **Melee Chopping, Weapon Durability, and Resource Spawning (`internal/game/game.go`)**:
   - **Fire Axe** (lines 814-874):
     - Cleave reach `128.0px`, radius `128.0px`, damage `2`.
     - Calls `DamageTile(tx, ty, 2)` (line 850).
     - Decrements `player.WeaponDurability--`.
     - Unequips to fists if durability drops to `<= 0` (`WeaponEquipped = false`, `WeaponType = ""`, `WeaponDurability = 0`).
     - Spawns `ecs.Item{Type: dropType}` at `(float64(tx)*128 + 64, float64(ty)*128 + 64)`.
   - **Club / Melee Weapon** (lines 875-935):
     - Reach `96.0px`, radius `96.0px`, damage `1`.
     - Calls `DamageTile(tx, ty, 1)` (line 911).
     - Decrements `player.WeaponDurability--` and unequips on `<= 0`.
     - Spawns `ecs.Item{Type: dropType}` on destruction.
   - **Shotgun Blast** (lines 745-784):
     - Spread cone up to `640.0px` range, damage `2`.
     - Calls `DamageTile(tx, ty, 2)` (line 772).
     - Spawns `ecs.Item{Type: dropType}` on destruction.
   - **Unarmed Shove** (lines 936-953):
     - Deals `0` damage to barriers, leaving durability intact.
   - **Item Pickup & Rendering**:
     - `processItems()` (lines 385-430) gathers dropped `"wood"` into the first empty inventory slot when player is within `64.0px`.
     - `DrawSystem.Draw` (lines 1465-1467) renders `assets.WoodImage` for `"wood"` items.

3. **Verification Command Results**:
   - `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...`: 100% tests pass (all unit, combat, stress, autotiling, and empirical test suites).
   - `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game`: Binary compiles cleanly with exit code 0.

---

## 2. Logic Chain

1. **Durability Model & Perimeter Boundary Protection**:
   - `IsDestructible(x, y)` explicitly guards boundary coordinates `(x <= 0 || x >= Width-1 || y <= 0 || y >= Height-1)` before inspecting tile type.
   - This ensures boundary walls remain indestructible, preventing entities from falling out of bounds into unrendered memory space.
   - Interior obstacle HP values (`TileFence=2`, `TileTree=3`, `TileStump=2`, `TileBench=2`, `TileWall=3`) match tactical survival design.

2. **Physical Collision & Vision Clearing**:
   - Destroyed walls convert to `TileWoodFloor`; destroyed outdoor obstacles convert to `TileGrass`.
   - `TileWoodFloor.IsSolid() == false` and `TileGrass.IsSolid() == false`, clearing `IsColliding` immediately.
   - `TileWoodFloor.BlocksVision() == false` and `TileGrass.BlocksVision() == false`, allowing FOV raycasting to pass through breached obstacles.

3. **Attack Geometry & Weapon Balancing**:
   - Axe deals 2 damage (destroying fences in 1 hit, walls/trees in 2 hits).
   - Club deals 1 damage (destroying fences in 2 hits, walls/trees in 3 hits).
   - Shotgun deals 2 damage across spread cone.
   - Unarmed shove deals 0 damage.
   - Weapon durability properly decrements upon hitting obstacles and unequips when depleted.

4. **Resource Drops & Collection Lifecycle**:
   - On destruction, `DamageTile` returns `true, "wood"`.
   - ECS item entity with component `ecs.Item{Type: "wood"}` is created at the destroyed tile center.
   - Player stepping within 64px triggers `processItems()`, transferring `"wood"` into the player's 9-slot inventory and removing the ground item entity.
   - `DrawSystem` renders `assets.WoodImage` at the item's world position.

5. **Forensic Integrity Verification**:
   - No hardcoded test responses or bypasses.
   - All tests use genuine procedural maps, ECS world components, and simulated combat cycles.

---

## 3. Adversarial Challenges & Stress Testing

| # | Adversarial Scenario / Stress Test | Expected Behavior | Actual Behavior | Result |
|---|---|---|---|---|
| 1 | Attack with zero or negative damage (`DamageTile(x, y, 0)` / `DamageTile(x, y, -5)`) | No change to tile durability, returns `false, ""` | Intact durability, returns `false, ""` | **PASS** |
| 2 | Massive damage to perimeter boundary (`DamageTile(0, 20, 999)`) | Tile remains `TileWall`, returns `false, ""` | Boundary remains indestructible | **PASS** |
| 3 | Out-of-bounds coordinate attacks (`DamageTile(-10, 50, 5)`) | Safe handling without panic or memory corruption | Returns `false, ""` safely | **PASS** |
| 4 | Excess single-hit damage (`DamageTile(x, y, 10)` on 2 HP fence) | Fence destroyed in 1 hit, drops wood, replaces with grass | Replaced with `TileGrass`, drops `"wood"` | **PASS** |
| 5 | Multi-barrier breach and corridor traversal | Player blocked by intact fence $\to$ destroys obstacle $\to$ walks through breach $\to$ collects wood drop $\to$ continues past breach | 100% collision and pickup invariants pass | **PASS** |
| 6 | Weapon durability depletion to 0 via barrier chopping | Weapon durability decrements to 0, weapon unequips to fists (`WeaponEquipped=false`, `WeaponType=""`, `WeaponDurability=0`) | Successfully unequips to unarmed fists | **PASS** |
| 7 | Unarmed attack against barrier | Shove plays sound, barrier durability remains 100% intact | Barrier takes 0 damage | **PASS** |

---

## 4. Caveats

- **Perimeter Boundaries**: Perimeter boundaries are intentionally and strictly indestructible to preserve world confinement.
- **Weapon Wear**: Striking barriers consumes weapon durability per successful swing, correctly enforcing resource management.
- **No caveats** impacting functionality or release readiness.

---

## 5. Conclusion

**Verdict**: **APPROVE**

Milestone 4 (Requirement R4: Environmental Destruction & Resource Drops) is completely and correctly implemented. All acceptance criteria and interface contracts are satisfied with zero regressions, zero integrity violations, and full test suite validation.

---

## 6. Verification Method

To independently verify this implementation:

1. **Run Full Test Suite**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...
   ```
   *Expected Output*: 100% pass across all packages (`PASS ok github.com/BryceWayne/go-zomboid/...`).

2. **Build Game Executable**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
   ```
   *Expected Output*: Exit code 0, binary created at `bin/game`.

3. **Inspect Core Files**:
   - `internal/game/world/map.go`: Durability tracking, `IsDestructible`, `DamageTile`, perimeter boundaries.
   - `internal/game/game.go`: Melee chopping logic, weapon damage values, wood item rendering.
   - `internal/game/world/destruction_test.go`: Durability unit tests.
   - `internal/game/destruction_combat_test.go`: Combat chopping, breach traversal, and item collection tests.
