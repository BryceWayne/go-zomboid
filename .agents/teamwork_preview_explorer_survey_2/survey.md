# Comprehensive Technical Survey: World Map, Tile Systems & Asset Mapping (Requirement R3)

**Author:** `teamwork_preview_explorer_survey_2`  
**Date:** 2026-08-29  
**Target Repository:** `github.com/BryceWayne/go-zomboid`  
**Scope:** `internal/game/world/`, `internal/game/`, `internal/assets/`, `context/`

---

## 1. Executive Summary

This survey provides a technical investigation of the world map, tile definitions, procedural generation pipeline, collision/FOV mechanics, and rendering systems in `go-zomboid`. It is conducted to support **Requirement R3** ("Infer and Implement New Logic: Analyze the imported assets [e.g., Benches, Chests, Sculptures] and automatically infer their mapping into the game world. Create new `TileType` constants in `internal/game/world/map.go` and update the `DrawSystem` in `internal/game/game.go` to properly render and depth-sort any objects that did not previously exist").

### Key Findings
1. **World Grid Architecture:** `internal/game/world/map.go` uses a flat 2D tile grid (`Width * Height` flat slices for `Tiles`, `Visible`, `Explored`) with standard tile size `TileSize = 128` (128x128 orthogonal world pixels). Coordinate transformations between 2D World pixels and 2.5D Isometric screen space follow `isoX = wx - wy` and `isoY = (wx + wy) / 2.0` with a 2:1 isometric ratio.
2. **Current `TileType` System:** There are currently 16 `TileType` constants (IDs 0 through 15) categorized into floor surfaces (`IsFloor() == true`), solid vertical obstacles (`IsSolid() == true`), vision blockers (`BlocksVision() == true`), and non-solid props.
3. **Asset Inventory in `context/`:** External assets include rich prop sets in `Small Forest` (`Bench.png`, `Chest.png`, `Sculpture-1.png`, `Sculture-2.png`, `Stone-1..2.png`, `Bush-1..4.png`, `Flower-1..3.png`, `Stump.png`, `Grass-1..2.png`, fences, trees, and ground tilesets), `Lab` (`Inside_C.png`), and `Zombie Apocalypse Tileset` (Urban Assets, Modular Barns, Roads, Pickable Items, Character and Zombie animation sheets).
4. **New `TileType` Mapping:** We formulate 6 new `TileType` constants: `TileBench` (16), `TileChest` (17), `TileSculpture` (18), `TileBush` (19), `TileFlower` (20), `TileStone` (21), with complete property mappings (`IsSolid`, `BlocksVision`, `IsFloor`, `String()`).
5. **Rendering & Depth-Sorting Integration:** In `internal/game/game.go`, `DrawSystem.Draw` executes a 2-pass isometric rendering pipeline: (Pass 1) Ground Diamond rendering on flat grid; (Pass 2) Y-sorted sprite rendering with depth key `Depth = worldX + worldY`. Integrating the new props requires adding underlying floor rendering in Pass 1 and registering the sprite definitions with `assets.<Asset>Image` in Pass 2.
6. **Test Suite Audit & Compatibility:** All 21 test suites across the repository (`map_test.go`, `world_empirical_stress_test.go`, `game_stress_test.go`, etc.) were analyzed. The core map tests verify that the first 10 foundational tile types exist in `NewMap(100, 100)`. Adding the new `TileType`s maintains 100% backward compatibility while extending game world fidelity.

---

## 2. World Map Architecture & Implementation

### 2.1 Map Data Structure (`internal/game/world/map.go`)

The world map is represented by the `Map` struct:

```go
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
```

