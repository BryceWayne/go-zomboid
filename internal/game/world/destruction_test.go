package world

import (
	"testing"
)

func TestDestruction_TileMaxDurability(t *testing.T) {
	m := NewMap(50, 50)

	expectedDurabilities := map[TileType]int{
		TileFence:          2,
		TileTree:           3,
		TileStump:          2,
		TileBench:          2,
		TileWall:           3,
		TileGrass:          0,
		TileDirt:           0,
		TileWoodFloor:      0,
		TileConcrete:       0,
		TileAsphalt:        0,
		TileTileFloor:      0,
		TileDebris:         0,
		TileChest:          0,
		TileElevationBlock: 0,
		TileRamp:           0,
	}

	for tileType, expected := range expectedDurabilities {
		actual := m.GetTileMaxDurability(tileType)
		if actual != expected {
			t.Errorf("GetTileMaxDurability(%v) = %d, expected %d", tileType, actual, expected)
		}
	}
}

func TestDestruction_IsDestructible(t *testing.T) {
	m := NewMap(40, 40)

	// Perimeter boundaries MUST be indestructible
	if m.IsDestructible(0, 0) {
		t.Errorf("Perimeter corner (0,0) should not be destructible")
	}
	if m.IsDestructible(0, 20) {
		t.Errorf("Perimeter West (0,20) should not be destructible")
	}
	if m.IsDestructible(39, 20) {
		t.Errorf("Perimeter East (39,20) should not be destructible")
	}
	if m.IsDestructible(20, 0) {
		t.Errorf("Perimeter North (20,0) should not be destructible")
	}
	if m.IsDestructible(20, 39) {
		t.Errorf("Perimeter South (20,39) should not be destructible")
	}

	// Out of bounds coordinates MUST not be destructible
	if m.IsDestructible(-5, 10) || m.IsDestructible(10, -5) || m.IsDestructible(50, 20) || m.IsDestructible(20, 50) {
		t.Errorf("Out-of-bounds coordinates should not be destructible")
	}

	// Interior destructible props
	m.SetTile(10, 10, TileFence)
	if !m.IsDestructible(10, 10) {
		t.Errorf("Interior TileFence at (10,10) must be destructible")
	}

	m.SetTile(11, 10, TileWall)
	if !m.IsDestructible(11, 10) {
		t.Errorf("Interior TileWall at (11,10) must be destructible")
	}

	m.SetTile(12, 10, TileTree)
	if !m.IsDestructible(12, 10) {
		t.Errorf("Interior TileTree at (12,10) must be destructible")
	}

	m.SetTile(13, 10, TileStump)
	if !m.IsDestructible(13, 10) {
		t.Errorf("Interior TileStump at (13,10) must be destructible")
	}

	m.SetTile(14, 10, TileBench)
	if !m.IsDestructible(14, 10) {
		t.Errorf("Interior TileBench at (14,10) must be destructible")
	}

	// Interior non-destructible tiles
	m.SetTile(15, 10, TileGrass)
	if m.IsDestructible(15, 10) {
		t.Errorf("TileGrass should not be destructible")
	}

	m.SetTile(16, 10, TileDirt)
	if m.IsDestructible(16, 10) {
		t.Errorf("TileDirt should not be destructible")
	}

	m.SetTile(17, 10, TileConcrete)
	if m.IsDestructible(17, 10) {
		t.Errorf("TileConcrete should not be destructible")
	}

	m.SetTile(18, 10, TileChest)
	if m.IsDestructible(18, 10) {
		t.Errorf("TileChest should not be destructible")
	}
}

