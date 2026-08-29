package world

import (
	"math"
	"testing"
)

// TestEmpirical_All10TileTypesGenerated verifies that every one of the 10 defined TileTypes
// is generated in the procedural world map with non-zero occurrence.
func TestEmpirical_All10TileTypesGenerated(t *testing.T) {
	m := NewMap(100, 100)

	allTileTypes := []TileType{
		TileGrass,
		TileWall,
		TileDirt,
		TileWoodFloor,
		TileTree,
		TileAsphalt,
		TileConcrete,
		TileTileFloor,
		TileFence,
		TileDebris,
	}

	counts := make(map[TileType]int)
	for _, tile := range m.Tiles {
		counts[tile]++
	}

	for _, tt := range allTileTypes {
		c, exists := counts[tt]
		if !exists || c == 0 {
			t.Errorf("FAIL: TileType %d (%s) was NOT generated in the map (count = %d)", tt, tt.String(), c)
		} else {
			t.Logf("PASS: TileType %-14s (ID %d): count = %5d", tt.String(), tt, c)
		}
	}

	// Verify total tile count matches dimensions
	expectedTotal := 100 * 100
	if len(m.Tiles) != expectedTotal {
		t.Fatalf("FAIL: Expected %d tiles, got %d", expectedTotal, len(m.Tiles))
	}
}

// TestEmpirical_All5BuildingArchetypesAndRooms verifies that all 5 building archetypes
// are present and that every building has valid, non-empty room bounds within its footprint.
func TestEmpirical_All5BuildingArchetypesAndRooms(t *testing.T) {
	m := NewMap(100, 100)

	requiredArchetypes := map[BuildingType]bool{
		BuildingResidential: false,
		BuildingGrocery:     false,
		BuildingPolice:      false,
		BuildingPharmacy:    false,
		BuildingWarehouse:   false,
	}

	if len(m.Buildings) == 0 {
		t.Fatalf("FAIL: No buildings generated in map")
	}

	for bIdx, b := range m.Buildings {
		if _, ok := requiredArchetypes[b.Type]; ok {
			requiredArchetypes[b.Type] = true
		}

		t.Logf("Building #%d: Type=%-12s Pos=(%2d,%2d) Size=%2dx%2d Rooms=%d Doors=%d",
			bIdx, b.Type, b.X, b.Y, b.W, b.H, len(b.Rooms), len(b.Doors))

		// Check building bounds
		if b.X <= 0 || b.Y <= 0 || b.W <= 0 || b.H <= 0 {
			t.Errorf("FAIL: Building #%d (%s) has invalid coordinates/dimensions: (%d,%d, %dx%d)",
				bIdx, b.Type, b.X, b.Y, b.W, b.H)
		}
		if b.X+b.W >= m.Width || b.Y+b.H >= m.Height {
			t.Errorf("FAIL: Building #%d (%s) exceeds map boundary: right=%d >= %d, bottom=%d >= %d",
				bIdx, b.Type, b.X+b.W, m.Width, b.Y+b.H, m.Height)
		}

		// Must have at least 1 room and 1 door
		if len(b.Rooms) == 0 {
			t.Errorf("FAIL: Building #%d (%s) has 0 rooms", bIdx, b.Type)
		}
		if len(b.Doors) == 0 {
			t.Errorf("FAIL: Building #%d (%s) has 0 doors", bIdx, b.Type)
		}

		// Validate each room
		for rIdx, r := range b.Rooms {
			if r.W <= 0 || r.H <= 0 {
				t.Errorf("FAIL: Building #%d (%s) Room #%d (%s) has non-positive dimensions: %dx%d",
					bIdx, b.Type, rIdx, r.Type, r.W, r.H)
			}

			// Room must be strictly within the building footprint
			if r.X < b.X || r.Y < b.Y || r.X+r.W > b.X+b.W || r.Y+r.H > b.Y+b.H {
				t.Errorf("FAIL: Building #%d (%s) Room #%d (%s) pos (%d,%d,%dx%d) exceeds building box (%d,%d,%dx%d)",
					bIdx, b.Type, rIdx, r.Type, r.X, r.Y, r.W, r.H, b.X, b.Y, b.W, b.H)
			}

			// Validate room contains walkable non-solid floor tiles
			walkableCount := 0
			for rx := r.X; rx < r.X+r.W; rx++ {
				for ry := r.Y; ry < r.Y+r.H; ry++ {
					tile := m.GetTile(rx, ry)
					if !tile.IsSolid() {
						walkableCount++
					}
				}
			}
			if walkableCount == 0 {
				t.Errorf("FAIL: Room #%d (%s) at (%d,%d,%dx%d) contains NO walkable tiles",
					rIdx, r.Type, r.X, r.Y, r.W, r.H)
			}
		}

		// Validate doors are non-solid
		for dIdx, d := range b.Doors {
			if d.X < b.X || d.X >= b.X+b.W || d.Y < b.Y || d.Y >= b.Y+b.H {
				t.Errorf("FAIL: Building #%d Door #%d at (%d,%d) is outside building footprint (%d,%d,%dx%d)",
					bIdx, dIdx, d.X, d.Y, b.X, b.Y, b.W, b.H)
			}
			doorTile := m.GetTile(d.X, d.Y)
			if doorTile.IsSolid() {
				t.Errorf("FAIL: Building #%d Door #%d at (%d,%d) has solid tile %v (%s)",
					bIdx, dIdx, d.X, d.Y, doorTile, doorTile.String())
			}
		}
	}

	for arch, present := range requiredArchetypes {
		if !present {
			t.Errorf("FAIL: Required building archetype %s was NOT generated", arch)
		} else {
			t.Logf("PASS: Required archetype %s is present", arch)
		}
	}
}