#### Key Characteristics:
- **Flat 1D Slices:** Memory efficiency and cache locality are maximized using flat slices of size `Width * Height`. Indexing is performed via `idx = y * Width + x`.
- **Spatial Resolution (`TileSize`):** `TileSize = 128` defines each cell as 128x128 pixels in 2D Cartesian world space. A standard 100x100 map spans 12,800 x 12,800 pixels.
- **Bounds Safety:** `GetTile(x, y)` automatically returns `TileWall` for any out-of-bounds coordinates, ensuring entities and raycasts cannot escape map boundaries.
- **Visibility & Exploration:** Two boolean slices track current field of view (`Visible`) and persistent fog-of-war memory (`Explored`).

### 2.2 Coordinate Spaces & Isometric Projection

`go-zomboid` operates across three distinct coordinate frames:

| Coordinate System | Units | Definition | Example / Conversion |
|---|---|---|---|
| **Tile Grid** | Integers $(x, y)$ | Grid cell indices $[0, \text{Width}-1] \times [0, \text{Height}-1]$ | Tile $(10, 8)$ |
| **World Space (2D)** | Pixels $(wx, wy)$ | Continuous Cartesian pixel space | $wx = x \times 128.0 + 64.0$ |
| **Isometric Space (2.5D)** | Pixels $(isoX, isoY)$ | Diamond projection space (2:1 aspect ratio) | $isoX = wx - wy$<br>$isoY = (wx + wy) / 2.0$ |
| **Screen Space** | Pixels $(sx, sy)$ | Render viewport $(1280 \times 720)$ with camera offset | $sx = (isoX - camX) \times 0.5 + 640.0$<br>$sy = (isoY - camY) \times 0.5 + 360.0$ |

Inverse transformation is handled cleanly in `IsoToWorld(isoX, isoY)`:
$$wx = isoY + \frac{isoX}{2.0}, \quad wy = isoY - \frac{isoX}{2.0}$$

### 2.3 Procedural Town Generation Pipeline (`NewMap`)

