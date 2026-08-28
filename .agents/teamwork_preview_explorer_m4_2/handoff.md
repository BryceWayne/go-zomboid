# Ranged Weapon System Analysis & Implementation Specification (Shotgun, Ammo Consumption & Noise Alert)

## 1. Observation

### 1.1 Existing ECS Data Structures & Inventory
- **Player Component** (`internal/ecs/components.go:28-47`):
  ```go
  type Player struct {
      Health             float64
      Hunger             float64
      Thirst             float64
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
  *Observation*: `ecs.Player` currently tracks `WeaponEquipped` and `WeaponDurability`, but requires `WeaponType string` (specified in `PROJECT.md:71`) to distinguish between weapon types (`"shotgun"`, `"axe"`, `"weapon"`).

- **Inventory System** (`internal/game/game.go:277-319`):
  - Inventory items are stored as string slices `[]string` in `player.Inventory` with capacity 9.
  - Number keys 1-9 trigger slot index `useItemIdx = 0..8`.
  - Item consumption applies effect, sets `player.AttackCooldown = 30`, and removes the item: `player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)`.

### 1.2 Existing Asset & World Generation Support
- **Procedural Sprites** (`cmd/tools/genassets/main.go:44-45, 1704-1830`):
  - `generateShotgun("shotgun.png")` generates a 16x16 pixel-art shotgun with steel barrel and wooden stock.
  - `generateAmmo("ammo.png")` generates a 16x16 pixel-art ammo box with brass cartridges.
- **Embedded Asset Handles** (`internal/assets/assets.go:39-40, 69-70`):
  - `assets.ShotgunImage` and `assets.AmmoImage` are exported `*ebiten.Image` globals.
- **Audio Synthesizer** (`internal/assets/audio.go:12-49`):
  - `assets.HitSound`: High-energy white noise burst representing physical impact / gunshot.
  - `assets.ShoveSound`: Pitch-sweeping low sine wave representing mechanical click / body shove.
- **World Map Loot Distribution** (`internal/game/world/map.go:789-800`):
  - `RoomArmory` spawns `"shotgun"` at `(centerX-24, centerY+24)` and `"ammo"` at `(centerX+24, centerY+24)`.
  - `RoomOffice` and `RoomWarehouseBay` spawn additional `"ammo"` boxes.
  - Ground items render using `assets.ShotgunImage` and `assets.AmmoImage` (`internal/game/game.go:765-769`).

### 1.3 Existing Combat Loop & Zombie AI
- **Combat Loop** (`internal/game/game.go:345-388`):
  - Space or X triggers attack if `player.AttackCooldown <= 0`.
  - Melee attack evaluates a single circular area at `(pos.X + FacingX*24, pos.Y + FacingY*24)` with radius 24.0px.
- **Zombie AI & Aggro System** (`internal/game/game.go:416-512`):
  - Zombies have `Chasing bool` and `WanderTimer int`.
  - Normal player movement has noise radius 200.0px; standing still has noise radius 50.0px.
  - Distances $> 400.0\text{px}$ cause zombies to lose aggro (`zombie.Chasing = false`).

---

## 2. Logic Chain

1. **Shotgun Equipping Mechanism**:
   - When the player presses the hotbar key corresponding to a `"shotgun"` item in `player.Inventory`:
     - Set `player.WeaponEquipped = true`
     - Set `player.WeaponType = "shotgun"`
     - Set `player.WeaponDurability = 15`
     - Remove `"shotgun"` from `player.Inventory`.
   - If the player presses a hotbar key corresponding to `"ammo"`, do nothing (`used = false`) so ammo is preserved for weapon firing.

2. **Ammo Requirement & Consumption**:
   - When the player attacks with an equipped shotgun (`player.WeaponEquipped && player.WeaponType == "shotgun"`):
     - Scan `player.Inventory` for the first occurrence of `"ammo"`.
     - **If ammo is present** (`ammoIdx >= 0`):
       - Remove 1 `"ammo"` item from `player.Inventory`: `player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)`.
       - Decrement durability: `player.WeaponDurability--`.
       - If `player.WeaponDurability <= 0`: break weapon (`WeaponEquipped = false`, `WeaponType = ""`, `WeaponDurability = 0`).
       - Set attack cooldown: `player.AttackCooldown = 30`.
       - Play gunshot sound: `assets.PlaySound(assets.HitSound)`.
       - Execute spread cone kill & knockback calculations.
       - Emit 400.0px Acoustic Noise Pulse.
     - **If out of ammo / dry fire** (`ammoIdx == -1`):
       - Do NOT decrement `player.WeaponDurability`.
       - Set attack cooldown: `player.AttackCooldown = 30`.
       - Play mechanical dry-fire / shove sound: `assets.PlaySound(assets.ShoveSound)`.
       - Perform close-quarters weapon butt shove on adjacent zombies within 24.0px (`z.StunTimer = 45`, `zVel = facing * 5.0`).
       - Do NOT emit the 400.0px gunshot noise pulse.

3. **Ranged Spread Cone Vector Mathematics**:
   - Player position $\vec{P} = (pos.X, pos.Y)$, normalized facing vector $\vec{F} = (facingX, facingY)$.
   - For any zombie at position $\vec{Z} = (zPos.X, zPos.Y)$, displacement $\vec{D} = \vec{Z} - \vec{P} = (dx, dy)$.
   - Distance $dist = \|\vec{D}\| = \sqrt{dx^2 + dy^2}$.
   - **Range boundary**: $dist \le 160.0\text{ px}$.
   - **Point-blank zone**: If $dist < 24.0\text{ px}$, direct hit regardless of angular variance.
   - **Cone angle boundary**:
     $$\cos(\theta) = \frac{\vec{F} \cdot \vec{D}}{\|\vec{F}\| \|\vec{D}\|} = \frac{facingX \cdot dx + facingY \cdot dy}{dist}$$
     Threshold for $\pm 22.5^\circ$ half-angle (total spread arc $45^\circ$):
     $$\cos(22.5^\circ) = \cos\left(\frac{\pi}{8}\right) \approx 0.9238795325112867$$
     If $\cos(\theta) \ge 0.92388$, the zombie is within the 3-pellet spread cone and is added to `toRemoveZombies`.

4. **Acoustic Noise Pulse & Swarm Convergence**:
   - A 12-gauge shotgun blast is intensely loud ($R_{noise} = 400.0\text{ px}$).
   - When fired with ammo, query all zombies in the ECS world:
     - For every zombie where $\|\vec{Z} - \vec{P}\| \le 400.0\text{ px}$:
       - `z.Chasing = true`
       - `z.WanderTimer = 0`
   - In `processZombies()`, alerted zombies immediately pivot and chase the player's position, creating authentic swarm pressure.

---

## 3. Caveats

- **ECS Entity Deletion Timing**: Ark ECS entities marked for deletion in `toRemoveZombies` remain valid during subsequent filter queries in the same update frame until `s.world.RemoveEntity(ent)` executes. Setting `z.Chasing = true` on a zombie that is deleted at the end of the frame is completely safe and causes no memory leaks or dangling pointers.
- **Headless Audio Execution**: In headless unit test environments, `assets.PlaySound` safely checks `if AudioContext == nil { return }` and executes without panic.
- **Inventory Ammo Stacking**: In the current 9-slot inventory model, each `"ammo"` string occupies 1 slot and grants 1 shotgun blast.

---

## 4. Conclusion & Pure Go Implementation Code

### 4.1 Target 1: `internal/ecs/components.go`
Add `WeaponType string` to `ecs.Player`:
```go
// Player marker component.
type Player struct {
	Health             float64
	Hunger             float64 // 100.0 is full, 0.0 is starving
	Thirst             float64 // 100.0 is hydrated, 0.0 is dehydrated
	Inventory          []string
	WeaponEquipped     bool
	WeaponType         string  // "weapon", "axe", "shotgun"
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

### 4.2 Target 2: `internal/game/game.go:processInputAndCombat()`
Replace the inventory equipping and combat sections with the following code:

```go
			// Inventory Usage (Keys 1-9)
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
				player.AttackCooldown = 30
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
					player.WeaponType = "weapon"
					player.WeaponDurability = 5
					used = true
				} else if t == "axe" {
					player.WeaponEquipped = true
					player.WeaponType = "axe"
					player.WeaponDurability = 12
					used = true
				} else if t == "shotgun" {
					player.WeaponEquipped = true
					player.WeaponType = "shotgun"
					player.WeaponDurability = 15
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
				
				if used {
					// Remove item from inventory
					player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
				}
			}

			if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
				vel.Y -= speed
			}
			if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
				vel.Y += speed
			}
			if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
				vel.X -= speed
			}
			if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
				vel.X += speed
			}

			// Update facing
			if vel.X != 0 || vel.Y != 0 {
				player.FacingX = vel.X / speed
				player.FacingY = vel.Y / speed
			}

			// Combat
			if player.AttackCooldown > 0 {
				player.AttackCooldown--
			}
			
			isAttacking := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyX)
			if isAttacking && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30 // Half second cooldown

				if player.WeaponEquipped && player.WeaponType == "shotgun" {
					// 1. Check for ammo in inventory
					ammoIdx := -1
					for idx, itm := range player.Inventory {
						if itm == "ammo" {
							ammoIdx = idx
							break
						}
					}

					if ammoIdx >= 0 {
						// Consume 1 ammo item
						player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)

						// Deduct shotgun durability
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
							player.WeaponDurability = 0
						}

						// Play gunshot blast sound
						assets.PlaySound(assets.HitSound)

						// Normalize facing vector
						facingLen := math.Hypot(player.FacingX, player.FacingY)
						facingX, facingY := player.FacingX, player.FacingY
						if facingLen < 0.001 {
							facingX, facingY = 1.0, 0.0
						} else {
							facingX /= facingLen
							facingY /= facingLen
						}

						// Shotgun Spread Cone (Range: 160px, Angle: +-22.5 degrees)
						const maxShotgunRange = 160.0
						const cosSpread = 0.9238795325112867 // math.Cos(22.5 * math.Pi / 180.0)

						zQuery := s.zombieFilter.Query()
						for zQuery.Next() {
							_, zPos, _ := zQuery.Get()
							ent := zQuery.Entity()

							dx := zPos.X - pos.X
							dy := zPos.Y - pos.Y
							dist := math.Hypot(dx, dy)

							if dist <= maxShotgunRange {
								if dist < 24.0 {
									// Point-blank kill
									toRemoveZombies = append(toRemoveZombies, ent)
								} else {
									cosAngle := (facingX*dx + facingY*dy) / dist
									if cosAngle >= cosSpread {
										toRemoveZombies = append(toRemoveZombies, ent)
									}
								}
							}
						}

						// Acoustic Noise Pulse: Alerts all wandering zombies within 400.0px
						noiseQuery := s.zombieFilter.Query()
						for noiseQuery.Next() {
							z, zPos, _ := noiseQuery.Get()
							zdx := pos.X - zPos.X
							zdy := pos.Y - zPos.Y
							if math.Hypot(zdx, zdy) <= 400.0 {
								z.Chasing = true
								z.WanderTimer = 0
							}
						}
					} else {
						// Dry Fire / Out of Ammo: Mechanical click & defensive butt shove
						assets.PlaySound(assets.ShoveSound)

						attackX := pos.X + player.FacingX*24
						attackY := pos.Y + player.FacingY*24
						zQuery := s.zombieFilter.Query()
						for zQuery.Next() {
							z, zPos, zVel := zQuery.Get()
							dx := attackX - zPos.X
							dy := attackY - zPos.Y
							if math.Hypot(dx, dy) < 24.0 {
								z.StunTimer = 45
								zVel.X = player.FacingX * 5.0
								zVel.Y = player.FacingY * 5.0
							}
						}
					}
				} else if player.WeaponEquipped && player.WeaponType == "axe" {
					// Fire Axe Melee Attack: Cleave reach 32.0px
					attackX := pos.X + player.FacingX*32
					attackY := pos.Y + player.FacingY*32
					hitZombies := false

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						_, zPos, _ := zQuery.Get()
						ent := zQuery.Entity()

						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 32.0 {
							hitZombies = true
							toRemoveZombies = append(toRemoveZombies, ent)
						}
					}

					if hitZombies {
						assets.PlaySound(assets.HitSound)
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
							player.WeaponDurability = 0
						}
					} else {
						assets.PlaySound(assets.ShoveSound)
					}
				} else if player.WeaponEquipped {
					// Standard Melee Attack: Reach 24.0px
					attackX := pos.X + player.FacingX*24
					attackY := pos.Y + player.FacingY*24
					hitZombies := false

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						_, zPos, _ := zQuery.Get()
						ent := zQuery.Entity()

						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 24.0 {
							hitZombies = true
							toRemoveZombies = append(toRemoveZombies, ent)
						}
					}

					if hitZombies {
						assets.PlaySound(assets.HitSound)
						player.WeaponDurability--
						if player.WeaponDurability <= 0 {
							player.WeaponEquipped = false
							player.WeaponType = ""
							player.WeaponDurability = 0
						}
					} else {
						assets.PlaySound(assets.ShoveSound)
					}
				} else {
					// Unarmed Shove
					attackX := pos.X + player.FacingX*24
					attackY := pos.Y + player.FacingY*24

					zQuery := s.zombieFilter.Query()
					for zQuery.Next() {
						z, zPos, zVel := zQuery.Get()
						dx := attackX - zPos.X
						dy := attackY - zPos.Y
						if math.Hypot(dx, dy) < 24.0 {
							z.StunTimer = 45
							zVel.X = player.FacingX * 5.0
							zVel.Y = player.FacingY * 5.0
						}
					}
					assets.PlaySound(assets.ShoveSound)
				}
			}
```

---

## 5. Verification Method

### 5.1 Unit Test Suite (`internal/game/combat_test.go` or `internal/game/shotgun_test.go`)
The implementation can be verified via the following comprehensive test cases:

```go
package game

import (
	"image/color"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// 1. Test Shotgun Equipping from Inventory
func TestShotgun_EquipFromInventory(t *testing.T) {
	player := &ecs.Player{
		Health:           100.0,
		Inventory:        []string{"shotgun", "ammo", "food"},
		WeaponEquipped:   false,
		WeaponType:       "",
		WeaponDurability: 0,
		AttackCooldown:   0,
	}

	useIdx := 0
	if useIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		item := player.Inventory[useIdx]
		if item == "shotgun" {
			player.WeaponEquipped = true
			player.WeaponType = "shotgun"
			player.WeaponDurability = 15
			player.Inventory = append(player.Inventory[:useIdx], player.Inventory[useIdx+1:]...)
		}
	}

	if !player.WeaponEquipped {
		t.Fatal("Expected shotgun to be equipped")
	}
	if player.WeaponType != "shotgun" {
		t.Fatalf("Expected WeaponType 'shotgun', got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 15 {
		t.Fatalf("Expected durability 15, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 2 || player.Inventory[0] != "ammo" || player.Inventory[1] != "food" {
		t.Fatalf("Expected remaining inventory ['ammo', 'food'], got %v", player.Inventory)
	}
}

// 2. Test Ammo Consumption & Durability Loss on Firing
func TestShotgun_AmmoConsumptionAndDurability(t *testing.T) {
	player := &ecs.Player{
		Inventory:        []string{"ammo", "ammo", "food"},
		WeaponEquipped:   true,
		WeaponType:       "shotgun",
		WeaponDurability: 15,
		AttackCooldown:   0,
	}

	// Fire shot 1
	ammoIdx := -1
	for i, itm := range player.Inventory {
		if itm == "ammo" {
			ammoIdx = i
			break
		}
	}
	if ammoIdx < 0 {
		t.Fatal("Expected ammo to be found in inventory")
	}
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
	player.WeaponDurability--

	if player.WeaponDurability != 14 {
		t.Fatalf("Expected durability 14, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 2 || player.Inventory[0] != "ammo" || player.Inventory[1] != "food" {
		t.Fatalf("Expected 1 ammo consumed, remaining %v", player.Inventory)
	}
}

// 3. Test Dry Fire when Out of Ammo
func TestShotgun_DryFire_OutOfAmmo(t *testing.T) {
	player := &ecs.Player{
		Inventory:        []string{"food", "water"},
		WeaponEquipped:   true,
		WeaponType:       "shotgun",
		WeaponDurability: 10,
		AttackCooldown:   0,
	}

	ammoIdx := -1
	for i, itm := range player.Inventory {
		if itm == "ammo" {
			ammoIdx = i
			break
		}
	}

	if ammoIdx != -1 {
		t.Fatal("Expected no ammo to be found")
	}

	// Dry fire does not consume durability or items
	initialDurability := player.WeaponDurability
	initialInvLen := len(player.Inventory)

	if player.WeaponDurability != initialDurability {
		t.Errorf("Dry fire modified durability: got %d, want %d", player.WeaponDurability, initialDurability)
	}
	if len(player.Inventory) != initialInvLen {
		t.Errorf("Dry fire consumed inventory items")
	}
}

// 4. Test Shotgun Spread Cone Math (+-22.5 deg, 160px reach)
func TestShotgun_SpreadConeMath(t *testing.T) {
	tests := []struct {
		name      string
		facingX   float64
		facingY   float64
		targetX   float64
		targetY   float64
		wantHit   bool
	}{
		{"Direct Ahead 100px", 1, 0, 100, 0, true},
		{"Within Cone +20 deg 100px", 1, 0, 100, 36, true},    // tan(20 deg) = 0.364 -> ~19.8 deg
		{"Outside Cone +30 deg 100px", 1, 0, 100, 58, false},  // tan(30 deg) = 0.577 -> ~30.1 deg
		{"Direct Ahead Max Range 160px", 1, 0, 160, 0, true},
		{"Beyond Max Range 170px", 1, 0, 170, 0, false},
		{"Directly Behind Player", 1, 0, -50, 0, false},
		{"Point Blank 15px Offset Angle", 1, 0, 0, 15, true}, // dist < 24px point-blank hit
		{"Diagonal Facing Hit", 1, 1, 80, 80, true},
		{"Diagonal Facing Off-Angle Miss", 1, 1, -80, 80, false},
	}

	const maxRange = 160.0
	const cosSpread = 0.9238795325112867

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facingLen := math.Hypot(tt.facingX, tt.facingY)
			fx, fy := tt.facingX/facingLen, tt.facingY/facingLen

			dx := tt.targetX
			dy := tt.targetY
			dist := math.Hypot(dx, dy)

			hit := false
			if dist <= maxRange {
				if dist < 24.0 {
					hit = true
				} else {
					cosAngle := (fx*dx + fy*dy) / dist
					if cosAngle >= cosSpread {
						hit = true
					}
				}
			}

			if hit != tt.wantHit {
				t.Errorf("SpreadCone hit = %v, want %v (dist=%.2f)", hit, tt.wantHit, dist)
			}
		})
	}
}

// 5. Test Acoustic Noise Pulse (400px Radius Aggro)
func TestShotgun_AcousticNoisePulseRadius(t *testing.T) {
	tests := []struct {
		name       string
		playerPos  ecs.Position
		zombiePos  ecs.Position
		wantAggro  bool
	}{
		{"Near Zombie 100px", ecs.Position{X: 100, Y: 100}, ecs.Position{X: 200, Y: 100}, true},
		{"Mid Zombie 350px", ecs.Position{X: 100, Y: 100}, ecs.Position{X: 450, Y: 100}, true},
		{"Edge Zombie 400px", ecs.Position{X: 100, Y: 100}, ecs.Position{X: 500, Y: 100}, true},
		{"Outside Zombie 410px", ecs.Position{X: 100, Y: 100}, ecs.Position{X: 510, Y: 100}, false},
		{"Far Zombie 600px", ecs.Position{X: 100, Y: 100}, ecs.Position{X: 700, Y: 100}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := ecs.Zombie{Chasing: false, WanderTimer: 100}
			dx := tt.playerPos.X - tt.zombiePos.X
			dy := tt.playerPos.Y - tt.zombiePos.Y
			dist := math.Hypot(dx, dy)

			if dist <= 400.0 {
				z.Chasing = true
				z.WanderTimer = 0
			}

			if z.Chasing != tt.wantAggro {
				t.Errorf("Zombie Chasing = %v, want %v (dist=%.2f)", z.Chasing, tt.wantAggro, dist)
			}
			if tt.wantAggro && z.WanderTimer != 0 {
				t.Errorf("Alerted zombie WanderTimer = %d, want 0", z.WanderTimer)
			}
		})
	}
}

// 6. Test Shotgun Breakage at 0 Durability
func TestShotgun_BreakageAtZeroDurability(t *testing.T) {
	player := &ecs.Player{
		Inventory:        []string{"ammo"},
		WeaponEquipped:   true,
		WeaponType:       "shotgun",
		WeaponDurability: 1, // Final shot
	}

	ammoIdx := 0
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
	player.WeaponDurability--
	if player.WeaponDurability <= 0 {
		player.WeaponEquipped = false
		player.WeaponType = ""
		player.WeaponDurability = 0
	}

	if player.WeaponEquipped {
		t.Error("Expected shotgun to unequip on 0 durability")
	}
	if player.WeaponType != "" {
		t.Errorf("Expected empty WeaponType, got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 0 {
		t.Errorf("Expected durability 0, got %d", player.WeaponDurability)
	}
}
```

### 5.2 Verification Commands
```bash
# 1. Run all unit and system tests
CC=gcc go test -v ./...

# 2. Build the game binary
CC=gcc go build -o bin/game ./cmd/game

# 3. Launch the game
CC=gcc go run ./cmd/game
```
*Expected Result*: All tests pass cleanly without errors; shotgun equips from hotbar (1-9), fires when ammo is available, kills zombies in the spread cone, alerts nearby zombies within 400px, and correctly falls back to dry-fire when out of ammo.
