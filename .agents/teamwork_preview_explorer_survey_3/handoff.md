# Survey Analysis Report: Items, Weapons, Armor, Combat, Damage Mechanics, and Test Suites

## 1. Observation

### 1.1 Project Structure & Inventory/Item System
- **Entity Component System**: Powered by Ark ECS (`github.com/mlange-42/ark v0.8.3`) and Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2 v2.9.10`).
- **Data Models** (`internal/ecs/components.go:28-57`):
  - `ecs.Player`: Contains `Health` (100.0), `Hunger` (100.0), `Thirst` (100.0), `Inventory` (`[]string`), `WeaponEquipped` (`bool`), `WeaponDurability` (`int`), `AttackCooldown` (`int`), `Dead` (`bool`), `Infected` (`bool`), `FacingX` (`float64`), `FacingY` (`float64`).
  - `ecs.Item`: Contains `Type string`.
  - `ecs.Zombie`: Contains `Speed` (`float64`), `Chasing` (`bool`), `IsRunner` (`bool`), `WanderTimer` (`int`), `WanderDirX`/`WanderDirY` (`float64`), `StunTimer` (`int`).
- **Item Spawning & World Distribution** (`internal/game/game.go:68-79`):
  - 3 items guaranteed near player spawn: `"weapon"`, `"food"`, `"water"`.
  - 20 random items placed across the world (5 `"weapon"`, 8 `"food"`, 7 `"water"`).
  - Ground items are drawn via `DrawSystem.Draw` (`internal/game/game.go:687-724`) with depth sorting based on `iPos.X + iPos.Y`.
- **Item Pickup Mechanics** (`internal/game/game.go:195-236`):
  - Processed in `UpdateSystem.processItems()`.
  - If Euclidean distance `math.Sqrt(dx*dx + dy*dy) < 16.0` between player and item:
    - If `len(player.Inventory) < 9`, appends `item.Type` to `player.Inventory` and removes item entity via `s.world.RemoveEntity(ent)`.
- **Inventory Usage Mechanics** (`internal/game/game.go:277-323`):
  - Number keys 1-9 (`ebiten.Key1` through `Key9`) map to slot indices 0-8.
  - When pressed and `player.AttackCooldown <= 0`: sets `player.AttackCooldown = 30` (frame throttle).
  - `"food"`: Restores +50 Hunger (clamped at 100).
  - `"water"`: Restores +50 Thirst (clamped at 100).
  - `"weapon"`: Sets `player.WeaponEquipped = true` and `player.WeaponDurability = 5`.
  - Consumed item is removed from `player.Inventory` slice.

### 1.2 Combat, Zombie Attacks & Damage Mechanics
- **Player Attack Execution** (`internal/game/game.go:345-392`):
  - Triggered by `ebiten.IsKeyPressed(ebiten.KeySpace)` or `ebiten.KeyX` when `player.AttackCooldown <= 0`.
  - Sets `player.AttackCooldown = 30` (0.5s at 60 FPS).
  - Attack center position: `attackX = pos.X + player.FacingX*24`, `attackY = pos.Y + player.FacingY*24`.
  - Hit radius: 24.0 pixels around attack center.
  - Armed Mode (`player.WeaponEquipped == true`):
    - Instant kill: all zombies within radius are removed (`toRemoveZombies = append(toRemoveZombies, ent)`).
    - Plays `assets.HitSound`.
    - `player.WeaponDurability--`. If `<= 0`, sets `player.WeaponEquipped = false`.
    - If no zombies hit, plays `assets.ShoveSound`.
  - Unarmed Mode (`player.WeaponEquipped == false`):
    - Shove / Stun: sets `z.StunTimer = 45`, `zVel.X = player.FacingX * 5.0`, `zVel.Y = player.FacingY * 5.0`.
    - Plays `assets.ShoveSound`.
- **Zombie Attacks & Player Damage** (`internal/game/game.go:250-271`, `462-468`):
  - Proximity check: in `processZombies()`, if `dist(player, zombie) < 14.0`:
    - `playerComp.Infected = true`.
  - Continuous Damage:
    - If `player.Infected`: `player.Health -= 0.05` per frame (lose 3.0 HP/s; fatal in ~33.3s).
    - If `player.Hunger == 0` or `player.Thirst == 0`: `player.Health -= 0.05` per frame.
    - When `player.Health <= 0`: `player.Dead = true`.
  - Observation: There is currently **no discrete instant bite/scratch damage** and **no armor mitigation mechanism** whatsoever.

### 1.3 Weapon Types & Asset Generation
- **Current Weapons**: Only one generic melee club weapon (`"weapon"`) exists.
- **Asset Pipeline** (`cmd/tools/genassets/main.go:22, 208-216`):
  - `generateWeapon("weapon.png")` creates a 16x16 brown diagonal stick.
  - Saved to `internal/assets/images/weapon.png` and embedded into `internal/assets/assets.go` via `//go:embed images/*`.
  - `assets.Load()` loads images into `*ebiten.Image` globals.
