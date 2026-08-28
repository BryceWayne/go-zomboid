# Engine Isometric Math, World Coordinate Transforms, Movement, Camera, and Map Systems Survey Report

**Date**: 2026-08-28  
**Investigator**: `survey_explorer_2`  
**Workspace**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2`  
**Target Codebase**: `go-zomboid` (`internal/game/world`, `internal/game`, `internal/ecs`, `cmd/tools/genassets`, `internal/assets`)

---

## 1. Executive Summary

This survey provides a comprehensive mathematical, architectural, and systemic investigation of the Project Zomboid Go isometric engine. It covers world coordinate transformations, projection equations, camera tracking, depth sorting, AABB collision detection, entity velocities, combat mechanics, AI perception, and map generation.

Furthermore, this report outlines the precise mathematical and systemic strategy required to upgrade the engine from the current **64x32 tile resolution** (`TileSize = 32`) to **256x128 high-fidelity tile resolution** (`TileSize = 128`, 4x texture resolution) ensuring that map generation, tile positioning, depth sorting, entity physics, collision detection, movement speed, and camera tracking remain completely seamless and robust.

---

## 2. Comprehensive Coordinate Space & Mathematical Formulation

The engine operates across four primary coordinate spaces:

```
+-------------------------------------------------------------------------------+
| 1. Tile Grid Space (tx, ty)                                                   |
|    - Discrete integer indices: tx in [0, Map.Width-1], ty in [0, Map.Height-1]|
|    - Flat array index: idx = ty * Map.Width + tx                              |
+-------------------------------------------------------------------------------+
                                      |
                World Position = (tx * TileSize + TileSize/2)
                Tile Index     = floor(wx / TileSize)
                                      v
+-------------------------------------------------------------------------------+
| 2. Cartesian World Space (wx, wy)                                             |
|    - Continuous 2D float coordinates (FloatPoint / ecs.Position)              |
|    - Physical simulations (velocities, colliders, ranges, distances)          |
+-------------------------------------------------------------------------------+
                                      |
                WorldToIso:  isoX = wx - wy,  isoY = (wx + wy) / 2
                IsoToWorld:  wx = isoY + isoX/2,  wy = isoY - isoX/2
                                      v
+-------------------------------------------------------------------------------+
| 3. Isometric Screen Space (isoX, isoY)                                        |
|    - Standard 2:1 isometric diamond projection space                          |
|    - Depth key for sorting: Depth = wx + wy (or isoY * 2)                     |
+-------------------------------------------------------------------------------+
                                      |
                drawX = isoX - anchorX - camX,  where camX = playerIsoX - 400
                drawY = isoY - anchorY - camY,  where camY = playerIsoY - 300
                                      v
