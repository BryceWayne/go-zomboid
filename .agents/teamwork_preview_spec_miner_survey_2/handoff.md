# Handoff Report: Environment & Procedural Town Generation Specification

## 1. Observation

### 1.1 Codebase Structure and Key Files
- **World & Map Package**:
  - `internal/game/world/map.go`: Defines `TileType`, `Map`, `NewMap`, `CalculateFOV`, `GetTile`, `SetTile`, `IsColliding`.
  - `internal/game/world/map_test.go`: Unit tests for map instantiation (`TestNewMap`) and AABB collision detection (`TestIsColliding`).
- **Game Engine & Systems**:
  - `internal/game/game.go`: Contains `Game`, `Reset`, `UpdateSystem` (including `processMovement`, `processItems`, `processZombies`), `DrawSystem` (isometric projection `WorldToIso`, ground tile rendering, depth-sorted vertical object rendering, FOV memory tinting, day/night cycle overlay).
  - `internal/game/game_test.go`: Unit tests for `WorldToIso` projection.
- **ECS Components**:
  - `internal/ecs/components.go`: Defines `Position`, `Velocity`, `Sprite`, `Collider`, `Player`, `Zombie`, `Item`.
- **Assets & Procedural Sprites Generator**:
  - `internal/assets/assets.go`: Embeds `images/*`, exports `PlayerImage`, `ZombieImage`, `RunnerImage`, `GrassImage`, `WallImage`, `DirtImage`, `WoodImage`, `TreeImage`, `WeaponImage`, `FoodImage`, `WaterImage`.
  - `cmd/tools/genassets/main.go`: Generates 2.5D isometric tiles (`generateIsoFloor`, `generateIsoWall`, `generateIsoTree`) and entities/items (`generateEntity`, `generateWeapon`, `generateItem`).

---

### 1.2 Current World & Town Generation Implementation
From `internal/game/world/map.go:8-106`:
- **Data Structures**:
  - `TileType int`: `TileGrass` (0), `TileWall` (1), `TileDirt` (2), `TileWoodFloor` (3), `TileTree` (4).
  - `TileSize = 32`: Grid resolution (1 tile = 32x32 world coordinate units).
  - `Map`:
    - `Width, Height int`: Grid dimensions (default 100x100 = 3200x3200 world units).
    - `Tiles []TileType`: 1D slice of size `Width * Height`, indexed by `y*Width + x`.
    - `Visible []bool`: Visibility mask calculated each frame from player FOV.
    - `Explored []bool`: Fog of war persistent memory mask.
- **Map Generation Procedure (`NewMap(width, height)`)**:
  1. *Perimeter & Base Fill* (lines 38-47): Fills outer boundary (`x=0`, `x=width-1`, `y=0`, `y=height-1`) with `TileWall`, interior with `TileGrass`.
  2. *Road Generation* (lines 49-59): Places a static crossroad of `TileDirt` (3 tiles wide):
     - Vertical road at columns `width/2 - 1`, `width/2`, `width/2 + 1` across all `y`.
     - Horizontal road at rows `height/2 - 1`, `height/2`, `height/2 + 1` across all `x`.
  3. *House Generation (`buildHouse(hx, hy, hw, hh)`)* (lines 61-82):
     - Fills perimeter `[hx, hx+hw-1] x [hy, hy+hh-1]` with `TileWall`.
     - Places a single south door at `(hx + hw/2, hy + hh - 1)` with `TileWoodFloor`.
     - Fills interior with `TileWoodFloor`.
     - Hardcoded 7 houses placed in the 4 quadrants:
       - `(10, 10, 10, 8)`, `(30, 15, 8, 12)`, `(60, 12, 12, 10)`, `(75, 30, 10, 10)`
       - `(15, 60, 12, 8)`, `(35, 70, 10, 10)`, `(70, 65, 15, 12)`
  4. *Vegetation Generation* (lines 94-104): Spawns 150 `TileTree` instances at uniform random coordinates `(2 + rand.Intn(width-4), 2 + rand.Intn(height-4))` on `TileGrass`.

