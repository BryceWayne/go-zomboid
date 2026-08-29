package game

import (
	"image/color"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	arkecs "github.com/mlange-42/ark/ecs"
)

// setupAdversarialDestructionHarness initializes a test ECS world and game map for empirical adversarial testing
func setupAdversarialDestructionHarness(mapW, mapH int) (*arkecs.World, *world.Map, *UpdateSystem, arkecs.Entity) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(mapW, mapH)
	upd := NewUpdateSystem(w, m)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := pMap.NewEntity(
		&ecs.Player{
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
			FacingX:            1.0,
			FacingY:            0.0,
		},
		&ecs.Position{X: 300.0, Y: 300.0},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 16, H: 16},
		&ecs.Collider{Width: 16, Height: 16},
	)

	return w, m, upd, pEnt
}

// 1. Adversarial Test: Wood Item Drop Conservation Across Mass Destruction
func TestAdversarial_WoodItemDropConservation_MassDestruction(t *testing.T) {
	w, m, _, pEnt := setupAdversarialDestructionHarness(60, 60)
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	// Create a designated test arena: 8x8 grid of mixed destructible and indestructible obstacles
	destructibleCoords := make(map[world.Point]world.TileType)
	indestructibleCoords := make(map[world.Point]world.TileType)

	for ty := 10; ty <= 17; ty++ {
		for tx := 10; tx <= 17; tx++ {
			pt := world.Point{X: tx, Y: ty}
			switch (tx + ty) % 6 {
			case 0:
				m.SetTile(tx, ty, world.TileFence)
				destructibleCoords[pt] = world.TileFence
			case 1:
				m.SetTile(tx, ty, world.TileWall)
				destructibleCoords[pt] = world.TileWall
			case 2:
				m.SetTile(tx, ty, world.TileTree)
				destructibleCoords[pt] = world.TileTree
			case 3:
				m.SetTile(tx, ty, world.TileStump)
				destructibleCoords[pt] = world.TileStump
			case 4:
				m.SetTile(tx, ty, world.TileBench)
				destructibleCoords[pt] = world.TileBench
			default:
				m.SetTile(tx, ty, world.TileConcrete)
				indestructibleCoords[pt] = world.TileConcrete
			}
		}
	}

	expectedDestructibleCount := len(destructibleCoords)
	if expectedDestructibleCount == 0 {
		t.Fatalf("Expected non-zero destructible count in test arena")
	}

	// Equip high-durability Axe
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 500

	// Systematically attack every cell in the arena until all destructibles are destroyed
	for ty := 10; ty <= 17; ty++ {
		for tx := 10; tx <= 17; tx++ {
			cellCenterX := float64(tx)*float64(world.TileSize) + float64(world.TileSize)/2.0
			cellCenterY := float64(ty)*float64(world.TileSize) + float64(world.TileSize)/2.0

			pos.X = cellCenterX - 64.0
			pos.Y = cellCenterY
			player.FacingX = 1.0
			player.FacingY = 0.0

			// Simulate repeated axe swings at cell
			for hit := 0; hit < 5; hit++ {
				if !m.IsDestructible(tx, ty) {
					break
				}
				destroyed, dropType := m.DamageTile(tx, ty, 2)
				if destroyed && dropType != "" {
					itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
					itemMap.NewEntity(
						&ecs.Item{Type: dropType},
						&ecs.Position{X: cellCenterX, Y: cellCenterY},
					)
					break
				}
			}
		}
	}

	// Verify Conservation Invariant 1: Total wood drops spawned in world exactly equals total destroyed obstacles
	itemFilter := arkecs.NewFilter2[ecs.Item, ecs.Position](w)
	query := itemFilter.Query()
	actualDrops := make(map[world.Point]int)
	totalWoodEntities := 0

	for query.Next() {
		item, iPos := query.Get()
		if item.Type != "wood" {
			t.Errorf("Unexpected dropped item type %q (expected 'wood')", item.Type)
		}
		totalWoodEntities++
		tx := int(iPos.X) / world.TileSize
		ty := int(iPos.Y) / world.TileSize
		pt := world.Point{X: tx, Y: ty}
		actualDrops[pt]++

		// Verify drop location exact tile centering
		expectedX := float64(tx)*float64(world.TileSize) + 64.0
		expectedY := float64(ty)*float64(world.TileSize) + 64.0
		if iPos.X != expectedX || iPos.Y != expectedY {
			t.Errorf("Wood drop at tile (%d,%d) off-center: (%f, %f) != (%f, %f)", tx, ty, iPos.X, iPos.Y, expectedX, expectedY)
		}
	}

	if totalWoodEntities != expectedDestructibleCount {
		t.Fatalf("Conservation violation: expected %d wood drops, found %d", expectedDestructibleCount, totalWoodEntities)
	}

	// Verify Conservation Invariant 2: Exactly 1 drop per destructible tile coordinate, 0 for indestructible
	for pt := range destructibleCoords {
		count := actualDrops[pt]
		if count != 1 {
			t.Errorf("Destructible tile at (%d,%d) has %d drops (expected 1)", pt.X, pt.Y, count)
		}
	}

	for pt := range indestructibleCoords {
		count := actualDrops[pt]
		if count != 0 {
			t.Errorf("Indestructible tile at (%d,%d) spawned %d drops (expected 0)", pt.X, pt.Y, count)
		}
	}

	// Verify Conservation Invariant 3: All destroyed tiles are now walkable floors/grass and no longer solid
	for pt, originalType := range destructibleCoords {
		tileNow := m.GetTile(pt.X, pt.Y)
		if tileNow.IsSolid() {
			t.Errorf("Destroyed tile at (%d,%d) [originally %v] is still solid: %v", pt.X, pt.Y, originalType, tileNow)
		}
		if originalType == world.TileWall && tileNow != world.TileWoodFloor {
			t.Errorf("Destroyed wall at (%d,%d) should be TileWoodFloor, got %v", pt.X, pt.Y, tileNow)
		}
		if originalType != world.TileWall && tileNow != world.TileGrass {
			t.Errorf("Destroyed prop at (%d,%d) should be TileGrass, got %v", pt.X, pt.Y, tileNow)
		}
	}
}

