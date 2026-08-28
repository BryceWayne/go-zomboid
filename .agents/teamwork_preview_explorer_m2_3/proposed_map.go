package world_proposed

import (
	"math"
	"math/rand"
)

// TileType represents the type of terrain or obstacle in a map cell.
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

// IsSolid returns true if the tile blocks physical entity movement and collision.
func (t TileType) IsSolid() bool {
	switch t {
	case TileWall, TileTree, TileFence, TileDebris:
		return true
	default:
		return false
	}
}

// BlocksVision returns true if the tile occludes FOV raycasting.
func (t TileType) BlocksVision() bool {
	return t == TileWall
}

// IsFloor returns true if the tile is a flat surface drawn during the ground diamond pass.
func (t TileType) IsFloor() bool {
	switch t {
	case TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor:
		return true
	default:
		return false
	}
}

// String returns the name of the tile type for debugging and logging.
func (t TileType) String() string {
	switch t {
	case TileGrass:
		return "Grass"
	case TileWall:
		return "Wall"
	case TileDirt:
		return "Dirt"
	case TileWoodFloor:
		return "WoodFloor"
	case TileTree:
		return "Tree"
	case TileAsphalt:
		return "Asphalt"
	case TileConcrete:
		return "Concrete"
	case TileTileFloor:
		return "TileFloor"
	case TileFence:
		return "Fence"
	case TileDebris:
		return "Debris"
	default:
		return "Unknown"
	}
}

// Point represents a 2D integer coordinate in tile or pixel space.
type Point struct {
	X, Y int
}

// FloatPoint represents a 2D float coordinate in world pixel space.
type FloatPoint struct {
	X, Y float64
}

// LootSpawn represents a structured contextual loot placement.
type LootSpawn struct {
	Type string  // "food", "water", "weapon", "axe", "shotgun", "ammo", "armor"
	X, Y float64 // world pixel coordinates
}

// BuildingType identifies the architectural archetype of a building.
type BuildingType string

const (
	BuildingResidential BuildingType = "Residential"
	BuildingGrocery     BuildingType = "Grocery"
	BuildingPolice      BuildingType = "Police"
	BuildingPharmacy    BuildingType = "Pharmacy"
	BuildingWarehouse   BuildingType = "Warehouse"
)

// RoomType identifies the functional zone within a building.
type RoomType string

const (
	RoomLiving   RoomType = "Living"
	RoomKitchen  RoomType = "Kitchen"
	RoomBedroom  RoomType = "Bedroom"
	RoomStore    RoomType = "Store"
	RoomStorage  RoomType = "Storage"
	RoomOffice   RoomType = "Office"
	RoomArmory   RoomType = "Armory"
	RoomPharmacy RoomType = "Pharmacy"
)

// Room represents a partitioned sub-area within a building.
type Room struct {
	Type RoomType
	X, Y int // tile grid coordinates
	W, H int // tile grid dimensions
}

// Building represents a structured structure with rooms and doors.
type Building struct {
	Type  BuildingType
	X, Y  int // tile grid coordinates
	W, H  int // tile grid dimensions
	Rooms []Room
	Doors []Point
}

// Map stores the world tile grid, visibility/exploration state, and structured spawn metadata.
type Map struct {
	Width, Height int
	Tiles         []TileType
	Visible       []bool
	Explored      []bool

	// Contextual Town Metadata & Spawn Points
	PlayerSpawn  FloatPoint
	Buildings    []Building
	LootSpawns   []LootSpawn
	ZombieSpawns []FloatPoint
}

