# Specification Mining & Architecture Survey: Rendering System, Depth Sorting, Lifecycle & Test Suite

**Surveyor:** `teamwork_preview_spec_miner_survey_3`  
**Date:** 2026-08-29  
**Target Repository:** `/home/bryce/code/go-zomboid`  
**Authoritative Sources Inspected:**
- `/home/bryce/code/go-zomboid/cmd/game/main.go`
- `/home/bryce/code/go-zomboid/cmd/tools/genassets/main.go` & `genassets_test.go`
- `/home/bryce/code/go-zomboid/internal/assets/assets.go` & test suite (`assets_test.go`, `assets_stress_test.go`, `challenger_stress_test.go`, `empirical_challenger_test.go`)
- `/home/bryce/code/go-zomboid/internal/game/game.go` & test suite (14 test files across combat, camera, armor, rendering, simulation)
- `/home/bryce/code/go-zomboid/internal/game/world/map.go` & test suite (`map_test.go`, `world_empirical_stress_test.go`)
- `/home/bryce/code/go-zomboid/context/` external PNG assets
- `/home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md`

---

## Features Discovered

| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Rendering | 2:1 Isometric Projection | Converts 2D world coordinates $(w_x, w_y)$ into 2:1 isometric screen space $(iso_x, iso_y)$ where $iso_x = w_x - w_y$ and $iso_y = (w_x + w_y)/2$. | World coordinates $(w_x, w_y \in \mathbb{R})$ | Isometric coordinates $(iso_x, iso_y)$ | Deterministic linear mapping, no singular points | `internal/game/game.go:818-822` |
| 2 | Rendering | Inverse Isometric Unprojection | Converts isometric coordinates $(iso_x, iso_y)$ back to world coordinates $(w_x, w_y)$ where $w_x = iso_y + iso_x/2$ and $w_y = iso_y - iso_x/2$. | Isometric coordinates $(iso_x, iso_y \in \mathbb{R})$ | World coordinates $(w_x, w_y)$ | Exact inverse of `WorldToIso` | `internal/game/game.go:824-828` |
| 3 | Rendering | Screen-to-World Coordinate Mapping | Transforms mouse screen coordinates $(s_x, s_y)$ through camera offset $(cam_x, cam_y)$ and $0.5\times$ viewport zoom into world coordinates. | Screen $(s_x, s_y)$, Cam $(cam_x, cam_y)$ | World coordinates $(w_x, w_y)$ | Linear transform centered at screen center $(640, 360)$ | `internal/game/game.go:198-208` |
| 4 | Rendering | Ground Diamond Tile Pass | Renders flat ground terrain tiles in a base layer without depth sorting, culling tiles outside player vision radius (2200 px) or unvisited fog-of-war. | Map grid, `Visible[]`, `Explored[]`, Player pos | Drawn ground diamonds on `ebiten.Image` | Walls skipped; non-visible explored tiles tinted with memory shade `(0.2, 0.2, 0.3, 1)` | `internal/game/game.go:879-927` |
| 5 | Rendering | Depth-Sorted Sprite Pass (Y-Sorting) | Collects all vertical obstacles/props, dropped items, character entities (player & zombies), and aim indicators into a `Renderable` slice sorted by `Depth = worldX + worldY`. | Renderable sprites with `Depth` values | Back-to-front depth-ordered drawing | Stable slice sort (`sort.SliceStable`) prevents Z-fighting | `internal/game/game.go:929-1168` |
| 6 | Rendering | Fog of War & Darkness Occlusion | Raycasts FOV (22 tiles) from player; unvisited cells are fully black, previously visited but out-of-sight cells have memory tint, entities/items in darkness are culled. | Player $(x,y)$, Map tile occlusion | `Visible[]` and `Explored[]` arrays | Zombies in unlit tiles are invisible; items in darkness not rendered | `internal/game/world/map.go:907-947`, `internal/game/game.go:897,1019,1079` |
| 7 | Rendering | Dynamic Day-Night Ambient Lighting | Fullscreen ambient lighting overlay modulated sinusoidally by `timeOfDay` ($\alpha = 0.45 + 0.45\cos(\frac{t}{24}2\pi)$). | `timeOfDay \in [0.0, 24.0)` | Fullscreen colored rectangle overlay | Blends darkness at midnight ($\alpha \approx 0.90$) to full sunlight at noon ($\alpha = 0.0$) | `internal/game/game.go:1175-1181` |
| 8 | Rendering | Bezier Curve Combat Swoosh Trails | Renders dynamic quadratic Bezier arcs and radial muzzle rays for melee swings (Axe, Club, Shove) and Shotgun blasts during attack frames (`cooldown \in [17, 30]`). | `playerFacing`, `weaponType`, `attackCooldown` | 2-pass anti-aliased vector stroke paths (outer glow + core) | Alpha quadratically fades over 14 active frames $(30 \to 16)$ | `internal/game/game.go:1267-1447` |
| 9 | Rendering | Screen-Space HUD / UI | Draws 2D vector HUD overlay including Health, Hunger, Thirst, Armor durability bars, weapon hits/ammo counter, 9-slot inventory hotbar, infection warning, and death screen. | Player ECS components | Vector rectangles & debug text at fixed screen coords | Clamps bar widths $\ge 0$ and $\le 200$; handles empty inventory slots gracefully | `internal/game/game.go:1183-1264` |
| 10 | World / Map | Isometric Tile Types & Properties | `TileType` enum defining physical behavior (`IsSolid()`), vision occlusion (`BlocksVision()`), and floor rendering (`IsFloor()`). | `TileType` int | Boolean capabilities & string names | Unknown types return `IsSolid=false`, `BlocksVision=false`, `IsFloor=false`, `String="Unknown"` | `internal/game/world/map.go:8-95` |
| 11 | World / Map | Procedural Town Architecture | Generates $100\times100$ town grid with asphalt road network, concrete sidewalks, boundary walls, 5 building archetypes (Residential, Grocery, Pharmacy, Police, Warehouse), fenced yards, and contextual props. | Grid dimensions `(width, height)` | `*world.Map` with tiles, spawns, buildings | Fallback generator used if $W < 30$ or $H < 30$ | `internal/game/world/map.go:174-352` |
| 12 | World / Map | Safe Contextual Spawning | Places player safely inside residential living room; generates contextual loot based on room type; spawns 140 zombies $\ge 1400$ px away on non-solid tiles. | Building/room metadata, target counts | `PlayerSpawn`, `LootSpawns`, `ZombieSpawns` | Skips solid tiles; retries zombie placement up to $30\times$ target count | `internal/game/world/map.go:328-350, 805-905` |
| 13 | World / Map | AABB Tile Collision Detection | Checks whether a bounding box $[x, x+w] \times [y, y+h]$ in world pixel space intersects any solid tile (`TileType.IsSolid() == true`). | World bounding box $(x, y, w, h)$ | `bool` (true if colliding) | Out-of-bounds bounding boxes return `true` (solid collision) | `internal/game/world/map.go:970-990` |
| 14 | Assets | Static PNG Asset Loading | Embeds images via `go:embed images/*` and decodes PNG bytes into `*ebiten.Image` global handles wrapped in `sync.Once`. | Embedded PNG files | Exported `*ebiten.Image` pointers | `log.Fatalf` on missing or undecodable image | `internal/assets/assets.go:1-108` |
| 15 | Assets | External PNG Context Ingestion | External PNG assets provided in `context/` (`Small Forest`, `Lab`, `Zombie Apocalypse Tileset`) including Benches, Chests, Sculptures, Bushes, Stones, Trees, Fences. | PNG files in `context/` subdirectories | `*ebiten.Image` variables in `internal/assets` | Must replace obsolete procedural generator pipeline | `ORIGINAL_REQUEST.md`, `context/` |
| 16 | Lifecycle | Game Startup & Main Loop | Configures 1280x720 window, loads audio & assets, constructs ECS world & map, binds camera, and runs `ebiten.RunGame`. | None | Running interactive game loop | Fatal error log on `ebiten.RunGame` failure | `cmd/game/main.go:1-23`, `internal/game/game.go:29-34` |
| 17 | Lifecycle | Game Reset & State Recycling | `Reset()` rebuilds ECS world, generates fresh map, initializes player stats (100 HP/Hunger/Thirst), spawns contextual loot & zombies, snaps camera to player. | None (invoked on init or 'R' press when dead) | Fresh game instance state | Preserves audio state while fully resetting ECS entities and map | `internal/game/game.go:36-120` |
| 18 | Camera | Smooth Exponential Lerp Camera | Isometric camera smoothly tracking player target position with exponential lerp factor (0.10) and sub-pixel snapping ($< 0.01$). | Target isometric coords $(iso_x, iso_y)$ | Camera position $(X, Y)$ | Uninitialized camera snaps instantly on first frame | `internal/game/game.go:158-196` |

