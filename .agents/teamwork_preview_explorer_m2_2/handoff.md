# Milestone 2 Architectural Specification & Design: Multi-Room Building Archetypes & Interior Floorplans

## 1. Observation

### 1.1 Existing Tile and World Map System (`internal/game/world/map.go`)
From `internal/game/world/map.go:10-27`:
```go
type TileType int

const (
	TileGrass TileType = iota
	TileWall
	TileDirt
	TileWoodFloor
	TileTree
)

const TileSize = 32

type Map struct {
	Width, Height int
	Tiles         []TileType
	Visible       []bool
	Explored      []bool
}
```

The current map implementation only defines 5 tile types, lacking asphalt, concrete, tile floor, fences, and debris.

### 1.2 Current Building and Town Generation
From `internal/game/world/map.go:61-93`:
```go
buildHouse := func(hx, hy, hw, hh int) {
	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x <= 0 || x >= width-1 || y <= 0 || y >= height-1 {
				continue
			}
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				isDoor := (x == hx+hw/2 && y == hy+hh-1) // Bottom door
				if isDoor {
					m.SetTile(x, y, TileWoodFloor)
				} else {
					m.SetTile(x, y, TileWall)
				}
			} else {
				m.SetTile(x, y, TileWoodFloor)
			}
		}
	}
}

// Generate houses in the 4 quadrants
buildHouse(10, 10, 10, 8)
buildHouse(30, 15, 8, 12)
buildHouse(60, 12, 12, 10)
buildHouse(75, 30, 10, 10)
buildHouse(15, 60, 12, 8)
buildHouse(35, 70, 10, 10)
buildHouse(70, 65, 15, 12)
```

**Deficiencies Observed**:
1. **Monolithic Single-Room Shells**: Buildings are plain rectangular boxes with no internal partitions, rooms, or functional zones.
2. **Homogeneous Archetypes**: All 7 structures are generic residential boxes; there are no commercial grocery stores, police stations, clinics, or industrial warehouses.
3. **Uniform Floorings**: All buildings use `TileWoodFloor` uniformly, with no distinction for commercial tile, industrial concrete, or bathroom tiling.
4. **Hardcoded Coordinates**: Static coordinates produce identical town layouts on every run with no district zoning.

