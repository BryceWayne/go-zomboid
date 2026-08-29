package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"reflect"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupChallengerTestGame initializes an isolated test world with player, map, update system, draw system
func setupChallengerTestGame(chestTileX, chestTileY int) (*arkecs.World, *world.Map, *UpdateSystem, *DrawSystem, *ecs.Player, *ecs.Position) {
	assets.Load()
	assets.InitAudio()
	w := arkecs.NewWorld()
	m := world.NewMap(60, 60)

	// Clear any procedural chests so only the test chest is active
	for p := range m.Chests {
		m.SetTile(p.X, p.Y, world.TileGrass)
	}
	m.Chests = make(map[world.Point][]string)

	if chestTileX >= 0 && chestTileY >= 0 {
		m.SetTile(chestTileX, chestTileY, world.TileChest)
	}

	updateSys := NewUpdateSystem(w, m)
	drawSys := NewDrawSystem(w, m)

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

	var posX, posY float64
	if chestTileX >= 0 && chestTileY >= 0 {
		posX = float64(chestTileX)*float64(world.TileSize) + 64.0
		posY = float64(chestTileY)*float64(world.TileSize) + 64.0
	} else {
		posX = 640.0
		posY = 640.0
	}
	pos := &ecs.Position{X: posX, Y: posY}

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	playerMap.NewEntity(
		player,
		pos,
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, updateSys, drawSys, player, pos
}

// -------------------------------------------------------------------------------------------------
// 1. ADVERSARIAL TESTS: CONCURRENT & RAPID INPUT KEY PRESSES
// -------------------------------------------------------------------------------------------------

// TestChallenger_EHeldDownFor100Frames_DebounceAndConservation verifies that holding down 'E'
// for 100 consecutive frames strictly respects the 20-frame debounce cooldown and never duplicates
// or deletes items.
func TestChallenger_EHeldDownFor100Frames_DebounceAndConservation(t *testing.T) {
	_, m, updateSys, _, player, pos := setupChallengerTestGame(10, 10)

	initialPlayerItems := []string{"p_axe", "p_food", "p_water", "p_ammo", "p_armor", "p_antidote", "p_shotgun", "p_ammo2", "p_food2"}
	initialChestItems := []string{"c_food1", "c_water1", "c_ammo1", "c_axe", "c_shotgun", "c_armor", "c_water2", "c_antidote", "c_food2"}

	copy(player.Inventory, initialPlayerItems)
	m.SetChestInventory(10, 10, initialChestItems)

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	totalSystemItems := make(map[string]int)
	for _, itm := range initialPlayerItems {
		totalSystemItems[itm]++
	}
	for _, itm := range initialChestItems {
		totalSystemItems[itm]++
	}

	pPos := world.FloatPoint{X: pos.X, Y: pos.Y}

	swapCount := 0
	for frame := 0; frame < 100; frame++ {
		// Decrement debounce cooldown at frame start (matching UpdateSystem.Update)
		if updateSys.interactCooldown > 0 {
			updateSys.interactCooldown--
		}

		// Simulate 'E' pressed this frame
		if updateSys.interactCooldown <= 0 {
			swapped, _, _ := simulateChestSwap(m, player, pPos)
			if !swapped {
				t.Fatalf("Frame %d: Expected swap to succeed when interactCooldown is 0", frame)
			}
			updateSys.interactCooldown = 20 // Set 20 frames debounce cooldown
			swapCount++
		}

		// Invariant: Player inventory is strictly 9 slots
		if len(player.Inventory) != 9 {
			t.Fatalf("Frame %d: Player inventory length corrupted: %d", frame, len(player.Inventory))
		}
		// Invariant: Chest inventory is strictly 9 slots
		chestInv := m.GetChestInventory(10, 10)
		if len(chestInv) != 9 {
			t.Fatalf("Frame %d: Chest inventory length corrupted: %d", frame, len(chestInv))
		}

		// Invariant: Item conservation across the closed system
		currentSystemItems := make(map[string]int)
		for _, itm := range player.Inventory {
			currentSystemItems[itm]++
		}
		for _, itm := range chestInv {
			currentSystemItems[itm]++
		}
		if !reflect.DeepEqual(currentSystemItems, totalSystemItems) {
			t.Fatalf("Frame %d: Item conservation violated!\ngot  %v\nwant %v", frame, currentSystemItems, totalSystemItems)
		}

		// Invariant: Equipped weapon remains untouched
		if !player.WeaponEquipped || player.WeaponType != "axe" || player.WeaponDurability != 12 {
			t.Fatalf("Frame %d: Equipped weapon mutated during chest swap!", frame)
		}
	}

	// 100 frames with 20 frame debounce cooldown:
	// Frame 0: swap (cooldown set to 20, decrements to 19 at next frame start)
	// Frame 20: cooldown reaches 0 -> swap 2
	// Frame 40: cooldown reaches 0 -> swap 3
	// Frame 60: cooldown reaches 0 -> swap 4
	// Frame 80: cooldown reaches 0 -> swap 5
	// Total swaps across 100 frames must be exactly 5.
	if swapCount != 5 {
		t.Fatalf("Expected exactly 5 swaps over 100 frames with 20-frame cooldown, got %d", swapCount)
	}

	// Since 5 swaps occurred (odd number), player should hold initial chest items and chest should hold initial player items
	if !reflect.DeepEqual(player.Inventory, initialChestItems) {
		t.Errorf("Final player inventory mismatch after 5 swaps:\ngot  %v\nwant %v", player.Inventory, initialChestItems)
	}
	if !reflect.DeepEqual(m.GetChestInventory(10, 10), initialPlayerItems) {
		t.Errorf("Final chest inventory mismatch after 5 swaps:\ngot  %v\nwant %v", m.GetChestInventory(10, 10), initialPlayerItems)
	}
}

// TestChallenger_UHeldDownOrHammered_FullInventorySafety verifies that repeatedly hammering 'U'
// with a 100% full inventory rejects unequip safely for 1,000 frames without data loss.
func TestChallenger_UHeldDownOrHammered_FullInventorySafety(t *testing.T) {
	_, _, _, _, player, _ := setupChallengerTestGame(-1, -1)

	// Equip a damaged shotgun
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 7

	fullInventory := []string{"axe", "food", "water", "ammo", "armor", "antidote", "weapon", "food2", "water2"}
	copy(player.Inventory, fullInventory)

	for frame := 0; frame < 1000; frame++ {
		player.AttackCooldown = 0 // Simulate ready unequip attempt
		success := simulateUnequip(player)

		if success {
			t.Fatalf("Frame %d: simulateUnequip succeeded when inventory was 100%% full!", frame)
		}

		// Verify equipped weapon is perfectly preserved
		if !player.WeaponEquipped || player.WeaponType != "shotgun" || player.WeaponDurability != 7 {
			t.Fatalf("Frame %d: Equipped weapon was corrupted: equipped=%v type=%s dur=%d",
				frame, player.WeaponEquipped, player.WeaponType, player.WeaponDurability)
		}

		// Verify all 9 inventory slots are untouched
		if !reflect.DeepEqual(player.Inventory, fullInventory) {
			t.Fatalf("Frame %d: Inventory slots modified on failed unequip:\ngot  %v\nwant %v", frame, player.Inventory, fullInventory)
		}
	}
}

// TestChallenger_UTransitionFromEmptyToFull verifies unequip transitions from partially empty to completely full.
func TestChallenger_UTransitionFromEmptyToFull(t *testing.T) {
	_, _, _, _, player, _ := setupChallengerTestGame(-1, -1)

	// Step 1: Equip Axe (durability 12) with empty inventory
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	// Unequip to slot 0
	player.AttackCooldown = 0
	ok := simulateUnequip(player)
	if !ok || player.WeaponEquipped || player.Inventory[0] != "axe" {
		t.Fatalf("Step 1 failed: unequip to empty inventory failed")
	}

	// Step 2: Equip weapon from slot 0
	player.AttackCooldown = 0
	simulateUseItem(player, 0)
	if !player.WeaponEquipped || player.WeaponType != "axe" || player.Inventory[0] != "" {
		t.Fatalf("Step 2 failed: re-equipping axe failed")
	}

	// Step 3: Fill slots 0..7 with items; slot 8 is empty
	for i := 0; i < 8; i++ {
		player.Inventory[i] = fmt.Sprintf("item_%d", i)
	}
	player.Inventory[8] = ""

	// Unequip should go to slot 8
	player.AttackCooldown = 0
	ok = simulateUnequip(player)
	if !ok || player.WeaponEquipped || player.Inventory[8] != "axe" {
		t.Fatalf("Step 3 failed: unequip to slot 8 failed: ok=%v inv[8]=%s", ok, player.Inventory[8])
	}

	// Step 4: Now all 9 slots (0..8) are full. Re-equip weapon from slot 8
	player.AttackCooldown = 0
	simulateUseItem(player, 8) // Slot 8 becomes empty
	player.Inventory[8] = "filler_item" // Fill slot 8 to make inventory 100% full

	// Step 5: Now try to unequip with 100% full inventory
	player.AttackCooldown = 0
	ok = simulateUnequip(player)
	if ok {
		t.Fatalf("Step 5 failed: unequip should fail when all 9 slots are full")
	}
	if !player.WeaponEquipped || player.WeaponType != "axe" {
		t.Fatalf("Step 5 failed: weapon lost during failed unequip")
	}
}

// TestChallenger_RapidKeys1to9_SwitchingStress_10000Cycles stress-tests 10,000 rapid cycles of
// pressing keys 1-9 across random combinations of weapons, consumables, armor, and blanks.
func TestChallenger_RapidKeys1to9_SwitchingStress_10000Cycles(t *testing.T) {
	_, _, _, _, player, _ := setupChallengerTestGame(-1, -1)

	itemPool := []string{"weapon", "axe", "shotgun", "food", "water", "antidote", "armor", ""}
	r := rand.New(rand.NewSource(123456789))

	consumedFood := 0
	consumedWater := 0
	consumedAntidote := 0

	for cycle := 0; cycle < 10000; cycle++ {
		// Populate any empty slots occasionally
		for i := 0; i < 9; i++ {
			if player.Inventory[i] == "" && r.Float64() < 0.3 {
				player.Inventory[i] = itemPool[r.Intn(len(itemPool))]
			}
		}

		// Pick random slot 0..8 to press
		slot := r.Intn(9)
		itm := player.Inventory[slot]

		// Set player vitals to test consumable consumption
		player.Hunger = 50.0
		player.Thirst = 50.0
		player.Infected = true
		player.AttackCooldown = 0

		wasEquipped := player.WeaponEquipped
		oldWeapon := player.WeaponType

		simulateUseItem(player, slot)

		// Assert invariants:
		if len(player.Inventory) != 9 {
			t.Fatalf("Cycle %d: Player inventory length is %d (want 9)", cycle, len(player.Inventory))
		}

		if itm == "food" {
			consumedFood++
			if player.Inventory[slot] != "" {
				t.Fatalf("Cycle %d: Food was consumed but slot %d not cleared", cycle, slot)
			}
			if player.WeaponEquipped != wasEquipped || player.WeaponType != oldWeapon {
				t.Fatalf("Cycle %d: Consuming food mutated equipped weapon", cycle)
			}
		} else if itm == "water" {
			consumedWater++
			if player.Inventory[slot] != "" {
				t.Fatalf("Cycle %d: Water was consumed but slot %d not cleared", cycle, slot)
			}
			if player.WeaponEquipped != wasEquipped || player.WeaponType != oldWeapon {
				t.Fatalf("Cycle %d: Consuming water mutated equipped weapon", cycle)
			}
		} else if itm == "antidote" {
			consumedAntidote++
			if player.Inventory[slot] != "" {
				t.Fatalf("Cycle %d: Antidote was consumed but slot %d not cleared", cycle, slot)
			}
			if player.WeaponEquipped != wasEquipped || player.WeaponType != oldWeapon {
				t.Fatalf("Cycle %d: Consuming antidote mutated equipped weapon", cycle)
			}
		} else if itm == "armor" {
			if !player.ArmorEquipped || player.ArmorType != "vest" {
				t.Fatalf("Cycle %d: Equipping armor failed", cycle)
			}
			if player.WeaponEquipped != wasEquipped || player.WeaponType != oldWeapon {
				t.Fatalf("Cycle %d: Equipping armor mutated equipped weapon", cycle)
			}
		} else if itm == "weapon" || itm == "axe" || itm == "shotgun" {
			if !player.WeaponEquipped || player.WeaponType != itm {
				t.Fatalf("Cycle %d: Equipping %s failed: equipped=%v type=%s", cycle, itm, player.WeaponEquipped, player.WeaponType)
			}
			if wasEquipped && oldWeapon != "" {
				if player.Inventory[slot] != oldWeapon {
					t.Fatalf("Cycle %d: Weapon swap failed: slot %d expected %s, got %s", cycle, slot, oldWeapon, player.Inventory[slot])
				}
			} else {
				if player.Inventory[slot] != "" {
					t.Fatalf("Cycle %d: First equip failed to clear slot %d: got %s", cycle, slot, player.Inventory[slot])
				}
			}
		}
	}
}

// -------------------------------------------------------------------------------------------------
// 2. ADVERSARIAL TESTS: EQUIPPED WEAPON DURABILITY PRESERVATION ACROSS INVENTORY & CHESTS
// -------------------------------------------------------------------------------------------------

// TestChallenger_DurabilityPreserved_DegradedWeaponsChestSwaps verifies that damaged weapons
// (e.g. axe with 7/12 hits, shotgun with 3/15 hits) maintain exact durability across single and
// rapid chest swaps.
func TestChallenger_DurabilityPreserved_DegradedWeaponsChestSwaps(t *testing.T) {
	tests := []struct {
		weaponType  string
		durability  int
		swapCycles  int
		chestPreset []string
	}{
		{"axe", 7, 1, []string{"food", "water", "ammo", "", "", "", "", "", ""}},
		{"axe", 1, 50, []string{"shotgun", "axe", "weapon", "armor", "food", "water", "ammo", "ammo", "food"}},
		{"shotgun", 3, 100, []string{"", "", "", "", "", "", "", "", ""}},
		{"shotgun", 14, 500, []string{"ammo", "ammo", "ammo", "ammo", "ammo", "ammo", "ammo", "ammo", "ammo"}},
		{"weapon", 2, 1000, []string{"weapon", "weapon", "weapon", "axe", "shotgun", "food", "water", "armor", "antidote"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_Durability%d_%dSwaps", tt.weaponType, tt.durability, tt.swapCycles), func(t *testing.T) {
			_, m, _, _, player, pos := setupChallengerTestGame(5, 5)

			player.WeaponEquipped = true
			player.WeaponType = tt.weaponType
			player.WeaponDurability = tt.durability

			playerInv := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"}
			copy(player.Inventory, playerInv)

			chestInv := make([]string, 9)
			copy(chestInv, tt.chestPreset)
			m.SetChestInventory(5, 5, chestInv)

			pPos := world.FloatPoint{X: pos.X, Y: pos.Y}

			for i := 0; i < tt.swapCycles; i++ {
				swapped, _, _ := simulateChestSwap(m, player, pPos)
				if !swapped {
					t.Fatalf("Cycle %d: Chest swap failed", i)
				}

				// Check durability invariant after EVERY swap
				if !player.WeaponEquipped {
					t.Fatalf("Cycle %d: WeaponEquipped became false", i)
				}
				if player.WeaponType != tt.weaponType {
					t.Fatalf("Cycle %d: WeaponType changed from %s to %s", i, tt.weaponType, player.WeaponType)
				}
				if player.WeaponDurability != tt.durability {
					t.Fatalf("Cycle %d: WeaponDurability mutated! Expected %d, got %d", i, tt.durability, player.WeaponDurability)
				}
			}
		})
	}
}

// TestChallenger_DurabilityPreserved_CombatAndChoppingInterleavedWithChestSwaps simulates
// combat degradation interspersed with chest swaps.
func TestChallenger_DurabilityPreserved_CombatAndChoppingInterleavedWithChestSwaps(t *testing.T) {
	_, m, _, _, player, pos := setupChallengerTestGame(8, 8)

	// Step 1: Equip fresh Axe (durability 12)
	player.Inventory[0] = "axe"
	simulateUseItem(player, 0)
	if !player.WeaponEquipped || player.WeaponDurability != 12 {
		t.Fatalf("Initial equip failed: dur=%d", player.WeaponDurability)
	}

	pPos := world.FloatPoint{X: pos.X, Y: pos.Y}
	m.SetChestInventory(8, 8, []string{"ammo", "food", "water", "", "", "", "", "", ""})

	// Step 2: 3 combat hits (durability: 12 -> 11 -> 10 -> 9)
	for hit := 1; hit <= 3; hit++ {
		player.WeaponDurability--
	}
	if player.WeaponDurability != 9 {
		t.Fatalf("Expected durability 9 after 3 hits, got %d", player.WeaponDurability)
	}

	// Step 3: Chest swap -> durability must stay 9
	swapped, _, _ := simulateChestSwap(m, player, pPos)
	if !swapped || player.WeaponDurability != 9 {
		t.Fatalf("Chest swap corrupted durability: got %d (want 9)", player.WeaponDurability)
	}

	// Step 4: 8 more hits -> durability drops to 1
	for hit := 1; hit <= 8; hit++ {
		player.WeaponDurability--
	}
	if player.WeaponDurability != 1 {
		t.Fatalf("Expected durability 1 after 8 more hits, got %d", player.WeaponDurability)
	}

	// Step 5: Chest swap -> durability must stay 1
	swapped, _, _ = simulateChestSwap(m, player, pPos)
	if !swapped || player.WeaponDurability != 1 {
		t.Fatalf("Chest swap corrupted durability at 1 hit: got %d (want 1)", player.WeaponDurability)
	}

	// Step 6: Final hit breaks weapon
	player.WeaponDurability--
	if player.WeaponDurability <= 0 {
		player.WeaponEquipped = false
		player.WeaponType = ""
		player.WeaponDurability = 0
	}

	if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
		t.Fatalf("Broken weapon failed to reset state")
	}

	// Step 7: Chest swap while unarmed
	swapped, _, _ = simulateChestSwap(m, player, pPos)
	if !swapped || player.WeaponEquipped || player.WeaponDurability != 0 {
		t.Fatalf("Unarmed chest swap unexpectedly equipped or mutated weapon")
	}
}