---

## Edge Cases

| # | Feature | Input / Condition | Observed / Required Behavior |
|---|---------|-------------------|-----------------------------|
| 1 | Depth Sorting | Entity at $(X, Y)$ overlaps Prop/Object at $(X_o, Y_o)$ | When $X + Y < X_o + Y_o$, entity renders behind the prop; when $X + Y > X_o + Y_o$, entity renders in front of prop. Stable sort preserves relative ordering when depth values are identical. |
| 2 | Depth Sorting | Newly introduced props (Bench, Chest, Sculpture) with varying sprite dimensions | Custom sprite dimensions (e.g. 52x37 Bench, 22x21 Chest, 23x31 Sculpture) require anchor adjustments so the sprite bottom aligns with the isometric tile center $(iso_X, iso_Y)$, preventing floating or ground-sinking. Depth remains `worldX + worldY`. |
| 3 | Asset Loading | Embedded image path missing or invalid | `assets.loadEbitenImage` calls `log.Fatalf` terminating the process. All referenced assets must be present in `internal/assets/images/` and embedded via `embed.FS`. |
| 4 | Asset Loading | Asset files containing spaces or special characters in source directory | Source paths in `context/` contain spaces (e.g. `context/Small Forest/Bench and chest/Bench.png`). When ingested into `internal/assets/images/`, filenames should be normalized/flattened or cleanly matched (e.g., `bench.png`, `chest.png`, `sculpture_1.png`). |
| 5 | Fog of War | Tile has `Explored=true` but `Visible=false` | Floor tile and props are drawn with memory tint `op.ColorScale.Scale(0.2, 0.2, 0.3, 1)`. Entities (zombies) and items in non-visible tiles are completely suppressed from rendering. |
| 6 | Fog of War | Tile has `Explored=false` and `Visible=false` (unvisited darkness) | Tile is completely skipped in ground and prop passes; background clear color `color.RGBA{15, 15, 15, 255}` remains visible. |
| 7 | Day/Night Cycle | Extreme hours: Noon ($t=12.0$) vs Midnight ($t=0.0 / 24.0$) | At noon, $\alpha = 0.0 \le 0.05$, lighting overlay is skipped (full brightness). At midnight, $\alpha = 0.90$, rendering semi-transparent deep navy rectangle `RGBA{0, 0, 15, 229}` over all world sprites. UI drawn afterwards remains 100% bright and readable. |
| 8 | Player State Rendering | Player Dead (`Dead=true`) vs Infected (`Infected=true`) vs Armor Equipped | Dead player rendered with dark gray tint `(0.3, 0.3, 0.3, 1)`. Infected player pulses green/red dynamically via `0.5 + 0.5*math.Sin(Health)`. Armored player gains metallic steel-blue tint `(0.75, 0.85, 1.25, 1.0)`. Active attack cooldown $> 20$ flashes white `(2, 2, 2, 1)`. |
| 9 | Map Collision | Non-solid Floor vs Solid Prop Collision | `TileType.IsSolid()` must return `true` for solid obstacles (Wall, Tree, Fence, Debris, Tent, Stump, Sign, ElevationBlock, Bench, Chest, Sculpture) and `false` for walkable floors (Grass, Dirt, WoodFloor, Asphalt, Concrete, TileFloor, Ramp, Flower, Bush). |
| 10 | Vision Occlusion | Solid Obstacle vs Vision Occlusion | `TileType.BlocksVision()` must return `true` ONLY for tall opaque structures (`TileWall`). Obstacles like `TileFence`, `TileTree`, `TileDebris`, `TileBench`, `TileChest`, `TileSculpture` must return `false` so the player can see over them. |
| 11 | Fallback Map | Map dimensions $< 30 \times 30$ | `NewMap()` falls back to `generateSmallFallback()` which creates dirt cross-roads, centers player spawn, and places 3 starter loot items (weapon, food, water). |
| 12 | Headless Test Execution | Running `go test ./...` in headless CI without display / GPU | Ebiten supports headless rendering into `ebiten.NewImage()`. All tests calling `g.Draw(screen)` or `drawSys.Draw(screen, timeOfDay)` execute headlessly without X11/Wayland dependencies. |
| 13 | Asset Retirement | Deletion of `cmd/tools/genassets` | Any test calling `exec.Command("go", "run", "./cmd/tools/genassets")` (such as `TestEmpiricalGenerationDeterminism` in `empirical_challenger_test.go`) will fail once `genassets` is deleted unless retired. Existing tests must reflect native asset loading. |