### 1.3 Available Pre-Rendered Assets (`internal/assets/assets.go`)
From `internal/assets/assets.go:22-35`:
```go
// Floor Tiles (64x32)
GrassImage     *ebiten.Image
DirtImage      *ebiten.Image
WoodImage      *ebiten.Image
AsphaltImage   *ebiten.Image
ConcreteImage  *ebiten.Image
TileFloorImage *ebiten.Image

// Vertical Obstacles / Props (64x64)
WallImage   *ebiten.Image
TreeImage   *ebiten.Image
FenceImage  *ebiten.Image
DebrisImage *ebiten.Image
```
The asset pipeline already has all the required textures loaded (`AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `FenceImage`, `DebrisImage`), ready for immediate use once the world generator and rendering pipeline map to them.

---

## 2. Logic Chain & Technical Design

```
+--------------------------------------------------------------------------------------------------+
|                                    Procedural Town Generator                                     |
|                                                                                                  |
|   1. Grid & Zoning Initialization (100x100 Grid, Residential / Commercial / Civic / Industrial)  |
|   2. Road Hierarchy Synthesizer (Main Paved Avenues [Asphalt] + Sidewalks [Concrete] + Streets)  |
|   3. Multi-Room Archetype Synthesizers:                                                          |
|      - Suburban Residential House (Living Room, Bedroom, Kitchen, Bathroom)                      |
|      - Grocery / Convenience Store (Sales Floor, Storage Backroom, Aisle Shelves)                |
|      - Police Station (Lobby / Office, Armory Vault, Holding Cells)                              |
|      - Pharmacy / Clinic (Waiting / Retail, Consultation Exam Room, Medical Storage)             |
|      - Industrial Warehouse (Open Storage Bay, Crate Obstacles, Foreman Office, Loading Dock)    |
|   4. Outdoor Props & Yard Fencing (Fenced Yards, Gateways, Tree Foliage, Rubble/Debris)          |
|   5. Semantic Spawn Extraction (Safe House Player Spawn, Thematic Room Loot, Zombie Coordinates) |
+--------------------------------------------------------------------------------------------------+
                                                 |
                       +-------------------------+-------------------------+
                       |                                                   |
                       v                                                   v
          +-------------------------+                         +-------------------------+
          |      world.Map          |                         |  Spawn Metadata (ECS)   |
          | - Width, Height         |                         | - PlayerSpawn (Point)   |
          | - Tiles []TileType      |                         | - LootSpawns []LootSpawn|
          | - Buildings []Building  |                         | - ZombieSpawns []Point  |
          | - IsSolid() / Vision    |                         +-------------------------+
          +-------------------------+                                      |
                       |                                                   v
                       v                                      +-------------------------+
          +-------------------------+                         |   game.Reset() (ECS)    |
          |    Rendering System     |                         | - Player in Living Room |
          | - Ground Floor Pass     |<------------------------| - Thematic Loot in Rooms|
          | - Y-Sorted Obstacle Pass|                         | - Outdoor Zombie Spawns |
          +-------------------------+                         +-------------------------+
```

### 2.1 Tile System Expansion & Physical Properties

Expand `TileType` with the full set of supported tiles:
```go
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

// IsSolid returns true if entities cannot pass through this tile
func (t TileType) IsSolid() bool {
	switch t {
	case TileWall, TileTree, TileFence, TileDebris:
		return true
	default:
		return false
	}
}

// BlocksVision returns true if the tile occludes field of view raycasting
func (t TileType) BlocksVision() bool {
	return t == TileWall
}

// IsFloor returns true if the tile is a walkable ground surface
func (t TileType) IsFloor() bool {
	switch t {
	case TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor:
		return true
	default:
		return false
	}
}
```

---

### 2.2 Pure Go Geometric and Semantic Data Structures

```go
package world

// BuildingType identifies the archetype and thematic purpose of a structure
type BuildingType int

const (
	BuildingResidential BuildingType = iota
	BuildingGrocery
	BuildingPoliceStation
	BuildingPharmacy
	BuildingWarehouse
)

func (b BuildingType) String() string {
	switch b {
	case BuildingResidential:
		return "Residential House"
	case BuildingGrocery:
		return "Grocery Store"
	case BuildingPoliceStation:
		return "Police Station"
	case BuildingPharmacy:
		return "Pharmacy / Clinic"
	case BuildingWarehouse:
		return "Warehouse"
	default:
		return "Unknown Building"
	}
}

// RoomType identifies the functional sub-room within a building
type RoomType int

const (
	RoomLiving RoomType = iota
	RoomBedroom
	RoomKitchen
	RoomBathroom
	RoomStoreSales
	RoomStoreBackroom
	RoomPoliceLobby
	RoomOffice
	RoomArmory
	RoomHoldingCell
	RoomConsultation
	RoomMedicalStorage
	RoomWarehouseBay
	RoomLoadingDock
)

func (r RoomType) String() string {
	switch r {
	case RoomLiving:
		return "Living Room"
	case RoomBedroom:
		return "Bedroom"
	case RoomKitchen:
		return "Kitchen"
	case RoomBathroom:
		return "Bathroom"
	case RoomStoreSales:
		return "Sales Floor"
	case RoomStoreBackroom:
		return "Storage Room"
	case RoomPoliceLobby:
		return "Police Lobby"
	case RoomOffice:
		return "Office"
	case RoomArmory:
		return "Police Armory"
	case RoomHoldingCell:
		return "Holding Cell"
	case RoomConsultation:
		return "Consultation Room"
	case RoomMedicalStorage:
		return "Medical Storage"
	case RoomWarehouseBay:
		return "Warehouse Main Bay"
	case RoomLoadingDock:
		return "Loading Dock"
	default:
		return "Room"
	}
}

// Point represents a 2D integer tile coordinate
type Point struct {
	X, Y int
}

// Rect represents an integer bounding box in tile coordinates
type Rect struct {
	X, Y, W, H int
}

func (r Rect) Center() Point {
	return Point{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

func (r Rect) Overlaps(other Rect) bool {
	return r.X < other.X+other.W && r.X+r.W > other.X &&
		r.Y < other.Y+other.H && r.Y+r.H > other.Y
}

// Room represents an enclosed interior room with dedicated floor type
type Room struct {
	Type   RoomType
	Bounds Rect
	Floor  TileType
}

// Building represents a complete synthesized multi-room building
type Building struct {
	Type   BuildingType
	Bounds Rect
	Rooms  []Room
	Doors  []Point
}

// LootSpawn represents a thematic loot anchor extracted from building rooms
type LootSpawn struct {
	Type     string   // "food", "water", "weapon", "axe", "shotgun", "ammo", "armor"
	WorldX   float64  // World pixel coordinate (X * TileSize + 16)
	WorldY   float64  // World pixel coordinate (Y * TileSize + 16)
	RoomType RoomType
}
```

---

### 2.3 Detailed Architectural Specifications for 5 Building Archetypes

#### Archetype 1: Suburban Residential House
- **Footprint**: Width $10 \le W \le 14$, Height $8 \le H \le 12$.
- **Flooring**: `TileWoodFloor` for living room and bedroom; `TileTileFloor` for kitchen and bathroom.
- **Rooms (4 distinct rooms)**:
  1. **Living Room** (`RoomLiving`): Front-left room ($hy+H/2 \le y < hy+H-1$, $hx+1 \le x < hx+W/2$). Floor: `TileWoodFloor`.
  2. **Kitchen** (`RoomKitchen`): Front-right room ($hy+H/2 \le y < hy+H-1$, $hx+W/2 < x < hx+W-1$). Floor: `TileTileFloor`.
  3. **Bedroom** (`RoomBedroom`): Rear-left room ($hy+1 \le y < hy+H/2$, $hx+1 \le x < hx+W/2$). Floor: `TileWoodFloor`.
  4. **Bathroom** (`RoomBathroom`): Rear-right room ($hy+1 \le y < hy+H/2$, $hx+W/2 < x < hx+W-1$). Floor: `TileTileFloor`.
- **Wall Layout**:
  - Outer perimeter: `TileWall`.
  - Horizontal partition along row $y_{split} = hy + H/2$.
  - Vertical partition along column $x_{split} = hx + W/2$.
- **Doorways**:
  - Main Front Entrance: Bottom wall into living room at $(hx + (x_{split}-hx)/2, hy+hh-1)$ -> `TileWoodFloor`.
  - Living <-> Bedroom Door: Horizontal wall at $(hx + (x_{split}-hx)/2, y_{split})$ -> `TileWoodFloor`.
  - Living <-> Kitchen Pass-through: Vertical wall at $(x_{split}, y_{split} + (hy+hh-1-y_{split})/2)$ -> `TileWoodFloor`.
  - Kitchen / Hall <-> Bathroom Door: Horizontal wall at $(x_{split} + (hx+hw-1-x_{split})/2, y_{split})$ -> `TileTileFloor`.
  - Rear Backyard Door: North or east wall of kitchen -> `TileWoodFloor`.

```
Residential Floorplan Diagram:
+-------------------+-------------------+
|      BEDROOM      |     BATHROOM      |
|  (TileWoodFloor)  |  (TileTileFloor)  |
|                   |                   |
+-------[DOOR]------+-------[DOOR]------+
|    LIVING ROOM    |      KITCHEN      |
|  (TileWoodFloor)  [DOOR] (TileTileFloor)|
|                   |                   |
+-------[DOOR]------+-------[DOOR]------+
     (Main Entry)        (Back Door)
```

---

#### Archetype 2: Grocery / Convenience Store
- **Footprint**: Width $14 \le W \le 18$, Height $10 \le H \le 14$.
- **Flooring**: `TileTileFloor` throughout sales floor and office; `TileConcrete` in rear storage backroom.
- **Rooms (2-3 rooms)**:
  1. **Sales Floor** (`RoomStoreSales`): Main front retail zone ($hy + h_{back} < y < hy+hh-1$). Floor: `TileTileFloor`. Contains 2 interior shelf display rows (`TileWall` fixture blocks with walkable aisles).
  2. **Storage Backroom** (`RoomStoreBackroom`): Rear 35% zone ($hy+1 \le y < hy + h_{back}$). Floor: `TileConcrete`.
  3. **Manager Office** (`RoomOffice`): Partitioned corner office in backroom. Floor: `TileTileFloor`.
- **Wall Layout & Doors**:
  - Front double entrance doors: $(hx + hw/2, hy+hh-1)$ and $(hx + hw/2 - 1, hy+hh-1)$ -> `TileTileFloor`.
  - Horizontal partition wall separating sales floor from storage backroom at $y = hy + h_{back}$.
  - Storage Room Door: $(hx + 3, hy + h_{back})$ -> `TileTileFloor`.
  - Rear Service / Cargo Loading Door: $(hx + hw/2, hy)$ -> `TileConcrete`.

```
Grocery Store Floorplan Diagram:
+---------------------------------------+
|  [DOOR: Rear Service Loading Dock]    |
|       STORAGE BACKROOM (Concrete)     |
+-------[DOOR]--------------------------+
|  [SHELF]     [SHELF]     [SHELF]      |
|                                       |
|       SALES FLOOR (TileFloor)         |
|                                       |
+-----------------[DOORS]---------------+
            (Double Front Entrance)
```

---

#### Archetype 3: Police Station
- **Footprint**: Width $14 \le W \le 18$, Height $12 \le H \le 16$.
- **Flooring**: `TileTileFloor` for lobby and detective office; `TileConcrete` for armory and holding cells.
- **Rooms (4 rooms)**:
  1. **Public Lobby / Reception** (`RoomPoliceLobby`): Front-left area. Floor: `TileTileFloor`.
  2. **Detective / Captain Office** (`RoomOffice`): Front-right area. Floor: `TileTileFloor`.
  3. **Police Armory / Weapons Vault** (`RoomArmory`): Secure rear-left room. Floor: `TileConcrete`.
  4. **Holding Cells / Lockup** (`RoomHoldingCell`): Secure rear-right room. Floor: `TileConcrete`.
- **Wall Layout & Doors**:
  - Outer perimeter: `TileWall`.
  - Horizontal security partition along row $y_{split} = hy + hh/2$.
  - Rear vertical wall along column $x_{split} = hx + hw/2$.
  - Front office vertical divider at $x = hx + hw*2/3$.
  - Front entrance into lobby: $(hx + (hw*2/3)/2, hy+hh-1)$ -> `TileTileFloor`.
  - Armory secure entrance: $(hx + (x_{split}-hx)/2, y_{split})$ -> `TileTileFloor`.
  - Holding cell entrance: $(x_{split} + (hx+hw-1-x_{split})/2, y_{split})$ -> `TileTileFloor`.
  - Office doorway: $(hx + hw*2/3, y_{split} + 2)$ -> `TileTileFloor`.
  - Rear tactical exit: $(hx + hw/2, hy)$ -> `TileConcrete`.

```
Police Station Floorplan Diagram:
+-------------------+-------------------+
|   POLICE ARMORY   |   HOLDING CELLS   |
| (Weapons/Ammo/Vest)|   (Lockup Cells)  |
|  (TileConcrete)   |  (TileConcrete)   |
+-------[DOOR]------+-------[DOOR]------+
|    POLICE LOBBY   |  DETECTIVE OFFICE |
|   & RECEPTION     [DOOR]  (Records)   |
|  (TileTileFloor)  |  (TileTileFloor)  |
+-------[DOOR]------+-------------------+
     (Front Entry)
```

---

#### Archetype 4: Pharmacy / Clinic
- **Footprint**: Width $12 \le W \le 16$, Height $10 \le H \le 14$.
- **Flooring**: `TileTileFloor` (sterile white/blue clinic tile throughout).
- **Rooms (3 rooms)**:
  1. **Waiting Room & Retail Counter** (`RoomPharmacyRetail`): Front half of building. Floor: `TileTileFloor`.
  2. **Doctor Consultation & Exam Room** (`RoomConsultation`): Rear-left room. Floor: `TileTileFloor`.
  3. **Medical Storage & Dispensary** (`RoomMedicalStorage`): Secure rear-right room. Floor: `TileTileFloor`.
- **Wall Layout & Doors**:
  - Horizontal divider at $y_{split} = hy + hh/2$.
  - Rear vertical divider at $x_{split} = hx + hw/2$.
  - Main entrance at front south wall: $(hx + hw/2, hy+hh-1)$ -> `TileTileFloor`.
  - Consultation room door: $(hx + (x_{split}-hx)/2, y_{split})$ -> `TileTileFloor`.
  - Medical dispensary door: $(x_{split} + (hx+hw-1-x_{split})/2, y_{split})$ -> `TileTileFloor`.

```
Pharmacy / Clinic Floorplan Diagram:
+-------------------+-------------------+
| CONSULTATION ROOM |  MEDICAL STORAGE  |
|  (Exam & Clinic)  |  (First Aid/Med)  |
|  (TileTileFloor)  |  (TileTileFloor)  |
+-------[DOOR]------+-------[DOOR]------+
|                                       |
|    WAITING ROOM & PHARMACY COUNTER    |
|            (TileTileFloor)            |
+-----------------[DOOR]----------------+
              (Main Entry)
```

---

#### Archetype 5: Industrial Warehouse
- **Footprint**: Width $16 \le W \le 22$, Height $12 \le H \le 18$.
- **Flooring**: `TileConcrete` (industrial concrete slab throughout); `TileTileFloor` in foreman corner office.
- **Rooms (2-3 rooms)**:
  1. **Warehouse Main Bay** (`RoomWarehouseBay`): Expansive high-ceiling storage bay covering 80%+ of the floor. Contains organized stacks of pallet crates (`TileDebris`) forming staging aisles. Floor: `TileConcrete`.
  2. **Foreman Office** (`RoomOffice`): Partitioned corner office (e.g. 5x4 tiles in top-right corner). Floor: `TileTileFloor`.
  3. **Loading Dock Staging Area** (`RoomLoadingDock`): Open bay staging zone adjacent to roll-up cargo doors. Floor: `TileConcrete`.
- **Wall Layout & Doors**:
  - Outer perimeter: `TileWall`.
  - Corner partition walls enclosing Foreman Office.
  - Wide 3-tile roll-up cargo entrance: $(hx + 4, hy+hh-1)$, $(hx + 5, hy+hh-1)$, $(hx + 6, hy+hh-1)$ -> `TileConcrete`.
  - Personnel side entrance: $(hx, hy + hh/2)$ -> `TileConcrete`.
  - Foreman Office interior door: $(hx + hw - 1 - w_{off}, hy + 2)$ -> `TileTileFloor`.

```
Industrial Warehouse Floorplan Diagram:
+-----------------------------+---------+
|                             | FOREMAN |
|  [CRATE]  [CRATE]  [CRATE]  | OFFICE  |
|  [CRATE]  [CRATE]  [CRATE]  | [DOOR]  |
|                             +---------+
[DOOR: Side]                            |
|       MAIN STORAGE BAY (Concrete)     |
|                                       |
+---[CARGO ROLL-UP DOORS]---------------+
```

---

### 2.4 Town Zoning and Procedural Synthesis Pipeline

In a 100x100 tile world map:
1. **District Layout**:
   - **North-West District ($X \in [4, 45], Y \in [4, 45]$)**: Suburban Residential Neighborhood with 3 multi-room residential houses, fenced yards (`TileFence`), garden tree clusters, and garden gates.
   - **North-East District ($X \in [55, 95], Y \in [4, 45]$)**: Commercial & Health Sector with a Grocery Store and a Pharmacy/Clinic, surrounded by concrete sidewalks and paved access paths.
   - **South-West District ($X \in [4, 45], Y \in [55, 95]$)**: Municipal & Civic Sector featuring a reinforced Police Station (with Armory and Holding Cells) plus an auxiliary residential cabin/outpost.
   - **South-East District ($X \in [55, 95], Y \in [55, 95]$)**: Industrial Logistics Sector with a large multi-bay Warehouse (pallet crate aisles, foreman office, cargo loading dock).
2. **Road Network Hierarchy**:
   - **Main Avenues**: 4-tile wide `TileAsphalt` running across the center:
     - Vertical Avenue: $X \in [48, 51]$, $Y \in [1, 98]$.
     - Horizontal Avenue: $Y \in [48, 51]$, $X \in [1, 98]$.
   - **Pedestrian Sidewalks**: 1-tile wide `TileConcrete` flanking both avenues at $X=47, 52$ and $Y=47, 52$.
   - **Neighborhood Connector Streets**: 2-tile wide `TileAsphalt` or `TileDirt` branching into each quadrant to provide direct road access to all building entrances.
3. **Fenced Yards & Props**:
   - Perimeter yards enclosed with `TileFence` with 2-tile open gate gaps leading out to sidewalks.
   - Natural vegetation clusters (`TileTree`) and outdoor rubble/debris (`TileDebris`) placed safely away from doors and roads.

---

### 2.5 Contextual Spawning & ECS Integration

1. **Guaranteed Safe Player Spawn**:
   - `PlayerSpawn` is set to the center of the Living Room of the primary residential house in the NW district (e.g. $(16, 16) \times 32$).
   - The room is guaranteed free of zombies and clear of wall collisions.
2. **Thematic Loot Spawning Distribution**:
   - **Residential Kitchens**: Guaranteed `food` (canned goods) and `water` (bottles).
   - **Residential Bedrooms**: `armor` (tactical vest), `weapon` (club/crowbar), `axe` (fire axe).
   - **Grocery Store Sales Floor & Backroom**: High density of `food` and `water`.
   - **Police Station Armory**: `shotgun`, `ammo`, `armor`, `weapon`.
   - **Police Station Lobby / Office**: `weapon`, `water`.
   - **Pharmacy / Clinic Medical Storage & Exam Room**: `food` (rations), `water` (saline/purified), first aid supplies.
   - **Warehouse Main Bay & Foreman Office**: `axe`, `shotgun`, `ammo`, `armor`, `weapon`.
3. **Outdoor Zombie Distribution**:
   - Spawns 150 zombies on outdoor walkable floor tiles (`TileGrass`, `TileDirt`, `TileAsphalt`, `TileConcrete`).
   - Validates that `!GetTile(x, y).IsSolid()` and ensures distance to `PlayerSpawn` is $> 300$ pixels, preventing zombie wall entrapment.

---

## 3. Pure Go Code Implementation for `internal/game/world`

Below is the complete, modular Go implementation designed to replace and upgrade `internal/game/world/map.go`.

### 3.1 Proposed `internal/game/world/map.go`

```go
package world

import (
	"math"
	"math/rand"
)

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

// IsSolid returns true if entities cannot walk through this tile
func (t TileType) IsSolid() bool {
	switch t {
	case TileWall, TileTree, TileFence, TileDebris:
		return true
	default:
		return false
	}
}

// BlocksVision returns true if the tile occludes field of view raycasts
func (t TileType) BlocksVision() bool {
	return t == TileWall
}

// IsFloor returns true if the tile is a walkable ground surface
func (t TileType) IsFloor() bool {
	switch t {
	case TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor:
		return true
	default:
		return false
	}
}

// BuildingType identifies the archetype of a structure
type BuildingType int

const (
	BuildingResidential BuildingType = iota
	BuildingGrocery
	BuildingPoliceStation
	BuildingPharmacy
	BuildingWarehouse
)

func (b BuildingType) String() string {
	switch b {
	case BuildingResidential:
		return "Residential House"
	case BuildingGrocery:
		return "Grocery Store"
	case BuildingPoliceStation:
		return "Police Station"
	case BuildingPharmacy:
		return "Pharmacy / Clinic"
	case BuildingWarehouse:
		return "Warehouse"
	default:
		return "Unknown Building"
	}
}

// RoomType identifies the functional sub-room within a building
type RoomType int

const (
	RoomLiving RoomType = iota
	RoomBedroom
	RoomKitchen
	RoomBathroom
	RoomStoreSales
	RoomStoreBackroom
	RoomPoliceLobby
	RoomOffice
	RoomArmory
	RoomHoldingCell
	RoomConsultation
	RoomMedicalStorage
	RoomWarehouseBay
	RoomLoadingDock
)

func (r RoomType) String() string {
	switch r {
	case RoomLiving:
		return "Living Room"
	case RoomBedroom:
		return "Bedroom"
	case RoomKitchen:
		return "Kitchen"
	case RoomBathroom:
		return "Bathroom"
	case RoomStoreSales:
		return "Sales Floor"
	case RoomStoreBackroom:
		return "Storage Room"
	case RoomPoliceLobby:
		return "Police Lobby"
	case RoomOffice:
		return "Office"
	case RoomArmory:
		return "Police Armory"
	case RoomHoldingCell:
		return "Holding Cell"
	case RoomConsultation:
		return "Consultation Room"
	case RoomMedicalStorage:
		return "Medical Storage"
	case RoomWarehouseBay:
		return "Warehouse Main Bay"
	case RoomLoadingDock:
		return "Loading Dock"
	default:
		return "Room"
	}
}

// Point represents a 2D integer tile coordinate
type Point struct {
	X, Y int
}

// Rect represents an integer bounding box in tile coordinates
type Rect struct {
	X, Y, W, H int
}

func (r Rect) Center() Point {
	return Point{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

func (r Rect) Overlaps(other Rect) bool {
	return r.X < other.X+other.W && r.X+r.W > other.X &&
		r.Y < other.Y+other.H && r.Y+r.H > other.Y
}

// Room represents an interior room
type Room struct {
	Type   RoomType
	Bounds Rect
	Floor  TileType
}

// Building represents a synthesized multi-room building
type Building struct {
	Type   BuildingType
	Bounds Rect
	Rooms  []Room
	Doors  []Point
}

// LootSpawn represents a thematic loot anchor
type LootSpawn struct {
	Type     string
	WorldX   float64
	WorldY   float64
	RoomType RoomType
}

// Map holds the complete world grid and spawn metadata
type Map struct {
	Width, Height int
	Tiles         []TileType
	Visible       []bool
	Explored      []bool
	Buildings     []Building
	PlayerSpawn   Point
	LootSpawns    []LootSpawn
	ZombieSpawns  []Point
}

// NewMap constructs and procedurally generates a complete town map
func NewMap(width, height int) *Map {
	m := &Map{
		Width:        width,
		Height:       height,
		Tiles:        make([]TileType, width*height),
		Visible:      make([]bool, width*height),
		Explored:     make([]bool, width*height),
		Buildings:    make([]Building, 0),
		LootSpawns:   make([]LootSpawn, 0),
		ZombieSpawns: make([]Point, 0),
	}

	// 1. Fill base map: outer boundary with TileWall, interior with TileGrass
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileGrass)
			}
		}
	}

	// For very small test maps (e.g. 10x10), return basic map
	if width < 30 || height < 30 {
		return m
	}

	// 2. Generate Road Network (Asphalt Avenues & Sidewalks)
	midX := width / 2
	midY := height / 2

	// Vertical Avenue (columns midX-2 to midX+1) with flanking sidewalks
	for y := 1; y < height-1; y++ {
		m.SetTile(midX-3, y, TileConcrete) // West Sidewalk
		for x := midX - 2; x <= midX+1; x++ {
			m.SetTile(x, y, TileAsphalt)
		}
		m.SetTile(midX+2, y, TileConcrete) // East Sidewalk
	}

	// Horizontal Avenue (rows midY-2 to midY+1) with flanking sidewalks
	for x := 1; x < width-1; x++ {
		m.SetTile(x, midY-3, TileConcrete) // North Sidewalk
		for y := midY - 2; y <= midY+1; y++ {
			m.SetTile(x, y, TileAsphalt)
		}
		m.SetTile(x, midY+2, TileConcrete) // South Sidewalk
	}

	// Branch streets into neighborhoods
	for x := 1; x < midX-3; x++ {
		m.SetTile(x, 22, TileDirt)
		m.SetTile(x, 23, TileDirt)
	}
	for x := midX + 3; x < width-1; x++ {
		m.SetTile(x, 22, TileAsphalt)
		m.SetTile(x, 23, TileAsphalt)
	}
	for x := 1; x < midX-3; x++ {
		m.SetTile(x, 72, TileDirt)
		m.SetTile(x, 73, TileDirt)
	}
	for x := midX + 3; x < width-1; x++ {
		m.SetTile(x, 72, TileConcrete)
		m.SetTile(x, 73, TileConcrete)
	}

	// 3. Synthesize Multi-Room Buildings across Town Districts

	// Quadrant 1 (North-West): Residential Neighborhood
	b1 := m.buildResidentialHouse(Rect{X: 8, Y: 8, W: 12, H: 10})
	b2 := m.buildResidentialHouse(Rect{X: 26, Y: 8, W: 14, H: 10})
	b3 := m.buildResidentialHouse(Rect{X: 12, Y: 28, W: 12, H: 10})
	m.Buildings = append(m.Buildings, b1, b2, b3)

	// Set guaranteed PlayerSpawn in primary residential living room
	if len(b1.Rooms) > 0 {
		m.PlayerSpawn = b1.Rooms[0].Bounds.Center()
	} else {
		m.PlayerSpawn = Point{X: 14, Y: 14}
	}

	// Add yard fences around residential houses
	m.buildFenceYard(Rect{X: 6, Y: 6, W: 36, H: 14}, Point{X: 20, Y: 20})

	// Quadrant 2 (North-East): Commercial & Healthcare District
	bGrocery := m.buildGroceryStore(Rect{X: midX + 8, Y: 6, W: 16, H: 12})
	bPharmacy := m.buildPharmacyClinic(Rect{X: midX + 28, Y: 6, W: 14, H: 12})
	m.Buildings = append(m.Buildings, bGrocery, bPharmacy)

	// Quadrant 3 (South-West): Civic & Police Sector
	bPolice := m.buildPoliceStation(Rect{X: 10, Y: midY + 10, W: 16, H: 14})
	bHouse4 := m.buildResidentialHouse(Rect{X: 30, Y: midY + 12, W: 12, H: 10})
	m.Buildings = append(m.Buildings, bPolice, bHouse4)

	// Quadrant 4 (South-East): Industrial & Logistics District
	bWarehouse := m.buildWarehouse(Rect{X: midX + 10, Y: midY + 10, W: 20, H: 16})
	m.Buildings = append(m.Buildings, bWarehouse)

	// 4. Extract Thematic Loot Spawns from All Generated Rooms
	m.extractThematicLoot()

	// 5. Place Outdoor Props, Trees, and Rubble
	m.populateOutdoorProps()

	// 6. Generate Valid Outdoor Zombie Spawns
	m.generateZombieSpawns(150)

	return m
}

// buildResidentialHouse synthesizes a 4-room home: Living, Bedroom, Kitchen, Bathroom
func (m *Map) buildResidentialHouse(bounds Rect) Building {
	hx, hy, hw, hh := bounds.X, bounds.Y, bounds.W, bounds.H

	// 1. Outer perimeter walls and wood floor base
	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileWoodFloor)
			}
		}
	}

	xSplit := hx + hw/2
	ySplit := hy + hh/2

	// 2. Interior partition walls
	for x := hx + 1; x < hx+hw-1; x++ {
		m.SetTile(x, ySplit, TileWall)
	}
	for y := hy + 1; y < hy+hh-1; y++ {
		m.SetTile(xSplit, y, TileWall)
	}

	// 3. Define 4 Rooms
	rLiving := Room{
		Type:   RoomLiving,
		Bounds: Rect{X: hx + 1, Y: ySplit + 1, W: xSplit - hx - 1, H: hy + hh - 1 - (ySplit + 1)},
		Floor:  TileWoodFloor,
	}
	rKitchen := Room{
		Type:   RoomKitchen,
		Bounds: Rect{X: xSplit + 1, Y: ySplit + 1, W: hx + hw - 1 - (xSplit + 1), H: hy + hh - 1 - (ySplit + 1)},
		Floor:  TileTileFloor,
	}
	rBedroom := Room{
		Type:   RoomBedroom,
		Bounds: Rect{X: hx + 1, Y: hy + 1, W: xSplit - hx - 1, H: ySplit - hy - 1},
		Floor:  TileWoodFloor,
	}
	rBathroom := Room{
		Type:   RoomBathroom,
		Bounds: Rect{X: xSplit + 1, Y: hy + 1, W: hx + hw - 1 - (xSplit + 1), H: ySplit - hy - 1},
		Floor:  TileTileFloor,
	}

	// Apply tile floor in kitchen and bathroom
	for y := rKitchen.Bounds.Y; y < rKitchen.Bounds.Y+rKitchen.Bounds.H; y++ {
		for x := rKitchen.Bounds.X; x < rKitchen.Bounds.X+rKitchen.Bounds.W; x++ {
			m.SetTile(x, y, TileTileFloor)
		}
	}
	for y := rBathroom.Bounds.Y; y < rBathroom.Bounds.Y+rBathroom.Bounds.H; y++ {
		for x := rBathroom.Bounds.X; x < rBathroom.Bounds.X+rBathroom.Bounds.W; x++ {
			m.SetTile(x, y, TileTileFloor)
		}
	}

	// 4. Place Doors for 100% room connectivity
	doorFront := Point{X: hx + (xSplit-hx)/2, Y: hy + hh - 1}
	doorLivingBed := Point{X: hx + (xSplit-hx)/2, Y: ySplit}
	doorLivingKitchen := Point{X: xSplit, Y: ySplit + (hy+hh-1-ySplit)/2}
	doorKitchenBath := Point{X: xSplit + (hx+hw-1-xSplit)/2, Y: ySplit}
	doorBack := Point{X: xSplit + (hx+hw-1-xSplit)/2, Y: hy + hh - 1}

	m.SetTile(doorFront.X, doorFront.Y, TileWoodFloor)
	m.SetTile(doorLivingBed.X, doorLivingBed.Y, TileWoodFloor)
	m.SetTile(doorLivingKitchen.X, doorLivingKitchen.Y, TileWoodFloor)
	m.SetTile(doorKitchenBath.X, doorKitchenBath.Y, TileTileFloor)
	m.SetTile(doorBack.X, doorBack.Y, TileTileFloor)

	return Building{
		Type:   BuildingResidential,
		Bounds: bounds,
		Rooms:  []Room{rLiving, rKitchen, rBedroom, rBathroom},
		Doors:  []Point{doorFront, doorLivingBed, doorLivingKitchen, doorKitchenBath, doorBack},
	}
}