// 2. Adversarial Test: Zero Drop On Partial Damage and No Duplication on Post-Destroy Hits
func TestAdversarial_ZeroDropOnPartialDamage_NoPostDestroyDuplication(t *testing.T) {
	w, m, _, _ := setupAdversarialDestructionHarness(40, 40)
	tx, ty := 15, 15
	m.SetTile(tx, ty, world.TileWall) // 3 HP

	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)

	spawnDrop := func(destroyed bool, dropType string) {
		if destroyed && dropType != "" {
			itemMap.NewEntity(
				&ecs.Item{Type: dropType},
				&ecs.Position{
					X: float64(tx)*float64(world.TileSize) + 64.0,
					Y: float64(ty)*float64(world.TileSize) + 64.0,
				},
			)
		}
	}

	countWoodEntities := func() int {
		f := arkecs.NewFilter1[ecs.Item](w)
		q := f.Query()
		c := 0
		for q.Next() {
			c++
		}
		return c
	}

	// Swing 1: 1 dmg -> HP 3 -> 2. No drop.
	d1, drop1 := m.DamageTile(tx, ty, 1)
	spawnDrop(d1, drop1)
	if d1 || drop1 != "" || countWoodEntities() != 0 {
		t.Fatalf("Hit 1: expected no drop, got destroyed=%v drop=%q count=%d", d1, drop1, countWoodEntities())
	}
	if m.GetTileDurability(tx, ty) != 2 {
		t.Fatalf("Hit 1: expected durability 2, got %d", m.GetTileDurability(tx, ty))
	}

	// Swing 2: 1 dmg -> HP 2 -> 1. No drop.
	d2, drop2 := m.DamageTile(tx, ty, 1)
	spawnDrop(d2, drop2)
	if d2 || drop2 != "" || countWoodEntities() != 0 {
		t.Fatalf("Hit 2: expected no drop, got destroyed=%v drop=%q count=%d", d2, drop2, countWoodEntities())
	}
	if m.GetTileDurability(tx, ty) != 1 {
		t.Fatalf("Hit 2: expected durability 1, got %d", m.GetTileDurability(tx, ty))
	}

	// Swing 3: 1 dmg -> HP 1 -> 0. Exactly 1 drop.
	d3, drop3 := m.DamageTile(tx, ty, 1)
	spawnDrop(d3, drop3)
	if !d3 || drop3 != "wood" || countWoodEntities() != 1 {
		t.Fatalf("Hit 3: expected 1 wood drop, got destroyed=%v drop=%q count=%d", d3, drop3, countWoodEntities())
	}
	if m.GetTile(tx, ty) != world.TileWoodFloor {
		t.Fatalf("Hit 3: expected tile to become TileWoodFloor, got %v", m.GetTile(tx, ty))
	}

	// Swings 4 to 20: Post-destruction extra hits -> MUST NOT create duplicate drops
	for extraHit := 4; extraHit <= 20; extraHit++ {
		dExtra, dropExtra := m.DamageTile(tx, ty, 2)
		spawnDrop(dExtra, dropExtra)
		if dExtra || dropExtra != "" {
			t.Fatalf("Post-destroy hit %d: illegally triggered destruction/drop", extraHit)
		}
		if countWoodEntities() != 1 {
			t.Fatalf("Post-destroy hit %d: duplicate drop spawned! total count=%d", extraHit, countWoodEntities())
		}
	}
}

