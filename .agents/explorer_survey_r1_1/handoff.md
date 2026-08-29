# Handoff Report: Explorer 1 — Requirements R1 (Tile Rendering Upgrade & Autotiling)

## 1. Observation

### Codebase Inspection & Asset Audit
1. **Coordinate System and Math** (`internal/game/game.go:901-907`, `internal/game/game.go:248-258`):
   - `WorldToIso(wx, wy float64) (isoX, isoY float64)` and `IsoToWorld` return identity `(wx, wy)` for 2D orthogonal Cartesian coordinates.
   - `WorldToScreen(wx, wy, camX, camY)` projects Cartesian world coordinates to the 1280x720 screen surface with `DefaultZoom = 0.5`:
     ```go
     screenX = (wx-camX)*DefaultZoom + 640.0
     screenY = (wy-camY)*DefaultZoom + 360.0
     ```
   - All coordinate conversions are 1:1 bijective Cartesian projections.

2. **Map Model and Tile Generation** (`internal/game/world/map.go:8-36`, `internal/game/world/map.go:194-370`):
   - `TileType` enums: `TileGrass` (0), `TileWall` (1), `TileDirt` (2), `TileWoodFloor` (3), `TileTree` (4), `TileAsphalt` (5), `TileConcrete` (6), `TileTileFloor` (7), `TileFence` (8), `TileDebris` (9), `TileTent` (10), `TileElevationBlock` (11), `TileRamp` (12), `TileStump` (13), `TileMushroom` (14), `TileSign` (15), `TileBench` (16), `TileChest` (17), `TileSculpture` (18), `TileBush` (19), `TileFlower` (20), `TileStone` (21).
   - Grid cell dimension: `TileSize = 128`.
   - Generation initializes the 100x100 grid to `TileGrass` surrounded by boundary `TileWall` (lines 207-215).
   - Procedural generation overlays roads (`TileAsphalt` flanked by `TileConcrete` sidewalks), dirt paths (`TileDirt`), multi-room buildings (`TileWall` with `TileWoodFloor`, `TileTileFloor`, and `TileConcrete`), fenced yards (`TileFence`), and props (`TileTree`, `TileBush`, `TileStone`, `TileChest`, etc.).

3. **Current Ground Tile Rendering Loop** (`internal/game/game.go:958-1022`):
   - In `DrawSystem.Draw`:
     ```go
     for y := 0; y < s.gameMap.Height; y++ {
         for x := 0; x < s.gameMap.Width; x++ {
             t := s.gameMap.GetTile(x, y)
             if t == world.TileWall {
                 continue
             }
             // ... selects img pointer based on TileType ...
             scaleX := (float64(world.TileSize) / imgW) * DefaultZoom
             scaleY := (float64(world.TileSize) / imgH) * DefaultZoom
             screenX, screenY := WorldToScreen(worldX, worldY, camX, camY)

             op := &ebiten.DrawImageOptions{}
             op.GeoM.Scale(scaleX, scaleY)
             op.GeoM.Translate(screenX, screenY)
             screen.DrawImage(img, op)
         }
     }
     ```
   - Each tile is rendered as a standalone 128x128 square drawn at `(x*128, y*128)` transformed to screen coordinates.
   - There is zero checking of neighboring tiles (no cardinal or diagonal neighbor query).

4. **Asset Texture Inspection** (`internal/assets/images/`):
   - Pixel data analysis of base floor images:
     - `grass.png`: (128x128), RGB, single uniform color `[115, 133, 37]` (flat olive green).
     - `dirt.png`: (128x128), RGB, uniform solid color `[125, 100, 64]` (flat brown).
     - `wood.png`, `asphalt.png`, `concrete.png`: (128x128), RGB, flat solid color fills.
     - `wall.png`: (256x256), RGBA obstacle sprite drawn as a standalone box.
   - External tilesets present in repository (`internal/assets/images/` and `context/`):
     - `Small Forest/Ground tileset/Bright-grass-tileset.png` (365x331 RGBA)
     - `Small Forest/Ground tileset/Dark-grass-tileset.png` (365x331 RGBA)
     - `Small Forest/Ground tileset/Earth-tileset.png` (206x92 RGBA)
     - `Small Forest/Ground tileset/Stone-path-tileset-horizontal.png` (182x37 RGBA)
     - `Small Forest/Ground tileset/Stone-path-tileset-vertical.png` (37x182 RGBA)
     - `Lab/Inside_C.png` (768x768 RGBA, 16x16 grid of 48x48 tiles)
     - `Zombie Apocalypse Tileset/Organized separated sprites/Modular Terrain Path` (12 16x16 tiles)
     - `Zombie Apocalypse Tileset/Organized separated sprites/Modular Road` (26 16x16 tiles)
     - `Zombie Apocalypse Tileset/Organized separated sprites/Modular Fences` (10 fence tiles)

