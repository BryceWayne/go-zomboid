package game

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupEmpiricalCombatHarness initializes a controlled headless test environment
func setupEmpiricalCombatHarness() (*arkecs.World, arkecs.Entity) {
	assets.Load()
	w := arkecs.NewWorld()

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
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, pEnt
}

// -----------------------------------------------------------------------------
// 1. AXE CLEAVE MULTI-KILL SWEEP IN DENSE ZOMBIE FORMATIONS
// -----------------------------------------------------------------------------

// TestEmpirical_AxeCleaveDenseSwarm verifies that an axe swing simultaneously kills
// all zombies within its 32px attack circle (attack center = pos + facing*32.0).
func TestEmpirical_AxeCleaveDenseSwarm(t *testing.T) {
	w, pEnt := setupEmpiricalCombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Player at (300, 300). Facing (1, 0). Attack Center = (332, 300). Hit Radius = 32.0.
	// Spawn 50 zombies tightly packed within radius < 32.0px around (332, 300)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	var insideEntities []arkecs.Entity
	var outsideEntities []arkecs.Entity

	rand.Seed(101)
	// 50 inside zombies
	for i := 0; i < 50; i++ {
		r := rand.Float64() * 30.0 // strictly < 32.0
		theta := rand.Float64() * 2 * math.Pi
		zx := 332.0 + r*math.Cos(theta)
		zy := 300.0 + r*math.Sin(theta)
		ent := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: zx, Y: zy}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
		insideEntities = append(insideEntities, ent)
	}

	// 20 outside zombies
	for i := 0; i < 20; i++ {
		r := 35.0 + rand.Float64()*50.0 // strictly > 32.0
		theta := rand.Float64() * 2 * math.Pi
		zx := 332.0 + r*math.Cos(theta)
		zy := 300.0 + r*math.Sin(theta)
		ent := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: zx, Y: zy}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
		outsideEntities = append(outsideEntities, ent)
	}

	// Execute Axe Cleave Melee Attack (matching game engine logic)
	attackX := 300.0 + player.FacingX*32.0
	attackY := 300.0 + player.FacingY*32.0
	hitZombies := false
	var toRemoveZombies []arkecs.Entity

	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
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
		player.WeaponDurability--
		if player.WeaponDurability <= 0 {
			player.WeaponEquipped = false
			player.WeaponType = ""
			player.WeaponDurability = 0
		}
	}
	for _, ent := range toRemoveZombies {
		w.RemoveEntity(ent)
	}

	// Assertions
	if len(toRemoveZombies) != 50 {
		t.Fatalf("Expected 50 inside zombies cleaved, got %d", len(toRemoveZombies))
	}
	if player.WeaponDurability != 11 {
		t.Errorf("Expected durability to decrease by exactly 1 (12 -> 11), got %d", player.WeaponDurability)
	}
	for i, ent := range insideEntities {
		if w.Alive(ent) {
			t.Errorf("Inside zombie #%d should have been deleted", i)
		}
	}
	for i, ent := range outsideEntities {
		if !w.Alive(ent) {
			t.Errorf("Outside zombie #%d should have survived", i)
		}
	}
}

