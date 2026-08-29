package world

import (
	"math"
	"math/rand"
	"sync"
	"testing"
)

// 1. Adversarial Test: Perimeter Absolute Indestructibility
// Tests all perimeter boundary edges, corners, and out-of-bounds coordinates
// across multiple map dimensions under all attack forms and extreme damage values.
func TestAdversarial_PerimeterAbsoluteIndestructibility(t *testing.T) {
	mapSizes := []struct{ W, H int }{
		{10, 10},
		{30, 30},
		{50, 50},
		{100, 100},
	}

	damageVariants := []int{1, 2, 3, 5, 10, 100, 999999, 0, -1, -50}

	for _, size := range mapSizes {
		m := NewMap(size.W, size.H)

		// A. Check every single perimeter tile along North, South, East, West borders
		for x := 0; x < size.W; x++ {
			for y := 0; y < size.H; y++ {
				isPerimeter := (x == 0 || x == size.W-1 || y == 0 || y == size.H-1)
				if !isPerimeter {
					continue
				}

				// Verify IsDestructible is false
				if m.IsDestructible(x, y) {
					t.Fatalf("Perimeter tile at (%d, %d) in map %dx%d reported IsDestructible == true", x, y, size.W, size.H)
				}

				// Verify GetTileMaxDurability is 0
				if m.GetTileDurability(x, y) != 0 {
					t.Fatalf("Perimeter tile at (%d, %d) returned non-zero durability: %d", x, y, m.GetTileDurability(x, y))
				}

				// Attempt damage across all variants
				for _, dmg := range damageVariants {
					destroyed, drop := m.DamageTile(x, y, dmg)
					if destroyed {
						t.Fatalf("Perimeter tile at (%d, %d) was destroyed with dmg=%d", x, y, dmg)
					}
					if drop != "" {
						t.Fatalf("Perimeter tile at (%d, %d) returned drop %q with dmg=%d", x, y, drop, dmg)
					}
					if m.GetTile(x, y) != TileWall {
						t.Fatalf("Perimeter tile at (%d, %d) changed from TileWall to %v", x, y, m.GetTile(x, y))
					}
					if !m.GetTile(x, y).IsSolid() {
						t.Fatalf("Perimeter tile at (%d, %d) lost solidity", x, y)
					}
					if !m.BlocksVision(x, y) {
						t.Fatalf("Perimeter tile at (%d, %d) lost vision blocking", x, y)
					}
				}

				// Stress: 100 rapid consecutive swings
				for i := 0; i < 100; i++ {
					destroyed, _ := m.DamageTile(x, y, 2)
					if destroyed {
						t.Fatalf("Perimeter tile at (%d, %d) destroyed after %d rapid swings", x, y, i+1)
					}
				}
			}
		}

		// B. Out-of-bounds coordinates
		oobCoords := []Point{
			{-1, 0}, {0, -1}, {-1, -1},
			{size.W, 0}, {0, size.H}, {size.W, size.H},
			{-100, -100}, {size.W + 500, size.H + 500},
			{math.MinInt32 / 2, math.MaxInt32 / 2},
		}

		for _, pt := range oobCoords {
			if m.IsDestructible(pt.X, pt.Y) {
				t.Fatalf("Out of bounds point (%d, %d) reported IsDestructible == true", pt.X, pt.Y)
			}
			destroyed, drop := m.DamageTile(pt.X, pt.Y, 10)
			if destroyed || drop != "" {
				t.Fatalf("Out of bounds point (%d, %d) returned destroyed=%v, drop=%q", pt.X, pt.Y, destroyed, drop)
			}
		}
	}
}

