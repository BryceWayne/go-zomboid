# Milestone 3 Investigation Report: Damage Mitigation, Infection Deflection & Durability Decay

## 1. Observation

### 1.1 Existing Zombie Attack and Infection Logic
- **Location**: `internal/game/game.go:446-453` in `UpdateSystem.processZombies()`
```go
		// Infection Check
		if dist < 14.0 && !playerDead {
			pMap := arkecs.NewMap1[ecs.Player](s.world)
			if playerComp := pMap.Get(playerEnt); playerComp != nil {
				playerComp.Infected = true
			}
		}
```
- **Observation**:
  - Whenever a zombie is within Euclidean distance `dist < 14.0` of a living player, `playerComp.Infected = true` is set unconditionally on every frame of contact.
  - There is currently no armor deflection roll, no defense calculation, and no durability tracking on contact.

### 1.2 Existing Health Drain Logic
- **Location**: `internal/game/game.go:234-255` in `UpdateSystem.processInputAndCombat()`
```go
		if player.Infected {
			player.Health -= 0.05 // Lose 3 health per second (takes ~33 seconds to die)
			if player.Health <= 0 {
				player.Dead = true
			}
		}

		// Drain Hunger and Thirst
		if !player.Dead {
			player.Hunger -= 0.003
			player.Thirst -= 0.005
			
			if player.Hunger < 0 { player.Hunger = 0 }
			if player.Thirst < 0 { player.Thirst = 0 }

			if player.Hunger == 0 || player.Thirst == 0 {
				player.Health -= 0.05
				if player.Health <= 0 {
					player.Dead = true
				}
			}
		}
```
- **Observation**:
  - When `player.Infected` is true, the player loses a constant `0.05` health per frame (at 60 FPS, 3.0 HP/s, taking ~33.3 seconds to deplete 100 HP).
  - Hunger/thirst depletion also drains health at `0.05` HP/frame when at 0.
  - There is no armor damage mitigation applied to the infection health drain or direct combat damage.

### 1.3 Target Player Component Definition
- **Location**: `PROJECT.md:62-83` (Milestone 3 Contract) and `internal/ecs/components.go:29-41`
- `ecs.Player` component specification:
```go
type Player struct {
	Health             float64
	Hunger             float64
	Thirst             float64
	Inventory          []string
	WeaponEquipped     bool
	WeaponType         string
	WeaponDurability   int
	ArmorEquipped      bool
	ArmorType          string
	ArmorDefense       float64 // e.g. 0.50 for 50% damage reduction
	ArmorDurability    int     // e.g. 10 hits
	ArmorMaxDurability int     // e.g. 10
	InfectionResist    float64 // e.g. 0.70 for 70% chance to block infection
	AttackCooldown     int
	Dead               bool
	Infected           bool
	FacingX, FacingY   float64
}
```

---

## 2. Logic Chain

1. **Zombie Contact Evaluation (< 14px)**:
   - In `processZombies()`, when `dist < 14.0 && !playerDead`, a zombie makes physical contact with the player entity.
   - We query the player's component `playerComp := pMap.Get(playerEnt)`.

2. **Infection Deflection Roll & Durability Decay**:
   - If `playerComp.ArmorEquipped == true`:
     - If the player is not yet infected (`!playerComp.Infected`):
       - Roll a random float: `roll := rand.Float64()`.
       - If `roll < playerComp.InfectionResist` (e.g. `0.70`), the armor successfully blocks and deflects the zombie's bite/scratch. Infection is not applied.
       - If `roll >= playerComp.InfectionResist`, the zombie bite penetrates the armor weave, setting `playerComp.Infected = true`.
     - In both deflection outcomes (and if already infected), the armor absorbs the physical impact, deducting durability:
       `playerComp.ArmorDurability--`.
     - If `playerComp.ArmorDurability <= 0`:
       The armor breaks and is destroyed:
       ```go
       playerComp.ArmorEquipped = false
       playerComp.ArmorType = ""
       playerComp.ArmorDefense = 0.0
       playerComp.ArmorDurability = 0
       playerComp.ArmorMaxDurability = 0
       playerComp.InfectionResist = 0.0
       ```
   - If `playerComp.ArmorEquipped == false`:
     - The player is unarmored and is directly exposed to the zombie bite:
       `playerComp.Infected = true`.

