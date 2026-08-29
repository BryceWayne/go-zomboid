# Handoff Report: 2D Orthogonal (Top-Down) Engine Architecture & Migration Plan

## 1. Observation

### 1.1 Existing Codebase Architecture & Key Files
The `go-zomboid` project is organized into the following active packages:
- `cmd/game/main.go`: Application entrypoint setting 1280x720 window, calling `assets.Load()`, `game.NewGame()`, and `ebiten.RunGame(g)`.
- `internal/ecs/components.go`: Ark ECS component definitions (`Position`, `Velocity`, `Sprite`, `Collider`, `Player`, `Item`, `Zombie`).
- `internal/assets/assets.go`: Embedded filesystem asset loader (`//go:embed images/*`). Contains asset image handles (`PlayerImage`, `ZombieImage`, `GrassImage`, `WallImage`, `TreeImage`, `LabTilesetImage`, etc.).
- `internal/assets/audio.go`: Procedural PCM white noise and sweep generator for sound effects (`HitSound`, `ShoveSound`).
- `internal/game/world/map.go`: 100x100 tile grid map generator, AABB collision detection, FOV raycasting, procedural buildings (Residential, Grocery, Police, Pharmacy, Warehouse), loot spawns, and zombie spawns. `TileSize = 128`.
- `internal/game/game.go`: Core game state (`Game`), ECS systems (`UpdateSystem`, `DrawSystem`), camera controller (`Camera`), isometric coordinate transforms (`WorldToIso`, `IsoToWorld`, `ScreenToIso`, `ScreenToWorld`), and Bezier curve combat renderer (`DrawAttackSwingArc`).

### 1.2 Coordinate Transformation Functions (Exact Line Numbers & Code)
In `internal/game/game.go`:
- **Lines 818–828**: Current 2:1 dimetric isometric projection:
  ```go
  func WorldToIso(wx, wy float64) (isoX, isoY float64) {
      isoX = wx - wy
      isoY = (wx + wy) / 2.0
      return
  }

  func IsoToWorld(isoX, isoY float64) (wx, wy float64) {
      wx = isoY + isoX/2.0
      wy = isoY - isoX/2.0
      return
  }
  ```
- **Lines 198–207**: Current screen unprojection functions:
  ```go
  func ScreenToIso(screenX, screenY, camX, camY float64) (isoX, isoY float64) {
      isoX = camX + (screenX-640.0)/0.5
      isoY = camY + (screenY-360.0)/0.5
      return
  }

  func ScreenToWorld(screenX, screenY, camX, camY float64) (wx, wy float64) {
      isoX, isoY := ScreenToIso(screenX, screenY, camX, camY)
      return IsoToWorld(isoX, isoY)
  }
  ```
- **Lines 421–431**: Mouse cursor unprojection for player movement and aiming:
  ```go
  camX := 0.0
  camY := 0.0
  if s.camera != nil {
      camX = s.camera.X
      camY = s.camera.Y
  } else {
      camX, camY = WorldToIso(pos.X, pos.Y)
  }
  mx, my := ebiten.CursorPosition()
  mouseWorldX, mouseWorldY := ScreenToWorld(float64(mx), float64(my), camX, camY)
  ```

### 1.3 Camera System
In `internal/game/game.go`:
- **Lines 158–196**:
  ```go
  type Camera struct {
      X, Y             float64
      TargetX, TargetY float64
      LerpFactor       float64
      Initialized      bool
  }
  ```
- Camera is initialized by snapping to player isometric coordinates:
  `Lines 51-53`:
  ```go
  playerIsoX, playerIsoY := WorldToIso(playerStartX, playerStartY)
  g.camera = NewCamera()
  g.camera.Snap(playerIsoX, playerIsoY)
  ```
- Camera is updated every tick in `UpdateSystem.Update()`:
  `Lines 237-240`:
  ```go
  targetIsoX, targetIsoY := WorldToIso(pPos.X, pPos.Y)
  if s.camera != nil {
      s.camera.Update(targetIsoX, targetIsoY)
  }
  ```

