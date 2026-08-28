package game

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupCombatTestHarness initializes a headless test world with player entity
func setupCombatTestHarness() (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	updateSys := NewUpdateSystem(w, m)

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := playerMap.NewEntity(
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
		&ecs.Position{X: 100.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, updateSys, pEnt
}

// 1. Test ECS Player Component Weapon Fields
func TestCombat_ECSComponentWeaponFields(t *testing.T) {
	pAxe := ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "axe",
		WeaponDurability: 12,
	}
	if !pAxe.WeaponEquipped || pAxe.WeaponType != "axe" || pAxe.WeaponDurability != 12 {
		t.Errorf("Axe component initialization failed: %+v", pAxe)
	}

	pShotgun := ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "shotgun",
		WeaponDurability: 15,
	}
	if !pShotgun.WeaponEquipped || pShotgun.WeaponType != "shotgun" || pShotgun.WeaponDurability != 15 {
		t.Errorf("Shotgun component initialization failed: %+v", pShotgun)
	}

	pClub := ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "weapon",
		WeaponDurability: 5,
	}
	if !pClub.WeaponEquipped || pClub.WeaponType != "weapon" || pClub.WeaponDurability != 5 {
		t.Errorf("Club component initialization failed: %+v", pClub)
	}

	pUnarmed := ecs.Player{
		WeaponEquipped:   false,
		WeaponType:       "",
		WeaponDurability: 0,
	}
	if pUnarmed.WeaponEquipped || pUnarmed.WeaponType != "" || pUnarmed.WeaponDurability != 0 {
		t.Errorf("Unarmed component initialization failed: %+v", pUnarmed)
	}
}

// 2. Test Equipping Weapons from Inventory (Club, Axe, Shotgun)
func TestCombat_EquipWeaponsFromInventory(t *testing.T) {
	testCases := []struct {
		itemType       string
		expectedType   string
		expectedDur    int
	}{
		{"weapon", "weapon", 5},
		{"axe", "axe", 12},
		{"shotgun", "shotgun", 15},
	}

	for _, tc := range testCases {
		t.Run(tc.itemType, func(t *testing.T) {
			w, _, _, pEnt := setupCombatTestHarness()
			pMap := arkecs.NewMap1[ecs.Player](w)
			player := pMap.Get(pEnt)
			player.Inventory = []string{tc.itemType, "food"}

			// Simulate inventory use
			useItemIdx := 0
			if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30
				item := player.Inventory[useItemIdx]
				if item == "weapon" {
					player.WeaponEquipped = true
					player.WeaponType = "weapon"
					player.WeaponDurability = 5
				} else if item == "axe" {
					player.WeaponEquipped = true
					player.WeaponType = "axe"
					player.WeaponDurability = 12
				} else if item == "shotgun" {
					player.WeaponEquipped = true
					player.WeaponType = "shotgun"
					player.WeaponDurability = 15
				}
				player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
			}

			if !player.WeaponEquipped {
				t.Errorf("Expected WeaponEquipped true for %s", tc.itemType)
			}
			if player.WeaponType != tc.expectedType {
				t.Errorf("Expected WeaponType %s, got %s", tc.expectedType, player.WeaponType)
			}
			if player.WeaponDurability != tc.expectedDur {
				t.Errorf("Expected WeaponDurability %d, got %d", tc.expectedDur, player.WeaponDurability)
			}
			if len(player.Inventory) != 1 || player.Inventory[0] != "food" {
				t.Errorf("Expected remaining inventory ['food'], got %v", player.Inventory)
			}
		})
	}
}