---

### 1.3 Loot and Zombie Spawning
From `internal/game/game.go:44-115`:
- **Player Spawn**: Hardcoded at center `(50.0 * TileSize, 50.0 * TileSize) = (1600.0, 1600.0)`.
- **Initial Loot**: 3 guaranteed starting items near player:
  - `weapon` at `(playerStartX - 200, playerStartY)`
  - `food` at `(playerStartX + 180, playerStartY + 100)`
  - `water` at `(playerStartX + 200, playerStartY - 50)`
- **Random Loot**: 20 items (5 weapons, 8 food, 7 water) placed at uniform random coordinates `(100 + rand.Intn(3000), 100 + rand.Intn(3000))`. No collision or room validation is performed.
- **Zombie Spawns**: 150 zombies spawned uniformly at `(100 + rand.Intn(3000), 100 + rand.Intn(3000))` until Euclidean distance to player > 300. 20% chance of runner variant (`speed 2.2-2.6` vs normal `1.0-1.5`). No collision check is performed against walls or trees.

---

### 1.4 Collision Map System
From `internal/game/world/map.go:168-183`:
- `IsColliding(rectX, rectY, rectW, rectH float64) bool`:
  - Converts AABB bounding box to tile bounds: `minTileX = int(rectX)/TileSize`, `maxTileX = int(rectX+rectW)/TileSize`, `minTileY = int(rectY)/TileSize`, `maxTileY = int(rectY+rectH)/TileSize`.
  - Scans all covered tiles; returns `true` if any tile is `TileWall` or `TileTree`.
  - Boundary safety: `GetTile(x, y)` returns `TileWall` when out of bounds, preventing entity escape.
- Movement resolution (`internal/game/game.go:514-530`):
  - Evaluates X and Y velocity components separately against `IsColliding`, providing smooth wall sliding.

---

### 1.5 Field of View (FOV) and Visibility
From `internal/game/world/map.go:108-150`:
- `CalculateFOV(playerX, playerY float64, radiusTiles int)`:
  - Resets `m.Visible` slice to `false`.
  - Marks player's current tile as `Visible` and `Explored`.
  - Casts `rays = radiusTiles * 8` (120 rays for 15-tile radius) in 360 degrees.
  - Steps along ray in unit increments; marks intersected tiles as `Visible` and `Explored`.
  - Ray terminates upon encountering `TileWall`.
- Rendering integration (`internal/game/game.go:588-750`):
  - View distance cutoff: `visionRadius = 250.0` pixels from player.
  - Explored but non-visible tiles render with darkened memory tint: `ColorScale.Scale(0.2, 0.2, 0.3, 1)`.
  - Unexplored tiles are omitted.
  - Zombies and items in non-visible tiles (`!Visible`) are culled from rendering.

---

### 1.6 Isometric Projection and Rendering Pipeline
From `internal/game/game.go:546-829`:
- **Coordinate Transformation**:
  $$\text{isoX} = wx - wy, \quad \text{isoY} = \frac{wx + wy}{2.0}$$
- **Camera Offset**:
  $$\text{camX} = \text{playerIsoX} - 400, \quad \text{camY} = \text{playerIsoY} - 300$$