### 1.4 DrawSystem Rendering Passes & Anchors
In `internal/game/game.go:830–1282`:
- **Ground Pass (`Lines 880–928`)**:
  Computes `isoX, isoY := WorldToIso(worldX, worldY)`.
  Applies translation offset `op.GeoM.Translate(-128, 0)` (intended for 256x128 diamond tiles).
  Applies camera offset: `op.GeoM.Translate(isoX-camX, isoY-camY)`.
  Applies global 50% scale: `op.GeoM.Scale(0.5, 0.5)`.
  Applies screen centering: `op.GeoM.Translate(640, 360)`.
- **Props / Obstacles Pass (`Lines 938–1021`)**:
  Applies translation offset `op.GeoM.Translate(-imgW/2.0, 128.0-imgH)`.
  Sets `Depth = worldX + worldY`.
- **Entities & Items Passes (`Lines 1024–1143`)**:
  - Items: `op.GeoM.Translate(-32, -32)`, `Depth = iPos.X + iPos.Y`.
  - Entities: `op.GeoM.Translate(-32, -128)`, `Depth = pos.X + pos.Y`.
- **Depth Sorting (`Lines 1180–1186`)**:
  Sorted via `sort.SliceStable` by `Depth` ($wx + wy$).
- **Bezier Combat Swoosh (`Lines 1285–1465`)**:
  `DrawAttackSwingArc` calculates world control points $P_0, P_1, P_2$, projects them to isometric screen coordinates via `WorldToIso(P.x, P.y)`, and strokes vector paths.

### 1.5 External Assets in `context/` and `internal/assets/images/`
By direct image inspection:
- `context/Lab/Inside_C.png`: 768x768 PNG. Standard RPG Maker tileset format (16x16 tiles of 48x48 pixels, or 24x24 tiles of 32x32 pixels, or 48x48 tiles of 16x16 pixels).
- `context/Zombie Apocalypse Tileset/Zombie Apocalypse Tileset Reference.png`: 764x300 PNG.
- `context/Zombie Apocalypse Tileset/Organized separated sprites/`: Contains hundreds of individual 16x16 pixel orthogonal sprites (e.g. `Modular Road` (16x16), `Modular Fences` (16x16), `Modular Bushes` (16x16), `Trees` (16x16), `Pickable Items` (8x8 to 16x16), `Player Character Walking Animation Frames` (14x16), `Skinny Walking Zombie Animation` (10x16)).
- `context/Small Forest/`: Contains pixel art foliage and fences of variable non-uniform dimensions (e.g. `Grass-1.png` is 25x24, `Grass-2.png` is 31x15, `Tree-1-1.png` is 15x19, `Wooden-fence-2.png` is 32x17).
- `internal/assets/images/*.png`: Contains legacy 4x isometric sprites (`grass.png` 256x128 diamond, `wall.png` 256x256, `player.png` 64x128).

### 1.6 Current Test Failure Baseline
Running `CC=gcc go test ./...` currently produces failures in:
1. `internal/assets`: `assets_test.go`, `assets_stress_test.go`, `challenger_stress_test.go`, `empirical_challenger_test.go` fail because `assets.go` was assigned Small Forest image snippets (e.g. 25x24) while test assertions expected 256x128 floor tiles.
2. `internal/game`: `draw_depth_test.go` fails on anchor expectations (`wantTransX = -128, wantTransY = -128`) and `GrassImage` dimensions.
3. `internal/game/world` and `internal/ecs`: 100% PASS.

---

## 2. Logic Chain & Root Cause Analysis

