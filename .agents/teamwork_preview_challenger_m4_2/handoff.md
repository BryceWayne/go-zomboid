# Milestone 4 Adversarial Challenge Report: Requirement R4 (Environmental Destruction & Resource Drops)

**Challenger 2 (Milestone 4)**  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_challenger_m4_2`  
**Date/Timestamp**: `2026-08-29T17:11:45Z`  
**Parent Agent ID**: `8fd0f6a8-cb46-4ae5-8024-c99358e741e1`  
**Final Verdict**: **APPROVE**

---

## 1. Observation

1. **Environmental Destruction & Combat Implementation**:
   - `internal/game/world/map.go`:
     - `TileDurability map[Point]int` tracks health across destructible barrier instances.
     - `IsDestructible(x, y int) bool` strictly guards boundary coordinates (`x <= 0 || x >= Width-1 || y <= 0 || y >= Height-1`), returning `false` for perimeter walls and interior non-destructibles, while returning `true` for interior fences, walls, trees, stumps, and benches.
     - `DamageTile(x, y int, amount int) (destroyed bool, dropType string)` decrements durability, cleans up map state on 0 HP, converts `TileWall` to walkable `TileWoodFloor` or other obstacles to `TileGrass`, and returns `true, "wood"`.
   - `internal/game/game.go`:
     - `processInputAndCombat`: Handles Fire Axe (cleave range `128.0px`, damage `2`), Club/Melee (range `96.0px`, damage `1`), Shotgun blast (cone `640.0px`, damage `2`), and Unarmed shove (damage `0`). On tile destruction, spawns `ecs.Item{Type: dropType}` at tile center `(tx*128 + 64, ty*128 + 64)`.
     - `processItems`: Collects ground items within `64.0px` into the first empty slot of `player.Inventory`, deleting collected ECS entities from the world.
     - Weapon durability depletion transitions cleanly to fists (`WeaponEquipped = false, WeaponType = "", WeaponDurability = 0`) when durability reaches `<= 0`.

2. **Adversarial Test Suite Implementation in `internal/game/destruction_adversarial_test.go`**:
   - Implemented 8 dedicated adversarial stress tests:
     - `TestAdversarial_WoodItemDropConservation_MassDestruction`: Mass destruction of 8x8 mixed destructible/indestructible arena. Confirms exact 1:1 drop conservation, center alignment, and walkability transitions.
     - `TestAdversarial_ZeroDropOnPartialDamage_NoPostDestroyDuplication`: Verifies zero drops on intermediate hits, single drop on death, and zero drops on post-destruction hits (anti-item-duplication).
     - `TestAdversarial_InventoryConsecutivePickups_SaturationAndRetention`: Tests sequential corridor pickup across 15 tiles; slots 0..8 saturate with wood, remaining 6 items stay unharmed in world.
     - `TestAdversarial_InventoryBatchClusterPickup_Saturation`: Tests instant cluster pickup of 15 drops at single point; 9 items collected, exactly 6 remain on ground.
     - `TestAdversarial_InventoryFragmentedPickup_PreservesExistingItems`: Validates sparse backpack slots (empty at 1,3,5,7) are filled with wood while pre-existing items in even slots remain untampered.
     - `TestAdversarial_WeaponBreakdownTransitions_ZeroDurabilityStress`: Tests weapon durability hitting 0 mid-chopping, clean transition to unarmed, 100 rapid unarmed swings without panic, and subsequent backup weapon equipping.
     - `TestAdversarial_Autotiling_EndcapTransitionsOnDestruction`: Tests real-time bitmask recalculation and endcap redrawing across horizontal lines, vertical lines, and complex cross-junctions (+).
     - `TestAdversarial_ShotgunConeCleave_MultiBarrierDestruction`: Tests multi-barrier cleave with shotgun blast, ammo consumption, durability degradation, and multiple drop creations.

3. **Empirical Execution Results**:
   - `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...` -> 100% tests passed across all packages with exit code 0.
   - `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game` -> Compiled binary successfully with 0 errors.

---

## 2. Logic Chain

1. **Wood Drop Conservation**:
   - Observed: When destroying an arena with $N$ destructibles, `DamageTile` triggers `destroyed == true` and returns `"wood"` exactly once per cell when durability reaches `<= 0`.
   - Invariant: Exactly $N$ wood items were instantiated in the ECS world. Coordinates matched `(tx*128+64, ty*128+64)` without floating-point drift. No drops were generated on partial damage (HP > 0) or post-destruction attacks.

2. **Inventory Saturation & Item Retention**:
   - Observed: `processItems()` searches for the first available `""` slot in `player.Inventory`. Once all 9 slots are populated, the loop terminates without modifying the remaining ground entities.
   - Invariant: Player inventory is capped strictly at 9 items. Surpassing capacity leaves excess items alive and untampered on the ground in ECS world space. Pre-existing inventory items are never overwritten.

3. **Weapon Breakdown Transitions**:
   - Observed: When `player.WeaponDurability--` reaches 0 upon striking a barrier, `WeaponEquipped` is set to `false`, `WeaponType` to `""`, and `WeaponDurability` to `0`.
   - Invariant: Subsequent attack attempts cleanly branch to unarmed shove (dealing 0 damage to barriers), preventing negative durability, integer underflows, or null pointer panics.

4. **Dynamic Autotiling Transitions**:
   - Observed: `GetCardinalBitmask4`, `GetWallBitmask`, and `GetFenceBitmask` dynamically inspect adjacent tile types in real-time.
   - Invariant: Breaching a barrier immediately triggers adjacent remaining wall/fence segments to recompute their bitmasks and render terminating endcaps (e.g. horizontal straight segment with mask 10 transitioning to endcaps 8 or 2; cross-junction 15 transitioning through 7, 5, 4, to 0).

---

## 3. Caveats

- **Perimeter Boundaries**: Perimeter boundary tiles (`x=0, y=0, x=Width-1, y=Height-1`) are intentionally hard-coded to be indestructible to preserve map topology and prevent entities from escaping the orthogonal grid.
- **Weapon Wear**: Hitting barriers consumes 1 point of weapon durability per swing that connects with at least one destructible barrier.

---

## 4. Conclusion

**Verdict: APPROVE**

The Environmental Destruction & Resource Drops system (Requirement R4) satisfies all functional and non-functional specifications. All 8 adversarial challenge tests in `internal/game/destruction_adversarial_test.go` pass with 100% reliability, confirming item drop conservation, inventory saturation handling, weapon breakdown lifecycle stability, and autotiling topological responsiveness.

---

## 5. Verification Method

1. **Execute All Adversarial Tests**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -run "TestAdversarial.*(Destruction|Wood|Inventory|Weapon|Autotiling|Shotgun)" ./internal/game
   ```

2. **Run Full Test Suite (Clean Count=1)**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...
   ```

3. **Build Binary**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
   ```