- **Render Passes**:
  1. *Ground Pass* (flat tiles):
     - `TileGrass`, `TileTree` $\rightarrow$ `assets.GrassImage` ($64 \times 32$ diamond)
     - `TileDirt` $\rightarrow$ `assets.DirtImage` ($64 \times 32$ diamond)
     - `TileWoodFloor` $\rightarrow$ `assets.WoodImage` ($64 \times 32$ diamond)
     - Drawn directly to screen at `drawX = isoX - 32 - camX`, `drawY = isoY - camY`.
  2. *Depth-Sorted Sprite Pass* (vertical obstacles, items, entities):
     - Walls: `WallImage` ($64 \times 64$), drawn at `(isoX - 32 - camX, isoY - 32 - camY)`, `Depth = worldX + worldY`.
     - Trees: `TreeImage` ($64 \times 64$), drawn at `(isoX - 32 - camX, isoY - 32 - camY)`, `Depth = worldX + worldY`.
     - Items: $16 \times 16$ sprites, drawn at `(isoX - 8 - camX, isoY - 8 - camY)`, `Depth = itemX + itemY`.
     - Entities: $16 \times 32$ sprites, drawn at `(isoX - 8 - camX, isoY - 32 - camY)`, `Depth = entX + entY`.
     - Facing indicator: scaled sprite drawn at `(targetX + targetY)`.
     - Sorted ascending by `Depth` via `sort.SliceStable`.
  3. *Lighting Overlay*:
     - Ambient darkness rect: $\alpha = 0.45 + 0.45 \cos((\text{timeOfDay}/24.0) \times 2\pi)$.
  4. *HUD / UI Pass*:
     - Health, Hunger, Thirst bars, weapon durability, inventory slots (1-9), game over overlay.

---

## 2. Logic Chain & Analysis

### 2.1 Limitations of Current Town Generation
1. **Lack of Procedural Variety**:
   - The map generator uses 7 hardcoded house positions and 1 static crossroad. Every new game has the exact same layout.
2. **Monolithic Single-Room Buildings**:
   - Houses are plain hollow boxes with a single bottom door and no interior partitions, windows, or varied room types.
3. **No Building Archetypes or Thematic Zoning**:
   - No distinction between residential homes, grocery stores, pharmacies, police stations, gun shops, or industrial warehouses.
4. **Tile Diversity Constraints**:
   - Only 5 tile types exist (`TileGrass`, `TileWall`, `TileDirt`, `TileWoodFloor`, `TileTree`). Missing paved roads (asphalt, concrete, sidewalks), water bodies (ponds/rivers), fences, debris/rubble, and interior floorings (tiles/linoleum).
5. **Decoupled Entity Spawning**:
   - Items and zombies are scattered randomly by coordinates `rand.Intn(3000)` without collision checking, causing items and zombies to spawn inside solid walls or trees.
   - Loot has no contextual placement (e.g. food in kitchens/supermarkets, weapons in police stations/gun shops).

---

### 2.2 Architectural Integration Points for Town Generation Expansion

```
+--------------------------------------------------------------------------------+
|                          Procedural Town Generator                             |
|                                                                                |
|  1. District / Zoning Planner (Commercial, Residential, Industrial, Parks)    |
|  2. Road Network Generator (Avenues, Neighborhood Streets, Alleys, Sidewalks)  |
|  3. Lot Subdivider (Parcel allocation along street frontages)                  |
|  4. Building Synthesizer (Archetypes: House, Store, Police, Pharmacy, Storage) |
|  5. Interior Floorplanner (Multi-room BSP: Kitchen, Bedroom, Storage, Aisle)  |
|  6. Exterior Props / Landmarks (Yards, Fences, Driveways, Debris, Vegetation)  |
|  7. Thematic Spawn Extractor (Safe Player Start, Zone Loot, Horde Placements)  |
+--------------------------------------------------------------------------------+
                                       |
                   +-------------------+-------------------+
                   |                                       |
                   v                                       v
      +-------------------------+             +-------------------------+
      |        world.Map        |             |   Town Spawn Metadata   |
      | - Tiles []TileType      |             | - Player Start Pos      |
      | - IsSolid() Collision   |             | - Contextual Loot Spawns|
      | - BlocksVision() FOV    |             | - Zombie Horde Spawns   |
      +-------------------------+             +-------------------------+
                   |                                       |
                   |                                       v
                   |                          +-------------------------+
                   |                          |   game.Reset() ECS      |
                   |                          | - Spawn Player in House |
                   |                          | - Spawn Loot in Rooms   |
                   |                          | - Spawn Zombies in Town |
                   |                          +-------------------------+
                   v                                       |
      +-------------------------+                          |
      |    Rendering System     |                          |
      | - Ground Pass (Floors)  |<-------------------------+
      | - Y-Sorted Vertical Pass|
      +-------------------------+
```