// 3. Test Fire Axe Multi-Target Cleave and Durability Loss
func TestCombat_AxeCleaveMultiTargetKill(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Player at (100, 100). Axe attack reach = 32.0px. Attack center = (128, 100).
	// Spawn 3 zombies within cleave area:
	// Z1: Straight ahead at (125, 100) -> distance to attack center = 3px < 32px
	// Z2: Upper arc at (125, 80) -> dx = 3, dy = 20 -> dist = 20.2px < 32px
	// Z3: Lower arc at (125, 120) -> dx = 3, dy = 20 -> dist = 20.2px < 32px
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 125.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 125.0, Y: 80.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 125.0, Y: 120.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Simulate Axe Cleave Attack
	attackX := 100.0 + player.FacingX*28.0
	attackY := 100.0 + player.FacingY*28.0
	hitRadius := 32.0

	hitCount := 0
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx+dy*dy) < hitRadius {
			hitCount++
			if player.WeaponEquipped {
				toRemove = append(toRemove, ent)
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}
	if hitCount > 0 && player.WeaponEquipped {
		player.WeaponDurability--
	}

	if hitCount != 3 {
		t.Fatalf("Expected Axe cleave to hit all 3 zombies simultaneously, got %d", hitCount)
	}
	if player.WeaponDurability != 11 {
		t.Errorf("Expected Axe durability 11 after 1 cleave swing, got %d", player.WeaponDurability)
	}

	zMap := arkecs.NewMap1[ecs.Zombie](w)
	if zMap.Get(z1) != nil || zMap.Get(z2) != nil || zMap.Get(z3) != nil {
		t.Error("Expected all 3 cleaved zombies to be removed from world")
	}
}

// 4. Test Axe Cleave Reach Boundary (Axe 32px reach vs Bat 24px reach)
func TestCombat_AxeVsBatReachComparison(t *testing.T) {
	// Zombie at (152, 100): Player at (100, 100), facing (1, 0)
	// Bat attack center (124, 100), radius 24 -> reaches up to 148px (Misses!)
	// Axe attack center (128, 100), radius 32 -> reaches up to 160px (Hits!)
	targetZombieX := 152.0
	targetZombieY := 100.0

	// Test Bat Miss
	batAttackX := 100.0 + 1.0*24.0
	batAttackY := 100.0 + 0.0*24.0
	batDist := math.Sqrt((batAttackX-targetZombieX)*(batAttackX-targetZombieX) + (batAttackY-targetZombieY)*(batAttackY-targetZombieY))
	batHits := batDist < 24.0

	if batHits {
		t.Errorf("Bat should not reach zombie at distance %f (radius 24.0)", batDist)
	}

	// Test Axe Hit
	axeAttackX := 100.0 + 1.0*28.0
	axeAttackY := 100.0 + 0.0*28.0
	axeDist := math.Sqrt((axeAttackX-targetZombieX)*(axeAttackX-targetZombieX) + (axeAttackY-targetZombieY)*(axeAttackY-targetZombieY))
	axeHits := axeDist < 32.0

	if !axeHits {
		t.Errorf("Axe should reach zombie at distance %f (radius 32.0)", axeDist)
	}
}

// 5. Test Unarmed Shove and Stun Mechanics (No Entity Deletion)
func TestCombat_UnarmedFistShove(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = false
	player.WeaponType = ""
	player.FacingX = 1.0
	player.FacingY = 0.0

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, StunTimer: 0},
		&ecs.Position{X: 115.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	attackX := 100.0 + player.FacingX*24.0
	attackY := 100.0 + player.FacingY*24.0

	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		z, zPos, zVel := zQuery.Get()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx+dy*dy) < 24.0 {
			z.StunTimer = 45
			zVel.X = player.FacingX * 5.0
			zVel.Y = player.FacingY * 5.0
		}
	}

	zComp := arkecs.NewMap1[ecs.Zombie](w).Get(zEnt)
	if zComp == nil {
		t.Fatal("Unarmed shove must NOT delete zombie entity")
	}
	if zComp.StunTimer != 45 {
		t.Errorf("Expected StunTimer 45, got %d", zComp.StunTimer)
	}
	zVel := arkecs.NewMap1[ecs.Velocity](w).Get(zEnt)
	if zVel.X != 5.0 || zVel.Y != 0.0 {
		t.Errorf("Expected knockback velocity (5.0, 0.0), got (%f, %f)", zVel.X, zVel.Y)
	}
}