// NewMap constructs a procedurally generated town map with zoned districts, roads, multi-room buildings, and contextual spawns.
func NewMap(width, height int) *Map {
	m := &Map{
		Width:        width,
		Height:       height,
		Tiles:        make([]TileType, width*height),
		Visible:      make([]bool, width*height),
		Explored:     make([]bool, width*height),
		Buildings:    make([]Building, 0),
		LootSpawns:   make([]LootSpawn, 0),
		ZombieSpawns: make([]FloatPoint, 0),
	}

	// 1. Fill base terrain with grass, surrounded by boundary perimeter walls
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileGrass)
			}
		}
	}

	if width < 30 || height < 30 {
		m.generateSmallFallback(width, height)
		return m
	}

	// 2. Procedural Road Network
	midX := width / 2
	midY := height / 2

	// East-West Main Avenue (Asphalt with Concrete sidewalks)
	for x := 1; x < width-1; x++ {
		m.SetTile(x, midY-2, TileConcrete)
		m.SetTile(x, midY-1, TileAsphalt)
		m.SetTile(x, midY, TileAsphalt)
		m.SetTile(x, midY+1, TileAsphalt)
		m.SetTile(x, midY+2, TileConcrete)
	}

	// North-South Main Boulevard
	for y := 1; y < height-1; y++ {
		m.SetTile(midX-2, y, TileConcrete)
		m.SetTile(midX-1, y, TileAsphalt)
		m.SetTile(midX, y, TileAsphalt)
		m.SetTile(midX+1, y, TileAsphalt)
		m.SetTile(midX+2, y, TileConcrete)
	}

	// Secondary Residential Access Road (North)
	resRoadY := int(float64(height) * 0.25)
	if resRoadY > 2 && resRoadY < midY-4 {
		for x := 4; x < midX-3; x++ {
			m.SetTile(x, resRoadY-1, TileConcrete)
			m.SetTile(x, resRoadY, TileAsphalt)
			m.SetTile(x, resRoadY+1, TileConcrete)
		}
	}

	// Secondary Industrial Access Road (South)
	indRoadY := int(float64(height) * 0.75)
	if indRoadY > midY+4 && indRoadY < height-3 {
		for x := midX + 3; x < width-4; x++ {
			m.SetTile(x, indRoadY-1, TileConcrete)
			m.SetTile(x, indRoadY, TileAsphalt)
			m.SetTile(x, indRoadY+1, TileConcrete)
		}
	}

	// Dirt walking trails in parks and backyards
	for y := 4; y < resRoadY; y++ {
		m.SetTile(6, y, TileDirt)
	}
	for y := indRoadY + 1; y < height-4; y++ {
		m.SetTile(width-8, y, TileDirt)
	}

	// 3. Synthesize Multi-Room Building Archetypes

	// A. Residential District (North-West Quadrant)
	// House 1 (Player Safe Starting House - 14x11)
	h1X, h1Y := 10, 8
	if h1X+14 < midX-3 && h1Y+11 < resRoadY {
		h1 := m.buildResidentialHouse(h1X, h1Y, 14, 11)
		m.Buildings = append(m.Buildings, h1)
		m.buildFencedYard(h1X-2, h1Y-2, 18, 15)
	}

	// House 2 (Neighbor House - 12x10)
	h2X, h2Y := 28, 8
	if h2X+12 < midX-3 && h2Y+10 < resRoadY {
		h2 := m.buildResidentialHouse(h2X, h2Y, 12, 10)
		m.Buildings = append(m.Buildings, h2)
		m.buildFencedYard(h2X-2, h2Y-2, 16, 14)
	}

	// House 3 (Corner Home - 12x10)
	h3X, h3Y := 10, resRoadY+3
	if h3X+12 < midX-3 && h3Y+10 < midY-3 {
		h3 := m.buildResidentialHouse(h3X, h3Y, 12, 10)
		m.Buildings = append(m.Buildings, h3)
	}

	// B. Commercial District (North-East Quadrant)
	// Grocery Store (18x12)
	gX, gY := midX+6, 8
	if gX+18 < width-3 && gY+12 < midY-3 {
		bGrocery := m.buildCommercialBuilding(gX, gY, 18, 12, BuildingGrocery)
		m.Buildings = append(m.Buildings, bGrocery)
		// Concrete storefront apron
		m.fillArea(gX, gY+12, 18, 2, TileConcrete)
	}

	// Pharmacy / Clinic (14x10)
	pX, pY := midX+28, 8
	if pX+14 < width-2 && pY+10 < midY-3 {
		bPharmacy := m.buildCommercialBuilding(pX, pY, 14, 10, BuildingPharmacy)
		m.Buildings = append(m.Buildings, bPharmacy)
		m.fillArea(pX, pY+10, 14, 2, TileConcrete)
	}

	// C. Municipal & Defense District (South-West Quadrant)
	// Police Station & Armory (16x13)
	polX, polY := 10, midY+6
	if polX+16 < midX-3 && polY+13 < height-3 {
		bPolice := m.buildPoliceStation(polX, polY, 16, 13)
		m.Buildings = append(m.Buildings, bPolice)

		// Police motor pool courtyard
		m.fillArea(polX+17, polY, 10, 10, TileAsphalt)
		m.buildFencedYard(polX+17, polY, 10, 10)
	}

	// House 4 (Suburban House South - 14x10)
	h4X, h4Y := 10, height-18
	if h4X+14 < midX-3 && h4Y+10 < height-2 {
		h4 := m.buildResidentialHouse(h4X, h4Y, 14, 10)
		m.Buildings = append(m.Buildings, h4)
	}

	// D. Industrial District (South-East Quadrant)
	// Warehouse Depot (20x14)
	wX, wY := midX+6, midY+6
	if wX+20 < width-3 && wY+14 < height-3 {
		bWarehouse := m.buildWarehouse(wX, wY, 20, 14)
		m.Buildings = append(m.Buildings, bWarehouse)
		m.buildFencedYard(wX-2, wY-2, 24, 18)
	}

	// 4. Outdoor Props: Debris, Trees, and Vegetation
	m.placeEnvironmentalProps(width, height, midX, midY)

	// 5. Contextual Spawn Points Extraction
	// Safe Player Spawn: inside House 1 living room (hx=10, hy=8 -> x=13, y=11 is living room center)
	playerTileX := 13
	playerTileY := 11
	if m.GetTile(playerTileX, playerTileY).IsSolid() {
		playerTileX = midX
		playerTileY = midY
	}
	m.PlayerSpawn = FloatPoint{
		X: float64(playerTileX)*TileSize + 16.0,
		Y: float64(playerTileY)*TileSize + 16.0,
	}

	// Generate Thematic Loot Spawns
	m.generateThematicLoot(playerTileX, playerTileY)

	// Generate Safe Zombie Spawns
	m.generateZombieSpawns(140)

	return m
}