- **Audio System** (`internal/assets/audio.go:16-58`):
  - Procedurally synthesizes raw PCM audio buffers for `HitSound` (white noise burst) and `ShoveSound` (pitch-sweeping sine wave).
  - Played via `AudioContext.NewPlayerFromBytes(data)`.

### 1.4 Test Suites, Build Commands & Game Loop
- **Build Command**: `CC=gcc go build -o bin/game ./cmd/game` succeeds without errors.
- **Test Command**: `CC=gcc go test -v ./...` executes and passes:
  - `internal/game`: `TestWorldToIso` (isometric projection coordinate transformations).
  - `internal/game/world`: `TestNewMap`, `TestIsColliding` (map bounds and AABB collision).
  - Output: All tests pass (`ok 0.022s`).
- **Asset Generation Command**: `go run ./cmd/tools/genassets` generates 11 PNG files in `internal/assets/images/`.
- **Game Entry Point** (`cmd/game/main.go:11-22`):
  - Initializes window (800x600), loads assets, calls `game.NewGame()`, and enters `ebiten.RunGame(g)`.
  - `Game.Reset()` initializes ECS world, spawns 150 zombies, player, items, and map.

---

## 2. Logic Chain

1. **Item & Inventory Extensibility**:
   - Items are string-tagged (`ecs.Item{Type: string}`).
   - The inventory is a slice `[]string` with 9 slots.
   - When a slot key (1-9) is pressed in `processInputAndCombat()`, it inspects `player.Inventory[useItemIdx]`.
   - Extending item types simply requires handling additional string identifiers (e.g. `"armor"`, `"vest"`, `"shotgun"`, `"axe"`, `"spear"`, `"ammo"`) in:
     a. `processInputAndCombat()` (consumption & equip logic).
     b. `cmd/tools/genassets` (sprite generator functions).
     c. `internal/assets` (embedded images and globals).
     d. `DrawSystem.Draw` (item world drop rendering and UI slot text).

2. **Weapon System Architecture**:
   - Currently, `ecs.Player` only holds `WeaponEquipped bool` and `WeaponDurability int`.
   - To support new melee and ranged weapon types:
     - Introduce a structured weapon registry or extended player fields (e.g. `EquippedWeaponType string`, `WeaponDurability int`, `WeaponAmmo int`, `WeaponRange float64`, `WeaponCooldown int`, `WeaponNoiseRadius float64`).
     - **Melee Archetypes**:
       - Axe (`"axe"`): Heavy damage/durability (10-15 hits), wider sweep angle, cleaves multiple zombies.
       - Spear (`"spear"`): Extended reach (40-48px vs standard 24px), narrow thrust hitbox, lower cooldown.
       - Knife (`"knife"`): Fast attack (15 frame cooldown vs 30), single target, low durability.
     - **Ranged Archetypes**:
       - Shotgun (`"shotgun"`) & Ammo (`"ammo"`): Consumes ammo from inventory, shoots a multi-pellet spread cone (e.g., 3 rays over 30 degrees up to 160px), knocks back and kills zombies, generates large noise pulse (`noiseRadius = 400px`) that alerts the zombie swarm.
       - Pistol (`"pistol"`): Fast linear raycast up to 220px, single target, moderate noise (`noiseRadius = 220px`).
       - Crossbow (`"crossbow"`): Silent ranged shot (`noiseRadius = 25px`), high precision, recoverable ammo.

3. **Armor System Architecture (Damage Mitigation, Durability & UI)**:
   - **Data Representation**:
     - Extend `ecs.Player` in `internal/ecs/components.go`:
       ```go
       ArmorEquipped      bool
       ArmorType          string   // e.g. "vest", "jacket", "riot"
       ArmorDefense       float64  // percentage mitigation, e.g. 0.50 (50% reduction)
       ArmorDurability    int      // hits / durability points remaining
       ArmorMaxDurability int
       InfectionResist    float64  // probability of deflecting infection on contact (e.g., 0.70)
       ```
   - **Damage & Infection Mitigation**:
     - In `processZombies()`, when zombie touches player (`dist < 14.0`):
       - If `player.ArmorEquipped`:
         - Roll infection deflection: if `rand.Float64() < player.InfectionResist`, infection is deflected!
         - Deduct durability: `player.ArmorDurability--`.
         - If `player.ArmorDurability <= 0`, break armor: `player.ArmorEquipped = false`, `player.ArmorType = ""`, play break sound.
       - If `!player.ArmorEquipped`:
         - `player.Infected = true`.
     - Direct Damage / Decay Mitigation:
       - When taking damage from zombie attacks or environmental sources:
         `effectiveDamage := rawDamage * (1.0 - player.ArmorDefense)`
   - **UI Indicator**:
     - In `DrawSystem.Draw()` (`internal/game/game.go`):
       - Render an Armor bar below Health/Hunger/Thirst (X: 10, Y: 75, W: 200, H: 15).
       - Color: Steel Blue (`color.RGBA{70, 130, 180, 255}`).
       - Text: `fmt.Sprintf("Armor: %d/%d (Def: %d%%)", player.ArmorDurability, player.ArmorMaxDurability, int(player.ArmorDefense*100))`.
       - Adjust vertical positions of Weapon text (`y = 95`) and Infected status (`y = 115`).
       - Visual player indicator: Draw armor tint or chest plate highlight when `player.ArmorEquipped == true`.

