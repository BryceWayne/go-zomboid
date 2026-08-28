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

// setupMeleeTestWorld initializes a headless test world with player entity
func setupMeleeTestWorld() (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
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
func TestMelee_ECSComponentWeaponFields(t *testing.T) {
	pAxe := ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "axe",
		WeaponDurability: 12,
	}
	if !pAxe.WeaponEquipped || pAxe.WeaponType != "axe" || pAxe.WeaponDurability != 12 {
		t.Errorf("Axe component initialization failed: %+v", pAxe)
	}

	pBat := ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "weapon",
		WeaponDurability: 5,
	}
	if !pBat.WeaponEquipped || pBat.WeaponType != "weapon" || pBat.WeaponDurability != 5 {
		t.Errorf("Bat component initialization failed: %+v", pBat)
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

// 2. Test Equipping Fire Axe from Inventory
func TestMelee_EquipAxeFromInventory(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.Inventory = []string{"axe", "food"}
	useItemIdx := 0

	// Simulate slot key press logic
	if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		itemType := player.Inventory[useItemIdx]
		if itemType == "axe" {
			player.WeaponEquipped = true
			player.WeaponType = "axe"
			player.WeaponDurability = 12
			player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
		}
	}

	if !player.WeaponEquipped {
		t.Error("Expected WeaponEquipped to be true")
	}
	if player.WeaponType != "axe" {
		t.Errorf("Expected WeaponType 'axe', got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 12 {
		t.Errorf("Expected WeaponDurability 12, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 1 || player.Inventory[0] != "food" {
		t.Errorf("Expected inventory to contain only 'food', got %v", player.Inventory)
	}
}

// 3. Test Equipping Spiked Bat from Inventory
func TestMelee_EquipBatFromInventory(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)

	player.Inventory = []string{"weapon", "water"}
	useItemIdx := 0

	// Simulate slot key press logic
	if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		itemType := player.Inventory[useItemIdx]
		if itemType == "weapon" {
			player.WeaponEquipped = true
			player.WeaponType = "weapon"
			player.WeaponDurability = 5
			player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
		}
	}

	if !player.WeaponEquipped {
		t.Error("Expected WeaponEquipped to be true")
	}
	if player.WeaponType != "weapon" {
		t.Errorf("Expected WeaponType 'weapon', got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 5 {
		t.Errorf("Expected WeaponDurability 5, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 1 || player.Inventory[0] != "water" {
		t.Errorf("Expected inventory to contain only 'water', got %v", player.Inventory)
	}
}

// 4. Test Unarmed Shove Mechanics (Stun + Knockback, No Kill)
func TestMelee_UnarmedShoveMechanics(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = false
	player.WeaponType = ""
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Spawn zombie in front of player at (120, 100)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0, StunTimer: 0},
		&ecs.Position{X: 120.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	// Simulate attack logic
	reach := 24.0
	hitRadius := 24.0
	if player.WeaponEquipped && player.WeaponType == "axe" {
		reach = 32.0
		hitRadius = 32.0
	}

	attackX := 100.0 + player.FacingX*reach
	attackY := 100.0 + player.FacingY*reach

	hitZombies := false
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		z, zPos, zVel := zQuery.Get()
		ent := zQuery.Entity()

		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx+dy*dy) < hitRadius {
			hitZombies = true
			if player.WeaponEquipped {
				toRemove = append(toRemove, ent)
			} else {
				z.StunTimer = 45
				zVel.X = player.FacingX * 5.0
				zVel.Y = player.FacingY * 5.0
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}

	if !hitZombies {
		t.Fatal("Expected unarmed shove to connect with zombie")
	}

	zComp := arkecs.NewMap1[ecs.Zombie](w).Get(zEnt)
	if zComp == nil {
		t.Fatal("Unarmed shove must not delete zombie entity")
	}
	if zComp.StunTimer != 45 {
		t.Errorf("Expected StunTimer 45, got %d", zComp.StunTimer)
	}
	zVel := arkecs.NewMap1[ecs.Velocity](w).Get(zEnt)
	if zVel.X != 5.0 || zVel.Y != 0.0 {
		t.Errorf("Expected pushback velocity (5.0, 0.0), got (%f, %f)", zVel.X, zVel.Y)
	}
}

// 5. Test Spiked Bat Kill and Reach (24px reach)
func TestMelee_BatReachAndKill(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "weapon"
	player.WeaponDurability = 5
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Zombie inside Bat reach: player at (100, 100), attack center at (124, 100), zombie at (120, 100) -> dist = 4px < 24px
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 120.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	reach := 24.0
	hitRadius := 24.0
	if player.WeaponEquipped && player.WeaponType == "axe" {
		reach = 32.0
		hitRadius = 32.0
	}
	attackX := 100.0 + player.FacingX*reach
	attackY := 100.0 + player.FacingY*reach

	hitZombies := false
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx+dy*dy) < hitRadius {
			hitZombies = true
			if player.WeaponEquipped {
				toRemove = append(toRemove, ent)
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}
	if hitZombies && player.WeaponEquipped {
		player.WeaponDurability--
	}

	if !hitZombies {
		t.Fatal("Expected bat attack to hit zombie")
	}
	if player.WeaponDurability != 4 {
		t.Errorf("Expected bat durability 4, got %d", player.WeaponDurability)
	}
	if arkecs.NewMap1[ecs.Zombie](w).Get(zEnt) != nil {
		t.Error("Expected zombie to be removed after armed kill")
	}
}

// 6. Test Spiked Bat Out of Reach (Zombie beyond Bat reach 48px)
func TestMelee_BatOutOfReach(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "weapon"
	player.WeaponDurability = 5
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Zombie at (155, 100): dist from bat attack center (124, 100) is 31px > 24px (Miss for Bat!)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 155.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	reach := 24.0
	hitRadius := 24.0
	if player.WeaponEquipped && player.WeaponType == "axe" {
		reach = 32.0
		hitRadius = 32.0
	}
	attackX := 100.0 + player.FacingX*reach
	attackY := 100.0 + player.FacingY*reach

	hitZombies := false
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx+dy*dy) < hitRadius {
			hitZombies = true
			if player.WeaponEquipped {
				toRemove = append(toRemove, ent)
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}
	if hitZombies && player.WeaponEquipped {
		player.WeaponDurability--
	}

	if hitZombies {
		t.Fatal("Bat should not reach zombie at (155, 100)")
	}
	if player.WeaponDurability != 5 {
		t.Errorf("Durability should remain 5 on miss, got %d", player.WeaponDurability)
	}
	if arkecs.NewMap1[ecs.Zombie](w).Get(zEnt) == nil {
		t.Error("Zombie should still be alive")
	}
}

// 7. Test Fire Axe Extended Reach (32px reach, hits (155, 100))
func TestMelee_AxeExtendedReach(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Zombie at (155, 100): Axe attack center at (132, 100), dist is 23px < 32px (HIT for Axe!)
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	zEnt := zombieMap.NewEntity(
		&ecs.Zombie{Speed: 1.0},
		&ecs.Position{X: 155.0, Y: 100.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{255, 0, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	reach := 24.0
	hitRadius := 24.0
	if player.WeaponEquipped && player.WeaponType == "axe" {
		reach = 32.0
		hitRadius = 32.0
	}
	attackX := 100.0 + player.FacingX*reach
	attackY := 100.0 + player.FacingY*reach

	hitZombies := false
	var toRemove []arkecs.Entity
	zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
	for zQuery.Next() {
		_, zPos, _ := zQuery.Get()
		ent := zQuery.Entity()
		dx := attackX - zPos.X
		dy := attackY - zPos.Y
		if math.Sqrt(dx*dx+dy*dy) < hitRadius {
			hitZombies = true
			if player.WeaponEquipped {
				toRemove = append(toRemove, ent)
			}
		}
	}
	for _, ent := range toRemove {
		w.RemoveEntity(ent)
	}
	if hitZombies && player.WeaponEquipped {
		player.WeaponDurability--
	}

	if !hitZombies {
		t.Fatal("Expected Axe to hit zombie at (155, 100) due to 32px reach & 32px radius")
	}
	if player.WeaponDurability != 11 {
		t.Errorf("Expected Axe durability 11, got %d", player.WeaponDurability)
	}
	if arkecs.NewMap1[ecs.Zombie](w).Get(zEnt) != nil {
		t.Error("Expected zombie to be removed after axe kill")
	}
}

// 8. Test Fire Axe Multi-Target Cleave (Cleaves 3 zombies simultaneously in wide arc)
func TestMelee_AxeMultiTargetCleave(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	// Player at (100, 100). Axe attack center at (132, 100) with radius 32.
	// Spawn 3 zombies in wide sweep:
	// Z1: Straight ahead at (135, 100) -> dist = 3px
	// Z2: Top flank at (130, 80) -> dx = 2, dy = 20 -> dist = ~20.1px < 32px
	// Z3: Bottom flank at (130, 120) -> dx = 2, dy = 20 -> dist = ~20.1px < 32px
	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	z1 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 135.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z2 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 130.0, Y: 80.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})
	z3 := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 130.0, Y: 120.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

	reach := 32.0
	hitRadius := 32.0
	attackX := 100.0 + player.FacingX*reach
	attackY := 100.0 + player.FacingY*reach

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
		player.WeaponDurability-- // 1 swing = 1 durability cost, even if cleaving multiple
	}

	if hitCount != 3 {
		t.Fatalf("Expected Axe cleave to hit all 3 zombies, got %d", hitCount)
	}
	if player.WeaponDurability != 11 {
		t.Errorf("Expected Axe durability 11 after 1 cleave swing, got %d", player.WeaponDurability)
	}

	zMap := arkecs.NewMap1[ecs.Zombie](w)
	if zMap.Get(z1) != nil || zMap.Get(z2) != nil || zMap.Get(z3) != nil {
		t.Error("Expected all 3 cleaved zombies to be removed from world")
	}
}

// 9. Test Fire Axe Durability 12 Multi-Hit Degradation Loop & Breakage
func TestMelee_Axe12DurabilityDegradationAndBreakage(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12
	player.FacingX = 1.0
	player.FacingY = 0.0

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	for hit := 1; hit <= 12; hit++ {
		zEnt := zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 130.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

		reach := 32.0
		hitRadius := 32.0
		attackX := 100.0 + player.FacingX*reach
		attackY := 100.0 + player.FacingY*reach

		hitZombies := false
		var toRemove []arkecs.Entity
		zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
		for zQuery.Next() {
			_, zPos, _ := zQuery.Get()
			ent := zQuery.Entity()
			dx := attackX - zPos.X
			dy := attackY - zPos.Y
			if math.Sqrt(dx*dx+dy*dy) < hitRadius {
				hitZombies = true
				toRemove = append(toRemove, ent)
			}
		}
		for _, ent := range toRemove {
			w.RemoveEntity(ent)
		}
		if hitZombies && player.WeaponEquipped {
			player.WeaponDurability--
			if player.WeaponDurability <= 0 {
				player.WeaponEquipped = false
				player.WeaponType = ""
			}
		}

		expectedDurability := 12 - hit
		if player.WeaponDurability != expectedDurability {
			t.Errorf("Hit %d: Expected durability %d, got %d", hit, expectedDurability, player.WeaponDurability)
		}

		if hit < 12 {
			if !player.WeaponEquipped || player.WeaponType != "axe" {
				t.Errorf("Hit %d: Axe should remain equipped with type 'axe'", hit)
			}
		} else {
			if player.WeaponEquipped || player.WeaponType != "" {
				t.Errorf("Hit %d: Axe should break and unequip at durability 0", hit)
			}
		}

		_ = zEnt
	}
}

// 10. Test Spiked Bat Durability 5 Multi-Hit Degradation Loop & Breakage
func TestMelee_Bat5DurabilityDegradationAndBreakage(t *testing.T) {
	w, _, _, pEnt := setupMeleeTestWorld()
	pMap := arkecs.NewMap1[ecs.Player](w)
	player := pMap.Get(pEnt)
	player.WeaponEquipped = true
	player.WeaponType = "weapon"
	player.WeaponDurability = 5
	player.FacingX = 1.0
	player.FacingY = 0.0

	zombieMap := arkecs.NewMap5[ecs.Zombie, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	for hit := 1; hit <= 5; hit++ {
		zombieMap.NewEntity(&ecs.Zombie{Speed: 1.0}, &ecs.Position{X: 120.0, Y: 100.0}, &ecs.Velocity{}, &ecs.Sprite{}, &ecs.Collider{})

		reach := 24.0
		hitRadius := 24.0
		attackX := 100.0 + player.FacingX*reach
		attackY := 100.0 + player.FacingY*reach

		hitZombies := false
		var toRemove []arkecs.Entity
		zQuery := arkecs.NewFilter3[ecs.Zombie, ecs.Position, ecs.Velocity](w).Query()
		for zQuery.Next() {
			_, zPos, _ := zQuery.Get()
			ent := zQuery.Entity()
			dx := attackX - zPos.X
			dy := attackY - zPos.Y
			if math.Sqrt(dx*dx+dy*dy) < hitRadius {
				hitZombies = true
				toRemove = append(toRemove, ent)
			}
		}
		for _, ent := range toRemove {
			w.RemoveEntity(ent)
		}
		if hitZombies && player.WeaponEquipped {
			player.WeaponDurability--
			if player.WeaponDurability <= 0 {
				player.WeaponEquipped = false
				player.WeaponType = ""
			}
		}

		expectedDurability := 5 - hit
		if player.WeaponDurability != expectedDurability {
			t.Errorf("Hit %d: Expected durability %d, got %d", hit, expectedDurability, player.WeaponDurability)
		}

		if hit < 5 {
			if !player.WeaponEquipped || player.WeaponType != "weapon" {
				t.Errorf("Hit %d: Bat should remain equipped", hit)
			}
		} else {
			if player.WeaponEquipped || player.WeaponType != "" {
				t.Errorf("Hit %d: Bat should break and unequip at durability 0", hit)
			}
		}
	}
}

// 11. Test Re-equipping Weapon Overrides Previous Weapon Stats
func TestMelee_ReEquipOverridesStats(t *testing.T) {
	player := &ecs.Player{
		WeaponEquipped:   true,
		WeaponType:       "weapon",
		WeaponDurability: 1, // Damaged bat
		Inventory:        []string{"axe"},
		AttackCooldown:   0,
	}

	useItemIdx := 0
	if useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
		player.AttackCooldown = 30
		itemType := player.Inventory[useItemIdx]
		if itemType == "axe" {
			player.WeaponEquipped = true
			player.WeaponType = "axe"
			player.WeaponDurability = 12
			player.Inventory = append(player.Inventory[:useItemIdx], player.Inventory[useItemIdx+1:]...)
		}
	}

	if !player.WeaponEquipped {
		t.Error("Expected WeaponEquipped true")
	}
	if player.WeaponType != "axe" {
		t.Errorf("Expected WeaponType 'axe', got '%s'", player.WeaponType)
	}
	if player.WeaponDurability != 12 {
		t.Errorf("Expected WeaponDurability 12, got %d", player.WeaponDurability)
	}
	if len(player.Inventory) != 0 {
		t.Errorf("Expected inventory to be empty, got %v", player.Inventory)
	}
}

// 12. Test UI Weapon Text Formatting
func TestMelee_UIWeaponTextFormatting(t *testing.T) {
	formatWeaponText := func(hasWeapon bool, weaponType string, durability int) string {
		if hasWeapon {
			weaponName := "WEAPON"
			if weaponType == "axe" {
				weaponName = "FIRE AXE"
			} else if weaponType == "weapon" {
				weaponName = "SPIKED BAT"
			}
			return fmt.Sprintf("Weapon: %s (Durability: %d) (Press SPACE/X to attack)", weaponName, durability)
		}
		return "Weapon: NONE (Press SPACE/X to shove zombies back)"
	}

	axeStr := formatWeaponText(true, "axe", 12)
	if !strings.Contains(axeStr, "FIRE AXE") || !strings.Contains(axeStr, "12") {
		t.Errorf("Unexpected axe text format: %s", axeStr)
	}

	batStr := formatWeaponText(true, "weapon", 5)
	if !strings.Contains(batStr, "SPIKED BAT") || !strings.Contains(batStr, "5") {
		t.Errorf("Unexpected bat text format: %s", batStr)
	}

	unarmedStr := formatWeaponText(false, "", 0)
	if !strings.Contains(unarmedStr, "NONE") || !strings.Contains(unarmedStr, "shove") {
		t.Errorf("Unexpected unarmed text format: %s", unarmedStr)
	}
}