---

## Comprehensive Rendering System Architecture & Analysis

### 1. Mathematical Foundations & Coordinate Systems
The game uses a **2:1 Isometric Dimetric Projection** system.

```
                  (0, 0)
                   /\
                  /  \
                 /    \
 (-w/2, h/2)    /      \    (w/2, h/2)
                \      /
                 \    /
                  \  /
                   \/
                 (0, h)
```

- **World Space $(w_x, w_y)$**: Continuous Cartesian plane measured in pixels. Each map cell is $TileSize = 128$ pixels wide and high.
- **Grid Cell Coordinates $(t_x, t_y)$**: Integer tile coordinates where $t_x = \lfloor w_x / TileSize \rfloor$ and $t_y = \lfloor w_y / TileSize \rfloor$.
- **Isometric Projection Formulas**:
  $$\begin{aligned}
  iso_x &= w_x - w_y \\
  iso_y &= \frac{w_x + w_y}{2}
  \end{aligned}$$
- **Inverse Projection Formulas**:
  $$\begin{aligned}
  w_x &= iso_y + \frac{iso_x}{2} \\
  w_y &= iso_y - \frac{iso_x}{2}
  \end{aligned}$$
- **Screen & Camera Transformations**:
  At default window resolution $1280 \times 720$ with zoom scale $s = 0.5$:
  $$\begin{aligned}
  screen_x &= (iso_x - cam_x) \times 0.5 + 640.0 \\
  screen_y &= (iso_y - cam_y) \times 0.5 + 360.0
  \end{aligned}$$