### 2.1 Why Black Gaps Occur (5 Compounding Root Causes)
1. **Isometric Projection Skew**: `WorldToIso` projects a tile at grid index $(x, y)$ to screen position $(x - y, (x + y)/2)$. On an orthogonal screen, adjacent square tiles must be drawn at $(x \cdot S, y \cdot S)$. Projecting orthogonal square sprites onto 45-degree diamond axes leaves triangular void regions between cells.
2. **Tile Size Spacing Disconnect**: `world.TileSize` is $128$. Tiles are positioned at Cartesian intervals of $128$ units ($0, 128, 256, \dots$). When an unscaled $25 \times 24$ or $16 \times 16$ sprite is rendered into a $128 \times 128$ cell, only 12–25 pixels are covered, leaving over $100$ pixels of empty black space per tile.
3. **Anchor Offsets (`Translate(-128, 0)`)**: Isometric diamond rendering requires offsetting by half the tile width $(-128, 0)$ because the diamond apex is centered at the top. On a top-down orthogonal grid, square tiles must be anchored at their top-left corner $(0, 0)$. Applying $(-128, 0)$ shifts the sprite completely out of its grid cell.
4. **Lack of Tileset Sub-Image Slicing / Scaling**: Multi-tile atlas sheets (like `Inside_C.png` or `Zombie Apocalypse Tileset Reference.png`) were loaded as single images rather than sliced into discrete uniform tiles (`SubImage(image.Rect(tx*grid, ty*grid, (tx+1)*grid, (ty+1)*grid))`) or scaled to match `world.TileSize`.
5. **Camera Zoom & Viewport Alignment**: Camera tracking was operating in Isometric space (`camX, camY = WorldToIso(px, py)`). When drawing orthogonally, the camera must track the player's Cartesian World coordinates $(px, py)$ directly.

### 2.2 Mathematical Structure of Seamless 2D Orthogonal (Top-Down) Rendering

#### 1. Discrete Tile Grid to World Cartesian Coordinates:
For a grid cell $(tx, ty)$ where $tx \in [0, \text{Width}-1]$ and $ty \in [0, \text{Height}-1]$:
$$\text{worldX} = tx \cdot \text{TileSize}$$
$$\text{worldY} = ty \cdot \text{TileSize}$$

#### 2. World Cartesian to Screen Viewport Coordinates:
Let $(wx, wy)$ be world coordinates, $(camX, camY)$ be the camera's Cartesian world position, $Z$ be the zoom / render scale factor, and $(W_{screen}, H_{screen}) = (1280, 720)$:
$$\text{screenX} = (wx - camX) \cdot Z + \frac{W_{screen}}{2} = (wx - camX) \cdot Z + 640.0$$
$$\text{screenY} = (wy - camY) \cdot Z + \frac{H_{screen}}{2} = (wy - camY) \cdot Z + 360.0$$

#### 3. Mouse Cursor Unprojection (Screen to World):
Let $(mx, my)$ be the mouse cursor pixel position in the window:
$$\text{mouseWorldX} = camX + \frac{mx - 640.0}{Z}$$
$$\text{mouseWorldY} = camY + \frac{my - 360.0}{Z}$$

#### 4. Seamless Tile Adjacency Proof (Zero Black Gaps):
For two adjacent horizontal tiles $(tx, ty)$ and $(tx+1, ty)$ with width $W = \text{TileSize}$:
- Right boundary of tile $tx$:
  $$\text{Right}(tx) = (tx \cdot \text{TileSize} + \text{TileSize} - camX) \cdot Z + 640.0 = ((tx+1) \cdot \text{TileSize} - camX) \cdot Z + 640.0$$
- Left boundary of tile $tx+1$:
  $$\text{Left}(tx+1) = ((tx+1) \cdot \text{TileSize} - camX) \cdot Z + 640.0$$
- Since $\text{Right}(tx) \equiv \text{Left}(tx+1)$ exactly, there is zero space between any two tiles across the entire grid.

#### 5. Orthogonal Y-Sorting (Depth):
In a top-down orthogonal view with slight vertical perspective for standing objects:
$$\text{Depth} = \text{pos.Y} \quad (\text{or } \text{worldY} + \text{TileSize})$$
Objects with higher $Y$ (further south) are rendered in front of objects with lower $Y$ (further north).

