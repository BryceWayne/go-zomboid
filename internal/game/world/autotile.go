package world

// Quadrant identifies one of the four 64x64 sub-quadrants of a 128x128 tile cell.
type Quadrant int

const (
	QuadNW Quadrant = iota // Top-Left quadrant (x: 0..64, y: 0..64)
	QuadNE                 // Top-Right quadrant (x: 64..128, y: 0..64)
	QuadSW                 // Bottom-Left quadrant (x: 0..64, y: 64..128)
	QuadSE                 // Bottom-Right quadrant (x: 64..128, y: 64..128)
)

// String returns a human-readable name for the quadrant.
func (q Quadrant) String() string {
	switch q {
	case QuadNW:
		return "NW"
	case QuadNE:
		return "NE"
	case QuadSW:
		return "SW"
	case QuadSE:
		return "SE"
	default:
		return "Unknown"
	}
}

// SubtileState represents the boundary/corner topology configuration of a quadrant.
type SubtileState int

const (
	SubtileFull           SubtileState = iota // 0: Completely solid interior quadrant
	SubtileHorizontalEdge                     // 1: Boundary transition along horizontal edge (North or South)
	SubtileVerticalEdge                       // 2: Boundary transition along vertical edge (East or West)
	SubtileOuterCorner                        // 3: Convex outer corner tip
	SubtileInnerCorner                        // 4: Concave inner corner notch
)

// String returns a human-readable name for the subtile state.
func (s SubtileState) String() string {
	switch s {
	case SubtileFull:
		return "Full"
	case SubtileHorizontalEdge:
		return "HorizontalEdge"
	case SubtileVerticalEdge:
		return "VerticalEdge"
	case SubtileOuterCorner:
		return "OuterCorner"
	case SubtileInnerCorner:
		return "InnerCorner"
	default:
		return "Unknown"
	}
}

// GetCardinalBitmask4 computes a 4-bit cardinal neighbor mask for a given cell (x, y)
// matching the specified matchType.
// Bit 0 (1): North (x, y - 1)
// Bit 1 (2): East  (x + 1, y)
// Bit 2 (4): South (x, y + 1)
// Bit 3 (8): West  (x - 1, y)
// Range: [0..15]
func GetCardinalBitmask4(m *Map, x, y int, matchType TileType) uint8 {
	if m == nil {
		return 0
	}
	var mask uint8
	if m.GetTile(x, y-1) == matchType {
		mask |= (1 << 0)
	}
	if m.GetTile(x+1, y) == matchType {
		mask |= (1 << 1)
	}
	if m.GetTile(x, y+1) == matchType {
		mask |= (1 << 2)
	}
	if m.GetTile(x-1, y) == matchType {
		mask |= (1 << 3)
	}
	return mask
}

// GetWallBitmask returns the 4-bit cardinal bitmask for walls at (x, y).
func GetWallBitmask(m *Map, x, y int) uint8 {
	return GetCardinalBitmask4(m, x, y, TileWall)
}

// GetFenceBitmask returns the 4-bit cardinal bitmask for fences at (x, y).
func GetFenceBitmask(m *Map, x, y int) uint8 {
	return GetCardinalBitmask4(m, x, y, TileFence)
}

// GroundType returns the effective underlying ground floor type for any tile.
// Obstacles and foliage placed outdoors on grass return TileGrass.
// Debris placed on concrete returns TileConcrete.
// Floor tiles return their own tile type.
func GroundType(t TileType) TileType {
	switch t {
	case TileGrass, TileTree, TileTent, TileElevationBlock, TileRamp, TileStump,
		TileMushroom, TileSign, TileBench, TileChest, TileSculpture, TileBush, TileFlower, TileStone:
		return TileGrass
	case TileDirt:
		return TileDirt
	case TileConcrete, TileDebris:
		return TileConcrete
	case TileAsphalt:
		return TileAsphalt
	case TileWoodFloor:
		return TileWoodFloor
	case TileTileFloor:
		return TileTileFloor
	case TileWall:
		return TileWall
	case TileFence:
		return TileFence
	default:
		return TileGrass
	}
}

// TerrainPriority defines the layer rendering hierarchy for autotiling and terrain blending.
// Higher priority terrains render smooth transition overlays on top of lower priority terrains.
func TerrainPriority(t TileType) int {
	gt := GroundType(t)
	switch gt {
	case TileDirt:
		return 10
	case TileGrass:
		return 20
	case TileConcrete:
		return 30
	case TileAsphalt:
		return 40
	case TileWoodFloor:
		return 50
	case TileTileFloor:
		return 50
	default:
		return 0
	}
}