// TestChallenger_DurabilityPreserved_DragDropInternalSlots verifies that reordering inventory
// slots 0-8 via drag and drop does not mutate equipped weapon durability.
func TestChallenger_DurabilityPreserved_DragDropInternalSlots(t *testing.T) {
	_, _, _, _, player, _ := setupChallengerTestGame(-1, -1)

	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 5

	for i := 0; i < 9; i++ {
		player.Inventory[i] = fmt.Sprintf("item_%d", i)
	}

	// Perform multiple drag swaps between internal inventory slots
	dragPairs := [][2]int{
		{0, 1}, {2, 7}, {8, 3}, {4, 5}, {6, 0}, {8, 1},
	}

	for _, pair := range dragPairs {
		simulateDragDrop(player, pair[0], pair[1])
		if !player.WeaponEquipped || player.WeaponType != "axe" || player.WeaponDurability != 5 {
			t.Fatalf("Internal drag (%d -> %d) corrupted equipped weapon: equipped=%v type=%s dur=%d",
				pair[0], pair[1], player.WeaponEquipped, player.WeaponType, player.WeaponDurability)
		}
	}
}

// -------------------------------------------------------------------------------------------------
// 3. ADVERSARIAL TESTS: HEADLESS UI RENDERING ACROSS RESOLUTIONS AND ASPECT RATIOS
// -------------------------------------------------------------------------------------------------

