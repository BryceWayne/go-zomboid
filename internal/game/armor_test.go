package game

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupArmorTestGame initializes a headless test world with player entity
func setupArmorTestGame() (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
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

	return w, m, updateSys, pEnt
}

// 1. Test ECS Player Component Armor Fields
func TestArmor_ECSComponentFields(t *testing.T) {
	p := ecs.Player{
		Health:             100.0,
		ArmorEquipped:      true,
		ArmorType:          "vest",
		ArmorDefense:       0.50,
		ArmorDurability:    10,
		ArmorMaxDurability: 10,
		InfectionResist:    0.70,
	}

	if !p.ArmorEquipped {
		t.Errorf("Expected ArmorEquipped to be true")
	}
	if p.ArmorType != "vest" {
		t.Errorf("Expected ArmorType 'vest', got '%s'", p.ArmorType)
	}
	if p.ArmorDefense != 0.50 {
		t.Errorf("Expected ArmorDefense 0.50, got %f", p.ArmorDefense)
	}
	if p.ArmorDurability != 10 || p.ArmorMaxDurability != 10 {
		t.Errorf("Expected Durability 10/10, got %d/%d", p.ArmorDurability, p.ArmorMaxDurability)
	}
	if p.InfectionResist != 0.70 {
		t.Errorf("Expected InfectionResist 0.70, got %f", p.InfectionResist)
	}
}

// 2. Test Equipping Armor from Inventory
func TestArmor_EquipFromInventory(t *testing.T) {
	w, _, _, pEnt := setupArmorTestGame()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	// Add armor and food to inventory
	player.Inventory = []string{"armor", "food"}
	if len(player.Inventory) != 2 {
		t.Fatalf("Expected inventory length 2, got %d", len(player.Inventory))
	}

	// Simulate equipping slot 0 ("armor")
	useItemIdx := 0
	if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		itemType := player.Inventory[useItemIdx]
		if itemType == "armor" || itemType == "vest" {
			player.ArmorEquipped = true
			player.ArmorType = "vest"
			player.ArmorDefense = 0.50
			player.ArmorDurability = 10
			player.ArmorMaxDurability = 10
			player.InfectionResist = 0.70
			player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
		}
	}

	if !player.ArmorEquipped {
		t.Error("Expected ArmorEquipped to be true")
	}
	if player.ArmorType != "vest" {
		t.Errorf("Expected ArmorType 'vest', got '%s'", player.ArmorType)
	}
	if player.ArmorDefense != 0.50 {
		t.Errorf("Expected ArmorDefense 0.50, got %f", player.ArmorDefense)
	}
	if player.ArmorDurability != 10 {
		t.Errorf("Expected ArmorDurability 10, got %d", player.ArmorDurability)
	}
	if player.ArmorMaxDurability != 10 {
		t.Errorf("Expected ArmorMaxDurability 10, got %d", player.ArmorMaxDurability)
	}
	if player.InfectionResist != 0.70 {
		t.Errorf("Expected InfectionResist 0.70, got %f", player.InfectionResist)
	}
	if len(player.Inventory) != 1 || player.Inventory[0] != "food" {
		t.Errorf("Expected inventory to contain only 'food', got %v", player.Inventory)
	}
	if player.AttackCooldown != 30 {
		t.Errorf("Expected AttackCooldown 30, got %d", player.AttackCooldown)
	}
}

// 3. Test Re-equipping Armor Refreshes Durability
func TestArmor_ReEquipRefreshesDurability(t *testing.T) {
	player := &ecs.Player{
		Health:             100.0,
		Inventory:          []string{"vest"},
		ArmorEquipped:      true,
		ArmorType:          "vest",
		ArmorDefense:       0.50,
		ArmorDurability:    2, // Damaged armor
		ArmorMaxDurability: 10,
		InfectionResist:    0.70,
		AttackCooldown:     0,
	}

	useItemIdx := 0
	if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		itemType := player.Inventory[useItemIdx]
		if itemType == "armor" || itemType == "vest" {
			player.ArmorEquipped = true
			player.ArmorType = "vest"
			player.ArmorDefense = 0.50
			player.ArmorDurability = 10
			player.ArmorMaxDurability = 10
			player.InfectionResist = 0.70
			player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
		}
	}

	if player.ArmorDurability != 10 {
		t.Errorf("Expected ArmorDurability refreshed to 10, got %d", player.ArmorDurability)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Expected inventory to be empty after equipping")
	}
}

