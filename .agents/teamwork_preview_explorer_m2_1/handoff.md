# Handoff Report: Milestone 2 — Procedural Town Layout, Road Networks & District Zoning

## 1. Observation

### 1.1 Existing World Generation Baseline
From `internal/game/world/map.go:8-106`:
- **Current Tile Palette** (`map.go:10-18`):
  ```go
  type TileType int
  const (
      TileGrass TileType = iota
      TileWall
      TileDirt
      TileWoodFloor
      TileTree
  )
  ```
- **Static World Generation** (`map.go:29-106`):
  - Fills outer boundary with `TileWall` and interior with `TileGrass`.
  - Places a static dirt crossroad (3 tiles wide) centered at `width/2` and `height/2`.
  - Places exactly 7 hardcoded rectangular houses in static positions:
    - Quadrant 1: `(10, 10, 10, 8)`, `(30, 15, 8, 12)`, `(60, 12, 12, 10)`, `(75, 30, 10, 10)`
    - Quadrant 2: `(15, 60, 12, 8)`, `(35, 70, 10, 10)`, `(70, 65, 15, 12)`
  - Uniform random tree scatter: 150 `TileTree` instances randomly placed on `TileGrass`.

### 1.2 Available Asset Palette from Milestone 1
From `internal/assets/assets.go:17-44` and `cmd/tools/genassets/main.go:20-47`:
- **Floor Sprites ($64 \times 32$ diamonds)**: `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`.
- **Vertical Sprites ($64 \times 64$ blocks)**: `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`.
- **Item / Equipment Sprites ($16 \times 16$)**: `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `FoodImage`, `WaterImage`.

### 1.3 Collision & Raycast FOV Systems
From `internal/game/world/map.go:108-183`:
- `IsColliding(rectX, rectY, rectW, rectH float64) bool` checks whether any covered tile is `TileWall` or `TileTree`.
- `CalculateFOV(playerX, playerY float64, radiusTiles int)` casts 360-degree rays from player tile, terminating rays upon encountering `TileWall`.
- `GetTile(x, y int)` returns `TileWall` when out of bounds, preventing out-of-bounds entity traversal.

### 1.4 Spawning Coupling Limitations in `internal/game/game.go`
From `internal/game/game.go:44-115`:
- Player spawn is hardcoded to center `(1600.0, 1600.0)`.
- 20 loot items and 150 zombies are spawned at uniform random coordinates `(100 + rand.Intn(3000), 100 + rand.Intn(3000))` without bounding box checks, causing entities to spawn trapped inside solid walls and trees.

---

## 2. Logic Chain & Architectural Design

### 2.1 Expanded Tile Representation
To support urban road networks, sidewalks, commercial plazas, and fenced yards, `TileType` must be expanded from 5 to 10 types:
1. `TileGrass` (passable ground)
2. `TileWall` (solid obstacle, blocks FOV vision)
3. `TileDirt` (passable ground trail)
4. `TileWoodFloor` (passable interior floor)
5. `TileTree` (solid obstacle)
6. `TileAsphalt` (passable paved roadway)
7. `TileConcrete` (passable paved sidewalk, driveway, plaza, warehouse floor)
8. `TileTileFloor` (passable commercial / kitchen / bathroom interior floor)
9. `TileFence` (solid obstacle, transparent to FOV vision)
10. `TileDebris` (solid obstacle, transparent to FOV vision)

Helper predicates on `TileType`:
- `IsSolid() bool`: `t == TileWall || t == TileTree || t == TileFence || t == TileDebris`
- `BlocksVision() bool`: `t == TileWall`
- `IsFloor() bool`: `t == TileGrass || t == TileDirt || t == TileWoodFloor || t == TileAsphalt || t == TileConcrete || t == TileTileFloor`

### 2.2 Hierarchical Procedural Town Layout Algorithm

```
+-----------------------------------------------------------------------------------+
|                        5-Phase Town Generation Pipeline                           |
|                                                                                   |
|  Phase 1: Road Network                                                            |
|    - Major Avenues (4-tile TileAsphalt + 2-tile TileConcrete sidewalks = 6-wide) |
|    - Secondary Cross-Streets (2-tile TileAsphalt + 2-tile TileConcrete = 4-wide) |
|    - Paved Intersections & Crosswalk Transitions                                  |
|                                                                                   |
|  Phase 2: District Zoning & Block Subdivision                                     |
|    - Quadrant NW: DistrictResidential (Suburban Neighborhood)                     |
|    - Quadrant NE: DistrictCommercial (Downtown Business & Municipal Core)         |
|    - Quadrant SW: DistrictPark (Nature Reserve, Gazebo Pavilion & Trails)         |
|    - Quadrant SE: DistrictIndustrial (Logistics Warehouses & Scrap Yards)         |
|                                                                                   |
|  Phase 3: Parcel Subdivision & Building Synthesis                                |
|    - Multi-Room Houses (Living Room, Kitchen, Bedroom, Bathroom, Driveways)       |
|    - Commercial Centers (Supermarket, Pharmacy, Police Station & Armory, Plaza)   |
|    - Industrial Sites (Warehouses, Security Fences, Loading Docks, Scrap Sheds)   |
|    - Parks (Pavilion, Dirt Cross-Trails, Organic Tree Copses)                     |
|                                                                                   |
|  Phase 4: Environmental Props & Boundaries                                        |
|    - Impassable Border Wall & Wilderness Tree Perimeter                           |
|    - Backyard & Security Fences (TileFence)                                       |
|    - Loading Bay Rubble & Alleys (TileDebris)                                     |
|                                                                                   |
|  Phase 5: Contextual Thematic Entity Spawns                                       |
|    - Player Safe Spawn inside Residential House Bedroom/Living Room               |
|    - Thematic Loot Spawns categorized by room type                                |
|    - Zombie Spawns (150 count) with district danger weighting & collision safety  |
+-----------------------------------------------------------------------------------+
```

### 2.3 Road Network Specification
1. **Major Avenues**:
   - Vertical Main Avenue: centered at $midX = W/2$. Roadway columns $[midX-2 .. midX+1]$ (`TileAsphalt`), sidewalk at $midX-3$ and $midX+2$ (`TileConcrete`).
   - Horizontal Main Avenue: centered at $midY = H/2$. Roadway rows $[midY-2 .. midY+1]$ (`TileAsphalt`), sidewalk at $midY-3$ and $midY+2$ (`TileConcrete`).
   - Central intersection: $4 \times 4$ asphalt crossway with 4 corner sidewalk joints.
2. **Secondary Cross-Streets**:
   - Vertical cross-streets placed at midpoints of left and right quadrants ($secX_1 \approx W/4, secX_2 \approx 3W/4$).
   - Horizontal cross-streets placed at midpoints of top and bottom quadrants ($secY_1 \approx H/4, secY_2 \approx 3H/4$).
   - 2-tile wide asphalt with 1-tile wide concrete sidewalks.
   - Subdivides each quadrant into 4 discrete urban blocks (16 total city blocks).

### 2.4 District Zoning & Building Archetypes

| District | Quadrant | Theme | Building Archetypes | Floorings & Props | Loot Profile | Zombie Danger |
|----------|----------|-------|---------------------|-------------------|--------------|---------------|
| **Residential** | Top-Left (NW) | Suburbs | Multi-room single-family homes ($11 \times 9$) | `TileWoodFloor`, `TileTileFloor`, front `TileConcrete` driveways, `TileFence` backyards | Food, water in kitchen; Armor vest, baseball bat in bedroom | Low (35 outdoor/yard zombies, $>350\text{px}$ from start) |
| **Commercial** | Top-Right (NE) | Downtown | Supermarket ($15 \times 14$), Pharmacy ($14 \times 12$), Police Station ($15 \times 13$), Retail Plaza | `TileTileFloor`, `TileConcrete` plaza walkways, back alley dumpsters (`TileDebris`) | Abundant food/water in grocery; Medical in pharmacy; Shotguns, ammo, axes, armor in police armory | Moderate-High (45 zombies, 20% runners) |
| **Industrial** | Bottom-Right (SE) | Logistics | Logistics Warehouse A ($15 \times 12$), Heavy Warehouse B ($14 \times 12$), Storage Units, Scrap Shed | `TileConcrete` loading bays, perimeter `TileFence` with gates, `TileDebris` scrap piles | Fire axes, shotguns, ammo crates, armor vests, tools | High (45 zombies, 30% runners) |
| **Parks** | Bottom-Left (SW) | Nature | Wooden Gazebo Pavilion ($7 \times 7$), Picnic Clearings | `TileGrass`, winding `TileDirt` trails, clustered `TileTree` copses, `TileWoodFloor` pavilion | Outdoor food/water rations | Low-Moderate (25 roaming zombies) |

### 2.5 Lot Subdivision & Parcel Placement
- **Residential Lots**: Front setback facing street with a 2-tile wide `TileConcrete` driveway connecting the road to the house carport. House is set back by 2 tiles. The rear of the lot is enclosed by a `TileFence` perimeter forming a private backyard.
- **Commercial Lots**: Zero front setback with main double-door entrance opening directly onto the concrete sidewalk. Rear service alley with rear exit door for deliveries and debris bins.
- **Industrial Lots**: Perimeter security fence (`TileFence`) enclosing the entire lot, with an open double-wide concrete gate facing the access road, leading to a concrete loading apron and roll-up warehouse door.
- **Park Lots**: Open frontage with intersecting dirt pathways leading from street sidewalks into central tree groves and pavilion clearings.

### 2.6 Pure Go Data Structures

```go
// Data structures formulated in proposed_map.go:

type DistrictType int
const (
    DistrictResidential DistrictType = iota
    DistrictCommercial
    DistrictIndustrial
    DistrictPark
)

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

type Rect struct {
    X, Y, W, H int
}

type Point struct {
    X, Y float64
}

type LootSpawn struct {
    ItemType string  // "food", "water", "weapon", "axe", "shotgun", "ammo", "armor"
    X, Y     float64 // World coordinates in pixels
    RoomType string  // "kitchen", "bedroom", "armory", "grocery", "pharmacy", "warehouse", "outdoor"
}

type ZombieSpawn struct {
    X, Y     float64 // World coordinates in pixels
    IsRunner bool
    District DistrictType
}

type Room struct {
    Name   string
    Bounds Rect
    Floor  TileType
}

type Building struct {
    Type     BuildingType
    District DistrictType
    Bounds   Rect
    Rooms    []Room
    Doors    []Point
}

type Lot struct {
    District DistrictType
    Bounds   Rect
    Building *Building
}

type District struct {
    Type   DistrictType
    Bounds Rect
    Lots   []Lot
}

type Map struct {
    Width, Height int
    Tiles         []TileType
    Visible       []bool
    Explored      []bool

    Districts    []District
    Buildings    []Building
    PlayerSpawn  Point
    LootSpawns   []LootSpawn
    ZombieSpawns []ZombieSpawn
}
```

---

## 3. Caveats
- No source code files in `internal/game/world/` were modified during this investigation phase (read-only compliance).
- Reference code has been saved in `.agents/teamwork_preview_explorer_m2_1/proposed_map.go` for direct adoption by the implementing agent.
- Coordinate system consistency: Grid tiles are indexed at `TileSize = 32` pixels. World coordinates are $X = \text{tileX} \times 32 + 16, Y = \text{tileY} \times 32 + 16$. Isometric rendering maps $(X, Y) \rightarrow (X - Y, (X + Y)/2)$.

---

## 4. Conclusion
The procedural town layout algorithm designed for Milestone 2 resolves all limitations of the prototype:
1. **Rich Road Hierarchy**: Implements major avenues with sidewalks, secondary cross-streets, driveways, and intersections using `TileAsphalt` and `TileConcrete`.
2. **Distinct District Zoning**: Divides the town into 4 thematic quadrants (Residential, Commercial, Industrial, Park).
3. **Structured Lot Subdivision**: Generates street frontages, driveways, fenced backyards, commercial storefronts, and industrial loading yards.
4. **Architectural Building Archetypes**: Constructs multi-room buildings (Suburban Houses, Supermarkets, Pharmacies, Police Stations, Warehouses, Storage Facilities, Park Pavilions) with realistic room partitions and floorings.
5. **Contextual & Collision-Safe Spawns**: Extracts structured metadata guaranteeing safe player placement, thematic loot distribution, and collision-verified zombie placement.

---

## 5. Verification Method

### 5.1 Reference Model Verification
Inspect the complete proposed implementation:
```sh
cat /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_1/proposed_map.go
```

### 5.2 Unit Test Execution
When applied to `internal/game/world/map.go`:
```sh
CC=gcc go test -v ./internal/game/world
```
Expected output:
- `TestNewMap` passes (validates map dimensions, tile count, boundary walls).
- `TestIsColliding` passes (validates AABB collision detection against solid tiles).

### 5.3 Full Test Suite & Asset Pipeline
```sh
go run ./cmd/tools/genassets
CC=gcc go test -v ./...
```
Expected output: All unit tests across all packages pass cleanly.
