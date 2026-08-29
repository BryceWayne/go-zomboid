# Handoff Report: Requirement R4 (Environmental Destruction) & Comprehensive Test Suite Verification

**Explorer 3**: Survey Findings & Implementation Proposal  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/explorer_survey_r4_1`  
**Date/Timestamp**: `2026-08-29T16:52:00Z`  
**Target Milestone**: 2D Engine Enhancement (R1 Autotiling, R2 Equip/Unequip, R3 Storage Chests, R4 Environmental Destruction)

---

## 1. Observation

### 1.1 Wooden Barriers & Obstacle Definitions in `internal/game/world/map.go`
- **Tile Types** (`map.go:11-34`):
  ```go
  const (
      TileGrass TileType = iota
      TileWall
      TileDirt
      TileWoodFloor
      TileTree
      TileAsphalt
      TileConcrete
      TileTileFloor
      TileFence
      TileDebris
      TileTent
      TileElevationBlock
      TileRamp
      TileStump
      TileMushroom
      TileSign
      TileBench
      TileChest
      TileSculpture
      TileBush
      TileFlower
      TileStone
  )
  ```
- **Solidity & Vision Occlusion** (`map.go:38-53`):
  - `IsSolid()` returns `true` for `TileWall`, `TileTree`, `TileFence`, `TileDebris`, `TileTent`, `TileElevationBlock`, `TileStump`, `TileSign`, `TileBench`, `TileChest`, `TileSculpture`, `TileStone`.
  - `BlocksVision()` returns `true` exclusively for `TileWall` (`return t == TileWall`). `TileFence`, `TileTree`, `TileBench`, and `TileStump` allow raycast vision.
- **Procedural Barrier Generation**:
  - `buildFencedYard(fx, fy, fw, fh int)` (`map.go:710-728`): Surrounds residential homes (House 1, House 2), warehouse yards, and the police motor pool with `TileFence` perimeter enclosures.
  - `buildResidentialHouse`, `buildGroceryStore`, `buildPharmacyClinic`, `buildPoliceStation`, `buildWarehouse` (`map.go:393-708`): Generate interior partition walls and exterior building walls with `TileWall`.
  - `placeEnvironmentalProps` (`map.go:740-890`): Generates `TileTree`, `TileStump`, `TileBench`, `TileSculpture`, `TileChest`, `TileDebris` across outdoor parks and plazas.
- **Current Barrier State**:
  - Map tiles are stored in a simple 1D slice: `m.Tiles = make([]TileType, width*height)` (`map.go:198`).
  - **No Durability or Health System**: There is currently no health tracking, durability counter, or destructibility classification for tiles in `map.go`. All solid obstacles are permanent and indestructible.

### 1.2 Combat, Attack Range, and Collision Mechanics in `internal/game/game.go`
- **Attack Inputs & Timing** (`game.go:520-538`):
  - Attacks trigger on `KeySpace`, `KeyX`, or `MouseButtonRight` with a 30-tick cooldown (`player.AttackCooldown = 30`).
  - Weapon types: `"axe"`, `"weapon"` (club/bat), `"shotgun"`, and Unarmed (fists).
- **Melee Attack Geometry** (`game.go:629-689`):
  - **Axe Attack** (`game.go:629-658`):
    ```go
    attackX := pos.X + player.FacingX*128.0
    attackY := pos.Y + player.FacingY*128.0
    // Cleave reach = 128.0px, radius = 128.0px
    ```
    Iterates over `s.zombieFilter.Query()`. If a zombie is within 128.0px of `attackX, attackY`, it is killed and `player.WeaponDurability--`.
  - **Standard Club/Weapon Attack** (`game.go:659-689`):
    ```go
    attackX := pos.X + player.FacingX*96.0
    attackY := pos.Y + player.FacingY*96.0
    // Reach = 96.0px, radius = 96.0px
    ```
    Kills zombies within 96.0px and deducts weapon durability.
  - **Unarmed Shove** (`game.go:690-707`):
    Applies stun and knockback velocity to zombies within 96.0px.
- **Missing Environmental Collision Check in Combat**:
  - In `processInputAndCombat` (`game.go:362-720`), attack routines **only query `s.zombieFilter`**.
  - **Zero interaction with the map**: Weapon swings completely pass through map obstacles without checking tile coordinates or reducing barrier HP.

### 1.3 ECS Components & Item Drops in `internal/ecs/components.go` and `internal/assets/assets.go`
- **ECS Player & Item Components** (`components.go:28-53`):
  - `Player` has `Inventory []string` (9 slots), `WeaponEquipped bool`, `WeaponType string`, `WeaponDurability int`.
  - `Item` has `Type string` (spawned with `Position{X, Y}`).
- **Item Collection Routine** (`game.go:314-360`):
  - `processItems()` queries all `ecs.Item` entities within 64.0px of player position `(pX, pY)`.
  - Inserts `item.Type` into the first empty slot in `player.Inventory` and removes the item entity from `s.world`.
- **Item Asset Definitions** (`assets.go:45-54`):
  - Current item textures: `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `AntidoteImage`, `FoodImage`, `WaterImage`, and `WoodImage`.
  - In `DrawSystem.Draw` (`game.go:1140-1178`), ground items are rendered centered at `(iPos.X, iPos.Y)`. A `"wood"` item drop can seamlessly utilize `assets.WoodImage` (or a dedicated wood resource sprite).

