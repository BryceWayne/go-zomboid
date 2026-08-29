package world

import (
	"sync"
	"testing"
)

// TestChallenger_All256NeighborPermutations_GetQuadrantSubtile exhaustively verifies
// that GetQuadrantSubtile handles all 2^8 = 256 possible 8-neighbor terrain configurations
// for every quadrant (NW, NE, SW, SE) strictly respecting the topological invariants.
func TestChallenger_All256NeighborPermutations_GetQuadrantSubtile(t *testing.T) {
	quads := []Quadrant{QuadNW, QuadNE, QuadSW, QuadSE}

	// 3x3 grid around center (1, 1).
	// Neighbors mapping in 3x3:
	// (0,0)=NW(bit0), (1,0)=N(bit1),  (2,0)=NE(bit2)
	// (0,1)=W(bit3),                  (2,1)=E(bit4)
	// (0,2)=SW(bit5), (1,2)=S(bit6),  (2,2)=SE(bit7)
	for mask := 0; mask < 256; mask++ {
		m := NewMap(3, 3)
		for i := range m.Tiles {
			m.Tiles[i] = TileDirt
		}
		m.SetTile(1, 1, TileGrass)

		// Set 8 neighbors according to mask bits
		if (mask & (1 << 0)) != 0 {
			m.SetTile(0, 0, TileGrass)
		}
		if (mask & (1 << 1)) != 0 {
			m.SetTile(1, 0, TileGrass)
		}
		if (mask & (1 << 2)) != 0 {
			m.SetTile(2, 0, TileGrass)
		}
		if (mask & (1 << 3)) != 0 {
			m.SetTile(0, 1, TileGrass)
		}
		if (mask & (1 << 4)) != 0 {
			m.SetTile(2, 1, TileGrass)
		}
		if (mask & (1 << 5)) != 0 {
			m.SetTile(0, 2, TileGrass)
		}
		if (mask & (1 << 6)) != 0 {
			m.SetTile(1, 2, TileGrass)
		}
		if (mask & (1 << 7)) != 0 {
			m.SetTile(2, 2, TileGrass)
		}

		for _, q := range quads {
			state := GetQuadrantSubtile(m, 1, 1, q, TileGrass)

			// Determine expected state based on quadrant-specific H, V, D neighbors
			var hMatch, vMatch, dMatch bool
			switch q {
			case QuadNW:
				hMatch = (mask & (1 << 3)) != 0 // W (0,1)
				vMatch = (mask & (1 << 1)) != 0 // N (1,0)
				dMatch = (mask & (1 << 0)) != 0 // NW (0,0)
			case QuadNE:
				hMatch = (mask & (1 << 4)) != 0 // E (2,1)
				vMatch = (mask & (1 << 1)) != 0 // N (1,0)
				dMatch = (mask & (1 << 2)) != 0 // NE (2,0)
			case QuadSW:
				hMatch = (mask & (1 << 3)) != 0 // W (0,1)
				vMatch = (mask & (1 << 6)) != 0 // S (1,2)
				dMatch = (mask & (1 << 5)) != 0 // SW (0,2)
			case QuadSE:
				hMatch = (mask & (1 << 4)) != 0 // E (2,1)
				vMatch = (mask & (1 << 6)) != 0 // S (1,2)
				dMatch = (mask & (1 << 7)) != 0 // SE (2,2)
			}

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

			if state != expected {
				t.Fatalf("Mask %08b Quad %v: expected state %v (%s), got %v (%s)",
					mask, q, expected, expected.String(), state, state.String())
			}
		}
	}
}