#### 6. Bezier Combat Swoosh in Orthogonal Space:
Control points $P_0, P_1, P_2$ are formulated directly in Cartesian coordinates $(x, y)$:
- $P_0 = (P_x + R_{in} \cos(\theta - \Delta\theta/2), P_y + R_{in} \sin(\theta - \Delta\theta/2))$
- $P_1 = (P_x + R_{apex} \cos(\theta), P_y + R_{apex} \sin(\theta))$
- $P_2 = (P_x + R_{out} \cos(\theta + \Delta\theta/2), P_y + R_{out} \sin(\theta + \Delta\theta/2))$
Direct projection to screen:
$$\text{screen}_i = (P_{i,x} - camX) \cdot Z + 640.0, \quad (P_{i,y} - camY) \cdot Z + 360.0$$
No isometric skewing is applied.

---

## 3. Complete Transition Specification (Files, Data Structures, Functions, and Formulas to Rewrite)

### 3.1 Mathematical & Coordinate Functions (`internal/game/game.go`)

| Current Function | Current Formula | New Orthogonal Formula | Rationale |
|---|---|---|---|
| `WorldToIso(wx, wy)` | `isoX = wx - wy`<br>`isoY = (wx + wy)/2.0` | `isoX = wx`<br>`isoY = wy` | Retains signature for backward compatibility or replace with `WorldToOrtho` identity mapping. |
| `IsoToWorld(isoX, isoY)` | `wx = isoY + isoX/2.0`<br>`wy = isoY - isoX/2.0` | `wx = isoX`<br>`wy = isoY` | Bijective identity inverse. |
| `ScreenToIso(sx, sy, camX, camY)` | `isoX = camX + (sx-640)/0.5`<br>`isoY = camY + (sy-360)/0.5` | `isoX = camX + (sx-640)/zoom`<br>`isoY = camY + (sy-360)/zoom` | Direct Cartesian viewport unprojection. |
| `ScreenToWorld(sx, sy, camX, camY)` | Calls `ScreenToIso` then `IsoToWorld` | `wx = camX + (sx-640)/zoom`<br>`wy = camY + (sy-360)/zoom` | Exact mouse-to-world mapping. |

### 3.2 Camera System (`internal/game/game.go`)
- **Struct Fields**:
  `Camera.X, Camera.Y` represent the camera's Cartesian World position $(wx, wy)$ (not isometric coordinates).
- **`Camera.Snap(targetX, targetY)`**: Snaps directly to target Cartesian world position $(wx, wy)$.
- **`Camera.Update(targetX, targetY)`**: Exponential lerping towards target $(wx, wy)$:
  $$\Delta X = \text{TargetX} - X, \quad \Delta Y = \text{TargetY} - Y$$
  $$X \mathrel{+}= \Delta X \cdot \text{LerpFactor}, \quad Y \mathrel{+}= \Delta Y \cdot \text{LerpFactor}$$
- **`Game.Reset()`**: Snaps camera directly to `(gameMap.PlayerSpawn.X, gameMap.PlayerSpawn.Y)`.
- **`UpdateSystem.Update()`**: Updates camera target to player's `(pPos.X, pPos.Y)`.

### 3.3 DrawSystem (`internal/game/game.go`)

#### 1. Ground Pass:
- Remove `WorldToIso` call and `op.GeoM.Translate(-128, 0)`.
- For each tile $(x, y)$:
  ```go
  worldX := float64(x * world.TileSize)
  worldY := float64(y * world.TileSize)
  screenX := (worldX - camX) * zoom + 640.0
  screenY := (worldY - camY) * zoom + 360.0

  op := &ebiten.DrawImageOptions{}
  // Scale texture to fill exactly TileSize * zoom
  scaleX := (float64(world.TileSize) / float64(imgW)) * zoom
  scaleY := (float64(world.TileSize) / float64(imgH)) * zoom
  op.GeoM.Scale(scaleX, scaleY)
  op.GeoM.Translate(screenX, screenY)
  ```