### 1.4 Existing Test Suite Inventory
The codebase contains 22 existing test suites covering:
1. `internal/game/orthogonal_engine_test.go`: Coordinate transforms (`WorldToIso`, `ScreenToWorld`, `WorldToScreen`), Cartesian camera lerping, tile adjacency, depth sorting.
2. `internal/game/combat_test.go`: 15 comprehensive unit tests for weapon equipping, axe cleave, shotgun spread cones, noise pulses, and weapon breakdown.
3. `internal/game/armor_test.go`: Armor equipping, deflection rolls, durability degradation.
4. `internal/game/dm_test.go`: Dynamic wave spawning, threat scaling, loot tables, day/night lighting.
5. `internal/game/world/map_test.go`: Map dimensions, procedural buildings, player spawn safety, collision detection (`IsColliding`), FOV occlusion (`BlocksVision`).
6. `internal/game/challenger_tile_render_test.go`: Rendering all tile types through `DrawSystem`.
7. `internal/assets/m1_stress_verification_test.go`: Image asset loading and audio initialization.

**Build & Test Environment Finding**:
Running `go test ./...` natively requires C headers. The execution command is:
`C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test ./...`
which passes 100% of existing tests.

---

## 2. Logic Chain

### 2.1 Environmental Destruction (Requirement R4) Mechanics
1. **Barrier Identification**:
   - `TileFence` is the primary wooden barrier surrounding residential yards, warehouse lots, and courtyards.
   - `TileWall` (wood walls / interior partitions) and wooden props (`TileTree`, `TileStump`, `TileBench`) also represent destructible wooden objects.
   - Perimeter boundary walls (`x=0, x=Width-1, y=0, y=Height-1`) must be flagged as indestructible boundary anchors to prevent entities from walking outside world limits.
2. **Durability & Health System**:
   - Introduce tile durability tracking in `world.Map` via a `TileDurability map[Point]int` (or `[]int` health grid).
   - Define baseline durability values:
     - `TileFence`: Max HP = 2 (Axe destroys in 1 swing with 2 damage, Club takes 2 swings with 1 damage).
     - `TileWall` (non-perimeter): Max HP = 3 (Axe takes 2 swings, Club takes 3 swings).
     - `TileTree`: Max HP = 3 (Axe takes 2 swings).
     - `TileStump` / `TileBench`: Max HP = 2.
     - Unarmed shove: 0 damage (cannot chop wooden barriers).
3. **Attack Sweep & Collision Detection**:
   - When the player attacks with a melee weapon or axe:
     - Attack center is calculated at `attackX := pos.X + player.FacingX * reach`, `attackY := pos.Y + player.FacingY * reach`.
     - Scan all tile grid coordinates $(tx, ty)$ intersecting the attack bounding box / radius:
       $$|tx \cdot \text{TileSize} + 64 - \text{attackX}| \le \text{radius} + 64 \quad \land \quad |ty \cdot \text{TileSize} + 64 - \text{attackY}| \le \text{radius} + 64$$
     - If `m.IsDestructible(tx, ty)` is true:
       - Apply damage `dmg` (Axe: 2, Weapon/Club: 1, Shotgun: 2).
       - Play hit sound (`assets.PlaySound(assets.HitSound)`).
       - Deduct weapon durability (`player.WeaponDurability--`).
