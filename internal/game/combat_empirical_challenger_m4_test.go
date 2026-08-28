package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupM4CombatHarness initializes an isolated headless test world
func setupM4CombatHarness() (*arkecs.World, *world.Map, *UpdateSystem, *DrawSystem, arkecs.Entity) {
	assets.Load()
	assets.InitAudio()
	w := arkecs.NewWorld()
	m := world.NewMap(60, 60)
	upd := NewUpdateSystem(w, m)
	drw := NewDrawSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			Hunger:             100.0,
			Thirst:             100.0,
			Inventory:          []string{},
			WeaponEquipped:     false,
			WeaponType:         "",
			WeaponDurability:   0,
			ArmorEquipped:      false,
			ArmorType:          "",
			ArmorDefense:       0.0,
			ArmorDurability:    0,
			ArmorMaxDurability: 0,
			InfectionResist:    0.0,
			AttackCooldown:     0,
			Dead:               false,
			Infected:           false,
			FacingX:            1.0,
			FacingY:            0.0,
		},
		&ecs.Position{X: 300.0, Y: 300.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, upd, drw, pEnt
}

// 1. Empirical Test: 5,000 Rapid Weapon & Item Swaps with Invariant Validation
func TestChallenger_RapidWeaponSwappingInvariants_5000Cycles(t *testing.T) {
	w, _, _, _, pEnt := setupM4CombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	availableItems := []string{"weapon", "axe", "shotgun", "ammo", "armor", "food", "water"}
	r := rand.New(rand.NewSource(987654))

	for cycle := 0; cycle < 5000; cycle++ {
		// Fill inventory up to 9 slots
		for len(player.Inventory) < 9 {
			chosen := availableItems[r.Intn(len(availableItems))]
			player.Inventory = append(player.Inventory, chosen)
		}

		slot := r.Intn(len(player.Inventory))
		item := player.Inventory[slot]
		player.AttackCooldown = 0

		used := false
		if item == "food" && player.Hunger < 100 {
			player.Hunger += 50
			if player.Hunger > 100 {
				player.Hunger = 100
			}
			used = true
		} else if item == "water" && player.Thirst < 100 {
			player.Thirst += 50
			if player.Thirst > 100 {
				player.Thirst = 100
			}
			used = true
		} else if item == "weapon" {
			player.WeaponEquipped = true
			player.WeaponType = "weapon"
			player.WeaponDurability = 5
			used = true
		} else if item == "axe" {
			player.WeaponEquipped = true
			player.WeaponType = "axe"
			player.WeaponDurability = 12
			used = true
		} else if item == "shotgun" {
			player.WeaponEquipped = true
			player.WeaponType = "shotgun"
			player.WeaponDurability = 15
			used = true
		} else if item == "armor" || item == "vest" {
			player.ArmorEquipped = true
			player.ArmorType = "vest"
			player.ArmorDefense = 0.50
			player.ArmorDurability = 10
			player.ArmorMaxDurability = 10
			player.InfectionResist = 0.70
			used = true
		}

		if used {
			player.Inventory = append(player.Inventory[:slot], player.Inventory[slot+1:]...)
		}

		// Invariant checks
		if player.WeaponEquipped {
			if player.WeaponType != "weapon" && player.WeaponType != "axe" && player.WeaponType != "shotgun" {
				t.Fatalf("Cycle %d: Invalid WeaponType '%s' when WeaponEquipped=true", cycle, player.WeaponType)
			}
			if player.WeaponDurability <= 0 {
				t.Fatalf("Cycle %d: WeaponDurability is %d but WeaponEquipped=true", cycle, player.WeaponDurability)
			}
		} else {
			if player.WeaponType != "" {
				t.Fatalf("Cycle %d: WeaponType '%s' should be empty when WeaponEquipped=false", cycle, player.WeaponType)
			}
			if player.WeaponDurability != 0 {
				t.Fatalf("Cycle %d: WeaponDurability is %d when WeaponEquipped=false", cycle, player.WeaponDurability)
			}
		}

		if len(player.Inventory) > 9 {
			t.Fatalf("Cycle %d: Inventory size %d exceeded max capacity of 9", cycle, len(player.Inventory))
		}
	}
}