5. **Test Status and Build State**:
   - `CC=gcc go test ./...` exits with code 0 (all existing asset, ECS, combat, camera, DM, and world tests pass).
   - `CC=gcc go build -o bin/game ./cmd/game` compiles cleanly with zero errors.

---

## 2. Logic Chain

1. **Root Cause of Harsh Square Borders**:
   - In commit `c396496`, the previous 2.5D isometric diamond tiles (which had transparent corners) were replaced with solid 128x128 square color fills to eliminate black void gaps on the new orthogonal grid.
   - Because each tile is rendered as an isolated 128x128 square without neighbor querying, any boundary between two different terrain types (e.g. `TileDirt` path cutting through `TileGrass`, or `TileConcrete` sidewalk bordering `TileAsphalt` road) results in a hard, pixel-straight 90-degree square transition.
   - Furthermore, the base textures are flat solid colors with zero edge transition alpha masks, accentuating the blocky grid boundaries.

2. **Why Isolated Tile Rendering Fails for Walls and Fences**:
   - `TileWall` and `TileFence` are drawn using a single static sprite (`wall.png` / `fence.png`).
   - Walls lack connectivity logic for horizontal segments, vertical segments, corner turns (NW, NE, SW, SE), T-junctions, cross junctions, front facades, and vertical depth drop-shadows onto adjacent floors.
   - Fences lack post-to-post connection autotiling.

3. **Optimal Technical Solution — 2D Autotiling & Transition Blending**:
   - **Autotiling Strategy 1: Terrain Hierarchy & 2x2 Sub-Tile (Blob) Autotiling**:
     - Assign terrain priority layering: Layer 0 (`TileDirt`) $\to$ Layer 1 (`TileGrass`) $\to$ Layer 2 (`TileConcrete`) $\to$ Layer 3 (`TileAsphalt`) $\to$ Layer 4 (`TileWoodFloor`, `TileTileFloor`).
     - Split each 128x128 tile into four 64x64 sub-quadrants (NW, NE, SW, SE).
     - Each quadrant checks 3 adjacent neighbors (Horizontal, Vertical, Diagonal).
     - Resolving the 3 neighbors maps each quadrant to one of 5 canonical sub-tile configurations:
       1. *Full Inner*: Solid terrain.
       2. *Horizontal Edge*: Boundary transition along top or bottom.
       3. *Vertical Edge*: Boundary transition along left or right.
       4. *Outer Corner*: Convex outer corner.
       5. *Inner Corner*: Concave corner notch.
     - With 5 sub-tile shapes per quadrant (20 sub-tiles or 5 mirrored/rotated sub-tiles per terrain pair), 100% seamless, mathematically complete blob autotiling is achieved for any map shape without corner seams.
   - **Autotiling Strategy 2: Stylized Vector Transition Overlay Fringes**:
     - Following `ART_STYLE_GUIDE.md`, render a base substrate (e.g., warm dirt/earth), then draw the primary terrain (e.g., grass with stylized chevron blade accents).
     - For border edges, draw procedural or pre-computed stylized fringe masks (scalloped grass edges, curb bevels, subtle drop shadows) that overlap adjacent lower-priority tiles.
   - **Autotiling Strategy 3: 4-Bit Cardinal Bitmasking for Walls & Fences**:
     - For walls and fences, calculate a 4-bit neighbor mask:
       $$\text{Mask} = (N \ll 0) | (E \ll 1) | (S \ll 2) | (W \ll 3) \in [0..15]$$
     - Distinguish wall top-caps (roofline) from South-facing vertical facades (wall depth facing rooms) with a subtle drop-shadow on the floor below.
     - Seamlessly connect fence posts and wall corners.

---

## 3. Caveats