// TestEmpirical_PlayerSpawnSafetyAndZombieDistance verifies that:
// 1. Player spawn is on a non-solid tile.
// 2. Player's 16x16 bounding box does not collide with solid geometry.
// 3. Player spawn is strictly >= 350.0 units away from ALL zombie spawns.
func TestEmpirical_PlayerSpawnSafetyAndZombieDistance(t *testing.T) {
	for seed := 0; seed < 30; seed++ {
		m := NewMap(100, 100)

		px := int(m.PlayerSpawn.X) / TileSize
		py := int(m.PlayerSpawn.Y) / TileSize

		spawnTile := m.GetTile(px, py)
		if spawnTile.IsSolid() {
			t.Fatalf("FAIL: Iter %d: Player spawn tile at (%d,%d) is solid: %v (%s)",
				seed, px, py, spawnTile, spawnTile.String())
		}

		// Test AABB collision at player spawn with 16x16 collider (centered or offset)
		playerBoxX := m.PlayerSpawn.X - 8.0
		playerBoxY := m.PlayerSpawn.Y - 8.0
		if m.IsColliding(playerBoxX, playerBoxY, 16.0, 16.0) {
			t.Fatalf("FAIL: Iter %d: Player 16x16 AABB at spawn (%f, %f) collides with solid obstacle",
				seed, playerBoxX, playerBoxY)
		}

		// Check distance to all zombie spawns
		minDist := math.MaxFloat64
		for zIdx, zs := range m.ZombieSpawns {
			dx := zs.X - m.PlayerSpawn.X
			dy := zs.Y - m.PlayerSpawn.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < minDist {
				minDist = dist
			}
			if dist < 350.0 {
				t.Fatalf("FAIL: Iter %d: Zombie #%d at (%f, %f) is too close to player spawn (%f, %f): dist=%.2f < 350.0",
					seed, zIdx, zs.X, zs.Y, m.PlayerSpawn.X, m.PlayerSpawn.Y, dist)
			}
		}
		if seed == 0 {
			t.Logf("PASS: Player spawn at (%.1f, %.1f) [tile %d,%d, %s], closest zombie distance = %.2f (target min >= 350.0)",
				m.PlayerSpawn.X, m.PlayerSpawn.Y, px, py, spawnTile.String(), minDist)
		}
	}
}

// TestEmpirical_100PercentZombieSpawnsNonSolid verifies that 100% of zombie spawns across
// multiple map generations land strictly on non-solid tiles.
func TestEmpirical_100PercentZombieSpawnsNonSolid(t *testing.T) {
	totalZombiesChecked := 0

	for seed := 0; seed < 30; seed++ {
		m := NewMap(100, 100)

		if len(m.ZombieSpawns) == 0 {
			t.Fatalf("FAIL: Iter %d: Map has 0 zombie spawns", seed)
		}

		for zIdx, zs := range m.ZombieSpawns {
			totalZombiesChecked++
			tx := int(zs.X) / TileSize
			ty := int(zs.Y) / TileSize

			tile := m.GetTile(tx, ty)
			if tile.IsSolid() {
				t.Fatalf("FAIL: Iter %d Zombie #%d at (%f, %f) [tile %d,%d] spawned on SOLID tile %v (%s)",
					seed, zIdx, zs.X, zs.Y, tx, ty, tile, tile.String())
			}

			// Also verify within map boundaries
			if zs.X < 0 || zs.X >= float64(m.Width*TileSize) || zs.Y < 0 || zs.Y >= float64(m.Height*TileSize) {
				t.Fatalf("FAIL: Iter %d Zombie #%d at (%f, %f) is out of map boundaries", seed, zIdx, zs.X, zs.Y)
			}
		}
	}

	t.Logf("PASS: Verified %d zombie spawns across 30 map generations: 100.0%% are non-solid and within bounds", totalZombiesChecked)
}