func (m *Map) generateSmallFallback(width, height int) {
	midX := width / 2
	midY := height / 2
	for x := 1; x < width-1; x++ {
		m.SetTile(x, midY, TileDirt)
	}
	for y := 1; y < height-1; y++ {
		m.SetTile(midX, y, TileDirt)
	}
	m.PlayerSpawn = FloatPoint{
		X: float64(midX)*TileSize + 16.0,
		Y: float64(midY)*TileSize + 16.0,
	}
	m.LootSpawns = append(m.LootSpawns,
		LootSpawn{Type: "weapon", X: m.PlayerSpawn.X + 32, Y: m.PlayerSpawn.Y},
		LootSpawn{Type: "food", X: m.PlayerSpawn.X, Y: m.PlayerSpawn.Y + 32},
		LootSpawn{Type: "water", X: m.PlayerSpawn.X - 32, Y: m.PlayerSpawn.Y},
	)
}

func (m *Map) buildResidentialHouse(hx, hy, hw, hh int) Building {
	b := Building{
		Type:  BuildingResidential,
		X:     hx,
		Y:     hy,
		W:     hw,
		H:     hh,
		Rooms: make([]Room, 0),
		Doors: make([]Point, 0),
	}

	// Outer walls
	m.fillArea(hx, hy, hw, hh, TileWall)

	// Interior floors
	m.fillArea(hx+1, hy+1, hw-2, hh-2, TileWoodFloor)

	// Internal vertical partition wall
	splitX := hx + hw/2
	for y := hy + 1; y < hy+hh-1; y++ {
		m.SetTile(splitX, y, TileWall)
	}

	// Kitchen floor on right side
	m.fillArea(splitX+1, hy+1, hw-(splitX-hx)-2, hh-2, TileTileFloor)

	// Doors: Front entrance at bottom left, interior door in partition wall
	doorFront := Point{X: hx + hw/4, Y: hy + hh - 1}
	doorInterior := Point{X: splitX, Y: hy + hh/2}
	m.SetTile(doorFront.X, doorFront.Y, TileWoodFloor)
	m.SetTile(doorInterior.X, doorInterior.Y, TileWoodFloor)

	b.Doors = append(b.Doors, doorFront, doorInterior)

	// Rooms metadata
	b.Rooms = append(b.Rooms,
		Room{Type: RoomLiving, X: hx + 1, Y: hy + 1, W: splitX - hx - 1, H: hh - 2},
		Room{Type: RoomKitchen, X: splitX + 1, Y: hy + 1, W: hw - (splitX - hx) - 2, H: hh - 2},
	)

	return b
}

