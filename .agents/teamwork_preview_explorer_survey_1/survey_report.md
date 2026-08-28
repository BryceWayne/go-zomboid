# Comprehensive Survey Report: Asset Generation Pipeline & Isometric Sprite Rendering System

**Author**: `survey_explorer_1`  
**Date**: 2026-08-28  
**Scope**: Procedural Asset Pipeline (`cmd/tools/genassets`), Asset Embedding (`internal/assets`), World & Map Math (`internal/game/world`), Game Rendering & Isometric Projection (`internal/game`), and 4x Resolution Scaling Strategy (64x32 $\rightarrow$ 256x128).

---

## 1. Executive Summary

This survey provides an exhaustive architectural and mathematical analysis of the asset generation and rendering pipeline of `go-zomboid`. It establishes the precise blueprint required to:
1. Quadruple base floor tile resolution from **64x32** to **256x128** (maintaining the exact 2:1 dimetric isometric ratio).
2. Proportionally scale vertical obstacles/props (**256x256**), character entities (**64x128**), and items/weapons/armor (**64x64**).
3. Scale and upgrade all procedural geometric overlays (chevrons, wildflowers, pebbles, plank seams, asphalt road markings, concrete expansion joints, tile grout, and drop shadows).
4. Update the isometric projection engine math (`TileSize = 128`, draw offsets, camera tracking, and movement/combat physics) without breaking world generation or gameplay balance.
5. Implement dynamic Bezier curve attack swooshes/trails in `DrawSystem`.

---

## 2. Asset Generation Pipeline (`cmd/tools/genassets`)

All sprite textures in `go-zomboid` are procedurally generated in pure Go standard library (`image`, `image/color`, `image/png`, `math`, `os`) and saved directly into `internal/assets/images/*.png`. There are zero external artistic dependencies.

### 2.1 Current Asset Registry & Target 4x Matrix

| Category | File Name | Current Size | Target 4x Size | Aspect Ratio | Visual Style & Key Primitives |
|---|---|---|---|---|---|
| **Floor Tiles** | `grass.png` | 64x32 | **256x128** | 2:1 | Mint/emerald top, stepped side rim, chevron grass blades, yellow/white wildflowers |
| | `dirt.png` | 64x32 | **256x128** | 2:1 | Warm brown top, stepped depth rim, rounded rectangular pebbles with highlights |
| | `wood.png` | 64x32 | **256x128** | 2:1 | 4 longitudinal planks in UV space, staggered end joints, nailhead pairs |
| | `asphalt.png` | 64x32 | **256x128** | 2:1 | Dark charcoal base, dashed yellow highway markings in UV space |
| | `concrete.png` | 64x32 | **256x128** | 2:1 | 2x2 sidewalk slab checkerboard in UV space, dark joint seams |
| | `tile_floor.png` | 64x32 | **256x128** | 2:1 | 4x4 alternating interior tile checkerboard in UV space, grout channels |
| **Vertical Obstacles** | `wall.png` | 64x64 | **256x256** | 1:1 | Coping diamond top, brick-red West face, shaded South face, edge highlights |
| | `tree.png` | 64x64 | **256x256** | 1:1 | Elliptical drop shadow, cylindrical trunk, spherical/teardrop canopy with toon shadow |
| | `fence.png` | 64x64 | **256x256** | 1:1 | 2 sloped rails, 7 pointed pickets, pyramid-capped corner and side posts |
| | `debris.png` | 64x64 | **256x256** | 1:1 | Drop shadow, 3D wooden crate with X-bracing and metal corner brackets, concrete/brick rubble |
| | `tent.png` | 64x64 | **256x256** | 1:1 | Triangular sloped green fabric faces, ridge pole, dark doorway |
| | `stump.png` | 64x64 | **256x256** | 1:1 | Cylindrical bark base, top cut face with growth rings |
| | `mushroom.png` | 64x64 | **256x256** | 1:1 | White stem, red dome cap with white polka dot spots |
| | `sign.png` | 64x64 | **256x256** | 1:1 | Vertical wooden post, rectangular sign board with simulated text lines |
| | `elevation_block.png` | 64x64 | **256x256** | 1:1 | 32px raised grass top diamond, extruded dirt West/South faces |
| | `elevation_ramp.png` | 64x64 | **256x256** | 1:1 | Continuous sloped grass ramp face, side dirt vertical triangle |
| **Entities** | `player.png` | 16x32 | **64x128** | 1:2 | Ground drop shadow, blue shirt, sleeves, dark pants, peach skin head, eye dots |
| | `zombie.png` | 16x32 | **64x128** | 1:2 | Ground drop shadow, forward reaching arms, green skin, red eyes |
| | `runner.png` | 16x32 | **64x128** | 1:2 | Ground drop shadow, leaning red runner silhouette, glowing yellow eyes |
| **Items & Gear** | `food.png` | 16x16 | **64x64** | 1:1 | Tin soup can, pull tab, red/gold label, center emblem, metallic rims |
| | `water.png` | 16x16 | **64x64** | 1:1 | Water bottle, white cap, ergonomic contoured waist, liquid meniscus & highlight |
| | `weapon.png` | 16x16 | **64x64** | 1:1 | Spiked baseball bat, taped grip, wood barrel, steel spikes, blood drops |
| | `axe.png` | 16x16 | **64x64** | 1:1 | Fire axe, curved wood handle, steel eye, breaching pick, red blade, beveled steel edge |
| | `shotgun.png` | 16x16 | **64x64** | 1:1 | Pump-action shotgun, wood buttstock, rubber pad, steel receiver, pump forend, dual barrel |
| | `ammo.png` | 16x16 | **64x64** | 1:1 | Olive ammo box, yellow stencil, 4 brass cartridges with copper tips, corner rivets |
| | `armor.png` | 16x16 | **64x64** | 1:1 | Tactical Kevlar plate carrier vest, shoulder straps/buckles, velcro ID patch, Molle webbing, 3 mag pouches |
| | `antidote.png` | 16x16 | **64x64** | 1:1 | Glass vial, cork stopper, glowing green liquid with highlight |