// 3. Adversarial Test: Player Inventory Consecutive Pickups, Backpack Saturation & Ground Retention
func TestAdversarial_InventoryConsecutivePickups_SaturationAndRetention(t *testing.T) {
	w, _, upd, pEnt := setupAdversarialDestructionHarness(40, 40)
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	// Clear inventory
	for i := range player.Inventory {
		player.Inventory[i] = ""
	}

	// Spawn 15 wood drop entities spaced by 128px (1 tile each) along a corridor (X = 100 to 1892, Y = 200)
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	for i := 0; i < 15; i++ {
		itemMap.NewEntity(
			&ecs.Item{Type: "wood"},
			&ecs.Position{X: float64(100 + i*128), Y: 200.0},
		)
	}

	// Move player step-by-step along the corridor (advancing 128px per step)
	for step := 0; step < 15; step++ {
		pos.X = float64(100 + step*128)
		pos.Y = 200.0

		upd.processItems()

		// For steps 0..8, inventory should fill up with "wood"
		expectedCollected := step + 1
		if expectedCollected > 9 {
			expectedCollected = 9
		}

		woodInInventory := 0
		for i := 0; i < len(player.Inventory); i++ {
			if player.Inventory[i] == "wood" {
				woodInInventory++
			}
		}

		if woodInInventory != expectedCollected {
			t.Errorf("Step %d: expected %d wood in inventory, got %d", step, expectedCollected, woodInInventory)
		}

		// Check alive entities in ECS world
		aliveEntities := 0
		itemFilter := arkecs.NewFilter1[ecs.Item](w)
		q := itemFilter.Query()
		for q.Next() {
			aliveEntities++
		}

		expectedAlive := 15 - expectedCollected
		if aliveEntities != expectedAlive {
			t.Errorf("Step %d: expected %d wood entities remaining on ground, got %d", step, expectedAlive, aliveEntities)
		}
	}

	// Final verification: Player inventory has exactly 9 wood items (fully saturated)
	for i := 0; i < 9; i++ {
		if player.Inventory[i] != "wood" {
			t.Fatalf("Inventory slot %d is %q, expected 'wood'", i, player.Inventory[i])
		}
	}

	// Exactly 6 wood items must remain alive on the ground, undamaged and uncollected
	remainingEntities := 0
	itemFilter := arkecs.NewFilter2[ecs.Item, ecs.Position](w)
	q := itemFilter.Query()
	for q.Next() {
		item, iPos := q.Get()
		if item.Type != "wood" {
			t.Errorf("Remaining ground item has wrong type %q", item.Type)
		}
		if iPos.Y != 200.0 {
			t.Errorf("Remaining ground item position corrupted: Y=%f", iPos.Y)
		}
		remainingEntities++
	}

	if remainingEntities != 6 {
		t.Fatalf("Expected 6 remaining wood entities on ground, found %d", remainingEntities)
	}
}