func (m *Map) buildCommercialBuilding(bx, by, bw, bh int, bType BuildingType) Building {
	b := Building{
		Type:  bType,
		X:     bx,
		Y:     by,
		W:     bw,
		H:     bh,
		Rooms: make([]Room, 0),
		Doors: make([]Point, 0),
	}

	// Outer walls
	m.fillArea(bx, by, bw, bh, TileWall)

	// Main sales floor
	floorType := TileTileFloor
	if bType == BuildingPharmacy {
		floorType = TileTileFloor
	}
	m.fillArea(bx+1, by+1, bw-2, bh-2, floorType)

	// Storage / back room partition wall
	splitY := by + int(float64(bh)*0.4)
	for x := bx + 1; x < bx+bw-1; x++ {
		m.SetTile(x, splitY, TileWall)
	}
	m.fillArea(bx+1, by+1, bw-2, splitY-by-1, TileConcrete)

	// Doors: Front main double-door, back room door
	doorFront1 := Point{X: bx + bw/2, Y: by + bh - 1}
	doorFront2 := Point{X: bx + bw/2 - 1, Y: by + bh - 1}
	doorBack := Point{X: bx + bw/2, Y: splitY}

	m.SetTile(doorFront1.X, doorFront1.Y, floorType)
	m.SetTile(doorFront2.X, doorFront2.Y, floorType)
	m.SetTile(doorBack.X, doorBack.Y, TileConcrete)

	b.Doors = append(b.Doors, doorFront1, doorFront2, doorBack)

	rStoreType := RoomStore
	rBackType := RoomStorage
	if bType == BuildingPharmacy {
		rStoreType = RoomPharmacy
		rBackType = RoomStorage
	}

	b.Rooms = append(b.Rooms,
		Room{Type: rBackType, X: bx + 1, Y: by + 1, W: bw - 2, H: splitY - by - 1},
		Room{Type: rStoreType, X: bx + 1, Y: splitY + 1, W: bw - 2, H: bh - (splitY - by) - 2},
	)

	return b
}

func (m *Map) buildPoliceStation(bx, by, bw, bh int) Building {
	b := Building{
		Type:  BuildingPolice,
		X:     bx,
		Y:     by,
		W:     bw,
		H:     bh,
		Rooms: make([]Room, 0),
		Doors: make([]Point, 0),
	}

	// Outer walls
	m.fillArea(bx, by, bw, bh, TileWall)

	// Office floor
	m.fillArea(bx+1, by+1, bw-2, bh-2, TileTileFloor)

	// Heavy Armory partition (top-right quarter)
	splitX := bx + int(float64(bw)*0.6)
	splitY := by + int(float64(bh)*0.5)

	for x := splitX; x < bx+bw-1; x++ {
		m.SetTile(x, splitY, TileWall)
	}
	for y := by + 1; y <= splitY; y++ {
		m.SetTile(splitX, y, TileWall)
	}

	// Armory floor: Concrete
	m.fillArea(splitX+1, by+1, bw-(splitX-bx)-2, splitY-by-1, TileConcrete)

	// Doors: Main entrance, Armory door
	doorFront := Point{X: bx + bw/3, Y: by + bh - 1}
	doorArmory := Point{X: splitX, Y: by + (splitY-by)/2}

	m.SetTile(doorFront.X, doorFront.Y, TileTileFloor)
	m.SetTile(doorArmory.X, doorArmory.Y, TileConcrete)

	b.Doors = append(b.Doors, doorFront, doorArmory)

	b.Rooms = append(b.Rooms,
		Room{Type: RoomOffice, X: bx + 1, Y: by + 1, W: bw - 2, H: bh - 2},
		Room{Type: RoomArmory, X: splitX + 1, Y: by + 1, W: bw - (splitX - bx) - 2, H: splitY - by - 1},
	)

	return b
}