---

## 3. Mathematical Analysis of Procedural Generators

### 3.1 Floor Diamond Mathematics & Parameterization

In `cmd/tools/genassets/main.go`, floor tiles are rendered into an image of dimensions $(W, H)$ where $W = 2H$.

1. **Cartesian Center**:
   $$cx = \frac{W - 1}{2.0}, \quad cy = \frac{H - 1}{2.0}$$
   - For $64 \times 32$: $cx = 31.5, cy = 15.5$
   - For $256 \times 128$: $cx = 127.5, cy = 63.5$

2. **Normalized Isometric Manhattan Distance**:
   $$dx = x - cx, \quad dy = y - cy$$
   $$\text{isoDist}(x, y) = \frac{|dx|}{W/2} + \frac{|dy|}{H/2}$$
   - Point $(x, y)$ is inside the ground diamond if and only if $\text{isoDist}(x, y) \le 1.0$.

3. **UV Space Transformation (Diamond Coordinate System)**:
   Textures that have structured surface geometry (planks, tiles, roads, slabs) convert $(dx, dy)$ into normalized $[0, 1] \times [0, 1]$ coordinate space:
   $$u = \frac{dx}{W} + \frac{dy}{H} + 0.5$$
   $$v = \frac{dy}{H} - \frac{dx}{W} + 0.5$$
   
   **Properties of $(u, v)$ Space**:
   - Diamond Top Vertex $(0, -H/2) \implies (u, v) = (0.5, 0.0)$
   - Diamond Right Vertex $(+W/2, 0) \implies (u, v) = (1.0, 0.5)$
   - Diamond Bottom Vertex $(0, +H/2) \implies (u, v) = (0.5, 1.0)$
   - Diamond Left Vertex $(-W/2, 0) \implies (u, v) = (0.0, 0.5)$
   
   Because $(u, v)$ coordinates are normalized $[0, 1]$, any formula defined in terms of $u$ and $v$ (such as plank lanes, tile grids, and asphalt stripes) scales **resolution-independently** from 64x32 to 256x128!