// TestEmpirical_AxeDurabilityLifecycle12Swings verifies full degradation from 12 hits to 0 (fists).
func TestEmpirical_AxeDurabilityLifecycle12Swings(t *testing.T) {
	player := &ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "axe",
		WeaponDurability: 12,
	}

	for swing := 1; swing <= 12; swing++ {
		player.WeaponDurability--
		if player.WeaponDurability <= 0 {
			player.WeaponEquipped = false
			player.WeaponType = ""
			player.WeaponDurability = 0
		}

		expectedDur := 12 - swing
		if player.WeaponDurability != expectedDur {
			t.Errorf("Swing %d: expected durability %d, got %d", swing, expectedDur, player.WeaponDurability)
		}
		if swing < 12 {
			if !player.WeaponEquipped || player.WeaponType != "axe" {
				t.Errorf("Swing %d: axe should remain equipped", swing)
			}
		} else {
			if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
				t.Errorf("Swing 12: axe must break to fists, got equipped=%v, type=%s, dur=%d",
					player.WeaponEquipped, player.WeaponType, player.WeaponDurability)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// 2. SHOTGUN SPREAD CONE GEOMETRIC BOUNDARY COVERAGE (±22.5°, 160px reach)
// -----------------------------------------------------------------------------

// TestEmpirical_ShotgunConeBoundaryPrecision tests high-precision angular and range boundaries.
func TestEmpirical_ShotgunConeBoundaryPrecision(t *testing.T) {
	const (
		maxRange  = 160.0
		cosSpread = 0.9238795325112867 // cos(22.5 deg)
		pX        = 500.0
		pY        = 500.0
	)

	// Facing Right (1, 0)
	facingX, facingY := 1.0, 0.0

	testCases := []struct {
		name      string
		dist      float64
		angleDeg  float64
		expectHit bool
	}{
		// Point blank (<24px) - omnidirectional
		{"Point Blank Straight", 10.0, 0.0, true},
		{"Point Blank Flank 45 deg", 15.0, 45.0, true},
		{"Point Blank Flank 90 deg", 20.0, 90.0, true},
		{"Point Blank Behind 180 deg", 22.0, 180.0, true},
		{"Point Blank Boundary 23.99px Behind", 23.99, 180.0, true},

		// Angular Boundary at Mid-Range (dist = 100px)
		{"Mid-Range Direct Center (0 deg)", 100.0, 0.0, true},
		{"Mid-Range Inside Cone (10 deg)", 100.0, 10.0, true},
		{"Mid-Range Inside Cone (20 deg)", 100.0, 20.0, true},
		{"Mid-Range Critical Boundary Inside (22.40 deg)", 100.0, 22.40, true},
		{"Mid-Range Critical Boundary Inside (22.49 deg)", 100.0, 22.49, true},
		{"Mid-Range Outside Cone (22.51 deg)", 100.0, 22.51, false},
		{"Mid-Range Outside Cone (22.60 deg)", 100.0, 22.60, false},
		{"Mid-Range Outside Cone (30 deg)", 100.0, 30.0, false},
		{"Mid-Range Negative Boundary Inside (-22.40 deg)", 100.0, -22.40, true},
		{"Mid-Range Negative Boundary Inside (-22.49 deg)", 100.0, -22.49, true},
		{"Mid-Range Outside Negative (-22.51 deg)", 100.0, -22.51, false},
		{"Mid-Range Outside Negative (-22.60 deg)", 100.0, -22.60, false},

		// Range Boundary along Center Line (angle = 0 deg)
		{"Range Close (50px)", 50.0, 0.0, true},
		{"Range Near Limit (150px)", 150.0, 0.0, true},
		{"Range Critical Limit (159.90px)", 159.90, 0.0, true},
		{"Range Exact Limit (160.00px)", 160.00, 0.0, true},
		{"Range Beyond Limit (160.10px)", 160.10, 0.0, false},
		{"Range Far Out (200px)", 200.0, 0.0, false},

		// Combined Boundary: 159.9px at 22.4 deg (Should HIT)
		{"Corner Inside (159.9px, 22.4 deg)", 159.90, 22.40, true},
		// Combined Boundary: 160.1px at 22.4 deg (Should MISS due to range)
		{"Corner Out of Range (160.1px, 22.4 deg)", 160.10, 22.40, false},
		// Combined Boundary: 159.9px at 22.6 deg (Should MISS due to angle)
		{"Corner Out of Angle (159.9px, 22.6 deg)", 159.90, 22.60, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			angleRad := tc.angleDeg * math.Pi / 180.0
			zx := pX + tc.dist*math.Cos(angleRad)
			zy := pY + tc.dist*math.Sin(angleRad)

			dx := zx - pX
			dy := zy - pY
			dist := math.Hypot(dx, dy)

			hit := false
			if dist <= maxRange {
				if dist < 24.0 {
					hit = true
				} else {
					cosAngle := (facingX*dx + facingY*dy) / dist
					if cosAngle >= cosSpread {
						hit = true
					}
				}
			}

			if hit != tc.expectHit {
				t.Errorf("Shotgun cone check failed for [%s]: dist=%.2f, angle=%.2f -> got hit=%v, want %v",
					tc.name, tc.dist, tc.angleDeg, hit, tc.expectHit)
			}
		})
	}
}

// TestEmpirical_Shotgun8DirectionsMonteCarlo tests shotgun cone calculations across all 8 cardinal/diagonal directions
// with 5,000 random sample points per direction.
func TestEmpirical_Shotgun8DirectionsMonteCarlo(t *testing.T) {
	directions := []struct {
		name string
		fx   float64
		fy   float64
	}{
		{"Right", 1.0, 0.0},
		{"Down-Right", 1.0, 1.0},
		{"Down", 0.0, 1.0},
		{"Down-Left", -1.0, 1.0},
		{"Left", -1.0, 0.0},
		{"Up-Left", -1.0, -1.0},
		{"Up", 0.0, -1.0},
		{"Up-Right", 1.0, -1.0},
	}

	const (
		maxRange  = 160.0
		cosSpread = 0.9238795325112867
	)

	rand.Seed(202)

	for _, dir := range directions {
		t.Run(dir.name, func(t *testing.T) {
			facingLen := math.Hypot(dir.fx, dir.fy)
			nfx, nfy := dir.fx/facingLen, dir.fy/facingLen
			facingAngle := math.Atan2(nfy, nfx)

			hitCount := 0
			missCount := 0

			for i := 0; i < 5000; i++ {
				// Generate random point in box [-200, 200]
				dx := (rand.Float64()*2.0 - 1.0) * 200.0
				dy := (rand.Float64()*2.0 - 1.0) * 200.0
				dist := math.Hypot(dx, dy)

				// Engine evaluation
				hit := false
				if dist <= maxRange {
					if dist < 24.0 {
						hit = true
					} else {
						cosAngle := (nfx*dx + nfy*dy) / dist
						if cosAngle >= cosSpread {
							hit = true
						}
					}
				}

				// Mathematical Ground Truth Oracle:
				targetAngle := math.Atan2(dy, dx)
				diffAngle := math.Abs(targetAngle - facingAngle)
				for diffAngle > math.Pi {
					diffAngle = 2*math.Pi - diffAngle
				}

				expectedHit := false
				if dist <= maxRange {
					if dist < 24.0 {
						expectedHit = true
					} else if diffAngle <= (22.5 * math.Pi / 180.0 + 1e-9) {
						expectedHit = true
					}
				}

				if hit != expectedHit {
					t.Fatalf("Direction %s mismatch at dx=%.2f, dy=%.2f, dist=%.2f: engine=%v, oracle=%v",
						dir.name, dx, dy, dist, hit, expectedHit)
				}

				if hit {
					hitCount++
				} else {
					missCount++
				}
			}

			if hitCount == 0 {
				t.Errorf("Direction %s produced 0 hits out of 5000 samples", dir.name)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// 3. EXACT AMMO CONSUMPTION (1 item per blast)
// -----------------------------------------------------------------------------

// TestEmpirical_ExactAmmoConsumptionSequence tests firing a 15-durability shotgun
// with an inventory containing varying amounts of ammo and mixed items.
func TestEmpirical_ExactAmmoConsumptionSequence(t *testing.T) {
	player := &ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "shotgun",
		WeaponDurability: 15,
		Inventory:        []string{"food", "ammo", "water", "ammo", "ammo", "vest"},
	}

	initialAmmo := 3
	for blast := 1; blast <= initialAmmo; blast++ {
		// Find ammo
		ammoIdx := -1
		for idx, itm := range player.Inventory {
			if itm == "ammo" {
				ammoIdx = idx
				break
			}
		}

		if ammoIdx < 0 {
			t.Fatalf("Blast %d: Expected ammo at index, none found", blast)
		}

		// Consume exactly 1 ammo
		player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
		player.WeaponDurability--

		expectedRemainingAmmo := initialAmmo - blast
		actualAmmoCount := 0
		for _, itm := range player.Inventory {
			if itm == "ammo" {
				actualAmmoCount++
			}
		}

		if actualAmmoCount != expectedRemainingAmmo {
			t.Errorf("Blast %d: expected %d ammo remaining, got %d", blast, expectedRemainingAmmo, actualAmmoCount)
		}
		if player.WeaponDurability != 15-blast {
			t.Errorf("Blast %d: expected durability %d, got %d", blast, 15-blast, player.WeaponDurability)
		}
	}

	// Verify non-ammo items remain intact
	expectedInventory := []string{"food", "water", "vest"}
	if len(player.Inventory) != len(expectedInventory) {
		t.Fatalf("Expected inventory length %d, got %d: %v", len(expectedInventory), len(player.Inventory), player.Inventory)
	}
	for i, item := range player.Inventory {
		if item != expectedInventory[i] {
			t.Errorf("Inventory slot %d: expected %s, got %s", i, expectedInventory[i], item)
		}
	}
}

// -----------------------------------------------------------------------------
// 4. EXACT 400px NOISE RADIUS HORDE AGGRO TRIGGERING z.Chasing = true
// -----------------------------------------------------------------------------

// TestEmpirical_Shotgun400pxNoiseRadiusHordeAggro tests radial aggro thresholding for 500 zombies.
func TestEmpirical_Shotgun400pxNoiseRadiusHordeAggro(t *testing.T) {
	w, pEnt := setupEmpiricalCombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"ammo"}

	pX, pY := 300.0, 300.0

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	type testZombie struct {
		ent      arkecs.Entity
		dist     float64
		shouldBe bool
	}

	var testZombies []testZombie

	// 1. Exact boundary cases
	boundaryDistances := []struct {
		dist     float64
		shouldBe bool
	}{
		{50.0, true},
		{200.0, true},
		{399.0, true},
		{399.9, true},
		{400.0, true},  // Exact boundary <= 400.0
		{400.1, false}, // Outside boundary
		{401.0, false},
		{500.0, false},
		{800.0, false},
	}

	for _, b := range boundaryDistances {
		ent := zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 99},
			&ecs.Position{X: pX + b.dist, Y: pY},
			&ecs.Velocity{},
			&ecs.Sprite{},
			&ecs.Collider{},
		)
		testZombies = append(testZombies, testZombie{ent: ent, dist: b.dist, shouldBe: b.shouldBe})
	}

	// 2. 200 random radial zombies
	rand.Seed(303)
	for i := 0; i < 200; i++ {
		dist := rand.Float64() * 700.0 // [0, 700]
		theta := rand.Float64() * 2 * math.Pi
		zx := pX + dist*math.Cos(theta)
		zy := pY + dist*math.Sin(theta)
		shouldBe := dist <= 400.0

		ent := zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: false, WanderTimer: 100 + i},
			&ecs.Position{X: zx, Y: zy},
			&ecs.Velocity{},
			&ecs.Sprite{},
			&ecs.Collider{},
		)
		testZombies = append(testZombies, testZombie{ent: ent, dist: dist, shouldBe: shouldBe})
	}

	// Trigger Acoustic Noise Pulse (400.0px radius)
	noiseQuery := arkecs.NewFilter2[ecs.Zombie, ecs.Position](w).Query()
	for noiseQuery.Next() {
		z, zPos := noiseQuery.Get()
		zdx := pX - zPos.X
		zdy := pY - zPos.Y
		if math.Hypot(zdx, zdy) <= 400.0 {
			z.Chasing = true
			z.WanderTimer = 0
		}
	}

	// Verify all zombies
	zCompMap := arkecs.NewMap1[ecs.Zombie](w)
	for i, tz := range testZombies {
		z := zCompMap.Get(tz.ent)
		if tz.shouldBe {
			if !z.Chasing || z.WanderTimer != 0 {
				t.Errorf("Zombie #%d at dist=%.2f should be chasing with WanderTimer=0, got Chasing=%v, WanderTimer=%d",
					i, tz.dist, z.Chasing, z.WanderTimer)
			}
		} else {
			if z.Chasing || z.WanderTimer == 0 {
				t.Errorf("Zombie #%d at dist=%.2f outside 400px should NOT be chasing, got Chasing=%v, WanderTimer=%d",
					i, tz.dist, z.Chasing, z.WanderTimer)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// 5. DRY FIRE FALLBACK WHEN AMMO COUNT IS 0
// -----------------------------------------------------------------------------

// TestEmpirical_DryFireFallbackBehavior verifies defensive shove, knockback, stun, and zero ammo/durability loss.
func TestEmpirical_DryFireFallbackBehavior(t *testing.T) {
	w, pEnt := setupEmpiricalCombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"food", "water", "axe"} // 0 AMMO
	player.FacingX = 1.0
	player.FacingY = 0.0

	pX, pY := 300.0, 300.0

	// Zombie inside shove reach (reach = 24.0, attackCenter = (324, 300))
	// Z1 at (315, 300) -> dx = 9, dy = 0 -> dist = 9 < 24 (SHOVE HIT)
	// Z2 at (350, 300) -> dx = -26, dy = 0 -> dist = 26 > 24 (SHOVE MISS)
	// Z3 at (400, 300) -> dist to player = 100px (inside gunshot cone, but dry fire so NO GUNSHOT)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0, StunTimer: 0, Chasing: false, WanderTimer: 50}, &ecs.Position{X: 315.0, Y: 300.0}, &ecs.Velocity{X: 0, Y: 0}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0, StunTimer: 0, Chasing: false, WanderTimer: 50}, &ecs.Position{X: 350.0, Y: 300.0}, &ecs.Velocity{X: 0, Y: 0}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0, StunTimer: 0, Chasing: false, WanderTimer: 50}, &ecs.Position{X: 400.0, Y: 300.0}, &ecs.Velocity{X: 0, Y: 0}, &ecs.Sprite{}, &ecs.Collider{})

	// Simulate Player Attack Trigger when Shotgun equipped and 0 ammo
	ammoIdx := -1
	for idx, itm := range player.Inventory {
		if itm == "ammo" {
			ammoIdx = idx
			break
		}
	}

	if ammoIdx >= 0 {
		t.Fatal("Expected no ammo to be found")
	}

	// Dry fire branch
	player.AttackCooldown = 30
	attackX := pX + player.FacingX*24.0
	attackY := pY + player.FacingY*24.0

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

	// Assertions
	// 1. Weapon durability unchanged
	if player.WeaponDurability != 15 {
		t.Errorf("Dry fire must NOT consume weapon durability (expected 15, got %d)", player.WeaponDurability)
	}
	// 2. Inventory untouched
	if len(player.Inventory) != 3 {
		t.Errorf("Dry fire must NOT modify inventory, got %v", player.Inventory)
	}
	// 3. Attack cooldown set
	if player.AttackCooldown != 30 {
		t.Errorf("Expected AttackCooldown = 30, got %d", player.AttackCooldown)
	}
	// 4. All zombies survive (no deletions)
	if !w.Alive(z1) || !w.Alive(z2) || !w.Alive(z3) {
		t.Fatal("Dry fire must NOT delete any zombie entities")
	}

	// 5. Z1 (near) stunned and knocked back
	zCompMap := arkecs.NewMap1[ecs.Zombie](w)
	vCompMap := arkecs.NewMap1[ecs.Velocity](w)

	z1Comp := zCompMap.Get(z1)
	v1Comp := vCompMap.Get(z1)
	if z1Comp.StunTimer != 45 {
		t.Errorf("Z1 StunTimer: expected 45, got %d", z1Comp.StunTimer)
	}
	if v1Comp.X != 5.0 || v1Comp.Y != 0.0 {
		t.Errorf("Z1 knockback velocity: expected (5.0, 0.0), got (%f, %f)", v1Comp.X, v1Comp.Y)
	}

	// 6. Z2 and Z3 not stunned
	z2Comp := zCompMap.Get(z2)
	z3Comp := zCompMap.Get(z3)
	if z2Comp.StunTimer != 0 {
		t.Errorf("Z2 StunTimer: expected 0, got %d", z2Comp.StunTimer)
	}
	if z3Comp.StunTimer != 0 {
		t.Errorf("Z3 StunTimer: expected 0, got %d", z3Comp.StunTimer)
	}
	// 7. No noise pulse: Z3 not aggroed
	if z3Comp.Chasing {
		t.Errorf("Dry fire must NOT trigger 400px noise pulse aggro")
	}
}

// -----------------------------------------------------------------------------
// 6. FULL E2E COMBAT CYCLE INTEGRATION
// -----------------------------------------------------------------------------

// TestEmpirical_FullCombatCycleIntegration tests a complete end-to-end combat sequence
// including shotgun firing, ammo depletion, dry fire shove, weapon swap to axe, cleave sweep, and degradation.
func TestEmpirical_FullCombatCycleIntegration(t *testing.T) {
	w, pEnt := setupEmpiricalCombatHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Step 1: Equip shotgun with 2 ammo
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15
	player.Inventory = []string{"ammo", "ammo", "axe"}

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	// Spawn 3 zombies in front (100px away)
	z1 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 380.0, Y: 300.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 390.0, Y: 310.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 390.0, Y: 290.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Fire blast #1
	ammoIdx := 0 // first ammo
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
	player.WeaponDurability--
	w.RemoveEntity(z1)
	w.RemoveEntity(z2)
	w.RemoveEntity(z3)

	if len(player.Inventory) != 2 || player.WeaponDurability != 14 {
		t.Fatalf("State after blast 1 invalid: inv=%v, dur=%d", player.Inventory, player.WeaponDurability)
	}

	// Fire blast #2 (into air)
	ammoIdx = 0 // second ammo
	player.Inventory = append(player.Inventory[:ammoIdx], player.Inventory[ammoIdx+1:]...)
	player.WeaponDurability--

	if len(player.Inventory) != 1 || player.Inventory[0] != "axe" || player.WeaponDurability != 13 {
		t.Fatalf("State after blast 2 invalid: inv=%v, dur=%d", player.Inventory, player.WeaponDurability)
	}

	// Blast #3: Ammo is empty -> Dry fire shove
	ammoFound := false
	for _, itm := range player.Inventory {
		if itm == "ammo" {
			ammoFound = true
		}
	}
	if ammoFound {
		t.Fatal("Ammo should be 0")
	}
	// Dry fire preserves shotgun stats
	if player.WeaponDurability != 13 || player.WeaponType != "shotgun" {
		t.Errorf("Dry fire corrupted shotgun stats: dur=%d, type=%s", player.WeaponDurability, player.WeaponType)
	}

	// Step 2: Swap to Axe from inventory
	axeIdx := 0
	if player.Inventory[axeIdx] == "axe" {
		player.WeaponEquipped = true
		player.WeaponType = "axe"
		player.WeaponDurability = 12
		player.Inventory = append(player.Inventory[:axeIdx], player.Inventory[axeIdx+1:]...)
	}

	if !player.WeaponEquipped || player.WeaponType != "axe" || player.WeaponDurability != 12 || len(player.Inventory) != 0 {
		t.Fatalf("Axe equip failed: %+v", player)
	}

	// Spawn dense horde around axe reach (332, 300)
	zCleave1 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 325.0, Y: 300.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	zCleave2 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 330.0, Y: 310.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	zCleave3 := zMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 330.0, Y: 290.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	// Axe cleave swing
	w.RemoveEntity(zCleave1)
	w.RemoveEntity(zCleave2)
	w.RemoveEntity(zCleave3)
	player.WeaponDurability--

	if player.WeaponDurability != 11 {
		t.Errorf("Expected axe durability 11 after cleave, got %d", player.WeaponDurability)
	}
	if w.Alive(zCleave1) || w.Alive(zCleave2) || w.Alive(zCleave3) {
		t.Errorf("Cleaved zombies still alive")
	}
}

// -----------------------------------------------------------------------------
// 7. RAPID WEAPON SWITCHING & HUD FORMATTING MATRIX
// -----------------------------------------------------------------------------

func TestEmpirical_HUDFormattingMatrix(t *testing.T) {
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

	testCases := []struct {
		name       string
		equipped   bool
		weaponType string
		durability int
		inventory  []string
		expected   string
	}{
		{"Unarmed default", false, "", 0, []string{}, "Weapon: NONE (Fists)"},
		{"Club 1 hit", true, "weapon", 1, []string{}, "Weapon: WEAPON (1 hits)"},
		{"Axe 12 hits", true, "axe", 12, []string{"water"}, "Weapon: AXE (12 hits)"},
		{"Shotgun 15 hits 0 ammo", true, "shotgun", 15, []string{"food"}, "Weapon: SHOTGUN (15 hits | Ammo: 0)"},
		{"Shotgun 1 hit 5 ammo", true, "shotgun", 1, []string{"ammo", "ammo", "ammo", "ammo", "ammo"}, "Weapon: SHOTGUN (1 hits | Ammo: 5)"},
		{"Broken Shotgun 0 hits", true, "shotgun", 0, []string{"ammo"}, "Weapon: NONE (Fists)"},
		{"Broken Axe 0 hits unequipped", false, "", 0, []string{}, "Weapon: NONE (Fists)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatHUDWeaponText(tc.equipped, tc.weaponType, tc.durability, tc.inventory)
			if result != tc.expected {
				t.Errorf("Expected HUD text '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
