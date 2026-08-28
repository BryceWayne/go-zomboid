package world

import (
	"math"
	"math/rand"
)

// TileType represents the type of a world grid cell.
type TileType int

const (
	TileGrass TileType = iota
	TileWall
	TileDirt
	TileWoodFloor
	TileTree
	TileAsphalt
	TileConcrete
	TileTileFloor
	TileFence
	TileDebris
)

const TileSize = 32

// IsSolid returns true if the tile physically blocks entity movement and collisions.
func (t TileType) IsSolid() bool {
	return t == TileWall || t == TileTree || t == TileFence || t == TileDebris
}

// BlocksVision returns true if the tile blocks line of sight in raycast FOV calculation.
func (t TileType) BlocksVision() bool {
	return t == TileWall
}

// IsFloor returns true if the tile is a passable ground surface.
func (t TileType) IsFloor() bool {
	return t == TileGrass || t == TileDirt || t == TileWoodFloor ||
		t == TileAsphalt || t == TileConcrete || t == TileTileFloor
}

// DistrictType represents the functional zoning of an urban area.
type DistrictType int

const (
	DistrictResidential DistrictType = iota
	DistrictCommercial
	DistrictIndustrial
	DistrictPark
)

func (d DistrictType) String() string {
	switch d {
	case DistrictResidential:
		return "Residential"
	case DistrictCommercial:
		return "Commercial"
	case DistrictIndustrial:
		return "Industrial"
	case DistrictPark:
		return "Park"
	default:
		return "Unknown"
	}
}

// BuildingType classifies the architectural template and function of a structure.
type BuildingType int

const (
	BuildingResidentialHouse BuildingType = iota
	BuildingGroceryStore
	BuildingPharmacy
	BuildingPoliceStation
	BuildingWarehouse
	BuildingStorageUnit
	BuildingParkPavilion
)

// Rect represents a 2D integer bounding box in tile grid coordinates.
type Rect struct {
	X, Y, W, H int
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

func (r Rect) Center() (int, int) {
	return r.X + r.W/2, r.Y + r.H/2
}

// Point represents a 2D float64 coordinate in world pixel units.
type Point struct {
	X, Y float64
}

// LootSpawn describes a designated contextual loot placement.
type LootSpawn struct {
	ItemType string  // "food", "water", "weapon", "axe", "shotgun", "ammo", "armor"
	X, Y     float64 // World coordinate in pixels
	RoomType string  // "kitchen", "bedroom", "armory", "grocery", "pharmacy", "warehouse", "outdoor"
}

// ZombieSpawn describes a planned zombie entity placement.
type ZombieSpawn struct {
	X, Y     float64 // World coordinate in pixels
	IsRunner bool
	District DistrictType
}

// Room represents a partitioned interior chamber of a building.
type Room struct {
	Name   string
	Bounds Rect
	Floor  TileType
}

// Building represents a synthesized multi-room building on a parcel.
type Building struct {
	Type     BuildingType
	District DistrictType
	Bounds   Rect
	Rooms    []Room
	Doors    []Point
}

// Lot represents a subdivided urban parcel with street frontage.
type Lot struct {
	District DistrictType
	Bounds   Rect
	Building *Building
}

// District represents a zoned cluster of city blocks.
type District struct {
	Type   DistrictType
	Bounds Rect
	Lots   []Lot
}

// Map encapsulates the tile grid, visibility/fog states, and procedural town layout metadata.
type Map struct {
	Width, Height int
	Tiles         []TileType
	Visible       []bool
	Explored      []bool

	// Procedural Layout & Spawn Metadata
	Districts    []District
	Buildings    []Building
	PlayerSpawn  Point
	LootSpawns   []LootSpawn
	ZombieSpawns []ZombieSpawn
}

// NewMap constructs and procedurally generates a fully zoned town map.
func NewMap(width, height int) *Map {
	m := &Map{
		Width:        width,
		Height:       height,
		Tiles:        make([]TileType, width*height),
		Visible:      make([]bool, width*height),
		Explored:     make([]bool, width*height),
		Districts:    make([]District, 0),
		Buildings:    make([]Building, 0),
		LootSpawns:   make([]LootSpawn, 0),
		ZombieSpawns: make([]ZombieSpawn, 0),
	}

	// 1. Fill base terrain and perimeter boundary
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				m.SetTile(x, y, TileWall)
			} else if (x == 1 || x == width-2 || y == 1 || y == height-2) && width > 10 && height > 10 {
				m.SetTile(x, y, TileTree)
			} else {
				m.SetTile(x, y, TileGrass)
			}
		}
	}

	// Fallback for micro test maps (e.g. 10x10 or 20x20 in minimal unit tests)
	if width < 40 || height < 40 {
		m.generateMicroMap()
		return m
	}

	// 2. Procedural Town Layout Generation
	m.generateTownLayout()

	return m
}

// generateMicroMap handles small test maps safely
func (m *Map) generateMicroMap() {
	midX, midY := m.Width/2, m.Height/2
	for y := 1; y < m.Height-1; y++ {
		m.SetTile(midX, y, TileDirt)
	}
	for x := 1; x < m.Width-1; x++ {
		m.SetTile(x, midY, TileDirt)
	}
	m.PlayerSpawn = Point{
		X: float64(midX) * float64(TileSize),
		Y: float64(midY) * float64(TileSize),
	}
}