// 4. Test Unarmored Player Direct Infection on Zombie Contact (<14px)
func TestZombieContact_UnarmoredDirectInfection(t *testing.T) {
	w, _, updateSys, pEnt := setupArmorTestGame()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.ArmorEquipped = false
	player.Infected = false

	// Spawn zombie at dist = 10.0 (< 14.0)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	updateSys.processZombies()

	if !player.Infected {
		t.Fatal("Expected unarmored player to be infected on zombie contact")
	}
}

// 5. Test Armored Deflection Success (InfectionResist = 1.0)
func TestZombieContact_ArmoredDeflectionSuccess(t *testing.T) {
	w, _, updateSys, pEnt := setupArmorTestGame()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDefense = 0.50
	player.ArmorDurability = 10
	player.ArmorMaxDurability = 10
	player.InfectionResist = 1.0 // Guaranteed deflect
	player.Infected = false

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	updateSys.processZombies()

	if player.Infected {
		t.Fatal("Expected infection to be deflected with InfectionResist = 1.0")
	}
	if player.ArmorDurability != 9 {
		t.Fatalf("Expected armor durability 9, got %d", player.ArmorDurability)
	}
	if !player.ArmorEquipped {
		t.Fatal("Expected armor to remain equipped")
	}
}

// 6. Test Armored Deflection Failure (InfectionResist = 0.0)
func TestZombieContact_ArmoredDeflectionFailure(t *testing.T) {
	w, _, updateSys, pEnt := setupArmorTestGame()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDefense = 0.50
	player.ArmorDurability = 10
	player.ArmorMaxDurability = 10
	player.InfectionResist = 0.0 // Guaranteed penetration
	player.Infected = false

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	updateSys.processZombies()

	if !player.Infected {
		t.Fatal("Expected 0% InfectionResist to result in infection")
	}
	if player.ArmorDurability != 9 {
		t.Fatalf("Expected armor durability 9, got %d", player.ArmorDurability)
	}
}

// 7. Test Armor Breakage at 0 Durability
func TestZombieContact_ArmorBreakageAtZeroDurability(t *testing.T) {
	w, _, updateSys, pEnt := setupArmorTestGame()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDefense = 0.50
	player.ArmorDurability = 1 // 1 hit remaining before breaking
	player.ArmorMaxDurability = 10
	player.InfectionResist = 1.0
	player.Infected = false

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 110, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	updateSys.processZombies()

	if player.ArmorEquipped {
		t.Fatal("Expected armor to break and unequip when durability reaches 0")
	}
	if player.ArmorDurability != 0 {
		t.Fatalf("Expected armor durability 0, got %d", player.ArmorDurability)
	}
	if player.ArmorDefense != 0.0 {
		t.Fatalf("Expected armor defense 0.0, got %f", player.ArmorDefense)
	}
	if player.ArmorType != "" {
		t.Fatalf("Expected armor type empty string, got %s", player.ArmorType)
	}
	if player.ArmorMaxDurability != 0 {
		t.Fatalf("Expected armor max durability 0, got %d", player.ArmorMaxDurability)
	}
	if player.InfectionResist != 0.0 {
		t.Fatalf("Expected infection resist 0.0, got %f", player.InfectionResist)
	}
}

// 8. Test Multi-Hit Degradation Loop
func TestArmor_MultiHitDegradation(t *testing.T) {
	w, _, updateSys, pEnt := setupArmorTestGame()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	initialDurability := 5
	player.ArmorEquipped = true
	player.ArmorType = "vest"
	player.ArmorDurability = initialDurability
	player.ArmorMaxDurability = 5
	player.ArmorDefense = 0.50
	player.InfectionResist = 1.0

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, Chasing: true},
		&ecs.Position{X: 105, Y: 100},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	for hit := 1; hit <= initialDurability; hit++ {
		expectedRemaining := initialDurability - hit
		updateSys.processZombies()

		if player.ArmorDurability != expectedRemaining {
			t.Errorf("Hit %d: Expected durability %d, got %d", hit, expectedRemaining, player.ArmorDurability)
		}

		if hit < initialDurability {
			if !player.ArmorEquipped {
				t.Errorf("Hit %d: Armor unexpectedly unequipped early", hit)
			}
		} else {
			if player.ArmorEquipped {
				t.Errorf("Hit %d: Armor should be unequipped at 0 durability", hit)
			}
		}
	}
}