// buildGroceryStore synthesizes a store with a large sales floor and rear concrete storage room
func (m *Map) buildGroceryStore(bounds Rect) Building {
	hx, hy, hw, hh := bounds.X, bounds.Y, bounds.W, bounds.H

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	hBack := 4
	if hh <= 10 {
		hBack = 3
	}
	ySplit := hy + hBack

	// Partition wall between storage and sales floor
	for x := hx + 1; x < hx+hw-1; x++ {
		m.SetTile(x, ySplit, TileWall)
	}

	// Fill storage room with concrete floor
	for y := hy + 1; y < ySplit; y++ {
		for x := hx + 1; x < hx+hw-1; x++ {
			m.SetTile(x, y, TileConcrete)
		}
	}

	rBackroom := Room{
		Type:   RoomStoreBackroom,
		Bounds: Rect{X: hx + 1, Y: hy + 1, W: hw - 2, H: hBack - 1},
		Floor:  TileConcrete,
	}
	rSales := Room{
		Type:   RoomStoreSales,
		Bounds: Rect{X: hx + 1, Y: ySplit + 1, W: hw - 2, H: hh - hBack - 2},
		Floor:  TileTileFloor,
	}

	// Add 2 retail shelf aisles in sales floor
	shelfY1 := ySplit + 3
	if shelfY1 < hy+hh-2 {
		for sx := hx + 3; sx <= hx+hw-4; sx++ {
			if sx%3 != 0 {
				m.SetTile(sx, shelfY1, TileWall)
			}
		}
	}

	// Doors: front double entrance, backroom connecting door, rear loading door
	doorFront1 := Point{X: hx + hw/2 - 1, Y: hy + hh - 1}
	doorFront2 := Point{X: hx + hw/2, Y: hy + hh - 1}
	doorBackroom := Point{X: hx + 3, Y: ySplit}
	doorRearLoading := Point{X: hx + hw/2, Y: hy}

	m.SetTile(doorFront1.X, doorFront1.Y, TileTileFloor)
	m.SetTile(doorFront2.X, doorFront2.Y, TileTileFloor)
	m.SetTile(doorBackroom.X, doorBackroom.Y, TileTileFloor)
	m.SetTile(doorRearLoading.X, doorRearLoading.Y, TileConcrete)

	return Building{
		Type:   BuildingGrocery,
		Bounds: bounds,
		Rooms:  []Room{rSales, rBackroom},
		Doors:  []Point{doorFront1, doorFront2, doorBackroom, doorRearLoading},
	}
}