The `NewMap(width, height)` constructor executes 5 ordered generation phases:

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Base Terrain Fill (TileGrass) & Perimeter (TileWall)      │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 2. Road Network (Main Ave, Boulevard, Access Roads, Trails) │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 3. Multi-Room Buildings (Residential, Grocery, Police, etc.)│
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 4. Outdoor Environmental Props (Debris, Trees, Campsite)    │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 5. Contextual Spawns (Player Spawn, Room Loot, Safe Zombies)│
└─────────────────────────────────────────────────────────────┘
```

1. **Base Terrain:** Fills the interior with `TileGrass` and perimeters with `TileWall`.
2. **Road Network:**
   - East-West Main Avenue at `midY` (3 tiles `TileAsphalt`, bordered by `TileConcrete` sidewalks).
   - North-South Boulevard at `midX` (3 tiles `TileAsphalt`, bordered by `TileConcrete` sidewalks).
   - Secondary residential access road at $Y \approx 0.25 \times \text{Height}$ and industrial road at $Y \approx 0.75 \times \text{Height}$.
   - Dirt walking trails (`TileDirt`).
3. **District Architecture:**
   - **Residential District (NW):** Houses 1–3 with partitioned living rooms, kitchens, bedrooms, bathrooms, and fenced yards (`TileFence`).
   - **Commercial District (NE):** Grocery Store (sales floor, concrete apron, shelves), Pharmacy / Clinic (consultation room, medical storage).
   - **Municipal & Defense District (SW):** Police Station & Armory (holding cell, armory with concrete floors, motor pool courtyard), House 4.
   - **Industrial District (SE):** Warehouse Depot (20x14, concrete flooring, loading bay doors, foreman office, fenced yard).
4. **Environmental Props (`placeEnvironmentalProps`):**
   - Fixed debris clusters (`TileDebris`).
   - Park tree groves (`TileTree`).
   - 120 randomized tree placements (90% `TileTree`, 10% `TileStump`, 30% `TileMushroom` proximity).
   - Campsite in NE area: `TileTent`, `TileSign`, `TileElevationBlock`, `TileRamp`.
   - Roadside signs (`TileSign`).
5. **Spawn Point Extraction:**
   - Safe player spawn centered in House 1 living room.
   - Room-specific loot spawns (`food`, `water`, `weapon`, `axe`, `shotgun`, `ammo`, `armor`, `antidote`).
   - 140 zombie spawns placed exclusively on non-solid tiles at least 1400.0px away from player spawn.

---

## 3. Tile System Mechanics & Properties

### 3.1 Existing `TileType` Constants

`internal/game/world/map.go` currently defines 16 tile types:

```go
const (
	TileGrass TileType = iota // 0
	TileWall                  // 1
	TileDirt                  // 2
	TileWoodFloor             // 3
	TileTree                  // 4
	TileAsphalt               // 5
	TileConcrete              // 6
	TileTileFloor             // 7
	TileFence                 // 8
	TileDebris                // 9
	TileTent                  // 10
	TileElevationBlock        // 11
	TileRamp                  // 12
	TileStump                 // 13
	TileMushroom              // 14
	TileSign                  // 15
)
```

### 3.2 Property Method Matrix

Each `TileType` implements 4 core behavioral methods:

| TileType | ID | `IsSolid()` | `BlocksVision()` | `IsFloor()` | `String()` | Description / Role |
|---|---|:---:|:---:|:---:|---|---|
| `TileGrass` | 0 | `false` | `false` | `true` | `"Grass"` | Base outdoor terrain |
| `TileWall` | 1 | `true` | `true` | `false` | `"Wall"` | Structural & perimeter wall (blocks raycast) |
| `TileDirt` | 2 | `false` | `false` | `true` | `"Dirt"` | Walking trails and garden patches |
| `TileWoodFloor` | 3 | `false` | `false` | `true` | `"WoodFloor"` | Residential interior flooring |
| `TileTree` | 4 | `true` | `false` | `false` | `"Tree"` | Forest & park obstacle (transparent to vision) |
| `TileAsphalt` | 5 | `false` | `false` | `true` | `"Asphalt"` | Roadway pavement |
| `TileConcrete` | 6 | `false` | `false` | `true` | `"Concrete"` | Sidewalks, commercial & warehouse floor |
| `TileTileFloor` | 7 | `false` | `false` | `true` | `"TileFloor"` | Kitchen, bathroom, store & clinic tile |
| `TileFence` | 8 | `true` | `false` | `false` | `"Fence"` | Yard & perimeter fence |
| `TileDebris` | 9 | `true` | `false` | `false` | `"Debris"` | Rubble & obstacle pile |
| `TileTent` | 10 | `true` | `false` | `false` | `"Tent"` | Campsite shelter |
| `TileElevationBlock` | 11 | `true` | `false` | `false` | `"ElevationBlock"` | Raised elevation terrain |
| `TileRamp` | 12 | `false` | `false` | `true` | `"Ramp"` | Elevation transition floor |
| `TileStump` | 13 | `true` | `false` | `false` | `"Stump"` | Tree stump obstacle |
| `TileMushroom` | 14 | `false` | `false` | `false` | `"Mushroom"` | Non-solid forest vegetation prop |
| `TileSign` | 15 | `true` | `false` | `false` | `"Sign"` | Roadside & campsite marker |

### 3.3 Collision & FOV Subsystems

1. **AABB Collision Engine (`Map.IsColliding`):**
   - Takes arbitrary entity bounding box $[rectX, rectX+rectW] \times [rectY, rectY+rectH]$.
   - Converts to tile index range:
     $$\text{minTileX} = \lfloor rectX / 128 \rfloor, \quad \text{maxTileX} = \lfloor (rectX+rectW) / 128 \rfloor$$
     $$\text{minTileY} = \lfloor rectY / 128 \rfloor, \quad \text{maxTileY} = \lfloor (rectY+rectH) / 128 \rfloor$$
   - Iterates through all overlapping cells and returns `true` if any cell satisfies `tile.IsSolid() == true` or crosses map boundaries.
2. **FOV Raycasting Engine (`Map.CalculateFOV`):**
   - Casts $radiusTiles \times 8$ radial rays from player position $(px, py)$ across $360^\circ$.
   - Steps outward along unit direction vectors $(\cos\theta, \sin\theta)$.
   - Marks each intersected tile as `Visible` and `Explored`.
   - Halts ray progression immediately when `Map.BlocksVision(tx, ty)` returns `true` (occlusion). Only `TileWall` occludes FOV; all fences, trees, debris, and props permit line-of-sight penetration.

---

## 4. Analysis of New Assets in `context/` & New `TileType` Specifications

### 4.1 Asset Packages in `context/`

Inspection of `context/` reveals 3 asset directories:

```
context/
├── Small Forest/
│   ├── Bench and chest/
│   │   ├── Bench.png (52x37)
│   │   └── Chest.png (22x21)
│   ├── Sculptures/
│   │   ├── Sculpture-1.png (23x31)
│   │   └── Sculture-2.png (29x32)
│   ├── Stones/
│   │   ├── Stone-1.png (28x19)
│   │   └── Stone-2.png (29x25)
│   ├── Bushes/
│   │   ├── Bush-1.png (24x18), Bush-2.png (19x15), Bush-3.png (25x19), Bush-4.png (28x19)
│   │   └── Stump.png (29x19)
│   ├── Flowers/
│   │   ├── Flower-1.png (26x25), Flower-2.png (24x22), Flower-3.png (26x18)
│   │   └── Grass-1.png (25x24), Grass-2.png (31x15)
│   ├── Fences/ (Wooden fence, Stone fence, Big wooden fence)
│   ├── Trees/ (Tree-1, Tree-2, Tree-3 stage variations)
│   └── Ground tileset/ (Bright-grass, Dark-grass, Earth, Stone-path tilesets)
├── Lab/
│   └── Inside_C.png (768x768)
└── Zombie Apocalypse Tileset/
    └── Organized separated sprites/
        ├── Urban Assets/ (vending machines, street lights, hydrants)
        ├── Broken Cars and Tires/
        ├── Pickable Items and Weapons/
        └── Character & Zombie Walking / Attack Animation Sheets