4. **Destruction & Map State Transition**:
   - When durability reaches 0:
     - Replace tile with walkable ground: `m.SetTile(tx, ty, TileGrass)` (or `TileWoodFloor` if indoors).
     - Because `TileGrass` has `IsSolid() == false` and `BlocksVision() == false`, the tile's physical solidity and vision blocking are immediately cleared!
     - `m.IsColliding(...)` will now return `false` through the newly opened pathway.
     - Recalculate FOV raycasts so sight lines immediately penetrate the cleared gap.
5. **Resource Item Drops & Collection**:
   - At the destroyed tile's center `(float64(tx)*TileSize + 64.0, float64(ty)*TileSize + 64.0)`, spawn an `ecs.Item{Type: "wood"}` entity in the ECS world.
   - When the player steps within 64px, `processItems()` picks up the `"wood"` item into their inventory.
   - `DrawSystem` renders the dropped wood item on the ground and displays `"wood"` in the inventory HUD.

---

## 3. Caveats

1. **Perimeter Boundary Walls**:
   Perimeter walls (`x == 0 || x == Width-1 || y == 0 || y == Height-1`) MUST be exempt from destruction (`IsDestructible` returns false). Otherwise, destroying perimeter walls would allow the player and zombies to walk out of bounds and trigger camera/render out-of-range exceptions.
2. **Indoors vs. Outdoors Ground Replacement**:
   When destroying an interior wall in a residential house or store, the replacement floor tile should ideally match the room's floor (`TileWoodFloor`, `TileConcrete`, or `TileTileFloor`). Defaulting to `TileWoodFloor` indoors and `TileGrass` outdoors ensures seamless visual aesthetics.
3. **Weapon Durability Consumption**:
   Chopping barriers should deduct weapon durability per hit (just like hitting zombies). When weapon durability reaches 0, the weapon breaks and unequips to fists, preventing infinite chopping with a broken tool.
4. **Unarmed Attacks**:
   Unarmed fists/shoves must NOT damage wooden barriers. Fists should play the shove sound and leave barrier durability untouched.
5. **Autotiling Integration (R1 Interaction)**:
   When a barrier (fence or wall) is destroyed, neighboring autotiling bitmasks should update dynamically so adjacent walls/fences redraw with clean terminating end-caps rather than broken connections.

---

## 4. Conclusion & Technical Implementation Proposal

### 4.1 Proposed Changes to `internal/game/world/map.go`
Add durability tracking and destruction helper methods to `world.Map`:

```go
// Map additions:
type Map struct {
    Width, Height int
    Tiles         []TileType
    Visible       []bool
    Explored      []bool
    TileDurability map[Point]int // Tracks current damage to destructible tiles
    
    // Spawns and buildings metadata...
    PlayerSpawn   FloatPoint
    Buildings     []Building
    LootSpawns    []LootSpawn
    ZombieSpawns  []FloatPoint
}

// IsDestructible returns true if the tile at (x, y) can be damaged and destroyed by weapons.
func (m *Map) IsDestructible(x, y int) bool {
    if x <= 0 || x >= m.Width-1 || y <= 0 || y >= m.Height-1 {
        return false // Perimeter boundary walls are indestructible
    }
    t := m.GetTile(x, y)
    switch t {
    case TileFence, TileTree, TileStump, TileBench:
        return true
    case TileWall:
        return true // Interior / destructible building walls
    default:
        return false
    }
}

// GetTileMaxDurability returns the maximum durability hits for a destructible tile.
func (m *Map) GetTileMaxDurability(t TileType) int {
    switch t {
    case TileFence:
        return 2
    case TileTree:
        return 3
    case TileStump, TileBench:
        return 2
    case TileWall:
        return 3
    default:
        return 0
    }
}

// GetTileDurability returns remaining health of tile at (x, y).
func (m *Map) GetTileDurability(x, y int) int {
    if m.TileDurability == nil {
        m.TileDurability = make(map[Point]int)
    }
    p := Point{X: x, Y: y}
    if cur, exists := m.TileDurability[p]; exists {
        return cur
    }
    return m.GetTileMaxDurability(m.GetTile(x, y))
}

// DamageTile reduces tile durability by amount. Returns true if destroyed along with dropped resource type.
func (m *Map) DamageTile(x, y int, amount int) (destroyed bool, dropType string) {
    if !m.IsDestructible(x, y) || amount <= 0 {
        return false, ""
    }
    if m.TileDurability == nil {
        m.TileDurability = make(map[Point]int)
    }
    p := Point{X: x, Y: y}
    cur := m.GetTileDurability(x, y)
    cur -= amount
    if cur <= 0 {
        delete(m.TileDurability, p)
        t := m.GetTile(x, y)
        // Replace with walkable ground
        if t == TileWall {
            m.SetTile(x, y, TileWoodFloor)
        } else {
            m.SetTile(x, y, TileGrass)
        }
        return true, "wood"
    }
    m.TileDurability[p] = cur
    return false, ""
}
```