// generateTownLayout runs the multi-phase procedural town synthesis
func (m *Map) generateTownLayout() {
	midX := m.Width / 2
	midY := m.Height / 2

	// Phase 1: Construct Road Network
	m.generateRoadNetwork(midX, midY)

	// Phase 2: Create District Zones & Subdivided Blocks
	m.generateDistricts(midX, midY)

	// Phase 3: Synthesize Buildings & Multi-Room Interiors
	m.generateAllBuildings()

	// Phase 4: Environmental Props, Fences & Outdoor Vegetation
	m.generateEnvironmentalProps()

	// Phase 5: Extract Thematic Contextual Spawns (Player, Loot, Zombies)
	m.extractSpawns()
}

// Phase 1: Road Network
func (m *Map) generateRoadNetwork(midX, midY int) {
	// Avenues: 4-tile wide TileAsphalt + flanking 1-tile TileConcrete sidewalks (total corridor = 6 tiles)
	// Vertical Main Avenue: columns [midX-2 .. midX+1], sidewalks at midX-3 and midX+2
	for y := 2; y < m.Height-2; y++ {
		m.SetTile(midX-3, y, TileConcrete)
		m.SetTile(midX-2, y, TileAsphalt)
		m.SetTile(midX-1, y, TileAsphalt)
		m.SetTile(midX, y, TileAsphalt)
		m.SetTile(midX+1, y, TileAsphalt)
		m.SetTile(midX+2, y, TileConcrete)
	}

	// Horizontal Main Avenue: rows [midY-2 .. midY+1], sidewalks at midY-3 and midY+2
	for x := 2; x < m.Width-2; x++ {
		m.SetTile(x, midY-3, TileConcrete)
		m.SetTile(x, midY-2, TileAsphalt)
		m.SetTile(x, midY-1, TileAsphalt)
		m.SetTile(x, midY, TileAsphalt)
		m.SetTile(x, midY+1, TileAsphalt)
		m.SetTile(x, midY+2, TileConcrete)
	}

	// Clean intersection center (pure asphalt)
	for y := midY - 2; y <= midY+1; y++ {
		for x := midX - 2; x <= midX+1; x++ {
			m.SetTile(x, y, TileAsphalt)
		}
	}

	// Secondary Cross-Streets (2-tile TileAsphalt + 1-tile TileConcrete sidewalks on each side = 4 tiles total)
	secX1 := 2 + (midX-3-2)/2 // ~24 in 100x100
	secX2 := (midX + 3) + (m.Width-2-(midX+3))/2
	secY1 := 2 + (midY-3-2)/2 // ~24
	secY2 := (midY + 3) + (m.Height-2-(midY+3))/2

	// Vertical Secondary Streets
	for _, sx := range []int{secX1, secX2} {
		for y := 2; y < m.Height-2; y++ {
			// Skip crossing main avenue asphalt
			if y >= midY-2 && y <= midY+1 {
				continue
			}
			m.SetTile(sx-1, y, TileConcrete)
			m.SetTile(sx, y, TileAsphalt)
			m.SetTile(sx+1, y, TileAsphalt)
			m.SetTile(sx+2, y, TileConcrete)
		}
	}

	// Horizontal Secondary Streets
	for _, sy := range []int{secY1, secY2} {
		for x := 2; x < m.Width-2; x++ {
			// Skip crossing main avenue asphalt
			if x >= midX-2 && x <= midX+1 {
				continue
			}
			m.SetTile(x, sy-1, TileConcrete)
			m.SetTile(x, sy, TileAsphalt)
			m.SetTile(x, sy+1, TileAsphalt)
			m.SetTile(x, sy+2, TileConcrete)
		}
	}

	// Clean secondary intersections
	for _, sx := range []int{secX1, secX2} {
		for _, sy := range []int{secY1, secY2} {
			for y := sy; y <= sy+1; y++ {
				for x := sx; x <= sx+1; x++ {
					m.SetTile(x, y, TileAsphalt)
				}
			}
		}
	}
}