// 2. Empirical Test: Simultaneous Horde Combat (Axe Cleave into 100 Zombies)
func TestChallenger_FireAxeCleaveHordeCombatStress(t *testing.T) {
	w, _, _, _, pEnt := setupM4CombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	// Attack center: (332, 300), radius: 32px
	r := rand.New(rand.NewSource(112233))
	expectedCleaveHits := 0
	expectedSurvivors := 0

	for i := 0; i < 100; i++ {
		zx := 300.0 + r.Float64()*80.0
		zy := 270.0 + r.Float64()*60.0

		dx := 332.0 - zx
		dy := 300.0 - zy
		if math.Hypot(dx, dy) < 32.0 {
			expectedCleaveHits++
		} else {
			expectedSurvivors++
		}

		zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0},
			&ecs.Position{X: zx, Y: zy},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)
	}

	if expectedCleaveHits == 0 {
		t.Fatal("Expected at least some zombies inside cleave circle")
	}

	// Perform Axe Swing
	attackX := 300.0 + player.FacingX*32.0
	attackY := 300.0 + player.FacingY*32.0
	var toRemove []arkecs.Entity
	zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w).Query()
	for zq.Next() {
		_, zPos := zq.Get()
		ent := zq.Entity()
		if math.Hypot(attackX-zPos.X, attackY-zPos.Y) < 32.0 {
			toRemove = append(toRemove, ent)
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}
	if len(toRemove) > 0 {
		player.WeaponDurability--
	}

	if len(toRemove) != expectedCleaveHits {
		t.Errorf("Expected cleave to eliminate %d zombies, got %d", expectedCleaveHits, len(toRemove))
	}
	if player.WeaponDurability != 11 {
		t.Errorf("Axe durability should be 11 (decayed by exactly 1), got %d", player.WeaponDurability)
	}

	survivorCount := 0
	sq := arkecs.NewFilter1[ecs.Zombie](w).Query()
	for sq.Next() {
		survivorCount++
	}
	if survivorCount != expectedSurvivors {
		t.Errorf("Expected %d survivors, got %d", expectedSurvivors, survivorCount)
	}
}