4. **Testing Harness & Headless Game Loop Verification**:
   - `internal/game` systems can be instantiated headlessly without GUI/audio devices.
   - `assets.InitAudio()` initializes safely and `PlaySound` handles nil/unsupported audio gracefully.
   - Unit tests can construct `arkecs.World`, instantiate `UpdateSystem`, spawn entities, manipulate inputs/states, and assert game logic with 100% determinism.

---

## 3. Caveats

- **Ebitengine Headless Execution**: Ebitengine `RunGame()` blocks and requires a display window; however, game logic inside `UpdateSystem`, `DrawSystem`, and map generators can be fully exercised in standard Go unit tests (`go test ./...`) without opening a window.
- **Zombie Health Pool vs Instant Kill**: Currently zombies have no `Health` component and are deleted immediately upon weapon hit. If new weapons introduce variable damage (e.g. shotgun pellets dealing partial damage vs sniper rifle instant kill), adding a `Health float64` field to `ecs.Zombie` may be beneficial, though retaining instant-kill on standard weapons maintains simplicity and balance.
- **Audio Generation in CI**: Audio is procedurally synthesized in Go; no external sound files are needed.

---

## 4. Conclusion & Concrete Architectural Specification

### 4.1 Recommended Implementation Plan

#### Component 1: Armor System Implementation
1. **ECS Extension** (`internal/ecs/components.go`):
   - Add armor fields to `ecs.Player`: `ArmorEquipped bool`, `ArmorType string`, `ArmorDefense float64`, `ArmorDurability int`, `ArmorMaxDurability int`, `InfectionResist float64`.
2. **Item & Asset Generation** (`cmd/tools/genassets/main.go` & `internal/assets/assets.go`):
   - Add `generateArmor("armor.png", color.RGBA{70, 130, 180, 255})`.
   - Add `ArmorImage *ebiten.Image` in `assets.go`.
3. **Item Spawning & Pickup** (`internal/game/game.go`):
   - Include `"armor"` in world item spawns (`itemTypes`).
   - Equip armor on key press (1-9): sets `ArmorEquipped = true`, `ArmorDurability = 10`, `ArmorMaxDurability = 10`, `ArmorDefense = 0.50`, `InfectionResist = 0.70`.
4. **Mitigation Logic** (`internal/game/game.go:processZombies`):
   - Mitigate contact damage/infection, decrease armor durability, handle armor destruction when durability reaches 0.
5. **UI & HUD** (`internal/game/game.go:Draw`):
   - Add Armor HUD status bar and visual indicator on player character.

#### Component 2: New Weapon Types (Melee & Ranged)
1. **Weapon Definitions**:
   - Ranged Option: Shotgun (`"shotgun"`) with Ammo (`"ammo"`).
     - Range: 160px, cone spread (3 pellets), high knockback, noise radius 350px.
   - Melee Option: Axe (`"axe"`) and/or Spear (`"spear"`).
     - Axe: Durability 12, cleave attack.
     - Spear: Range 40px (1.7x normal club reach), single target thrust.
2. **Asset Generation**:
   - Add procedural generators in `genassets` for `shotgun.png`, `axe.png`, `ammo.png`.
   - Add asset bindings in `internal/assets/assets.go`.
3. **Combat Execution**:
   - Handle weapon-specific range, ammunition requirements, durability decay, and noise generation for zombie swarm aggro.

#### Component 3: Test Suite Expansion
1. Create comprehensive unit tests in `internal/game/game_test.go` and `internal/game/combat_test.go`:
   - `TestItemPickup_CapacityAndRemoval`: Verifies 9-slot inventory limit and world removal.
   - `TestEquipAndConsumeItems`: Verifies food/water replenishment, weapon equip, and armor equip.
   - `TestWeaponCombat_DurabilityAndCleave`: Verifies weapon durability loss, zombie removal, and shove stun/knockback.
   - `TestArmor_DamageMitigationAndDeflection`: Verifies armor durability loss, infection deflection probability, and armor breakage.
   - `TestRangedCombat_AmmoConsumption`: Verifies ammo requirement, projectile raycast hit, and noise alert radius.

---

## 5. Verification Method

To independently verify these findings and all subsequent implementations, run the following verification steps:

1. **Asset Generation Verification**:
   ```bash
   go run ./cmd/tools/genassets
   ```
   *Expected Result*: Exits 0, generates all PNG sprites in `internal/assets/images/`.

2. **Compilation & Build Verification**:
   ```bash
   CC=gcc go build -v -o bin/game ./cmd/game
   ```
   *Expected Result*: Compiles cleanly with exit code 0.

3. **Full Test Suite Execution**:
   ```bash
   CC=gcc go test -v ./...
   ```
   *Expected Result*: All package tests execute and pass.

4. **Interactive Game Loop Launch**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   *Expected Result*: Ebitengine initializes window, renders isometric map, player HUD, inventory, items, and zombie swarm without runtime panics or rendering glitches.
