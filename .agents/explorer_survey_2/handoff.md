# Handoff Report: Gameplay Systems & Dungeon Master Architecture Survey

**Investigator**: Explorer Survey 2  
**Date**: 2026-08-29  
**Target Repository**: `go-zomboid` (`/home/bryce/code/go-zomboid`)  
**Scope**: World Generation & Map Structures, Ark ECS Entity Management, Player & Zombie AI, Combat & Inventory, Day/Night Cycle & Lighting Pipeline, Dungeon Master Simulation Architecture.

---

## 1. Observation

Direct code inspection of the `go-zomboid` codebase revealed the following architecture and implementation facts:

### 1.1 World Generation & Map Data Structures (`internal/game/world/map.go`)
- **Map Model (`map.go:180-191`)**:
  - Grid dimensions: `Width int`, `Height int` (default `100 x 100` cells).
  - Cell size: `world.TileSize = 128` world coordinate units (`map.go:36`). The default map spans `12,800 x 12,800` Cartesian world units.
  - Flattened 1D slices: `Tiles []TileType` (indexed by `y * Width + x`), `Visible []bool` (real-time FOV visibility), `Explored []bool` (fog of war persistence).
- **Tile Types & Solid Properties (`map.go:8-63`)**:
  - 22 `TileType` definitions: `TileGrass` (0), `TileWall` (1), `TileDirt` (2), `TileWoodFloor` (3), `TileTree` (4), `TileAsphalt` (5), `TileConcrete` (6), `TileTileFloor` (7), `TileFence` (8), `TileDebris` (9), `TileTent` (10), `TileElevationBlock` (11), `TileRamp` (12), `TileStump` (13), `TileMushroom` (14), `TileSign` (15), `TileBench` (16), `TileChest` (17), `TileSculpture` (18), `TileBush` (19), `TileFlower` (20), `TileStone` (21).
  - `IsSolid()` (`map.go:39-47`): Solid obstacles that block entity movement: `TileWall`, `TileTree`, `TileFence`, `TileDebris`, `TileTent`, `TileElevationBlock`, `TileStump`, `TileSign`, `TileBench`, `TileChest`, `TileSculpture`, `TileStone`.
  - `BlocksVision()` (`map.go:49-52`): Line-of-sight occlusion: only `TileWall` blocks vision.
  - `IsFloor()` (`map.go:54-62`): Flat floor surfaces: `TileGrass`, `TileDirt`, `TileWoodFloor`, `TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileRamp`.
- **Procedural Town Generation (`NewMap`, `map.go:194-371`)**:
  - Outer perimeter boundary walls enclosing `[0..Width-1, 0..Height-1]`.
  - Road Network (`map.go:222-271`): East-West Main Avenue (`midY`), North-South Main Boulevard (`midX`), secondary residential access road (North ~25% height), secondary industrial access road (South ~75% height), dirt trails.
  - Zoned District Buildings (`map.go:273-343`):
    - *Residential (NW)*: House 1 (Player safe start + fenced yard), House 2, House 3, House 4 (South).
    - *Commercial (NE)*: Grocery Store (`18x12`, sales floor + storage), Pharmacy/Clinic (`14x10`, consultation + medical storage).
    - *Municipal/Defense (SW)*: Police Station (`16x13`, lobby, office, armory, holding cells, motor pool courtyard).
    - *Industrial (SE)*: Warehouse Depot (`20x14`, loading bay, foreman office, crate clusters).
  - Environmental Props & Foliage (`map.go:740-917`): 180 procedural foliage placement rolls (trees, mushrooms, stumps, bushes, flowers, stones), campsite with elevation ramps, sculptures, benches, chests.
  - Spawns & Placements (`map.go:347-370`):
    - `PlayerSpawn`: Placed inside House 1 living room (`x=13, y=14`) at `(float64(tx)*128 + 64.0, float64(ty)*128 + 64.0)`.
    - `generateThematicLoot` (`map.go:919-976`): Guaranteed starter loot in starting house, plus contextual loot in rooms (food/water in kitchens/stores, weapons/ammo/armor in armory, antidotes in clinic, axes/armor in warehouse).
    - `generateZombieSpawns(140)` (`map.go:992-1019`): 140 non-colliding positions with distance $\ge 1400.0$ pixels from player spawn.