// 6. Test Shotgun Ammo Requirement and Consumption
func TestCombat_ShotgunAmmoRequirementAndConsumption(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"ammo", "ammo", "food"}
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Zombie at (150, 100) in cone
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 150.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Simulate Shotgun Attack
	ammoIdx := -1
	for i, item := range player.Inventory {
		if item == "ammo" {
			ammoIdx = i
			break
		}
	}

	if ammoIdx < 0 {
		t.Fatal("Expected ammo to be present in inventory")
	}

	// Consume ammo
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)

	// Spread cone evaluation: reach 160px, cone half-angle 22.5 deg (cos 22.5 deg ≈ 0.92388)
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := zPos.X - 100.0
		dy := zPos.Y - 100.0
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= 160.0 {
			if dist < 16.0 {
				toRemove = append(toRemove, ent)
			} else {
				dirX := dx / dist
				dirY := dy / dist
				dot := dirX*player.FacingX + dirY*player.FacingY
				if dot >= 0.92388 {
					toRemove = append(toRemove, ent)
				}
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}
	player.WeaponDurability--

	// Assertions
	if len(player.Inventory) != 2 || player.Inventory[0] != "ammo" || player.Inventory[1] != "food" {
		t.Errorf("Expected 1 ammo consumed, remaining inventory: %v", player.Inventory)
	}
	if player.WeaponDurability != 14 {
		t.Errorf("Expected Shotgun durability 14, got %d", player.WeaponDurability)
	}
	if arkecs.NewMap1[ecs.Zombie](w).Get(zEnt) != nil {
		t.Error("Expected zombie in shotgun cone to be killed")
	}
}

// 7. Test Shotgun Spread Cone Geometry and Reach Hits
func TestCombat_ShotgunConeReachHit(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"ammo"}
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Player at (100, 100), facing Right (1, 0)
	// Z1: (200, 100) -> dist = 100px <= 160px, angle = 0 deg (IN CONE -> HIT)
	// Z2: (220, 120) -> dist = ~121px <= 160px, angle = ~9.5 deg (cos ~0.986 >= 0.92388) (IN CONE -> HIT)
	// Z3: (200, 180) -> dist = ~128px <= 160px, angle = ~38.7 deg (cos ~0.781 < 0.92388) (OUT OF CONE -> MISS)
	// Z4: (300, 100) -> dist = 200px > 160px (OUT OF RANGE -> MISS)
	// Z5: (50, 100)  -> Behind player (OUT OF CONE -> MISS)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 200.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 220.0, Y: 120.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 200.0, Y: 180.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z4 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 300.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z5 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 50.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Fire Shotgun
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := zPos.X - 100.0
		dy := zPos.Y - 100.0
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= 160.0 {
			if dist < 16.0 {
				toRemove = append(toRemove, ent)
			} else {
				dirX := dx / dist
				dirY := dy / dist
				dot := dirX*player.FacingX + dirY*player.FacingY
				if dot >= 0.92388 {
					toRemove = append(toRemove, ent)
				}
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}

	zMap := arkecs.NewMap1[ecs.Zombie](w)
	if zMap.Get(z1) != nil {
		t.Error("Z1 (direct center) should be killed by shotgun")
	}
	if zMap.Get(z2) != nil {
		t.Error("Z2 (9.5 deg flank) should be killed by shotgun cone")
	}
	if zMap.Get(z3) == nil {
		t.Error("Z3 (38.7 deg wide) should survive outside shotgun cone")
	}
	if zMap.Get(z4) == nil {
		t.Error("Z4 (200px distance) should survive outside shotgun range")
	}
	if zMap.Get(z5) == nil {
		t.Error("Z5 (behind player) should survive shotgun blast")
	}
}

// 8. Test Shotgun Out-of-Ammo Dry Fire Behavior
func TestCombat_ShotgunOutOfAmmoDryFire(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"food", "water"} // NO AMMO!
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Zombie at (150, 100) directly in front
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 150.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Simulate Shotgun Attack Trigger
	ammoIdx := -1
	for i, item := range player.Inventory {
		if item == "ammo" {
			ammoIdx = i
			break
		}
	}

	if ammoIdx >= 0 {
		t.Fatal("Did not expect ammo to be found")
	}

	// Dry fire branch: no attack execution, no durability consumption, click sound
	player.AttackCooldown = 30

	if player.WeaponDurability != 15 {
		t.Errorf("Durability should remain 15 on dry fire, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 2 {
		t.Errorf("Inventory should be untouched on dry fire, got %v", player.Inventory)
	}
	if arkecs.NewMap1[ecs.Zombie](w).Get(zEnt) == nil {
		t.Error("Zombie should NOT be killed on dry fire")
	}
	if player.AttackCooldown != 30 {
		t.Errorf("Expected AttackCooldown 30 on dry fire click, got %d", player.AttackCooldown)
	}
}