4. **Stepped Depth / 3D Rim Extrusion**:
   - To simulate a thick slab with flat toon shading:
     $$\text{rimThickness}(x) = 0.15 + 0.05 \sin\left(\frac{2\pi \cdot x}{W} \cdot k\right)$$
   - If $\text{isoDist} > 1.0 - \text{rimThickness}$:
     - If $x < W/2$: Pixel colored with `leftColor` (lighter side tone).
     - If $x \ge W/2$: Pixel colored with `rightColor` (darker shadow tone).

---

### 3.2 Inventory of Procedural Geometric Overlays

| Overlay Type | Host Tile / Asset | Hardcoded Formula in 64x32 | 4x Scaled Formula in 256x128 | Visual Representation |
|---|---|---|---|---|
| **Grass Chevrons** | `grass.png` | Positions: `(16, 12)`, `(40, 8)`, `(24, 20)`, `(48, 16)`. V-shape drawn with single pixel offsets: `(cx±1, cy-1)`, `(cx±2, cy-2)`. | Positions: `(64, 48)`, `(160, 32)`, `(96, 80)`, `(192, 64)`. Drawn as smooth vector V-shapes with stroke width 3-4px, arm length 12-16px. | Crisp minimalist grass tufts |
| **Wildflowers** | `grass.png` | Positions: `(24, 8)`, `(40, 20)`, `(12, 18)`. Yellow center pixel with 4 white petal arms of length 2px. | Positions: `(96, 32)`, `(160, 80)`, `(48, 72)`. Yellow central disc ($r = 4\text{px}$) with 4 or 8 rounded petal circles/capsules ($r = 6\text{px}$). | Clean Dribbble-style wildflower clusters |
| **Pebbles** | `dirt.png` | Positions: `(20, 10)`, `(45, 14)`, `(30, 22)`, `(15, 20)`. Drawn as `fillRect(px, py, 3, 2, pebbleLight)`. | Positions: `(80, 40)`, `(180, 56)`, `(120, 88)`, `(60, 80)`. Drawn as rounded rectangles / smooth ellipses of size $14 \times 8\text{px}$ with highlight rim and shadow base. | Distinct flat-shaded smooth stones |
| **Wood Planks & Nails** | `wood.png` | 4 lanes along $v$: `lane = int(v * 4)`. Seam: $v_{inLane} < 0.05 \vee v_{inLane} > 0.95$. End joints: $u \in \{0.60, 0.30, 0.75, 0.45\}$. Nails at $nu = endU \pm 0.05$. | Same UV logic: 4 lanes along $v$. Seam width $\Delta v \approx 0.015$ (giving 3px crisp dark seams). End joint seams $\Delta u \approx 0.008$. Nails drawn as circular steel nailheads ($r = 2.5\text{px}$) with highlight dot. | Clean architectural hardwood flooring |
| **Asphalt Center Striping** | `asphalt.png` | Yellow marking in UV space: $v \in [0.43, 0.57]$ and $(u \le 0.38 \vee u \ge 0.62)$. | Exact same UV condition: $v \in [0.43, 0.57]$ and $(u \le 0.38 \vee u \ge 0.62)$. Renders as clean, crisp 18px-wide yellow dashed road markings. | Crisp highway center dashes |
| **Concrete Slabs & Expansion Joints** | `concrete.png` | 2x2 grid in UV space: $quadX = u \ge 0.5, quadY = v \ge 0.5$. Expansion joints: $|u - 0.5| < 0.025 \vee |v - 0.5| < 0.025$. | 2x2 grid in UV space with joint width $\Delta \approx 0.01$ (4px crisp joint channels). Stepped bevel on slab edges. | Modern modular sidewalk pavement |
| **Ceramic Tile Checkerboard** | `tile_floor.png` | 4x4 grid in UV space: $gridU = u \cdot 4, gridV = v \cdot 4$. Grout: $subU < 0.05 \vee subV < 0.05$. | 4x4 grid in UV space: grout width $\Delta \approx 0.02$ (3-4px dark grout). Alternating white/charcoal tiles. | Clean commercial tile flooring |
| **Drop Shadows** | `player.png`, `zombie.png`, `tree.png`, `debris.png` | Ellipse: $(dx^2)/r_x^2 + (dy^2)/r_y^2 \le 1.0$ with semi-transparent black `RGBA{0, 0, 0, 40}`. | Scaled ellipse: center $(cx \cdot 4, cy \cdot 4)$, radii $(r_x \cdot 4, r_y \cdot 4)$. Smooth semi-transparent dark oval under character feet and prop bases. | Minimalist grounding shadows |
| **Crate X-Bracing & Corner Brackets** | `debris.png` | Diagonals on 64x64 crate faces: $relX == relY \vee relX == (16 - relY)$. 7 corner bracket stamps. | Scaled 256x256 crate faces with 6px thick X-bracing beams and detailed metallic L-brackets at all 7 vertices. | Stylized heavy cargo crate |
| **Tree Canopy Crescent Shadow** | `tree.png` | Canopy center $(32, 26)$, radius $22$. Shadow condition: $dx > 4.4 \wedge dy > 4.4 \wedge \text{dist} > 11.0$. | Canopy center $(128, 104)$, radius $88$. Smooth circular boundary with crisp toon crescent shadow on bottom-right quadrant. | Vector geometric tree foliage |