```

### 4.2 Proposed New `TileType` Constants

To fulfill **Requirement R3** cleanly, we recommend adding the following dedicated `TileType` constants:

```go
const (
	TileGrass TileType = iota // 0
	TileWall                  // 1
	TileDirt                  // 2
	TileWoodFloor             // 3
	TileTree                  // 4
	TileAsphalt               // 5
	TileConcrete              // 6
	TileTileFloor             // 7
	TileFence                 // 8
	TileDebris                // 9
	TileTent                  // 10
	TileElevationBlock        // 11
	TileRamp                  // 12
	TileStump                 // 13
	TileMushroom              // 14
	TileSign                  // 15
	
	// New World Tile & Prop Constants (Requirement R3)
	TileBench                 // 16
	TileChest                 // 17
	TileSculpture             // 18
	TileBush                  // 19
	TileFlower                // 20
	TileStone                 // 21
)
```

### 4.3 Behavioral Specification for New Tile Types

| TileType | Constant | Value | `IsSolid()` | `BlocksVision()` | `IsFloor()` | `String()` | Physical & Visual Semantics |
|---|---|:---:|:---:|:---:|:---:|---|---|
| **Bench** | `TileBench` | 16 | `true` | `false` | `false` | `"Bench"` | Park/street furniture. Blocks movement, transparent to FOV, renders in Y-sorted sprite pass over grass/concrete. |
| **Chest** | `TileChest` | 17 | `true` | `false` | `false` | `"Chest"` | Storage container / loot cache. Solid obstacle, transparent to FOV, renders in Y-sorted sprite pass. |
| **Sculpture** | `TileSculpture` | 18 | `true` | `false` | `false` | `"Sculpture"` | Monument statue. Solid obstacle, transparent to FOV, renders in Y-sorted sprite pass as park/plaza centerpiece. |
| **Bush** | `TileBush` | 19 | `false` | `false` | `false` | `"Bush"` | Foliage shrub. Walkable (non-solid), transparent to FOV, renders in Y-sorted sprite pass over grass. |
| **Flower** | `TileFlower` | 20 | `false` | `false` | `false` | `"Flower"` | Wildflower / garden cluster. Walkable (non-solid), transparent to FOV, renders in Y-sorted sprite pass over grass. |
| **Stone** | `TileStone` | 21 | `true` | `false` | `false` | `"Stone"` | Natural rock / boulder. Solid obstacle, transparent to FOV, renders in Y-sorted sprite pass over grass/dirt. |

#### Method Implementation in `internal/game/world/map.go`:

```go
func (t TileType) IsSolid() bool {
	switch t {
	case TileWall, TileTree, TileFence, TileDebris, TileTent, TileElevationBlock, TileStump, TileSign,
		TileBench, TileChest, TileSculpture, TileStone:
		return true
	default:
		return false
	}
}