// TestChallenger_MapBorderAndCornerBitmasks verifies autotiling queries against all border
// edges, corners, degenerate maps (1x1, 1xN, Nx1), and out-of-bounds coordinates.
func TestChallenger_MapBorderAndCornerBitmasks(t *testing.T) {
	maps := []*Map{
		NewMap(1, 1),
		NewMap(1, 20),
		NewMap(20, 1),
		NewMap(2, 2),
		NewMap(50, 50),
	}

	testCoords := [][2]int{
		{-1000, -1000},
		{-1, -1},
		{-1, 0},
		{0, -1},
		{0, 0},
		{0, 1},
		{1, 0},
		{19, 0},
		{0, 19},
		{19, 19},
		{49, 49},
		{50, 50},
		{100, 100},
		{10000, 10000},
	}

	for mapIdx, m := range maps {
		// Fill with alternating walls, fences, grass, and dirt
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				switch (x + y) % 4 {
				case 0:
					m.SetTile(x, y, TileWall)
				case 1:
					m.SetTile(x, y, TileFence)
				case 2:
					m.SetTile(x, y, TileGrass)
				case 3:
					m.SetTile(x, y, TileDirt)
				}
			}
		}

		for _, pt := range testCoords {
			x, y := pt[0], pt[1]

			// Ensure zero panics and results are bounded in [0..15]
			wMask := GetWallBitmask(m, x, y)
			if wMask > 15 {
				t.Fatalf("Map %d at (%d, %d): wall mask %d out of bounds [0..15]", mapIdx, x, y, wMask)
			}

			fMask := GetFenceBitmask(m, x, y)
			if fMask > 15 {
				t.Fatalf("Map %d at (%d, %d): fence mask %d out of bounds [0..15]", mapIdx, x, y, fMask)
			}

			cardMask := GetCardinalBitmask4(m, x, y, TileGrass)
			if cardMask > 15 {
				t.Fatalf("Map %d at (%d, %d): cardinal mask %d out of bounds [0..15]", mapIdx, x, y, cardMask)
			}

			for _, q := range []Quadrant{QuadNW, QuadNE, QuadSW, QuadSE} {
				st := GetQuadrantSubtile(m, x, y, q, TileGrass)
				if st < SubtileFull || st > SubtileInnerCorner {
					t.Fatalf("Map %d at (%d, %d): invalid subtile state %v", mapIdx, x, y, st)
				}
			}

			trans := GetTileTransitions(m, x, y)
			// Nil or bounded transitions slice
			if len(trans) > 16 {
				t.Fatalf("Map %d at (%d, %d): unexpectedly huge transition count %d", mapIdx, x, y, len(trans))
			}
		}
	}

	// Test nil map safety
	if mask := GetCardinalBitmask4(nil, 0, 0, TileWall); mask != 0 {
		t.Fatalf("Expected 0 on nil map, got %d", mask)
	}
	if mask := GetWallBitmask(nil, 0, 0); mask != 0 {
		t.Fatalf("Expected 0 on nil map, got %d", mask)
	}
	if mask := GetFenceBitmask(nil, 0, 0); mask != 0 {
		t.Fatalf("Expected 0 on nil map, got %d", mask)
	}
	if st := GetQuadrantSubtile(nil, 0, 0, QuadNW, TileGrass); st != SubtileFull {
		t.Fatalf("Expected SubtileFull on nil map, got %v", st)
	}
	if trans := GetTileTransitions(nil, 0, 0); trans != nil {
		t.Fatalf("Expected nil on nil map, got %v", trans)
	}
}