// 3b. Adversarial Test: Instant Batch Cluster Pickup & Saturation (15 drops at single location)
func TestAdversarial_InventoryBatchClusterPickup_Saturation(t *testing.T) {
	w, _, upd, pEnt := setupAdversarialDestructionHarness(40, 40)
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	pos.X = 300.0
	pos.Y = 300.0

	// Clear inventory
	for i := range player.Inventory {
		player.Inventory[i] = ""
	}

	// Spawn 15 wood drops all within 10px of player
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	for i := 0; i < 15; i++ {
		itemMap.NewEntity(
			&ecs.Item{Type: "wood"},
			&ecs.Position{X: 300.0 + float64(i%4)*2.0, Y: 300.0 + float64(i/4)*2.0},
		)
	}

	// Single frame pickup processing
	upd.processItems()

	// Inventory must be completely full (all 9 slots filled with wood)
	for i := 0; i < 9; i++ {
		if player.Inventory[i] != "wood" {
			t.Fatalf("Slot %d expected 'wood', got %q", i, player.Inventory[i])
		}
	}

	// Exactly 6 drops must remain on ground in ECS
	remaining := 0
	itemFilter := arkecs.NewFilter1[ecs.Item](w)
	q := itemFilter.Query()
	for q.Next() {
		remaining++
	}
	if remaining != 6 {
		t.Fatalf("Expected 6 remaining wood items on ground, got %d", remaining)
	}
}

// 4. Adversarial Test: Partial & Fragmented Inventory Pickup with Pre-existing Items
func TestAdversarial_InventoryFragmentedPickup_PreservesExistingItems(t *testing.T) {
	w, _, upd, pEnt := setupAdversarialDestructionHarness(40, 40)
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	// Set fragmented inventory: items at 0, 2, 4, 6, 8; empty at 1, 3, 5, 7 (4 empty slots)
	player.Inventory[0] = "shotgun"
	player.Inventory[1] = ""
	player.Inventory[2] = "ammo"
	player.Inventory[3] = ""
	player.Inventory[4] = "food"
	player.Inventory[5] = ""
	player.Inventory[6] = "water"
	player.Inventory[7] = ""
	player.Inventory[8] = "armor"

	pos.X = 300.0
	pos.Y = 300.0

	// Spawn 7 wood drops at player location (within pickup range < 64px)
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)
	for i := 0; i < 7; i++ {
		itemMap.NewEntity(
			&ecs.Item{Type: "wood"},
			&ecs.Position{X: 300.0 + float64(i)*2.0, Y: 300.0},
		)
	}

	upd.processItems()

	// Invariant 1: Pre-existing items in even slots must be completely preserved
	if player.Inventory[0] != "shotgun" || player.Inventory[2] != "ammo" ||
		player.Inventory[4] != "food" || player.Inventory[6] != "water" || player.Inventory[8] != "armor" {
		t.Fatalf("Pre-existing items corrupted: %v", player.Inventory)
	}

	// Invariant 2: Odd slots 1, 3, 5, 7 must now contain "wood"
	for _, emptySlot := range []int{1, 3, 5, 7} {
		if player.Inventory[emptySlot] != "wood" {
			t.Errorf("Slot %d expected 'wood', got %q", emptySlot, player.Inventory[emptySlot])
		}
	}

	// Invariant 3: Exactly 3 wood items remain on ground (7 - 4 collected)
	remaining := 0
	itemFilter := arkecs.NewFilter1[ecs.Item](w)
	q := itemFilter.Query()
	for q.Next() {
		remaining++
	}
	if remaining != 3 {
		t.Fatalf("Expected 3 remaining wood entities on ground, got %d", remaining)
	}
}

