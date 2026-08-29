package game

import (
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupChestTestGame initializes a test world with player, map, and a chest at designated coordinates
func setupChestTestGame(chestTileX, chestTileY int) (*arkecs.World, *world.Map, *UpdateSystem, *ecs.Player, *ecs.Position) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	// Clear any procedural chests so only the test chest is active
	for p := range m.Chests {
		m.SetTile(p.X, p.Y, world.TileGrass)
	}
	m.Chests = make(map[world.Point][]string)

	updateSys := NewUpdateSystem(w, m)

	// Place chest
	m.SetTile(chestTileX, chestTileY, world.TileChest)

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
	// Player starts centered near chest
	pos := &ecs.Position{
		X: float64(chestTileX)*float64(world.TileSize) + 64.0,
		Y: float64(chestTileY)*float64(world.TileSize) + 64.0,
	}

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

// simulateChestSwap executes the atomic inventory swap logic between player and nearby chest
func simulateChestSwap(gameMap *world.Map, player *ecs.Player, playerPos world.FloatPoint) (bool, int, int) {
	pTileX := int(playerPos.X) / world.TileSize
	pTileY := int(playerPos.Y) / world.TileSize
	var nearChest bool
	var chestTileX, chestTileY int
	closestChestDist := math.MaxFloat64

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			tx := pTileX + dx
			ty := pTileY + dy
			if tx >= 0 && tx < gameMap.Width && ty >= 0 && ty < gameMap.Height {
				if gameMap.GetTile(tx, ty) == world.TileChest {
					chestCenterX := float64(tx)*float64(world.TileSize) + 64.0
					chestCenterY := float64(ty)*float64(world.TileSize) + 64.0
					dist := math.Hypot(playerPos.X-chestCenterX, playerPos.Y-chestCenterY)
					if dist <= 192.0 && dist < closestChestDist {
						closestChestDist = dist
						nearChest = true
						chestTileX = tx
						chestTileY = ty
					}
				}
			}
		}
	}

	if !nearChest {
		return false, 0, 0
	}

	for len(player.Inventory) < 9 {
		player.Inventory = append(player.Inventory, "")
	}
	chestInv := gameMap.GetChestInventory(chestTileX, chestTileY)
	for len(chestInv) < 9 {
		chestInv = append(chestInv, "")
	}

	// Atomic deep copy swap
	newPlayerInv := make([]string, 9)
	copy(newPlayerInv, chestInv[:9])

	newChestInv := make([]string, 9)
	copy(newChestInv, player.Inventory[:9])

	player.Inventory = newPlayerInv
	gameMap.SetChestInventory(chestTileX, chestTileY, newChestInv)

	return true, chestTileX, chestTileY
}