// GetQuadrantSubtile computes the sub-tile corner/edge state for 4-quadrant terrain blending.
// It evaluates the horizontal, vertical, and diagonal neighbors of the given quadrant against primaryType.
func GetQuadrantSubtile(m *Map, x, y int, quad Quadrant, primaryType TileType) SubtileState {
	if m == nil {
		return SubtileFull
	}

	var hx, hy int // horizontal neighbor
	var vx, vy int // vertical neighbor
	var dx, dy int // diagonal neighbor

	switch quad {
	case QuadNW:
		hx, hy = x-1, y
		vx, vy = x, y-1
		dx, dy = x-1, y-1
	case QuadNE:
		hx, hy = x+1, y
		vx, vy = x, y-1
		dx, dy = x+1, y-1
	case QuadSW:
		hx, hy = x-1, y
		vx, vy = x, y+1
		dx, dy = x-1, y+1
	case QuadSE:
		hx, hy = x+1, y
		vx, vy = x, y+1
		dx, dy = x+1, y+1
	}

	hMatch := GroundType(m.GetTile(hx, hy)) == primaryType
	vMatch := GroundType(m.GetTile(vx, vy)) == primaryType
	dMatch := GroundType(m.GetTile(dx, dy)) == primaryType

	if hMatch && vMatch {
		if dMatch {
			return SubtileFull
		}
		return SubtileInnerCorner
	}
	if hMatch && !vMatch {
		return SubtileHorizontalEdge
	}
	if !hMatch && vMatch {
		return SubtileVerticalEdge
	}
	return SubtileOuterCorner
}

// TerrainTransition describes a transition overlay piece to render over a base tile.
type TerrainTransition struct {
	Quad        Quadrant
	TerrainType TileType
	State       SubtileState
	IsDiagonal  bool
}

// GetTileTransitions returns all quadrant transition overlays that should be rendered
// on top of the base tile at (x, y) to blend seamlessly with higher-priority neighboring terrain.
func GetTileTransitions(m *Map, x, y int) []TerrainTransition {
	if m == nil {
		return nil
	}

	baseTile := m.GetTile(x, y)
	baseGround := GroundType(baseTile)
	basePriority := TerrainPriority(baseGround)

	if baseTile == TileWall {
		return nil
	}

	var transitions []TerrainTransition
	quads := []Quadrant{QuadNW, QuadNE, QuadSW, QuadSE}

	for _, q := range quads {
		var hx, hy, vx, vy, dx, dy int
		switch q {
		case QuadNW:
			hx, hy = x-1, y
			vx, vy = x, y-1
			dx, dy = x-1, y-1
		case QuadNE:
			hx, hy = x+1, y
			vx, vy = x, y-1
			dx, dy = x+1, y-1
		case QuadSW:
			hx, hy = x-1, y
			vx, vy = x, y+1
			dx, dy = x-1, y+1
		case QuadSE:
			hx, hy = x+1, y
			vx, vy = x, y+1
			dx, dy = x+1, y+1
		}

		hTile := GroundType(m.GetTile(hx, hy))
		vTile := GroundType(m.GetTile(vx, vy))
		dTile := GroundType(m.GetTile(dx, dy))

		hPrio := TerrainPriority(hTile)
		vPrio := TerrainPriority(vTile)
		dPrio := TerrainPriority(dTile)

		// Collect candidate higher-priority terrain types affecting this quadrant
		candidates := make(map[TileType]bool)
		if hPrio > basePriority {
			candidates[hTile] = true
		}
		if vPrio > basePriority {
			candidates[vTile] = true
		}
		if dPrio > basePriority {
			candidates[dTile] = true
		}

		for cand := range candidates {
			hMatch := (hTile == cand)
			vMatch := (vTile == cand)
			dMatch := (dTile == cand)

			if hMatch && vMatch {
				// Both cardinal neighbors match candidate terrain -> Inner corner or Full blend
				if dMatch {
					transitions = append(transitions, TerrainTransition{
						Quad:        q,
						TerrainType: cand,
						State:       SubtileOuterCorner, // Full convex outer quadrant from adjacent terrain
						IsDiagonal:  false,
					})
				} else {
					transitions = append(transitions, TerrainTransition{
						Quad:        q,
						TerrainType: cand,
						State:       SubtileInnerCorner,
						IsDiagonal:  false,
					})
				}
			} else if hMatch && !vMatch {
				transitions = append(transitions, TerrainTransition{
					Quad:        q,
					TerrainType: cand,
					State:       SubtileHorizontalEdge,
					IsDiagonal:  false,
				})
			} else if !hMatch && vMatch {
				transitions = append(transitions, TerrainTransition{
					Quad:        q,
					TerrainType: cand,
					State:       SubtileVerticalEdge,
					IsDiagonal:  false,
				})
			} else if !hMatch && !vMatch && dMatch {
				transitions = append(transitions, TerrainTransition{
					Quad:        q,
					TerrainType: cand,
					State:       SubtileOuterCorner,
					IsDiagonal:  true,
				})
			}
		}
	}

	return transitions
}