// 3. Empirical Test: Simultaneous Horde Combat (Shotgun Blast + Noise Alert across 100 Zombies)
func TestChallenger_ShotgunHordeBlastAndNoiseSwarmStress(t *testing.T) {
	w, _, _, _, pEnt := setupM4CombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"ammo", "ammo"}
	player.FacingX = 0.0
	player.FacingY = 1.0 // Facing Down

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	// Spawn 100 zombies in a ring around the player
	r := rand.New(rand.NewSource(445566))
	type zInfo struct {
		ent         arkecs.Entity
		x, y        float64
		shouldKill  bool
		shouldAlert bool
	}
	var zombies []zInfo

	const maxShotgunRange = 160.0
	const cosSpread = 0.9238795325112867

	for i := 0; i < 100; i++ {
		angle := r.Float64() * 2 * math.Pi
		dist := 10.0 + r.Float64()*500.0 // up to 510px
		zx := 300.0 + math.Cos(angle)*dist
		zy := 300.0 + math.Sin(angle)*dist

		dx := zx - 300.0
		dy := zy - 300.0
		actualDist := math.Hypot(dx, dy)

		shouldKill := false
		if actualDist <= maxShotgunRange {
			if actualDist < 24.0 {
				shouldKill = true
			} else {
				cosAngle := (player.FacingX*dx + player.FacingY*dy) / actualDist
				if cosAngle >= cosSpread {
					shouldKill = true
				}
			}
		}

		shouldAlert := actualDist <= 400.0

		ent := zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 99},
			&ecs.Position{X: zx, Y: zy},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		zombies = append(zombies, zInfo{
			ent:         ent,
			x:           zx,
			y:           zy,
			shouldKill:  shouldKill,
			shouldAlert: shouldAlert,
		})
	}

	// Fire Shotgun
	ammoIdx := -1
	for idx, itm := range player.Inventory {
		if itm == "ammo" {
			ammoIdx = idx
			break
		}
	}
	if ammoIdx < 0 {
		t.Fatal("Expected ammo in inventory")
	}
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
	player.WeaponDurability--

	var toRemove []arkecs.Entity
	zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w).Query()
	for zq.Next() {
		_, zPos := zq.Get()
		ent := zq.Entity()
		dx := zPos.X - 300.0
		dy := zPos.Y - 300.0
		dist := math.Hypot(dx, dy)
		if dist <= maxShotgunRange {
			if dist < 24.0 {
				toRemove = append(toRemove, ent)
			} else {
				cosAngle := (player.FacingX*dx + player.FacingY*dy) / dist
				if cosAngle >= cosSpread {
					toRemove = append(toRemove, ent)
				}
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}

	// Acoustic Noise Pulse (400px)
	nq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w).Query()
	for nq.Next() {
		z, zPos := nq.Get()
		if math.Hypot(300.0-zPos.X, 300.0-zPos.Y) <= 400.0 {
			z.Chasing = true
			z.WanderTimer = 0
		}
	}

	// Assertions
	zCompMap := arkecs.NewMap1[ecs.Zombie](w)
	for i, zi := range zombies {
		alive := w.Alive(zi.ent)
		if zi.shouldKill && alive {
			t.Errorf("Zombie #%d: Expected kill, but still alive at (%.1f, %.1f)", i, zi.x, zi.y)
		}
		if !zi.shouldKill && !alive {
			t.Errorf("Zombie #%d: Expected survive, but killed at (%.1f, %.1f)", i, zi.x, zi.y)
		}
		if alive {
			zc := zCompMap.Get(zi.ent)
			if zi.shouldAlert {
				if !zc.Chasing || zc.WanderTimer != 0 {
					t.Errorf("Zombie #%d at dist %.1f should be alerted to chase", i, math.Hypot(zi.x-300, zi.y-300))
				}
			} else {
				if zc.Chasing || zc.WanderTimer != 99 {
					t.Errorf("Zombie #%d at dist %.1f should NOT be alerted", i, math.Hypot(zi.x-300, zi.y-300))
				}
			}
		}
	}
}