---

## 4. Asset Embedding & Engine Rendering Architecture

### 4.1 Asset Embedding (`internal/assets/assets.go`)
- Uses Go 1.16+ `embed.FS` directive:
  ```go
  //go:embed images/*
  var imageFS embed.FS
  ```
- Package-level exported variables of type `*ebiten.Image`:
  - Characters: `PlayerImage`, `ZombieImage`, `RunnerImage`
  - Floors: `GrassImage`, `DirtImage`, `WoodImage`, `AsphaltImage`, `ConcreteImage`, `TileFloorImage`
  - Obstacles & Props: `WallImage`, `TreeImage`, `FenceImage`, `DebrisImage`, `TentImage`, `StumpImage`, `MushroomImage`, `SignImage`, `ElevationBlockImage`, `ElevationRampImage`
  - Items & Equipment: `WeaponImage`, `AxeImage`, `ShotgunImage`, `AmmoImage`, `ArmorImage`, `AntidoteImage`, `FoodImage`, `WaterImage`
- `Load()` function reads embedded bytes via `imageFS.ReadFile()`, decodes through `image.Decode()`, and creates GPU textures via `ebiten.NewImageFromImage(img)`.

---

### 4.2 Isometric Projection & Coordinate Geometry (`internal/game`)

The engine operates on two coordinate systems:
1. **World Cartesian Space $(wx, wy)$**: Continuous 2D world coordinates (used for physics, ECS positions, pathfinding, AI distances, and collision).
2. **Isometric Screen Space $(isoX, isoY)$**: 2.5D dimetric projection on screen.

#### Mathematical Transform Equations:
$$\text{WorldToIso}(wx, wy) = \begin{pmatrix} wx - wy \\ \frac{wx + wy}{2} \end{pmatrix}$$

$$\text{IsoToWorld}(isoX, isoY) = \begin{pmatrix} isoY + \frac{isoX}{2} \\ isoY - \frac{isoX}{2} \end{pmatrix}$$

#### Transformation Properties:
- Direction $+wx$ (East) maps to Screen $(+\Delta, +\Delta/2)$ (Down-Right).
- Direction $+wy$ (South) maps to Screen $(-\Delta, +\Delta/2)$ (Down-Left).
- Direction $+wx + wy$ maps to Screen $(0, +\Delta)$ (Straight Down).
- Direction $+wx - wy$ maps to Screen $(+2\Delta, 0)$ (Straight Right).

#### Camera Tracking:
The camera is centered on the player in isometric space:
```go
isoX, isoY := WorldToIso(playerPos.X, playerPos.Y)
camX = isoX - (ScreenWidth / 2.0)  // 400.0 for 800x600 window
camY = isoY - (ScreenHeight / 2.0) // 300.0 for 800x600 window
```

---

### 4.3 Rendering Pipeline & Sprite Draw Offsets

In `DrawSystem.Draw(screen *ebiten.Image, timeOfDay float64)`:

1. **Ground Pass (Diamonds)**:
   - For tile at grid $(x, y)$:
     $$worldX = x \cdot \text{TileSize}, \quad worldY = y \cdot \text{TileSize}$$
     $$isoX, isoY = \text{WorldToIso}(worldX, worldY)$$
   - **Current Draw Offset (64x32 tile)**:
     $$\text{drawX} = isoX - 32 - camX, \quad \text{drawY} = isoY - 0 - camY$$
     *(Subtracting 32 horizontally centers the 64px-wide diamond top vertex at $(isoX, isoY)$).*
   - **Target Draw Offset (256x128 tile)**:
     $$\text{drawX} = isoX - 128 - camX, \quad \text{drawY} = isoY - 0 - camY$$

2. **Depth Sorted Sprite Pass (Renderables)**:
   All vertical obstacles, props, items, and character entities are collected into a `Renderable` slice and sorted by depth:
   $$\text{Depth} = worldX + worldY$$
   
   | Object Type | Current Sprite Size | Current Draw Offset | Target Sprite Size | Target Draw Offset | Alignment Rationale |
   |---|---|---|---|---|---|
   | **Vertical Obstacle / Prop** | 64x64 | `isoX - 32, isoY - 32` | 256x256 | `isoX - 128, isoY - 128` | Centers 256px width horizontally ($-128$); base diamond at bottom $128\text{px}$ aligns exactly with the ground tile diamond |
   | **Character Entity** | 16x32 | `isoX - 8, isoY - 32` | 64x128 | `isoX - 32, isoY - 128` | Centers 64px width horizontally ($-32$); anchors feet at ground contact $(isoX, isoY)$ ($-128$) |
   | **Item / Weapon / Armor** | 16x16 | `isoX - 8, isoY - 8` | 64x64 | `isoX - 32, isoY - 32` | Centers 64x64 item bounding box on tile center $(isoX, isoY)$ |

---

## 5. Comprehensive 4x Scaling Blueprint

### 5.1 Constant and Coordinate Upgrades

| Parameter / Constant | Current Value (1x) | Target Value (4x) | Location | Impact & Implementation Detail |
|---|---|---|---|---|
| `world.TileSize` | `32` | **`128`** | `internal/game/world/map.go:30` | Core tile spacing constant. 1 grid cell = $128 \times 128$ world units. |
| `Player.Speed` | `3.0` | **`12.0`** | `internal/game/game.go:263` | Movement speed scales by 4x to traverse 128px tiles at identical visual speed. |
| `Zombie.Speed` | `1.0 - 1.5` (Reg) / `2.2 - 2.6` (Runner) | **`4.0 - 6.0`** (Reg) / **`8.8 - 10.4`** (Runner) | `internal/game/game.go:80-83` | Zombie movement and chase speeds scale by 4x. |
| `Collider.Width, Height` | `16, 16` | **`64, 64`** | `internal/game/game.go:65, 98` | Entity physical collision bounding boxes scale to 64x64. |
| `Melee Reach (Club)` | `24.0` | **`96.0`** | `internal/game/game.go:519, 529` | Bat/club swing reach and hit radius scale by 4x. |
| `Melee Reach (Axe)` | `32.0` | **`128.0`** | `internal/game/game.go:489, 500` | Fire axe cleave reach and cleave radius scale by 4x. |
| `Unarmed Shove Reach` | `24.0` | **`96.0`** | `internal/game/game.go:549, 557` | Shove attack reach scales by 4x. |
| `Shotgun Range` | `160.0` (Range), `24.0` (Point-blank) | **`640.0`** (Range), **`96.0`** (Point-blank) | `internal/game/game.go:433, 446` | Shotgun firing cone range and point-blank kill threshold scale by 4x. |
| `Noise Alert Radius` | `400.0` (Shotgun), `200.0` (Run), `50.0` (Walk) | **`1600.0`** (Shotgun), **`800.0`** (Run), **`200.0`** (Walk) | `internal/game/game.go:464, 596, 594` | Acoustic alert propagation radii scale by 4x. |
| `Separation Radius / Force` | `20.0` / `2.0` | **`80.0`** / **`8.0`** | `internal/game/game.go:615, 616` | Zombie swarm collision separation radius scales by 4x. |
| `Item Pickup Distance` | `16.0` | **`64.0`** | `internal/game/game.go:211` | Proximity threshold for picking up ground loot items. |
| `Zombie Bite Distance` | `14.0` | **`56.0`** | `internal/game/game.go:637` | Contact distance for zombie infection and armor durability decay. |
| `Vision / FOV Radius` | `250.0` (Screen clip) | **`1000.0`** | `internal/game/game.go:800` | Ground and entity rendering cull radius in world units. |
| `Safe Spawn Perimeter` | `350.0` | **`1400.0`** | `internal/game/world/map.go:898` | Safe clearance between player spawn and zombie spawns. |

