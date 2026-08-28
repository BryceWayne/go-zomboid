# Milestone 4 Melee Weapon Expansion: Fire Axe Cleave, Reach, Durability, and Inventory Equipping

## 1. Observation

### 1.1 Existing Component Definition
In `internal/ecs/components.go:29-47`:
```go
type Player struct {
	Health             float64
	Hunger             float64 // 100.0 is full, 0.0 is starving
	Thirst             float64 // 100.0 is hydrated, 0.0 is dehydrated
	Inventory          []string
	WeaponEquipped     bool
	WeaponDurability   int
	ArmorEquipped      bool
	ArmorType          string
	ArmorDefense       float64
	ArmorDurability    int
	ArmorMaxDurability int
	InfectionResist    float64
	AttackCooldown     int
	Dead               bool
	Infected           bool
	FacingX            float64
	FacingY            float64
}
```
Observation: `ecs.Player` currently lacks `WeaponType string`. To distinguish `"axe"` vs `"weapon"` (bat) vs unarmed fists, `WeaponType` is required.

### 1.2 Existing Item Usage and Inventory Key Handling
In `internal/game/game.go:277-319`:
```go
useItemIdx := -1
if ebiten.IsKeyPressed(ebiten.Key1) { useItemIdx = 0 }
if ebiten.IsKeyPressed(ebiten.Key2) { useItemIdx = 1 }
if ebiten.IsKeyPressed(ebiten.Key3) { useItemIdx = 2 }
if ebiten.IsKeyPressed(ebiten.Key4) { useItemIdx = 3 }
if ebiten.IsKeyPressed(ebiten.Key5) { useItemIdx = 4 }
if ebiten.IsKeyPressed(ebiten.Key6) { useItemIdx = 5 }
if ebiten.IsKeyPressed(ebiten.Key7) { useItemIdx = 6 }
if ebiten.IsKeyPressed(ebiten.Key8) { useItemIdx = 7 }
if ebiten.IsKeyPressed(ebiten.Key9) { useItemIdx = 8 }

if useItemIdx >= 0 && useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
	player.AttackCooldown = 30 // Small cooldown so it doesn't instantly consume everything if held
	t := player.Inventory[useItemIdx]
	
	used := false
	if t == "food" && player.Hunger < 100 {
		player.Hunger += 50
		if player.Hunger > 100 { player.Hunger = 100 }
		used = true
	} else if t == "water" && player.Thirst < 100 {
		player.Thirst += 50
		if player.Thirst > 100 { player.Thirst = 100 }
		used = true
	} else if t == "weapon" {
		player.WeaponEquipped = true
		player.WeaponDurability = 5
		used = true
	} else if t == "armor" || t == "vest" {
...
```
Observation:
- Pressing 1-9 checks `player.Inventory[useItemIdx]`.
- `"weapon"` sets `WeaponEquipped = true` and `WeaponDurability = 5`.
- `"axe"` is currently not handled in the equipping `if/else` ladder.

### 1.3 Existing Combat Processing
In `internal/game/game.go:345-388`:
```go
isAttacking := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyX)
if isAttacking && player.AttackCooldown <= 0 {
	player.AttackCooldown = 30 // Half second cooldown

	attackX := pos.X + player.FacingX*24
	attackY := pos.Y + player.FacingY*24
	
	hitZombies := false
	zQuery := s.zombieFilter.Query()
	for zQuery.Next() {
		z, zPos, zVel := zQuery.Get()
		ent := zQuery.Entity()
		
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx + dy*dy) < 24.0 { // Hit radius
			hitZombies = true
			if player.WeaponEquipped {
				toRemoveZombies = append(toRemoveZombies, ent)
			} else {
				// Shove!
				z.StunTimer = 45
				zVel.X = player.FacingX * 5.0
				zVel.Y = player.FacingY * 5.0
			}
		}
	}
	
	if player.WeaponEquipped {
		if hitZombies {
			assets.PlaySound(assets.HitSound)
			player.WeaponDurability--
			if player.WeaponDurability <= 0 {
				player.WeaponEquipped = false
			}
		} else {
			// Swoosh sound? Just play shove sound for now
			assets.PlaySound(assets.ShoveSound)
		}
	} else {
		assets.PlaySound(assets.ShoveSound)
	}
}
```
Observation:
- Hardcoded fixed offset `reach = 24.0` and `hitRadius = 24.0`.
- Does not distinguish weapon type reach, sweep angle, or durability parameters.
- When durability reaches 0, `WeaponEquipped` is set to `false`, but `WeaponType` should also be reset to `""`.