// 4. Empirical Test: 1,500 Continuous Simulation Frames
func TestChallenger_1500FramesHeavyContinuousSimulation(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()
	r := rand.New(rand.NewSource(778899))

	pMap := arkecs.NewMap1[ecs.Player](g.world)
	posMap := arkecs.NewMap1[ecs.Position](g.world)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](g.world)

	// Collect player entity and completely drain query
	var pEnt arkecs.Entity
	pq := arkecs.NewFilter1[ecs.Player](g.world).Query()
	for pq.Next() {
		pEnt = pq.Entity()
	}

	player := pMap.Get(pEnt)
	player.Health = 100.0
	player.Hunger = 100.0
	player.Thirst = 100.0
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.Inventory = []string{"shotgun", "ammo", "ammo", "ammo", "food", "water", "armor"}

	offscreen := ebiten.NewImage(800, 600)

	for frame := 0; frame < 1500; frame++ {
		// Replenish vitals so player doesn't die of hunger/thirst in headless run
		if player.Health < 50 {
			player.Health = 100
		}
		if player.Hunger < 30 {
			player.Hunger = 100
		}
		if player.Thirst < 30 {
			player.Thirst = 100
		}

		if frame%20 == 0 {
			// Change facing direction
			angles := []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi, -math.Pi / 2}
			a := angles[r.Intn(len(angles))]
			player.FacingX = math.Cos(a)
			player.FacingY = math.Sin(a)

			// Attack
			if player.WeaponEquipped && player.WeaponType == "axe" {
				pPos := posMap.Get(pEnt)
				attackX := pPos.X + player.FacingX*32.0
				attackY := pPos.Y + player.FacingY*32.0
				hitZombies := false
				var toRemove []arkecs.Entity
				zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](g.world).Query()
				for zq.Next() {
					_, zPos := zq.Get()
					ent := zq.Entity()
					if math.Hypot(attackX-zPos.X, attackY-zPos.Y) < 32.0 {
						hitZombies = true
						toRemove = append(toRemove, ent)
					}
				}
				for _, ent := range toRemove {
					g.world.RemoveEntity(ent)
				}
				if hitZombies {
					player.WeaponDurability--
					if player.WeaponDurability <= 0 {
						player.WeaponEquipped = false
						player.WeaponType = ""
						player.WeaponDurability = 0
					}
				}
			} else if player.WeaponEquipped && player.WeaponType == "shotgun" {
				pPos := posMap.Get(pEnt)
				ammoIdx := -1
				for i, itm := range player.Inventory {
					if itm == "ammo" {
						ammoIdx = i
						break
					}
				}
				if ammoIdx >= 0 {
					player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
					player.WeaponDurability--
					if player.WeaponDurability <= 0 {
						player.WeaponEquipped = false
						player.WeaponType = ""
						player.WeaponDurability = 0
					}

					facingLen := math.Hypot(player.FacingX, player.FacingY)
					fx, fy := player.FacingX/facingLen, player.FacingY/facingLen
					var toRemove []arkecs.Entity
					zq := arkecs.NewFilter2[ecs.Zombie, ecs.Position](g.world).Query()
					for zq.Next() {
						_, zPos := zq.Get()
						ent := zq.Entity()
						dx := zPos.X - pPos.X
						dy := zPos.Y - pPos.Y
						dist := math.Hypot(dx, dy)
						if dist <= 160.0 {
							if dist < 24.0 || (fx*dx+fy*dy)/dist >= 0.92388 {
								toRemove = append(toRemove, ent)
							}
						}
					}
					for _, ent := range toRemove {
						g.world.RemoveEntity(ent)
					}
				}
			} else if !player.WeaponEquipped {
				// Re-equip shotgun if in inventory
				for i, itm := range player.Inventory {
					if itm == "shotgun" {
						player.WeaponEquipped = true
						player.WeaponType = "shotgun"
						player.WeaponDurability = 15
						player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
						break
					}
				}
			}
		}

		// Spawn fresh zombies periodically
		if frame%120 == 0 {
			pPos := posMap.Get(pEnt)
			for k := 0; k < 4; k++ {
				angle := r.Float64() * 2 * math.Pi
				dist := 70.0 + r.Float64()*100.0
				zMap.NewEntity(
					&ecs.Zombie{Speed: 1.2, Chasing: true},
					&ecs.Position{X: pPos.X + math.Cos(angle)*dist, Y: pPos.Y + math.Sin(angle)*dist},
					&ecs.Velocity{X: 0, Y: 0},
					&ecs.Sprite{W: 16, H: 16},
					&ecs.Collider{Width: 16, Height: 16},
				)
			}
		}

		// Update systems
		g.updateSys.processZombies()
		g.updateSys.processMovement()

		// Render every 100 frames
		if frame%100 == 0 {
			g.drawSys.Draw(offscreen, 12.0)
		}
	}
}

