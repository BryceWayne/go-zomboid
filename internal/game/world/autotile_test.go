package world

import (
	"testing"
)

// TestCardinalBitmask4_All16Combinations verifies that GetCardinalBitmask4
// correctly identifies all 16 possible cardinal neighbor combinations.
func TestCardinalBitmask4_All16Combinations(t *testing.T) {
	// 5x5 map, test center cell at (2, 2)
	for mask := 0; mask < 16; mask++ {
		m := NewMap(5, 5)
		// Clear map to grass
		for i := range m.Tiles {
			m.Tiles[i] = TileGrass
		}

		hasN := (mask & (1 << 0)) != 0
		hasE := (mask & (1 << 1)) != 0
		hasS := (mask & (1 << 2)) != 0
		hasW := (mask & (1 << 3)) != 0

		if hasN {
			m.SetTile(2, 1, TileWall)
		}
		if hasE {
			m.SetTile(3, 2, TileWall)
		}
		if hasS {
			m.SetTile(2, 3, TileWall)
		}
		if hasW {
			m.SetTile(1, 2, TileWall)
		}

		computed := GetCardinalBitmask4(m, 2, 2, TileWall)
		if computed != uint8(mask) {
			t.Fatalf("Mask %d: expected %d, got %d (N:%v E:%v S:%v W:%v)",
				mask, mask, computed, hasN, hasE, hasS, hasW)
		}

		wallMask := GetWallBitmask(m, 2, 2)
		if wallMask != uint8(mask) {
			t.Fatalf("GetWallBitmask: expected %d, got %d", mask, wallMask)
		}
	}
}

// TestFenceBitmask_Connections verifies fence bitmask calculations.
func TestFenceBitmask_Connections(t *testing.T) {
	m := NewMap(10, 10)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Create a 4-cell horizontal fence at y=4, x=2..5
	for x := 2; x <= 5; x++ {
		m.SetTile(x, 4, TileFence)
	}

	// (2,4) has East neighbor only -> mask = 2
	if mask := GetFenceBitmask(m, 2, 4); mask != 2 {
		t.Fatalf("Fence start: expected 2, got %d", mask)
	}
	// (3,4) has West and East neighbors -> mask = 2 | 8 = 10
	if mask := GetFenceBitmask(m, 3, 4); mask != 10 {
		t.Fatalf("Fence mid: expected 10, got %d", mask)
	}
	// (5,4) has West neighbor only -> mask = 8
	if mask := GetFenceBitmask(m, 5, 4); mask != 8 {
		t.Fatalf("Fence end: expected 8, got %d", mask)
	}
}

// TestGroundType_All22TileTypes verifies that GroundType correctly classifies
// all 22 TileTypes into their underlying floor terrain substrates.
func TestGroundType_All22TileTypes(t *testing.T) {
	tests := []struct {
		tile     TileType
		expected TileType
	}{
		{TileGrass, TileGrass},
		{TileDirt, TileDirt},
		{TileWoodFloor, TileWoodFloor},
		{TileAsphalt, TileAsphalt},
		{TileConcrete, TileConcrete},
		{TileTileFloor, TileTileFloor},
		{TileWall, TileWall},
		{TileFence, TileFence},
		// Props on Grass
		{TileTree, TileGrass},
		{TileTent, TileGrass},
		{TileElevationBlock, TileGrass},
		{TileRamp, TileGrass},
		{TileStump, TileGrass},
		{TileMushroom, TileGrass},
		{TileSign, TileGrass},
		{TileBench, TileGrass},
		{TileChest, TileGrass},
		{TileSculpture, TileGrass},
		{TileBush, TileGrass},
		{TileFlower, TileGrass},
		{TileStone, TileGrass},
		// Prop on Concrete
		{TileDebris, TileConcrete},
	}

	for _, tt := range tests {
		got := GroundType(tt.tile)
		if got != tt.expected {
			t.Errorf("GroundType(%v %s): expected %v, got %v", tt.tile, tt.tile.String(), tt.expected, got)
		}
	}
}