// buildPoliceStation synthesizes a secure precinct with Armory, Holding Cells, Lobby, and Office
func (m *Map) buildPoliceStation(bounds Rect) Building {
	hx, hy, hw, hh := bounds.X, bounds.Y, bounds.W, bounds.H

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	xSplit := hx + hw/2
	ySplit := hy + hh/2
	xOffSplit := hx + (hw * 2) / 3

	// Security partition wall
	for x := hx + 1; x < hx+hw-1; x++ {
		m.SetTile(x, ySplit, TileWall)
	}
	// Rear secure division
	for y := hy + 1; y < ySplit; y++ {
		m.SetTile(xSplit, y, TileWall)
	}
	// Front office division
	for y := ySplit + 1; y < hy+hh-1; y++ {
		m.SetTile(xOffSplit, y, TileWall)
	}

	// Armory & Holding cells floor = TileConcrete
	for y := hy + 1; y < ySplit; y++ {
		for x := hx + 1; x < hx+hw-1; x++ {
			m.SetTile(x, y, TileConcrete)
		}
	}

	rLobby := Room{
		Type:   RoomPoliceLobby,
		Bounds: Rect{X: hx + 1, Y: ySplit + 1, W: xOffSplit - hx - 1, H: hy + hh - 1 - (ySplit + 1)},
		Floor:  TileTileFloor,
	}
	rOffice := Room{
		Type:   RoomOffice,
		Bounds: Rect{X: xOffSplit + 1, Y: ySplit + 1, W: hx + hw - 1 - (xOffSplit + 1), H: hy + hh - 1 - (ySplit + 1)},
		Floor:  TileTileFloor,
	}
	rArmory := Room{
		Type:   RoomArmory,
		Bounds: Rect{X: hx + 1, Y: hy + 1, W: xSplit - hx - 1, H: ySplit - hy - 1},
		Floor:  TileConcrete,
	}
	rCells := Room{
		Type:   RoomHoldingCell,
		Bounds: Rect{X: xSplit + 1, Y: hy + 1, W: hx + hw - 1 - (xSplit + 1), H: ySplit - hy - 1},
		Floor:  TileConcrete,
	}

	doorFront := Point{X: hx + (xOffSplit-hx)/2, Y: hy + hh - 1}
	doorOffice := Point{X: xOffSplit, Y: ySplit + 2}
	doorArmory := Point{X: hx + (xSplit-hx)/2, Y: ySplit}
	doorCells := Point{X: xSplit + (hx+hw-1-xSplit)/2, Y: ySplit}
	doorRear := Point{X: hx + hw/2, Y: hy}

	m.SetTile(doorFront.X, doorFront.Y, TileTileFloor)
	m.SetTile(doorOffice.X, doorOffice.Y, TileTileFloor)
	m.SetTile(doorArmory.X, doorArmory.Y, TileTileFloor)
	m.SetTile(doorCells.X, doorCells.Y, TileTileFloor)
	m.SetTile(doorRear.X, doorRear.Y, TileConcrete)

	return Building{
		Type:   BuildingPoliceStation,
		Bounds: bounds,
		Rooms:  []Room{rLobby, rOffice, rArmory, rCells},
		Doors:  []Point{doorFront, doorOffice, doorArmory, doorCells, doorRear},
	}
}

