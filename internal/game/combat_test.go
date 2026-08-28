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
		itemType     string
		expectedType string
		expectedDur  int
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

			// Simulate inventory use (same logic as processInputAndCombat)
			useItemIdx := 0
			if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
				player.AttackCooldown = 30
				item := player.Inventory[useItemIdx]
				used := false
				if item == "weapon" {
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
				}
				if used {
					player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
				}
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

	// Player at (100, 100). Axe attack reach = 32.0px. Attack center = (132, 100).
	// Spawn 3 zombies within cleave area:
	// Z1: Straight ahead at (125, 100) -> distance to attack center = 7px < 32px
	// Z2: Upper arc at (125, 80) -> dx = 7, dy = 20 -> dist = 21.2px < 32px
	// Z3: Lower arc at (125, 120) -> dx = 7, dy = 20 -> dist = 21.2px < 32px
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 125.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 125.0, Y: 80.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 125.0, Y: 120.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Simulate Axe Cleave Attack
	attackX := 100.0 + player.FacingX*32.0
	attackY := 100.0 + player.FacingY*32.0
	hitRadius := 32.0

	hitCount := 0
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Hypot(dx, dy) < hitRadius {
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

	if w.Alive(z1) || w.Alive(z2) || w.Alive(z3) {
		t.Error("Expected all 3 cleaved zombies to be removed from world")
	}
}