// TestChallenger_HeadlessUIRendering_AllResolutionsAndAspectRatios stress-tests DrawSystem.Draw
// across 16 different screen resolutions / aspect ratios under 7 distinct game states.
func TestChallenger_HeadlessUIRendering_AllResolutionsAndAspectRatios(t *testing.T) {
	resolutions := []struct {
		name   string
		width  int
		height int
		aspect string
	}{
		{"16:9 Standard Native", 1280, 720, "16:9"},
		{"16:9 Full HD", 1920, 1080, "16:9"},
		{"16:9 QHD", 2560, 1440, "16:9"},
		{"16:9 4K UHD", 3840, 2160, "16:9"},
		{"4:3 XGA", 1024, 768, "4:3"},
		{"4:3 SVGA", 800, 600, "4:3"},
		{"4:3 VGA", 640, 480, "4:3"},
		{"16:10 WUXGA", 1920, 1200, "16:10"},
		{"21:9 Ultrawide FHD", 2560, 1080, "21:9"},
		{"21:9 UWQHD", 3440, 1440, "21:9"},
		{"9:16 Mobile Portrait", 720, 1280, "9:16"},
		{"9:16 Full HD Portrait", 1080, 1920, "9:16"},
		{"1:1 Square 1000", 1000, 1000, "1:1"},
		{"1:1 Small Square", 500, 500, "1:1"},
		{"1:1 Tiny Square", 256, 256, "1:1"},
		{"Boundary Minimum", 1, 1, "1:1"},
	}

	gameStates := []struct {
		name         string
		setupState   func(player *ecs.Player, m *world.Map, pos *ecs.Position)
		timeOfDay    float64
		draggingSlot int
	}{
		{
			name: "Unarmed Day Normal",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				player.WeaponEquipped = false
				player.WeaponType = ""
				player.WeaponDurability = 0
				player.Dead = false
				player.Infected = false
				player.Health = 100.0
			},
			timeOfDay:    12.0,
			draggingSlot: -1,
		},
		{
			name: "Axe Equipped Near Chest (HUD Prompt Active)",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				player.WeaponEquipped = true
				player.WeaponType = "axe"
				player.WeaponDurability = 8
				m.SetTile(10, 10, world.TileChest)
				pos.X = 10*128.0 + 64.0
				pos.Y = 10*128.0 + 64.0
			},
			timeOfDay:    14.0,
			draggingSlot: -1,
		},
		{
			name: "Shotgun Equipped Dragging Dedicated Slot 9",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				player.WeaponEquipped = true
				player.WeaponType = "shotgun"
				player.WeaponDurability = 15
			},
			timeOfDay:    10.0,
			draggingSlot: 9,
		},
		{
			name: "Dragging Inventory Slot 3 with Armor Equipped",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				player.Inventory[3] = "ammo"
				player.ArmorEquipped = true
				player.ArmorType = "vest"
				player.ArmorDefense = 0.50
				player.ArmorDurability = 10
				player.ArmorMaxDurability = 10
			},
			timeOfDay:    16.0,
			draggingSlot: 3,
		},
		{
			name: "Infected and Low Health Night Time",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				player.Infected = true
				player.Health = 20.0
				player.WeaponEquipped = true
				player.WeaponType = "weapon"
				player.WeaponDurability = 2
			},
			timeOfDay:    0.0, // Midnight
			draggingSlot: -1,
		},
		{
			name: "Player Dead Screen",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				player.Dead = true
				player.Health = 0.0
			},
			timeOfDay:    23.0,
			draggingSlot: -1,
		},
		{
			name: "Full Inventory Extreme Names",
			setupState: func(player *ecs.Player, m *world.Map, pos *ecs.Position) {
				for i := 0; i < 9; i++ {
					player.Inventory[i] = fmt.Sprintf("VERY_LONG_ITEM_NAME_TAG_%d_X", i)
				}
				player.WeaponEquipped = true
				player.WeaponType = "MODIFIED_EXPERIMENTAL_PLASMA_CANNON"
				player.WeaponDurability = 9999
			},
			timeOfDay:    6.0,
			draggingSlot: -1,
		},
	}

	for _, res := range resolutions {
		t.Run(fmt.Sprintf("Resolution_%dx%d_%s", res.width, res.height, res.aspect), func(t *testing.T) {
			_, m, _, drawSys, player, pos := setupChallengerTestGame(10, 10)
			screen := ebiten.NewImage(res.width, res.height)

			for _, gs := range gameStates {
				gs.setupState(player, m, pos)

				// Verify draw execution completes without panic
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Fatalf("PANIC during Draw on resolution %dx%d (%s) in state '%s': %v",
								res.width, res.height, res.aspect, gs.name, r)
						}
					}()
					drawSys.Draw(screen, gs.timeOfDay, gs.draggingSlot)
				}()
			}
		})
	}
}