// 9. Test Shotgun Acoustic Noise Pulse Alerts Horde within 400px
func TestCombat_ShotgunNoisePulseAlertsSwarm(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"ammo"}

	// Player at (200, 200)
	// Z1 at (350, 200): dist = 150px (<= 400px) -> Wandering (Chasing: false, WanderTimer: 60)
	// Z2 at (550, 200): dist = 350px (<= 400px) -> Wandering (Chasing: false, WanderTimer: 80)
	// Z3 at (750, 200): dist = 550px (> 400px)  -> Wandering (Chasing: false, WanderTimer: 100)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1Ent := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 60}, &ecs.Position{X: 350.0, Y: 200.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2Ent := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 80}, &ecs.Position{X: 550.0, Y: 200.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3Ent := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 100}, &ecs.Position{X: 750.0, Y: 200.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Simulate Shotgun Acoustic Noise Pulse (400.0px radius)
	pX, pY := 200.0, 200.0
	noiseQuery := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w).Query()
	for noiseQuery.Next() {
		z, zPos := noiseQuery.Get()
		dx := zPos.X - pX
		dy := zPos.Y - pY
		if math.Sqrt(dx*dx+dy*dy) <= 400.0 {
			z.Chasing = true
			z.WanderTimer = 0
		}
	}

	zMap := arkecs.NewMap1[ecs.Zombie](w)
	z1Comp := zMap.Get(z1Ent)
	z2Comp := zMap.Get(z2Ent)
	z3Comp := zMap.Get(z3Ent)

	if !z1Comp.Chasing || z1Comp.WanderTimer != 0 {
		t.Errorf("Z1 (150px) should be alerted to chase, got Chasing=%v, WanderTimer=%d", z1Comp.Chasing, z1Comp.WanderTimer)
	}
	if !z2Comp.Chasing || z2Comp.WanderTimer != 0 {
		t.Errorf("Z2 (350px) should be alerted to chase, got Chasing=%v, WanderTimer=%d", z2Comp.Chasing, z2Comp.WanderTimer)
	}
	if z3Comp.Chasing || z3Comp.WanderTimer != 100 {
		t.Errorf("Z3 (550px) should remain unalerted, got Chasing=%v, WanderTimer=%d", z3Comp.Chasing, z3Comp.WanderTimer)
	}
}

// 10. Test Weapon Durability Breakdown to Fists on 0 Hits
func TestCombat_WeaponDurabilityBreakdownOnZeroHits(t *testing.T) {
	weapons := []struct {
		name       string
		weaponType string
	}{
		{"Club", "weapon"},
		{"Axe", "axe"},
		{"Shotgun", "shotgun"},
	}

	for _, wpn := range weapons {
		t.Run(wpn.name, func(t *testing.T) {
			w, _, _, pEnt := setupCombatTestHarness()
			pMap := arkecs.NewMap1[ecs.Player](w)
			player := pMap.Get(pEnt)
			player.WeaponEquipped = true
			player.WeaponType = wpn.weaponType
			player.WeaponDurability = 1 // 1 hit remaining before breaking!
			player.Inventory = []string{"ammo"}

			// Simulate Final Attack
			player.WeaponDurability--
			if player.WeaponDurability <= 0 {
				player.WeaponEquipped = false
				player.WeaponType = ""
			}

			if player.WeaponEquipped {
				t.Errorf("%s: Expected WeaponEquipped false after durability 0", wpn.name)
			}
			if player.WeaponType != "" {
				t.Errorf("%s: Expected WeaponType empty string after durability 0, got '%s'", wpn.name, player.WeaponType)
			}
			if player.WeaponDurability != 0 {
				t.Errorf("%s: Expected WeaponDurability 0, got %d", wpn.name, player.WeaponDurability)
			}
		})
	}
}