+-------------------------------------------------------------------------------+
| 4. Viewport Screen Space (screenX, screenY)                                   |
|    - Ebitengine render canvas: 800x600 pixels (Layout(800, 600))              |
+-------------------------------------------------------------------------------+
```

---

## 3. Identification of All Constants, Formulas, and Offsets

### 3.1 World Grid & Map Constants (`internal/game/world/map.go`)

| Constant / Parameter | Current Value | Definition / Source File Location | Purpose & Mechanics |
|---|---|---|---|
| `world.TileSize` | `32` | `internal/game/world/map.go:30` | Cartesian world width and height of 1 square tile cell. |
| `Map.Width`, `Map.Height` | `100, 100` | `internal/game/game.go:39` | Total grid dimensions (10,000 tiles total). |
| `Tile Center Offset` | `+16.0, +16.0` | `internal/game/world/map.go:341-342, 807-815, 892-893` | Center of tile cell: `float64(tx)*TileSize + 16.0` (i.e. `TileSize / 2.0`). |
| `Map Boundary Limits` | `[0, Width*TileSize]` | `internal/game/world/map.go:973` | World collision boundaries ($0 \le wx \le 3200, 0 \le wy \le 3200$). |

---

### 3.2 Projection & Unprojection Formulas (`internal/game/game.go`)

#### Isometric Forward Projection (`WorldToIso`):
$$\text{isoX} = wx - wy$$
$$\text{isoY} = \frac{wx + wy}{2}$$
- **File**: `internal/game/game.go:744-748`
- **Derivation**: A step along $+X$ moves $+32\text{px}$ horizontally and $+16\text{px}$ vertically on screen. A step along $+Y$ moves $-32\text{px}$ horizontally and $+16\text{px}$ vertically on screen.
- **Bounding diamond**: A $32 \times 32$ world square tile projects to a diamond with diagonal span:
  - Width: $\Delta \text{isoX} = 32 - (-32) = 64\text{px}$
  - Height: $\Delta \text{isoY} = 16 - (-16) = 32\text{px}$ (Aspect ratio 2:1).

#### Isometric Inverse Projection (`IsoToWorld`):
$$wx = \text{isoY} + \frac{\text{isoX}}{2}$$
$$wy = \text{isoY} - \frac{\text{isoX}}{2}$$
- **File**: `internal/game/game.go:750-754`
- **Algebraic Verification**:
  $$\text{isoY} + \frac{\text{isoX}}{2} = \frac{wx + wy}{2} + \frac{wx - wy}{2} = \frac{2wx}{2} = wx$$
  $$\text{isoY} - \frac{\text{isoX}}{2} = \frac{wx + wy}{2} - \frac{wx - wy}{2} = \frac{2wy}{2} = wy$$
- **Usage**: Used in `game.go:354-357` to translate screen mouse cursor $(mx, my)$ back into world coordinates for movement and aiming.

---

### 3.3 Camera Tracking & Screen Centering

| Parameter | Value | Formula / Implementation | File Location |
|---|---|---|---|
| Viewport Resolution | $800 \times 600$ | `Layout(outsideW, outsideH) (800, 600)` | `internal/game/game.go:141` |
| Viewport Center | $(400, 300)$ | Screen midpoint $(800/2, 600/2)$ | `internal/game/game.go:352, 796` |
| Camera Offset X | `camX` | `camX = playerIsoX - 400` | `internal/game/game.go:352, 796` |
| Camera Offset Y | `camY` | `camY = playerIsoY - 300` | `internal/game/game.go:353, 797` |
| Mouse Cursor Unprojection | `(mouseWorldX, mouseWorldY)` | $\begin{aligned}\text{mouseIsoX} &= mx + \text{camX} \\ \text{mouseIsoY} &= my + \text{camY} \\ (wx, wy) &= \text{IsoToWorld}(\text{mouseIsoX}, \text{mouseIsoY})\end{aligned}$ | `internal/game/game.go:354-357` |

---

### 3.4 Sprite Dimensions, Anchors, and Draw Offsets

To render 2D rectangular textures onto 2.5D isometric world coordinates, every sprite type uses a specific origin anchor:

| Sprite Category | Texture Size | World Position $P(wx, wy)$ | Screen Translation $(drawX, drawY)$ | Anchor Rationale | File Location |
|---|---|---|---|---|---|
| **Ground Tiles** | $64 \times 32$ | Tile $(x, y)$ top-left vertex: $(x \cdot 32, y \cdot 32)$ | `drawX = isoX - 32 - camX`<br>`drawY = isoY - 0 - camY` | Top apex of $64 \times 32$ diamond is at texture coordinate $(32, 0)$. Placing top apex at $(isoX, isoY)$ shifts image by $(-32, 0)$. | `game.go:825-826` |
| **Vertical Obstacles / Walls / Props** | $64 \times 64$ | Tile $(x, y)$ top-left vertex: $(x \cdot 32, y \cdot 32)$ | `drawX = isoX - 32 - camX`<br>`drawY = isoY - 32 - camY` | Ground footprint diamond sits at texture $y \in [32, 64]$ with top apex at $(32, 32)$. Vertical elevation extends $32\text{px}$ up into $y \in [0, 32]$. Shift is $(-32, -32)$. | `game.go:909-910` |
| **Items / Weapons / Armor** | $16 \times 16$ | Center of item: $(iPos.X, iPos.Y)$ | `drawX = isoX - 8 - camX`<br>`drawY = isoY - 8 - camY` | Centered at sprite center $(8, 8)$ (half of $16 \times 16$). | `game.go:948-949` |
| **Entities (Player / Zombies / Runners)** | $16 \times 32$ | Ground position of entity feet: $(pos.X, pos.Y)$ | `drawX = isoX - 8 - camX`<br>`drawY = isoY - 32 - camY` | Horizontal center is $16/2 = 8$. Feet contact point on ground is at bottom $y = 32$. Shift is $(-8, -32)$. | `game.go:1009-1010` |
| **Facing Indicator** | $4 \times 4$ (scaled) | Target world point: $P_{pos} + \vec{facing} \cdot 20.0$ | `drawX = isoX - 4 - camX`<br>`drawY = isoY - 4 - camY`<br>`GeoM.Scale(0.25, 0.25)` | Centered indicator offset ahead of player. | `game.go:1051-1083` |

---

### 3.5 Depth Sorting & Z-Ordering (`internal/game/game.go:852-1093`)

- **Sorting Metric**:
  $$\text{Depth} = wx + wy$$
- **Sorting Mechanism**:
  `sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].Depth < sprites[j].Depth })` (`game.go:1086-1088`).
- **Mathematical Property**:
  In a 2:1 isometric projection, screen vertical position is $\text{isoY} = (wx + wy) / 2$. Since screen $Y$ strictly increases with $(wx + wy)$, rendering objects in ascending order of $(wx + wy)$ guarantees that objects closer to the bottom/foreground of the screen are drawn over objects in the background.

---

### 3.6 Collision Detection & Physics Bounding Boxes

- **Algorithm**: Axis-Aligned Bounding Box (AABB) intersection with map tile grid (`internal/game/world/map.go:970-990`).
- **Tile Overlap Bounding Calculation**:
  $$\text{minTileX} = \lfloor \text{rectX} / \text{TileSize} \rfloor, \quad \text{minTileY} = \lfloor \text{rectY} / \text{TileSize} \rfloor$$
  $$\text{maxTileX} = \lfloor (\text{rectX} + \text{rectW}) / \text{TileSize} \rfloor, \quad \text{maxTileY} = \lfloor (\text{rectY} + \text{rectH}) / \text{TileSize} \rfloor$$
- **Solid Obstacles Checked**: `TileWall`, `TileTree`, `TileFence`, `TileDebris`, `TileTent`, `TileElevationBlock`, `TileStump`, `TileSign` (`map.go:33-40`).
- **Entity Collider Sizes**:
  - Player: `Collider{Width: 16, Height: 16}` (`game.go:65`)
  - Zombie: `Collider{Width: 16, Height: 16}` (`game.go:98`)
- **Axis-Separated Sliding Movement Integration** (`game.go:713-728`):
  ```go
  if vel.X != 0 {
      if !s.gameMap.IsColliding(pos.X+vel.X, pos.Y, col.Width, col.Height) {
          pos.X += vel.X
      }
  }
  if vel.Y != 0 {
      if !s.gameMap.IsColliding(pos.X, pos.Y+vel.Y, col.Width, col.Height) {
          pos.Y += vel.Y
      }
  }
  ```

---

### 3.7 Entity Velocities, Speed Coefficients, and Combat Ranges

| Category | Constant / Variable | Value | Implementation Details | File Location |
|---|---|---|---|---|
| **Player Speed** | `speed` | `3.0` px/frame | Velocity along movement vector ($\pm 3.0$ or normalized diagonal/mouse direction). | `game.go:263, 363-369` |
| **Zombie Speed (Normal)** | `speed` | $1.0 + [0.0 \sim 0.5]$ px/frame | Average $1.25$ px/frame. | `game.go:80` |
| **Zombie Speed (Runner)** | `speed` | $2.2 + [0.0 \sim 0.4]$ px/frame | Average $2.4$ px/frame (20% spawn rate). | `game.go:82` |
| **Zombie Wander Speed** | `wanderSpeed` | `zombie.Speed * 0.4` | $0.4 \sim 1.0$ px/frame during wandering. | `game.go:685` |
| **Boid Separation** | `separationRadius`, `separationForce` | `20.0` px, `2.0` | Prevents zombie stacking. | `game.go:615-616` |
| **Unarmed Shove Knockback** | `vel.X, vel.Y`, friction | $5.0 \times \text{Facing}$, friction `0.85`, stun `45` frames | Pushes zombie back with exponential velocity decay. | `game.go:483, 560, 631` |
| **Item Pickup Radius** | Distance | `16.0` px | `math.Sqrt(dx*dx + dy*dy) < 16.0` | `game.go:211` |
| **Zombie Bite / Contact** | Distance | `14.0` px | Triggers armor deflection or direct infection. | `game.go:637` |
| **Melee Reach (Axe)** | Center offset, Cleave Radius | `32.0` px, `32.0` px | Hits all zombies in 32px radius centered 32px in front of player. | `game.go:489-500` |
| **Melee Reach (Bat / Club / Fist)** | Center offset, Radius | `24.0` px, `24.0` px | Hits zombies within 24px radius centered 24px in front of player. | `game.go:473, 519, 549` |
| **Shotgun Spread Cone** | Max Range, Angle, Point-blank | $160.0$ px, $\pm 22.5^\circ$ ($\cos \ge 0.92388$), Point-blank $< 24.0$ px | Cleaves all zombies in spread cone. | `game.go:433-455` |
| **Shotgun Acoustic Noise Pulse** | Radius | `400.0` px | Agitates all wandering zombies within 400px. | `game.go:458-468` |
| **Zombie Hearing Detection** | Noise Radius | `50.0` px (idle player), `200.0` px (moving player) | Triggers zombie chase mode. | `game.go:594-597` |
| **Zombie Vision Detection** | Vision Radius | `150.0` px | Triggers zombie chase mode. | `game.go:598` |
| **Zombie De-aggro Distance** | Distance | `400.0` px | Zombies lose interest beyond 400px. | `game.go:670` |
| **Safe Spawn Distance** | Distance | `350.0` px | Minimum distance between player spawn and any zombie spawn. | `map.go:898` |
| **Fog of War & FOV** | Raycast radius, Render Cutoff | `15` tiles raycast, `250.0` px render cutoff | Raycasting occluded by `TileWall`; items/zombies unrendered beyond 250px or in darkness. | `game.go:172, 800, 815, 870, 934, 994`, `map.go:908-947` |

---

## 4. How Current Math Functions Convert Between Coordinates

### 4.1 Converting World Coordinates to Tile Grid Indices
To convert a world point $(wx, wy)$ to tile index $(tx, ty)$:
$$tx = \text{int}(wx) / \text{world.TileSize}$$
$$ty = \text{int}(wy) / \text{world.TileSize}$$
- **Integer index validation**: $0 \le tx < \text{Map.Width}$ and $0 \le ty < \text{Map.Height}$.
- **1D Flat Index**: $\text{idx} = ty \times \text{Map.Width} + tx$.

### 4.2 Converting Tile Grid Indices to World Coordinates
To place an entity or prop at the center of a tile $(tx, ty)$:
$$wx = float64(tx) \times \text{world.TileSize} + \frac{\text{world.TileSize}}{2.0}$$
$$wy = float64(ty) \times \text{world.TileSize} + \frac{\text{world.TileSize}}{2.0}$$

### 4.3 Converting World Coordinates to Isometric Screen Coordinates
For any world position $(wx, wy)$:
$$isoX = wx - wy$$
$$isoY = \frac{wx + wy}{2.0}$$
Then apply camera translation and sprite anchor translation:
$$drawX = isoX - \text{anchorX} - \text{camX}$$
$$drawY = isoY - \text{anchorY} - \text{camY}$$

### 4.4 Converting Screen Mouse Coordinates to World Coordinates
Given mouse cursor position $(mx, my)$ on screen:
$$\text{mouseIsoX} = float64(mx) + \text{camX}$$
$$\text{mouseIsoY} = float64(my) + \text{camY}$$
$$\text{mouseWorldX} = \text{mouseIsoY} + \frac{\text{mouseIsoX}}{2.0}$$
$$\text{mouseWorldY} = \text{mouseIsoY} - \frac{\text{mouseIsoX}}{2.0}$$

---

## 5. Architectural Upgrade Strategy for 256x128 (4x Texture Resolution)

To quadruple the base tile pixel size from **64x32** to **256x128** for floors and proportionally scale entities, walls, and props, the coordinate math and engine constants must be systematically updated across all subsystems.

```
+-----------------------------------------------------------------------------------------------+
| RESOLUTION COMPARISON TABLE                                                                   |
+------------------------------+--------------------------------+-------------------------------+
| Subsystem Attribute          | Baseline (64x32)               | Upgraded 4x (256x128)         |
+------------------------------+--------------------------------+-------------------------------+
| Floor Tile Texture Size      | 64 x 32 px                     | 256 x 128 px                  |
| Vertical Obstacles / Props   | 64 x 64 px                     | 256 x 256 px                  |
| Character Entities           | 16 x 32 px                     | 64 x 128 px                   |
| Item / Weapon / Armor Icons  | 16 x 16 px                     | 64 x 64 px                    |
| world.TileSize Constant      | 32                             | 128                           |
| Tile Center Offset           | +16.0, +16.0                   | +64.0, +64.0                  |
| Ground Tile Draw Anchor      | (-32, 0)                       | (-128, 0)                     |
| Obstacle / Wall Draw Anchor  | (-32, -32)                     | (-128, -128)                  |
| Character Entity Draw Anchor | (-8, -32)                      | (-32, -128)                   |
| Item Sprite Draw Anchor      | (-8, -8)                       | (-32, -32)                    |
| Facing Indicator Draw Anchor | (-4, -4), Scale 0.25           | (-16, -16), Scale 0.25 (or 1) |
| Facing Indicator Offset      | 20.0 px                        | 80.0 px                       |
| Player Collider (AABB)       | 16 x 16 px                     | 64 x 64 px                    |
| Zombie Collider (AABB)       | 16 x 16 px                     | 64 x 64 px                    |
| Player Movement Speed        | 3.0 px/frame                   | 12.0 px/frame                 |
| Zombie Movement Speed        | 1.0 ~ 1.5 px/frame             | 4.0 ~ 6.0 px/frame            |
| Runner Movement Speed        | 2.2 ~ 2.6 px/frame             | 8.8 ~ 10.4 px/frame           |
| Zombie Wander Speed          | zombie.Speed * 0.4             | zombie.Speed * 0.4            |
| Boid Separation Radius/Force | 20.0 px, 2.0                   | 80.0 px, 8.0                  |
| Shove Knockback Force        | 5.0                            | 20.0                          |
| Item Pickup Distance         | 16.0 px                        | 64.0 px                       |
| Zombie Bite Contact Distance | 14.0 px                        | 56.0 px                       |
| Fire Axe Cleave Reach/Radius | 32.0 px / 32.0 px              | 128.0 px / 128.0 px           |
| Bat / Club / Shove Reach     | 24.0 px / 24.0 px              | 96.0 px / 96.0 px             |
| Shotgun Range & Point-blank  | 160.0 px, 24.0 px              | 640.0 px, 96.0 px             |
| Shotgun Acoustic Pulse       | 400.0 px                       | 1600.0 px                     |
| Zombie Hearing (Idle/Moving) | 50.0 px / 200.0 px             | 200.0 px / 800.0 px           |
| Zombie Vision Detection      | 150.0 px                       | 600.0 px                      |
| Zombie De-aggro Distance     | 400.0 px                       | 1600.0 px                     |
| Safe Zombie Spawn Perimeter  | 350.0 px                       | 1400.0 px                     |
| Render Vision Cutoff Radius  | 250.0 px                       | 1000.0 px                     |
+------------------------------+--------------------------------+-------------------------------+
```

---

### 5.1 Step-by-Step Upgrades by File

#### A. Procedural Asset Generator (`cmd/tools/genassets/main.go`)
1. **Floor Tiles**:
   - Change canvas size from `w, h := 64, 32` to `w, h := 256, 128`.
   - Scale isometric diamond equation: $\frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0} \le 1.0$.
   - Scale vector chevrons, wildflowers, pebble sizes, and plank seams proportionally by $4\times$.
2. **Vertical Obstacles / Walls / Props**:
   - Change canvas size from `w, h := 64, 64` to `w, h := 256, 256`.
   - Top face diamond: $y \in [0, 112]$, Left face: $x \in [0, 128], y \in [60+x/2, 188+x/2]$, Right face: $x \in [128, 256], y \in [124-(x-128)/2, 252-(x-128)/2]$.
3. **Character Entities**:
   - Change canvas size from `16x32` to `64x128`.
   - Scale capsule head, torso, limbs, and shadow ellipses $4\times$.
4. **Items & Equipment**:
   - Change canvas size from `16x16` to `64x64`.

#### B. World Map & Procedural Generator (`internal/game/world/map.go`)
1. **TileSize Constant**:
   - Update `const TileSize = 128` (was `32`).
2. **Spawns and Offsets**:
   - `PlayerSpawn`: `X: float64(playerTileX)*TileSize + 64.0, Y: float64(playerTileY)*TileSize + 64.0`.
   - `generateThematicLoot`: `centerX := float64(r.X+r.W/2)*TileSize + 64.0, centerY := float64(r.Y+r.H/2)*TileSize + 64.0`. Scale sub-tile loot placement offsets ($16 \to 64, 20 \to 80, 24 \to 96, 32 \to 128$).
   - `generateZombieSpawns`: `zx := float64(tx)*TileSize + 64.0, zy := float64(ty)*TileSize + 64.0`, distance check `dist < 1400.0` (was `350.0`).
3. **Raycast FOV & Collision**:
   - `CalculateFOV`: `px := int(playerX) / TileSize, py := int(playerY) / TileSize`.
   - `IsColliding`: checks `minTileX := int(rectX) / TileSize` through `maxTileX := int(rectX+rectW) / TileSize`.

#### C. Game Engine & Systems (`internal/game/game.go`)
1. **Projection Math**:
   - `WorldToIso` and `IsoToWorld` formulas remain identical:
     $$\text{isoX} = wx - wy, \quad \text{isoY} = \frac{wx + wy}{2.0}$$
     $$\text{Inverse}: wx = \text{isoY} + \frac{\text{isoX}}{2.0}, \quad wy = \text{isoY} - \frac{\text{isoX}}{2.0}$$
   - *Why?* Because when $wx, wy$ are scaled $4\times$, $\text{isoX}, \text{isoY}$ automatically scale $4\times$ ($128 \times 128 \to 256 \times 128$), maintaining perfect mathematical continuity.
2. **Player & Zombie Initialization**:
   - Player Sprite: `ecs.Sprite{W: 64, H: 128}`, Collider: `ecs.Collider{Width: 64, Height: 64}`.
   - Zombie Sprite: `ecs.Sprite{W: 64, H: 128}`, Collider: `ecs.Collider{Width: 64, Height: 64}`.
3. **Movement & Speed Coefficients**:
   - Player speed: `speed := 12.0` (was `3.0`).
   - Normal Zombie speed: `4.0 + rand.Float64()*2.0` (was $1.0 + \text{rand} \cdot 0.5$).
   - Runner Zombie speed: `8.8 + rand.Float64()*1.6` (was $2.2 + \text{rand} \cdot 0.4$).
   - Separation: `separationRadius = 80.0`, `separationForce = 8.0` (was $20.0, 2.0$).
   - Shove impulse: `zVel.X = player.FacingX * 20.0, zVel.Y = player.FacingY * 20.0` (was $5.0$).
4. **Combat & AI Ranges**:
   - Item pickup distance: `math.Sqrt(dx*dx + dy*dy) < 64.0` (was $16.0$).
   - Zombie bite contact: `dist < 56.0` (was $14.0$).
   - Fire Axe reach & radius: `player.FacingX*128.0`, cleave radius `128.0` (was $32.0$).
   - Club / Fist / Shove reach & radius: `player.FacingX*96.0`, radius `96.0` (was $24.0$).
   - Shotgun range: `640.0` px (was $160.0$), Point-blank: `96.0` px (was $24.0$).
   - Shotgun noise pulse: `1600.0` px (was $400.0$).
   - Zombie hearing: `200.0` px (idle) / `800.0` px (moving) (was $50.0 / 200.0$).
   - Zombie vision: `600.0` px (was $150.0$).
   - Zombie de-aggro: `1600.0` px (was $400.0$).
5. **Draw Anchors in `DrawSystem`**:
   - Ground tiles: `drawX := isoX - 128 - camX`, `drawY := isoY - 0 - camY`.
   - Vertical obstacles: `drawX := isoX - 128 - camX`, `drawY := isoY - 128 - camY`.
   - Items: `drawX := isoX - 32 - camX`, `drawY := isoY - 32 - camY`.
   - Entities: `drawX := isoX - 32 - camX`, `drawY := isoY - 128 - camY`.
   - Facing indicator: target distance $80.0\text{px}$, `drawX := isoX - 16 - camX`, `drawY := isoY - 16 - camY`.
   - Vision cutoff: `visionRadius := 1000.0` (was $250.0$).

---

## 6. Verification and Test Suite Impact Matrix

When the 256x128 upgrade is applied, the test suite must be updated to align with the new dimensions and thresholds:

| Test File | Test Name | Required Updates |
|---|---|---|
| `internal/assets/assets_test.go` | `TestEmbeddedAssetDimensionsAndValidity`, `TestAssetsLoadAllPointersNonNil` | Update expected dimensions: Floors $\to 256\times 128$, Obstacles $\to 256\times 256$, Entities $\to 64\times 128$, Items $\to 64\times 64$. |
| `internal/game/world/map_test.go` | `TestCollisionDetection`, `TestPlayerSafeSpawn`, `TestFOVAndOcclusion` | Update test coordinate inputs and distances ($350.0 \to 1400.0$, collision box coordinates $64 \to 256$, etc.). |
| `internal/game/world/world_empirical_stress_test.go` | `TestEmpirical_PlayerSpawnSafetyAndZombieDistance`, `TestEmpirical_AABBCollisionSolidVsFloor` | Update safe distance check ($350.0 \to 1400.0$) and collider test offsets. |
| `internal/game/combat_test.go` | `TestCombat_AxeCleaveMultiTargetKill`, `TestCombat_ShotgunConeReachHit`, `TestCombat_ShotgunNoisePulseAlertsSwarm`, etc. | Scale test positions, reach thresholds ($32 \to 128, 24 \to 96, 160 \to 640, 400 \to 1600$). |
| `internal/game/armor_test.go` | `TestZombieContact_UnarmoredDirectInfection`, etc. | Scale zombie contact test distance ($10.0 \to 40.0 < 56.0$). |
| `internal/game/game_stress_test.go` | `TestIsometricProjectionMathStress`, `TestIsometricRenderingAllTileTypesAndPropsStress` | Verify fuzzed transforms and update safe distance ($350.0 \to 1400.0$). |
| `internal/game/adversarial_challenger_m5_test.go` | `TestAdversarial_WeaponBreakOnShotgunBlastEmittingNoisePulse`, `TestAdversarial_InventoryManipulationUnderMaxCapacity` | Scale shotgun cone, noise pulse, and pickup test positions ($4\times$). |

---

## 7. Recommendations for Milestone Implementation

1. **M1 (Asset Generation)**:
   - Ensure all generators in `cmd/tools/genassets` produce clean anti-aliased/vector-style $256 \times 128$ floor diamonds and $256 \times 256$ obstacle cubes without hard pixel borders.
2. **M2 (Isometric Math & World Coordinates)**:
   - Update `TileSize = 128` in `internal/game/world/map.go`.
   - Update `DrawSystem` draw offsets and entity anchors in `internal/game/game.go`.
   - Update speed coefficients and combat/AI ranges in `internal/game/game.go`.
3. **M3 (Bezier Combat Curves)**:
   - Use the player's world position $(pos.X, pos.Y)$, facing vector $(FacingX, FacingY)$, and attack radius ($128.0\text{px}$ for axe) projected through `WorldToIso` to calculate control points $P_0, P_1, P_2$ for drawing the weapon swing trail swoosh in the `DrawSystem`.