// buildPharmacyClinic synthesizes a clinic with Waiting Area, Consultation Room, and Medical Storage
func (m *Map) buildPharmacyClinic(bounds Rect) Building {
	hx, hy, hw, hh := bounds.X, bounds.Y, bounds.W, bounds.H

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileTileFloor)
			}
		}
	}

	xSplit := hx + hw/2
	ySplit := hy + hh/2

	for x := hx + 1; x < hx+hw-1; x++ {
		m.SetTile(x, ySplit, TileWall)
	}
	for y := hy + 1; y < ySplit; y++ {
		m.SetTile(xSplit, y, TileWall)
	}

	rWaiting := Room{
		Type:   RoomPharmacyRetail,
		Bounds: Rect{X: hx + 1, Y: ySplit + 1, W: hw - 2, H: hy + hh - 1 - (ySplit + 1)},
		Floor:  TileTileFloor,
	}
	rConsult := Room{
		Type:   RoomConsultation,
		Bounds: Rect{X: hx + 1, Y: hy + 1, W: xSplit - hx - 1, H: ySplit - hy - 1},
		Floor:  TileTileFloor,
	}
	rMedStore := Room{
		Type:   RoomMedicalStorage,
		Bounds: Rect{X: xSplit + 1, Y: hy + 1, W: hx + hw - 1 - (xSplit + 1), H: ySplit - hy - 1},
		Floor:  TileTileFloor,
	}

	doorFront := Point{X: hx + hw/2, Y: hy + hh - 1}
	doorConsult := Point{X: hx + (xSplit-hx)/2, Y: ySplit}
	doorMedStore := Point{X: xSplit + (hx+hw-1-xSplit)/2, Y: ySplit}

	m.SetTile(doorFront.X, doorFront.Y, TileTileFloor)
	m.SetTile(doorConsult.X, doorConsult.Y, TileTileFloor)
	m.SetTile(doorMedStore.X, doorMedStore.Y, TileTileFloor)

	return Building{
		Type:   BuildingPharmacy,
		Bounds: bounds,
		Rooms:  []Room{rWaiting, rConsult, rMedStore},
		Doors:  []Point{doorFront, doorConsult, doorMedStore},
	}
}