// Phase 2: District Zoning & Block Subdivision
func (m *Map) generateDistricts(midX, midY int) {
	secX1 := 2 + (midX-3-2)/2
	secX2 := (midX + 3) + (m.Width-2-(midX+3))/2
	secY1 := 2 + (midY-3-2)/2
	secY2 := (midY + 3) + (m.Height-2-(midY+3))/2

	// Define 4 Districts mapped to Quadrants
	// Quadrant 0: Top-Left (Residential Suburbs)
	m.Districts = append(m.Districts, District{
		Type:   DistrictResidential,
		Bounds: Rect{X: 2, Y: 2, W: midX - 5, H: midY - 5},
		Lots: []Lot{
			{District: DistrictResidential, Bounds: Rect{X: 3, Y: 3, W: secX1 - 4, H: secY1 - 4}},
			{District: DistrictResidential, Bounds: Rect{X: secX1 + 3, Y: 3, W: midX - 6 - secX1, H: secY1 - 4}},
			{District: DistrictResidential, Bounds: Rect{X: 3, Y: secY1 + 3, W: secX1 - 4, H: midY - 6 - secY1}},
			{District: DistrictResidential, Bounds: Rect{X: secX1 + 3, Y: secY1 + 3, W: midX - 6 - secX1, H: midY - 6 - secY1}},
		},
	})

	// Quadrant 1: Top-Right (Commercial Downtown)
	m.Districts = append(m.Districts, District{
		Type:   DistrictCommercial,
		Bounds: Rect{X: midX + 3, Y: 2, W: m.Width - midX - 5, H: midY - 5},
		Lots: []Lot{
			{District: DistrictCommercial, Bounds: Rect{X: midX + 3, Y: 3, W: secX2 - 4 - midX, H: secY1 - 4}},
			{District: DistrictCommercial, Bounds: Rect{X: secX2 + 3, Y: 3, W: m.Width - 5 - secX2, H: secY1 - 4}},
			{District: DistrictCommercial, Bounds: Rect{X: midX + 3, Y: secY1 + 3, W: secX2 - 4 - midX, H: midY - 6 - secY1}},
			{District: DistrictCommercial, Bounds: Rect{X: secX2 + 3, Y: secY1 + 3, W: m.Width - 5 - secX2, H: midY - 6 - secY1}},
		},
	})

	// Quadrant 2: Bottom-Left (Parks & Nature Reserve)
	m.Districts = append(m.Districts, District{
		Type:   DistrictPark,
		Bounds: Rect{X: 2, Y: midY + 3, W: midX - 5, H: m.Height - midY - 5},
		Lots: []Lot{
			{District: DistrictPark, Bounds: Rect{X: 3, Y: midY + 3, W: secX1 - 4, H: secY2 - 4 - midY}},
			{District: DistrictPark, Bounds: Rect{X: secX1 + 3, Y: midY + 3, W: midX - 6 - secX1, H: secY2 - 4 - midY}},
			{District: DistrictPark, Bounds: Rect{X: 3, Y: secY2 + 3, W: secX1 - 4, H: m.Height - 5 - secY2}},
			{District: DistrictPark, Bounds: Rect{X: secX1 + 3, Y: secY2 + 3, W: midX - 6 - secX1, H: m.Height - 5 - secY2}},
		},
	})

	// Quadrant 3: Bottom-Right (Industrial & Logistics Yard)
	m.Districts = append(m.Districts, District{
		Type:   DistrictIndustrial,
		Bounds: Rect{X: midX + 3, Y: midY + 3, W: m.Width - midX - 5, H: m.Height - midY - 5},
		Lots: []Lot{
			{District: DistrictIndustrial, Bounds: Rect{X: midX + 3, Y: midY + 3, W: secX2 - 4 - midX, H: secY2 - 4 - midY}},
			{District: DistrictIndustrial, Bounds: Rect{X: secX2 + 3, Y: midY + 3, W: m.Width - 5 - secX2, H: secY2 - 4 - midY}},
			{District: DistrictIndustrial, Bounds: Rect{X: midX + 3, Y: secY2 + 3, W: secX2 - 4 - midX, H: m.Height - 5 - secY2}},
			{District: DistrictIndustrial, Bounds: Rect{X: secX2 + 3, Y: secY2 + 3, W: m.Width - 5 - secX2, H: m.Height - 5 - secY2}},
		},
	})
}

// Phase 3: Building Synthesis
func (m *Map) generateAllBuildings() {
	for _, dist := range m.Districts {
		for i, lot := range dist.Lots {
			switch dist.Type {
			case DistrictResidential:
				b := m.buildResidentialHouse(lot.Bounds, i)
				m.Buildings = append(m.Buildings, b)
			case DistrictCommercial:
				var b Building
				switch i {
				case 0:
					b = m.buildGroceryStore(lot.Bounds)
				case 1:
					b = m.buildPharmacy(lot.Bounds)
				case 2:
					b = m.buildPoliceStation(lot.Bounds)
				default:
					b = m.buildCommercialPlaza(lot.Bounds)
				}
				m.Buildings = append(m.Buildings, b)
			case DistrictIndustrial:
				var b Building
				switch i {
				case 0:
					b = m.buildWarehouse(lot.Bounds, "Warehouse A (Logistics)")
				case 1:
					b = m.buildWarehouse(lot.Bounds, "Warehouse B (Heavy Storage)")
				case 2:
					b = m.buildStorageUnits(lot.Bounds)
				default:
					b = m.buildScrapYardShed(lot.Bounds)
				}
				m.Buildings = append(m.Buildings, b)
			case DistrictPark:
				if i == 0 {
					b := m.buildParkPavilion(lot.Bounds)
					m.Buildings = append(m.Buildings, b)
				} else {
					m.decorateParkBlock(lot.Bounds, i)
				}
			}
		}
	}
}

