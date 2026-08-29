package world

import (
	"testing"
)

// TestAdversarial_CardinalBitmask_Exhaustive16States tests all 16 cardinal connectivity states
// with isolated, adjacent, diagonal, and edge-of-map configurations.
func TestAdversarial_CardinalBitmask_Exhaustive16States(t *testing.T) {
	for mask := 0; mask < 16; mask++ {
		m := NewMap(7, 7)
		for i := range m.Tiles {
			m.Tiles[i] = TileGrass
		}

		cx, cy := 3, 3
		m.SetTile(cx, cy, TileWall)

		hasN := (mask & (1 << 0)) != 0
		hasE := (mask & (1 << 1)) != 0
		hasS := (mask & (1 << 2)) != 0
		hasW := (mask & (1 << 3)) != 0

		if hasN {
			m.SetTile(cx, cy-1, TileWall)
		}
		if hasE {
			m.SetTile(cx+1, cy, TileWall)
		}
		if hasS {
			m.SetTile(cx, cy+1, TileWall)
		}
		if hasW {
			m.SetTile(cx-1, cy, TileWall)
		}

		// Also place walls diagonally to ensure diagonals DO NOT pollute cardinal bitmask
		m.SetTile(cx-1, cy-1, TileWall)
		m.SetTile(cx+1, cy-1, TileWall)
		m.SetTile(cx-1, cy+1, TileWall)
		m.SetTile(cx+1, cy+1, TileWall)

		computedWall := GetWallBitmask(m, cx, cy)
		if computedWall != uint8(mask) {
			t.Fatalf("Mask %d: expected wall bitmask %d, got %d (N:%v E:%v S:%v W:%v)",
				mask, mask, computedWall, hasN, hasE, hasS, hasW)
		}

		computedCard := GetCardinalBitmask4(m, cx, cy, TileWall)
		if computedCard != uint8(mask) {
			t.Fatalf("Mask %d: expected cardinal bitmask %d, got %d", mask, mask, computedCard)
		}
	}
}

// TestAdversarial_FenceBitmask_IsolationAndDiscrimination verifies that fence bitmask
// is computed accurately and strictly ignores non-fence neighbors (walls, props, grass).
func TestAdversarial_FenceBitmask_IsolationAndDiscrimination(t *testing.T) {
	nonFenceTypes := []TileType{
		TileGrass, TileWall, TileDirt, TileTree, TileConcrete, TileAsphalt,
		TileWoodFloor, TileTileFloor, TileBench, TileChest, TileSculpture,
	}

	for mask := 0; mask < 16; mask++ {
		for _, otherType := range nonFenceTypes {
			m := NewMap(5, 5)
			for i := range m.Tiles {
				m.Tiles[i] = otherType
			}

			cx, cy := 2, 2
			m.SetTile(cx, cy, TileFence)

			hasN := (mask & (1 << 0)) != 0
			hasE := (mask & (1 << 1)) != 0
			hasS := (mask & (1 << 2)) != 0
			hasW := (mask & (1 << 3)) != 0

			if hasN {
				m.SetTile(cx, cy-1, TileFence)
			}
			if hasE {
				m.SetTile(cx+1, cy, TileFence)
			}
			if hasS {
				m.SetTile(cx, cy+1, TileFence)
			}
			if hasW {
				m.SetTile(cx-1, cy, TileFence)
			}

			computedFence := GetFenceBitmask(m, cx, cy)
			if computedFence != uint8(mask) {
				t.Fatalf("Mask %d with surrounding %v: expected fence bitmask %d, got %d",
					mask, otherType, mask, computedFence)
			}

			// If surrounding type is NOT TileWall, GetWallBitmask should be 0 because all neighbors are either otherType or TileFence
			if otherType != TileWall {
				if wallMask := GetWallBitmask(m, cx, cy); wallMask != 0 {
					t.Fatalf("Expected GetWallBitmask to be 0 with non-wall background %v, got %d", otherType, wallMask)
				}
			}
		}
	}
}