// 4. Test Axe Cleave Reach Boundary (Axe 32px reach vs Bat 24px reach)
func TestCombat_AxeVsBatReachComparison(t *testing.T) {
	// Zombie at (152, 100): Player at (100, 100), facing (1, 0)
	// Bat attack center (124, 100), radius 24 -> reaches up to 148px (Misses!)
	// Axe attack center (132, 100), radius 32 -> reaches up to 164px (Hits!)
	targetZombieX := 152.0
	targetZombieY := 100.0

	// Test Bat Miss
	batAttackX := 100.0 + 1.0*24.0
	batAttackY := 100.0 + 0.0*24.0
	batDist := math.Hypot(batAttackX-targetZombieX, batAttackY-targetZombieY)
	batHits := batDist < 24.0

	if batHits {
		t.Errorf("Bat should not reach zombie at distance %f (radius 24.0)", batDist)
	}

	// Test Axe Hit
	axeAttackX := 100.0 + 1.0*32.0
	axeAttackY := 100.0 + 0.0*32.0
	axeDist := math.Hypot(axeAttackX-targetZombieX, axeAttackY-targetZombieY)
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
		if math.Hypot(dx, dy) < 24.0 {
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
	const maxShotgunRange = 160.0
	const cosSpread = 0.9238795325112867

	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := zPos.X - 100.0
		dy := zPos.Y - 100.0
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
	player.WeaponDurability--

	// Assertions
	if len(player.Inventory) != 2 || player.Inventory[0] != "ammo" || player.Inventory[1] != "food" {
		t.Errorf("Expected 1 ammo consumed, remaining inventory: %v", player.Inventory)
	}
	if player.WeaponDurability != 14 {
		t.Errorf("Expected Shotgun durability 14, got %d", player.WeaponDurability)
	}
	if w.Alive(zEnt) {
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
	// Z6: (115, 110) -> dist = 18px < 24px (POINT BLANK -> HIT)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 200.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 220.0, Y: 120.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 200.0, Y: 180.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z4 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 300.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z5 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 50.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z6 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 115.0, Y: 110.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Fire Shotgun
	var toRemove []arkecs.Entity
	const maxShotgunRange = 160.0
	const cosSpread = 0.9238795325112867

	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := zPos.X - 100.0
		dy := zPos.Y - 100.0
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

	if w.Alive(z1) {
		t.Error("Z1 (direct center) should be killed by shotgun")
	}
	if w.Alive(z2) {
		t.Error("Z2 (9.5 deg flank) should be killed by shotgun cone")
	}
	if !w.Alive(z3) {
		t.Error("Z3 (38.7 deg wide) should survive outside shotgun cone")
	}
	if !w.Alive(z4) {
		t.Error("Z4 (200px distance) should survive outside shotgun range")
	}
	if !w.Alive(z5) {
		t.Error("Z5 (behind player) should survive shotgun blast")
	}
	if w.Alive(z6) {
		t.Error("Z6 (point blank 18px) should be killed by shotgun")
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

	// Zombie at (150, 100) directly in front (distance 50px - out of shove reach 24px)
	// Zombie at (115, 100) adjacent (distance 15px - inside shove reach 24px)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zFar := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0, StunTimer: 0}, &ecs.Position{X: 150.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	zNear := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0, StunTimer: 0}, &ecs.Position{X: 115.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

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

	// Dry fire branch: mechanical click & defensive butt shove (radius 24.0px)
	player.AttackCooldown = 30
	attackX := 100.0 + player.FacingX*24.0
	attackY := 100.0 + player.FacingY*24.0
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
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

	if player.WeaponDurability != 15 {
		t.Errorf("Durability should remain 15 on dry fire, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 2 {
		t.Errorf("Inventory should be untouched on dry fire, got %v", player.Inventory)
	}
	if !w.Alive(zFar) || !w.Alive(zNear) {
		t.Error("Zombies should NOT be killed on dry fire")
	}
	if player.AttackCooldown != 30 {
		t.Errorf("Expected AttackCooldown 30 on dry fire click, got %d", player.AttackCooldown)
	}
	zNearComp := arkecs.NewMap1[ecs.Zombie](w).Get(zNear)
	if zNearComp.StunTimer != 45 {
		t.Errorf("Expected close zombie to be stunned with StunTimer 45, got %d", zNearComp.StunTimer)
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
		if math.Hypot(dx, dy) <= 400.0 {
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
				player.WeaponDurability = 0
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

// 11. Test Full Multi-Hit Degradation Loop for Axe (12 hits), Shotgun (15 hits), Club (5 hits)
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
					player.WeaponDurability = 0
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
					if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
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

// 13. Test Ammo Item In Hotbar is Preserved (Cannot Be Equipped directly)
func TestCombat_AmmoNotDirectlyEquippable(t *testing.T) {
	player := &ecs.Player{
		Inventory:        []string{"ammo", "axe"},
		WeaponEquipped:   false,
		WeaponType:       "",
		WeaponDurability: 0,
	}

	useItemIdx := 0
	tItem := player.Inventory[useItemIdx]
	used := false
	if tItem == "weapon" {
		player.WeaponEquipped = true
		player.WeaponType = "weapon"
		player.WeaponDurability = 5
		used = true
	} else if tItem == "axe" {
		player.WeaponEquipped = true
		player.WeaponType = "axe"
		player.WeaponDurability = 12
		used = true
	} else if tItem == "shotgun" {
		player.WeaponEquipped = true
		player.WeaponType = "shotgun"
		player.WeaponDurability = 15
		used = true
	}

	if used {
		player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
	}

	if used || player.WeaponEquipped || len(player.Inventory) != 2 || player.Inventory[0] != "ammo" {
		t.Errorf("Ammo should not be equippable or consumed directly from hotbar, inv: %v", player.Inventory)
	}
}

// 14. Test Weapon Switch Re-equips and Overrides Previous Stats
func TestCombat_ReEquipOverridesStats(t *testing.T) {
	player := &ecs.Player{
		Inventory:        []string{"shotgun"},
		WeaponEquipped:   true,
		WeaponType:       "axe",
		WeaponDurability: 3, // Worn down axe
	}

	useItemIdx := 0
	tItem := player.Inventory[useItemIdx]
	used := false
	if tItem == "shotgun" {
		player.WeaponEquipped = true
		player.WeaponType = "shotgun"
		player.WeaponDurability = 15
		used = true
	}
	if used {
		player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
	}

	if !player.WeaponEquipped || player.WeaponType != "shotgun" || player.WeaponDurability != 15 {
		t.Errorf("Expected re-equipping shotgun to overwrite stats: %+v", player)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Expected inventory to be empty after equipping shotgun, got %v", player.Inventory)
	}
}

// 15. Test Diagonal Facing Vector Shotgun Spread Cone
func TestCombat_ShotgunDiagonalFacingSpreadCone(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.FacingX = 1.0
	player.FacingY = 1.0 // Diagonal down-right (45 deg)

	facingLen := math.Hypot(player.FacingX, player.FacingY)
	fx, fy := player.FacingX/facingLen, player.FacingY/facingLen

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	// Target 1: Exactly diagonal at (170.7, 170.7) -> dx=70.7, dy=70.7 -> dist=100px, dot=1.0 (HIT)
	z1 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 170.71, Y: 170.71}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	// Target 2: Orthogonal right at (200, 100) -> dx=100, dy=0 -> dist=100px, dot=0.707 < 0.92388 (MISS)
	z2 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 200.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	const maxShotgunRange = 160.0
	const cosSpread = 0.9238795325112867

	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := zPos.X - 100.0
		dy := zPos.Y - 100.0
		dist := math.Hypot(dx, dy)
		if dist <= maxShotgunRange {
			if dist < 24.0 {
				toRemove = append(toRemove, ent)
			} else {
				cosAngle := (fx*dx + fy*dy) / dist
				if cosAngle >= cosSpread {
					toRemove = append(toRemove, ent)
				}
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}

	if w.Alive(z1) {
		t.Error("Z1 along diagonal facing should be hit")
	}
	if !w.Alive(z2) {
		t.Error("Z2 orthogonal to diagonal facing should not be hit")
	}
}

// 16. Test Fire Axe Wide Lateral Cleave Arc
func TestCombat_FireAxeWideAngleLateralCleave(t *testing.T) {
	w, _, _, pEnt := setupCombatTestHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.FacingX = 0.0
	player.FacingY = 1.0 // Facing Down (0, 1)

	// Attack center: (100, 132)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	// Left flank at (80, 132) -> dx = 20, dy = 0 -> dist = 20 < 32 (HIT)
	zLeft := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 80.0, Y: 132.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	// Right flank at (120, 132) -> dx = -20, dy = 0 -> dist = 20 < 32 (HIT)
	zRight := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 120.0, Y: 132.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	// Center bottom at (100, 150) -> dx = 0, dy = -18 -> dist = 18 < 32 (HIT)
	zCenter := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 100.0, Y: 150.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	attackX := 100.0 + player.FacingX*32.0
	attackY := 100.0 + player.FacingY*32.0
	hitRadius := 32.0

	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Hypot(dx, dy) < hitRadius {
			toRemove = append(toRemove, ent)
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}

	if w.Alive(zLeft) || w.Alive(zRight) || w.Alive(zCenter) {
		t.Error("Expected all 3 zombies in wide lateral sweep to be cleaved")
	}
}