func (m *Map) buildWarehouse(bx, by, bw, bh int) Building {
	b := Building{
		Type:  BuildingWarehouse,
		X:     bx,
		Y:     by,
		W:     bw,
		H:     bh,
		Rooms: make([]Room, 0),
		Doors: make([]Point, 0),
	}

	// Outer walls
	m.fillArea(bx, by, bw, bh, TileWall)

	// Concrete floor
	m.fillArea(bx+1, by+1, bw-2, bh-2, TileConcrete)

	// Wide Loading Bay Doors (bottom) and Side Door
	doorMain1 := Point{X: bx + bw/2 - 1, Y: by + bh - 1}
	doorMain2 := Point{X: bx + bw/2, Y: by + bh - 1}
	doorMain3 := Point{X: bx + bw/2 + 1, Y: by + bh - 1}
	doorSide := Point{X: bx, Y: by + bh/2}

	m.SetTile(doorMain1.X, doorMain1.Y, TileConcrete)
	m.SetTile(doorMain2.X, doorMain2.Y, TileConcrete)
	m.SetTile(doorMain3.X, doorMain3.Y, TileConcrete)
	m.SetTile(doorSide.X, doorSide.Y, TileConcrete)

	b.Doors = append(b.Doors, doorMain1, doorMain2, doorMain3, doorSide)

	// Add internal crates / debris clusters
	m.SetTile(bx+3, by+3, TileDebris)
	m.SetTile(bx+4, by+3, TileDebris)
	m.SetTile(bx+bw-4, by+3, TileDebris)
	m.SetTile(bx+bw-4, by+4, TileDebris)

	b.Rooms = append(b.Rooms,
		Room{Type: RoomStorage, X: bx + 1, Y: by + 1, W: bw - 2, H: bh - 2},
	)

	return b
}

func (m *Map) buildFencedYard(fx, fy, fw, fh int) {
	for x := fx; x < fx+fw; x++ {
		for y := fy; y < fy+fh; y++ {
			if x <= 0 || x >= m.Width-1 || y <= 0 || y >= m.Height-1 {
				continue
			}
			isBorder := (x == fx || x == fx+fw-1 || y == fy || y == fy+fh-1)
			if isBorder {
				cur := m.GetTile(x, y)
				if cur == TileGrass || cur == TileDirt {
					isGate := (x == fx+fw/2 && y == fy+fh-1) || (x == fx && y == fy+fh/2)
					if !isGate {
						m.SetTile(x, y, TileFence)
					}
				}
			}
		}
	}
}

func (m *Map) fillArea(x, y, w, h int, t TileType) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			if px > 0 && px < m.Width-1 && py > 0 && py < m.Height-1 {
				m.SetTile(px, py, t)
			}
		}
	}
}

func (m *Map) placeEnvironmentalProps(width, height, midX, midY int) {
	debrisList := []Point{
		{midX + 8, midY + 22},
		{midX + 9, midY + 22},
		{midX + 27, midY + 16},
		{midX + 27, midY + 17},
		{26, midY + 18},
		{27, midY + 18},
		{midX + 12, 23},
		{midX + 13, 23},
	}
	for _, p := range debrisList {
		if p.X > 0 && p.X < width-1 && p.Y > 0 && p.Y < height-1 {
			cur := m.GetTile(p.X, p.Y)
			if cur == TileGrass || cur == TileConcrete {
				m.SetTile(p.X, p.Y, TileDebris)
			}
		}
	}

	parkTrees := []Point{
		{4, 4}, {5, 4}, {4, 5}, {6, 6},
		{24, 4}, {25, 5}, {26, 4},
		{4, 30}, {5, 31}, {6, 30},
		{midX + 28, 4}, {midX + 29, 5}, {midX + 31, 6},
		{midX + 32, 28}, {midX + 33, 30}, {midX + 31, 31},
		{4, height - 8}, {5, height - 7}, {6, height - 9},
		{midX + 28, height - 8}, {midX + 30, height - 7},
	}
	for _, p := range parkTrees {
		if p.X > 0 && p.X < width-1 && p.Y > 0 && p.Y < height-1 {
			if m.GetTile(p.X, p.Y) == TileGrass {
				m.SetTile(p.X, p.Y, TileTree)
			}
		}
	}

	for i := 0; i < 120; i++ {
		tx := 2 + rand.Intn(width-4)
		ty := 2 + rand.Intn(height-4)
		if m.GetTile(tx, ty) == TileGrass {
			if math.Abs(float64(tx-midX)) > 3 && math.Abs(float64(ty-midY)) > 3 {
				m.SetTile(tx, ty, TileTree)
			}
		}
	}
}