// buildResidentialHouse constructs a multi-room suburban home with front driveway and fenced yard
func (m *Map) buildResidentialHouse(bounds Rect, houseIdx int) Building {
	// Calculate house footprint inside lot
	hw := 11
	hh := 9
	if hw > bounds.W-4 {
		hw = bounds.W - 4
	}
	if hh > bounds.H-4 {
		hh = bounds.H - 4
	}

	hx := bounds.X + 2
	hy := bounds.Y + 2

	// Pave Driveway from road across sidewalk into front yard
	for dy := bounds.Y; dy < hy+hh; dy++ {
		m.SetTile(hx+hw-2, dy, TileConcrete)
		m.SetTile(hx+hw-1, dy, TileConcrete)
	}

	// Floor & Outer Walls
	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileWoodFloor)
			}
		}
	}

	// Front entrance door (south)
	doorX := hx + 3
	doorY := hy + hh - 1
	m.SetTile(doorX, doorY, TileWoodFloor)

	// Back door (north) leading to backyard
	backDoorX := hx + 4
	backDoorY := hy
	m.SetTile(backDoorX, backDoorY, TileWoodFloor)

	// Interior Partition Walls: Living Room (left), Bedroom (top-right), Kitchen (bottom-right)
	splitX := hx + hw/2
	splitY := hy + hh/2

	// Vertical partition
	for y := hy + 1; y < hy+hh-1; y++ {
		m.SetTile(splitX, y, TileWall)
	}
	// Doorway in vertical partition
	m.SetTile(splitX, hy+2, TileWoodFloor)

	// Horizontal partition on right side (Kitchen vs Bedroom)
	for x := splitX + 1; x < hx+hw-1; x++ {
		m.SetTile(x, splitY, TileWall)
	}
	// Doorway to kitchen
	m.SetTile(splitX+2, splitY, TileWoodFloor)

	// Kitchen Floor: TileTileFloor
	for y := splitY + 1; y < hy+hh-1; y++ {
		for x := splitX + 1; x < hx+hw-1; x++ {
			m.SetTile(x, y, TileTileFloor)
		}
	}

	// Fenced Backyard (perimeter of back half of lot)
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		if m.GetTile(x, bounds.Y) == TileGrass {
			m.SetTile(x, bounds.Y, TileFence)
		}
	}
	for y := bounds.Y; y < hy; y++ {
		if m.GetTile(bounds.X, y) == TileGrass {
			m.SetTile(bounds.X, y, TileFence)
		}
		if m.GetTile(bounds.X+bounds.W-1, y) == TileGrass {
			m.SetTile(bounds.X+bounds.W-1, y, TileFence)
		}
	}

	b := Building{
		Type:     BuildingResidentialHouse,
		District: DistrictResidential,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Living Room", Bounds: Rect{X: hx + 1, Y: hy + 1, W: splitX - hx - 1, H: hh - 2}, Floor: TileWoodFloor},
			{Name: "Bedroom", Bounds: Rect{X: splitX + 1, Y: hy + 1, W: hx + hw - splitX - 2, H: splitY - hy - 1}, Floor: TileWoodFloor},
			{Name: "Kitchen", Bounds: Rect{X: splitX + 1, Y: splitY + 1, W: hx + hw - splitX - 2, H: hy + hh - splitY - 2}, Floor: TileTileFloor},
		},
		Doors: []Point{
			{X: float64(doorX)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
			{X: float64(backDoorX)*TileSize + 16, Y: float64(backDoorY)*TileSize + 16},
		},
	}

	return b
}

// buildGroceryStore creates a commercial supermarket with sales aisles and back storage
func (m *Map) buildGroceryStore(bounds Rect) Building {
	hw := bounds.W - 2
	hh := bounds.H - 2
	hx := bounds.X + 1
	hy := bounds.Y + 1

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	// Front entrance double door (south)
	doorX1 := hx + hw/2
	doorX2 := doorX1 + 1
	doorY := hy + hh - 1
	m.SetTile(doorX1, doorY, TileTileFloor)
	m.SetTile(doorX2, doorY, TileTileFloor)

	// Backroom Storage partition (north)
	splitY := hy + 4
	for x := hx + 1; x < hx+hw-1; x++ {
		m.SetTile(x, splitY, TileWall)
	}
	// Backroom interior door
	m.SetTile(hx+3, splitY, TileTileFloor)
	// Rear loading exit door
	m.SetTile(hx+3, hy, TileTileFloor)

	return Building{
		Type:     BuildingGroceryStore,
		District: DistrictCommercial,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Sales Floor", Bounds: Rect{X: hx + 1, Y: splitY + 1, W: hw - 2, H: hy + hh - splitY - 2}, Floor: TileTileFloor},
			{Name: "Back Storage", Bounds: Rect{X: hx + 1, Y: hy + 1, W: hw - 2, H: splitY - hy - 1}, Floor: TileTileFloor},
		},
		Doors: []Point{
			{X: float64(doorX1)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
		},
	}
}

// buildPharmacy creates a medical dispensary and clinic
func (m *Map) buildPharmacy(bounds Rect) Building {
	hw := bounds.W - 2
	hh := bounds.H - 2
	hx := bounds.X + 1
	hy := bounds.Y + 1

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	doorX := hx + hw/2
	doorY := hy + hh - 1
	m.SetTile(doorX, doorY, TileTileFloor)

	// Pharmacy Vault partition
	splitX := hx + hw/2
	for y := hy + 1; y < hy+hh-1; y++ {
		m.SetTile(splitX, y, TileWall)
	}
	m.SetTile(splitX, hy+3, TileTileFloor)

	return Building{
		Type:     BuildingPharmacy,
		District: DistrictCommercial,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Retail Clinic", Bounds: Rect{X: hx + 1, Y: hy + 1, W: splitX - hx - 1, H: hh - 2}, Floor: TileTileFloor},
			{Name: "Pharmacy Vault", Bounds: Rect{X: splitX + 1, Y: hy + 1, W: hx + hw - splitX - 2, H: hh - 2}, Floor: TileTileFloor},
		},
		Doors: []Point{
			{X: float64(doorX)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
		},
	}
}