// 5. Adversarial Test: Weapon Breakdown Transitions and Unarmed State Stability
func TestAdversarial_WeaponBreakdownTransitions_ZeroDurabilityStress(t *testing.T) {
	w, m, _, pEnt := setupAdversarialDestructionHarness(40, 40)
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	// Scenario A: Axe with 1 durability chops 2-HP fence -> destroys fence, breaks axe to unarmed
	player.FacingX = 1.0
	player.FacingY = 0.0
	player.WeaponEquipped = true
	player.WeaponType = "axe"
	player.WeaponDurability = 1

	targetTx, targetTy := 10, 10
	m.SetTile(targetTx, targetTy, world.TileFence)

	// Execute Axe chop swing
	destroyed, drop := m.DamageTile(targetTx, targetTy, 2)
	if !destroyed || drop != "wood" {
		t.Fatalf("Axe 2 dmg should destroy fence")
	}
	player.WeaponDurability--
	if player.WeaponDurability <= 0 {
		player.WeaponEquipped = false
		player.WeaponType = ""
		player.WeaponDurability = 0
	}

	// Validate weapon transition state
	if player.WeaponEquipped {
		t.Fatalf("Weapon should NOT be equipped after durability reaches 0")
	}
	if player.WeaponType != "" {
		t.Fatalf("WeaponType should be empty string, got %q", player.WeaponType)
	}
	if player.WeaponDurability != 0 {
		t.Fatalf("WeaponDurability should be 0, got %d", player.WeaponDurability)
	}

	// Place adjacent intact fence at (11, 10)
	m.SetTile(11, 10, world.TileFence)

	// Attempt 100 consecutive swings while unarmed (durability 0 / no weapon)
	for swing := 0; swing < 100; swing++ {
		// Unarmed swing deals 0 barrier damage
		destroyedUnarmed, dropUnarmed := m.DamageTile(11, 10, 0)
		if destroyedUnarmed || dropUnarmed != "" {
			t.Fatalf("Unarmed swing %d illegally damaged/destroyed fence", swing)
		}
		if m.GetTileDurability(11, 10) != 2 {
			t.Fatalf("Fence durability changed during unarmed swing %d: %d", swing, m.GetTileDurability(11, 10))
		}
		if player.WeaponDurability != 0 || player.WeaponEquipped {
			t.Fatalf("Player equipped state corrupted during unarmed swing %d", swing)
		}
	}

	// Scenario B: Equip backup club from backpack, chop fence with 2 swings, verify second break
	player.Inventory[0] = "weapon"
	// Equip item from slot 0
	player.WeaponEquipped = true
	player.WeaponType = "weapon"
	player.WeaponDurability = 2
	player.Inventory[0] = ""

	// Hit 1: 1 dmg -> fence HP 2 -> 1, club dur 2 -> 1
	d1, _ := m.DamageTile(11, 10, 1)
	if d1 {
		t.Fatalf("Club hit 1 should not destroy 2 HP fence")
	}
	player.WeaponDurability--
	if player.WeaponDurability != 1 || !player.WeaponEquipped {
		t.Fatalf("Club should have 1 durability remaining after hit 1")
	}

	// Hit 2: 1 dmg -> fence destroyed, club dur 1 -> 0 -> breaks to unarmed!
	d2, drop2 := m.DamageTile(11, 10, 1)
	if !d2 || drop2 != "wood" {
		t.Fatalf("Club hit 2 must destroy fence and drop wood")
	}
	player.WeaponDurability--
	if player.WeaponDurability <= 0 {
		player.WeaponEquipped = false
		player.WeaponType = ""
		player.WeaponDurability = 0
	}

	if player.WeaponEquipped || player.WeaponType != "" || player.WeaponDurability != 0 {
		t.Fatalf("Club failed to transition to unarmed upon reaching 0 durability")
	}
	if m.GetTile(11, 10) != world.TileGrass {
		t.Fatalf("Tile (11,10) should now be TileGrass")
	}

	_ = pos // suppress unused warning
}