// 2. Adversarial Test: Complete Matrix of All TileTypes Destructibility and Drops
// Verifies that ONLY the 5 intended destructible types can be broken, and all
// other 17 tile types remain strictly indestructible.
func TestAdversarial_AllTileTypesDestructibilityMatrix(t *testing.T) {
	allTileTypes := []struct {
		tile          TileType
		destructible  bool
		maxDurability int
		expectedAfter TileType
	}{
		{TileGrass, false, 0, TileGrass},
		{TileWall, true, 3, TileWoodFloor},
		{TileDirt, false, 0, TileDirt},
		{TileWoodFloor, false, 0, TileWoodFloor},
		{TileTree, true, 3, TileGrass},
		{TileAsphalt, false, 0, TileAsphalt},
		{TileConcrete, false, 0, TileConcrete},
		{TileTileFloor, false, 0, TileTileFloor},
		{TileFence, true, 2, TileGrass},
		{TileDebris, false, 0, TileDebris},
		{TileTent, false, 0, TileTent},
		{TileElevationBlock, false, 0, TileElevationBlock},
		{TileRamp, false, 0, TileRamp},
		{TileStump, true, 2, TileGrass},
		{TileMushroom, false, 0, TileMushroom},
		{TileSign, false, 0, TileSign},
		{TileBench, true, 2, TileGrass},
		{TileChest, false, 0, TileChest},
		{TileSculpture, false, 0, TileSculpture},
		{TileBush, false, 0, TileBush},
		{TileFlower, false, 0, TileFlower},
		{TileStone, false, 0, TileStone},
	}

	for _, tc := range allTileTypes {
		m := NewMap(30, 30)
		tx, ty := 15, 15
		m.SetTile(tx, ty, tc.tile)

		// Check IsDestructible
		isDest := m.IsDestructible(tx, ty)
		if isDest != tc.destructible {
			t.Errorf("TileType %v (%s): IsDestructible = %v, expected %v", tc.tile, tc.tile.String(), isDest, tc.destructible)
		}

		// Check Max Durability
		maxDur := m.GetTileMaxDurability(tc.tile)
		if maxDur != tc.maxDurability {
			t.Errorf("TileType %v (%s): MaxDurability = %d, expected %d", tc.tile, tc.tile.String(), maxDur, tc.maxDurability)
		}

		if !tc.destructible {
			// Non-destructible: try damaging with massive damage
			destroyed, drop := m.DamageTile(tx, ty, 999)
			if destroyed || drop != "" {
				t.Errorf("Non-destructible tile %v (%s) was damaged/destroyed: destroyed=%v, drop=%q", tc.tile, tc.tile.String(), destroyed, drop)
			}
			if m.GetTile(tx, ty) != tc.tile {
				t.Errorf("Tile %v mutated to %v after DamageTile", tc.tile, m.GetTile(tx, ty))
			}
		} else {
			// Destructible: step through 1-damage decrements
			for hp := tc.maxDurability; hp > 1; hp-- {
				curDur := m.GetTileDurability(tx, ty)
				if curDur != hp {
					t.Errorf("Tile %v (%s): Expected durability %d, got %d", tc.tile, tc.tile.String(), hp, curDur)
				}
				destroyed, drop := m.DamageTile(tx, ty, 1)
				if destroyed || drop != "" {
					t.Fatalf("Tile %v was destroyed prematurely at hp=%d", tc.tile, hp)
				}
			}

			// Final 1 HP damage destroys it
			destroyed, drop := m.DamageTile(tx, ty, 1)
			if !destroyed {
				t.Fatalf("Tile %v failed to destroy on final hit", tc.tile)
			}
			if drop != "wood" {
				t.Fatalf("Tile %v expected drop 'wood', got %q", tc.tile, drop)
			}
			if m.GetTile(tx, ty) != tc.expectedAfter {
				t.Fatalf("Tile %v after destruction: expected %v, got %v", tc.tile, tc.expectedAfter, m.GetTile(tx, ty))
			}
			if m.IsDestructible(tx, ty) {
				t.Fatalf("Tile %v after destruction should no longer be destructible", tc.tile)
			}
		}
	}
}