### 2. Multi-Pass Rendering Architecture (`DrawSystem.Draw`)
The rendering loop in `internal/game/game.go:830-1264` executes in 6 distinct sequential passes:

```
+-------------------------------------------------------------------------+
| Pass 1: Background Clear (color: #0F0F0F)                               |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Pass 2: Ground Tiles (Grass, Dirt, Wood, Asphalt, Concrete, TileFloor)  |
| - Rendered directly to screen (no depth sort needed)                    |
| - Distance culling: dx*dx + dy*dy <= (2200)^2                           |
| - Fog of war check: Visible[idx] || Explored[idx]                       |
| - Explored memory tint: ColorScale(0.2, 0.2, 0.3, 1)                    |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Pass 3: Depth-Sorted Sprite Pass (Y-Sorting)                            |
| 1. Collect Props/Obstacles (Walls, Trees, Fences, Debris, Benches, etc) |
| 2. Collect Ground Items (Food, Water, Weapons, Armor, Antidote, Ammo)   |
| 3. Collect Character Entities (Player, Standard Zombie, Runner Zombie)  |
| 4. Collect Player Aim/Facing Indicator Sprite                           |
| -> Sort all by Depth = worldX + worldY (Back-to-Front)                  |
| -> Draw sorted sprites to screen                                        |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Pass 4: Vector Combat Trails (Melee Cleave & Shotgun Wavefront Arcs)    |
| - Evaluated during attackCooldown in [17..30]                           |
| - 2-Pass Anti-aliased Quadratic Bezier Strokes (Outer Glow + Core)      |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Pass 5: Day-Night Ambient Darkness Overlay                              |
| - Fullscreen rect modulated by timeOfDay: alpha = 0.45 + 0.45*cos(2pi*t)|
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Pass 6: Screen-Space HUD / UI (Health, Hunger, Thirst, Armor, Hotbar)   |
| - Fixed screen coordinates (1280x720) drawn over all world elements     |
+-------------------------------------------------------------------------+
```

---

## Depth-Sorting & New World Object Ingestion Analysis

### 1. Depth Metric & Isometric Sorting
In 2:1 isometric projection, objects located further up and left on the screen have smaller world coordinate sums $(w_x + w_y)$, while objects further down and right have larger sums $(w_x + w_y)$.
Since $iso_y = (w_x + w_y) / 2$, sorting by $Depth = w_x + w_y$ is mathematically equivalent to sorting by $iso_y$ (pure Y-sorting).

When rendering:
- Ground tiles are drawn in Pass 2 as base terrain.
- All vertical structures, props, items, and entities are added to `sprites []Renderable` with:
  `Depth = worldX + worldY`
- `sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].Depth < sprites[j].Depth })` guarantees that:
  1. Objects with smaller $(w_x + w_y)$ are drawn first (in the background).
  2. Objects with larger $(w_x + w_y)$ are drawn later (in the foreground).
  3. Characters walking behind a prop (e.g. a bench, chest, or sculpture) have $p_x + p_y < prop_x + prop_y$, so the prop is drawn on top of the player.
  4. Characters walking in front of the prop have $p_x + p_y > prop_x + prop_y$, so the player is drawn on top of the prop.

### 2. Survey of External Assets in `context/`
The `context/` directory provides high-resolution external PNG assets across three collections:

#### Collection A: `context/Small Forest/`
- **Bench and Chest**:
  - `Bench.png` ($52 \times 37$ px): Park/yard bench prop.
  - `Chest.png` ($22 \times 21$ px): Wooden treasure/loot container prop.
- **Sculptures**:
  - `Sculpture-1.png` ($23 \times 31$ px): Classical stone statue prop.
  - `Sculture-2.png` ($29 \times 32$ px): Classical ornate stone sculpture prop.
- **Bushes & Stumps**:
  - `Bush-1.png` ($24 \times 18$ px) to `Bush-4.png` ($28 \times 19$ px): Dense foliage bushes.
  - `Stump.png` ($29 \times 19$ px): Tree stump prop.
- **Stones**:
  - `Stone-1.png` ($28 \times 19$ px), `Stone-2.png` ($29 \times 25$ px): Large landscape stones/boulders.
- **Flowers & Grass Tufts**:
  - `Flower-1.png` ($26 \times 25$ px) to `Flower-3.png` ($26 \times 18$ px): Decorative ground flora.
  - `Grass-1.png` ($25 \times 24$ px), `Grass-2.png` ($31 \times 15$ px): Wild grass tufts.
- **Fences**:
  - `Wooden fence/` ($29\times17$ to $32\times17$ px, vertical $15\times36$ px).
  - `Stone fence/` ($32\times22$ px, vertical $13\times32$ px).
  - `Big wooden fence/` ($54\times23$ to $64\times23$ px, vertical $14\times44$ px).
- **Trees**:
  - `Tree-1/` ($15\times19$ to $37\times50$ px, 4 growth stages).
  - `Tree-2/` ($15\times18$ to $36\times50$ px, 4 growth stages).
  - `Tree-3/` ($15\times19$ to $55\times67$ px, 4 growth stages).
- **Ground Tilesets**:
  - `Bright-grass-tileset.png`, `Dark-grass-tileset.png`, `Earth-tileset.png`, `Stone-path-tileset-horizontal.png`, `Stone-path-tileset-vertical.png`.

#### Collection B: `context/Lab/`
- `Inside_C.png` ($768 \times 768$ px): Complete indoor tileset with laboratory equipment, consoles, containers, and partition walls.

#### Collection C: `context/Zombie Apocalypse Tileset/`
- Reference atlas ($436 \times 501$ px) and PSD source.
- Over 50 subcategories of separated sprites (Characters, Blood animations, Weapons, Vehicles, Urban Props, Fences, Signs, etc.).

### 3. Mapping New Objects into `TileType` and Game Logic
To fulfill Requirement R3 ("Infer and Implement New Logic: Analyze the imported assets (e.g., Benches, Chests, Sculptures) and automatically infer their mapping into the game world"):

| Inferred TileType | Source Context PNG | Category | `IsSolid()` | `BlocksVision()` | `IsFloor()` | Base Ground Pass | Recommended Map Placement |
|-------------------|--------------------|----------|-------------|------------------|-------------|------------------|---------------------------|
| `TileBench` | `Bench.png` ($52\times37$) | Obstacle/Prop | `true` | `false` | `false` | `GrassImage` / `ConcreteImage` | Parks, Sidewalks, Residential Yards, Storefronts |
| `TileChest` | `Chest.png` ($22\times21$) | Obstacle/Container | `true` | `false` | `false` | `WoodImage` / `TileFloorImage` | Bedrooms, Armory, Storage, Warehouse Bay |
| `TileSculpture` | `Sculpture-1.png` / `Sculture-2.png` ($29\times32$) | Obstacle/Prop | `true` | `false` | `false` | `GrassImage` / `ConcreteImage` | Town Square, Municipal Courtyards, Clinic Lobby |
| `TileBush` (optional prop) | `Bush-1.png` ($24\times18$) | Decorative Obstacle | `false` | `false` | `false` | `GrassImage` | Backyards, Park perimeters, Trails |
| `TileStone` (optional prop) | `Stone-1.png` ($28\times19$) | Solid Obstacle | `true` | `false` | `false` | `GrassImage` | Park wilderness, Road edges |

### 4. Sprite Anchoring and Transformation Matrix Calculation
For an obstacle sprite of native dimensions $(W, H)$ drawn at tile $(x, y)$ (world coords $worldX = x \cdot TileSize, worldY = y \cdot TileSize$):
- Center of the tile in isometric space is $(iso_X, iso_Y) = \text{WorldToIso}(worldX + 64, worldY + 64)$.
- For 2:1 isometric tiles, the base contact point should be centered horizontally at $-W/2$ and anchored vertically at $-H + \text{tileGroundOffset}$.
- The transformation sequence:
  ```go
  op := &ebiten.DrawImageOptions{}
  // Anchor sprite foot at base
  op.GeoM.Translate(-float64(imgW)/2.0, -float64(imgH) + 32.0)
  // Translate to camera-relative isometric position
  op.GeoM.Translate(isoX - camX, isoY - camY)
  // Apply zoom scale (0.5)
  op.GeoM.Scale(0.5, 0.5)
  // Translate to viewport center (640, 360)
  op.GeoM.Translate(640, 360)
  ```
