# Handoff Report: Milestone 2 — Dungeon Master Simulation (R2)

**Author**: Worker M2 (implementer, qa, specialist)  
**Date**: 2026-08-29  
**Target Repository**: `go-zomboid` (`/home/bryce/code/go-zomboid`)  
**Scope**: Dynamic Zombie Wave Spawning, Threat Difficulty Scaling, Dynamic & Randomized Loot Drops, Ambient Supply Restock, Day/Night Cycle & Aggression Scaling, Ambient Lighting Overlay, Game Loop Integration, Unit Testing Suite.

---

## 1. Observation

### 1.1 Codebase State Before Implementation
- `internal/game/world/map.go`: Map grid size $100 \times 100$, tile unit size $128\text{px}$, procedural town with 5 building archetypes (Residential, Grocery, Pharmacy, Police, Warehouse) and contextual loot.
- `internal/game/game.go`:
  - Static population: spawned 140 zombies once in `Reset()`. Kills permanently reduced zombie count without wave reinforcements.
  - Resource starvation: no ground loot replenishment; zombie deaths produced zero drops.
  - Decoupled lighting and AI: `g.timeOfDay` modulated a simple cosine darkness alpha in `DrawSystem.Draw()`, but zombie speed, noise detection radius, and vision radius were completely static constants day and night.

### 1.2 Implemented Changes & Additions
1. **Dungeon Master Subsystem (`internal/game/dm.go`)**:
   - Implemented `DungeonMasterConfig` and `DungeonMaster` struct managing ECS mappers (`zombieMap`, `itemMap`, `playerMap`) and query filters.
   - **Dynamic Zombie Wave Spawning**:
     - Periodically triggers waves every 1800 ticks (30s) or dynamically when living zombie count drops below 15.
     - Threat scaling formula:
       $$\text{Threat}(t) = 1.0 + \frac{\text{TotalTicks}}{60 \times 180} + 0.25 \times (\text{DayCount} - 1) + (0.50 \text{ if Night else } 0.0)$$
     - Wave size: $\text{clamp}(\lfloor\text{BaseZombies} \times \text{Threat}(t)\rfloor, 3, 16)$.
     - Spawns candidate zombies at perimeter distance $[700\text{px}, 1600\text{px}]$ from player on non-solid walkable tiles (`!gameMap.GetTile(tx, ty).IsSolid()` and `!gameMap.IsColliding(...)`).
     - Runner probability scaling: 15% during Day (08:00 - 17:00), 45% at Night (20:00 - 05:00), with smooth linear interpolation during Dawn (05:00 - 08:00) and Dusk (17:00 - 20:00).
     - Caps active population to `MaxLivingZombies = 140`.
   - **Dynamic & Randomized Loot Drops**:
     - `HandleZombieDeath(wx, wy float64)`: 25% drop rate upon kill (`ZombieDropChance = 0.25`), weighted across 8 items (`ammo`: 30%, `food`: 25%, `water`: 20%, `weapon`: 10%, `antidote`: 8%, `axe`: 4%, `armor`: 2%, `shotgun`: 1%). Instantiates item entity at `(wx, wy)` when ground items $< \text{MaxMapItems}$ (60).
     - `SpawnAmbientSupplies(count int)`: Every 3600 ticks (60s), rolls 2–4 items in interior building rooms or walkable tiles if total items $< \text{MaxMapItems}$.
   - **Day/Night Cycle & Aggression Scaling**:
     - `GetAggressionModifiers(timeOfDay float64) (speedMult, noiseMult, visionMult float64)`:
       - Daytime (08:00 - 17:00): `1.0, 1.0, 1.0`.
       - Night (20:00 - 05:00): `speedMult >= 1.25` (peaking at `1.35` at Midnight), `noiseMult >= 1.50` (peaking at `1.75`), `visionMult >= 1.25` (peaking at `1.35`).
       - Dawn / Dusk: Smooth linear ramp between day base and night multipliers.
     - `GetAmbientLighting(timeOfDay float64) (color.RGBA, float64)`:
       - Dawn (05:00 - 08:00): Warm rose/gold tint (`color.RGBA{180, 140, 80, 255}`), alpha transitioning $0.55 \rightarrow 0.0$.
       - Day (08:00 - 17:00): Clear sunlight, alpha $= 0.0$.
       - Dusk (17:00 - 20:00): Amber twilight tint (`color.RGBA{140, 60, 50, 255}`), alpha transitioning $0.0 \rightarrow 0.60$.
       - Night (20:00 - 05:00): Midnight navy overlay (`color.RGBA{5, 10, 35, 255}`), alpha peaking at $0.88$ at 00:00 (Midnight).
