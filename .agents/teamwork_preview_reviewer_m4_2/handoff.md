# Milestone 4 Review Report: Requirement R4 (Environmental Destruction & Resource Drops)

**Reviewer 2 (Milestone 4)**  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_reviewer_m4_2`  
**Date/Timestamp**: `2026-08-29T17:11:30Z`  
**Parent Agent ID**: `8fd0f6a8-cb46-4ae5-8024-c99358e741e1`

---

## 1. Observation

1. **Environmental Destruction Engine Logic (`internal/game/world/map.go`)**:
   - `TileDurability map[Point]int` accurately tracks barrier damage and remaining durability.
   - `GetTileMaxDurability` sets standard HP: `TileFence` (2 HP), `TileWall` (3 HP), `TileTree` (3 HP), `TileStump` (2 HP), `TileBench` (2 HP), non-destructibles (0 HP).
   - `IsDestructible(x, y)` explicitly guards map boundaries (`x <= 0 || x >= m.Width-1 || y <= 0 || y >= m.Height-1`), ensuring world perimeter boundaries remain strictly indestructible.
   - `DamageTile(x, y, amount)` reduces tile durability, deletes the map entry on destruction, replaces `TileWall` with walkable `TileWoodFloor` or other obstacles with `TileGrass`, and returns `destroyed = true, dropType = "wood"`.

2. **Combat Integration & Resource Spawning (`internal/game/game.go`)**:
   - Axe swings (128.0px reach & radius) deal 2 damage to all destructible tiles within the arc, destroying fences in 1 swing and dropping wood items.
   - Club/standard weapon swings (96.0px reach & radius) deal 1 damage, destroying fences across 2 consecutive hits.
   - Shotgun blasts with ammo deal 2 damage across the 640.0px spread cone / point-blank radius to destructible tiles.
   - Unarmed shoves apply 0 damage, leaving barriers untouched.
   - Hitting barriers decrements weapon durability, properly breaking/unequipping weapons to fists upon reaching 0 HP.
   - Destroyed barriers spawn `ecs.Item{Type: "wood"}` entities at the center of the destroyed tile `(float64(tx)*TileSize + 64.0, float64(ty)*TileSize + 64.0)`.
   - `processItems()` collects `"wood"` into the first empty slot of `player.Inventory` when within 64.0px.
   - `DrawSystem.Draw` renders dropped wood items using `assets.WoodImage` and renders `"wood"` in HUD inventory slots.

3. **Dynamic Autotiling & Vision Occlusion Clearing**:
   - Because terrain transitions (`GetTileTransitions`) and cardinal bitmasks (`GetWallBitmask`, `GetFenceBitmask`) are calculated directly from live `Map.Tiles` during rendering, destroying a barrier instantly updates neighbor connectivity (e.g. converting a T-junction or segment into a terminating endcap) and blends newly revealed floor tiles without stale cache artifacts.
   - Solid collision (`IsColliding`) and raycasted FOV occlusion (`CalculateFOV`) immediately allow entities and vision to penetrate through breached tiles.

4. **Test & Build Execution**:
   - Ran `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v ./...`: Passed 100% of tests across all packages.
   - Ran `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game`: Compiled successfully with exit code 0.

---

## 2. Logic Chain

1. **Integrity & Authenticity Audit**:
   - Inspected implementation in `internal/game/world/map.go`, `internal/game/game.go`, `internal/game/world/destruction_test.go`, and `internal/game/destruction_combat_test.go`.
   - Confirmed zero hardcoded test outputs, zero facade/dummy methods, and zero external bypasses. All durability calculations, collision mutations, ECS item entity instantiations, and weapon degradation logic are genuine.

2. **Adversarial Edge Case Analysis**:
   - **Boundary Safety**: Attacking perimeter tiles (0, y), (x, 0), (Width-1, y), (x, Height-1) returns `destroyed = false, dropType = ""` and leaves tiles intact (`TileWall`), preventing entities from escaping bounds.
   - **Multiple Adjacent Barriers**: Cleaving multiple adjacent barriers (such as fence lines) with an Axe damages all barriers within the 128px radius, spawns separate `"wood"` drops at each tile's center, decrements weapon durability once per swing, and allows passage through breached tiles while adjacent intact tiles continue to block collision.
   - **Inventory Collection Invariants**: Spawning wood drops and walking over them places `"wood"` into the player's inventory. If all 9 slots are full, items remain safely on the ground without being deleted or corrupted.
   - **Weapon Breaking**: Chopping barriers to 0 weapon durability immediately unsets `WeaponEquipped`, resets `WeaponType = ""`, and resets `WeaponDurability = 0`, smoothly reverting the player to unarmed combat.

3. **Findings & Observations**:
   - **Finding (Minor - Test Isolation)**: In `internal/game/destruction_combat_test.go`, `setupDestructionTestHarness()` calls `world.NewMap(50, 50)` which generates random procedural props (trees/stumps). In `TestCombat_ShotgunBlastDestroysBarriers`, the 640px shotgun cone covers a wide grid area. If random procedural generation happens to place a 2-HP prop within the cone, `destroyedCount` counts both the test fence and the procedural prop (yielding 2 instead of 1).
   - *Recommendation for future test cleanup*: Pre-clear the shotgun cone region to `TileGrass` before placing the test fence in `TestCombat_ShotgunBlastDestroysBarriers` or assert `destroyedCount >= 1 && m.GetTile(targetTx, targetTy) == world.TileGrass`. Note: This is an isolated unit test harness condition and does not affect production game logic.

---

## 3. Caveats

- **Test Scope**: Verified all packages with `go test -v ./...` and `go build ./cmd/game`.
- **Procedural Map State**: The game map is generated with procedural props; non-solid floor tiles under destroyed walls become `TileWoodFloor` to preserve interior building aesthetics, while exterior destructibles become `TileGrass`.

---

## 4. Conclusion

**Verdict: APPROVE**

The implementation of Milestone 4: Requirement R4 (Environmental Destruction & Resource Drops) fulfills all functional requirements and acceptance criteria. Durability degradation, weapon chopping reach/radius, perimeter wall protection, collision and vision clearing, resource drop spawning, inventory collection, weapon durability wear, and dynamic autotiling recalculations operate correctly and robustly.

---

## 5. Verification Method

To independently verify this review:

1. **Run Full Test Suite**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v ./...
   ```
   *Expected Result*: All tests pass with exit code 0.

2. **Run Fresh (Uncached) Test Suite**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...
   ```
   *Expected Result*: 100% tests pass.

3. **Build Binary**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
   ```
   *Expected Result*: Compiles cleanly to `bin/game`.

4. **Inspect Source Files**:
   - `internal/game/world/map.go`: Lines 1153–1225 (`IsDestructible`, `GetTileMaxDurability`, `GetTileDurability`, `DamageTile`).
   - `internal/game/game.go`: Lines 745–784 (shotgun destruction), lines 833–862 (axe chopping), lines 894–923 (club chopping), lines 408–430 (item collection), lines 1466–1468 & 1702–1720 (rendering & HUD).
   - `internal/game/world/destruction_test.go` & `internal/game/destruction_combat_test.go`.