func TestDestruction_TileDurabilityDegradation(t *testing.T) {
	m := NewMap(40, 40)
	tx, ty := 15, 15
	m.SetTile(tx, ty, TileFence)

	// Check initial durability
	dur := m.GetTileDurability(tx, ty)
	if dur != 2 {
		t.Fatalf("Expected initial fence durability = 2, got %d", dur)
	}

	// Hit 1: 1 damage -> durability goes from 2 to 1, not destroyed
	destroyed, dropType := m.DamageTile(tx, ty, 1)
	if destroyed {
		t.Fatalf("Hit 1 should not destroy fence with 2 HP")
	}
	if dropType != "" {
		t.Fatalf("Hit 1 should not return dropType, got %q", dropType)
	}
	dur = m.GetTileDurability(tx, ty)
	if dur != 1 {
		t.Fatalf("Expected fence durability after 1 damage = 1, got %d", dur)
	}
	if m.GetTile(tx, ty) != TileFence {
		t.Fatalf("Tile should still be TileFence after 1 hit, got %v", m.GetTile(tx, ty))
	}

	// Hit 2: 1 damage -> durability goes from 1 to 0, destroyed!
	destroyed, dropType = m.DamageTile(tx, ty, 1)
	if !destroyed {
		t.Fatalf("Hit 2 should destroy fence with 1 remaining HP")
	}
	if dropType != "wood" {
		t.Fatalf("Expected dropType 'wood', got %q", dropType)
	}
	if m.GetTile(tx, ty) != TileGrass {
		t.Fatalf("Destroyed fence should be replaced with TileGrass, got %v", m.GetTile(tx, ty))
	}
	if m.IsDestructible(tx, ty) {
		t.Fatalf("Destroyed tile (now Grass) should no longer be destructible")
	}
}

func TestDestruction_WallDurabilityAndFloorReplacement(t *testing.T) {
	m := NewMap(40, 40)
	tx, ty := 20, 20
	m.SetTile(tx, ty, TileWall)

	// Wall has 3 max durability
	dur := m.GetTileDurability(tx, ty)
	if dur != 3 {
		t.Fatalf("Expected initial wall durability = 3, got %d", dur)
	}

	// Deal 2 damage (e.g. Axe hit)
	destroyed, dropType := m.DamageTile(tx, ty, 2)
	if destroyed {
		t.Fatalf("Wall should survive 2 damage, remaining HP should be 1")
	}
	if dropType != "" {
		t.Fatalf("DropType should be empty before destruction, got %q", dropType)
	}
	if m.GetTileDurability(tx, ty) != 1 {
		t.Fatalf("Expected remaining wall durability = 1, got %d", m.GetTileDurability(tx, ty))
	}

	// Deal 1 damage (e.g. Club hit) -> destroyed!
	destroyed, dropType = m.DamageTile(tx, ty, 1)
	if !destroyed {
		t.Fatalf("Wall should be destroyed on final hit")
	}
	if dropType != "wood" {
		t.Fatalf("Expected dropType 'wood', got %q", dropType)
	}
	if m.GetTile(tx, ty) != TileWoodFloor {
		t.Fatalf("Destroyed interior wall should be replaced with TileWoodFloor, got %v", m.GetTile(tx, ty))
	}
}

func TestDestruction_TreeStumpBenchDurability(t *testing.T) {
	m := NewMap(40, 40)

	// Tree (HP = 3)
	m.SetTile(10, 10, TileTree)
	if m.GetTileDurability(10, 10) != 3 {
		t.Fatalf("Tree max durability should be 3, got %d", m.GetTileDurability(10, 10))
	}
	destroyed, drop := m.DamageTile(10, 10, 3)
	if !destroyed || drop != "wood" || m.GetTile(10, 10) != TileGrass {
		t.Fatalf("Tree destruction failed: destroyed=%v, drop=%q, tile=%v", destroyed, drop, m.GetTile(10, 10))
	}

	// Stump (HP = 2)
	m.SetTile(11, 10, TileStump)
	if m.GetTileDurability(11, 10) != 2 {
		t.Fatalf("Stump max durability should be 2, got %d", m.GetTileDurability(11, 10))
	}
	destroyed, drop = m.DamageTile(11, 10, 2)
	if !destroyed || drop != "wood" || m.GetTile(11, 10) != TileGrass {
		t.Fatalf("Stump destruction failed: destroyed=%v, drop=%q, tile=%v", destroyed, drop, m.GetTile(11, 10))
	}

	// Bench (HP = 2)
	m.SetTile(12, 10, TileBench)
	if m.GetTileDurability(12, 10) != 2 {
		t.Fatalf("Bench max durability should be 2, got %d", m.GetTileDurability(12, 10))
	}
	destroyed, drop = m.DamageTile(12, 10, 2)
	if !destroyed || drop != "wood" || m.GetTile(12, 10) != TileGrass {
		t.Fatalf("Bench destruction failed: destroyed=%v, drop=%q, tile=%v", destroyed, drop, m.GetTile(12, 10))
	}
}