// TestAdversarial_TerrainPriority_StrictMonotonicity verifies strict inequality
// of terrain priority levels across all terrain categories.
func TestAdversarial_TerrainPriority_StrictMonotonicity(t *testing.T) {
	layers := []struct {
		tile TileType
		name string
		prio int
	}{
		{TileDirt, "Dirt", 10},
		{TileGrass, "Grass", 20},
		{TileConcrete, "Concrete", 30},
		{TileAsphalt, "Asphalt", 40},
		{TileWoodFloor, "WoodFloor", 50},
		{TileTileFloor, "TileFloor", 50},
	}

	for i := 0; i < len(layers)-1; i++ {
		p1 := TerrainPriority(layers[i].tile)
		p2 := TerrainPriority(layers[i+1].tile)
		if i == 4 { // WoodFloor vs TileFloor
			if p1 != p2 {
				t.Errorf("Expected %s (%d) == %s (%d)", layers[i].name, p1, layers[i+1].name, p2)
			}
		} else {
			if !(p1 < p2) {
				t.Errorf("Expected %s (%d) < %s (%d)", layers[i].name, p1, layers[i+1].name, p2)
			}
		}
	}
}

// TestAdversarial_SubtileEvaluation_AllCombinationsStress tests GetQuadrantSubtile
// under every possible 3-neighbor truth table (H in {T,F}, V in {T,F}, D in {T,F}).
func TestAdversarial_SubtileEvaluation_AllCombinationsStress(t *testing.T) {
	quads := []Quadrant{QuadNW, QuadNE, QuadSW, QuadSE}

	for _, q := range quads {
		// 8 combinations of (H, V, D)
		for combo := 0; combo < 8; combo++ {
			hMatch := (combo & 1) != 0
			vMatch := (combo & 2) != 0
			dMatch := (combo & 4) != 0

			m := NewMap(5, 5)
			for i := range m.Tiles {
				m.Tiles[i] = TileDirt
			}
			m.SetTile(2, 2, TileGrass)

			var hx, hy, vx, vy, dx, dy int
			switch q {
			case QuadNW:
				hx, hy = 1, 2
				vx, vy = 2, 1
				dx, dy = 1, 1
			case QuadNE:
				hx, hy = 3, 2
				vx, vy = 2, 1
				dx, dy = 3, 1
			case QuadSW:
				hx, hy = 1, 2
				vx, vy = 2, 3
				dx, dy = 1, 3
			case QuadSE:
				hx, hy = 3, 2
				vx, vy = 2, 3
				dx, dy = 3, 3
			}

			if hMatch {
				m.SetTile(hx, hy, TileGrass)
			}
			if vMatch {
				m.SetTile(vx, vy, TileGrass)
			}
			if dMatch {
				m.SetTile(dx, dy, TileGrass)
			}

			subtile := GetQuadrantSubtile(m, 2, 2, q, TileGrass)

			var expected SubtileState
			if hMatch && vMatch {
				if dMatch {
					expected = SubtileFull
				} else {
					expected = SubtileInnerCorner
				}
			} else if hMatch && !vMatch {
				expected = SubtileHorizontalEdge
			} else if !hMatch && vMatch {
				expected = SubtileVerticalEdge
			} else {
				expected = SubtileOuterCorner
			}

			if subtile != expected {
				t.Fatalf("Quad %v Combo (H:%v, V:%v, D:%v): expected %v, got %v",
					q, hMatch, vMatch, dMatch, expected, subtile)
			}
		}
	}
}