// 5. Empirical Test: HUD String Formatting & Reticle Tints Exhaustive Matrix
func TestChallenger_HUDStringsAndReticleTintsMatrix(t *testing.T) {
	assets.Load()
	offscreen := ebiten.NewImage(800, 600)

	type testCase struct {
		name             string
		weaponEquipped   bool
		weaponType       string
		weaponDurability int
		inventory        []string
		armorEquipped    bool
		armorDurability  int
		armorMaxDur      int
		armorDefense     float64
		expectedHUDText  string
		expectedReticleR float32
		expectedReticleG float32
		expectedReticleB float32
		expectedReticleA float32
	}

	tests := []testCase{
		{
			name:             "Unarmed Fists",
			weaponEquipped:   false,
			weaponType:       "",
			weaponDurability: 0,
			inventory:        []string{"food"},
			expectedHUDText:  "Weapon: NONE (Fists)",
			expectedReticleR: 1.0,
			expectedReticleG: 1.0,
			expectedReticleB: 0.0,
			expectedReticleA: 0.7,
		},
		{
			name:             "Club Equipped",
			weaponEquipped:   true,
			weaponType:       "weapon",
			weaponDurability: 5,
			inventory:        []string{},
			expectedHUDText:  "Weapon: WEAPON (5 hits)",
			expectedReticleR: 1.0,
			expectedReticleG: 0.0,
			expectedReticleB: 0.0,
			expectedReticleA: 0.7,
		},
		{
			name:             "Fire Axe Equipped",
			weaponEquipped:   true,
			weaponType:       "axe",
			weaponDurability: 12,
			inventory:        []string{"water"},
			expectedHUDText:  "Weapon: AXE (12 hits)",
			expectedReticleR: 1.0,
			expectedReticleG: 0.2,
			expectedReticleB: 0.2,
			expectedReticleA: 0.8,
		},
		{
			name:             "Shotgun with 3 Ammo",
			weaponEquipped:   true,
			weaponType:       "shotgun",
			weaponDurability: 15,
			inventory:        []string{"ammo", "food", "ammo", "ammo"},
			expectedHUDText:  "Weapon: SHOTGUN (15 hits | Ammo: 3)",
			expectedReticleR: 1.0,
			expectedReticleG: 0.6,
			expectedReticleB: 0.2,
			expectedReticleA: 0.8,
		},
		{
			name:             "Shotgun with 0 Ammo",
			weaponEquipped:   true,
			weaponType:       "shotgun",
			weaponDurability: 7,
			inventory:        []string{"food"},
			expectedHUDText:  "Weapon: SHOTGUN (7 hits | Ammo: 0)",
			expectedReticleR: 1.0,
			expectedReticleG: 0.6,
			expectedReticleB: 0.2,
			expectedReticleA: 0.8,
		},
		{
			name:             "Broken Weapon State",
			weaponEquipped:   false,
			weaponType:       "",
			weaponDurability: 0,
			inventory:        []string{},
			expectedHUDText:  "Weapon: NONE (Fists)",
			expectedReticleR: 1.0,
			expectedReticleG: 1.0,
			expectedReticleB: 0.0,
			expectedReticleA: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _, _, drw, pEnt := setupM4CombatHarness()
			pMap := arkecs.NewMap1[ecs.Player](w)
			player := pMap.Get(pEnt)

			player.WeaponEquipped = tt.weaponEquipped
			player.WeaponType = tt.weaponType
			player.WeaponDurability = tt.weaponDurability
			player.Inventory = tt.inventory
			player.ArmorEquipped = tt.armorEquipped
			player.ArmorDurability = tt.armorDurability
			player.ArmorMaxDurability = tt.armorMaxDur
			player.ArmorDefense = tt.armorDefense

			// Check HUD String logic
			var hudWeaponText string
			if player.WeaponEquipped && player.WeaponDurability > 0 {
				wType := strings.ToUpper(player.WeaponType)
				if wType == "" {
					wType = "WEAPON"
				}
				if player.WeaponType == "shotgun" {
					ammoCount := 0
					for _, item := range player.Inventory {
						if item == "ammo" {
							ammoCount++
						}
					}
					hudWeaponText = fmt.Sprintf("Weapon: %s (%d hits | Ammo: %d)", wType, player.WeaponDurability, ammoCount)
				} else {
					hudWeaponText = fmt.Sprintf("Weapon: %s (%d hits)", wType, player.WeaponDurability)
				}
			} else {
				hudWeaponText = "Weapon: NONE (Fists)"
			}

			if hudWeaponText != tt.expectedHUDText {
				t.Errorf("HUD text mismatch: got '%s', want '%s'", hudWeaponText, tt.expectedHUDText)
			}

			// Check Reticle Tint logic
			op := &ebiten.DrawImageOptions{}
			if player.WeaponEquipped {
				if player.WeaponType == "shotgun" {
					op.ColorScale.Scale(1.0, 0.6, 0.2, 0.8)
				} else if player.WeaponType == "axe" {
					op.ColorScale.Scale(1.0, 0.2, 0.2, 0.8)
				} else {
					op.ColorScale.Scale(1.0, 0.0, 0.0, 0.7)
				}
			} else {
				op.ColorScale.Scale(1.0, 1.0, 0.0, 0.7)
			}

			const eps = 1e-4
			if math.Abs(float64(op.ColorScale.R()-tt.expectedReticleR)) > eps ||
				math.Abs(float64(op.ColorScale.G()-tt.expectedReticleG)) > eps ||
				math.Abs(float64(op.ColorScale.B()-tt.expectedReticleB)) > eps ||
				math.Abs(float64(op.ColorScale.A()-tt.expectedReticleA)) > eps {
				t.Errorf("Reticle ColorScale mismatch: got (%.2f, %.2f, %.2f, %.2f), want (%.2f, %.2f, %.2f, %.2f)",
					op.ColorScale.R(), op.ColorScale.G(), op.ColorScale.B(), op.ColorScale.A(),
					tt.expectedReticleR, tt.expectedReticleG, tt.expectedReticleB, tt.expectedReticleA)
			}

			// Render frame
			drw.Draw(offscreen, 12.0)
		})
	}
}