// TestEmpirical_AABBCollisionSolidVsFloor checks AABB collision detection against
// all solid and floor tiles with sweeping sub-pixel positions.
func TestEmpirical_AABBCollisionSolidVsFloor(t *testing.T) {
	m := NewMap(30, 30)

	// Clear to grass
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			m.SetTile(x, y, TileGrass)
		}
	}

	solidTiles := []struct {
		tile TileType
		name string
		x, y int
	}{
		{TileWall, "TileWall", 5, 5},
		{TileTree, "TileTree", 10, 5},
		{TileFence, "TileFence", 15, 5},
		{TileDebris, "TileDebris", 20, 5},
	}

	for _, st := range solidTiles {
		m.SetTile(st.x, st.y, st.tile)

		tilePixelX := float64(st.x * TileSize)
		tilePixelY := float64(st.y * TileSize)

		// 1. Direct hit on solid tile
		if !m.IsColliding(tilePixelX+4, tilePixelY+4, 16, 16) {
			t.Errorf("FAIL: Expected collision on %s at tile (%d,%d)", st.name, st.x, st.y)
		}

		// 2. Overlap from top-left corner
		if !m.IsColliding(tilePixelX-8, tilePixelY-8, 16, 16) {
			t.Errorf("FAIL: Expected corner collision on %s at tile (%d,%d)", st.name, st.x, st.y)
		}

		// 3. Overlap from bottom-right corner
		if !m.IsColliding(tilePixelX+24, tilePixelY+24, 16, 16) {
			t.Errorf("FAIL: Expected corner collision on %s at tile (%d,%d)", st.name, st.x, st.y)
		}
	}

	floorTiles := []struct {
		tile TileType
		name string
		x, y int
	}{
		{TileGrass, "TileGrass", 5, 15},
		{TileDirt, "TileDirt", 10, 15},
		{TileWoodFloor, "TileWoodFloor", 15, 15},
		{TileAsphalt, "TileAsphalt", 20, 15},
		{TileConcrete, "TileConcrete", 5, 20},
		{TileTileFloor, "TileTileFloor", 10, 20},
	}

	for _, ft := range floorTiles {
		m.SetTile(ft.x, ft.y, ft.tile)

		tilePixelX := float64(ft.x * TileSize)
		tilePixelY := float64(ft.y * TileSize)

		// Entity inside floor tile must NOT collide
		if m.IsColliding(tilePixelX+8, tilePixelY+8, 16, 16) {
			t.Errorf("FAIL: Expected NO collision on floor tile %s at (%d,%d)", ft.name, ft.x, ft.y)
		}
	}

	// Boundary collision tests
	if !m.IsColliding(-1.0, 100.0, 16, 16) {
		t.Errorf("FAIL: Expected collision on negative X boundary")
	}
	if !m.IsColliding(100.0, -1.0, 16, 16) {
		t.Errorf("FAIL: Expected collision on negative Y boundary")
	}
	if !m.IsColliding(float64(30*TileSize)+1, 100.0, 16, 16) {
		t.Errorf("FAIL: Expected collision on max X boundary")
	}
	if !m.IsColliding(100.0, float64(30*TileSize)+1, 16, 16) {
		t.Errorf("FAIL: Expected collision on max Y boundary")
	}
}