// TestTerrainPriority_OrderingInvariants ensures strictly monotonic layering:
// Dirt (10) < Grass (20) < Concrete (30) < Asphalt (40) < Floors (50).
func TestTerrainPriority_OrderingInvariants(t *testing.T) {
	dirtPrio := TerrainPriority(TileDirt)
	grassPrio := TerrainPriority(TileGrass)
	concretePrio := TerrainPriority(TileConcrete)
	asphaltPrio := TerrainPriority(TileAsphalt)
	woodPrio := TerrainPriority(TileWoodFloor)
	tileFloorPrio := TerrainPriority(TileTileFloor)

	if !(dirtPrio < grassPrio) {
		t.Fatalf("Expected Dirt (%d) < Grass (%d)", dirtPrio, grassPrio)
	}
	if !(grassPrio < concretePrio) {
		t.Fatalf("Expected Grass (%d) < Concrete (%d)", grassPrio, concretePrio)
	}
	if !(concretePrio < asphaltPrio) {
		t.Fatalf("Expected Concrete (%d) < Asphalt (%d)", concretePrio, asphaltPrio)
	}
	if !(asphaltPrio < woodPrio) {
		t.Fatalf("Expected Asphalt (%d) < WoodFloor (%d)", asphaltPrio, woodPrio)
	}
	if woodPrio != tileFloorPrio {
		t.Fatalf("Expected WoodFloor (%d) == TileFloor (%d)", woodPrio, tileFloorPrio)
	}
}

// TestGetQuadrantSubtile_AllStatesAcrossAll4Quadrants verifies 4-quadrant blob neighbor evaluation.
func TestGetQuadrantSubtile_AllStatesAcrossAll4Quadrants(t *testing.T) {
	quads := []Quadrant{QuadNW, QuadNE, QuadSW, QuadSE}

	for _, q := range quads {
		// 1. SubtileFull (H, V, D all match)
		mFull := NewMap(5, 5)
		for i := range mFull.Tiles {
			mFull.Tiles[i] = TileGrass
		}
		if s := GetQuadrantSubtile(mFull, 2, 2, q, TileGrass); s != SubtileFull {
			t.Fatalf("Quad %v: expected SubtileFull, got %v", q, s)
		}

		// 2. SubtileInnerCorner (H and V match, D does NOT match)
		mInner := NewMap(5, 5)
		for i := range mInner.Tiles {
			mInner.Tiles[i] = TileGrass
		}
		switch q {
		case QuadNW:
			mInner.SetTile(1, 1, TileDirt) // diagonal
		case QuadNE:
			mInner.SetTile(3, 1, TileDirt)
		case QuadSW:
			mInner.SetTile(1, 3, TileDirt)
		case QuadSE:
			mInner.SetTile(3, 3, TileDirt)
		}
		if s := GetQuadrantSubtile(mInner, 2, 2, q, TileGrass); s != SubtileInnerCorner {
			t.Fatalf("Quad %v: expected SubtileInnerCorner, got %v", q, s)
		}

		// 3. SubtileHorizontalEdge (H matches, V does not)
		mH := NewMap(5, 5)
		for i := range mH.Tiles {
			mH.Tiles[i] = TileDirt
		}
		mH.SetTile(2, 2, TileGrass)
		switch q {
		case QuadNW:
			mH.SetTile(1, 2, TileGrass) // H match
		case QuadNE:
			mH.SetTile(3, 2, TileGrass) // H match
		case QuadSW:
			mH.SetTile(1, 2, TileGrass) // H match
		case QuadSE:
			mH.SetTile(3, 2, TileGrass) // H match
		}
		if s := GetQuadrantSubtile(mH, 2, 2, q, TileGrass); s != SubtileHorizontalEdge {
			t.Fatalf("Quad %v: expected SubtileHorizontalEdge, got %v", q, s)
		}

		// 4. SubtileVerticalEdge (V matches, H does not)
		mV := NewMap(5, 5)
		for i := range mV.Tiles {
			mV.Tiles[i] = TileDirt
		}
		mV.SetTile(2, 2, TileGrass)
		switch q {
		case QuadNW:
			mV.SetTile(2, 1, TileGrass) // V match
		case QuadNE:
			mV.SetTile(2, 1, TileGrass) // V match
		case QuadSW:
			mV.SetTile(2, 3, TileGrass) // V match
		case QuadSE:
			mV.SetTile(2, 3, TileGrass) // V match
		}
		if s := GetQuadrantSubtile(mV, 2, 2, q, TileGrass); s != SubtileVerticalEdge {
			t.Fatalf("Quad %v: expected SubtileVerticalEdge, got %v", q, s)
		}

		// 5. SubtileOuterCorner (Neither H nor V matches)
		mOuter := NewMap(5, 5)
		for i := range mOuter.Tiles {
			mOuter.Tiles[i] = TileDirt
		}
		mOuter.SetTile(2, 2, TileGrass)
		if s := GetQuadrantSubtile(mOuter, 2, 2, q, TileGrass); s != SubtileOuterCorner {
			t.Fatalf("Quad %v: expected SubtileOuterCorner, got %v", q, s)
		}
	}
}