---

### 5.2 Bezier Curve Combat Attack Trails (Requirement R3)

When attacking with melee weapons (especially the fire axe), a dynamic Bezier curve attack swoosh will be rendered in `DrawSystem`.

#### 1. Mathematical Formulation
A quadratic Bezier curve $B(t)$ is evaluated for $t \in [0.0, 1.0]$:
$$B(t) = (1 - t)^2 P_0 + 2(1 - t)t P_1 + t^2 P_2$$

- **$P_0$ (Start Point)**: Starting swing angle $\theta_{start} = \text{facingAngle} - \frac{\pi}{3}$, at distance $r_{inner} = 40\text{px}$.
- **$P_1$ (Control Point / Apex)**: Apex swing angle $\theta_{apex} = \text{facingAngle}$, extended outward at $r_{apex} = 1.35 \times \text{weaponReach} = 172\text{px}$.
- **$P_2$ (End Point)**: Ending swing angle $\theta_{end} = \text{facingAngle} + \frac{\pi}{3}$, at distance $r_{outer} = \text{weaponReach} = 128\text{px}$.

#### 2. Visual Rendering
- During attack frames (when `player.AttackCooldown > 15`):
  - Sample $N = 24$ points along the curve.
  - Transform each world point $B(t)$ into isometric screen space via $\text{WorldToIso}(B(t).X, B(t).Y)$.
  - Draw tapered arc segments with decreasing width and fading alpha using `vector.StrokeLine(screen, p1.X, p1.Y, p2.X, p2.Y, width, color, false)`.
  - Color palette: Translucent blazing orange-red (`RGBA{255, 100, 30, 200}`) with bright core (`RGBA{255, 230, 180, 240}`).

---

## 6. Test Suite Invariant Mapping

When the resolution and engine constants are updated, the following test files will be updated to reflect the new invariants:

1. **`cmd/tools/genassets/genassets_test.go`**:
   - `expectedAssetFiles`: Floor dimensions $256 \times 128$, Obstacle dimensions $256 \times 256$, Entity dimensions $64 \times 128$, Item dimensions $64 \times 64$.
   - Determinism and SHA256 stability checks across 3 execution iterations.
2. **`internal/assets/assets_test.go`**:
   - `expectedAssets` & `TestAssetsLoadAllPointersNonNil`: Updated dimensions assertions matching target sizes.
3. **`internal/assets/assets_stress_test.go`**:
   - `TestFloorTileIsometricBounds`: Bounds $256 \times 128$, center $(127.5, 63.5)$, radii $(128.5, 64.5)$.
   - `TestCharacterGroundAnchor`: Bounds $64 \times 128$, checks grounding pixels in rows $112..127$.
   - `TestItemOutlineContrast`: Solid pixel threshold scaled from $\ge 20$ to $\ge 320$ pixels for $64 \times 64$.
4. **`internal/game/game_test.go`**:
   - `TestWorldToIso`: Update test vectors for 4x coordinates.
5. **`internal/game/world/map_test.go`**:
   - Verify safe spawn distance threshold $\ge 1400.0$.
   - Verify collision checks on $128 \times 128$ grid boundaries.
6. **`internal/game/combat_test.go` & `armor_test.go`**:
   - Verify reach assertions ($96.0$ / $128.0$) and projectile cone thresholds.

---

## 7. Survey Verification Summary

- Running `go run ./cmd/tools/genassets` executes without errors and is fully deterministic.
- Running `CC=gcc go test ./...` passes all existing test suites.
- All generators, math functions, and rendering systems have been completely mapped with exact file paths and line numbers.