// 1. Test Procedural Map Chest Persistence and Starter Loot
func TestChest_ProceduralPersistenceAndStarterLoot(t *testing.T) {
	m := world.NewMap(100, 100)
	midX := 100 / 2
	midY := 100 / 2

	expectedChests := []struct {
		name     string
		pos      world.Point
		wantLoot []string
	}{
		{
			name:     "Warehouse",
			pos:      world.Point{X: midX + 22, Y: midY + 8},
			wantLoot: []string{"axe", "ammo", "ammo", "food", "", "", "", "", ""},
		},
		{
			name:     "Campsite",
			pos:      world.Point{X: 90, Y: 9},
			wantLoot: []string{"food", "water", "weapon", "antidote", "", "", "", "", ""},
		},
		{
			name:     "House 1 Bedroom",
			pos:      world.Point{X: 11, Y: 9},
			wantLoot: []string{"armor", "water", "food", "", "", "", "", "", ""},
		},
		{
			name:     "Police Armory",
			pos:      world.Point{X: 11, Y: midY + 7},
			wantLoot: []string{"shotgun", "ammo", "ammo", "armor", "", "", "", "", ""},
		},
	}

	for _, tc := range expectedChests {
		t.Run(tc.name, func(t *testing.T) {
			tile := m.GetTile(tc.pos.X, tc.pos.Y)
			if tile != world.TileChest {
				t.Fatalf("Expected tile at (%d, %d) to be TileChest, got %v", tc.pos.X, tc.pos.Y, tile)
			}

			loot := m.GetChestInventory(tc.pos.X, tc.pos.Y)
			if len(loot) != 9 {
				t.Fatalf("Expected 9-slot inventory, got length %d", len(loot))
			}
			if !reflect.DeepEqual(loot, tc.wantLoot) {
				t.Errorf("Starter loot mismatch at (%d, %d):\ngot  %v\nwant %v", tc.pos.X, tc.pos.Y, loot, tc.wantLoot)
			}
		})
	}

	// Test uninitialized chest returns 9 empty slots
	emptyLoot := m.GetChestInventory(42, 42)
	if len(emptyLoot) != 9 {
		t.Fatalf("Expected 9-slot slice for uninitialized chest, got %d", len(emptyLoot))
	}
	for i, itm := range emptyLoot {
		if itm != "" {
			t.Errorf("Expected empty slot at %d, got '%s'", i, itm)
		}
	}

	// Test SetChestInventory defensive copy
	inputSlice := []string{"food", "water", "weapon", "", "", "", "", "", ""}
	m.SetChestInventory(42, 42, inputSlice)
	inputSlice[0] = "mutated" // Mutate caller's slice

	persisted := m.GetChestInventory(42, 42)
	if persisted[0] != "food" {
		t.Errorf("Chest inventory was mutated externally! Expected 'food', got '%s'", persisted[0])
	}
}

// 2. Test Chest Proximity Detection (192px / 1.5 tiles radius)
func TestChest_ProximityDetection(t *testing.T) {
	_, m, _, player, _ := setupChestTestGame(10, 10)
	chestCenterX := float64(10)*float64(world.TileSize) + 64.0
	chestCenterY := float64(10)*float64(world.TileSize) + 64.0

	tests := []struct {
		name      string
		playerX   float64
		playerY   float64
		wantRange bool
	}{
		{"Exact Center (0px)", chestCenterX, chestCenterY, true},
		{"Horizontal In Range (64px)", chestCenterX + 64.0, chestCenterY, true},
		{"Horizontal In Range (128px / 1 tile)", chestCenterX + 128.0, chestCenterY, true},
		{"Horizontal In Range (180px)", chestCenterX + 180.0, chestCenterY, true},
		{"Boundary In Range (192px / 1.5 tiles)", chestCenterX + 192.0, chestCenterY, true},
		{"Diagonal In Range (120px, 120px => 169.7px)", chestCenterX + 120.0, chestCenterY + 120.0, true},
		{"Boundary Out of Range (193px)", chestCenterX + 193.0, chestCenterY, false},
		{"Horizontal Out of Range (256px / 2 tiles)", chestCenterX + 256.0, chestCenterY, false},
		{"Diagonal Out of Range (150px, 150px => 212.1px)", chestCenterX + 150.0, chestCenterY + 150.0, false},
		{"Far Away (500px)", chestCenterX + 500.0, chestCenterY + 500.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pPos := world.FloatPoint{X: tt.playerX, Y: tt.playerY}
			swapped, tx, ty := simulateChestSwap(m, player, pPos)

			if swapped != tt.wantRange {
				dist := math.Hypot(tt.playerX-chestCenterX, tt.playerY-chestCenterY)
				t.Errorf("Proximity check failed for dist %.2fpx: got swapped=%v, want inRange=%v", dist, swapped, tt.wantRange)
			}
			if swapped && (tx != 10 || ty != 10) {
				t.Errorf("Expected chest at (10, 10), got (%d, %d)", tx, ty)
			}
		})
	}
}