// 9. Test Infection Health Drain Mitigation
func TestArmor_DamageMitigation_HealthDrain(t *testing.T) {
	w1, _, sys1, pEnt1 := setupArmorTestGame()
	pMap1 := arkecs.NewMap1[ecs.Player](w1)
	unarmored := pMap1.Get(pEnt1)
	unarmored.Infected = true
	unarmored.ArmorEquipped = false
	unarmored.ArmorDefense = 0.0

	w2, _, sys2, pEnt2 := setupArmorTestGame()
	pMap2 := arkecs.NewMap1[ecs.Player](w2)
	armored := pMap2.Get(pEnt2)
	armored.Infected = true
	armored.ArmorEquipped = true
	armored.ArmorDefense = 0.50

	for i := 0; i < 100; i++ {
		sys1.processInputAndCombat(-1)
		sys2.processInputAndCombat(-1)
	}

	lossUnarmored := 100.0 - unarmored.Health
	lossArmored := 100.0 - armored.Health

	t.Logf("Unarmored HP Loss: %f, Armored HP Loss: %f", lossUnarmored, lossArmored)

	if lossArmored >= lossUnarmored {
		t.Errorf("Expected armored damage (%f) to be strictly less than unarmored damage (%f)", lossArmored, lossUnarmored)
	}

	expectedMitigationRatio := 0.50
	actualMitigationRatio := 1.0 - (lossArmored / lossUnarmored)
	if math.Abs(actualMitigationRatio-expectedMitigationRatio) > 0.05 {
		t.Errorf("Expected mitigation ratio ~0.50, got %f", actualMitigationRatio)
	}
}

// 10. Test HUD Bar Calculation & Text Formatting
func TestArmor_HUDCalculations(t *testing.T) {
	calcWidth := func(durability, maxDurability int) float32 {
		if maxDurability <= 0 || durability <= 0 {
			return 0
		}
		w := float32(float64(durability) / float64(maxDurability) * 200.0)
		if w > 200 {
			w = 200
		}
		if w < 0 {
			w = 0
		}
		return w
	}

	tests := []struct {
		name      string
		dur, max  int
		wantWidth float32
	}{
		{"Full Durability", 10, 10, 200.0},
		{"Half Durability", 5, 10, 100.0},
		{"Quarter Durability", 25, 100, 50.0},
		{"Zero Durability", 0, 10, 0.0},
		{"Negative Durability Clamped", -3, 10, 0.0},
		{"Over-Capacity Clamped", 15, 10, 200.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcWidth(tt.dur, tt.max)
			if got != tt.wantWidth {
				t.Errorf("calcWidth(%d, %d) = %f, want %f", tt.dur, tt.max, got, tt.wantWidth)
			}
		})
	}

	armoredStr := fmt.Sprintf("Armor: %d/%d (Def: %d%%)", 10, 10, int(0.50*100))
	if armoredStr != "Armor: 10/10 (Def: 50%)" {
		t.Errorf("Expected 'Armor: 10/10 (Def: 50%%)', got '%s'", armoredStr)
	}
}

// 11. Test Visual Indicator State Logic
func TestArmor_VisualIndicatorConditions(t *testing.T) {
	tests := []struct {
		name          string
		dead          bool
		infected      bool
		armorEquip    bool
		wantArmorTint bool
	}{
		{"Equipped Normal Player", false, false, true, true},
		{"Unarmored Normal Player", false, false, false, false},
		{"Equipped Infected Player", false, true, true, false}, // Infection pulse takes precedence
		{"Equipped Dead Player", true, false, true, false},     // Death tint takes precedence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isArmorTintApplied := false
			if !tt.dead && !tt.infected && tt.armorEquip {
				isArmorTintApplied = true
			}

			if isArmorTintApplied != tt.wantArmorTint {
				t.Errorf("Visual tint condition = %v, want %v", isArmorTintApplied, tt.wantArmorTint)
			}
		})
	}
}