2. **Game Loop Integration (`internal/game/game.go`)**:
   - Added `dm *DungeonMaster` to `Game`, `UpdateSystem`, and `DrawSystem`.
   - Initialized in `Game.Reset()`, `NewGame()`, `NewUpdateSystem()`, and `NewDrawSystem()`.
   - In `UpdateSystem.Update()`: Dispatches `dm.Update(timeOfDay, playerPos)` each tick.
   - In `UpdateSystem.processInputAndCombat()`: Dispatches `dm.HandleZombieDeath(zPos.X, zPos.Y)` on all zombie kills.
   - In `UpdateSystem.processZombies()`: Applies `speedMult`, `noiseMult`, and `visionMult` from `dm.GetAggressionModifiers(timeOfDay)` to zombie physics and detection radii.
   - In `DrawSystem.Draw()`: Employs `dm.GetAmbientLighting(timeOfDay)` to render day/night atmospheric tint and darkness.
3. **Unit Tests (`internal/game/dm_test.go`)**:
   - Implemented 9 test suites verifying wave scaling, perimeter validity, death drops, ambient restock, day/night lighting, aggression modifiers, runner scaling, max cap enforcement, and game loop execution.

---

## 2. Logic Chain

1. **Dynamic Wave Invariant**:
   - *Observation*: Without waves, game difficulty collapses as zombies are eliminated.
   - *Implementation*: DM evaluates wave triggers every 1800 ticks or when zombie count $< 15$. Threat scales continuously with ticks and day count, adding a $+0.50$ night threat multiplier.
   - *Result*: Waves dynamically scale from 3 base zombies up to 16 maximum, ensuring long-term challenge without overflowing ECS memory.

2. **Perimeter Spawn Non-Trapping Guarantee**:
   - *Observation*: Spawning zombies inside walls or on top of players causes instant collision or clipping bugs.
   - *Implementation*: DM candidate generator samples radial polar coordinates $r \in [700\text{px}, 1600\text{px}]$, computes tile coordinates, verifies map boundaries ($2 \le tx < W-2$), asserts `!GetTile(tx, ty).IsSolid()`, and checks 64x64 AABB collision against obstacles.
   - *Result*: 100% of candidate spawns occur on valid, open, walkable terrain outside the player's immediate viewport.

3. **Sustainable Resource Economy**:
   - *Observation*: Weapons have finite durability (Club: 5, Axe: 12, Shotgun: 15) and firearms require non-renewable ammo.
   - *Implementation*: Every zombie kill rolls a 25% chance of dropping loot from the 8-item weighted table (`ammo`, `food`, `water`, `weapon`, `antidote`, `axe`, `armor`, `shotgun`). Ambient supply drops periodically replenish 2–4 items in building rooms up to `MaxMapItems = 60`.
   - *Result*: Balanced risk/reward loop where combat and exploration sustain player survival.

4. **Atmospheric Day/Night Cycle & Aggression Shift**:
   - *Observation*: Night survival should demand heightened caution.
   - *Implementation*: At night (20:00 - 05:00), lighting shifts to deep navy tint ($\alpha \approx 0.88$), runner ratio increases from 15% to 45%, zombie speed scales up by $+25\%$ to $+35\%$, sound detection radius expands by $+50\%$ to $+75\%$, and vision expands by $+25\%$ to $+35\%$.
   - *Result*: Clear contrast between daytime scavenging and high-stakes nighttime horde defense.

---

## 3. Caveats

1. **ECS World Locking**:
   - Ark ECS query filters lock the world during iteration. All count methods (`CountZombies()`, `CountItems()`) explicitly invoke `defer q.Close()` or `q.Close()` to prevent locking panics during entity creation.
2. **Backwards Compatibility**:
   - Existing unit tests that instantiate `NewUpdateSystem(w, m)` or `NewDrawSystem(w, m)` without explicit DM arguments remain 100% supported by instantiating default DM instances internally.

---

## 4. Conclusion

Milestone 2 (Dungeon Master Simulation / R2) is fully implemented, verified, and integrated into the `go-zomboid` engine. All requirements — dynamic wave spawning, threat scaling, dynamic loot drops, ambient supply drops, day/night aggression modifiers, and day/night ambient lighting — are active, genuinely modeled, and covered by unit tests.

---

## 5. Verification Method

To independently verify the Milestone 2 implementation:

```bash
# 1. Run dedicated Dungeon Master and Orthogonal test suites
CC=gcc go test -count=1 -v -run "TestDungeonMaster|TestOrthogonal" ./internal/game

# 2. Run combat, armor, and AI regression suites
CC=gcc go test -count=1 -v -run "TestCombat|TestArmor|TestZombie" ./internal/game

# 3. Run world map test suite
CC=gcc go test -count=1 -v ./internal/game/world

# 4. Verify full package compilation and static analysis
CC=gcc go vet ./...
CC=gcc go build ./...
```