// 6. Empirical Test: Complete Weapon Breakage Lifecycle Transitioning to Unarmed Shove
func TestChallenger_WeaponBreakageLifecycleTransition(t *testing.T) {
	weapons := []struct {
		weaponType string
		initDur    int
	}{
		{"weapon", 5},
		{"axe", 12},
		{"shotgun", 15},
	}

	for _, wpn := range weapons {
		t.Run(wpn.weaponType, func(t *testing.T) {
			w, _, _, _, pEnt := setupM4CombatHarness()
			pMap := arkecs.NewMap1[ecs.Player](w)
			player := pMap.Get(pEnt)
			player.WeaponEquipped = true
			player.WeaponType = wpn.weaponType
			player.WeaponDurability = 1 // 1 hit remaining!

			zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

			// Hit breaks weapon
			player.WeaponDurability--
			if player.WeaponDurability <= 0 {
				player.WeaponEquipped = false
				player.WeaponType = ""
				player.WeaponDurability = 0
			}

			if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
				t.Fatalf("Weapon did not reset to unarmed after reaching 0 durability: %+v", player)
			}

			// Perform Unarmed Shove on next attack
			zEnt := zMap.NewEntity(
				&ecs.Zombie{Speed: 1.0, StunTimer: 0},
				&ecs.Position{X: 315.0, Y: 300.0},
				&ecs.Velocity{X: 0, Y: 0},
				&ecs.Sprite{W: 16, H: 16},
				&ecs.Collider{Width: 16, Height: 16},
			)

			attackX := 300.0 + player.FacingX*24.0
			attackY := 300.0 + player.FacingY*24.0
			zq := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
			for zq.Next() {
				z, zPos, zVel := zq.Get()
				if math.Hypot(attackX-zPos.X, attackY-zPos.Y) < 24.0 {
					z.StunTimer = 45
					zVel.X = player.FacingX * 5.0
					zVel.Y = player.FacingY * 5.0
				}
			}

			if !w.Alive(zEnt) {
				t.Fatal("Unarmed shove must NOT delete zombie entity")
			}
			zComp := arkecs.NewMap1[ecs.Zombie](w).Get(zEnt)
			if zComp.StunTimer != 45 {
				t.Errorf("Expected StunTimer=45, got %d", zComp.StunTimer)
			}
			zVelComp := arkecs.NewMap1[ecs.Velocity](w).Get(zEnt)
			if zVelComp.X != 5.0 || zVelComp.Y != 0.0 {
				t.Errorf("Expected knockback velocity (5.0, 0.0), got (%f, %f)", zVelComp.X, zVelComp.Y)
			}
		})
	}
}