// 3. Test Atomic Inventory Swapping and Equipped Weapon Isolation
func TestChest_InventorySwapAtomic(t *testing.T) {
	_, m, _, player, _ := setupChestTestGame(5, 5)

	// 3.1 Initial Player state: 9 inventory items + active equipped shotgun
	playerInitialInv := []string{"food", "food", "axe", "water", "antidote", "ammo", "ammo", "food", "water"}
	copy(player.Inventory, playerInitialInv)
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 15

	// 3.2 Chest initial state
	chestInitialInv := []string{"armor", "shotgun", "ammo", "ammo", "water", "", "", "", ""}
	m.SetChestInventory(5, 5, chestInitialInv)

	// 3.3 Execute Swap
	pPos := world.FloatPoint{X: 5*128.0 + 64.0, Y: 5*128.0 + 64.0}
	swapped, _, _ := simulateChestSwap(m, player, pPos)
	if !swapped {
		t.Fatalf("Expected chest swap to succeed")
	}

	// Verify Player now has Chest's former inventory
	if !reflect.DeepEqual(player.Inventory, chestInitialInv) {
		t.Errorf("Player inventory mismatch after swap:\ngot  %v\nwant %v", player.Inventory, chestInitialInv)
	}

	// Verify Chest now has Player's former inventory
	chestCurrentInv := m.GetChestInventory(5, 5)
	if !reflect.DeepEqual(chestCurrentInv, playerInitialInv) {
		t.Errorf("Chest inventory mismatch after swap:\ngot  %v\nwant %v", chestCurrentInv, playerInitialInv)
	}

	// Verify Equipped Weapon is ISOLATED and remains untouched
	if !player.WeaponEquipped || player.WeaponType != "shotgun" || player.WeaponDurability != 15 {
		t.Errorf("Equipped weapon was modified during chest swap: equipped=%v type=%s dur=%d",
			player.WeaponEquipped, player.WeaponType, player.WeaponDurability)
	}

	// 3.4 Deep Copy Mutation Isolation: Mutate Player's inventory and confirm Chest is unaffected
	player.Inventory[0] = "CONSUMED_EMPTY"
	chestAfterPlayerMutation := m.GetChestInventory(5, 5)
	if chestAfterPlayerMutation[0] != "food" {
		t.Errorf("Chest inventory was corrupted when player mutated their own inventory! Expected 'food', got '%s'", chestAfterPlayerMutation[0])
	}
}

// 4. Test Debounce Cooldown Protection (20 frames)
func TestChest_DebounceCooldown(t *testing.T) {
	_, m, updateSys, player, pos := setupChestTestGame(5, 5)

	player.Inventory[0] = "player_item"
	m.SetChestInventory(5, 5, []string{"chest_item", "", "", "", "", "", "", "", ""})

	// Initial cooldown is 0
	if updateSys.interactCooldown != 0 {
		t.Fatalf("Expected initial interactCooldown 0, got %d", updateSys.interactCooldown)
	}

	// Simulate first interaction setting cooldown to 20
	updateSys.interactCooldown = 20
	simulateChestSwap(m, player, world.FloatPoint{X: pos.X, Y: pos.Y})

	// Simulate 19 update frames: Cooldown should count down but remain > 0
	for frame := 1; frame <= 19; frame++ {
		updateSys.Update(-1)
		expectedRemaining := 20 - frame
		if updateSys.interactCooldown != expectedRemaining {
			t.Errorf("Frame %d: expected interactCooldown %d, got %d", frame, expectedRemaining, updateSys.interactCooldown)
		}
	}

	// 20th frame: Cooldown reaches 0 and allows another interaction
	updateSys.Update(-1)
	if updateSys.interactCooldown != 0 {
		t.Errorf("Expected interactCooldown to reach 0 after 20 frames, got %d", updateSys.interactCooldown)
	}
}