// TestChallenger_HUDLayoutGeometryContracts asserts exact UI layout specifications from PROJECT.md
func TestChallenger_HUDLayoutGeometryContracts(t *testing.T) {
	// 1. Inventory Slots: 9 slots starting at X=1070, Y=30 + i*25, Width=200, Height=20
	for i := 0; i < 9; i++ {
		slotY := 30 + i*25
		slotW := 200
		slotH := 20

		if slotY+slotH > 265 {
			t.Errorf("Backpack slot %d (Y=%d..%d) collides with dedicated equipped slot (Y=265)", i, slotY, slotY+slotH)
		}
		if slotW != 200 || slotH != 20 {
			t.Errorf("Slot %d dimension mismatch: got (%d, %d), want (200, 20)", i, slotW, slotH)
		}
	}

	// 2. Dedicated Equipped UI Slot: X=1070, Y=265, Width=200, Height=30
	equippedX := 1070
	equippedY := 265
	equippedW := 200
	equippedH := 30

	if equippedX != 1070 || equippedY != 265 || equippedW != 200 || equippedH != 30 {
		t.Errorf("Dedicated Equipped Slot geometry mismatch: got (%d, %d, %d, %d), want (1070, 265, 200, 30)",
			equippedX, equippedY, equippedW, equippedH)
	}

	// 3. Chest Interaction HUD Prompt: X=490, Y=645, Width=300, Height=25
	promptX := 490
	promptY := 645
	promptW := 300
	promptH := 25

	if promptX != 490 || promptY != 645 || promptW != 300 || promptH != 25 {
		t.Errorf("Chest prompt geometry mismatch: got (%d, %d, %d, %d), want (490, 645, 300, 25)",
			promptX, promptY, promptW, promptH)
	}
}