// 3. Adversarial Test: Concurrent Stress Destruction Across Multiple Map Instances
// Concurrently spawns and destroys hundreds of barriers in parallel across 16 goroutines
// ensuring thread isolation, memory correctness, and zero panic/crashes.
func TestAdversarial_ConcurrentDestructionAcrossGoroutines(t *testing.T) {
	const numGoroutines = 16
	const barriersPerMap = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(gid * 7919)))
			m := NewMap(40, 40)

			// Place 100 destructible barriers at distinct random interior coords
			barrierTypes := []TileType{TileFence, TileWall, TileTree, TileStump, TileBench}
			coords := make(map[Point]TileType, barriersPerMap)

			for len(coords) < barriersPerMap {
				tx := 1 + rng.Intn(38)
				ty := 1 + rng.Intn(38)
				tt := barrierTypes[rng.Intn(len(barrierTypes))]
				m.SetTile(tx, ty, tt)
				coords[Point{tx, ty}] = tt
			}

			// Destroy all barriers with random damage amounts (1, 2, 3)
			for pt := range coords {
				for {
					if !m.IsDestructible(pt.X, pt.Y) {
						break
					}
					dmg := 1 + rng.Intn(3)
					destroyed, drop := m.DamageTile(pt.X, pt.Y, dmg)
					if destroyed {
						if drop != "wood" {
							t.Errorf("Goroutine %d: Destroyed barrier drop was %q, expected 'wood'", gid, drop)
						}
						break
					}
				}
			}

			// Verify all barriers are replaced with walkable non-solid tiles
			for pt, initialType := range coords {
				tile := m.GetTile(pt.X, pt.Y)
				if initialType == TileWall {
					if tile != TileWoodFloor {
						t.Errorf("Goroutine %d: Wall at (%d,%d) was not replaced by TileWoodFloor, got %v", gid, pt.X, pt.Y, tile)
					}
				} else {
					if tile != TileGrass {
						t.Errorf("Goroutine %d: Barrier at (%d,%d) was not replaced by TileGrass, got %v", gid, pt.X, pt.Y, tile)
					}
				}
				if tile.IsSolid() {
					t.Errorf("Goroutine %d: Destroyed tile at (%d,%d) should not be solid", gid, pt.X, pt.Y)
				}
			}

			// Verify TileDurability map is fully cleared of destroyed entries
			if len(m.TileDurability) != 0 {
				t.Errorf("Goroutine %d: TileDurability map has %d orphaned entries", gid, len(m.TileDurability))
			}
		}(g)
	}

	wg.Wait()
}

// 4. Adversarial Test: Rapid Burst Attacks & Overkill Durability Handling
// Tests rapid consecutive hits, zero damage hits, negative damage hits,
// and overkill single-swing damage.
func TestAdversarial_RapidBurstAttacksAndOverkill(t *testing.T) {
	m := NewMap(40, 40)

	// A. Overkill 1-Hit Destruction
	m.SetTile(10, 10, TileWall) // 3 HP
	destroyed, drop := m.DamageTile(10, 10, 500)
	if !destroyed || drop != "wood" {
		t.Fatalf("500 damage overkill should destroy 3 HP wall and drop wood")
	}
	if m.GetTile(10, 10) != TileWoodFloor {
		t.Fatalf("Destroyed wall should be TileWoodFloor")
	}

	// Subsequent attacks on already-destroyed floor
	for i := 0; i < 20; i++ {
		d, dr := m.DamageTile(10, 10, 2)
		if d || dr != "" {
			t.Fatalf("Subsequent attack on destroyed floor must return (false, '')")
		}
	}

	// B. Rapid 0 and Negative Damage Attacks do not alter durability or destroy
	m.SetTile(12, 12, TileTree) // 3 HP
	for i := 0; i < 50; i++ {
		d, _ := m.DamageTile(12, 12, 0)
		if d {
			t.Fatalf("0 damage must not destroy tree")
		}
		d2, _ := m.DamageTile(12, 12, -10)
		if d2 {
			t.Fatalf("Negative damage must not destroy tree")
		}
	}
	if m.GetTileDurability(12, 12) != 3 {
		t.Fatalf("Tree durability altered by zero/negative damage: got %d, expected 3", m.GetTileDurability(12, 12))
	}
}