1. **Tile System Expansion (`internal/game/world/map.go`)**:
   - Add new `TileType` constants:
     - Floor types: `TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileWater`
     - Obstacle/Prop types: `TileFence`, `TileBush`, `TileDebris`
   - Implement helper methods:
     - `TileType.IsSolid() bool`: returns true for `TileWall`, `TileTree`, `TileFence`, `TileWater`.
     - `TileType.BlocksVision() bool`: returns true for `TileWall` (allows seeing over fences/water).
     - `TileType.IsFloor() bool`: returns true for passable ground tiles.
2. **Building & Room Metadata Schema**:
   - `Building` struct: `Type` (Residential, Grocery, PoliceStation, Pharmacy, Warehouse), `Bounds` (AABB), `Rooms []Room`, `Doors []Point`.
   - `Room` struct: `Type` (LivingRoom, Kitchen, Bedroom, Bathroom, StoreFloor, StorageRoom), `Bounds` (AABB).
   - Allows spawning systems to query rooms by type for realistic loot distribution.
3. **Procedural Town Generator Pipeline**:
   - Implement a modular generator in `internal/game/world/` or `internal/game/world/generator/`.
   - Generates grid of zones, connects them with a road hierarchy (asphalt roads + concrete sidewalks), places diverse building archetypes with multi-room layouts, adds fenced yards and outdoor debris/trees, and outputs both `*Map` and structured spawn lists.
4. **Collision & FOV Integration**:
   - Update `IsColliding` to check `m.GetTile(x, y).IsSolid()`.
   - Update `CalculateFOV` to check `m.GetTile(x, y).BlocksVision()`.
5. **Asset Pipeline Integration (`cmd/tools/genassets` & `internal/assets`)**:
   - Add procedural asset generators for new isometric tiles in `cmd/tools/genassets/main.go`:
     - Floor tiles ($64 \times 32$): `asphalt.png`, `concrete.png`, `tile_floor.png`, `water.png`
     - Vertical tiles ($64 \times 64$): `fence.png`, `bush.png`, `debris.png`
   - Update `internal/assets/assets.go` to load and expose the new images.
   - Update `internal/game/game.go` `DrawSystem` to render new ground tiles and depth-sort vertical tiles.

---

## 3. Features Discovered

| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Map Storage | `Map` Grid Representation | Flattened 1D slice storing tile types, visibility, and exploration state | `width, height int` | `*Map` instance | None (allocates $W \times H$ slices) | `internal/game/world/map.go:22-36` |
| 2 | Map Coordinates | `GetTile` & `SetTile` | Reads and writes tile types with boundary protection | `x, y int`, `TileType` | `TileType` (for get) | Out-of-bounds `GetTile` returns `TileWall`; `SetTile` ignores | `internal/game/world/map.go:152-164` |
| 3 | World Generation | `NewMap` Base Map | Generates perimeter wall, fills grass, static crossroad, 7 houses, 150 trees | `width, height int` | Initialized `*Map` | None | `internal/game/world/map.go:29-106` |
| 4 | Building Template | `buildHouse` Generator | Creates rectangular building with wood floor, wall border, and single south door | `hx, hy, hw, hh int` | Writes tiles to `Map` | Skips tiles out of bounds | `internal/game/world/map.go:61-82` |
| 5 | Collision Detection | `IsColliding` | Checks if AABB bounding box intersects solid tiles (`TileWall`, `TileTree`) | `rectX, rectY, rectW, rectH float64` | `bool` | Out-of-bounds tiles treated as `TileWall` | `internal/game/world/map.go:168-183` |
| 6 | Field of View | `CalculateFOV` | 360-degree raycasting with line-of-sight occlusion by `TileWall` | `playerX, playerY float64, radius int` | Updates `Visible[]`, `Explored[]` | Early return if player out of bounds | `internal/game/world/map.go:108-150` |
| 7 | Projection Math | `WorldToIso` | Transforms 2D world coordinates to 2.5D isometric screen coordinates | `wx, wy float64` | `isoX = wx - wy`, `isoY = (wx+wy)/2` | None | `internal/game/game.go:546-550` |
| 8 | Ground Rendering | Diamond Floor Pass | Renders $64 \times 32$ floor tiles (`TileGrass`, `TileDirt`, `TileWoodFloor`) | `screen *ebiten.Image`, camera pos | Blits floor images | Culls tiles $>250\text{px}$ from player | `internal/game/game.go:588-630` |
| 9 | Vertical Rendering | Y-Sorted Object Pass | Depth-sorts walls, trees, items, and entities by $wx + wy$ | `sprites []Renderable` | Blits sorted images | Culls non-visible items/zombies | `internal/game/game.go:632-828` |
| 10 | Fog of War | Exploration Memory Tint | Applies darkened tint $(0.2, 0.2, 0.3)$ to explored tiles outside current FOV | `Explored[idx] && !Visible[idx]` | ColorScale tint on image | Unexplored tiles omitted | `internal/game/game.go:618, 674` |
| 11 | Lighting | Day-Night Cycle Overlay | Renders ambient darkness rect based on `timeOfDay` cosine curve | `timeOfDay float64` | Screen fill with alpha $0.0-0.90$ | None | `internal/game/game.go:830-836` |
| 12 | Entity Spawning | World Initialization | Spawns Player at center, 3 close items, 20 random items, 150 zombies | `world *arkecs.World` | Entities in ECS | Does not validate collision bounds | `internal/game/game.go:34-115` |
| 13 | Asset Generation | Procedural Tile Generator | Creates PNG assets with noise texturing and shading in `genassets` | `cmd/tools/genassets` CLI | PNGs in `assets/images` | Fails on filesystem permissions | `cmd/tools/genassets/main.go:1-218` |

---

## 4. Edge Cases

| # | Feature | Input | Observed Behavior |
|---|---------|-------|-------------------|
| 1 | Map Boundary Access | `m.GetTile(-1, 50)` or `m.GetTile(100, 100)` | Returns `TileWall`, preventing entities and rays from escaping or crashing. |
| 2 | Out-of-Bounds Tile Writing | `m.SetTile(-5, 200, TileGrass)` | Ignored cleanly without panics or memory corruption. |
| 3 | FOV Player Out of Bounds | `CalculateFOV(-10.0, 5000.0, 15)` | Detects `px < 0 \|\| py >= Height` and immediately returns without updating visibility. |
| 4 | FOV Wall Occlusion | Ray hits `TileWall` at step 3 of 15 | Marks wall tile as `Visible` and `Explored`, then `break` terminates that ray to prevent seeing through walls. |
| 5 | Tree Spawn Filtering | Random coordinates pick a tile containing `TileWall` or `TileDirt` | `m.GetTile(tx, ty) == TileGrass` evaluates to false, skipping tree placement on roads and houses. |
| 6 | AABB Collision Edge Overlap | Box `(60, 60, 10, 10)` overlapping wall at tile `(2,2)` (coords 64-95) | `IsColliding` computes `maxTileX=2, maxTileY=2`, detects `TileWall`, and returns `true`. |
| 7 | Multi-Tile AABB Span | Entity moving across tile boundary spanning 4 tiles | Checks all tiles in `[minTileX..maxTileX] x [minTileY..maxTileY]`; collides if any tile is solid. |
| 8 | Entity Movement Wall Sliding | Moving diagonal with velocity `vel.X != 0, vel.Y != 0` into a vertical wall | X move is blocked, Y move succeeds, allowing smooth wall sliding. |
| 9 | Unchecked Item/Zombie Spawning | Random coordinate lands inside `TileWall` or `TileTree` | Spawns entity inside obstacle without validation; entity becomes trapped in collision. |
| 10 | Fog of War Darkness Culling | Zombie in tile with `Visible[idx] == false` within vision radius | Zombie is omitted from `sprites` render list and remains completely invisible. |
| 11 | Explored Wall/Floor Tinting | Camera near previously explored house while player is outside FOV | House walls and floors render with dim $(0.2, 0.2, 0.3)$ memory tint. |
| 12 | Depth Sorting Parity | Entity at tile center `(32*x+16, 32*y+16)` vs wall at `(x, y)` | Entity depth ($32x+32y+32$) > Wall depth ($32x+32y$); entity correctly renders in front of wall behind it. |
| 13 | Day-Night Midnight Transition | `timeOfDay = 0.0` or `24.0` | `math.Cos(0) = 1.0`, yielding maximum ambient darkness $\alpha = 0.90$. |
| 14 | Day-Night Noon Transition | `timeOfDay = 12.0` | `math.Cos(\pi) = -1.0`, yielding minimum darkness $\alpha = 0.0$. |

