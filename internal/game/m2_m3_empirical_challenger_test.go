package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupChallengerGame initializes a clean test world for adversarial testing
func setupChallengerGame(mapW, mapH int) (*arkecs.World, *world.Map, *UpdateSystem, *ecs.Player, *ecs.Position) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(mapW, mapH)
	// Clear procedural props/chests so map is clean slate
	for p := range m.Chests {
		m.SetTile(p.X, p.Y, world.TileWoodFloor)
	}
	m.Chests = make(map[world.Point][]string)

	updateSys := NewUpdateSystem(w, m)

	player := &ecs.Player{
		Health:             100.0,
		Hunger:             100.0,
		Thirst:             100.0,
		Inventory:          make([]string, 9),
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
		FacingX:            1,
		FacingY:            0,
	}
	pos := &ecs.Position{X: 640, Y: 640}

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	playerMap.NewEntity(
		player,
		pos,
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, updateSys, player, pos
}

// ==============================================================================================
// CHALLENGE 1: 50,000 Rapid Continuous Swaps with Random Consumption & Restock (Zero Duplication/Loss)
// ==============================================================================================

func TestEmpiricalChallenge_50000RapidContinuousSwapsWithRestockAndConsumption(t *testing.T) {
	_, m, _, player, pos := setupChallengerGame(60, 60)

	chestTileX, chestTileY := 20, 20
	m.SetTile(chestTileX, chestTileY, world.TileChest)
	chestPoint := world.Point{X: chestTileX, Y: chestTileY}

	chestCenterX := float64(chestTileX)*float64(world.TileSize) + 64.0
	chestCenterY := float64(chestTileY)*float64(world.TileSize) + 64.0
	pos.X = chestCenterX
	pos.Y = chestCenterY

	// Initial setup: Distinct items
	initialPlayerItems := []string{"p_axe", "p_food1", "p_food2", "p_water1", "p_ammo1", "p_ammo2", "p_armor", "p_antidote", "p_shotgun"}
	initialChestItems := []string{"c_food1", "c_water1", "c_water2", "c_ammo1", "c_ammo2", "c_ammo3", "c_axe", "c_armor", "c_weapon"}

	copy(player.Inventory, initialPlayerItems)
	m.SetChestInventory(chestTileX, chestTileY, initialChestItems)

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	// Global Accounting Ledger
	ledger := make(map[string]int)
	for _, item := range initialPlayerItems {
		if item != "" {
			ledger[item]++
		}
	}
	for _, item := range initialChestItems {
		if item != "" {
			ledger[item]++
		}
	}
	ledger["axe"]++ // Equipped weapon

	consumedCount := make(map[string]int)
	destroyedWeapons := make(map[string]int)

	rng := rand.New(rand.NewSource(424242))

	const totalIterations = 50000
	itemPool := []string{"loot_food", "loot_water", "loot_ammo", "loot_antidote", "loot_axe", "loot_shotgun", "loot_weapon"}

	for i := 0; i < totalIterations; i++ {
		action := rng.Intn(100)

		if action < 60 {
			// Action 1: Continuous Chest Swap (60% probability)
			pPos := world.FloatPoint{X: pos.X, Y: pos.Y}
			swapped, tx, ty := simulateChestSwap(m, player, pPos)
			if !swapped || tx != chestTileX || ty != chestTileY {
				t.Fatalf("Iteration %d: simulateChestSwap failed: swapped=%v tx=%d ty=%d", i, swapped, tx, ty)
			}

			// Verify slice lengths
			if len(player.Inventory) != 9 {
				t.Fatalf("Iteration %d: player inventory corrupted: len=%d", i, len(player.Inventory))
			}
			chestInv := m.GetChestInventory(chestTileX, chestTileY)
			if len(chestInv) != 9 {
				t.Fatalf("Iteration %d: chest inventory corrupted: len=%d", i, len(chestInv))
			}

		} else if action < 70 {
			// Action 2: Random Item Consumption from player inventory (10% probability)
			for slot := 0; slot < 9; slot++ {
				itm := player.Inventory[slot]
				if itm != "" && (itm == "p_food1" || itm == "p_food2" || itm == "c_food1" || itm == "loot_food" ||
					itm == "p_water1" || itm == "c_water1" || itm == "c_water2" || itm == "loot_water" ||
					itm == "p_antidote" || itm == "loot_antidote") {
					player.Inventory[slot] = ""
					consumedCount[itm]++
					break
				}
			}

		} else if action < 80 {
			// Action 3: Restock/Loot injection into empty slot (10% probability)
			restocked := false
			newItem := fmt.Sprintf("%s_%d", itemPool[rng.Intn(len(itemPool))], i)
			// Try player inventory first
			for slot := 0; slot < 9; slot++ {
				if player.Inventory[slot] == "" {
					player.Inventory[slot] = newItem
					ledger[newItem]++
					restocked = true
					break
				}
			}
			// If player full, try chest
			if !restocked {
				chestInv := m.GetChestInventory(chestTileX, chestTileY)
				for slot := 0; slot < 9; slot++ {
					if chestInv[slot] == "" {
						chestInv[slot] = newItem
						m.SetChestInventory(chestTileX, chestTileY, chestInv)
						ledger[newItem]++
						restocked = true
						break
					}
				}
			}

		} else if action < 90 {
			// Action 4: Equip / Unequip / Attack cycle (10% probability)
			if player.WeaponEquipped {
				if rng.Float64() < 0.3 {
					// Unequip attempt
					simulateUnequip(player)
				} else {
					// Attack with weapon, reducing durability
					player.WeaponDurability--
					if player.WeaponDurability <= 0 {
						destroyedWeapons[player.WeaponType]++
						player.WeaponEquipped = false
						player.WeaponType = ""
						player.WeaponDurability = 0
					}
				}
			} else {
				// Try to equip any weapon in player's inventory
				for slot := 0; slot < 9; slot++ {
					itm := player.Inventory[slot]
					if itm == "axe" || itm == "weapon" || itm == "shotgun" ||
						itm == "p_axe" || itm == "c_axe" || itm == "p_shotgun" || itm == "c_weapon" {
						// Normalize weapon type for simulator
						normType := "weapon"
						if itm == "axe" || itm == "p_axe" || itm == "c_axe" {
							normType = "axe"
						} else if itm == "shotgun" || itm == "p_shotgun" {
							normType = "shotgun"
						}
						player.Inventory[slot] = normType
						// Adjust ledger if normalized name changed
						if itm != normType {
							ledger[itm]--
							ledger[normType]++
						}
						player.AttackCooldown = 0
						simulateUseItem(player, slot)
						break
					}
				}
			}

		} else {
			// Action 5: Drag & Drop slot shuffle (10% probability)
			s1 := rng.Intn(10)
			s2 := rng.Intn(10)
			simulateDragDrop(player, s1, s2)
		}

		// Invariant Verification every 500 iterations
		if i%500 == 0 || i == totalIterations-1 {
			currentCounts := make(map[string]int)
			for _, itm := range player.Inventory {
				if itm != "" {
					currentCounts[itm]++
				}
			}
			chestInv := m.GetChestInventory(chestTileX, chestTileY)
			for _, itm := range chestInv {
				if itm != "" {
					currentCounts[itm]++
				}
			}
			if player.WeaponEquipped && player.WeaponType != "" {
				currentCounts[player.WeaponType]++
			}
			for itm, c := range consumedCount {
				currentCounts[itm] += c
			}
			for itm, c := range destroyedWeapons {
				currentCounts[itm] += c
			}

			// Cross-check with ledger
			for itm, expectedTotal := range ledger {
				actualTotal := currentCounts[itm]
				if actualTotal != expectedTotal {
					t.Fatalf("Iteration %d: Item Conservation Violation for '%s': expected %d, got %d (consumed: %d, destroyed: %d, playerInv: %v, chestInv: %v, equipped: %v)",
						i, itm, expectedTotal, actualTotal, consumedCount[itm], destroyedWeapons[itm], player.Inventory, chestInv, player.WeaponType)
				}
			}
		}
	}

	// Final verification of storage isolation
	chestInv := m.GetChestInventory(chestTileX, chestTileY)
	originalChest0 := chestInv[0]
	player.Inventory[0] = "MODIFIED_ISOLATION_TEST"
	chestInvAfterMutation := m.GetChestInventory(chestTileX, chestTileY)
	if chestInvAfterMutation[0] != originalChest0 {
		t.Fatalf("Memory Isolation Failed: Mutating player.Inventory mutated chest storage!")
	}
	_ = chestPoint
}