func (t TileType) BlocksVision() bool {
	return t == TileWall
}

func (t TileType) IsFloor() bool {
	switch t {
	case TileGrass, TileDirt, TileWoodFloor, TileAsphalt, TileConcrete, TileTileFloor, TileRamp:
		return true
	default:
		return false
	}
}

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
	case TileTent:
		return "Tent"
	case TileElevationBlock:
		return "ElevationBlock"
	case TileRamp:
		return "Ramp"
	case TileStump:
		return "Stump"
	case TileMushroom:
		return "Mushroom"
	case TileSign:
		return "Sign"
	case TileBench:
		return "Bench"
	case TileChest:
		return "Chest"
	case TileSculpture:
		return "Sculpture"
	case TileBush:
		return "Bush"
	case TileFlower:
		return "Flower"
	case TileStone:
		return "Stone"
	default:
		return "Unknown"
	}
}
```

---

## 5. World Map Placement & Generation Strategy

### 5.1 Placement Mapping Strategy

To ensure all new objects are visibly generated on the map without disrupting existing building footprints, road traffic corridors, or safe player/zombie spawns, we recommend placing them in `placeEnvironmentalProps`:

```
                                  [North]
                  +--------------------------------------+
                  | Residential District  Commercial     |
                  |  - Benches in yards    - Benches on  |
                  |  - Flower beds          storefronts  |
                  |                       - Sculpture in |
                  |                         NE Park      |
                  | [Main West-East Avenue]              |
 [West]           |--------------------------------------|           [East]
                  | [Main North-South Boulevard]         |
                  |                                      |
                  | Police / Municipal    Industrial     |
                  |  - Stones & Bushes     - Chests in   |
                  |    along perimeter       Warehouse   |
                  |  - Benches in court    - Crates/junk |
                  +--------------------------------------+
                                  [South]
```

### 5.2 Specific Placement Points in `placeEnvironmentalProps`

1. **Town Plaza & Memorial Park Centerpiece (`TileSculpture`):**
   - Place a grand sculpture in the NE park plaza at `(midX + 30, 5)` and another near the municipal courtyard at `(midX - 10, midY + 4)`.
   - Surround each sculpture with colorful `TileFlower` and `TileBench` installations.
2. **Park & Sidewalk Seating (`TileBench`):**
   - Place benches along the concrete sidewalks at `(midX - 3, midY - 6)` and `(midX + 3, midY + 6)`.
   - Place benches in residential front yards: House 1 yard `(12, 6)` and House 2 yard `(30, 6)`.
   - Place benches in the storefront apron in front of the Grocery store at `(gX + 4, gY + 13)` and Clinic at `(pX + 4, pY + 11)`.
3. **Loot Caches & Hidden Chests (`TileChest`):**
   - Place chests inside the warehouse back storage corner at `(wX + 16, wY + 2)`.
   - Place a chest in the top-right campsite at `(campX + 2, campY + 1)`.
   - Place a chest in House 1 backyard at `(h1X + 1, h1Y - 1)`.
4. **Natural Landscape Foliage & Rocks (`TileBush`, `TileFlower`, `TileStone`):**
   - In the randomized prop placement loop (120 iterations), distribute:
     - 40% `TileTree`
     - 15% `TileBush` (shrubbery)
     - 15% `TileStone` (rocks along walking paths)
     - 15% `TileFlower` (wildflower patches)
     - 10% `TileStump`
     - 5% `TileMushroom`

---

## 6. Isometric Rendering & Depth-Sorting Pipeline (`internal/game/game.go`)

### 6.1 Rendering Pipeline Architecture

`DrawSystem.Draw` executes a strict layered rendering sequence:

```
Screen.Fill(Background)
   │
   ▼