- **FOV Raycasting & Collision (`map.go:1021-1104`)**:
  - `CalculateFOV(playerX, playerY, radiusTiles)` casts `radiusTiles * 8` angular rays (called with `radiusTiles = 22` in `UpdateSystem`).
  - `IsColliding(rectX, rectY, rectW, rectH)` performs AABB bounding box collision against solid tiles.

---

### 1.2 Entity Management & Ark ECS (`internal/ecs/components.go` & `internal/game/game.go`)
- **Ark ECS Framework**: `github.com/mlange-42/ark/ecs`.
- **Component Definitions (`components.go:8-65`)**:
  1. `ecs.Position`: `{ X, Y float64 }` (continuous world pixel coordinates).
  2. `ecs.Velocity`: `{ X, Y float64 }` (per-frame movement vector).
  3. `ecs.Sprite`: `{ Color color.RGBA, W, H float64 }`.
  4. `ecs.Collider`: `{ Width, Height float64 }` (AABB dimensions, default 64x64).
  5. `ecs.Player`:
     - `Health float64` (100.0 max)
     - `Hunger float64` (100.0 full, drains `0.003`/frame; damage at 0)
     - `Thirst float64` (100.0 hydrated, drains `0.005`/frame; damage at 0)
     - `Inventory []string` (capacity 9 slots)
     - `WeaponEquipped bool`, `WeaponType string` ("weapon", "axe", "shotgun"), `WeaponDurability int`
     - `ArmorEquipped bool`, `ArmorType string` ("vest"), `ArmorDefense float64` (0.50), `ArmorDurability int` (10), `ArmorMaxDurability int` (10), `InfectionResist float64` (0.70)
     - `AttackCooldown int` (30 frames max)
     - `Dead bool`, `Infected bool`
     - `FacingX, FacingY float64`
  6. `ecs.Item`: `{ Type string }` ("food", "water", "weapon", "axe", "shotgun", "ammo", "armor", "antidote").
  7. `ecs.Zombie`:
     - `Speed float64` (Regular: 4.0–6.0 px/frame, Runner: 8.8–10.4 px/frame)
     - `Chasing bool`
     - `IsRunner bool`
     - `WanderTimer int`
     - `WanderDirX, WanderDirY float64`
     - `StunTimer int`
- **Entity Creation & Mappers (`game.go:42-112`)**:
  - `playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)`
  - `zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)`
  - `itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)`

---

### 1.3 Player Mechanics, Combat, and Zombie AI (`internal/game/game.go`)
- **Player Movement & Inventory Usage (`game.go:294-460`)**:
  - Player speed: `12.0` px/frame (WASD / arrow keys or Left Mouse Button towards cursor).
  - Item consumption via Number Keys `1`–`9`:
    - "food" (+50 Hunger), "water" (+50 Thirst), "antidote" (cures infection).
    - "weapon" (equips Spiked Club/Bat, durability 5).
    - "axe" (equips Fire Axe, durability 12).
    - "shotgun" (equips Shotgun, durability 15).
    - "armor" / "vest" (equips Tactical Vest, durability 10, defense 0.50, resist 0.70).
  - Infection damage: `0.05` HP/frame (mitigated by `ArmorDefense` if equipped).
- **Combat Execution (`game.go:462-645`)**:
  - Triggered by Space, Key 'X', or Right Mouse Button. Sets `AttackCooldown = 30`.
  - **Shotgun**: Consumes 1 "ammo" item from inventory + 1 weapon durability. Point-blank kill (`< 96px`) or spread cone (`Range: 640px`, angle $\pm 22.5^\circ$). Emits **acoustic noise pulse** alerting all wandering zombies within `1600.0px` (`z.Chasing = true`). If out of ammo, performs dry-fire butt shove.
  - **Fire Axe**: Cleave reach `128.0px`, radius `128.0px`. Multi-target cleave kill. Deducts 1 durability on hit.
  - **Weapon (Bat/Club)**: Reach `96.0px`, radius `96.0px`. Multi-target kill. Deducts 1 durability on hit.
  - **Unarmed Shove**: Reach `96.0px`, radius `96.0px`. Applies `StunTimer = 45` frames and pushback impulse `Facing * 20.0`.