// buildPoliceStation creates a reinforced police station with front desk, cells, and armory
func (m *Map) buildPoliceStation(bounds Rect) Building {
	hw := bounds.W - 2
	hh := bounds.H - 2
	hx := bounds.X + 1
	hy := bounds.Y + 1

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	doorX := hx + hw/2
	doorY := hy + hh - 1
	m.SetTile(doorX, doorY, TileTileFloor)

	// Split: Front Lobby vs Back Armory & Holding Cells
	splitY := hy + hh/2
	for x := hx + 1; x < hx+hw-1; x++ {
		m.SetTile(x, splitY, TileWall)
	}
	m.SetTile(hx+3, splitY, TileTileFloor)

	// Vertical split for Armory vs Holding Cell
	splitX := hx + hw/2
	for y := hy + 1; y < splitY; y++ {
		m.SetTile(splitX, y, TileWall)
	}
	m.SetTile(splitX, hy+2, TileTileFloor)

	return Building{
		Type:     BuildingPoliceStation,
		District: DistrictCommercial,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Lobby", Bounds: Rect{X: hx + 1, Y: splitY + 1, W: hw - 2, H: hy + hh - splitY - 2}, Floor: TileTileFloor},
			{Name: "Holding Cell", Bounds: Rect{X: hx + 1, Y: hy + 1, W: splitX - hx - 1, H: splitY - hy - 1}, Floor: TileTileFloor},
			{Name: "Police Armory", Bounds: Rect{X: splitX + 1, Y: hy + 1, W: hx + hw - splitX - 2, H: splitY - hy - 1}, Floor: TileTileFloor},
		},
		Doors: []Point{
			{X: float64(doorX)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
		},
	}
}

// buildCommercialPlaza creates retail shops
func (m *Map) buildCommercialPlaza(bounds Rect) Building {
	hw := bounds.W - 2
	hh := bounds.H - 2
	hx := bounds.X + 1
	hy := bounds.Y + 1

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	doorX := hx + hw/2
	doorY := hy + hh - 1
	m.SetTile(doorX, doorY, TileTileFloor)

	return Building{
		Type:     BuildingGroceryStore,
		District: DistrictCommercial,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Retail Store", Bounds: Rect{X: hx + 1, Y: hy + 1, W: hw - 2, H: hh - 2}, Floor: TileTileFloor},
		},
		Doors: []Point{
			{X: float64(doorX)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
		},
	}
}

// buildWarehouse creates an industrial logistics warehouse with concrete apron and loading doors
func (m *Map) buildWarehouse(bounds Rect, name string) Building {
	hw := bounds.W - 4
	hh := bounds.H - 4
	hx := bounds.X + 2
	hy := bounds.Y + 2

	// Concrete Apron
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		for x := bounds.X; x < bounds.X+bounds.W; x++ {
			m.SetTile(x, y, TileConcrete)
		}
	}

	// Security Perimeter Fence
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		m.SetTile(x, bounds.Y, TileFence)
		m.SetTile(x, bounds.Y+bounds.H-1, TileFence)
	}
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		m.SetTile(bounds.X, y, TileFence)
		m.SetTile(bounds.X+bounds.W-1, y, TileFence)
	}

	// Fence Entrance Gate (open)
	m.SetTile(bounds.X+bounds.W/2, bounds.Y+bounds.H-1, TileConcrete)
	m.SetTile(bounds.X+bounds.W/2+1, bounds.Y+bounds.H-1, TileConcrete)

	// Warehouse Building Walls & Floor
	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileConcrete)
			}
		}
	}

	// Wide Loading Door (2 tiles)
	doorX := hx + hw/2
	doorY := hy + hh - 1
	m.SetTile(doorX, doorY, TileConcrete)
	m.SetTile(doorX+1, doorY, TileConcrete)

	// Foreman Office partition
	splitX := hx + 5
	splitY := hy + 4
	for x := hx + 1; x < splitX; x++ {
		m.SetTile(x, splitY, TileWall)
	}
	for y := hy + 1; y < splitY; y++ {
		m.SetTile(splitX, y, TileWall)
	}
	m.SetTile(splitX, hy+2, TileConcrete)

	// Outdoor Debris in yard
	m.SetTile(bounds.X+1, bounds.Y+1, TileDebris)
	m.SetTile(bounds.X+2, bounds.Y+1, TileDebris)

	return Building{
		Type:     BuildingWarehouse,
		District: DistrictIndustrial,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Main Storage Floor", Bounds: Rect{X: hx + 1, Y: splitY + 1, W: hw - 2, H: hy + hh - splitY - 2}, Floor: TileConcrete},
			{Name: "Foreman Office", Bounds: Rect{X: hx + 1, Y: hy + 1, W: splitX - hx - 1, H: splitY - hy - 1}, Floor: TileConcrete},
		},
		Doors: []Point{
			{X: float64(doorX)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
		},
	}
}

