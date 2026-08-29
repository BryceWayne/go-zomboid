package world

import (
	"math"
	"testing"
)

func TestTileTypeProperties(t *testing.T) {
	solidTiles := []TileType{
		TileWall, TileTree, TileFence, TileDebris, TileTent, TileElevationBlock, TileStump, TileSign,
		TileBench, TileChest, TileSculpture, TileStone,
	}
	for _, tile := range solidTiles {
		if !tile.IsSolid() {
			t.Errorf("Expected tile %v (%s) to be solid", tile, tile.String())
		}
	}

	nonSolidTiles := []TileType{
		TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor, TileRamp, TileMushroom,
		TileBush, TileFlower,
	}
	for _, tile := range nonSolidTiles {
		if tile.IsSolid() {
			t.Errorf("Expected tile %v (%s) to NOT be solid", tile, tile.String())
		}
	}

	// Vision blocking: only TileWall blocks vision
	if !TileWall.BlocksVision() {
		t.Errorf("Expected TileWall to block vision")
	}
	transparentTiles := []TileType{
		TileGrass, TileDirt, TileWoodFloor, TileTree, TileAsphalt, TileConcrete, TileTileFloor,
		TileFence, TileDebris, TileTent, TileElevationBlock, TileRamp, TileStump, TileMushroom, TileSign,
		TileBench, TileChest, TileSculpture, TileBush, TileFlower, TileStone,
	}
	for _, tile := range transparentTiles {
		if tile.BlocksVision() {
			t.Errorf("Expected tile %v (%s) to NOT block vision", tile, tile.String())
		}
	}

	// Floor types
	floorTiles := []TileType{TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor, TileRamp}
	for _, tile := range floorTiles {
		if !tile.IsFloor() {
			t.Errorf("Expected tile %v (%s) to be floor", tile, tile.String())
		}
	}
	verticalTiles := []TileType{
		TileWall, TileTree, TileFence, TileDebris, TileTent, TileElevationBlock, TileStump, TileMushroom, TileSign,
		TileBench, TileChest, TileSculpture, TileBush, TileFlower, TileStone,
	}
	for _, tile := range verticalTiles {
		if tile.IsFloor() {
			t.Errorf("Expected tile %v (%s) to NOT be floor", tile, tile.String())
		}
	}

	// String representations
	for tile := TileGrass; tile <= TileStone; tile++ {
		if tile.String() == "Unknown" {
			t.Errorf("Expected known string for tile %d, got Unknown", tile)
		}
	}

	expectedStrings := map[TileType]string{
		TileBench:     "Bench",
		TileChest:     "Chest",
		TileSculpture: "Sculpture",
		TileBush:      "Bush",
		TileFlower:    "Flower",
		TileStone:     "Stone",
	}
	for tile, exp := range expectedStrings {
		if tile.String() != exp {
			t.Errorf("Expected tile %d string to be %q, got %q", tile, exp, tile.String())
		}
	}
}

func TestNewMapProceduralTown(t *testing.T) {
	m := NewMap(100, 100)

	if m.Width != 100 || m.Height != 100 {
		t.Fatalf("Expected 100x100 map, got %dx%d", m.Width, m.Height)
	}

	if len(m.Tiles) != 10000 {
		t.Fatalf("Expected 10000 tiles, got %d", len(m.Tiles))
	}

	// Verify perimeter walls
	for x := 0; x < 100; x++ {
		if m.GetTile(x, 0) != TileWall || m.GetTile(x, 99) != TileWall {
			t.Errorf("Perimeter wall missing at x=%d", x)
		}
	}
	for y := 0; y < 100; y++ {
		if m.GetTile(0, y) != TileWall || m.GetTile(99, y) != TileWall {
			t.Errorf("Perimeter wall missing at y=%d", y)
		}
	}

	// Verify all tile types are represented
	tileCount := make(map[TileType]int)
	for _, tile := range m.Tiles {
		tileCount[tile]++
	}

	expectedTiles := []TileType{
		TileGrass, TileWall, TileDirt, TileWoodFloor, TileTree,
		TileAsphalt, TileConcrete, TileTileFloor, TileFence, TileDebris,
	}
	for _, exp := range expectedTiles {
		count := tileCount[exp]
		if count == 0 {
			t.Errorf("Expected map to contain tile type %v (%s), but found 0", exp, exp.String())
		}
	}

	// Verify buildings
	if len(m.Buildings) < 4 {
		t.Errorf("Expected at least 4 buildings, got %d", len(m.Buildings))
	}
	buildingTypes := make(map[BuildingType]bool)
	for _, b := range m.Buildings {
		buildingTypes[b.Type] = true
	}
	if !buildingTypes[BuildingResidential] {
		t.Errorf("Expected BuildingResidential in buildings")
	}
	if !buildingTypes[BuildingGrocery] {
		t.Errorf("Expected BuildingGrocery in buildings")
	}
	if !buildingTypes[BuildingPolice] {
		t.Errorf("Expected BuildingPolice in buildings")
	}
	if !buildingTypes[BuildingPharmacy] {
		t.Errorf("Expected BuildingPharmacy in buildings")
	}
	if !buildingTypes[BuildingWarehouse] {
		t.Errorf("Expected BuildingWarehouse in buildings")
	}
}