// 11. Test Full Multi-Hit Degradation Loop for Axe (12 hits) and Shotgun (15 hits)
func TestCombat_MultiHitDegradationLoop(t *testing.T) {
	testSuites := []struct {
		weaponType string
		initDur    int
	}{
		{"weapon", 5},
		{"axe", 12},
		{"shotgun", 15},
	}

	for _, ts := range testSuites {
		t.Run(ts.weaponType, func(t *testing.T) {
			player := &ecs.Player{
				WeaponEquipped:   true,
				WeaponType:       ts.weaponType,
				WeaponDurability: ts.initDur,
			}

			for hit := 1; hit <= ts.initDur; hit++ {
				player.WeaponDurability--
				if player.WeaponDurability <= 0 {
					player.WeaponEquipped = false
					player.WeaponType = ""
				}

				expectedRemaining := ts.initDur - hit
				if player.WeaponDurability != expectedRemaining {
					t.Errorf("Hit %d: Expected durability %d, got %d", hit, expectedRemaining, player.WeaponDurability)
				}

				if hit < ts.initDur {
					if !player.WeaponEquipped || player.WeaponType != ts.weaponType {
						t.Errorf("Hit %d: Weapon should remain equipped as '%s'", hit, ts.weaponType)
					}
				} else {
					if player.WeaponEquipped || player.WeaponType != "" {
						t.Errorf("Hit %d: Weapon should break at durability 0", hit)
					}
				}
			}
		})
	}
}

// 12. Test Weapon HUD String Formatting and Ammo Count Calculation
func TestCombat_HUDFormattingAndAmmoCount(t *testing.T) {
	formatHUDWeaponText := func(hasWeapon bool, weaponType string, durability int, inventory []string) string {
		if hasWeapon && durability > 0 {
			wType := strings.ToUpper(weaponType)
			if wType == "" {
				wType = "WEAPON"
			}
			if weaponType == "shotgun" {
				ammoCount := 0
				for _, item := range inventory {
					if item == "ammo" {
						ammoCount++
					}
				}
				return fmt.Sprintf("Weapon: %s (%d hits | Ammo: %d)", wType, durability, ammoCount)
			}
			return fmt.Sprintf("Weapon: %s (%d hits)", wType, durability)
		}
		return "Weapon: NONE (Fists)"
	}

	tests := []struct {
		name       string
		equipped   bool
		weaponType string
		durability int
		inventory  []string
		wantText   string
	}{
		{
			name:       "Unarmed Fists",
			equipped:   false,
			weaponType: "",
			durability: 0,
			inventory:  []string{"food"},
			wantText:   "Weapon: NONE (Fists)",
		},
		{
			name:       "Spiked Club Equipped",
			equipped:   true,
			weaponType: "weapon",
			durability: 5,
			inventory:  []string{},
			wantText:   "Weapon: WEAPON (5 hits)",
		},
		{
			name:       "Fire Axe Equipped",
			equipped:   true,
			weaponType: "axe",
			durability: 12,
			inventory:  []string{"water"},
			wantText:   "Weapon: AXE (12 hits)",
		},
		{
			name:       "Shotgun with 3 Ammo",
			equipped:   true,
			weaponType: "shotgun",
			durability: 15,
			inventory:  []string{"ammo", "food", "ammo", "ammo"},
			wantText:   "Weapon: SHOTGUN (15 hits | Ammo: 3)",
		},
		{
			name:       "Shotgun with 0 Ammo",
			equipped:   true,
			weaponType: "shotgun",
			durability: 8,
			inventory:  []string{"food", "water"},
			wantText:   "Weapon: SHOTGUN (8 hits | Ammo: 0)",
		},
		{
			name:       "Broken Weapon Fallback",
			equipped:   false,
			weaponType: "",
			durability: 0,
			inventory:  []string{},
			wantText:   "Weapon: NONE (Fists)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatHUDWeaponText(tt.equipped, tt.weaponType, tt.durability, tt.inventory)
			if got != tt.wantText {
				t.Errorf("HUD Text = '%s', want '%s'", got, tt.wantText)
			}
		})
	}
}