// buildStorageUnits creates segmented storage bays
func (m *Map) buildStorageUnits(bounds Rect) Building {
	hw := bounds.W - 2
	hh := bounds.H - 2
	hx := bounds.X + 1
	hy := bounds.Y + 1

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileConcrete)
			}
		}
	}

	doorX := hx + hw/2
	doorY := hy + hh - 1
	m.SetTile(doorX, doorY, TileConcrete)

	return Building{
		Type:     BuildingStorageUnit,
		District: DistrictIndustrial,
		Bounds:   Rect{X: hx, Y: hy, W: hw, H: hh},
		Rooms: []Room{
			{Name: "Storage Bay", Bounds: Rect{X: hx + 1, Y: hy + 1, W: hw - 2, H: hh - 2}, Floor: TileConcrete},
		},
		Doors: []Point{
			{X: float64(doorX)*TileSize + 16, Y: float64(doorY)*TileSize + 16},
		},
	}
}

// buildScrapYardShed creates a scrap yard with outdoor debris and small tool shed
func (m *Map) buildScrapYardShed(bounds Rect) Building {
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		m.SetTile(x, bounds.Y, TileFence)
		m.SetTile(x, bounds.Y+bounds.H-1, TileFence)
	}
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		m.SetTile(bounds.X, y, TileFence)
		m.SetTile(bounds.X+bounds.W-1, y, TileFence)
	}
	m.SetTile(bounds.X+bounds.W/2, bounds.Y+bounds.H-1, TileGrass) // Gate

	// Debris clusters
	m.SetTile(bounds.X+2, bounds.Y+2, TileDebris)
	m.SetTile(bounds.X+3, bounds.Y+2, TileDebris)
	m.SetTile(bounds.X+2, bounds.Y+3, TileDebris)
	m.SetTile(bounds.X+bounds.W-3, bounds.Y+2, TileDebris)
	m.SetTile(bounds.X+bounds.W-4, bounds.Y+2, TileDebris)

	// Tool Shed in corner
	sx, sy, sw, sh := bounds.X+bounds.W-7, bounds.Y+bounds.H-6, 5, 4
	for y := sy; y < sy+sh; y++ {
		for x := sx; x < sx+sw; x++ {
			if x == sx || x == sx+sw-1 || y == sy || y == sy+sh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileWoodFloor)
			}
		}
	}
	m.SetTile(sx+2, sy+sh-1, TileWoodFloor) // Shed door

	return Building{
		Type:     BuildingStorageUnit,
		District: DistrictIndustrial,
		Bounds:   Rect{X: sx, Y: sy, W: sw, H: sh},
		Rooms: []Room{
			{Name: "Scrap Tool Shed", Bounds: Rect{X: sx + 1, Y: sy + 1, W: sw - 2, H: sh - 2}, Floor: TileWoodFloor},
		},
		Doors: []Point{
			{X: float64(sx+2)*TileSize + 16, Y: float64(sy+sh-1)*TileSize + 16},
		},
	}
}

// buildParkPavilion creates an open wooden park pavilion
func (m *Map) buildParkPavilion(bounds Rect) Building {
	pw := 7
	ph := 7
	px := bounds.X + (bounds.W-pw)/2
	py := bounds.Y + (bounds.H-ph)/2

	// Dirt trails leading to pavilion
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		m.SetTile(x, py+ph/2, TileDirt)
	}
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		m.SetTile(px+pw/2, y, TileDirt)
	}

	// Wood floor with perimeter fence
	for y := py; y < py+ph; y++ {
		for x := px; x < px+pw; x++ {
			m.SetTile(x, y, TileWoodFloor)
		}
	}
	for x := px; x < px+pw; x++ {
		m.SetTile(x, py, TileFence)
		m.SetTile(x, py+ph-1, TileFence)
	}
	for y := py; y < py+ph; y++ {
		m.SetTile(px, y, TileFence)
		m.SetTile(px+pw-1, y, TileFence)
	}

	// Openings on 4 sides
	m.SetTile(px+pw/2, py, TileWoodFloor)
	m.SetTile(px+pw/2, py+ph-1, TileWoodFloor)
	m.SetTile(px, py+ph/2, TileWoodFloor)
	m.SetTile(px+pw-1, py+ph/2, TileWoodFloor)

	return Building{
		Type:     BuildingParkPavilion,
		District: DistrictPark,
		Bounds:   Rect{X: px, Y: py, W: pw, H: ph},
		Rooms: []Room{
			{Name: "Pavilion", Bounds: Rect{X: px + 1, Y: py + 1, W: pw - 2, H: ph - 2}, Floor: TileWoodFloor},
		},
		Doors: []Point{
			{X: float64(px+pw/2)*TileSize + 16, Y: float64(py+ph-1)*TileSize + 16},
		},
	}
}