// TestChallenger_NonStandardTerrainLayouts_MultiPriorityClash tests multi-terrain conflicts
// where a low-priority tile is bordered by multiple distinct higher-priority terrains.
func TestChallenger_NonStandardTerrainLayouts_MultiPriorityClash(t *testing.T) {
	m := NewMap(5, 5)
	// Base tile at (2, 2) is Dirt (priority 10)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			m.SetTile(x, y, TileDirt)
		}
	}

	// Place 4 distinct higher-priority terrains on the 4 cardinal neighbors:
	// North: Grass (prio 20)
	// East:  Concrete (prio 30)
	// South: Asphalt (prio 40)
	// West:  WoodFloor (prio 50)
	m.SetTile(2, 1, TileGrass)
	m.SetTile(3, 2, TileConcrete)
	m.SetTile(2, 3, TileAsphalt)
	m.SetTile(1, 2, TileWoodFloor)

	transitions := GetTileTransitions(m, 2, 2)
	if len(transitions) == 0 {
		t.Fatalf("Expected multiple transitions on multi-clash tile, got 0")
	}

	// Verify that each neighboring higher-priority terrain generates appropriate edge transitions
	foundGrass := false
	foundConcrete := false
	foundAsphalt := false
	foundWood := false

	for _, tr := range transitions {
		switch tr.TerrainType {
		case TileGrass:
			foundGrass = true
			if tr.Quad != QuadNW && tr.Quad != QuadNE {
				t.Errorf("Grass transition generated for unexpected quadrant %v", tr.Quad)
			}
			if tr.State != SubtileVerticalEdge && tr.State != SubtileHorizontalEdge {
				t.Errorf("Grass transition unexpected state %v", tr.State)
			}
		case TileConcrete:
			foundConcrete = true
			if tr.Quad != QuadNE && tr.Quad != QuadSE {
				t.Errorf("Concrete transition generated for unexpected quadrant %v", tr.Quad)
			}
		case TileAsphalt:
			foundAsphalt = true
			if tr.Quad != QuadSW && tr.Quad != QuadSE {
				t.Errorf("Asphalt transition generated for unexpected quadrant %v", tr.Quad)
			}
		case TileWoodFloor:
			foundWood = true
			if tr.Quad != QuadNW && tr.Quad != QuadSW {
				t.Errorf("WoodFloor transition generated for unexpected quadrant %v", tr.Quad)
			}
		}
	}

	if !foundGrass || !foundConcrete || !foundAsphalt || !foundWood {
		t.Fatalf("Expected all 4 higher-priority terrains in transitions: Grass:%v Concrete:%v Asphalt:%v Wood:%v",
			foundGrass, foundConcrete, foundAsphalt, foundWood)
	}
}

// TestChallenger_CheckerboardTerrainPattern ensures checkerboard patterns of alternating
// terrain types generate consistent, non-overlapping diagonal/corner overlays across all cells.
func TestChallenger_CheckerboardTerrainPattern(t *testing.T) {
	size := 20
	m := NewMap(size, size)

	// Alternating Dirt and Grass checkerboard
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if (x+y)%2 == 0 {
				m.SetTile(x, y, TileDirt)
			} else {
				m.SetTile(x, y, TileGrass)
			}
		}
	}

	// Every Dirt cell has 4 Grass cardinal neighbors -> should produce transitions on all 4 quadrants
	for y := 1; y < size-1; y++ {
		for x := 1; x < size-1; x++ {
			if m.GetTile(x, y) == TileDirt {
				trans := GetTileTransitions(m, x, y)
				if len(trans) < 4 {
					t.Fatalf("Dirt tile at (%d, %d) in checkerboard: expected >= 4 transitions, got %d", x, y, len(trans))
				}
			}
		}
	}
}

// TestChallenger_RapidMapQueries_ThroughputAndConcurrency tests parallel high-throughput
// calls across multiple goroutines to verify race freedom and high performance.
func TestChallenger_RapidMapQueries_ThroughputAndConcurrency(t *testing.T) {
	m := NewMap(100, 100)
	// Procedurally generate map with buildings and paths
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			switch (x*3 + y*7) % 8 {
			case 0:
				m.SetTile(x, y, TileWall)
			case 1:
				m.SetTile(x, y, TileFence)
			case 2:
				m.SetTile(x, y, TileDirt)
			case 3:
				m.SetTile(x, y, TileConcrete)
			case 4:
				m.SetTile(x, y, TileAsphalt)
			case 5:
				m.SetTile(x, y, TileWoodFloor)
			case 6:
				m.SetTile(x, y, TileTileFloor)
			default:
				m.SetTile(x, y, TileGrass)
			}
		}
	}

	const goroutines = 8
	const queriesPerGoroutine = 50000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < queriesPerGoroutine; i++ {
				x := (i*13 + seed*17) % 100
				y := (i*29 + seed*31) % 100

				_ = GetWallBitmask(m, x, y)
				_ = GetFenceBitmask(m, x, y)
				_ = GetCardinalBitmask4(m, x, y, TileGrass)
				_ = GetQuadrantSubtile(m, x, y, QuadNW, TileGrass)
				_ = GetQuadrantSubtile(m, x, y, QuadNE, TileConcrete)
				_ = GetQuadrantSubtile(m, x, y, QuadSW, TileAsphalt)
				_ = GetQuadrantSubtile(m, x, y, QuadSE, TileWoodFloor)
				_ = GetTileTransitions(m, x, y)
			}
		}(g)
	}

	wg.Wait()
}