- Memory tint for explored tiles:
  ```go
  if !s.gameMap.Visible[idx] && s.gameMap.Explored[idx] {
      op.ColorScale.Scale(0.2, 0.2, 0.3, 1)
  }
  ```
- Depth assignment:
  `Depth = worldX + worldY`

---

## Game Lifecycle & Entry Point (`cmd/game/main.go`)

### 1. Initialization Sequence
```
main()
  |--> ebiten.SetWindowSize(1280, 720)
  |--> ebiten.SetWindowTitle("Go Zomboid")
  |--> assets.Load()                // Native PNG asset decoding via sync.Once
  |--> g := game.NewGame()
        |--> assets.InitAudio()     // Audio subsystem & sound synthesis
        |--> g.Reset()
              |--> w := arkecs.NewWorld()
              |--> gameMap := world.NewMap(100, 100)
              |--> Spawn Player at gameMap.PlayerSpawn
              |--> Initialize & Snap Camera to (playerIsoX, playerIsoY)
              |--> Spawn Contextual Loot Spawns
              |--> Spawn 140 Zombies with Safe Perimeter (>= 1400 px)
              |--> NewUpdateSystem(w, gameMap)
              |--> NewDrawSystem(w, gameMap)
  |--> ebiten.RunGame(g)            // Starts 60 FPS update/draw loop
```

### 2. Main Game Loop Execution
- `Game.Update() error`:
  1. Advances `timeOfDay += 24.0 / (60 * 5 * 60)` (5-minute real-time day/night cycle).
  2. Detects player restart: if `ebiten.IsKeyPressed(KeyR)` and player is dead, calls `g.Reset()`.
  3. Dispatches `g.updateSys.Update()`:
     - Calculates raycasting FOV ($radius = 22$ tiles).
     - Smoothly lerps Camera towards player isometric position.
     - Handles item pickup (distance $< 64$ px, max 9 inventory slots).
     - Processes user input (WASD / Arrow movement, mouse click vector, keys 1-9 item consumption).
     - Processes zombie AI (wander timer, chasing when player in sight/range, runner speeds).
     - Evaluates AABB collision against solid tiles.
     - Resolves combat hitboxes, armor durability degradation, infection progression, hunger/thirst depletion.
- `Game.Draw(screen *ebiten.Image)`:
  - Clears screen with `#0F0F0F`.
  - Dispatches `g.drawSys.Draw(screen, g.timeOfDay)`.
- `Game.Layout(outsideWidth, outsideHeight int) (int, int)`:
  - Returns fixed internal virtual canvas resolution: `1280, 720`.

---

## Comprehensive Repository Test Suite Catalog

Execution of `CC=gcc go test -v ./...` verifies all packages compile and execute.

### 1. Catalog of All 22 Test Files Across 5 Packages

#### Package 1: `github.com/BryceWayne/go-zomboid/internal/assets` (4 test files)
1. `assets_test.go`:
   - `TestEmbeddedAssetDimensionsAndValidity`: Verifies embedded PNGs decode properly and have non-zero pixels.
   - `TestAssetsLoadAllPointersNonNil`: Checks that all exported image pointers are populated non-nil with correct bounds after `assets.Load()`.
2. `assets_stress_test.go`:
   - `TestFloorTileIsometricBounds`: Validates that floor tiles fit within the 2:1 isometric diamond without bleeding.
   - `TestCharacterGroundAnchor`: Ensures character sprites have grounded foot pixels in bottom rows `[112..127]`.
   - `TestItemOutlineContrast`: Checks item icon pixel density and dark outline contrast.
   - `TestAssetsLoadIdempotency`: Calls `Load()` repeatedly across multiple iterations without corruption.