// decorateParkBlock adds winding dirt paths and natural tree clusters
func (m *Map) decorateParkBlock(bounds Rect, idx int) {
	// Winding dirt path
	midX, midY := bounds.Center()
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		m.SetTile(x, midY, TileDirt)
	}
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		m.SetTile(midX, y, TileDirt)
	}

	// Tree clusters in 4 quadrants of the park block
	coords := [][2]int{
		{bounds.X + 2, bounds.Y + 2},
		{bounds.X + 3, bounds.Y + 2},
		{bounds.X + 2, bounds.Y + 3},
		{bounds.X + bounds.W - 3, bounds.Y + 2},
		{bounds.X + bounds.W - 4, bounds.Y + 3},
		{bounds.X + 2, bounds.Y + bounds.H - 3},
		{bounds.X + 3, bounds.Y + bounds.H - 4},
		{bounds.X + bounds.W - 3, bounds.Y + bounds.H - 3},
		{bounds.X + bounds.W - 4, bounds.Y + bounds.H - 3},
	}
	for _, c := range coords {
		if m.GetTile(c[0], c[1]) == TileGrass {
			m.SetTile(c[0], c[1], TileTree)
		}
	}
}

// Phase 4: Environmental Props & Organic Tree Clusters
func (m *Map) generateEnvironmentalProps() {
	if m.Width < 20 || m.Height < 20 {
		return
	}

	// Place trees in open grass areas (avoiding roads, sidewalks, buildings)
	numTrees := (m.Width * m.Height) / 40
	for i := 0; i < numTrees; i++ {
		tx := 2 + rand.Intn(m.Width-4)
		ty := 2 + rand.Intn(m.Height-4)
		if m.GetTile(tx, ty) == TileGrass {
			m.SetTile(tx, ty, TileTree)
		}
	}
}

// Phase 5: Extract Structured Spawns
func (m *Map) extractSpawns() {
	// 1. Safe Player Spawn in House #1 (Bedroom or Living Room)
	foundPlayer := false
	for _, b := range m.Buildings {
		if b.Type == BuildingResidentialHouse && len(b.Rooms) > 0 {
			r := b.Rooms[0] // Living room
			cx, cy := r.Bounds.Center()
			m.PlayerSpawn = Point{
				X: float64(cx)*TileSize + 16,
				Y: float64(cy)*TileSize + 16,
			}
			foundPlayer = true
			break
		}
	}
	if !foundPlayer {
		m.PlayerSpawn = Point{
			X: float64(m.Width/2) * TileSize,
			Y: float64(m.Height/2) * TileSize,
		}
	}

	// 2. Thematic Contextual Loot Spawns
	for _, b := range m.Buildings {
		switch b.Type {
		case BuildingResidentialHouse:
			for _, r := range b.Rooms {
				cx, cy := r.Bounds.Center()
				px, py := float64(cx)*TileSize+16, float64(cy)*TileSize+16
				switch r.Name {
				case "Kitchen":
					m.LootSpawns = append(m.LootSpawns, LootSpawn{ItemType: "food", X: px, Y: py, RoomType: "kitchen"})
					m.LootSpawns = append(m.LootSpawns, LootSpawn{ItemType: "water", X: px + 10, Y: py, RoomType: "kitchen"})
				case "Bedroom":
					m.LootSpawns = append(m.LootSpawns, LootSpawn{ItemType: "armor", X: px, Y: py, RoomType: "bedroom"})
					m.LootSpawns = append(m.LootSpawns, LootSpawn{ItemType: "weapon", X: px + 10, Y: py, RoomType: "bedroom"})
				case "Living Room":
					m.LootSpawns = append(m.LootSpawns, LootSpawn{ItemType: "food", X: px, Y: py, RoomType: "living_room"})
				}
			}
		case BuildingGroceryStore:
			for _, r := range b.Rooms {
				cx, cy := r.Bounds.Center()
				px, py := float64(cx)*TileSize+16, float64(cy)*TileSize+16
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{ItemType: "food", X: px - 30, Y: py, RoomType: "grocery"},
					LootSpawn{ItemType: "food", X: px, Y: py, RoomType: "grocery"},
					LootSpawn{ItemType: "food", X: px + 30, Y: py, RoomType: "grocery"},
					LootSpawn{ItemType: "water", X: px - 15, Y: py + 20, RoomType: "grocery"},
					LootSpawn{ItemType: "water", X: px + 15, Y: py + 20, RoomType: "grocery"},
				)
			}
		case BuildingPharmacy:
			for _, r := range b.Rooms {
				cx, cy := r.Bounds.Center()
				px, py := float64(cx)*TileSize+16, float64(cy)*TileSize+16
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{ItemType: "water", X: px, Y: py, RoomType: "pharmacy"},
					LootSpawn{ItemType: "water", X: px + 15, Y: py, RoomType: "pharmacy"},
					LootSpawn{ItemType: "armor", X: px - 15, Y: py, RoomType: "pharmacy"},
				)
			}
		case BuildingPoliceStation:
			for _, r := range b.Rooms {
				cx, cy := r.Bounds.Center()
				px, py := float64(cx)*TileSize+16, float64(cy)*TileSize+16
				if r.Name == "Police Armory" {
					m.LootSpawns = append(m.LootSpawns,
						LootSpawn{ItemType: "shotgun", X: px - 15, Y: py, RoomType: "armory"},
						LootSpawn{ItemType: "ammo", X: px, Y: py, RoomType: "armory"},
						LootSpawn{ItemType: "ammo", X: px + 15, Y: py, RoomType: "armory"},
						LootSpawn{ItemType: "armor", X: px, Y: py + 15, RoomType: "armory"},
						LootSpawn{ItemType: "axe", X: px - 15, Y: py + 15, RoomType: "armory"},
					)
				} else {
					m.LootSpawns = append(m.LootSpawns,
						LootSpawn{ItemType: "weapon", X: px, Y: py, RoomType: "lobby"},
					)
				}
			}
		case BuildingWarehouse:
			for _, r := range b.Rooms {
				cx, cy := r.Bounds.Center()
				px, py := float64(cx)*TileSize+16, float64(cy)*TileSize+16
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{ItemType: "axe", X: px - 20, Y: py, RoomType: "warehouse"},
					LootSpawn{ItemType: "axe", X: px + 20, Y: py, RoomType: "warehouse"},
					LootSpawn{ItemType: "shotgun", X: px, Y: py - 20, RoomType: "warehouse"},
					LootSpawn{ItemType: "ammo", X: px, Y: py + 20, RoomType: "warehouse"},
					LootSpawn{ItemType: "armor", X: px, Y: py, RoomType: "warehouse"},
				)
			}
		case BuildingParkPavilion:
			for _, r := range b.Rooms {
				cx, cy := r.Bounds.Center()
				px, py := float64(cx)*TileSize+16, float64(cy)*TileSize+16
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{ItemType: "food", X: px - 10, Y: py, RoomType: "outdoor"},
					LootSpawn{ItemType: "water", X: px + 10, Y: py, RoomType: "outdoor"},
				)
			}
		}
	}

	// 3. Thematic Zombie Horde Distribution (150 total)
	targetZombies := 150
	spawnAttempts := 0
	for len(m.ZombieSpawns) < targetZombies && spawnAttempts < 2000 {
		spawnAttempts++
		tx := 3 + rand.Intn(m.Width-6)
		ty := 3 + rand.Intn(m.Height-6)
		t := m.GetTile(tx, ty)

		if t.IsSolid() {
			continue
		}

		zx := float64(tx)*TileSize + 16
		zy := float64(ty)*TileSize + 16

		// Distance check from player spawn (> 350 pixels)
		dx := zx - m.PlayerSpawn.X
		dy := zy - m.PlayerSpawn.Y
		if math.Hypot(dx, dy) < 350 {
			continue
		}

		// Determine district
		distType := DistrictResidential
		if tx >= m.Width/2 && ty < m.Height/2 {
			distType = DistrictCommercial
		} else if tx >= m.Width/2 && ty >= m.Height/2 {
			distType = DistrictIndustrial
		} else if tx < m.Width/2 && ty >= m.Height/2 {
			distType = DistrictPark
		}

		// Runner chance varies by district
		runnerChance := 0.15
		if distType == DistrictIndustrial {
			runnerChance = 0.30
		} else if distType == DistrictCommercial {
			runnerChance = 0.20
		}

		isRunner := rand.Float64() < runnerChance

		m.ZombieSpawns = append(m.ZombieSpawns, ZombieSpawn{
			X:        zx,
			Y:        zy,
			IsRunner: isRunner,
			District: distType,
		})
	}
}