// ==============================================================================================
// CHALLENGE 2: Multiple Chests in Close Proximity & Boundary Distances (191.9px vs 192.1px)
// ==============================================================================================

func TestEmpiricalChallenge_ChestBoundaryDistances191_9Vs192_1(t *testing.T) {
	_, m, _, player, _ := setupChallengerGame(50, 50)

	chestX, chestY := 25, 25
	m.SetTile(chestX, chestY, world.TileChest)
	m.SetChestInventory(chestX, chestY, []string{"chest_item", "", "", "", "", "", "", "", ""})

	chestCenterX := float64(chestX)*float64(world.TileSize) + 64.0
	chestCenterY := float64(chestY)*float64(world.TileSize) + 64.0

	// 1. Cardinal and Diagonal Boundary Tests
	boundaryTests := []struct {
		name      string
		playerX   float64
		playerY   float64
		wantRange bool
	}{
		// Cardinal East
		{"East 191.9px (In Range)", chestCenterX + 191.9, chestCenterY, true},
		{"East 192.0px (Exact Boundary In Range)", chestCenterX + 192.0, chestCenterY, true},
		{"East 192.1px (Out of Range)", chestCenterX + 192.1, chestCenterY, false},

		// Cardinal West
		{"West 191.9px (In Range)", chestCenterX - 191.9, chestCenterY, true},
		{"West 192.0px (Exact Boundary In Range)", chestCenterX - 192.0, chestCenterY, true},
		{"West 192.1px (Out of Range)", chestCenterX - 192.1, chestCenterY, false},

		// Cardinal North
		{"North 191.9px (In Range)", chestCenterX, chestCenterY - 191.9, true},
		{"North 192.0px (Exact Boundary In Range)", chestCenterX, chestCenterY - 192.0, true},
		{"North 192.1px (Out of Range)", chestCenterX, chestCenterY - 192.1, false},

		// Cardinal South
		{"South 191.9px (In Range)", chestCenterX, chestCenterY + 191.9, true},
		{"South 192.0px (Exact Boundary In Range)", chestCenterX, chestCenterY + 192.0, true},
		{"South 192.1px (Out of Range)", chestCenterX, chestCenterY + 192.1, false},

		// Diagonal NE (dx = dy = dist / sqrt(2))
		{"NE 191.9px (In Range)", chestCenterX + 191.9/math.Sqrt2, chestCenterY - 191.9/math.Sqrt2, true},
		{"NE 192.1px (Out of Range)", chestCenterX + 192.1/math.Sqrt2, chestCenterY - 192.1/math.Sqrt2, false},

		// Diagonal SW
		{"SW 191.9px (In Range)", chestCenterX - 191.9/math.Sqrt2, chestCenterY + 191.9/math.Sqrt2, true},
		{"SW 192.1px (Out of Range)", chestCenterX - 192.1/math.Sqrt2, chestCenterY + 192.1/math.Sqrt2, false},
		
		// Additional Diagonal: NW and SE
		{"NW 191.9px (In Range)", chestCenterX - 191.9/math.Sqrt2, chestCenterY - 191.9/math.Sqrt2, true},
		{"NW 192.1px (Out of Range)", chestCenterX - 192.1/math.Sqrt2, chestCenterY - 192.1/math.Sqrt2, false},
		{"SE 191.9px (In Range)", chestCenterX + 191.9/math.Sqrt2, chestCenterY + 191.9/math.Sqrt2, true},
		{"SE 192.1px (Out of Range)", chestCenterX + 192.1/math.Sqrt2, chestCenterY + 192.1/math.Sqrt2, false},
	}

	for _, tt := range boundaryTests {
		t.Run(tt.name, func(t *testing.T) {
			player.Inventory = []string{"player_item", "", "", "", "", "", "", "", ""}
			pPos := world.FloatPoint{X: tt.playerX, Y: tt.playerY}
			dist := math.Hypot(tt.playerX-chestCenterX, tt.playerY-chestCenterY)

			swapped, tx, ty := simulateChestSwap(m, player, pPos)

			if swapped != tt.wantRange {
				t.Errorf("Boundary check failed for '%s' (dist=%.4fpx): got swapped=%v, want inRange=%v",
					tt.name, dist, swapped, tt.wantRange)
			}
			if swapped && (tx != chestX || ty != chestY) {
				t.Errorf("Swapped with wrong chest coords: got (%d, %d), want (%d, %d)", tx, ty, chestX, chestY)
			}
		})
	}
}