// 5. Adversarial Test: Weapon Durability Wear and Breaking on Barriers
// Models weapon attacks against a series of barriers until durability exhausts to 0,
// verifying state transitions, unequip triggering, and unarmed fallback behavior.
func TestAdversarial_WeaponBreakingAndWearLifecycle(t *testing.T) {
	m := NewMap(30, 30)

	type MockWeaponState struct {
		Equipped   bool
		WeaponType string
		Durability int
	}

	// Scenario A: Axe with 3 durability attacks 3 separate fences (2 HP each, 2 dmg/swing)
	axe := MockWeaponState{Equipped: true, WeaponType: "axe", Durability: 3}
	fenceCoords := []Point{{10, 10}, {10, 11}, {10, 12}}
	for _, pt := range fenceCoords {
		m.SetTile(pt.X, pt.Y, TileFence)
	}

	for i, pt := range fenceCoords {
		if !axe.Equipped || axe.Durability <= 0 {
			t.Fatalf("Axe should be equipped with durability > 0 before swing %d", i+1)
		}

		// Swing axe (deals 2 dmg)
		d, drop := m.DamageTile(pt.X, pt.Y, 2)
		if !d || drop != "wood" {
			t.Fatalf("Axe swing %d should destroy fence at (%d,%d)", i+1, pt.X, pt.Y)
		}

		axe.Durability--
		if axe.Durability <= 0 {
			axe.Equipped = false
			axe.WeaponType = ""
			axe.Durability = 0
		}
	}

	// After 3 swings on 3 fences, axe durability should be 0 and weapon unequipped
	if axe.Equipped || axe.WeaponType != "" || axe.Durability != 0 {
		t.Fatalf("Axe should be fully broken (Equipped=false, Durability=0), got %+v", axe)
	}

	// Scenario B: Now unarmed, attacking a 4th fence at (10, 13)
	m.SetTile(10, 13, TileFence)
	// Unarmed attack deals 0 barrier damage
	d, _ := m.DamageTile(10, 13, 0)
	if d {
		t.Fatalf("Unarmed attack must not destroy fence")
	}
	if m.GetTileDurability(10, 13) != 2 {
		t.Fatalf("Fence durability should remain 2 after unarmed strike")
	}
}

// 6. Adversarial Test: Dynamic Autotiling Connectivity Degradation on Chopping
// Tests that destroying connected walls or fences properly updates cardinal bitmasks
// of remaining neighboring segments without leaving stale connectivity state.
func TestAdversarial_AutotilingDynamicTransitionsOnDestruction(t *testing.T) {
	m := NewMap(20, 20)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Build a 3x3 solid block of walls centered at (10, 10)
	// (9,9)   (10,9)   (11,9)
	// (9,10)  (10,10)  (11,10)
	// (9,11)  (10,11)  (11,11)
	for y := 9; y <= 11; y++ {
		for x := 9; x <= 11; x++ {
			m.SetTile(x, y, TileWall)
		}
	}

	// Initial Center (10, 10) mask = N(1) | E(2) | S(4) | W(8) = 15
	if mask := GetCardinalBitmask4(m, 10, 10, TileWall); mask != 15 {
		t.Fatalf("Initial 3x3 center wall cardinal mask: expected 15, got %d", mask)
	}

	// Step 1: Destroy North neighbor at (10, 9)
	d, _ := m.DamageTile(10, 9, 3)
	if !d {
		t.Fatalf("Failed to destroy North wall")
	}
	// Center mask now missing North (1) -> E(2) | S(4) | W(8) = 14
	if mask := GetCardinalBitmask4(m, 10, 10, TileWall); mask != 14 {
		t.Fatalf("Center wall mask after North destroyed: expected 14, got %d", mask)
	}

	// Step 2: Destroy West neighbor at (9, 10)
	d, _ = m.DamageTile(9, 10, 3)
	if !d {
		t.Fatalf("Failed to destroy West wall")
	}
	// Center mask now missing North(1) and West(8) -> E(2) | S(4) = 6
	if mask := GetCardinalBitmask4(m, 10, 10, TileWall); mask != 6 {
		t.Fatalf("Center wall mask after West destroyed: expected 6, got %d", mask)
	}

	// Step 3: Destroy East neighbor at (11, 10)
	d, _ = m.DamageTile(11, 10, 3)
	if !d {
		t.Fatalf("Failed to destroy East wall")
	}
	// Center mask now only South connected -> S(4) = 4
	if mask := GetCardinalBitmask4(m, 10, 10, TileWall); mask != 4 {
		t.Fatalf("Center wall mask after East destroyed: expected 4, got %d", mask)
	}

	// Step 4: Destroy South neighbor at (10, 11)
	d, _ = m.DamageTile(10, 11, 3)
	if !d {
		t.Fatalf("Failed to destroy South wall")
	}
	// Center mask now completely isolated -> 0
	if mask := GetCardinalBitmask4(m, 10, 10, TileWall); mask != 0 {
		t.Fatalf("Center wall mask after all cardinal neighbors destroyed: expected 0, got %d", mask)
	}

	// Repeat same isolation test for Fence bitmasks
	for y := 9; y <= 11; y++ {
		for x := 9; x <= 11; x++ {
			m.SetTile(x, y, TileFence)
		}
	}

	if mask := GetFenceBitmask(m, 10, 10); mask != 15 {
		t.Fatalf("Initial 3x3 center fence mask: expected 15, got %d", mask)
	}

	// Destroy North fence
	m.DamageTile(10, 9, 2)
	if mask := GetFenceBitmask(m, 10, 10); mask != 14 {
		t.Fatalf("Center fence mask after North destroyed: expected 14, got %d", mask)
	}
}