// TestGetTileTransitions_DirtPathThroughGrass verifies that dirt tiles adjacent to grass
// produce grass transition overlays to blend smoothly.
func TestGetTileTransitions_DirtPathThroughGrass(t *testing.T) {
	m := NewMap(10, 10)
	for i := range m.Tiles {
		m.Tiles[i] = TileGrass
	}

	// Place a vertical dirt trail down x=4
	for y := 0; y < 10; y++ {
		m.SetTile(4, y, TileDirt)
	}

	// Dirt tile at (4, 4) should have Grass transitions from West (x=3) and East (x=5)
	transitions := GetTileTransitions(m, 4, 4)
	if len(transitions) == 0 {
		t.Fatalf("Expected transitions on dirt tile bordering grass, got 0")
	}

	hasNW := false
	hasNE := false
	hasSW := false
	hasSE := false

	for _, tr := range transitions {
		if tr.TerrainType != TileGrass {
			t.Errorf("Expected transition terrain TileGrass, got %v", tr.TerrainType)
		}
		switch tr.Quad {
		case QuadNW:
			hasNW = true
			if tr.State != SubtileHorizontalEdge && tr.State != SubtileVerticalEdge {
				t.Errorf("QuadNW unexpected state %v", tr.State)
			}
		case QuadNE:
			hasNE = true
		case QuadSW:
			hasSW = true
		case QuadSE:
			hasSE = true
		}
	}

	if !hasNW || !hasNE || !hasSW || !hasSE {
		t.Errorf("Missing quadrant transitions: NW:%v NE:%v SW:%v SE:%v", hasNW, hasNE, hasSW, hasSE)
	}
}

// TestAutotiling_OutOfBoundsAndExtremeCoordinates ensures no crashes on boundary queries.
func TestAutotiling_OutOfBoundsAndExtremeCoordinates(t *testing.T) {
	m := NewMap(10, 10)

	// Coordinates out of bounds
	oobCoords := [][2]int{
		{-1, -1},
		{-10, 5},
		{5, -10},
		{10, 10},
		{100, 100},
		{0, 0},
		{9, 9},
	}

	for _, coord := range oobCoords {
		x, y := coord[0], coord[1]
		_ = GetCardinalBitmask4(m, x, y, TileWall)
		_ = GetWallBitmask(m, x, y)
		_ = GetFenceBitmask(m, x, y)
		_ = GetQuadrantSubtile(m, x, y, QuadNW, TileGrass)
		_ = GetQuadrantSubtile(m, x, y, QuadNE, TileGrass)
		_ = GetQuadrantSubtile(m, x, y, QuadSW, TileGrass)
		_ = GetQuadrantSubtile(m, x, y, QuadSE, TileGrass)
		_ = GetTileTransitions(m, x, y)
	}
}