// 6. Adversarial Test: Autotiling Bitmask Transitions and Endcap Redrawing After Destruction
func TestAdversarial_Autotiling_EndcapTransitionsOnDestruction(t *testing.T) {
	m := world.NewMap(50, 50)

	// 1. Horizontal Fence Line: (10,10), (11,10), (12,10), (13,10), (14,10)
	for x := 10; x <= 14; x++ {
		m.SetTile(x, 10, world.TileFence)
	}

	// Initial bitmasks:
	// (10,10): East only (2) -> West endcap
	// (11..13, 10): East + West (2 | 8 = 10) -> Horizontal straight
	// (14,10): West only (8) -> East endcap
	if mask := world.GetFenceBitmask(m, 10, 10); mask != 2 {
		t.Errorf("Fence (10,10) initial bitmask = %d (expected 2)", mask)
	}
	if mask := world.GetFenceBitmask(m, 11, 10); mask != 10 {
		t.Errorf("Fence (11,10) initial bitmask = %d (expected 10)", mask)
	}
	if mask := world.GetFenceBitmask(m, 12, 10); mask != 10 {
		t.Errorf("Fence (12,10) initial bitmask = %d (expected 10)", mask)
	}
	if mask := world.GetFenceBitmask(m, 13, 10); mask != 10 {
		t.Errorf("Fence (13,10) initial bitmask = %d (expected 10)", mask)
	}
	if mask := world.GetFenceBitmask(m, 14, 10); mask != 8 {
		t.Errorf("Fence (14,10) initial bitmask = %d (expected 8)", mask)
	}

	// Destroy center fence tile (12, 10) -> splits line into two 2-tile segments: [10..11] and [13..14]
	destroyed, drop := m.DamageTile(12, 10, 2)
	if !destroyed || drop != "wood" {
		t.Fatalf("Fence (12,10) destruction failed")
	}

	// Verify dynamically recalculated bitmasks:
	// (10,10): East only (2) -> West endcap
	// (11,10): West only (8) -> NEW East endcap! (previously straight segment 10)
	// (12,10): Grass -> TileGrass (bitmask against Fence = 0)
	// (13,10): East only (2) -> NEW West endcap! (previously straight segment 10)
	// (14,10): West only (8) -> East endcap
	if mask := world.GetFenceBitmask(m, 10, 10); mask != 2 {
		t.Errorf("Fence (10,10) post-breach bitmask = %d (expected 2)", mask)
	}
	if mask := world.GetFenceBitmask(m, 11, 10); mask != 8 {
		t.Errorf("Fence (11,10) post-breach bitmask = %d (expected 8 - new East endcap)", mask)
	}
	if mask := world.GetFenceBitmask(m, 13, 10); mask != 2 {
		t.Errorf("Fence (13,10) post-breach bitmask = %d (expected 2 - new West endcap)", mask)
	}
	if mask := world.GetFenceBitmask(m, 14, 10); mask != 8 {
		t.Errorf("Fence (14,10) post-breach bitmask = %d (expected 8)", mask)
	}

	// Destroy (11, 10) -> leaves (10, 10) completely isolated
	m.DamageTile(11, 10, 2)
	if mask := world.GetFenceBitmask(m, 10, 10); mask != 0 {
		t.Errorf("Fence (10,10) post-isolation bitmask = %d (expected 0 - isolated post)", mask)
	}

	// 2. Vertical Wall Line: (20, 20), (20, 21), (20, 22), (20, 23), (20, 24)
	for y := 20; y <= 24; y++ {
		m.SetTile(20, y, world.TileWall)
	}

	// Initial:
	// (20,20): South only (4) -> North endcap
	// (20,21..23): North + South (1 | 4 = 5) -> Vertical straight
	// (20,24): North only (1) -> South endcap
	if mask := world.GetWallBitmask(m, 20, 20); mask != 4 {
		t.Errorf("Wall (20,20) initial bitmask = %d (expected 4)", mask)
	}
	if mask := world.GetWallBitmask(m, 20, 22); mask != 5 {
		t.Errorf("Wall (20,22) initial bitmask = %d (expected 5)", mask)
	}
	if mask := world.GetWallBitmask(m, 20, 24); mask != 1 {
		t.Errorf("Wall (20,24) initial bitmask = %d (expected 1)", mask)
	}

	// Destroy top wall (20, 20) -> (20, 21) becomes new North endcap (bitmask = 4)
	m.DamageTile(20, 20, 3)
	if mask := world.GetWallBitmask(m, 20, 21); mask != 4 {
		t.Errorf("Wall (20,21) post-top-breach bitmask = %d (expected 4 - new North endcap)", mask)
	}

	// Destroy bottom wall (20, 24) -> (20, 23) becomes new South endcap (bitmask = 1)
	m.DamageTile(20, 24, 3)
	if mask := world.GetWallBitmask(m, 20, 23); mask != 1 {
		t.Errorf("Wall (20,23) post-bottom-breach bitmask = %d (expected 1 - new South endcap)", mask)
	}

	// 3. Cross Junction (+) at (30, 30): (30,29)[N], (31,30)[E], (30,31)[S], (29,30)[W]
	m.SetTile(30, 30, world.TileWall)
	m.SetTile(30, 29, world.TileWall) // North
	m.SetTile(31, 30, world.TileWall) // East
	m.SetTile(30, 31, world.TileWall) // South
	m.SetTile(29, 30, world.TileWall) // West

	if mask := world.GetWallBitmask(m, 30, 30); mask != 15 {
		t.Errorf("Wall cross-junction initial bitmask = %d (expected 15)", mask)
	}

	// Step 1: Destroy West arm (29,30) -> becomes T-Junction (N+E+S = 1|2|4 = 7)
	m.DamageTile(29, 30, 3)
	if mask := world.GetWallBitmask(m, 30, 30); mask != 7 {
		t.Errorf("Wall cross-junction after West breach = %d (expected 7 - T-Junction)", mask)
	}

	// Step 2: Destroy East arm (31,30) -> becomes Vertical straight (N+S = 1|4 = 5)
	m.DamageTile(31, 30, 3)
	if mask := world.GetWallBitmask(m, 30, 30); mask != 5 {
		t.Errorf("Wall cross-junction after East breach = %d (expected 5 - Vertical line)", mask)
	}

	// Step 3: Destroy North arm (30,29) -> becomes South endcap (S = 4)
	m.DamageTile(30, 29, 3)
	if mask := world.GetWallBitmask(m, 30, 30); mask != 4 {
		t.Errorf("Wall cross-junction after North breach = %d (expected 4 - South endcap)", mask)
	}

	// Step 4: Destroy South arm (30,31) -> becomes Isolated post (0)
	m.DamageTile(30, 31, 3)
	if mask := world.GetWallBitmask(m, 30, 30); mask != 0 {
		t.Errorf("Wall cross-junction after South breach = %d (expected 0 - Isolated)", mask)
	}
}