- **Zombie AI System (`game.go:647-783`)**:
  - Detection radii: Base noise radius = `200.0px` (or `800.0px` if player is moving); Vision radius = `600.0px`.
  - If `dist < noiseRadius || dist < visionRadius` $\rightarrow$ `zombie.Chasing = true`.
  - If `dist > 1600.0 || playerDead` $\rightarrow$ `zombie.Chasing = false`.
  - Contact / Bite (`dist < 56.0px`): Deflection roll against `playerComp.InfectionResist` (70%). Unarmored player takes immediate infection. Armor durability decays by 1 on contact hit.
  - Flocking & Boid Separation: `separationRadius = 80.0px`, `separationForce = 8.0px`.
- **Item Pickup (`game.go:251-292`)**:
  - Distance check: `dist < 64.0px`. If `len(player.Inventory) < 9`, adds item to inventory and calls `s.world.RemoveEntity(ent)`.

---

### 1.4 Time, Day/Night Cycle, and Lighting Pipeline (`internal/game/game.go`)
- **Game Clock (`game.go:37, 123-128`)**:
  - `g.timeOfDay float64` initialized to `8.0` (8:00 AM).
  - Advanced in `Game.Update()`: `g.timeOfDay += 24.0 / (60.0 * 5.0 * 60.0)` (5 real minutes = 24 in-game hours; 0.001333 hours/frame).
- **Ambient Lighting in `DrawSystem.Draw` (`game.go:1193-1200`)**:
  - Current formula: `alpha := 0.45 + 0.45 * math.Cos((timeOfDay / 24.0) * math.Pi * 2)`.
  - At Noon (`12.0`), `alpha = 0.0` (bright clear sunlight).
  - At Midnight (`0.0` / `24.0`), `alpha = 0.90` (darkness).
  - Drawn via `vector.DrawFilledRect(screen, 0, 0, 1280, 720, color.RGBA{0, 0, 15, uint8(alpha * 255)}, false)`.
- **Critical Gap**:
  - **Enemy AI currently has zero awareness of `timeOfDay`!** Zombie speed, vision range, hearing range, and aggression are static constants day and night (`game.go:667-675`).
  - **No wave spawning exists**: Once initial 140 zombies are killed, the map permanently stays empty.
  - **No loot restocking exists**: Once initial contextual loot items are consumed, no new items ever spawn, and zombies drop no loot upon death.

---

## 2. Logic Chain

1. **Static Population Limitation**:
   - *Observation*: `Game.Reset()` spawns exactly 140 zombies once (`game.go:89-112`). When killed, `w.RemoveEntity(ent)` destroys them (`game.go:643`).
   - *Logic*: As players clear buildings, zombie population monotonically drops to zero. To satisfy requirement §R2 ("dynamic zombie wave spawning scaling difficulty over time"), a simulation controller must periodically evaluate the living zombie count, calculate elapsed game difficulty, and spawn new zombie waves outside the player's active camera/FOV on valid walkable tiles.

2. **Resource Starvation Limitation**:
   - *Observation*: `gameMap.LootSpawns` places initial items in buildings (`game.go:74-80`). Pickup destroys the entity (`game.go:290`). Combat requires non-renewable consumables (ammo for shotgun, food/water for hunger/thirst, weapons with finite durability).
   - *Logic*: Without loot replenishment, player survival is strictly capped by initial map loot. To satisfy requirement §R2 ("randomized loot drops across the map"), the Dungeon Master must introduce two loot injection mechanisms: (a) zombie death drops (weighted drop table on kill), and (b) periodic ambient supply drops in unvisited rooms/plazas.