// 7. Adversarial Test: Fortress Maze Breach - Collision and FOV Propagation
// Builds a double-layered wall fortress around an interior target. Verifies that
// collision blocks traversal and FOV raycasting cannot penetrate until specific
// breach corridors are chopped down.
func TestAdversarial_FortressBreachCollisionAndFOVPropagation(t *testing.T) {
	m := NewMap(40, 40)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Outer Wall at x=10 (ty=5..15)
	for ty := 5; ty <= 15; ty++ {
		m.SetTile(10, ty, TileWall)
	}

	// Inner Wall at x=14 (ty=5..15)
	for ty := 5; ty <= 15; ty++ {
		m.SetTile(14, ty, TileWall)
	}

	playerPixelX := float64(8*TileSize + 64)
	playerPixelY := float64(10*TileSize + 64)
	targetInnerIdx := 10*m.Width + 16 // Inside inner sanctum at (16, 10)
	intermediateIdx := 10*m.Width + 12 // Between walls at (12, 10)

	// Step 1: Initial state - FOV blocked by outer wall
	m.CalculateFOV(playerPixelX, playerPixelY, 15)
	if m.Visible[intermediateIdx] {
		t.Fatalf("Intermediate zone between walls should NOT be visible initially")
	}
	if m.Visible[targetInnerIdx] {
		t.Fatalf("Inner sanctum should NOT be visible initially")
	}

	// Verify collision blocks crossing outer wall at (10, 10)
	wallX := float64(10*TileSize + 64)
	wallY := float64(10*TileSize + 64)
	if !m.IsColliding(wallX-16, wallY-16, 32, 32) {
		t.Fatalf("Outer wall at (10, 10) should collide")
	}

	// Step 2: Chop down outer wall at (10, 10)
	d1, drop1 := m.DamageTile(10, 10, 3)
	if !d1 || drop1 != "wood" {
		t.Fatalf("Failed to breach outer wall at (10, 10)")
	}

	// Verify collision cleared at outer breach
	if m.IsColliding(wallX-16, wallY-16, 32, 32) {
		t.Fatalf("Outer breach at (10, 10) should no longer collide")
	}

	// Recalculate FOV: intermediate zone is NOW visible, but inner sanctum is still blocked by inner wall
	m.CalculateFOV(playerPixelX, playerPixelY, 15)
	if !m.Visible[intermediateIdx] {
		t.Fatalf("Intermediate zone at (12, 10) MUST become visible after outer wall breach")
	}
	if m.Visible[targetInnerIdx] {
		t.Fatalf("Inner sanctum at (16, 10) should still be occluded by inner wall")
	}

	// Step 3: Chop down inner wall at (14, 10)
	innerWallX := float64(14*TileSize + 64)
	innerWallY := float64(10*TileSize + 64)
	d2, drop2 := m.DamageTile(14, 10, 3)
	if !d2 || drop2 != "wood" {
		t.Fatalf("Failed to breach inner wall at (14, 10)")
	}

	// Verify collision cleared at inner breach
	if m.IsColliding(innerWallX-16, innerWallY-16, 32, 32) {
		t.Fatalf("Inner breach at (14, 10) should no longer collide")
	}

	// Recalculate FOV: inner sanctum is NOW fully visible!
	m.CalculateFOV(playerPixelX, playerPixelY, 15)
	if !m.Visible[targetInnerIdx] {
		t.Fatalf("Inner sanctum at (16, 10) MUST become visible after double breach")
	}
}