// TestChallenger_TopologicalInvariants_CorridorsAndJunctions tests bitmasks along straight
// runs, corners, T-junctions, cross-junctions, and dead-ends for both walls and fences.
func TestChallenger_TopologicalInvariants_CorridorsAndJunctions(t *testing.T) {
	testTypes := []TileType{TileWall, TileFence}

	for _, tt := range testTypes {
		m := NewMap(10, 10)
		for i := range m.Tiles {
			m.Tiles[i] = TileGrass
		}

		// Create a Cross-Junction at (5, 5) with arms reaching (5,4), (6,5), (5,6), (4,5)
		m.SetTile(5, 5, tt)
		m.SetTile(5, 4, tt) // N
		m.SetTile(6, 5, tt) // E
		m.SetTile(5, 6, tt) // S
		m.SetTile(4, 5, tt) // W

		maskCenter := GetCardinalBitmask4(m, 5, 5, tt)
		if maskCenter != 15 { // N(1) | E(2) | S(4) | W(8) = 15
			t.Fatalf("Cross junction center for %v: expected 15, got %d", tt, maskCenter)
		}

		maskN := GetCardinalBitmask4(m, 5, 4, tt)
		if maskN != 4 { // Only S neighbor connected
			t.Fatalf("North arm for %v: expected 4 (S), got %d", tt, maskN)
		}

		maskE := GetCardinalBitmask4(m, 6, 5, tt)
		if maskE != 8 { // Only W neighbor connected
			t.Fatalf("East arm for %v: expected 8 (W), got %d", tt, maskE)
		}

		maskS := GetCardinalBitmask4(m, 5, 6, tt)
		if maskS != 1 { // Only N neighbor connected
			t.Fatalf("South arm for %v: expected 1 (N), got %d", tt, maskS)
		}

		maskW := GetCardinalBitmask4(m, 4, 5, tt)
		if maskW != 2 { // Only E neighbor connected
			t.Fatalf("West arm for %v: expected 2 (E), got %d", tt, maskW)
		}

		// Remove South arm to form T-junction (N, E, W)
		m.SetTile(5, 6, TileGrass)
		maskTJunction := GetCardinalBitmask4(m, 5, 5, tt)
		if maskTJunction != 11 { // N(1) | E(2) | W(8) = 11
			t.Fatalf("T-junction (N+E+W) for %v: expected 11, got %d", tt, maskTJunction)
		}

		// Remove West arm to form L-Corner (N, E)
		m.SetTile(4, 5, TileGrass)
		maskLCorner := GetCardinalBitmask4(m, 5, 5, tt)
		if maskLCorner != 3 { // N(1) | E(2) = 3
			t.Fatalf("L-Corner (N+E) for %v: expected 3, got %d", tt, maskLCorner)
		}

		// Remove North arm to form Single-End (E)
		m.SetTile(5, 4, TileGrass)
		maskEnd := GetCardinalBitmask4(m, 5, 5, tt)
		if maskEnd != 2 { // E(2) = 2
			t.Fatalf("Single-end (E) for %v: expected 2, got %d", tt, maskEnd)
		}

		// Remove East arm to form Isolated (0)
		m.SetTile(6, 5, TileGrass)
		maskIsolated := GetCardinalBitmask4(m, 5, 5, tt)
		if maskIsolated != 0 {
			t.Fatalf("Isolated tile for %v: expected 0, got %d", tt, maskIsolated)
		}
	}
}