### 1.4 Asset & World Loot Status
- `cmd/tools/genassets/main.go:43, 1597-1675`: Generates pixel-art `axe.png` (red and steel double-beveled fire axe).
- `internal/assets/assets.go:38, 68`: Declares and loads `assets.AxeImage`.
- `internal/game/world/map.go:773, 788, 799`: Contextually spawns `"axe"` items in residential bedrooms, armories, and warehouse bays.
- `internal/game/game.go:763-765`: `DrawSystem.Draw` already maps `"axe"` ground items to `assets.AxeImage`.

---

## 2. Logic Chain

1. **ECS Data Model**:
   - Adding `WeaponType string` to `ecs.Player` (`internal/ecs/components.go`) allows runtime differentiation between:
     - Unarmed: `WeaponEquipped == false`, `WeaponType == ""`
     - Spiked Bat: `WeaponEquipped == true`, `WeaponType == "weapon"`, `WeaponDurability <= 5`
     - Fire Axe: `WeaponEquipped == true`, `WeaponType == "axe"`, `WeaponDurability <= 12`

2. **Inventory Equipping**:
   - When keys 1-9 are pressed and `t == "axe"`:
     - Sets `player.WeaponEquipped = true`
     - Sets `player.WeaponType = "axe"`
     - Sets `player.WeaponDurability = 12`
     - Sets `used = true`, removing the item from `player.Inventory` and setting cooldown `player.AttackCooldown = 30`.
   - When `t == "weapon"`:
     - Sets `player.WeaponEquipped = true`
     - Sets `player.WeaponType = "weapon"`
     - Sets `player.WeaponDurability = 5`
     - Sets `used = true`, removing the item from `player.Inventory`.

3. **Melee Reach and Cleave Sweep Mechanics**:
   - **Spiked Bat (`"weapon"`)**:
     - Reach: `24.0` px along facing vector: `attackX = pos.X + player.FacingX * 24.0`, `attackY = pos.Y + player.FacingY * 24.0`.
     - Hit Radius: `24.0` px around attack center.
     - Coverage: forward reach up to 48px from player position.
   - **Fire Axe (`"axe"`)**:
     - Reach: `32.0` px along facing vector: `attackX = pos.X + player.FacingX * 32.0`, `attackY = pos.Y + player.FacingY * 32.0`.
     - Hit Radius: `32.0` px around attack center.
     - Coverage: forward reach up to 64px from player position, and lateral sweep $\pm 32$ px (wide $180^\circ$ swing sweep), catching all multiple zombies within the arc and killing them simultaneously (cleave).
   - **Unarmed (Fists / Shove)**:
     - Reach: `24.0` px, Radius: `24.0` px.
     - Does not remove zombies; instead applies stun (`z.StunTimer = 45`) and pushback velocity (`zVel.X = player.FacingX * 5.0`, `zVel.Y = player.FacingY * 5.0`).
     - Plays `assets.ShoveSound`.

4. **Durability Degradation & Breakage**:
   - When an armed swing connects with 1 or more zombies:
     - Plays `assets.HitSound`.
     - Deducts 1 durability point: `player.WeaponDurability--`.
     - If `player.WeaponDurability <= 0`:
       - `player.WeaponEquipped = false`
       - `player.WeaponType = ""`
       - `player.WeaponDurability = 0`
   - If an armed swing misses (no zombies in range):
     - Plays `assets.ShoveSound` (swoosh).
     - Durability is NOT deducted on a miss.

5. **HUD UI Feedback**:
   - In `DrawSystem.Draw()`:
     - If `hasWeapon`:
       - Formats weapon name: `"FIRE AXE"` for `"axe"`, `"SPIKED BAT"` for `"weapon"`.
       - Displays: `fmt.Sprintf("Weapon: %s (Durability: %d) (Press SPACE/X to attack)", weaponName, playerDurability)` at Y: 95.
     - If unarmed:
       - Displays: `"Weapon: NONE (Press SPACE/X to shove zombies back)"` at Y: 95.
     - Facing reticle:
       - If `"axe"`: Tint `(1.0, 0.2, 0.2, 0.8)`.
       - If `"weapon"`: Tint `(1.0, 0.0, 0.0, 0.7)`.
       - If unarmed: Tint `(1.0, 1.0, 0.0, 0.7)`.