3. **Decoupled Lighting & AI Limitation**:
   - *Observation*: `g.timeOfDay` updates in `Game.Update()` and feeds into `DrawSystem.Draw` to modulate an overlay alpha (`game.go:1196`), but `UpdateSystem.processZombies()` does not receive or use `timeOfDay`.
   - *Logic*: To satisfy requirement §R2 ("day/night cycle that darkens ambient lighting and increases enemy aggression at night"), `timeOfDay` must be passed into AI update logic to dynamically scale zombie speed, acoustic hearing range, vision range, runner spawn ratio, and chase tenacity during night hours (20:00 to 05:00).

4. **Orthogonal Grid Alignment (R1 ↔ R2)**:
   - *Observation*: Under R1, coordinate transformations `WorldToIso` and `IsoToWorld` become 1:1 orthogonal mappings (`screenX = (wx - camX) * scale + center`).
   - *Logic*: All entity physics, distances (`math.Hypot(dx, dy)`), collision boxes, and AI ranges operate natively in Cartesian world coordinates `(wx, wy)`. The Dungeon Master system operates strictly in Cartesian world space `(wx, wy)` and tile space `(tx, ty)`, making it 100% decoupled from projection math and fully compatible with the 2D orthogonal overhaul.

---

## 3. Caveats

1. **ECS Entity Density & Frame Rate Budget**:
   - Ark ECS is high-performance, but spawning thousands of active entities would stress distance queries. The Dungeon Master must enforce a maximum active zombie cap (`MaxLivingZombies = 120-150`) and maximum active items cap (`MaxMapItems = 60-80`).
2. **Spawn Proximity & Visual Popping**:
   - Spawning zombies too close to the player causes visible "popping". The Dungeon Master must enforce a minimum spawn exclusion radius ($R_{min} \ge 700\text{px}$) and verify target tiles are outside the player's active FOV or behind walls.
3. **Headless Test Compatibility**:
   - All Dungeon Master logic (wave calculations, drop rolls, time scaling, difficulty curves) must execute deterministically in headless unit tests without requiring active Ebitengine window contexts.

---

## 4. Conclusion: Dungeon Master Architecture Specification

To satisfy all requirements of **R2 (Dungeon Master Simulation)**, the system architecture is structured as follows:

```
                  +----------------------------------------------+
                  |                 Game Loop                    |
                  |                (cmd/game)                    |
                  +----------------------+-----------------------+
                                         |
                       +-----------------+-----------------+
                       |                                   |
                       v                                   v
             +--------------------+               +------------------+
             |   Dungeon Master   |               |   UpdateSystem   |
             |   (Simulation DM)  |               | (Combat/Physics) |
             +---------+----------+               +--------+---------+
                       |                                   |
         +-------------+-------------+                     |
         |             |             |                     |
         v             v             v                     v
   +-----------+ +-----------+ +-----------+     +-------------------+
   |Dynamic    | |Dynamic    | |Day/Night  |     | Ark ECS World     |
   |Zombie Wave| |Loot Drops | |Aggression |<--->| - Player (1)      |
   |Spawning   | |& Restock  | |Scaling    |     | - Zombies (N)     |
   +-----------+ +-----------+ +-----------+     | - Items (M)       |
                                                 +-------------------+
```

### 4.1 System Data Structures (`internal/game/dm.go`)