1. **Test Invariants**:
   - Existing tests (`TestChallenger_All16ObstaclesPropsAnd6FloorsRenderNoPanic`, `TestEmpirical_All10TileTypesGenerated`, `TestOrthogonal_SeamlessTileAdjacency`) assert that all tile types continue to exist and render without panics or coordinate misalignments. The autotiling engine must maintain full backwards compatibility with `world.TileType` enums and `world.TileSize = 128`.
2. **Performance Constraints**:
   - Ground rendering is executed 60 times per second across up to ~400-800 visible tiles within the FOV radius. All sub-tile images, bitmask lookup tables, and `ebiten.DrawImageOptions` must be pre-generated or cached at asset load time (`assets.Load()`) to avoid per-frame heap allocations in the render loop.
3. **No Direct Code Modifications in Survey Phase**:
   - In accordance with the Explorer archetype guidelines, no source code files were modified during this investigation.

---

## 4. Conclusion & Technical Implementation Proposal

### Proposed Architecture & Step-by-Step Implementation Plan

#### Component 1: Autotile Engine (`internal/game/world/autotile.go` or `internal/game/autotile.go`)
- Define bitmask lookup tables and neighbor evaluation helpers:
  - `GetCardinalBitmask4(m *Map, x, y int, matchType TileType) uint8`
  - `GetQuadrantSubtile(m *Map, x, y int, quad Quadrant, primaryType TileType) SubtileState`
- Define terrain priority rules for seamless transitions between:
  - Grass $\leftrightarrow$ Dirt (paths and trails)
  - Asphalt $\leftrightarrow$ Concrete (roads and sidewalks)
  - Concrete $\leftrightarrow$ Grass (sidewalk edges and lawns)
  - Building Floors $\leftrightarrow$ Exterior Grass/Concrete (doorways and perimeters)

#### Component 2: Vector & Tileset Asset Pipeline (`internal/assets/autotile_assets.go`)
- Create/load high-resolution, vector-styled base textures and transition sub-tiles matching `ART_STYLE_GUIDE.md`:
  - `Grass`: Soft vibrant green base with stylized geometric chevron blades.
  - `Dirt`: Warm earth tones with subtle geometric pebble accents.
  - `Concrete`: Clean light gray sidewalk tiles with crisp expansion joint lines.
  - `Asphalt`: Smooth dark slate gray road surface.
  - `WoodFloor` & `TileFloor`: Warm wooden planks and ceramic grid patterns.
  - `Grass-to-Dirt` & `Grass-to-Concrete` transition overlays (5 quadrant sub-tile masks: Full, North/South edge, East/West edge, Outer corner, Inner corner).
  - `Wall` 4-bit autotiling pieces: Horizontal wall, Vertical wall, L-corners (NW, NE, SW, SE), T-junctions, 4-way crosses, wall endcaps, and South-facing front facade drop-shadows.
  - `Fence` 4-bit autotiling pieces: Horizontal rails, Vertical rails, Corner posts, and T-joints.

#### Component 3: DrawSystem Upgrade (`internal/game/game.go`)
- Refactor `DrawSystem.Draw` ground rendering into a 3-layer pipeline:
  1. **Base Ground Pass**: Render base floor texture.
  2. **Transition Overlay Pass**: For any tile with adjacent differing terrain, render the corresponding 4-quadrant transition overlays using the pre-computed bitmask states.
  3. **Walls & Structures Pass**: In the Y-sorted renderable pass, render walls and fences with their connected bitmask sprites and front facade drop shadows.

---

## 5. Verification Method

To independently verify the implementation after code changes:
1. **Compilation Check**:
   ```bash
   CC=gcc go build -o bin/game ./cmd/game
   ```
   Must compile cleanly without errors.
2. **Automated Unit & Stress Test Verification**:
   ```bash
   CC=gcc go test -v ./...
   ```
   All existing and new test suites in `internal/assets`, `internal/ecs`, `internal/game`, and `internal/game/world` must pass with 100% success.
3. **Visual Inspection**:
   Run `CC=gcc go run ./cmd/game` and inspect:
   - Dirt paths running through grass have smooth, stylized transition fringes instead of harsh 90-degree square blocks.
   - Roads and sidewalks connect with clean curbs and transition lines.
   - Walls connect seamlessly at corners, T-junctions, and doorways with top-caps and vertical facades.
   - No black gaps, texture tearing, or frame rate stuttering during camera movement.