---

## 3. Caveats

- **Ranged Weapons Compatibility**: In Milestone 4, ranged weapons (`"shotgun"`, `"ammo"`) will also be handled in `processInputAndCombat()`. The clean branching designed here (`reach`, `hitRadius`, `player.WeaponType`) allows seamless addition of ranged raycast / pellet cone mechanics without altering melee logic.
- **Attack Cooldown Throttle**: `player.AttackCooldown = 30` (0.5s at 60 FPS) provides balance against rapid spamming, preventing instant depletion of weapon durability.
- No caveats regarding asset availability or world generation; both `axe.png` and map spawns are fully operational.

---

## 4. Conclusion & Pure Go Implementation

### 4.1 Target File Changes

#### File 1: `internal/ecs/components.go`
```go
// Player marker component.
type Player struct {
	Health             float64
	Hunger             float64 // 100.0 is full, 0.0 is starving
	Thirst             float64 // 100.0 is hydrated, 0.0 is dehydrated
	Inventory          []string
	WeaponEquipped     bool
	WeaponType         string  // "axe", "weapon", etc.
	WeaponDurability   int
	ArmorEquipped      bool
	ArmorType          string
	ArmorDefense       float64
	ArmorDurability    int
	ArmorMaxDurability int
	InfectionResist    float64
	AttackCooldown     int
	Dead               bool
	Infected           bool
	FacingX            float64
	FacingY            float64
}
```

#### File 2: `internal/game/game.go` (`processInputAndCombat` & `DrawSystem.Draw`)

**Inventory Item Equipping (`processInputAndCombat`)**:
```go
				used := false
				if t == "food" && player.Hunger < 100 {
					player.Hunger += 50
					if player.Hunger > 100 { player.Hunger = 100 }
					used = true
				} else if t == "water" && player.Thirst < 100 {
					player.Thirst += 50
					if player.Thirst > 100 { player.Thirst = 100 }
					used = true
				} else if t == "weapon" {
					player.WeaponEquipped = true
					player.WeaponType = "weapon"
					player.WeaponDurability = 5
					used = true
				} else if t == "axe" {
					player.WeaponEquipped = true
					player.WeaponType = "axe"
					player.WeaponDurability = 12
					used = true
				} else if t == "armor" || t == "vest" {
					player.ArmorEquipped = true
					player.ArmorType = "vest"
					player.ArmorDefense = 0.50
					player.ArmorDurability = 10
					player.ArmorMaxDurability = 10
					player.InfectionResist = 0.70
					used = true
				}
```

**Melee Combat Hit Testing (`processInputAndCombat`)**:
```go
			// Combat
			if player.AttackCooldown > 0 {
				player.AttackCooldown--
			}
			
			isAttacking := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyX)
			if isAttacking && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30 // Half second cooldown

				reach := 24.0
				hitRadius := 24.0
				if player.WeaponEquipped && player.WeaponType == "axe" {
					reach = 32.0
					hitRadius = 32.0
				}

				attackX := pos.X + player.FacingX*reach
				attackY := pos.Y + player.FacingY*reach
				
				hitZombies := false
				zQuery := s.zombieFilter.Query()
				for zQuery.Next() {
					z, zPos, zVel := zQuery.Get()
					ent := zQuery.Entity()
					
					dx := attackX - zPos.X
					dy := attackY - zPos.Y
					if math.Sqrt(dx*dx + dy*dy) < hitRadius {
						hitZombies = true
						if player.WeaponEquipped {
							toRemoveZombies = append(toRemoveZombies, ent)
						} else {
							// Shove!
							z.StunTimer = 45
							zVel.X = player.FacingX * 5.0
							zVel.Y = player.FacingY * 5.0
						}
					}
				}
				
				if player.WeaponEquipped {
					if hitZombies {
						assets.PlaySound(assets.HitSound)
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
						}
					} else {
						// Swoosh sound
						assets.PlaySound(assets.ShoveSound)
					}
				} else {
					assets.PlaySound(assets.ShoveSound)
				}
			}
```