func TestPlayerSafeSpawn(t *testing.T) {
	m := NewMap(100, 100)

	px := int(m.PlayerSpawn.X) / TileSize
	py := int(m.PlayerSpawn.Y) / TileSize

	if px <= 0 || px >= m.Width-1 || py <= 0 || py >= m.Height-1 {
		t.Fatalf("Player spawn (%f, %f) is out of bounds", m.PlayerSpawn.X, m.PlayerSpawn.Y)
	}

	spawnTile := m.GetTile(px, py)
	if spawnTile.IsSolid() {
		t.Errorf("Player spawn is on solid tile: %v (%s)", spawnTile, spawnTile.String())
	}

	// Check distance to all zombie spawns
	for i, zs := range m.ZombieSpawns {
		dx := zs.X - m.PlayerSpawn.X
		dy := zs.Y - m.PlayerSpawn.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1400.0 {
			t.Errorf("Zombie spawn %d (%f, %f) is too close to player spawn (%f, %f): dist=%f < 1400",
				i, zs.X, zs.Y, m.PlayerSpawn.X, m.PlayerSpawn.Y, dist)
		}
	}
}

func TestContextualLootSpawns(t *testing.T) {
	m := NewMap(100, 100)

	if len(m.LootSpawns) < 10 {
		t.Fatalf("Expected at least 10 loot spawns, got %d", len(m.LootSpawns))
	}

	lootTypes := make(map[string]int)
	for i, ls := range m.LootSpawns {
		lootTypes[ls.Type]++

		// Verify coordinates are on walkable tiles
		tx := int(ls.X) / TileSize
		ty := int(ls.Y) / TileSize
		tile := m.GetTile(tx, ty)
		if tile.IsSolid() {
			t.Errorf("Loot %d (%s) at (%f, %f) is on solid tile %v (%s)",
				i, ls.Type, ls.X, ls.Y, tile, tile.String())
		}
	}

	// Check variety of loot items
	requiredTypes := []string{"food", "water", "weapon", "axe", "shotgun", "ammo", "armor"}
	for _, req := range requiredTypes {
		if lootTypes[req] == 0 {
			t.Errorf("Missing contextual loot type: %s", req)
		}
	}
}

func TestZombieSpawnsNoTrapping(t *testing.T) {
	m := NewMap(100, 100)

	if len(m.ZombieSpawns) < 50 {
		t.Fatalf("Expected at least 50 zombie spawns, got %d", len(m.ZombieSpawns))
	}

	for i, zs := range m.ZombieSpawns {
		tx := int(zs.X) / TileSize
		ty := int(zs.Y) / TileSize
		tile := m.GetTile(tx, ty)
		if tile.IsSolid() {
			t.Errorf("Zombie %d at (%f, %f) is trapped on solid tile %v (%s)",
				i, zs.X, zs.Y, tile, tile.String())
		}
	}
}

func TestCollisionDetection(t *testing.T) {
	m := NewMap(50, 50)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Test non-solid grass
	if m.IsColliding(256, 256, 64, 64) {
		t.Errorf("Expected no collision on grass")
	}

	// Test solid wall at (2,2) -> (256,256) to (383,383)
	m.SetTile(2, 2, TileWall)
	if !m.IsColliding(240, 240, 40, 40) {
		t.Errorf("Expected collision overlapping TileWall")
	}

	// Test solid tree at (4,4) -> (512, 512)
	m.SetTile(4, 4, TileTree)
	if !m.IsColliding(512, 512, 64, 64) {
		t.Errorf("Expected collision on TileTree")
	}

	// Test solid fence at (6,6) -> (768, 768)
	m.SetTile(6, 6, TileFence)
	if !m.IsColliding(768, 768, 64, 64) {
		t.Errorf("Expected collision on TileFence")
	}

	// Test solid debris at (8,8) -> (1024, 1024)
	m.SetTile(8, 8, TileDebris)
	if !m.IsColliding(1024, 1024, 64, 64) {
		t.Errorf("Expected collision on TileDebris")
	}

	// Out of bounds collision
	if !m.IsColliding(-10, -10, 64, 64) {
		t.Errorf("Expected collision out of bounds negative")
	}
	if !m.IsColliding(8000, 8000, 64, 64) {
		t.Errorf("Expected collision out of bounds positive")
	}
}