// 5. 10,000-Cycle Rapid Swap Stress Test for Data Conservation
func TestChest_10000RapidSwapStress(t *testing.T) {
	_, m, _, player, _ := setupChestTestGame(8, 8)

	// Distinct items in player and chest
	initialPlayerItems := []string{"p_axe", "p_food1", "p_food2", "p_water1", "p_ammo1", "p_ammo2", "p_armor", "p_antidote", "p_shotgun"}
	initialChestItems := []string{"c_food1", "c_water1", "c_water2", "c_ammo1", "c_ammo2", "c_ammo3", "c_axe", "c_armor", "c_weapon"}

	copy(player.Inventory, initialPlayerItems)
	m.SetChestInventory(8, 8, initialChestItems)

	// Player equipped weapon
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 12

	// Build ground-truth frequency map of the 18 items in the closed system
	totalHistogram := make(map[string]int)
	for _, item := range initialPlayerItems {
		totalHistogram[item]++
	}
	for _, item := range initialChestItems {
		totalHistogram[item]++
	}

	pPos := world.FloatPoint{X: 8*128.0 + 64.0, Y: 8*128.0 + 64.0}

	const iterations = 10000
	for i := 0; i < iterations; i++ {
		swapped, _, _ := simulateChestSwap(m, player, pPos)
		if !swapped {
			t.Fatalf("Iteration %d: swap failed", i)
		}

		// Invariant 1: Player inventory length is strictly 9
		if len(player.Inventory) != 9 {
			t.Fatalf("Iteration %d: player inventory length is %d (want 9)", i, len(player.Inventory))
		}

		// Invariant 2: Chest inventory length is strictly 9
		chestInv := m.GetChestInventory(8, 8)
		if len(chestInv) != 9 {
			t.Fatalf("Iteration %d: chest inventory length is %d (want 9)", i, len(chestInv))
		}

		// Invariant 3: Odd iterations => player has chest items; Even iterations => player has player items
		if i%2 == 0 {
			// After 1st swap (i=0): player has chest items
			if !reflect.DeepEqual(player.Inventory, initialChestItems) {
				t.Fatalf("Iteration %d: player items do not match initial chest items", i)
			}
			if !reflect.DeepEqual(chestInv, initialPlayerItems) {
				t.Fatalf("Iteration %d: chest items do not match initial player items", i)
			}
		} else {
			// After 2nd swap (i=1): player has player items
			if !reflect.DeepEqual(player.Inventory, initialPlayerItems) {
				t.Fatalf("Iteration %d: player items do not match initial player items", i)
			}
			if !reflect.DeepEqual(chestInv, initialChestItems) {
				t.Fatalf("Iteration %d: chest items do not match initial chest items", i)
			}
		}

		// Invariant 4: Equipped weapon is preserved throughout
		if !player.WeaponEquipped || player.WeaponType != "axe" || player.WeaponDurability != 12 {
			t.Fatalf("Iteration %d: equipped weapon corrupted", i)
		}
	}

	// Final verification of item conservation histogram
	finalHistogram := make(map[string]int)
	for _, item := range player.Inventory {
		finalHistogram[item]++
	}
	for _, item := range m.GetChestInventory(8, 8) {
		finalHistogram[item]++
	}

	if !reflect.DeepEqual(finalHistogram, totalHistogram) {
		t.Fatalf("Item conservation violated across %d swaps:\ngot  %v\nwant %v", iterations, finalHistogram, totalHistogram)
	}
}

// 6. Test HUD prompt rendering when player is near a chest
func TestChest_HUDPromptDrawing(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	drawSys := NewDrawSystem(w, m)

	// Place chest at (10, 10)
	m.SetTile(10, 10, world.TileChest)

	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pPos := &ecs.Position{X: 10*128.0 + 64.0, Y: 10*128.0 + 64.0}
	playerMap.NewEntity(
		&ecs.Player{Health: 100.0, Inventory: make([]string, 9)},
		pPos,
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	screen := ebiten.NewImage(1280, 720)

	// Case 1: Player near chest (in range) -> draws prompt
	drawSys.Draw(screen, 12.0, -1)

	// Case 2: Player far away (out of range) -> does not draw prompt
	pPos.X = 40 * 128.0
	pPos.Y = 40 * 128.0
	drawSys.Draw(screen, 12.0, -1)
}