### 4.2 Proposed Changes to `internal/game/game.go` (`processInputAndCombat`)
Integrate barrier detection and chopping into melee attack routines:

```go
// Inside processInputAndCombat:
// Helper to chop barriers in attack radius:
chopBarriers := func(attackCenterX, attackCenterY, attackRadius float64, damage int) bool {
    if damage <= 0 {
        return false
    }
    hitBarrier := false
    minTx := int(attackCenterX-attackRadius) / world.TileSize
    maxTx := int(attackCenterX+attackRadius) / world.TileSize
    minTy := int(attackCenterY-attackRadius) / world.TileSize
    maxTy := int(attackCenterY+attackRadius) / world.TileSize

    for ty := minTy; ty <= maxTy; ty++ {
        for tx := minTx; tx <= maxTx; tx++ {
            if tx < 0 || tx >= s.gameMap.Width || ty < 0 || ty >= s.gameMap.Height {
                continue
            }
            tileCenterX := float64(tx)*float64(world.TileSize) + float64(world.TileSize)/2.0
            tileCenterY := float64(ty)*float64(world.TileSize) + float64(world.TileSize)/2.0
            dist := math.Hypot(attackCenterX-tileCenterX, attackCenterY-tileCenterY)
            if dist <= attackRadius+float64(world.TileSize)/2.0 {
                if s.gameMap.IsDestructible(tx, ty) {
                    destroyed, dropType := s.gameMap.DamageTile(tx, ty, damage)
                    hitBarrier = true
                    if destroyed && dropType != "" {
                        itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](s.world)
                        itemMap.NewEntity(
                            &ecs.Item{Type: dropType},
                            &ecs.Position{X: tileCenterX, Y: tileCenterY},
                        )
                    }
                }
            }
        }
    }
    return hitBarrier
}
```

- When player attacks with **Axe** (`damage = 2`):
  Calls `chopBarriers(attackX, attackY, 128.0, 2)`. If `hitZombies || hitBarrier`, plays hit sound and decrements `player.WeaponDurability`.
- When player attacks with **Club/Weapon** (`damage = 1`):
  Calls `chopBarriers(attackX, attackY, 96.0, 1)`. If `hitZombies || hitBarrier`, plays hit sound and decrements `player.WeaponDurability`.
- When player attacks with **Shotgun** (`damage = 2`):
  Can damage barriers within point-blank / spread cone.
- When player attacks **Unarmed** (`damage = 0`):
  Does not damage barriers.

### 4.3 Proposed Changes to `internal/game/game.go` (`DrawSystem`)
- Support `"wood"` item rendering in `DrawSystem.Draw`:
  ```go
  case "wood":
      img = assets.WoodImage
  ```
- Support `"wood"` text display in inventory HUD slots.

---

## 5. Comprehensive Test Suite Verification Plan (R1, R2, R3, R4)

### 5.1 Verification Commands
- **Run All Tests**:
  ```bash
  C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v ./...
  ```
- **Build Executable Binary**:
  ```bash
  C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
  ```

### 5.2 Required Test Cases per Requirement