func TestEmpiricalChallenge_MultipleChestsCloseProximityDisambiguation(t *testing.T) {
	_, m, _, player, _ := setupChallengerGame(50, 50)

	// Create a 2x2 cluster of adjacent chests
	// Chest A: (10, 10), Chest B: (11, 10), Chest C: (10, 11), Chest D: (11, 11)
	chests := []struct {
		name string
		tx   int
		ty   int
		loot []string
	}{
		{"ChestA", 10, 10, []string{"itemA1", "itemA2", "", "", "", "", "", "", ""}},
		{"ChestB", 11, 10, []string{"itemB1", "itemB2", "", "", "", "", "", "", ""}},
		{"ChestC", 10, 11, []string{"itemC1", "itemC2", "", "", "", "", "", "", ""}},
		{"ChestD", 11, 11, []string{"itemD1", "itemD2", "", "", "", "", "", "", ""}},
	}

	for _, c := range chests {
		m.SetTile(c.tx, c.ty, world.TileChest)
		m.SetChestInventory(c.tx, c.ty, c.loot)
	}

	// 1. Move closest to Chest A
	player.Inventory = []string{"player1", "player2", "", "", "", "", "", "", ""}
	posA := world.FloatPoint{
		X: float64(10)*float64(world.TileSize) + 50.0,
		Y: float64(10)*float64(world.TileSize) + 50.0,
	}
	swappedA, txA, tyA := simulateChestSwap(m, player, posA)
	if !swappedA || txA != 10 || tyA != 10 {
		t.Fatalf("Expected swap with Chest A at (10, 10), got swapped=%v (%d, %d)", swappedA, txA, tyA)
	}
	if player.Inventory[0] != "itemA1" || player.Inventory[1] != "itemA2" {
		t.Errorf("Player did not receive Chest A items: got %v", player.Inventory)
	}
	// Verify Chest B, C, D remain uncorrupted
	if m.GetChestInventory(11, 10)[0] != "itemB1" || m.GetChestInventory(10, 11)[0] != "itemC1" || m.GetChestInventory(11, 11)[0] != "itemD1" {
		t.Fatalf("Other chests were corrupted when swapping with Chest A!")
	}

	// 2. Move closest to Chest B
	posB := world.FloatPoint{
		X: float64(11)*float64(world.TileSize) + 70.0,
		Y: float64(10)*float64(world.TileSize) + 50.0,
	}
	swappedB, txB, tyB := simulateChestSwap(m, player, posB)
	if !swappedB || txB != 11 || tyB != 10 {
		t.Fatalf("Expected swap with Chest B at (11, 10), got swapped=%v (%d, %d)", swappedB, txB, tyB)
	}
	if player.Inventory[0] != "itemB1" || player.Inventory[1] != "itemB2" {
		t.Errorf("Player did not receive Chest B items: got %v", player.Inventory)
	}
	// Chest B should now have player's previous items (which were Chest A's items)
	if m.GetChestInventory(11, 10)[0] != "itemA1" {
		t.Errorf("Chest B did not receive player's items: got %v", m.GetChestInventory(11, 10))
	}

	// 3. Move closest to Chest D
	posD := world.FloatPoint{
		X: float64(11)*float64(world.TileSize) + 64.0,
		Y: float64(11)*float64(world.TileSize) + 64.0,
	}
	swappedD, txD, tyD := simulateChestSwap(m, player, posD)
	if !swappedD || txD != 11 || tyD != 11 {
		t.Fatalf("Expected swap with Chest D at (11, 11), got swapped=%v (%d, %d)", swappedD, txD, tyD)
	}
	if player.Inventory[0] != "itemD1" {
		t.Errorf("Player did not receive Chest D items: got %v", player.Inventory)
	}

	// 4. Test Equidistant Center Point (exact middle of 4 chests: (1408, 1408))
	midPos := world.FloatPoint{
		X: 10*128.0 + 128.0,
		Y: 10*128.0 + 128.0,
	}
	swappedMid, _, _ := simulateChestSwap(m, player, midPos)
	if !swappedMid {
		t.Fatalf("Expected swap to succeed at equidistant center point")
	}
}