3. `challenger_stress_test.go`:
   - `TestChallenger_All27ExportedPointersAndExactBounds`: Rigorous bounds check on all exported pointers.
   - `TestChallenger_MultiThreadedLoadAndPointerRace`: 20 concurrent loader goroutines + 30 reader goroutines stress-testing race conditions.
   - `TestChallenger_RepeatedSequentialLoads`: 100 consecutive `Load()` executions.
   - `TestChallenger_AssetPixelContrastAndColorSaturation`: Statistical analysis of perceived luminance, RMS contrast, and HSV saturation.
   - `TestChallenger_FloorTileGeometryDiamond`: Inner core opacity ($\ge 98\%$) and zero exterior bleeding.
   - `TestChallenger_CharacterGroundDropShadows`: Head pixel presence and ground shadow anchor presence.
   - `TestChallenger_ItemIconCenteringAndContour`: Centroid verification in $[20, 44]$ bounding window.
4. `empirical_challenger_test.go`:
   - `TestEmpiricalAssetCatalogCompleteness`: Catalog verification.
   - `TestEmpiricalAlphaFillRatios`: Fill ratio bounds for entities ($20-70\%$), floors ($45-55\%$), obstacles ($5-85\%$), items ($5-70\%$).
   - `TestEmpiricalFloorDiamondGeometry`: Manhattan distance boundary verification.
   - `TestEmpiricalCharacterGrounding`: Verifies solid boot pixels.
   - `TestEmpiricalGenerationDeterminism`: **Note: Runs `exec.Command("go", "run", "./cmd/tools/genassets")`**. Needs retirement when `genassets` is deleted.
   - `TestEmpiricalObstacleBoundsAndGrounding`: Checks lower half contact area for obstacles.
   - `TestEmpiricalItemIconQuality`: Verifies mass and centroid centering.

#### Package 2: `github.com/BryceWayne/go-zomboid/internal/game/world` (2 test files)
1. `map_test.go`:
   - `TestTileTypeProperties`: Validates `IsSolid()`, `BlocksVision()`, `IsFloor()`, and `String()` for all tile types.
   - `TestNewMapProceduralTown`: Verifies $100\times100$ map grid, perimeter walls, all tile type representations, and all 5 building archetypes.
   - `TestPlayerSafeSpawn`: Ensures player spawn is on non-solid tile and $\ge 1400$ px from all zombies.
   - `TestContextualLootSpawns`: Validates $\ge 10$ loot spawns on walkable tiles with full item variety (food, water, weapon, axe, shotgun, ammo, armor).
   - `TestZombieSpawnsNoTrapping`: Validates $\ge 50$ zombie spawns on non-solid tiles.
   - `TestCollisionDetection`: Tests `IsColliding` against non-solid grass, solid wall, tree, fence, debris, and out-of-bounds boundaries.
   - `TestFOVAndOcclusion`: Verifies wall occludes line of sight while fence allows raycast penetration.
   - `TestSmallFallbackMap`: Validates fallback generation for $20\times20$ map.
2. `world_empirical_stress_test.go`:
   - `TestEmpirical_All10TileTypesGenerated`: Asserts that all 10 core tile types (`Grass`, `Wall`, `Dirt`, `WoodFloor`, `Tree`, `Asphalt`, `Concrete`, `TileFloor`, `Fence`, `Debris`) have non-zero generation counts.
   - `TestEmpirical_All5BuildingArchetypesAndRooms`: Validates building dimensions and room containment.
   - `TestEmpirical_PlayerSpawnSafetyAndZombieDistance`: Verifies minimum zombie distance $\ge 350.0$ px across map seeds.
   - `TestEmpirical_100PercentZombieSpawnsNonSolid`: Verifies 4200 zombie spawns across 30 map generations are 100% non-solid.
   - `TestEmpirical_AABBCollisionSolidVsFloor`: Edge and corner overlap collision checks.
   - `TestEmpirical_FOVRaycastingWallVsFence`: Full circular raycasting verification.
   - `TestEmpirical_LootDistributionAndWalkability`: Validates counts of all 8 loot types.

#### Package 3: `github.com/BryceWayne/go-zomboid/internal/game` (14 test files)
1. `game_test.go`: `TestWorldToIso`, `TestNewGameInitialization`, `TestGameResetContextualSpawns`.
2. `game_stress_test.go`: `TestGameResetStress` (100 iterations), `TestIsometricProjectionMathStress` (fuzzing 5000 points), `TestIsometricRenderingAllTileTypesAndPropsStress` (headless draw across 24h day/night cycle, fog of war, and dead player), `TestGameLoopContinuousSimulationStress` (2500 consecutive simulation frames).
3. `game_empirical_stress_test.go`: `TestEmpirical_HeadlessContinuousGameLoop_1500Frames`, `TestEmpirical_GameResetStateInvariants`.
4. `camera_test.go` & `camera_empirical_challenger_test.go`: Camera lerp smoothing, snap behavior, viewport boundary culling, mouse navigation vectors.
5. `combat_test.go`, `combat_empirical_stress_test.go`, `combat_empirical_challenger_m4_test.go`: Melee/shotgun ranges, zombie damage/stuns, weapon durability, runner zombie behavior.
6. `armor_test.go`, `armor_empirical_challenge_test.go`, `armor_empirical_stress_test.go`: Tactical vest defense mitigation, infection resistance, durability degradation.
7. `bezier_combat_test.go`: Quadratic Bezier curves, control point math, swing arc rendering.
8. `e2e_tiers_test.go`: End-to-end integration tests for all 9 tier features.
9. `adversarial_challenger_m5_test.go`: Rapid world reset mid-combat, extreme entity densities, headless simulation stability.