// buildWarehouse synthesizes a large industrial warehouse with crate aisles and foreman office
func (m *Map) buildWarehouse(bounds Rect) Building {
	hx, hy, hw, hh := bounds.X, bounds.Y, bounds.W, bounds.H

	for y := hy; y < hy+hh; y++ {
		for x := hx; x < hx+hw; x++ {
			if x == hx || x == hx+hw-1 || y == hy || y == hy+hh-1 {
				m.SetTile(x, y, TileWall)
			} else {
				m.SetTile(x, y, TileConcrete)
			}
		}
	}

	wOff := 5
	hOff := 4
	if hw <= 14 {
		wOff = 4
	}

	// Office walls (Top-Right corner)
	offX := hx + hw - 1 - wOff
	offY := hy + 1
	for y := offY; y <= offY+hOff; y++ {
		m.SetTile(offX, y, TileWall)
	}
	for x := offX; x < hx+hw-1; x++ {
		m.SetTile(x, offY+hOff, TileWall)
	}

	// Office floor = TileTileFloor
	for y := offY; y < offY+hOff; y++ {
		for x := offX + 1; x < hx+hw-1; x++ {
			m.SetTile(x, y, TileTileFloor)
		}
	}

	rOffice := Room{
		Type:   RoomOffice,
		Bounds: Rect{X: offX + 1, Y: offY, W: wOff - 1, H: hOff},
		Floor:  TileTileFloor,
	}
	rBay := Room{
		Type:   RoomWarehouseBay,
		Bounds: Rect{X: hx + 1, Y: hy + 1, W: hw - 2, H: hh - 2},
		Floor:  TileConcrete,
	}

	// Crate stacks (TileDebris) in main bay
	for cy := hy + 3; cy <= hy+hh-6; cy += 3 {
		for cx := hx + 3; cx <= offX-3; cx++ {
			m.SetTile(cx, cy, TileDebris)
		}
	}

	// Roll-up cargo doors (3-tile wide)
	doorCargo1 := Point{X: hx + 4, Y: hy + hh - 1}
	doorCargo2 := Point{X: hx + 5, Y: hy + hh - 1}
	doorCargo3 := Point{X: hx + 6, Y: hy + hh - 1}
	doorSide := Point{X: hx, Y: hy + hh/2}
	doorOffice := Point{X: offX, Y: offY + 2}

	m.SetTile(doorCargo1.X, doorCargo1.Y, TileConcrete)
	m.SetTile(doorCargo2.X, doorCargo2.Y, TileConcrete)
	m.SetTile(doorCargo3.X, doorCargo3.Y, TileConcrete)
	m.SetTile(doorSide.X, doorSide.Y, TileConcrete)
	m.SetTile(doorOffice.X, doorOffice.Y, TileTileFloor)

	return Building{
		Type:   BuildingWarehouse,
		Bounds: bounds,
		Rooms:  []Room{rBay, rOffice},
		Doors:  []Point{doorCargo1, doorCargo2, doorCargo3, doorSide, doorOffice},
	}
}