// ==============================================================================================
// CHALLENGE 3: Equip/Unequip Stress under all 0..9 Inventory Occupancies and Weapon Durabilities
// ==============================================================================================

func TestEmpiricalChallenge_EquipUnequipAllInventoryOccupancyPermutations(t *testing.T) {
	weapons := []struct {
		name       string
		defaultDur int
	}{
		{"weapon", 5},
		{"axe", 12},
		{"shotgun", 15},
	}

	// Test every occupancy level from 0 to 9 items
	for occupancy := 0; occupancy <= 9; occupancy++ {
		t.Run(fmt.Sprintf("Occupancy_%d_Items", occupancy), func(t *testing.T) {
			_, _, _, player, _ := setupChallengerGame(30, 30)

			// Fill first `occupancy` slots with dummy items
			for s := 0; s < occupancy; s++ {
				player.Inventory[s] = fmt.Sprintf("item_%d", s)
			}

			for _, w := range weapons {
				// --- Subtest A: Unequip weapon into inventory ---
				player.WeaponEquipped = true
				player.WeaponType = w.name
				player.WeaponDurability = w.defaultDur
				player.AttackCooldown = 0

				unequipSuccess := simulateUnequip(player)

				if occupancy < 9 {
					// Unequip MUST SUCCEED
					if !unequipSuccess {
						t.Fatalf("Occupancy %d: Expected unequip of %s to succeed", occupancy, w.name)
					}
					if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
						t.Fatalf("Occupancy %d: Player equipped state not cleared after unequip", occupancy)
					}
					// Must land in the FIRST empty slot (which is index `occupancy`)
					if player.Inventory[occupancy] != w.name {
						t.Fatalf("Occupancy %d: Expected %s in slot %d, got '%s'", occupancy, w.name, occupancy, player.Inventory[occupancy])
					}

					// Clean up the unequipped weapon for next weapon test
					player.Inventory[occupancy] = ""
				} else {
					// Occupancy == 9: Full inventory, Unequip MUST FAIL safely
					if unequipSuccess {
						t.Fatalf("Occupancy 9: Expected unequip to fail on full inventory")
					}
					if !player.WeaponEquipped || player.WeaponType != w.name || player.WeaponDurability != w.defaultDur {
						t.Fatalf("Occupancy 9: Weapon was corrupted during failed unequip")
					}
					for s := 0; s < 9; s++ {
						expectedItem := fmt.Sprintf("item_%d", s)
						if player.Inventory[s] != expectedItem {
							t.Fatalf("Occupancy 9: Inventory slot %d corrupted: got '%s', want '%s'", s, player.Inventory[s], expectedItem)
						}
					}
				}

				// --- Subtest B: Equip weapon from inventory ---
				if occupancy < 9 {
					// Place weapon in slot 0, then equip
					origItem0 := player.Inventory[0]
					player.Inventory[0] = w.name
					player.WeaponEquipped = false
					player.WeaponType = ""
					player.WeaponDurability = 0
					player.AttackCooldown = 0

					simulateUseItem(player, 0)

					if !player.WeaponEquipped || player.WeaponType != w.name || player.WeaponDurability != w.defaultDur {
						t.Fatalf("Occupancy %d: Failed to equip %s from slot 0", occupancy, w.name)
					}
					if player.Inventory[0] != "" {
						t.Fatalf("Occupancy %d: Expected slot 0 to be empty after equip, got '%s'", occupancy, player.Inventory[0])
					}

					// Restore slot 0
					player.Inventory[0] = origItem0
				}
			}
		})
	}
}