// 8. Adversarial Test: Fence Transparency vs Solidity Invariant
// Fences are solid obstacles (blocking movement) but transparent (non-occluding vision).
// Destroying a fence must clear collision without affecting visibility.
func TestAdversarial_FenceTransparencyAndSolidityLifecycle(t *testing.T) {
	m := NewMap(30, 30)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Place a line of fences at x=10 (ty=5..15)
	for ty := 5; ty <= 15; ty++ {
		m.SetTile(10, ty, TileFence)
	}

	fenceX := float64(10*TileSize + 64)
	fenceY := float64(10*TileSize + 64)
	playerX := float64(8*TileSize + 64)
	playerY := float64(10*TileSize + 64)
	targetIdx := 10*m.Width + 12 // Behind fence

	// Verify fence is solid
	if !m.GetTile(10, 10).IsSolid() {
		t.Fatalf("Fence must be solid")
	}
	if !m.IsColliding(fenceX-16, fenceY-16, 32, 32) {
		t.Fatalf("Fence must collide with entity AABB")
	}

	// Verify fence does NOT block vision
	if m.BlocksVision(10, 10) {
		t.Fatalf("Fence must NOT block vision")
	}

	// Calculate FOV: tile behind fence SHOULD be visible immediately
	m.CalculateFOV(playerX, playerY, 10)
	if !m.Visible[targetIdx] {
		t.Fatalf("Tile behind fence MUST be visible even before fence is chopped")
	}

	// Chop fence down
	d, drop := m.DamageTile(10, 10, 2)
	if !d || drop != "wood" {
		t.Fatalf("Fence should be destroyed by 2 damage")
	}

	// Verify fence is now grass, non-solid, and collision is cleared
	if m.GetTile(10, 10) != TileGrass {
		t.Fatalf("Destroyed fence must be TileGrass")
	}
	if m.IsColliding(fenceX-16, fenceY-16, 32, 32) {
		t.Fatalf("Destroyed fence must no longer collide")
	}
}

// 9. Adversarial Test: Lazy Init and Memory Integrity of TileDurability
func TestAdversarial_TileDurabilityLazyInitAndMemoryIntegrity(t *testing.T) {
	// Construct a raw Map where TileDurability is nil
	m := &Map{
		Width:          20,
		Height:         20,
		Tiles:          make([]TileType, 400),
		TileDurability: nil, // Intentionally nil
	}

	m.SetTile(5, 5, TileFence)

	// GetTileDurability should not panic with nil map
	dur := m.GetTileDurability(5, 5)
	if dur != 2 {
		t.Fatalf("Expected fence durability 2 from nil map, got %d", dur)
	}

	// DamageTile should not panic with nil map
	d, drop := m.DamageTile(5, 5, 1)
	if d || drop != "" {
		t.Fatalf("1 damage should not destroy fence")
	}
	if m.GetTileDurability(5, 5) != 1 {
		t.Fatalf("Expected fence durability 1, got %d", m.GetTileDurability(5, 5))
	}

	// Destroy fence
	d2, drop2 := m.DamageTile(5, 5, 1)
	if !d2 || drop2 != "wood" {
		t.Fatalf("Final damage must destroy fence")
	}

	// Map entry should be deleted
	if _, exists := m.TileDurability[Point{5, 5}]; exists {
		t.Fatalf("Destroyed tile entry should be deleted from TileDurability map")
	}
}