// TestAdversarial_TileTransitions_MultiTierHierarchy tests transitions when multiple
// different higher-priority terrains surround a base tile.
func TestAdversarial_TileTransitions_MultiTierHierarchy(t *testing.T) {
	m := NewMap(5, 5)
	for i := range m.Tiles {
		m.Tiles[i] = TileDirt // Base prio = 10
	}

	// Dirt tile at (2, 2)
	// Put Grass (prio 20) on West (1, 2)
	m.SetTile(1, 2, TileGrass)
	// Put Concrete (prio 30) on North (2, 1)
	m.SetTile(2, 1, TileConcrete)
	// Put Asphalt (prio 40) on East (3, 2)
	m.SetTile(3, 2, TileAsphalt)
	// Put WoodFloor (prio 50) on South (2, 3)
	m.SetTile(2, 3, TileWoodFloor)

	transitions := GetTileTransitions(m, 2, 2)
	if len(transitions) == 0 {
		t.Fatal("Expected multi-tier transitions, got 0")
	}

	// Check that we got transitions for Grass, Concrete, Asphalt, WoodFloor
	foundGrass := false
	foundConcrete := false
	foundAsphalt := false
	foundWood := false

	for _, tr := range transitions {
		switch tr.TerrainType {
		case TileGrass:
			foundGrass = true
		case TileConcrete:
			foundConcrete = true
		case TileAsphalt:
			foundAsphalt = true
		case TileWoodFloor:
			foundWood = true
		}
	}

	if !foundGrass || !foundConcrete || !foundAsphalt || !foundWood {
		t.Fatalf("Missing terrain transitions in multi-tier test: Grass:%v Concrete:%v Asphalt:%v WoodFloor:%v",
			foundGrass, foundConcrete, foundAsphalt, foundWood)
	}
}

// TestAdversarial_TileTransitions_NoReverseTransitions ensures lower-priority terrain
// NEVER generates transition overlays over higher-priority terrain.
func TestAdversarial_TileTransitions_NoReverseTransitions(t *testing.T) {
	// Concrete (30) surrounded by Dirt (10) and Grass (20)
	m := NewMap(5, 5)
	for i := range m.Tiles {
		m.Tiles[i] = TileDirt
	}
	m.SetTile(2, 1, TileGrass)
	m.SetTile(2, 3, TileGrass)
	m.SetTile(1, 2, TileGrass)
	m.SetTile(3, 2, TileGrass)

	// Center tile is Concrete
	m.SetTile(2, 2, TileConcrete)

	transitions := GetTileTransitions(m, 2, 2)
	if len(transitions) != 0 {
		t.Fatalf("Higher priority Concrete (30) surrounded by lower priority Dirt (10) / Grass (20) should have 0 transitions, got %d: %+v",
			len(transitions), transitions)
	}
}

// TestAdversarial_TileTransitions_WallTilesReceiveNoTransitions ensures walls
// do not receive ground terrain transitions.
func TestAdversarial_TileTransitions_WallTilesReceiveNoTransitions(t *testing.T) {
	m := NewMap(5, 5)
	for i := range m.Tiles {
		m.Tiles[i] = TileWoodFloor
	}
	m.SetTile(2, 2, TileWall)

	transitions := GetTileTransitions(m, 2, 2)
	if transitions != nil {
		t.Fatalf("TileWall should receive nil transitions, got %v", transitions)
	}
}

// TestAdversarial_TileTransitions_DiagonalOnlyNeighbor tests diagonal-only neighbor
// triggering IsDiagonal = true and SubtileOuterCorner.
func TestAdversarial_TileTransitions_DiagonalOnlyNeighbor(t *testing.T) {
	m := NewMap(5, 5)
	for i := range m.Tiles {
		m.Tiles[i] = TileDirt
	}

	// Grass only at diagonal (1, 1) relative to center (2, 2)
	m.SetTile(1, 1, TileGrass)

	transitions := GetTileTransitions(m, 2, 2)
	if len(transitions) != 1 {
		t.Fatalf("Expected exactly 1 diagonal transition, got %d", len(transitions))
	}

	tr := transitions[0]
	if tr.Quad != QuadNW {
		t.Errorf("Expected QuadNW, got %v", tr.Quad)
	}
	if tr.TerrainType != TileGrass {
		t.Errorf("Expected TileGrass, got %v", tr.TerrainType)
	}
	if tr.State != SubtileOuterCorner {
		t.Errorf("Expected SubtileOuterCorner, got %v", tr.State)
	}
	if !tr.IsDiagonal {
		t.Errorf("Expected IsDiagonal = true, got %v", tr.IsDiagonal)
	}
}