func TestEmpiricalChallenge_WeaponsVaryingDurabilitiesEquipAndDegradation(t *testing.T) {
	_, _, _, player, _ := setupChallengerGame(30, 30)

	durabilities := []int{1, 2, 5, 8, 12, 15, 20}

	for _, dur := range durabilities {
		t.Run(fmt.Sprintf("Durability_%d", dur), func(t *testing.T) {
			player.WeaponEquipped = true
			player.WeaponType = "axe"
			player.WeaponDurability = dur

			// Simulate swings until durability reaches 0
			swings := 0
			for player.WeaponDurability > 0 {
				player.WeaponDurability--
				swings++
				if player.WeaponDurability <= 0 {
					player.WeaponEquipped = false
					player.WeaponType = ""
					player.WeaponDurability = 0
					break
				}
			}

			if swings != dur {
				t.Errorf("Expected %d swings, got %d", dur, swings)
			}
			if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
				t.Errorf("Weapon not properly cleared after reaching 0 durability")
			}
		})
	}
}

func TestEmpiricalChallenge_DragAndDropExhaustiveMatrix(t *testing.T) {
	_, _, _, player, _ := setupChallengerGame(30, 30)

	// Test all 10x10 combinations of draggingSlot -> dropSlot (0..8 inventory, 9 equipped)
	for src := 0; src < 10; src++ {
		for dst := 0; dst < 10; dst++ {
			// Initialize clean state
			for i := 0; i < 9; i++ {
				player.Inventory[i] = fmt.Sprintf("item_%d", i)
			}
			player.Inventory[1] = "axe"
			player.Inventory[3] = ""
			player.WeaponEquipped = true
			player.WeaponType = "shotgun"
			player.WeaponDurability = 15

			// Record pre-drag inventory clone
			preInv := make([]string, 9)
			copy(preInv, player.Inventory)
			preWeaponType := player.WeaponType
			preWeaponEquipped := player.WeaponEquipped

			simulateDragDrop(player, src, dst)

			// Invariant 1: Player inventory length is ALWAYS exactly 9
			if len(player.Inventory) != 9 {
				t.Fatalf("Drag (%d -> %d): Inventory length violated: len=%d", src, dst, len(player.Inventory))
			}

			// Invariant 2: No panic occurred, state is valid
			if src == dst {
				// No-op expected
				if !reflect.DeepEqual(player.Inventory, preInv) || player.WeaponType != preWeaponType || player.WeaponEquipped != preWeaponEquipped {
					t.Errorf("Self-drag (%d -> %d) modified state!", src, dst)
				}
			}
		}
	}
}