```go
package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

type DungeonMasterConfig struct {
	// Wave Spawning Config
	WaveIntervalFrames     int     // 1800 frames = 30.0s between wave evaluations
	BaseZombiesPerWave     int     // 3 base zombies
	MaxLivingZombies       int     // 140 max concurrent zombies
	MinSpawnDistance       float64 // 700.0px (outside immediate viewport)
	MaxSpawnDistance       float64 // 1600.0px (within active neighborhood)
	DayRunnerProbability   float64 // 0.15 (15% runners by day)
	NightRunnerProbability float64 // 0.45 (45% runners at night)

	// Loot Drop Config
	ZombieDropChance       float64 // 0.25 (25% chance of item drop on kill)
	MaxMapItems            int     // 60 max concurrent ground items
	SupplyDropInterval     int     // 3600 frames = 60.0s between ambient supply rolls

	// Day/Night & Aggression Config
	DayCycleMinutes        float64 // 5.0 real minutes per 24 in-game hours
	NightSpeedMultiplier   float64 // 1.25 (+25% zombie speed at night)
	NightNoiseMultiplier   float64 // 1.50 (+50% sound detection radius at night)
	NightVisionRadius      float64 // 750.0px (increased zombie night vision)
}

type DungeonMaster struct {
	world       *arkecs.World
	gameMap     *world.Map
	config      DungeonMasterConfig

	// State trackers
	TotalTicks         int64
	DayCount           int
	WaveCount          int
	NextWaveTick       int64
	NextSupplyDropTick int64
	DifficultyRating   float64 // Dynamic scale factor: 1.0 -> 3.5+

	// ECS Entity Mappers
	zombieMap arkecs.Map5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider]
	itemMap   arkecs.Map2[ecs.Item, ecs.Position]
	playerMap arkecs.Map2[ecs.Player, ecs.Position]
}
```

---

### 4.2 Dynamic Zombie Wave Spawning Algorithm

1. **Difficulty Scaling Formula**:
   $$\text{Threat}(t) = 1.0 + \frac{\text{TotalTicks}}{60 \times 180} + 0.25 \times \text{DayCount} + (0.50 \text{ if Night else } 0.0)$$
   $$\text{WaveSize} = \text{clamp}\left(\lfloor\text{BaseZombies} \times \text{Threat}(t)\rfloor, 3, 16\right)$$

2. **Off-Screen Walkable Coordinate Search**:
   - Generate candidate tile $(tx, ty)$ at angle $\theta \in [0, 2\pi)$ and radial distance $r \in [R_{min}, R_{max}]$ from player.
   - Assert $2 \le tx < \text{Width}-2$ and $2 \le ty < \text{Height}-2$.
   - Assert $!\text{gameMap.GetTile}(tx, ty).\text{IsSolid}()$.
   - If player FOV is available, prioritize tiles where $\text{Visible}[ty \times W + tx] == \text{false}$.

3. **Archetype Assignment**:
   - Roll runner probability: $P(\text{Runner}) = \text{NightRunnerProb}$ if night else $\text{DayRunnerProb}$.
   - Compute zombie speed: $\text{BaseSpeed} \times \text{AggressionSpeedMultiplier}$.
   - If spawned during high-threat night waves, set $\text{Chasing} = \text{true}$ to direct horde toward player's general district.

---

### 4.3 Dynamic & Randomized Loot Drop Tables

1. **On-Kill Death Drops (`HandleZombieDeath(wx, wy)`)**:
   - Roll drop chance: $r < \text{ZombieDropChance}$ (25%).
   - Weighted Rarity Table:
     | Item | Weight | Category | Purpose |
     |---|---|---|---|
     | `ammo` | 30% | Consumable | Refills shotgun ammunition |
     | `food` | 25% | Survival | Restores +50 hunger |
     | `water` | 20% | Survival | Restores +50 thirst |
     | `weapon` | 10% | Melee | Basic club (5 durability) |
     | `antidote` | 8% | Medical | Cures lethal zombie infection |
     | `axe` | 4% | Heavy Melee | Fire axe (12 durability, cleave) |
     | `armor` | 2% | Defense | Tactical vest (10 durability, 70% resist) |
     | `shotgun` | 1% | Ranged | Shotgun (15 durability, spread cone) |
   - Instantiate entity via `itemMap.NewEntity(&ecs.Item{Type: itemType}, &ecs.Position{X: wx, Y: wy})`.

2. **Periodic Ambient Restock / Supply Drops**:
   - Every 60 seconds (`NextSupplyDropTick`), if total items $< \text{MaxMapItems}$:
   - Locate valid indoor rooms (Store, Armory, Clinic, Kitchen) or outdoor plazas.
   - Spawn 2–4 themed supplies to maintain exploratory incentives.

---

### 4.4 Day/Night Lighting & Enemy Aggression Scaling

