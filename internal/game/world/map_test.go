package world

import (
	"testing"
)

func TestNewMap(t *testing.T) {
	m := NewMap(50, 50)
	
	if m.Width != 50 || m.Height != 50 {
		t.Errorf("Expected map dimensions 50x50, got %dx%d", m.Width, m.Height)
	}
	
	if len(m.Tiles) != 50*50 {
		t.Errorf("Expected %d tiles, got %d", 50*50, len(m.Tiles))
	}

	// Boundary walls
	if m.GetTile(0, 0) != TileWall {
		t.Errorf("Expected boundary tile (0,0) to be a wall")
	}
}

func TestIsColliding(t *testing.T) {
	m := NewMap(10, 10)
	// Clear the map just for predictable testing
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Add a wall at (2, 2)
	m.SetTile(2, 2, TileWall)
	
	// A bounding box clearly outside the wall shouldn't collide
	// 32 pixels per tile. (0,0) to (31,31) is tile (0,0)
	// Tile (2,2) is (64,64) to (95,95)
	
	if m.IsColliding(0, 0, 16, 16) {
		t.Errorf("Expected no collision at (0,0)")
	}

	// A bounding box overlapping the wall should collide
	if !m.IsColliding(60, 60, 10, 10) {
		t.Errorf("Expected collision at (60,60,10,10) overlapping tile (2,2)")
	}
}