func TestFOVAndOcclusion(t *testing.T) {
	m := NewMap(50, 50)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Place a wall at (15, 10)
	m.SetTile(15, 10, TileWall)
	// Place a fence at (10, 15)
	m.SetTile(10, 15, TileFence)

	// Calculate FOV from (10, 10)
	m.CalculateFOV(10.0*TileSize+64, 10.0*TileSize+64, 10)

	// Player tile must be visible
	if !m.Visible[10*m.Width+10] {
		t.Errorf("Expected player tile to be visible")
	}

	// Wall at (15, 10) should be visible
	if !m.Visible[10*m.Width+15] {
		t.Errorf("Expected wall at (15,10) to be visible")
	}

	// Tile behind wall at (17, 10) should NOT be visible
	if m.Visible[10*m.Width+17] {
		t.Errorf("Expected tile behind wall (17,10) to be occluded")
	}

	// Tile behind fence at (10, 17) SHOULD be visible (fence does not block vision)
	if !m.Visible[17*m.Width+10] {
		t.Errorf("Expected tile behind fence (10,17) to be visible (fences allow vision)")
	}
}

func TestSmallFallbackMap(t *testing.T) {
	m := NewMap(20, 20)
	if m.Width != 20 || m.Height != 20 {
		t.Fatalf("Expected 20x20 map, got %dx%d", m.Width, m.Height)
	}
	if len(m.LootSpawns) == 0 {
		t.Errorf("Expected fallback map to have starter loot spawns")
	}
	px := int(m.PlayerSpawn.X) / TileSize
	py := int(m.PlayerSpawn.Y) / TileSize
	if m.GetTile(px, py).IsSolid() {
		t.Errorf("Fallback player spawn is on solid tile")
	}
}

func TestNewMapProceduralPropsGeneration(t *testing.T) {
	m := NewMap(100, 100)

	newPropTiles := []TileType{
		TileBench,
		TileChest,
		TileSculpture,
		TileBush,
		TileFlower,
		TileStone,
	}

	counts := make(map[TileType]int)
	for _, tile := range m.Tiles {
		counts[tile]++
	}

	for _, pt := range newPropTiles {
		count := counts[pt]
		if count == 0 {
			t.Errorf("Expected procedural map to generate new prop tile %v (%s), but got 0", pt, pt.String())
		} else {
			t.Logf("Generated prop %-12s (ID %d): count = %d", pt.String(), pt, count)
		}
	}
}

func TestCollisionAndFOVNewProps(t *testing.T) {
	m := NewMap(40, 40)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Place solid props
	m.SetTile(5, 5, TileBench)
	m.SetTile(7, 5, TileChest)
	m.SetTile(9, 5, TileSculpture)
	m.SetTile(11, 5, TileStone)

	// Place non-solid props
	m.SetTile(13, 5, TileBush)
	m.SetTile(15, 5, TileFlower)

	// Verify collision for solid props
	solidPoints := []Point{{5, 5}, {7, 5}, {9, 5}, {11, 5}}
	for _, p := range solidPoints {
		px := float64(p.X * TileSize)
		py := float64(p.Y * TileSize)
		if !m.IsColliding(px+10, py+10, 16, 16) {
			t.Errorf("Expected collision on solid prop %v at (%d,%d)", m.GetTile(p.X, p.Y), p.X, p.Y)
		}
	}

	// Verify no collision for non-solid props
	nonSolidPoints := []Point{{13, 5}, {15, 5}}
	for _, p := range nonSolidPoints {
		px := float64(p.X * TileSize)
		py := float64(p.Y * TileSize)
		if m.IsColliding(px+10, py+10, 16, 16) {
			t.Errorf("Expected NO collision on non-solid prop %v at (%d,%d)", m.GetTile(p.X, p.Y), p.X, p.Y)
		}
	}

	// Verify FOV penetration through all new props
	playerX := 20.0*TileSize + 64.0
	playerY := 20.0*TileSize + 64.0
	m.SetTile(20, 15, TileBench)
	m.SetTile(20, 25, TileSculpture)
	m.SetTile(15, 20, TileChest)
	m.SetTile(25, 20, TileStone)

	m.CalculateFOV(playerX, playerY, 10)

	// All props and the tiles behind them must be visible
	behindCoords := []Point{
		{20, 14}, // Behind bench
		{20, 26}, // Behind sculpture
		{14, 20}, // Behind chest
		{26, 20}, // Behind stone
	}
	for _, bc := range behindCoords {
		if !m.Visible[bc.Y*m.Width+bc.X] {
			t.Errorf("Expected tile at (%d,%d) behind prop to be visible in FOV", bc.X, bc.Y)
		}
	}
}

