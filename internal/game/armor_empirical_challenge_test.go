package game

import (
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupEmpiricalArmorHarness creates a controlled single-player game environment
func setupEmpiricalArmorHarness() (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	sys := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			Hunger:             100.0,
			Thirst:             100.0,
			Inventory:          []string{},
			WeaponEquipped:     false,
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
			FacingX:            1,
			FacingY:            0,
		},
		&ecs.Position{X: 100, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, sys, pEnt
}

// 1. Empirical Challenge: Statistical Deflection Distribution over 10,000 rolls
// Verifies empirical deflection rate for InfectionResist = 0.70 matches ~70% (3-sigma bound [68.5%, 71.5%])
func TestEmpirical_StatisticalDeflectionDistribution_10000Rolls(t *testing.T) {
	assets.Load()
	rand.Seed(42) // Deterministic seed for reproducible statistical harness

	totalRolls := 10000
	deflectionCount := 0
	infectionCount := 0

	for i := 0; i < totalRolls; i++ {
		w := arkecs.NewWorld()
		m := world.NewMap(30, 30)
		sys := NewUpdateSystem(w, m)

		pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		pEnt := pMap.NewEntity(
			&ecs.Player{
				Health:             100.0,
				Hunger:             100.0,
				Thirst:             100.0,
				ArmorEquipped:      true,
				ArmorType:          "vest",
				ArmorDefense:       0.50,
				ArmorDurability:    10,
				ArmorMaxDurability: 10,
				InfectionResist:    0.70,
				Dead:               false,
				Infected:           false,
			},
			&ecs.Position{X: 100, Y: 100},
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: true},
			&ecs.Position{X: 105, Y: 100}, // Distance = 5.0 (< 14.0 contact range)
			&ecs.Velocity{X: 0, Y: 0},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		sys.processZombies()

		playerMap1 := arkecs.NewMap1[ecs.Player](w)
		player := playerMap1.Get(pEnt)
		if player.Infected {
			infectionCount++
		} else {
			deflectionCount++
		}

		// Verify durability was decremented regardless of deflection outcome
		if player.ArmorDurability != 9 {
			t.Fatalf("Roll %d: Expected armor durability 9 on contact, got %d", i, player.ArmorDurability)
		}
	}

	deflectionRate := float64(deflectionCount) / float64(totalRolls)
	infectionRate := float64(infectionCount) / float64(totalRolls)

	t.Logf("Empirical 10,000 Roll Deflection Results:")
	t.Logf("  Total Trials: %d", totalRolls)
	t.Logf("  Deflections: %d (%.4f%%)", deflectionCount, deflectionRate*100)
	t.Logf("  Infections:  %d (%.4f%%)", infectionCount, infectionRate*100)

	// Statistical 3-sigma bounds for N=10000, p=0.70:
	// sigma = sqrt(10000 * 0.70 * 0.30) = sqrt(2100) = 45.825 rolls (0.458%)
	// 3-sigma range = 70.0% +/- 1.37% => [68.63%, 71.37%]
	// Allowing generous tolerance of +/- 2.0% => [68.0%, 72.0%]
	expectedRate := 0.70
	tolerance := 0.02

	if math.Abs(deflectionRate-expectedRate) > tolerance {
		t.Fatalf("FAIL: Deflection rate %.4f outside acceptable range [%.4f, %.4f]",
			deflectionRate, expectedRate-tolerance, expectedRate+tolerance)
	}

	if deflectionCount+infectionCount != totalRolls {
		t.Fatalf("FAIL: Total rolls mismatch: %d + %d != %d", deflectionCount, infectionCount, totalRolls)
	}
}

// 2. Empirical Challenge: Exact Mathematical Health Drain Mitigation of 50%
// Unarmored infected drain = 0.05 per frame
// Armored (50% defense) infected drain = 0.05 * (1.0 - 0.50) = 0.025 per frame
func TestEmpirical_ExactHealthDrainMitigation_50Percent(t *testing.T) {
	w, _, sys, pEnt := setupEmpiricalArmorHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Subtest 2a: Single-tick exact step mathematical precision
	player.Health = 100.0
	player.Hunger = 100.0
	player.Thirst = 100.0
	player.Infected = true
	player.ArmorEquipped = true
	player.ArmorDefense = 0.50

	sys.processInputAndCombat()

	expectedSingleTickHealth := 100.0 - 0.025
	if math.Abs(player.Health-expectedSingleTickHealth) > 1e-9 {
		t.Fatalf("FAIL: Single tick armored health drain mismatch: expected %f, got %f (diff: %e)",
			expectedSingleTickHealth, player.Health, player.Health-expectedSingleTickHealth)
	}

	// Subtest 2b: Multi-tick comparison against unarmored baseline over 1,000 frames
	wUnarmored, _, sysUnarmored, pEntUnarmored := setupEmpiricalArmorHarness()
	pMapUnarmored := arkecs.NewMap1[ecs.Player](wUnarmored)
	unarmoredPlayer := pMapUnarmored.Get(pEntUnarmored)
	unarmoredPlayer.Health = 100.0
	unarmoredPlayer.Hunger = 100.0
	unarmoredPlayer.Thirst = 100.0
	unarmoredPlayer.Infected = true
	unarmoredPlayer.ArmorEquipped = false
	unarmoredPlayer.ArmorDefense = 0.0

	wArmored, _, sysArmored, pEntArmored := setupEmpiricalArmorHarness()
	pMapArmored := arkecs.NewMap1[ecs.Player](wArmored)
	armoredPlayer := pMapArmored.Get(pEntArmored)
	armoredPlayer.Health = 100.0
	armoredPlayer.Hunger = 100.0
	armoredPlayer.Thirst = 100.0
	armoredPlayer.Infected = true
	armoredPlayer.ArmorEquipped = true
	armoredPlayer.ArmorDefense = 0.50

	ticks := 1000
	for tick := 1; tick <= ticks; tick++ {
		// Keep hunger and thirst saturated so starvation damage does not contaminate infection drain
		unarmoredPlayer.Hunger = 100.0
		unarmoredPlayer.Thirst = 100.0
		armoredPlayer.Hunger = 100.0
		armoredPlayer.Thirst = 100.0

		sysUnarmored.processInputAndCombat()
		sysArmored.processInputAndCombat()

		expectedUnarmoredHealth := 100.0 - (float64(tick) * 0.05)
		expectedArmoredHealth := 100.0 - (float64(tick) * 0.025)

		if math.Abs(unarmoredPlayer.Health-expectedUnarmoredHealth) > 1e-9 {
			t.Fatalf("Tick %d: Unarmored health mismatch: expected %f, got %f",
				tick, expectedUnarmoredHealth, unarmoredPlayer.Health)
		}
		if math.Abs(armoredPlayer.Health-expectedArmoredHealth) > 1e-9 {
			t.Fatalf("Tick %d: Armored health mismatch: expected %f, got %f",
				tick, expectedArmoredHealth, armoredPlayer.Health)
		}
	}

	lossUnarmored := 100.0 - unarmoredPlayer.Health
	lossArmored := 100.0 - armoredPlayer.Health
	drainRatio := lossArmored / lossUnarmored

	t.Logf("Drain Mitigation over %d frames:", ticks)
	t.Logf("  Unarmored Loss: %f (exact: %f)", lossUnarmored, float64(ticks)*0.05)
	t.Logf("  Armored Loss:   %f (exact: %f)", lossArmored, float64(ticks)*0.025)
	t.Logf("  Mitigation Factor: %.6f (expected: 0.500000)", 1.0-drainRatio)

	if math.Abs(drainRatio-0.50) > 1e-9 {
		t.Fatalf("FAIL: Drain mitigation ratio mismatch: expected exact 0.50, got %f", drainRatio)
	}
}

// 3. Empirical Challenge: Exact 10-Hit Degradation Lifecycle Until Break
// Verifies exact 1-to-1 decrement per zombie contact across all 10 hits
func TestEmpirical_Exact10HitDegradationLifecycle(t *testing.T) {
	w, _, sys, pEnt := setupEmpiricalArmorHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Equip fresh armor with max durability 10
	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDefense = 0.50
	player.ArmorDurability = 10
	player.ArmorMaxDurability = 10
	player.InfectionResist = 1.0 // 100% resist so we observe pure durability degradation without infection interference
	player.Infected = false

	// Spawn single zombie at contact range (dist = 10.0 < 14.0)
	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Step through hits 1 to 10
	for hit := 1; hit <= 10; hit++ {
		sys.processZombies()

		expectedDurability := 10 - hit
		if player.ArmorDurability != expectedDurability {
			t.Fatalf("Hit %d: Expected durability %d, got %d", hit, expectedDurability, player.ArmorDurability)
		}

		if hit < 10 {
			// Hits 1 through 9: Armor must remain equipped
			if !player.ArmorEquipped {
				t.Fatalf("Hit %d: Armor unexpectedly broken prematurely! Durability: %d", hit, player.ArmorDurability)
			}
			if player.ArmorType != "vest" {
				t.Fatalf("Hit %d: ArmorType mutated prematurely: %s", hit, player.ArmorType)
			}
			if player.ArmorDefense != 0.50 {
				t.Fatalf("Hit %d: ArmorDefense mutated prematurely: %f", hit, player.ArmorDefense)
			}
			if player.InfectionResist != 1.0 {
				t.Fatalf("Hit %d: InfectionResist mutated prematurely: %f", hit, player.InfectionResist)
			}
		} else {
			// Hit 10: Armor must break and unequip
			if player.ArmorEquipped {
				t.Fatalf("Hit 10: Armor failed to unequip at 0 durability!")
			}
		}
	}
}

// 4. Empirical Challenge: Armor State Clean Reset Upon Breaking
// Verifies all 6 armor fields cleanly reset to zero values upon breaking
func TestEmpirical_ArmorStateCleanResetUponBreak(t *testing.T) {
	w, _, sys, pEnt := setupEmpiricalArmorHarness()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Initialize armor with 1 durability remaining
	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDefense = 0.50
	player.ArmorDurability = 1
	player.ArmorMaxDurability = 10
	player.InfectionResist = 0.70
	player.Infected = false

	zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 110, Y: 100}, // Contact distance 10.0 < 14.0
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Trigger the 1-hit breakage
	sys.processZombies()

	// Verify complete clean state reset
	if player.ArmorEquipped != false {
		t.Errorf("FAIL Clean Reset: ArmorEquipped = %v, want false", player.ArmorEquipped)
	}
	if player.ArmorType != "" {
		t.Errorf("FAIL Clean Reset: ArmorType = %q, want \"\"", player.ArmorType)
	}
	if player.ArmorDefense != 0.0 {
		t.Errorf("FAIL Clean Reset: ArmorDefense = %f, want 0.0", player.ArmorDefense)
	}
	if player.ArmorDurability != 0 {
		t.Errorf("FAIL Clean Reset: ArmorDurability = %d, want 0", player.ArmorDurability)
	}
	if player.ArmorMaxDurability != 0 {
		t.Errorf("FAIL Clean Reset: ArmorMaxDurability = %d, want 0", player.ArmorMaxDurability)
	}
	if player.InfectionResist != 0.0 {
		t.Errorf("FAIL Clean Reset: InfectionResist = %f, want 0.0", player.InfectionResist)
	}

	// Step 4b: Post-break zombie contact must behave as unarmored (durability stays 0, no underflow)
	player.Infected = false // Reset infection to verify unarmored contact triggers immediate infection
	sys.processZombies()

	if player.ArmorDurability != 0 {
		t.Errorf("FAIL Underflow: Post-break contact decremented durability to %d, want 0", player.ArmorDurability)
	}
	if !player.Infected {
		t.Errorf("FAIL Post-break contact: Player without armor should be immediately infected")
	}
}