// TestEmpirical_FOVRaycastingWallVsFence verifies that FOV raycasting is
// strictly blocked by TileWall, while penetrating TileFence, TileTree, TileDebris, and floors.
func TestEmpirical_FOVRaycastingWallVsFence(t *testing.T) {
	m := NewMap(40, 40)
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			m.SetTile(x, y, TileGrass)
		}
	}

	// Player at (20, 20)
	playerTileX, playerTileY := 20, 20
	playerPixelX := float64(playerTileX)*TileSize + 16.0
	playerPixelY := float64(playerTileY)*TileSize + 16.0

	// 1. Place TileWall to the East at (25, 20)
	m.SetTile(25, 20, TileWall)

	// 2. Place TileFence to the South at (20, 25)
	m.SetTile(20, 25, TileFence)

	// 3. Place TileTree to the North at (20, 15)
	m.SetTile(20, 15, TileTree)

	// 4. Place TileDebris to the West at (15, 20)
	m.SetTile(15, 20, TileDebris)

	// Calculate FOV with radius 10
	m.CalculateFOV(playerPixelX, playerPixelY, 10)

	// Check Player tile is visible
	if !m.Visible[playerTileY*m.Width+playerTileX] {
		t.Fatalf("FAIL: Player tile (%d,%d) is not visible", playerTileX, playerTileY)
	}

	// EAST (Wall):
	// Wall itself should be visible
	if !m.Visible[20*m.Width+25] {
		t.Errorf("FAIL: Wall tile at (25, 20) should be visible")
	}
	// Tiles behind Wall (26, 20), (27, 20), (28, 20) MUST NOT be visible (occluded by TileWall)
	for tx := 26; tx <= 29; tx++ {
		if m.Visible[20*m.Width+tx] {
			t.Errorf("FAIL: Tile behind wall at (%d, 20) should be occluded by TileWall, but is visible", tx)
		}
	}

	// SOUTH (Fence):
	// Fence itself is visible
	if !m.Visible[25*m.Width+20] {
		t.Errorf("FAIL: Fence tile at (20, 25) should be visible")
	}
	// Tiles behind Fence (20, 26), (20, 27), (20, 28) MUST be visible (penetrated by raycast)
	for ty := 26; ty <= 28; ty++ {
		if !m.Visible[ty*m.Width+20] {
			t.Errorf("FAIL: Tile behind fence at (20, %d) should be visible through Fence, but was blocked", ty)
		}
	}

	// NORTH (Tree):
	// Tree itself is visible
	if !m.Visible[15*m.Width+20] {
		t.Errorf("FAIL: Tree tile at (20, 15) should be visible")
	}
	// Tiles behind Tree (20, 14), (20, 13) should be visible (trees don't block vision)
	for ty := 13; ty <= 14; ty++ {
		if !m.Visible[ty*m.Width+20] {
			t.Errorf("FAIL: Tile behind tree at (20, %d) should be visible through Tree, but was blocked", ty)
		}
	}

	// WEST (Debris):
	// Debris itself is visible
	if !m.Visible[20*m.Width+15] {
		t.Errorf("FAIL: Debris tile at (15, 20) should be visible")
	}
	// Tiles behind Debris (14, 20), (13, 20) should be visible (debris doesn't block vision)
	for tx := 13; tx <= 14; tx++ {
		if !m.Visible[20*m.Width+tx] {
			t.Errorf("FAIL: Tile behind debris at (%d, 20) should be visible through Debris, but was blocked", tx)
		}
	}
}

// TestEmpirical_LootDistributionAndWalkability verifies that:
// 1. All 7 loot types ("food", "water", "weapon", "axe", "shotgun", "ammo", "armor") are present.
// 2. 100% of loot items are spawned on non-solid tiles.
func TestEmpirical_LootDistributionAndWalkability(t *testing.T) {
	requiredLoot := map[string]int{
		"food":    0,
		"water":   0,
		"weapon":  0,
		"axe":     0,
		"shotgun": 0,
		"ammo":    0,
		"armor":   0,
	}

	for seed := 0; seed < 20; seed++ {
		m := NewMap(100, 100)

		for lIdx, ls := range m.LootSpawns {
			requiredLoot[ls.Type]++

			tx := int(ls.X) / TileSize
			ty := int(ls.Y) / TileSize

			tile := m.GetTile(tx, ty)
			if tile.IsSolid() {
				t.Fatalf("FAIL: Loot #%d (%s) at (%.1f, %.1f) [tile %d,%d] is on solid tile %v (%s)",
					lIdx, ls.Type, ls.X, ls.Y, tx, ty, tile, tile.String())
			}
		}
	}

	for lootType, count := range requiredLoot {
		if count == 0 {
			t.Errorf("FAIL: Required loot type %s was never generated across 20 maps", lootType)
		} else {
			t.Logf("PASS: Loot type %-10s generated count = %4d", lootType, count)
		}
	}
}

// TestEmpirical_AllNewPropTileTypesGenerated verifies that all 6 new prop TileTypes
// (TileBench, TileChest, TileSculpture, TileBush, TileFlower, TileStone) are generated
// with non-zero occurrence across multiple map generations.
func TestEmpirical_AllNewPropTileTypesGenerated(t *testing.T) {
	newProps := []TileType{
		TileBench,
		TileChest,
		TileSculpture,
		TileBush,
		TileFlower,
		TileStone,
	}

	for iter := 0; iter < 20; iter++ {
		m := NewMap(100, 100)
		counts := make(map[TileType]int)
		for _, tile := range m.Tiles {
			counts[tile]++
		}

		for _, prop := range newProps {
			c := counts[prop]
			if c == 0 {
				t.Fatalf("FAIL: Iter %d: New prop %v (%s) was NOT generated (count = 0)", iter, prop, prop.String())
			}
			if iter == 0 {
				t.Logf("PASS: Iter 0: Prop %-12s (ID %d): count = %4d", prop.String(), prop, c)
			}
		}
	}
}

