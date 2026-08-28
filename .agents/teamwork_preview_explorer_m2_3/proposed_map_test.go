package world_proposed

import (
	"math"
	"testing"
)

func TestTileTypeProperties(t *testing.T) {
	solidTiles := []TileType{TileWall, TileTree, TileFence, TileDebris}
	for _, tile := range solidTiles {
		if !tile.IsSolid() {
			t.Errorf("Expected tile %v to be solid", tile)
		}
	}

	nonSolidTiles := []TileType{TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor}
	for _, tile := range nonSolidTiles {
		if tile.IsSolid() {
			t.Errorf("Expected tile %v to NOT be solid", tile)
		}
	}

	// Vision blocking
	if !TileWall.BlocksVision() {
		t.Errorf("Expected TileWall to block vision")
	}
	if TileFence.BlocksVision() {
		t.Errorf("Expected TileFence to NOT block vision")
	}
	if TileDebris.BlocksVision() {
		t.Errorf("Expected TileDebris to NOT block vision")
	}
	if TileGrass.BlocksVision() {
		t.Errorf("Expected TileGrass to NOT block vision")
	}

	// Floor types
	floorTiles := []TileType{TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor}
	for _, tile := range floorTiles {
		if !tile.IsFloor() {
			t.Errorf("Expected tile %v to be floor", tile)
		}
	}
	verticalTiles := []TileType{TileWall, TileTree, TileFence, TileDebris}
	for _, tile := range verticalTiles {
		if tile.IsFloor() {
			t.Errorf("Expected tile %v to NOT be floor", tile)
		}
	}

	// String representations
	for tile := TileGrass; tile <= TileDebris; tile++ {
		if tile.String() == "Unknown" {
			t.Errorf("Expected known string for tile %d, got Unknown", tile)
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
		if dist < 350.0 {
			t.Errorf("Zombie spawn %d (%f, %f) is too close to player spawn (%f, %f): dist=%f < 350",
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
	if m.IsColliding(64, 64, 16, 16) {
		t.Errorf("Expected no collision on grass")
	}

	// Test solid wall at (2,2) -> (64,64) to (95,95)
	m.SetTile(2, 2, TileWall)
	if !m.IsColliding(60, 60, 10, 10) {
		t.Errorf("Expected collision overlapping TileWall")
	}

	// Test solid tree at (4,4)
	m.SetTile(4, 4, TileTree)
	if !m.IsColliding(128, 128, 16, 16) {
		t.Errorf("Expected collision on TileTree")
	}

	// Test solid fence at (6,6)
	m.SetTile(6, 6, TileFence)
	if !m.IsColliding(192, 192, 16, 16) {
		t.Errorf("Expected collision on TileFence")
	}

	// Test solid debris at (8,8)
	m.SetTile(8, 8, TileDebris)
	if !m.IsColliding(256, 256, 16, 16) {
		t.Errorf("Expected collision on TileDebris")
	}

	// Out of bounds collision
	if !m.IsColliding(-10, -10, 16, 16) {
		t.Errorf("Expected collision out of bounds negative")
	}
	if !m.IsColliding(2000, 2000, 16, 16) {
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
	m.CalculateFOV(10.0*TileSize+16, 10.0*TileSize+16, 10)

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