- Memory tint and visibility checks remain intact.

#### 2. Obstacles / Props Pass:
- Remove isometric anchoring `Translate(-imgW/2.0, 128.0-imgH)`.
- Anchor standing props to bottom-center or top-left of tile:
  ```go
  op.GeoM.Scale(scaleX, scaleY)
  op.GeoM.Translate(screenX, screenY)
  sprites = append(sprites, Renderable{
      Image: img,
      Depth: worldY + float64(world.TileSize), // Top-down Y-sorting
      Op: op,
  })
  ```

#### 3. Entities & Items Pass:
- Entities (Player, Zombies) positioned at $(pos.X, pos.Y)$:
  ```go
  screenX := (pos.X - camX) * zoom + 640.0
  screenY := (pos.Y - camY) * zoom + 360.0
  op.GeoM.Scale(entScaleX, entScaleY)
  op.GeoM.Translate(screenX - float64(entW)/2.0 * entScaleX, screenY - float64(entH)/2.0 * entScaleY)
  sprites = append(sprites, Renderable{
      Image: img,
      Depth: pos.Y, // Y-sorting
      Op: op,
  })
  ```
- Items positioned at $(iPos.X, iPos.Y)$:
  `Depth = iPos.Y`.

#### 4. Bezier Combat Swoosh:
- In `DrawAttackSwingArc`:
  Remove `WorldToIso(wp0X, wp0Y)` etc.
  Directly convert control points to screen coordinates:
  ```go
  screen0X := float32((wp0X - camX) * zoom + 640.0)
  screen0Y := float32((wp0Y - camY) * zoom + 360.0)
  screen1X := float32((wp1X - camX) * zoom + 640.0)
  screen1Y := float32((wp1Y - camY) * zoom + 360.0)
  screen2X := float32((wp2X - camX) * zoom + 640.0)
  screen2Y := float32((wp2Y - camY) * zoom + 360.0)
  ```
  Path creation and `vector.StrokePath` remain unchanged.

### 3.4 World & Map Engine (`internal/game/world/map.go`)
- `TileSize`: Standardize `TileSize` to match asset design (e.g. `32`, `48`, `64`, or `128` with uniform asset scaling).
- Tile types, rooms, building generation, and AABB collision logic:
  `IsColliding(rectX, rectY, rectW, rectH)` already uses Cartesian grid math (`rectX / TileSize`), which natively matches 2D orthogonal space!
- `CalculateFOV(playerX, playerY, radius)`: Raycasting natively operates in 2D Cartesian radial coordinates ($dirX = \cos\theta, dirY = \sin\theta$). Fully compatible.

### 3.5 Asset Pipeline (`internal/assets/assets.go`)
- Assign clean 2D orthogonal tiles / textures to exported pointers (`GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`, `WallImage`, `TreeImage`, etc.).
- Provide sub-image slicing helper function for multi-tile tilesets (e.g. `LabTilesetImage.SubImage(...)`).

### 3.6 Test Suites Requiring Refactoring for 2D Orthogonal Math
The following test files in `internal/game/` and `internal/assets/` contain isometric-specific assertions that must be updated to orthogonal math:
1. `internal/game/game_test.go`:
   - `TestWorldToIso`: Update expected outputs from diamond $(32, 16)$ to orthogonal $(32, 0)$ or identity.
2. `internal/game/camera_test.go`:
   - `TestCamera_ScreenToIsoAndScreenToWorldRoundtrip`: Update projection formulas to orthogonal.
   - `TestCamera_ViewportCornersUnprojection`: Update expected corner delta coordinates.
   - `TestCamera_TileClickMovementTargetingAccuracy`: Update screen tile coordinate calculations.