---

## 5. Caveats
- No changes to existing code were made during this specification mining pass (strictly read-only).
- Existing game audio requires ALSA/sound context initialization which recovers gracefully in headless environments.
- Coordinate system uses world pixels ($1 \text{ tile} = 32 \text{ px}$), while isometric rendering expands ground diamonds to $64 \times 32$ pixels and wall/tree sprites to $64 \times 64$ pixels.

---

## 6. Conclusion
The current `go-zomboid` world generator is a minimalist prototype with 7 hardcoded houses, a single static crossroad, uniform random tree placement, and unconstrained loot/zombie spawns.

To fulfill the project requirements for expanding procedural town generation:
1. **Procedural Layout Generation**: Replace hardcoded `NewMap` with an algorithmic town generator featuring district zoning (Residential, Commercial, Industrial, Park), a multi-tier road network (asphalt avenues, streets, sidewalks), and lot subdivision.
2. **Multi-Room Building Archetypes**: Create diverse building types (Suburban Houses, Grocery Stores, Police Stations, Pharmacies, Warehouses) with partitioned room layouts (Kitchens, Bedrooms, Bathrooms, Storerooms, Sales Floors).
3. **Environment & Outdoor Landmarks**: Introduce fenced backyards, driveways, street debris/rubble, water ponds, and bush foliage.
4. **Expanded Tile & Asset Palette**: Add new tile types (`TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileWater`, `TileFence`, `TileBush`, `TileDebris`) generated via `cmd/tools/genassets` and rendered in the isometric pipeline.
5. **Contextual Loot & Zombie Spawning**: Connect ECS spawning to room metadata so loot spawns thematically inside containers/rooms and zombies spawn with proper density without getting trapped inside walls.

---

## 7. Verification Method

### 7.1 Automated Test Execution
Run test suite across all packages:
```sh
CC=gcc go test -v ./...
```
Expected output: All unit tests in `internal/game` and `internal/game/world` pass cleanly.

### 7.2 Asset Generation Execution
Run procedural asset generator:
```sh
go run ./cmd/tools/genassets
```
Expected output: Executes without errors and produces all required PNG sprite files in `internal/assets/images/`.

### 7.3 Game Loop Execution
Launch the game engine:
```sh
CC=gcc go run ./cmd/game
```
Expected output: Ebitengine initializes window, loads embedded assets, constructs the world map, and runs the 60 FPS update and render loops without panics.