[Pass 1] Ground Tiles (Ground diamonds, flat isometric projection)
   │
   ▼
[Pass 2] Y-Depth Sorted Sprites
   ├── Static Vertical Tiles (Walls, Trees, Fences, Debris, Tents, Benches, Chests, Sculptures, etc.)
   ├── Dropped World Items (Weapons, Food, Ammo, Armor, etc.)
   ├── Entities (Player, Standard Zombies, Runner Zombies)
   └── Player Facing Indicator
   │
   ▼
[Pass 3] Bezier Combat Swing & Blast Swoosh Arcs
   │
   ▼
[Pass 4] Day-Night Ambient Lighting Pass
   │
   ▼
[Pass 5] HUD / UI Overlay (Health, Hunger, Thirst, Armor, Weapon, Inventory)
```

### 6.2 Pass 1: Ground Tile Diamond Pass

For all vertical props (including the new `TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`), the underlying terrain diamond must be rendered first during Pass 1 so there are no black voids or holes in the terrain mesh:

```go
switch t {
case world.TileGrass, world.TileTree, world.TileFence, world.TileTent, world.TileElevationBlock,
	world.TileRamp, world.TileStump, world.TileMushroom, world.TileSign,
	world.TileBench, world.TileChest, world.TileSculpture, world.TileBush, world.TileFlower, world.TileStone:
	screen.DrawImage(assets.GrassImage, op)
case world.TileDirt:
	screen.DrawImage(assets.DirtImage, op)
case world.TileWoodFloor:
	screen.DrawImage(assets.WoodImage, op)
case world.TileAsphalt:
	screen.DrawImage(assets.AsphaltImage, op)
case world.TileConcrete, world.TileDebris:
	screen.DrawImage(assets.ConcreteImage, op)
case world.TileTileFloor:
	screen.DrawImage(assets.TileFloorImage, op)
}
```

### 6.3 Pass 2: Sprite Gathering & Y-Sorting

Vertical objects are appended to `sprites []Renderable` with depth key:
$$\text{Depth} = worldX + worldY$$

```go
for y := 0; y < s.gameMap.Height; y++ {
	for x := 0; x < s.gameMap.Width; x++ {
		t := s.gameMap.GetTile(x, y)
		if t == world.TileWall || t == world.TileTree || t == world.TileFence || t == world.TileDebris ||
			t == world.TileTent || t == world.TileElevationBlock || t == world.TileRamp || t == world.TileStump ||
			t == world.TileMushroom || t == world.TileSign ||
			t == world.TileBench || t == world.TileChest || t == world.TileSculpture ||
			t == world.TileBush || t == world.TileFlower || t == world.TileStone {

			worldX := float64(x * world.TileSize)
			worldY := float64(y * world.TileSize)

			// ... distance and visibility checks ...

			isoX, isoY := WorldToIso(worldX, worldY)

			var img *ebiten.Image
			switch t {
			case world.TileWall:
				img = assets.WallImage
			case world.TileTree:
				img = assets.TreeImage
			case world.TileFence:
				img = assets.FenceImage
			case world.TileDebris:
				img = assets.DebrisImage
			case world.TileTent:
				img = assets.TentImage
			case world.TileElevationBlock:
				img = assets.ElevationBlockImage
			case world.TileRamp:
				img = assets.ElevationRampImage
			case world.TileStump:
				img = assets.StumpImage
			case world.TileMushroom:
				img = assets.MushroomImage
			case world.TileSign:
				img = assets.SignImage
			case world.TileBench:
				img = assets.BenchImage
			case world.TileChest:
				img = assets.ChestImage
			case world.TileSculpture:
				img = assets.SculptureImage
			case world.TileBush:
				img = assets.BushImage
			case world.TileFlower:
				img = assets.FlowerImage
			case world.TileStone:
				img = assets.StoneImage
			}

			if img == nil {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-128, -128)
			op.GeoM.Translate(isoX-camX, isoY-camY)
			op.GeoM.Scale(0.5, 0.5)
			op.GeoM.Translate(640, 360)

			if !s.gameMap.Visible[idx] && s.gameMap.Explored[idx] {
				op.ColorScale.Scale(0.2, 0.2, 0.3, 1) // Memory tint
			}

			sprites = append(sprites, Renderable{
				Image: img,
				Depth: worldX + worldY,
				Op:    op,
			})
		}
	}
}
```

Sorting with `sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].Depth < sprites[j].Depth })` guarantees proper isometric occlusion between the player, moving zombies, dropped items, walls, benches, chests, sculptures, and trees.

---

## 7. Repository Map & World Test Suite Audit

All 21 test files across the repository were investigated for assertions touching `world`, `TileType`, and isometric rendering:

| Test File | Key Test Cases | Observations & Invariants |
|---|---|---|
| `internal/game/world/map_test.go` | `TestTileTypeProperties`, `TestNewMapProceduralTown`, `TestPlayerSafeSpawn`, `TestContextualLootSpawns`, `TestZombieSpawnsNoTrapping`, `TestCollisionDetection`, `TestFOVAndOcclusion`, `TestSmallFallbackMap` | Verifies `solidTiles`, `nonSolidTiles`, `floorTiles`, `verticalTiles`, `BlocksVision`, `String()`, map dimensions (100x100), 4+ buildings, player spawn $> 1400$px from zombies, loot walkability, collision AABB checks. |
| `internal/game/world/world_empirical_stress_test.go` | `TestEmpirical_All10TileTypesGenerated`, `TestEmpirical_All5BuildingArchetypesAndRooms`, `TestEmpirical_PlayerSpawnSafetyAndZombieDistance`, `TestEmpirical_100PercentZombieSpawnsNonSolid`, `TestEmpirical_AABBCollisionSolidVsFloor`, `TestEmpirical_FOVRaycastingWallVsFence`, `TestEmpirical_LootDistributionAndWalkability` | Validates that the original 10 TileTypes are generated ($> 0$ count across map), all 5 building types have valid bounds and doors, player spawn is safe across 30 seeds, 100% zombies on non-solid tiles, AABB sweeping collisions, FOV wall raycast blocking. |
| `internal/game/game_stress_test.go` | `TestIsometricRenderingAllTileTypesAndPropsStress`, `TestGameLoopContinuousSimulationStress` | Lays out all tile types in a test map grid, instantiates player, items, standard/runner/stunned zombies, runs 24h day-night cycle, fog of war, and dead player states. Continuous 2500-frame simulation. |
| `internal/game/camera_empirical_challenger_test.go` & `camera_test.go` | Camera lerp, snapping, Screen-to-World, World-to-Iso | Verifies `world.TileSize = 128` scale, camera centering, isometric transformations. |
| `internal/game/e2e_tiers_test.go` | `world.TileSize == 128`, map boundary collision | Verifies out-of-bounds collision handling. |
| `internal/assets/empirical_challenger_test.go` & `challenger_stress_test.go` | Determinism test invoking `exec.Command("go", "run", "./cmd/tools/genassets")` | **Important Caveat for Team:** These tests previously executed `cmd/tools/genassets`. When `cmd/tools/genassets` is retired per Requirement R1, the asset tests must test the loaded native PNG assets in `internal/assets/images/` without shelling out to `genassets`. |

### Test Invariant Safety
- Adding `TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone` will **NOT** break any existing tests in `internal/game/world/` because existing tests assert counts on `expectedTiles` (the 10 core tiles), which remain generated.
- `map_test.go` and `game_stress_test.go` should be updated to include the new tile types in `TestTileTypeProperties` and `TestIsometricRenderingAllTileTypesAndPropsStress` for 100% test coverage.

---

## 8. Concrete Recommendations & Implementation Roadmap (R3)

### Step 1: Update `internal/game/world/map.go`
1. Define constants `TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`.
2. Update `IsSolid()`: Add `TileBench`, `TileChest`, `TileSculpture`, `TileStone`.
3. Update `BlocksVision()`: Retain `TileWall` as only vision blocker.
4. Update `IsFloor()`: Floor tiles remain flat surfaces; new props return `false`.
5. Update `String()`: Add cases for `"Bench"`, `"Chest"`, `"Sculpture"`, `"Bush"`, `"Flower"`, `"Stone"`.
6. Update `placeEnvironmentalProps`:
   - Place `TileSculpture` in NE park and town square.
   - Place `TileBench` along sidewalks and storefronts.
   - Place `TileChest` in warehouse and campsite.
   - Distribute `TileBush`, `TileFlower`, `TileStone` across outdoor grass/dirt.

### Step 2: Update `internal/assets/assets.go`
1. Add exported image variables:
   ```go
   var (
       BenchImage     *ebiten.Image
       ChestImage     *ebiten.Image
       SculptureImage *ebiten.Image
       BushImage      *ebiten.Image
       FlowerImage    *ebiten.Image
       StoneImage     *ebiten.Image
   )
   ```
2. In `Load()`, initialize from imported native PNGs:
   ```go
   BenchImage = loadEbitenImage("images/bench.png")
   ChestImage = loadEbitenImage("images/chest.png")
   SculptureImage = loadEbitenImage("images/sculpture.png")
   BushImage = loadEbitenImage("images/bush.png")
   FlowerImage = loadEbitenImage("images/flower.png")
   StoneImage = loadEbitenImage("images/stone.png")
   ```

### Step 3: Update `internal/game/game.go` (`DrawSystem`)
1. In `DrawSystem.Draw` Pass 1 (Ground diamond pass), include new props in `world.TileGrass` fallback case.
2. In `DrawSystem.Draw` Pass 2 (Sprite pass), handle new `TileType` constants, map to `assets.<Name>Image`, and append to `sprites` with `Depth = worldX + worldY`.

### Step 4: Extend Map & Rendering Tests
1. In `internal/game/world/map_test.go`:
   - Add `TileBench`, `TileChest`, `TileSculpture`, `TileStone` to `solidTiles`.
   - Add `TileBush`, `TileFlower` to `nonSolidTiles`.
   - Add string verification up to `TileStone`.
2. In `internal/game/game_stress_test.go`:
   - Add all 6 new `TileType`s to `allTiles` in `TestIsometricRenderingAllTileTypesAndPropsStress` to verify headless rendering across 24h lighting cycles.

---

## 9. Conclusion

The architecture of `go-zomboid`'s world map and tile systems is well-structured, modular, and ready for the integration of new external PNG assets to fulfill Requirement R3. By introducing the recommended `TileType` constants, implementing their physical/visual properties, distributing them thematically in `placeEnvironmentalProps`, and hooking them into the Y-sorted `DrawSystem`, the team can deliver world fidelity while passing all verification suites.