3. `internal/game/draw_depth_test.go`:
   - `TestDrawSystem_SpriteGeometricAnchors`: Update expected sprite translation anchors from `(-128, -128)` to top-down anchors.
   - `TestDrawSystem_DepthSortingOrdering`: Update monotonicity check to test Y-depth ($wy$).
4. `internal/game/bezier_combat_test.go`:
   - `TestBezier_AxeControlPointsCalculation`: Update screen apex calculation from isometric $(140, 270)$ to orthogonal $(340, 200)$.
5. `internal/game/camera_empirical_challenger_test.go`:
   - Viewport corner distance and round-trip fuzzing tests: update formulas from $\text{sqrt}(1856000)$ to orthogonal corner Euclidean distances.
6. `internal/assets/assets_test.go`, `challenger_stress_test.go`, `empirical_challenger_test.go`:
   - Update expected texture dimensions to match the assigned 2D orthogonal sprites / sliced tiles.

---

## 4. Caveats & Architectural Considerations

1. **Tile Resolution & Zoom Factor**:
   - `Zombie Apocalypse Tileset` sprites are natively $16 \times 16$ pixels.
   - `Inside_C.png` (Lab tileset) is $768 \times 768$ (composed of $48 \times 48$ or $32 \times 32$ tiles).
   - If `world.TileSize` is set to $32$ (or $64$), sprites can either be drawn with integer scaling (e.g. $2\times$ or $4\times$ pixel art scaling) or the global rendering zoom can be set so that pixel art remains sharp and crisp.
2. **Texture Filtering**:
   - To preserve clean pixel art when scaling $16\times 16$ or $48\times 48$ tiles, ensure `ebiten.FilterNearest` (or default draw options) is maintained to prevent blurry bilinear filtering artifacts.
3. **Seamless Integration with Requirement 2 (Dungeon Master System)**:
   - Dynamic zombie wave spawning can inject new entities via `zombieMap.NewEntity(...)` using `gameMap.ZombieSpawns` or edge-of-map non-solid tiles.
   - Randomized loot drops can instantiate `itemMap.NewEntity(...)` dynamically across explored rooms.
   - Day/night ambient lighting in `DrawSystem` already uses `g.timeOfDay` and alpha overlays, which can modulate enemy aggression (`zombie.Speed`, detection radii) at night.

---

## 5. Conclusion

The transition from Isometric to 2D Orthogonal (top-down) grid requires:
1. Simplifying coordinate transforms (`WorldToIso`, `IsoToWorld`) from 2:1 dimetric skewing to direct Cartesian mapping $(wx, wy)$.
2. Centering the camera and mouse unprojection directly on Cartesian world space.
3. Rewriting `DrawSystem` ground pass to render rectangular grid cells $[x \cdot \text{TileSize}, y \cdot \text{TileSize}]$ with top-left origin $(0, 0)$ and proper tile scaling to guarantee **zero black gaps**.
4. Switching depth sorting from isometric diagonal $(wx + wy)$ to vertical top-down Y-sorting ($wy$).
5. Projecting Bezier combat swoosh arcs directly into orthogonal screen space.
6. Updating the asset loader and test suite assertions to match the orthogonal specifications.

All required formulas, file paths, line numbers, and data structures are fully mapped above, providing a complete, self-contained blueprint for the Technical Director and implementation team.

---

## 6. Verification Method

To verify the overhaul once implemented:
1. **Automated Test Suite**:
   ```bash
   CC=gcc go test ./...
   ```
   All tests in `internal/assets`, `internal/ecs`, `internal/game`, and `internal/game/world` must pass with zero failures.

2. **Visual Seamless Rendering & Gameplay Check**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   Verify:
   - World renders in a top-down orthogonal perspective.
   - Adjacent ground tiles seamlessly meet with zero black lines or voids.
   - Left-clicking unprojects accurately to target tiles and player moves to destination.
   - Right-clicking / attacking triggers Bezier swoosh trails aligned with player facing angle.
   - Day/night ambient lighting cycles over time.