// 5. Empirical Challenge: Non-Standard Edge Cases Stress Harness
// Tests edge cases: multiple simultaneous attackers, dead player contact, stunned zombie contact
func TestEmpirical_ArmorEdgeCasesStressHarness(t *testing.T) {
	t.Run("MultiZombieContactInSingleTick", func(t *testing.T) {
		w, _, sys, pEnt := setupEmpiricalArmorHarness()
		pMap := arkecs.NewMap1[ecs.Player](w)
		player := pMap.Get(pEnt)

		player.ArmorEquipped = true
		player.ArmorType = "vest"
		player.ArmorDefense = 0.50
		player.ArmorDurability = 10
		player.ArmorMaxDurability = 10
		player.InfectionResist = 1.0

		// Spawn 3 zombies around player within contact distance
		zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: true}, &ecs.Position{X: 105, Y: 100}, &ecs.Velocity{}, &ecs.Sprite{W: 16, H: 16}, &ecs.Collider{Width: 16, Height: 16})
		zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: true}, &ecs.Position{X: 95, Y: 100}, &ecs.Velocity{}, &ecs.Sprite{W: 16, H: 16}, &ecs.Collider{Width: 16, Height: 16})
		zMap.NewEntity(&ecs.Zombie{Speed: 1.0, Chasing: true}, &ecs.Position{X: 100, Y: 105}, &ecs.Velocity{}, &ecs.Sprite{W: 16, H: 16}, &ecs.Collider{Width: 16, Height: 16})

		sys.processZombies()

		// 3 zombies touching player in one tick = 3 durability lost
		if player.ArmorDurability != 7 {
			t.Errorf("Expected durability 7 after 3 simultaneous hits, got %d", player.ArmorDurability)
		}
	})

	t.Run("StunnedZombieDoesNotDamageArmor", func(t *testing.T) {
		w, _, sys, pEnt := setupEmpiricalArmorHarness()
		pMap := arkecs.NewMap1[ecs.Player](w)
		player := pMap.Get(pEnt)

		player.ArmorEquipped = true
		player.ArmorType = "vest"
		player.ArmorDefense = 0.50
		player.ArmorDurability = 10
		player.ArmorMaxDurability = 10

		// Spawn stunned zombie at contact distance
		zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: true, StunTimer: 30},
			&ecs.Position{X: 105, Y: 100},
			&ecs.Velocity{},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		sys.processZombies()

		// Stunned zombie should skip AI and attack logic
		if player.ArmorDurability != 10 {
			t.Errorf("Expected durability 10 (stunned zombie cannot attack), got %d", player.ArmorDurability)
		}
	})

	t.Run("DeadPlayerArmorDoesNotDegrade", func(t *testing.T) {
		w, _, sys, pEnt := setupEmpiricalArmorHarness()
		pMap := arkecs.NewMap1[ecs.Player](w)
		player := pMap.Get(pEnt)

		player.Dead = true
		player.ArmorEquipped = true
		player.ArmorType = "vest"
		player.ArmorDefense = 0.50
		player.ArmorDurability = 10
		player.ArmorMaxDurability = 10

		zMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
		zMap.NewEntity(
			&ecs.Zombie{Speed: 1.0, Chasing: true},
			&ecs.Position{X: 105, Y: 100},
			&ecs.Velocity{},
			&ecs.Sprite{W: 16, H: 16},
			&ecs.Collider{Width: 16, Height: 16},
		)

		sys.processZombies()

		if player.ArmorDurability != 10 {
			t.Errorf("Expected durability 10 (dead player cannot take armor degradation), got %d", player.ArmorDurability)
		}
	})
}