#### Package 4: `github.com/BryceWayne/go-zomboid/internal/ecs` (1 test file)
1. `components_test.go`: Component initialization, struct fields, and Ark ECS map queries.

#### Package 5: `github.com/BryceWayne/go-zomboid/cmd/tools/genassets` (1 test file)
1. `genassets_test.go`: Determinism and PNG dimension assertions for the procedural generator. (To be deleted with `cmd/tools/genassets` directory per R1).

---

## Edge Cases, Rendering Pitfalls & Acceptance Criteria Requirements

### 1. Procedural Generation Retirement (Requirement R1)
- **Action**: Completely delete `/home/bryce/code/go-zomboid/cmd/tools/genassets` directory and its contents (`main.go`, `genassets_test.go`).
- **Pitfall**: `internal/assets/empirical_challenger_test.go` contains `TestEmpiricalGenerationDeterminism` which invokes `exec.Command("go", "run", "./cmd/tools/genassets")`. Once `genassets` is deleted, this test will fail unless retired or updated to verify static asset loading determinism instead.

### 2. External PNG Ingestion & Embedding (Requirement R2)
- **Action**: Copy PNG files from `context/` into `internal/assets/images/` and update `internal/assets/assets.go` to load them natively into `*ebiten.Image` variables.
- **Pitfall**: Filenames in `context/` have varying capitalizations and spaces (e.g. `Bench and chest/Bench.png`, `Sculptures/Sculture-2.png`, `Big-wooden-fence-1.png`). Files placed into `internal/assets/images/` should be cleanly named (e.g. `bench.png`, `chest.png`, `sculpture_1.png`, `sculpture_2.png`).
- **Asset Dimensions**:
  - `Bench.png`: $52 \times 37$ px
  - `Chest.png`: $22 \times 21$ px
  - `Sculpture-1.png`: $23 \times 31$ px
  - `Sculture-2.png`: $29 \times 32$ px
  - `Bush-1.png` to `Bush-4.png`: $19\times15$ to $28\times19$ px
  - `Stone-1.png`, `Stone-2.png`: $28\times19$ to $29\times25$ px
- All loaded pointers must be non-nil after `assets.Load()`.

### 3. World Logic Inference, Tile Types & DrawSystem Depth-Sorting (Requirement R3)
- **Action**:
  1. Add `TileBench`, `TileChest`, `TileSculpture` (and optionally `TileBush`, `TileStone`) constants in `internal/game/world/map.go`.
  2. Implement `IsSolid()`, `BlocksVision()`, `IsFloor()`, and `String()` for each new `TileType`.
  3. Place new objects into `NewMap()` (e.g., Benches in residential yards/parks, Chests in bedrooms/armory/storage, Sculptures in town parks/courtyards).
  4. Ensure existing 10 `TileType` constants (`Grass`, `Wall`, `Dirt`, `WoodFloor`, `Tree`, `Asphalt`, `Concrete`, `TileFloor`, `Fence`, `Debris`) remain present with $>0$ counts so `TestEmpirical_All10TileTypesGenerated` continues to pass without regression.
  5. Update `DrawSystem.Draw` in `internal/game/game.go`:
     - In Ground Pass: Render base ground (`GrassImage` or `WoodImage` or `ConcreteImage`) under new props.
     - In Props Pass: Add new tile types to the props collector, assign corresponding `*ebiten.Image`, compute anchor translation, and set `Depth: worldX + worldY`.
     - Sorting: The existing `sort.SliceStable` on `Depth` automatically sorts new objects properly with players, zombies, and items.

### 4. Acceptance Criteria Verification Checklist
- [x] `cmd/tools/genassets` directory deletion requirement identified and isolated.
- [x] Native PNG loading in `internal/assets/assets.go` mapped to all new assets.
- [x] `CC=gcc go test ./...` test suite surveyed with all 22 test files cataloged.
- [x] `CC=gcc go run ./cmd/game` lifecycle traced from window initialization to rendering loop.