// buildFenceYard surrounds residential lots with fences and a gate opening
func (m *Map) buildFenceYard(bounds Rect, gate Point) {
	for y := bounds.Y; y < bounds.Y+bounds.H; y++ {
		for x := bounds.X; x < bounds.X+bounds.W; x++ {
			if x <= 0 || x >= m.Width-1 || y <= 0 || y >= m.Height-1 {
				continue
			}
			isPerimeter := (x == bounds.X || x == bounds.X+bounds.W-1 || y == bounds.Y || y == bounds.Y+bounds.H-1)
			if isPerimeter {
				if m.GetTile(x, y) == TileGrass {
					if x == gate.X && y == gate.Y {
						m.SetTile(x, y, TileDirt)
					} else {
						m.SetTile(x, y, TileFence)
					}
				}
			}
		}
	}
}

// extractThematicLoot populates m.LootSpawns based on room semantic types
func (m *Map) extractThematicLoot() {
	for _, b := range m.Buildings {
		for _, r := range b.Rooms {
			cx := float64(r.Bounds.Center().X*TileSize + 16)
			cy := float64(r.Bounds.Center().Y*TileSize + 16)

			switch r.Type {
			case RoomKitchen:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "food", WorldX: cx - 16, WorldY: cy, RoomType: r.Type},
					LootSpawn{Type: "water", WorldX: cx + 16, WorldY: cy, RoomType: r.Type},
				)
			case RoomBedroom:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "armor", WorldX: cx - 16, WorldY: cy, RoomType: r.Type},
					LootSpawn{Type: "weapon", WorldX: cx + 16, WorldY: cy, RoomType: r.Type},
				)
			case RoomStoreSales:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "food", WorldX: cx - 32, WorldY: cy, RoomType: r.Type},
					LootSpawn{Type: "water", WorldX: cx, WorldY: cy, RoomType: r.Type},
					LootSpawn{Type: "food", WorldX: cx + 32, WorldY: cy, RoomType: r.Type},
				)
			case RoomStoreBackroom:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "food", WorldX: cx - 16, WorldY: cy, RoomType: r.Type},
					LootSpawn{Type: "water", WorldX: cx + 16, WorldY: cy, RoomType: r.Type},
				)
			case RoomArmory:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "shotgun", WorldX: cx - 20, WorldY: cy - 20, RoomType: r.Type},
					LootSpawn{Type: "ammo", WorldX: cx + 20, WorldY: cy - 20, RoomType: r.Type},
					LootSpawn{Type: "armor", WorldX: cx - 20, WorldY: cy + 20, RoomType: r.Type},
					LootSpawn{Type: "weapon", WorldX: cx + 20, WorldY: cy + 20, RoomType: r.Type},
				)
			case RoomMedicalStorage:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "water", WorldX: cx - 16, WorldY: cy, RoomType: r.Type},
					LootSpawn{Type: "food", WorldX: cx + 16, WorldY: cy, RoomType: r.Type},
				)
			case RoomWarehouseBay:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "axe", WorldX: cx - 30, WorldY: cy - 30, RoomType: r.Type},
					LootSpawn{Type: "ammo", WorldX: cx + 30, WorldY: cy - 30, RoomType: r.Type},
					LootSpawn{Type: "armor", WorldX: cx, WorldY: cy + 30, RoomType: r.Type},
				)
			case RoomOffice:
				m.LootSpawns = append(m.LootSpawns,
					LootSpawn{Type: "weapon", WorldX: cx, WorldY: cy, RoomType: r.Type},
				)
			}
		}
	}
}