| Requirement | Test File | Test Name | Objective & Assertions |
|-------------|-----------|-----------|------------------------|
| **R1: Tile Autotiling** | `internal/game/world/map_test.go` | `TestAutotiling_BitmaskComputation` | Verifies 4-bit / 8-bit neighbor bitmasks (N, S, E, W, diagonals) for grass/dirt, grass/wall, and floor boundaries. |
| **R1: Tile Autotiling** | `internal/game/challenger_tile_render_test.go` | `TestAutotiling_SubImageUVMapping` | Verifies correct UV sub-rectangle coordinates mapped from autotile tilesets without out-of-bounds panics. |
| **R1: Tile Autotiling** | `internal/game/orthogonal_engine_test.go` | `TestAutotiling_SeamlessTerrainBlending` | Verifies adjacent disparate tiles share aligned border transitions without black gaps or square artifact borders. |
| **R2: Equip/Unequip** | `internal/game/combat_test.go` | `TestEquip_DedicatedSlotTransfer` | Equipping weapon from inventory slot `i` moves item to `Equipped` slot and clears `inventory[i]`. |
| **R2: Equip/Unequip** | `internal/game/combat_test.go` | `TestUnequip_ReturnToFirstEmptySlot` | Unequipping returns item to first available empty slot in main inventory (1-9) and resets equipped state. |
| **R2: Equip/Unequip** | `internal/game/combat_test.go` | `TestUnequip_FullInventoryBlocked` | Unequipping when all 9 inventory slots are occupied cleanly rejects or retains item without deletion. |
| **R2: Equip/Unequip** | `internal/game/combat_test.go` | `TestEquip_DurabilityDepletionClearsSlot` | Weapon breaking at 0 durability clears dedicated equipped slot back to fists. |
| **R3: Storage Chest** | `internal/game/game_test.go` | `TestStorageChest_ProximityDetection` | Pressing 'E' within 192px of `TileChest` triggers interaction; pressing 'E' > 250px away does nothing. |
| **R3: Storage Chest** | `internal/game/game_test.go` | `TestStorageChest_FullInventorySwap` | Entire 9-slot player inventory and 9-slot chest inventory swap element-for-element without loss or corruption. |
| **R3: Storage Chest** | `internal/game/game_test.go` | `TestStorageChest_MultipleChestsIndependent` | Interacting with Chest A vs Chest B maintains distinct isolated item storage for each coordinate. |
| **R3: Storage Chest** | `internal/game/game_test.go` | `TestStorageChest_KeyDebounce` | Holding 'E' continuously across frames does not oscillate inventory back and forth every tick. |
| **R4: Environmental Destruction** | `internal/game/world/map_test.go` | `TestDestruction_TileDurabilityDegradation` | `DamageTile` reduces fence HP from 2 to 1 on hit 1, and destroys on hit 2, returning `destroyed=true, dropType="wood"`. |
| **R4: Environmental Destruction** | `internal/game/world/map_test.go` | `TestDestruction_PerimeterIndestructible` | Perimeter border walls (`x=0, y=0, x=W-1, y=H-1`) return `IsDestructible=false` and cannot be damaged. |
| **R4: Environmental Destruction** | `internal/game/world/map_test.go` | `TestDestruction_SolidityAndVisionCleared` | Upon barrier destruction, `GetTile` is `TileGrass`, `IsSolid()` is `false`, `IsColliding` returns `false`, and FOV penetrates. |
| **R4: Environmental Destruction** | `internal/game/combat_test.go` | `TestCombat_AxeChopBarrierSpawnsWoodDrop` | Axe swing facing `TileFence` damages fence, destroys it, decrements axe durability, and spawns `ecs.Item{Type: "wood"}`. |
| **R4: Environmental Destruction** | `internal/game/combat_test.go` | `TestCombat_UnarmedCannotChopBarriers` | Unarmed shove against `TileFence` leaves barrier durability unchanged and spawns no items. |
| **R4: Environmental Destruction** | `internal/game/combat_test.go` | `TestCombat_WoodPickupIntoInventory` | Player moving within 64px of dropped wood item collects `"wood"` into first empty inventory slot. |
| **R4: Environmental Destruction** | `internal/game/game_stress_test.go` | `TestStress_MultiBarrierBreachAndTraversal` | Headless simulation where player chops down 5 consecutive fence segments, picks up all 5 wood drops, and walks through the gap. |

---

## 6. Verification Method

To verify this survey and all subsequent code changes:
1. Verify compiler passes without error:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test ./internal/game/world ./internal/ecs ./internal/game ./internal/assets
   ```
2. Verify binary build succeeds:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
   ```
3. Run the complete test suite:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v ./...
   ```