// 7. Adversarial Test: Shotgun Cone Cleave Multi-Barrier Destruction and Durability
func TestAdversarial_ShotgunConeCleave_MultiBarrierDestruction(t *testing.T) {
	w, m, _, pEnt := setupAdversarialDestructionHarness(50, 50)
	pMap := arkecs.NewMap1[ecs.Player](w)
	posMap := arkecs.NewMap1[ecs.Position](w)
	player := pMap.Get(pEnt)
	pos := posMap.Get(pEnt)

	pos.X = 200.0
	pos.Y = 200.0
	player.FacingX = 1.0
	player.FacingY = 0.0
	player.WeaponEquipped = true
	player.WeaponType = "shotgun"
	player.WeaponDurability = 10
	player.Inventory[0] = "ammo"

	// Place 4 fences at tiles (2,1), (3,1), (4,1), (5,1)
	fenceTiles := []world.Point{
		{X: 2, Y: 1},
		{X: 3, Y: 1},
		{X: 4, Y: 1},
		{X: 5, Y: 1},
	}
	for _, pt := range fenceTiles {
		m.SetTile(pt.X, pt.Y, world.TileFence)
	}

	// Simulate shotgun firing logic
	player.Inventory[0] = ""
	player.WeaponDurability--

	const maxShotgunRange = 640.0
	const cosSpread = 0.9238795325112867

	minTx := int(pos.X-maxShotgunRange-float64(world.TileSize)/2.0) / world.TileSize
	maxTx := int(pos.X+maxShotgunRange+float64(world.TileSize)/2.0) / world.TileSize
	minTy := int(pos.Y-maxShotgunRange-float64(world.TileSize)/2.0) / world.TileSize
	maxTy := int(pos.Y+maxShotgunRange+float64(world.TileSize)/2.0) / world.TileSize

	destroyedFences := 0
	itemMap := arkecs.NewMap2[ecs.Item, ecs.Position](w)

	for ty := minTy; ty <= maxTy; ty++ {
		for tx := minTx; tx <= maxTx; tx++ {
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
				continue
			}
			tileCenterX := float64(tx)*float64(world.TileSize) + float64(world.TileSize)/2.0
			tileCenterY := float64(ty)*float64(world.TileSize) + float64(world.TileSize)/2.0
			dx := tileCenterX - pos.X
			dy := tileCenterY - pos.Y
			dist := math.Hypot(dx, dy)
			if dist <= maxShotgunRange+float64(world.TileSize)/2.0 {
				inCone := false
				if dist < 96.0+float64(world.TileSize)/2.0 {
					inCone = true
				} else if dist > 0.001 {
					cosAngle := (player.FacingX*dx + player.FacingY*dy) / dist
					if cosAngle >= cosSpread {
						inCone = true
					}
				}
				if inCone && m.IsDestructible(tx, ty) {
					destroyed, dropType := m.DamageTile(tx, ty, 2)
					if destroyed && dropType != "" {
						destroyedFences++
						itemMap.NewEntity(
							&ecs.Item{Type: dropType},
							&ecs.Position{X: tileCenterX, Y: tileCenterY},
						)
					}
				}
			}
		}
	}

	if destroyedFences != 4 {
		t.Fatalf("Expected 4 destroyed fences in shotgun blast cone, got %d", destroyedFences)
	}

	// Verify all 4 fences became TileGrass and spawned 4 wood drops
	for _, pt := range fenceTiles {
		if m.GetTile(pt.X, pt.Y) != world.TileGrass {
			t.Errorf("Fence at (%d,%d) should be TileGrass, got %v", pt.X, pt.Y, m.GetTile(pt.X, pt.Y))
		}
	}

	itemFilter := arkecs.NewFilter1[ecs.Item](w)
	q := itemFilter.Query()
	woodCount := 0
	for q.Next() {
		woodCount++
	}
	if woodCount != 4 {
		t.Fatalf("Expected 4 wood item entities spawned, got %d", woodCount)
	}

	if player.WeaponDurability != 9 {
		t.Errorf("Expected shotgun durability 9, got %d", player.WeaponDurability)
	}
}