**HUD & Reticle Rendering (`DrawSystem.Draw`)**:
```go
	// Reticle indicator:
	if hasWeapon {
		if playerWeaponType == "axe" {
			op.ColorScale.Scale(1, 0.2, 0.2, 0.8) // Fire axe indicator
		} else {
			op.ColorScale.Scale(1, 0, 0, 0.7) // Red if weapon
		}
	} else {
		op.ColorScale.Scale(1, 1, 0, 0.7) // Yellow if shove
	}

	// Weapon Text UI (Y: 95):
	if hasWeapon {
		weaponName := "WEAPON"
		if playerWeaponType == "axe" {
			weaponName = "FIRE AXE"
		} else if playerWeaponType == "weapon" {
			weaponName = "SPIKED BAT"
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Weapon: %s (Durability: %d) (Press SPACE/X to attack)", weaponName, playerDurability), 10, 95)
	} else {
		ebitenutil.DebugPrintAt(screen, "Weapon: NONE (Press SPACE/X to shove zombies back)", 10, 95)
	}
```

### 4.2 Patch & Test Suite Artifacts
- **Patch file**: `.agents/teamwork_preview_explorer_m4_1/proposed_m4_melee.patch`
- **Unit test suite**: `.agents/teamwork_preview_explorer_m4_1/proposed_melee_test.go` (12 unit tests covering all components, equip logic, reach, cleave, degradation, and breakage).

---

## 5. Verification Method

1. **Compilation Check**:
   ```bash
   CC=gcc go build -v -o bin/game ./cmd/game
   ```
   *Expected Result*: Exit code 0, compiles with zero warnings or type errors.

2. **Unit Test Suite**:
   ```bash
   CC=gcc go test -v ./...
   ```
   *Expected Result*: All existing tests in `internal/game`, `internal/game/world`, `internal/assets`, and `cmd/tools/genassets` pass (`PASS`).

3. **Melee Test Assertions (to be added to `internal/game/melee_test.go`)**:
   - `TestMelee_ECSComponentWeaponFields`: Confirms `WeaponType` field stores `"axe"`, `"weapon"`, `""`.
   - `TestMelee_EquipAxeFromInventory`: Confirms equipping `"axe"` sets `WeaponEquipped = true`, `WeaponType = "axe"`, `WeaponDurability = 12`.
   - `TestMelee_EquipBatFromInventory`: Confirms equipping `"weapon"` sets `WeaponEquipped = true`, `WeaponType = "weapon"`, `WeaponDurability = 5`.
   - `TestMelee_UnarmedShoveMechanics`: Confirms fists stun zombies (timer 45) and apply pushback velocity (5.0, 0.0) without entity deletion.
   - `TestMelee_BatReachAndKill`: Confirms Bat hits at 24px reach, decrements durability from 5 to 4, and kills zombie.
   - `TestMelee_BatOutOfReach`: Confirms Bat misses zombie at distance 55px.
   - `TestMelee_AxeExtendedReach`: Confirms Axe hits zombie at distance 55px (within 32px reach + 32px radius = 64px) where Bat misses.
   - `TestMelee_AxeMultiTargetCleave`: Confirms Axe cleaves 3 zombies simultaneously across wide forward/lateral sweep in a single swing, consuming 1 durability point.
   - `TestMelee_Axe12DurabilityDegradationAndBreakage`: Confirms Axe degrades over 12 consecutive hits and breaks on the 12th hit.
   - `TestMelee_Bat5DurabilityDegradationAndBreakage`: Confirms Bat degrades over 5 hits and breaks on the 5th hit.
   - `TestMelee_ReEquipOverridesStats`: Confirms equipping a new weapon cleanly overwrites previous weapon durability and type.
   - `TestMelee_UIWeaponTextFormatting`: Confirms UI text formatting for Fire Axe, Spiked Bat, and Unarmed states.

4. **Interactive Game Loop Launch**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   *Expected Result*: Launches 800x600 window without panics, allows picking up and equipping Fire Axe / Bat via inventory keys 1-9, shows HUD status, and performs melee attacks against zombies.