func (m *Map) generateThematicLoot(playerTileX, playerTileY int) {
	// Guaranteed starter loot safely inside player living room / kitchen
	m.addLootIfWalkable("weapon", float64(playerTileX)*TileSize+16, float64(playerTileY-1)*TileSize+16)
	m.addLootIfWalkable("food", float64(playerTileX-1)*TileSize+16, float64(playerTileY)*TileSize+16)
	m.addLootIfWalkable("water", float64(playerTileX)*TileSize+16, float64(playerTileY+1)*TileSize+16)

	// Contextual room loot distribution
	for _, b := range m.Buildings {
		for _, r := range b.Rooms {
			centerX := float64(r.X+r.W/2)*TileSize + 16.0
			centerY := float64(r.Y+r.H/2)*TileSize + 16.0

			switch r.Type {
			case RoomKitchen:
				m.addLootIfWalkable("food", centerX-16, centerY)
				m.addLootIfWalkable("water", centerX+16, centerY)
			case RoomLiving, RoomBedroom:
				m.addLootIfWalkable("food", centerX-16, centerY-16)
				m.addLootIfWalkable("water", centerX+16, centerY+16)
				m.addLootIfWalkable("weapon", centerX, centerY)
			case RoomStore:
				m.addLootIfWalkable("food", centerX-32, centerY)
				m.addLootIfWalkable("food", centerX+32, centerY)
				m.addLootIfWalkable("water", centerX, centerY-32)
				m.addLootIfWalkable("water", centerX, centerY+32)
			case RoomPharmacy:
				m.addLootIfWalkable("food", centerX-24, centerY)
				m.addLootIfWalkable("water", centerX+24, centerY)
				m.addLootIfWalkable("armor", centerX, centerY)
			case RoomArmory:
				m.addLootIfWalkable("weapon", centerX-24, centerY-24)
				m.addLootIfWalkable("axe", centerX+24, centerY-24)
				m.addLootIfWalkable("shotgun", centerX-24, centerY+24)
				m.addLootIfWalkable("ammo", centerX+24, centerY+24)
				m.addLootIfWalkable("armor", centerX, centerY)
			case RoomOffice:
				m.addLootIfWalkable("food", centerX-20, centerY)
				m.addLootIfWalkable("weapon", centerX+20, centerY)
				m.addLootIfWalkable("ammo", centerX, centerY+20)
			case RoomStorage:
				m.addLootIfWalkable("axe", centerX-32, centerY)
				m.addLootIfWalkable("ammo", centerX+32, centerY)
				m.addLootIfWalkable("armor", centerX, centerY-32)
				m.addLootIfWalkable("weapon", centerX, centerY+32)
			}
		}
	}
}

func (m *Map) addLootIfWalkable(lootType string, x, y float64) {
	tx := int(x) / TileSize
	ty := int(y) / TileSize
	if tx >= 0 && tx < m.Width && ty >= 0 && ty < m.Height {
		if !m.GetTile(tx, ty).IsSolid() {
			m.LootSpawns = append(m.LootSpawns, LootSpawn{
				Type: lootType,
				X:    x,
				Y:    y,
			})
		}
	}
}

func (m *Map) generateZombieSpawns(targetCount int) {
	spawned := 0
	attempts := 0
	maxAttempts := targetCount * 30

	for spawned < targetCount && attempts < maxAttempts {
		attempts++
		tx := 2 + rand.Intn(m.Width-4)
		ty := 2 + rand.Intn(m.Height-4)

		if m.GetTile(tx, ty).IsSolid() {
			continue
		}

		zx := float64(tx)*TileSize + 16.0
		zy := float64(ty)*TileSize + 16.0

		dx := zx - m.PlayerSpawn.X
		dy := zy - m.PlayerSpawn.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 350.0 {
			continue // Safe player perimeter
		}

		m.ZombieSpawns = append(m.ZombieSpawns, FloatPoint{X: zx, Y: zy})
		spawned++
	}
}

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

			if m.BlocksVision(tx, ty) {
				break
			}
		}
	}
}

func (m *Map) BlocksVision(x, y int) bool {
	return m.GetTile(x, y).BlocksVision()
}

func (m *Map) GetTile(x, y int) TileType {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return TileWall
	}
	return m.Tiles[y*m.Width+x]
}

func (m *Map) SetTile(x, y int, t TileType) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}
	m.Tiles[y*m.Width+x] = t
}

func (m *Map) IsColliding(rectX, rectY, rectW, rectH float64) bool {
	// Boundary check
	if rectX < 0 || rectY < 0 || rectX+rectW > float64(m.Width*TileSize) || rectY+rectH > float64(m.Height*TileSize) {
		return true
	}

	minTileX := int(rectX) / TileSize
	minTileY := int(rectY) / TileSize
	maxTileX := int(rectX+rectW) / TileSize
	maxTileY := int(rectY+rectH) / TileSize

	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			if m.GetTile(x, y).IsSolid() {
				return true
			}
		}
	}
	return false
}