func TestDestruction_PerimeterIndestructible(t *testing.T) {
	m := NewMap(40, 40)

	// Perimeter boundaries (x=0, y=0, x=39, y=39)
	perimeterCoords := []Point{
		{X: 0, Y: 0},
		{X: 0, Y: 20},
		{X: 39, Y: 20},
		{X: 20, Y: 0},
		{X: 20, Y: 39},
	}

	for _, pt := range perimeterCoords {
		destroyed, drop := m.DamageTile(pt.X, pt.Y, 999)
		if destroyed {
			t.Errorf("Perimeter tile at (%d,%d) was illegally destroyed", pt.X, pt.Y)
		}
		if drop != "" {
			t.Errorf("Perimeter tile at (%d,%d) returned drop %q", pt.X, pt.Y, drop)
		}
		if m.GetTile(pt.X, pt.Y) != TileWall {
			t.Errorf("Perimeter tile at (%d,%d) changed tile type to %v", pt.X, pt.Y, m.GetTile(pt.X, pt.Y))
		}
	}
}

func TestDestruction_SolidityAndVisionCleared(t *testing.T) {
	m := NewMap(40, 40)
	tx, ty := 10, 10
	m.SetTile(tx, ty, TileWall)

	// 1. Before destruction: tile is solid, blocks vision, and collides
	if !m.GetTile(tx, ty).IsSolid() {
		t.Fatalf("Wall should be solid before destruction")
	}
	if !m.BlocksVision(tx, ty) {
		t.Fatalf("Wall should block vision before destruction")
	}
	rectX := float64(tx*TileSize + 10)
	rectY := float64(ty*TileSize + 10)
	if !m.IsColliding(rectX, rectY, 32, 32) {
		t.Fatalf("IsColliding should return true on wall before destruction")
	}

	// Check FOV does NOT penetrate wall
	m.CalculateFOV(float64((tx-1)*TileSize+64), float64(ty*TileSize+64), 10)
	targetIdx := ty*m.Width + (tx + 1)
	if m.Visible[targetIdx] {
		t.Fatalf("Target tile behind wall should NOT be visible before destruction")
	}

	// 2. Destroy the wall
	destroyed, drop := m.DamageTile(tx, ty, 3)
	if !destroyed || drop != "wood" {
		t.Fatalf("Wall destruction failed: destroyed=%v, drop=%q", destroyed, drop)
	}

	// 3. After destruction: tile is TileWoodFloor, non-solid, non-vision-blocking, and clear collision
	if m.GetTile(tx, ty).IsSolid() {
		t.Fatalf("Destroyed wall (now floor) should NOT be solid")
	}
	if m.BlocksVision(tx, ty) {
		t.Fatalf("Destroyed wall (now floor) should NOT block vision")
	}
	if m.IsColliding(rectX, rectY, 32, 32) {
		t.Fatalf("IsColliding should return false on destroyed wall")
	}

	// Check FOV NOW penetrates through destroyed tile
	for i := range m.Visible {
		m.Visible[i] = false
	}
	m.CalculateFOV(float64((tx-1)*TileSize+64), float64(ty*TileSize+64), 10)
	if !m.Visible[targetIdx] {
		t.Fatalf("Target tile behind destroyed wall MUST be visible after destruction")
	}
}

func TestDestruction_ZeroOrNegativeDamage(t *testing.T) {
	m := NewMap(40, 40)
	m.SetTile(10, 10, TileFence)

	destroyed, drop := m.DamageTile(10, 10, 0)
	if destroyed || drop != "" || m.GetTileDurability(10, 10) != 2 {
		t.Fatalf("0 damage should do nothing")
	}

	destroyed, drop = m.DamageTile(10, 10, -5)
	if destroyed || drop != "" || m.GetTileDurability(10, 10) != 2 {
		t.Fatalf("Negative damage should do nothing")
	}
}

func TestDestruction_ExcessDamageSingleHit(t *testing.T) {
	m := NewMap(40, 40)
	m.SetTile(10, 10, TileFence)

	// Deal 10 damage to a 2 HP fence
	destroyed, drop := m.DamageTile(10, 10, 10)
	if !destroyed || drop != "wood" {
		t.Fatalf("Excess damage should destroy fence and drop wood")
	}
	if m.GetTile(10, 10) != TileGrass {
		t.Fatalf("Destroyed fence should be TileGrass, got %v", m.GetTile(10, 10))
	}
}