1. **Day/Night Phases & Ambient Lighting**:
   - **Dawn (05:00 – 08:00)**: Warm golden-rose ambient tint (`color.RGBA{180, 140, 80, uint8(alpha*255)}`), alpha transitions $0.60 \rightarrow 0.05$.
   - **Day (08:00 – 17:00)**: Clear sunlight, alpha $= 0.0$.
   - **Dusk (17:00 – 20:00)**: Amber-crimson twilight (`color.RGBA{120, 40, 60, uint8(alpha*255)}`), alpha transitions $0.05 \rightarrow 0.60$.
   - **Night (20:00 – 05:00)**: Deep midnight navy overlay (`color.RGBA{5, 10, 35, uint8(alpha*255)}`), alpha peaking at $0.88$ at 00:00 (Midnight).

2. **Nighttime Aggression Multipliers**:
   ```go
   func (dm *DungeonMaster) GetAggressionModifiers(timeOfDay float64) (speedMult, noiseMult, visionMult float64) {
       if timeOfDay < 5.0 || timeOfDay > 20.0 { // Night
           nightFactor := 1.0
           if timeOfDay >= 22.0 || timeOfDay <= 3.0 { // Deep Midnight Peak
               nightFactor = 1.30
           } else {
               nightFactor = 1.15
           }
           return 1.25 * nightFactor, 1.50 * nightFactor, 1.25 * nightFactor
       }
       return 1.0, 1.0, 1.0 // Daytime base
   }
   ```
   - At Night:
     - Zombie speed increased by $+25\%$ to $+35\%$.
     - Player movement noise detection radius expands from $800\text{px} \rightarrow 1200\text{px}-1400\text{px}$.
     - Gunshot acoustic pulse radius expands from $1600\text{px} \rightarrow 2400\text{px}$.
     - Runner spawn ratio increases from $15\% \rightarrow 45\%-50\%$.

---

## 5. Verification Method

Independent verification of the Dungeon Master system and gameplay mechanics can be executed using the following commands and unit tests:

### 5.1 Verification Commands
```bash
# 1. Run all Go package tests across repository
CC=gcc go test -v ./...

# 2. Run dedicated Dungeon Master unit test suite
CC=gcc go test -v -run "TestDungeonMaster" ./internal/game

# 3. Run combat, armor, and AI stress tests
CC=gcc go test -v -run "TestCombat|TestArmor|TestZombie" ./internal/game

# 4. Launch full game binary
CC=gcc go run ./cmd/game
```

### 5.2 Test Invariants & Assertion Plan
1. **Dynamic Wave Invariant**:
   - In a headless test world, start with 0 zombies. Advance game time by 30 seconds (`1800` ticks). Assert `zombieCount > 0` and all spawned zombies are on non-solid walkable tiles $\ge 700\text{px}$ from player.
2. **Difficulty Progression Invariant**:
   - Evaluate wave size and runner ratio at `Day 1, 12:00` vs `Day 3, 00:00 (Night)`. Assert `WaveSize(Day 3 Night) > WaveSize(Day 1 Day)` and `RunnerRatio(Night) > RunnerRatio(Day)`.
3. **Zombie Death Drop Invariant**:
   - Kill 100 zombies in a headless test harness with `ZombieDropChance = 1.0`. Assert exactly 100 loot entities are spawned at zombie death coordinates with valid item types.
4. **Night Aggression Invariant**:
   - Verify `GetAggressionModifiers(0.0)` (Midnight) returns `speedMult >= 1.25` and `noiseMult >= 1.50`, whereas `GetAggressionModifiers(12.0)` (Noon) returns `1.0, 1.0, 1.0`.

### 5.3 Invalidation Conditions
- Any zombie wave spawning on solid obstacles (`TileWall`, `TileTree`, `TileFence`, etc.) or outside map bounds.
- Spawning zombies directly in the player's immediate view cone ($< 600\text{px}$).
- Frame rate degradation caused by unbounded entity counts ($> 200$ active entities).
- Inability to compile or execute under `CC=gcc go test ./...`.