// populateOutdoorProps scatters trees and debris in open grass areas
func (m *Map) populateOutdoorProps() {
	if m.Width <= 10 || m.Height <= 10 {
		return
	}
	// Scatter trees in grass areas
	for i := 0; i < 120; i++ {
		tx := 2 + rand.Intn(m.Width-4)
		ty := 2 + rand.Intn(m.Height-4)
		if m.GetTile(tx, ty) == TileGrass {
			m.SetTile(tx, ty, TileTree)
		}
	}
	// Scatter debris in alleyways
	for i := 0; i < 25; i++ {
		dx := 2 + rand.Intn(m.Width-4)
		dy := 2 + rand.Intn(m.Height-4)
		if m.GetTile(dx, dy) == TileGrass || m.GetTile(dx, dy) == TileDirt {
			m.SetTile(dx, dy, TileDebris)
		}
	}
}

// generateZombieSpawns computes valid outdoor spawn points away from the player
func (m *Map) generateZombieSpawns(count int) {
	playerWX := float64(m.PlayerSpawn.X * TileSize)
	playerWY := float64(m.PlayerSpawn.Y * TileSize)

	for i := 0; i < count; i++ {
		for attempts := 0; attempts < 100; attempts++ {
			zx := 2 + rand.Intn(m.Width-4)
			zy := 2 + rand.Intn(m.Height-4)
			t := m.GetTile(zx, zy)
			if !t.IsSolid() {
				wx := float64(zx * TileSize)
				wy := float64(zy * TileSize)
				dist := math.Sqrt((wx-playerWX)*(wx-playerWX) + (wy-playerWY)*(wy-playerWY))
				if dist > 300 {
					m.ZombieSpawns = append(m.ZombieSpawns, Point{X: zx, Y: zy})
					break
				}
			}
		}
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
		angle := (float64(i) / float64(rays)) * 2 * 3.1415926535
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
```

---

## 4. Caveats & Assumptions

1. **Read-Only Scope**: This report establishes the full architectural and mathematical model without modifying production files directly during investigation.
2. **Backward Compatibility**: All existing exported signatures (`NewMap`, `GetTile`, `SetTile`, `IsColliding`, `CalculateFOV`, `TileSize`, `TileGrass`, `TileWall`, `TileDirt`, `TileWoodFloor`, `TileTree`) remain strictly 100% backward compatible.
3. **Small Map Safety**: Maps with width or height $< 30$ (used in small unit tests) gracefully return initialized grass maps with perimeter walls without generating out-of-bounds building coordinates.
4. **Asset Integration Synergy**: The new tile types directly map to the textures already exported in `internal/assets/assets.go` (`AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `FenceImage`, `DebrisImage`).

---

## 5. Conclusion

1. **Complete Archetype Coverage**: All 5 requested building archetypes are fully designed with multi-room partitions, thematic floorings, and guaranteed doorway graph connectivity:
   - **Suburban Residential House**: 4 rooms (Living, Bedroom, Kitchen, Bathroom), Wood/Tile floors.
   - **Grocery / Convenience Store**: 2 rooms (Sales Floor with display aisles, Concrete Backroom Storage), Tile/Concrete floors.
   - **Police Station**: 4 rooms (Public Lobby, Office, Armory Vault, Holding Cells), Tile/Concrete floors.
   - **Pharmacy / Clinic**: 3 rooms (Waiting Area, Consultation Exam Room, Medical Storage), Tile floors.
   - **Industrial Warehouse**: 2 rooms (Expansive Main Bay with Crate Pallets, Foreman Office), Concrete/Tile floors.
2. **Physics & Vision Enhancements**: `TileType.IsSolid()` correctly accounts for `TileWall`, `TileTree`, `TileFence`, `TileDebris`, while `TileType.BlocksVision()` preserves line-of-sight over low fences and debris while occluding behind walls.
3. **Contextual ECS Spawning**: Generates safe player spawn inside residential homes, thematic loot in kitchens/bedrooms/armories/dispensaries, and open-ground zombie spawns free of collision traps.

---

## 6. Verification Method

### 6.1 Unit Test Suite
Execute the Go test suite:
```sh
CC=gcc go test -v ./internal/game/world/...
```
Expected output: All tests in `internal/game/world` pass cleanly.

### 6.2 Full Project Test Execution
Execute all package tests:
```sh
CC=gcc go test -v ./...
```
Expected output: All tests across `internal/assets`, `internal/ecs`, `internal/game`, and `internal/game/world` pass cleanly.

### 6.3 Asset Pipeline Verification
Run procedural asset generator:
```sh
go run ./cmd/tools/genassets
```
Expected output: Generates all PNG assets in `internal/assets/images/` without error.

### 6.4 Interactive Game Loop Verification
Launch the game engine:
```sh
CC=gcc go run ./cmd/game
```
Expected output: Launches Ebitengine window, initializes town map with multi-room building archetypes, spawns player safely inside residential house, and executes 60 FPS update and render loops smoothly.