3. **Ongoing Health Drain Mitigation**:
   - In `processInputAndCombat()`:
     - When `player.Infected == true`:
       - Base drain is `0.05` HP/frame.
       - If `player.ArmorEquipped && player.ArmorDefense > 0`:
         `drain *= (1.0 - player.ArmorDefense)`
       - For example, with `ArmorDefense = 0.50` (50% reduction), `drain = 0.025` HP/frame (1.5 HP/s, extending life from 33.3s to 66.7s).
       - Subtract `player.Health -= drain`.
       - If `player.Health <= 0`: `player.Dead = true`.
     - Starvation/dehydration health drain (`Hunger == 0 || Thirst == 0`) is internal physiological damage and is **not** mitigated by body armor.

4. **Swarm Multi-Hit Dynamics**:
   - In `processZombies()`, if multiple zombies are in contact range within the same frame:
     - The query loop processes each zombie against the shared `playerComp`.
     - If the first zombie hits and depletes the remaining durability (breaking the armor), `playerComp.ArmorEquipped` immediately becomes `false`.
     - The next zombie in the swarm during that same frame observes `playerComp.ArmorEquipped == false` and applies infection.
     - This creates realistic swarm mechanics where armor protects against initial hits but collapses under concentrated assault.

---

## 3. Caveats

1. **RNG Reproducibility in Tests**: When writing unit tests for deflection probability, tests should avoid nondeterministic flakiness by either testing boundary values (`InfectionResist = 1.0` for guaranteed deflection, `InfectionResist = 0.0` for guaranteed penetration) or seeding `rand.Seed` / testing statistical distributions over 1000+ iterations.
2. **Durability per Frame vs Invulnerability Window**: In real-time gameplay at 60 FPS, continuous contact across frames will tick durability down on each frame a zombie is within 14px. With starting durability 10, continuous contact without shoving/running will break armor in ~10 frames (~0.16s). This encourages active shoving (`SPACE`/`X`) to create distance and space.
3. **Equipping Replacement Armor**: When a player equips a new armor vest from inventory, it overwrites any existing armor state and resets `ArmorDurability` to `ArmorMaxDurability` (10).

---

## 4. Conclusion & Concrete Code Formulation

### 4.1 Implementation in `internal/game/game.go:processZombies()`

Replace lines 446-453 in `internal/game/game.go` with:

```go
		// Zombie Contact, Infection Deflection & Armor Durability Decay (<14px)
		if dist < 14.0 && !playerDead {
			pMap := arkecs.NewMap1[ecs.Player](s.world)
			if playerComp := pMap.Get(playerEnt); playerComp != nil {
				if playerComp.ArmorEquipped {
					// Deflection roll against player's InfectionResist (e.g. 0.70 = 70% chance to block infection)
					if !playerComp.Infected {
						if rand.Float64() < playerComp.InfectionResist {
							// Deflected! Armor blocked the zombie bite/scratch.
						} else {
							// Deflection failed! Zombie bite penetrated the armor.
							playerComp.Infected = true
						}
					}
					// Deduct armor durability on contact hit
					playerComp.ArmorDurability--
					if playerComp.ArmorDurability <= 0 {
						// Armor broke under the zombie attack!
						playerComp.ArmorEquipped = false
						playerComp.ArmorType = ""
						playerComp.ArmorDefense = 0.0
						playerComp.ArmorDurability = 0
						playerComp.ArmorMaxDurability = 0
						playerComp.InfectionResist = 0.0
					}
				} else {
					// Unarmored player takes immediate infection on contact
					playerComp.Infected = true
				}
			}
		}
```

### 4.2 Implementation in `internal/game/game.go:processInputAndCombat()`

Replace lines 234-240 in `internal/game/game.go` with:

```go
		if player.Infected {
			drain := 0.05 // Base: Lose 3 health per second (takes ~33 seconds to die)
			if player.ArmorEquipped && player.ArmorDefense > 0 {
				drain *= (1.0 - player.ArmorDefense)
			}
			player.Health -= drain
			if player.Health <= 0 {
				player.Dead = true
			}
		}
```

---

## 5. Verification Method

### 5.1 Unit Test Specifications (`internal/game/armor_mitigation_test.go`)

The following test suite directly verifies all 5 requirements:

```go
package game

import (
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestZombieContact_UnarmoredDirectInfection verifies unarmored player gets infected immediately on contact (<14px)
func TestZombieContact_UnarmoredDirectInfection(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:        100.0,
			ArmorEquipped: false,
			Infected:      false,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Spawn zombie at dist = 10.0 (< 14.0)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	sys.processZombies()

	player := arkecs.NewMap1[ecs.Player](w).Get(pEnt)
	if !player.Infected {
		t.Fatal("Expected unarmored player to be infected on zombie contact")
	}
}

// TestZombieContact_ArmoredDeflectionSuccess verifies armor with 100% resist blocks infection and reduces durability
func TestZombieContact_ArmoredDeflectionSuccess(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			ArmorEquipped:      true,
			ArmorType:          "vest",
			ArmorDefense:       0.50,
			ArmorDurability:    10,
			ArmorMaxDurability: 10,
			InfectionResist:    1.0, // 100% deflection guarantee
			Infected:           false,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Spawn zombie at dist = 10.0 (< 14.0)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	sys.processZombies()

	player := arkecs.NewMap1[ecs.Player](w).Get(pEnt)
	if player.Infected {
		t.Fatal("Expected armor with 1.0 InfectionResist to deflect infection")
	}
	if player.ArmorDurability != 9 {
		t.Fatalf("Expected armor durability 9, got %d", player.ArmorDurability)
	}
	if !player.ArmorEquipped {
		t.Fatal("Expected armor to remain equipped after 1 hit")
	}
}

// TestZombieContact_ArmoredDeflectionFailure verifies 0% resist fails deflection and infects player
func TestZombieContact_ArmoredDeflectionFailure(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			ArmorEquipped:      true,
			ArmorType:          "vest",
			ArmorDefense:       0.50,
			ArmorDurability:    10,
			ArmorMaxDurability: 10,
			InfectionResist:    0.0, // 0% deflection (guaranteed failure)
			Infected:           false,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	sys.processZombies()

	player := arkecs.NewMap1[ecs.Player](w).Get(pEnt)
	if !player.Infected {
		t.Fatal("Expected 0% InfectionResist to result in infection")
	}
	if player.ArmorDurability != 9 {
		t.Fatalf("Expected armor durability 9, got %d", player.ArmorDurability)
	}
}

// TestZombieContact_ArmorBreakageAtZeroDurability verifies armor breaks when durability reaches 0
func TestZombieContact_ArmorBreakageAtZeroDurability(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			ArmorEquipped:      true,
			ArmorType:          "vest",
			ArmorDefense:       0.50,
			ArmorDurability:    1, // 1 hit remaining before breaking
			ArmorMaxDurability: 10,
			InfectionResist:    1.0,
			Infected:           false,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	sys.processZombies()

	player := arkecs.NewMap1[ecs.Player](w).Get(pEnt)
	if player.ArmorEquipped {
		t.Fatal("Expected armor to break and unequip when durability reaches 0")
	}
	if player.ArmorDurability != 0 {
		t.Fatalf("Expected armor durability 0, got %d", player.ArmorDurability)
	}
	if player.ArmorDefense != 0.0 {
		t.Fatalf("Expected armor defense 0.0, got %f", player.ArmorDefense)
	}
	if player.InfectionResist != 0.0 {
		t.Fatalf("Expected infection resist 0.0, got %f", player.InfectionResist)
	}
}

// TestInfectionDrain_ArmorMitigation verifies health drain is reduced by (1.0 - ArmorDefense)
func TestInfectionDrain_ArmorMitigation(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	// Armored Player: 50% defense
	pArmoredEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			Hunger:             100.0,
			Thirst:             100.0,
			Infected:           true,
			ArmorEquipped:      true,
			ArmorDefense:       0.50,
			ArmorDurability:    10,
			ArmorMaxDurability: 10,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	sys.processInputAndCombat()

	armoredPlayer := arkecs.NewMap1[ecs.Player](w).Get(pArmoredEnt)
	expectedArmoredHealth := 100.0 - (0.05 * (1.0 - 0.50)) // 99.975
	if math.Abs(armoredPlayer.Health-expectedArmoredHealth) > 1e-6 {
		t.Fatalf("Expected armored health %f, got %f", expectedArmoredHealth, armoredPlayer.Health)
	}
}
```

### 5.2 Command Verification
1. **Compilation & Unit Tests**:
   ```bash
   CC=gcc go test -v ./...
   ```
2. **Game Launch Check**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