// CalculateFOV casts rays in 360 degrees to calculate visibility and persistent fog of war
func (m *Map) CalculateFOV(playerX, playerY float64, radiusTiles int) {
	for i := range m.Visible {
		m.Visible[i] = false
	}

	px := int(playerX) / TileSize
	py := int(playerY) / TileSize

	if px < 0 || px >= m.Width || py < 0 || py >= m.Height {
		return
	}

	m.Visible[py*m.Width+px] = true
	m.Explored[py*m.Width+px] = true

	rays := radiusTiles * 8
	for i := 0; i < rays; i++ {
		angle := (float64(i) / float64(rays)) * 2 * math.Pi
		dirX := math.Cos(angle)
		dirY := math.Sin(angle)

		cx, cy := float64(px)+0.5, float64(py)+0.5
		for step := 0; step < radiusTiles; step++ {
			cx += dirX
			cy += dirY

			tx, ty := int(cx), int(cy)
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
				break
			}

			m.Visible[ty*m.Width+tx] = true
			m.Explored[ty*m.Width+tx] = true

			if m.GetTile(tx, ty).BlocksVision() {
				break
			}
		}
	}
}

// GetTile returns the tile type at (x, y) with bounds safety
func (m *Map) GetTile(x, y int) TileType {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return TileWall
	}
	return m.Tiles[y*m.Width+x]
}

// SetTile writes the tile type at (x, y) with bounds safety
func (m *Map) SetTile(x, y int, t TileType) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}
	m.Tiles[y*m.Width+x] = t
}

// IsColliding checks if an AABB bounding box intersects any solid tile in world coordinates
func (m *Map) IsColliding(rectX, rectY, rectW, rectH float64) bool {
	minTileX := int(rectX) / TileSize
	minTileY := int(rectY) / TileSize
	maxTileX := int(rectX+rectW) / TileSize
	maxTileY := int(rectY+rectH) / TileSize

	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			t := m.GetTile(x, y)
			if t.IsSolid() {
				return true
			}
		}
	}
	return false
}
