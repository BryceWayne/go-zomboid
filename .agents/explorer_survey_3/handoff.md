# Handoff Report: Testing Infrastructure & Acceptance Criteria

**Author**: Explorer Survey 3 (Testing & Acceptance Criteria)  
**Date**: 2026-08-29T15:56:00Z  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/explorer_survey_3`  
**Target Milestone**: Survey / Transition Specification  
**Status**: COMPLETE  

---

## 1. Observation

### 1.1 Codebase Test Suite Inventory & Current Execution Results

An exhaustive audit of all test files across the repository was conducted using `CC=gcc go test -v ./...` and targeted package tests:

| Package | Test Files | Total Test Functions | Current Execution Status | Primary Failure / Pass Rationale |
|---|---|---|---|---|
| `internal/ecs` | `components_test.go` | 5 | **PASS** | Pure data structures for `Player`, `Zombie`, `Item`, `Position`, `Velocity`, `Sprite`, `Collider`. |
| `internal/game/world` | `map_test.go`<br>`world_empirical_stress_test.go` | 18 | **PASS** | Map generation, procedural zoning, AABB tile collision, and 2D FOV raycasting are implemented in discrete tile grid `(tx, ty)` and Cartesian coordinates `(wx, wy) = (tx*128, ty*128)`. |
| `internal/assets` | `assets_test.go`<br>`assets_stress_test.go`<br>`challenger_stress_test.go`<br>`empirical_challenger_test.go`<br>`m1_stress_verification_test.go` | 17 | **FAIL** | Tests assert legacy Milestone 2 procedural dimensions (Floors: 256x128 2:1 dimetric diamond, Obstacles: 256x256, Entities: 64x128, Items: 64x64) and diamond alpha masks. `assets.go` was updated to RPG Maker asset paths (variable sprite sizes such as 25x24, 26x25, 9x15), causing bounds mismatches (e.g. `GrassImage dimensions = 25x24, want 256x128`). |
| `internal/game` | `game_test.go`<br>`draw_depth_test.go`<br>`camera_test.go`<br>`camera_empirical_challenger_test.go`<br>`game_stress_test.go`<br>`game_empirical_stress_test.go`<br>`bezier_combat_test.go`<br>`combat_test.go`<br>`combat_empirical_stress_test.go`<br>`combat_empirical_challenger_m4_test.go`<br>`armor_test.go`<br>`armor_empirical_stress_test.go`<br>`armor_empirical_challenge_test.go`<br>`adversarial_challenger_m5_test.go`<br>`challenger_tile_render_test.go` | 64 | **PARTIAL FAIL** | Logic/combat/armor tests pass. Tests asserting `WorldToIso` projection math, isometric vertical sprite anchors `(-128, -128)`, and 256x128 diamond floor textures fail. |

#### Verbatim Failure Trace 1: Asset Dimensions and Diamond Alpha
```
=== RUN   TestDrawSystem_GroundPassUnderNewProps
    draw_depth_test.go:103: GrassImage dimensions = 25x24, want 256x128
--- FAIL: TestDrawSystem_GroundPassUnderNewProps (0.00s)

=== RUN   TestDrawSystem_SpriteGeometricAnchors
    --- FAIL: TestDrawSystem_SpriteGeometricAnchors/Wall (0.00s)
        draw_depth_test.go:54: Wall transX = -16.000000, want -128.000000
        draw_depth_test.go:57: Wall transY = 111.000000, want -128.000000
```

#### Verbatim Failure Trace 2: Assets Package Bounds Asserts
```
=== RUN   TestChallenger_AllExportedPointersAndExactBounds
    challenger_stress_test.go:142: Reader 15 detected bounds mismatch: 26x25 want 256x256
    challenger_stress_test.go:142: Reader 5 detected bounds mismatch: 31x15 want 256x128
    challenger_stress_test.go:142: Reader 2 detected bounds mismatch: 9x15 want 64x128
--- FAIL: TestChallenger_AllExportedPointersAndExactBounds (0.01s)
```

---

### 1.2 `cmd/game` Launch, Loop, Input, and Render Architecture

1. **Entry Point (`cmd/game/main.go:11-22`)**:
   - Window Dimensions: `ebiten.SetWindowSize(1280, 720)` with title `"Go Zomboid"`.
   - Asset Loader: `assets.Load()` loads embedded PNG textures from `internal/assets/images/*` via `embed.FS`.
   - Game Instance: `g := game.NewGame()` constructs the ECS world, map, systems, and camera.
   - Run: `ebiten.RunGame(g)` executes the standard Ebitengine game loop at 60 TPS / 60 FPS.

2. **Game Loop Lifecycle (`internal/game/game.go:20-156`)**:
   - `Layout(outsideWidth, outsideHeight int)`: Returns fixed logical viewport resolution `(1280, 720)`.
   - `Update() error`:
     - Advances `g.timeOfDay += 24.0 / (60.0 * 5.0 * 60.0)` (5 real-world minutes = 24 in-game hours).
     - Polls `ebiten.KeyR` for restart when dead.
     - Invokes `g.updateSys.Update()`.
   - `Draw(screen *ebiten.Image)`:
     - Fills screen background with dark base `color.RGBA{15, 15, 15, 255}`.
     - Invokes `g.drawSys.Draw(screen, g.timeOfDay)`.

3. **Subsystem Interaction & Input Pipeline**:
   - **`UpdateSystem` (`internal/game/game.go:211-800`)**:
     - `processItems()`: Picks up items within 64px radius into player inventory (max 9 slots).
     - `processInputAndCombat()`:
       - Keyboard Movement: W/S/A/D and Up/Down/Left/Right add directional velocity (`speed = 12.0 px/frame`).
       - Mouse Movement: Left click calculates vector towards unprojected world cursor `ScreenToWorld(mx, my, camX, camY)`.
       - Aiming: Right click sets `player.FacingX, player.FacingY` towards cursor.
       - Combat: Space / X / Right Click executes attack with active cooldown (30 frames). Handles Shotgun spread cone (640px, $\pm 22.5^\circ$), Fire Axe cleave (128px reach), Bat/Club melee (96px reach), and Unarmed shove (96px reach, 45 frames stun, 20px impulse).
       - Acoustic Pulse: Gunshots alert all zombies within 1600px radius (`z.Chasing = true`).
     - `processZombies()`: Boid flocking separation (80px radius), LOS / noise pursuit (200px silent, 800px moving, 600px vision), infection roll on contact (56px) against armor deflection.
     - `processMovement()`: AABB collision checks against `world.Map.IsColliding(pos.X+vel.X, pos.Y+vel.Y, col.W, col.H)`.

4. **Coordinate Transformations in Current Codebase (`internal/game/game.go:818-828`, `198-207`)**:
   ```go
   // Isometric 2:1 dimetric projection
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

5. **DrawSystem Depth Sorting & Anchors (`internal/game/game.go:879-1186`)**:
   - Ground tiles: Rendered with `GeoM.Translate(-128, 0)` followed by `GeoM.Translate(isoX-camX, isoY-camY)` scaled at 0.5.
   - Sprites (walls, props, entities, items): Rendered with anchor `GeoM.Translate(-imgW/2.0, 128.0-imgH)` (or `-32, -128` for characters).
   - Depth key: `Depth = worldX + worldY` (diagonal diamond order).

---

## 2. Logic Chain

### 2.1 Mathematical Incompatibilities of Isometric vs. 2D Orthogonal Top-Down

```
Observation: WorldToIso uses `isoX = wx - wy, isoY = (wx + wy)/2`.
Observation: DrawSystem renders floor diamonds with top-vertex anchor (-128, 0) and depth `worldX + worldY`.
Observation: ORIGINAL_REQUEST §R1 mandates strict 2D Orthogonal (top-down) grid without empty black spaces.
```

1. **Projection Math Transformation**:
   - In 2D Orthogonal space, Cartesian World space $(wx, wy)$ maps directly 1:1 to Screen/Grid space without dimetric shearing.
   - **Orthogonal World-to-Screen**:
     $$\text{screenX} = (wx - \text{camX}) \cdot S + \frac{W_{\text{viewport}}}{2}$$
     $$\text{screenY} = (wy - \text{camY}) \cdot S + \frac{H_{\text{viewport}}}{2}$$
     (where $S$ is camera zoom scale, default $1.0$ or $0.5$, and $(W_{\text{viewport}}, H_{\text{viewport}}) = (1280, 720)$).
   - **Orthogonal Screen-to-World**:
     $$wx = \text{camX} + \frac{\text{screenX} - W_{\text{viewport}}/2}{S}$$
     $$wy = \text{camY} + \frac{\text{screenY} - H_{\text{viewport}}/2}{S}$$

2. **Depth Sorting & Rendering Order**:
   - **Isometric**: Required sorting by $wx + wy$ because tiles and entities rendered diagonally from North to South.
   - **2D Orthogonal Top-Down**:
     - Ground/Floor Pass: Rendered in standard grid row/column order $(y, x)$ or as a single background pass.
     - Entity/Obstacle Pass: Sorted strictly by vertical Cartesian coordinate ($Y$ or $Y + H$) so foreground entities naturally occlude background props.

3. **Tile Anchors & Seamless Tiling**:
   - In 2D Orthogonal, adjacent tile cells $[tx \cdot \text{TileSize}, (tx+1) \cdot \text{TileSize}]$ must align edge-to-edge.
   - Any diamond offset (e.g. `-128, 0`) or 2:1 dimetric skew creates black diamond triangular gaps. Orthogonal draw operations must place tiles at top-left $(tx \cdot \text{TileSize}, ty \cdot \text{TileSize})$ with dimensions $\text{TileSize} \times \text{TileSize}$.

---

### 2.2 Detailed Inventory of All Broken Tests & Required Updates

The following table itemizes every test that fails or becomes invalid under 2D Orthogonal math:

| Test File | Test Identifier | Current (Broken) Assumption | Required Orthogonal Update |
|---|---|---|---|
| `internal/game/game_test.go` | `TestWorldToIso` | Tests $(32, 0) \to (32, 16)$ and $(0, 32) \to (-32, 16)$ (2:1 dimetric projection). | Replace with `TestWorldToOrtho` / `TestScreenToWorld` verifying orthogonal 1:1 bijective symmetry $(wx, wy) \leftrightarrow (sx, sy)$. |
| `internal/game/game_stress_test.go` | `TestIsometricProjectionMathStress` | Fuzzes $5000$ points against $isoX = wx - wy, isoY = (wx+wy)/2$. | Update to fuzz orthogonal projection and inverse unprojection roundtrip: $| \text{ScreenToWorld}(\text{WorldToScreen}(wx, wy)) - (wx, wy) | < 10^{-9}$. |
| `internal/game/draw_depth_test.go` | `TestDrawSystem_SpriteGeometricAnchors` | Asserts vertical translation `transX = -128.0, transY = -128.0` for 256x256 obstacles and `128 - imgH` for props. | Update to assert 2D top-down anchors: top-left `(0, 0)` for floor tiles and bottom-center/center `(-imgW/2, -imgH + TileSize)` for vertical obstacles. |
| `internal/game/draw_depth_test.go` | `TestDrawSystem_GroundPassUnderNewProps` | Asserts `GrassImage` is 256x128 diamond. | Update to assert RPG Maker / orthogonal tile dimensions (e.g. 128x128 or seamless tile dimensions). |
| `internal/game/draw_depth_test.go` | `TestDrawSystem_DepthSortingOrdering` | Asserts depth monotonicity via `WorldToIso` $(isoX_1+isoY_1 < isoX_2+isoY_2)$. | Update to verify Y-monotonic depth sorting: entity at $(x, y_1)$ renders before entity at $(x, y_2)$ when $y_1 < y_2$. |
| `internal/game/camera_test.go` | `TestCamera_ScreenToIsoAndScreenToWorldRoundtrip` | Tests `ScreenToIso` roundtrip using dimetric factor $0.5$. | Update to test orthogonal `ScreenToWorld` / `WorldToScreen` roundtrip across $5000$ random points. |
| `internal/game/camera_test.go` | `TestCamera_ViewportCornersUnprojection` | Asserts corner offset $\Delta isoX = \pm 1280, \Delta isoY = \pm 720$. | Update to assert orthogonal viewport deltas: $\Delta wx = (sx - 640)/S = \pm 640$, $\Delta wy = (sy - 360)/S = \pm 360$. |
| `internal/game/camera_test.go` | `TestCamera_TileClickMovementTargetingAccuracy` | Computes click unprojection via `WorldToIso`. | Update click destination targeting using orthogonal `ScreenToWorld`. |
| `internal/game/camera_test.go` | `TestCamera_GameResetSharedInstance` | Asserts `g.camera.X, g.camera.Y` snapped to `WorldToIso(PlayerSpawn)`. | Update to assert camera snaps directly to Cartesian `(PlayerSpawn.X, PlayerSpawn.Y)`. |
| `internal/game/camera_empirical_challenger_test.go` | `TestChallenger_ViewportCornerCullingDistanceAndInvariants` | Asserts isometric corner distance $\sqrt{1856000} \approx 1362.35\text{px}$. | Update to assert orthogonal corner distance $\sqrt{(640/S)^2 + (360/S)^2} \approx 734.30\text{px}$ (for $S=1.0$) or $1468.60\text{px}$ (for $S=0.5$). |
| `internal/game/camera_empirical_challenger_test.go` | `TestChallenger_MouseClickTileNavigationExactTileCenters` | Forward-projects tile centers via `WorldToIso`. | Update to orthogonal tile center projection $(tx \cdot \text{TileSize} + 64, ty \cdot \text{TileSize} + 64)$. |
| `internal/game/camera_empirical_challenger_test.go` | `TestChallenger_AdversarialFuzzingExtremeCoordinates` | Inversion checks using isometric formulas up to $\pm 10^7\text{px}$. | Update to orthogonal coordinate fuzzing up to $\pm 10^7\text{px}$. |
| `internal/game/bezier_combat_test.go` | `TestBezier_AxeControlPointsCalculation` | Lines 55-63 assert screen apex $S_1 = (140.0, 270.0)$ derived from `WorldToIso(340, 200)`. | Update screen apex assertion to orthogonal coordinates: $S_1 = (340.0, 200.0)$ in world space and transformed via `WorldToScreen`. |
| `internal/assets/assets_test.go` | `TestEmbeddedAssetDimensionsAndValidity`<br>`TestAssetsLoadAllPointersNonNil` | Asserts fixed 256x128 floors, 256x256 obstacles, 64x128 entities. | Update to validate all non-nil loaded pointers and bounds matching the current RPG Maker assets in `internal/assets/images/*`. |
| `internal/assets/assets_stress_test.go`<br>`challenger_stress_test.go`<br>`empirical_challenger_test.go` | `TestFloorTileIsometricBounds`<br>`TestEmpiricalFloorDiamondGeometry`<br>`TestEmpiricalAlphaFillRatios` | Asserts 2:1 diamond geometry and specific alpha transparency fill ratios. | Replace diamond geometry tests with orthogonal rectangular coverage tests (asserting no transparent gaps on floor boundaries). |

---

### 2.3 Dungeon Master System Testing Requirements

The Dungeon Master system specified in `ORIGINAL_REQUEST.md §R2` introduces dynamic game state variables that require dedicated test suites:

1. **Dynamic Zombie Wave Spawning**:
   - Time-indexed wave events (e.g. every $N$ ticks / minutes).
   - Difficulty scaling function: Wave size $W_k = f(k, \text{timeOfDay})$, runner ratio $R_k = g(k)$.
   - Spatial spawning safety: Waves must spawn in a ring outside current viewport/safe perimeter ($R_{\text{min}} \ge 1000\text{px}$, $R_{\text{max}} \le 2500\text{px}$) on non-solid walkable tiles.
2. **Randomized Loot Distribution**:
   - Dynamic item spawning over time in under-looted map zones/rooms.
   - Drop validation: 100% of dynamically spawned items must occupy non-solid, accessible tiles.
   - Loot table probability distribution across 8 item types (`weapon`, `axe`, `shotgun`, `ammo`, `armor`, `antidote`, `food`, `water`).
3. **Day/Night Cycle & Enemy Aggression**:
   - Ambient lighting: Alpha modulation from $0.0$ (noon) to $0.90$ (midnight).
   - Aggression scaling: Zombie speed multiplier at night (e.g. $1.25\times - 1.5\times$ base speed), expanded vision/noise detection radius during night hours ($20.0 \le t \le 24.0$ or $0.0 \le t \le 6.0$).

---

## 3. Caveats

1. **CGO and Headless Test Environment**:
   - Ebitengine and its audio dependency Oto require CGO compilation on Linux. When running `go test ./...`, the command must be prefixed with `CC=gcc` (i.e. `CC=gcc go test ./...`).
   - All tests in `internal/game` must execute headlessly by constructing `ebiten.NewImage(1280, 720)` as render targets without attempting to initialize a physical X11 window.
2. **Asset Dimensions in RPG Maker Packages**:
   - Unlike the previous procedural 4x assets which had uniform $256\times 128$ dimensions, RPG Maker tilesets have varied native dimensions ($24\times 24$, $25\times 24$, $26\times 25$, $768\times 768$ atlas sheets). Test assertions must verify non-nil pointer loading and seamless orthogonal tiling rather than hardcoding legacy $256\times 128$ constants.
3. **ECS Entity Performance During Wave Spawns**:
   - Spawning large zombie waves (e.g. 50+ entities) during live simulation must be verified to prevent entity ID leaks or unhandled nil pointer dereferences during player death or game reset.

---

## 4. Conclusion

### 4.1 Master Acceptance Criteria Checklist

| Category | Requirement Ref | Acceptance Criteria | Verification Method | Status |
|---|---|---|---|---|
| **Orthogonal Math** | `ORIGINAL_REQUEST §R1` | Core coordinate math (`WorldToScreen`, `ScreenToWorld`, `Camera.Update`) operates strictly on a 2D Orthogonal (top-down) grid. | `CC=gcc go test -v -run "TestOrthogonal|TestCamera" ./internal/game` | Specification Ready |
| **Seamless Rendering** | `ORIGINAL_REQUEST §R1` | Floor tiles tile edge-to-edge without black gaps, sheared diamond artifacts, or transparency seams. | `CC=gcc go test -v -run "TestDrawSystem|TestChallenger_Tile" ./internal/game` | Specification Ready |
| **Dynamic Wave Spawning** | `ORIGINAL_REQUEST §R2` | Dungeon Master spawns scaling zombie waves over time at safe perimeter distances ($> 1000\text{px}$) on non-solid tiles. | `CC=gcc go test -v -run "TestDungeonMaster_WaveSpawning" ./internal/game` | Specification Ready |
| **Dynamic Loot Drops** | `ORIGINAL_REQUEST §R2` | Dungeon Master dynamically distributes loot across map buildings/rooms; 100% non-solid tile placement. | `CC=gcc go test -v -run "TestDungeonMaster_LootDistribution" ./internal/game` | Specification Ready |
| **Day/Night Cycle & Aggression** | `ORIGINAL_REQUEST §R2` | Game cycles day/night with visible ambient lighting changes; zombie speed and detection radius scale up at night. | `CC=gcc go test -v -run "TestDungeonMaster_DayNightAggression" ./internal/game` | Specification Ready |
| **Full Suite Pass** | `ORIGINAL_REQUEST §Acceptance Criteria` | `CC=gcc go test ./...` exits with code 0 across all packages with 0 failures. | `CC=gcc go test ./...` | Specification Ready |
| **Application Launch** | `ORIGINAL_REQUEST §Acceptance Criteria` | `CC=gcc go run ./cmd/game` successfully initializes and launches game loop without errors. | `CC=gcc go test -run TestHeadlessGameLoop ./internal/game` | Specification Ready |

---

### 4.2 Comprehensive Test Suite Architecture Design

```
internal/
├── assets/
│   ├── assets_test.go               -> Validates non-nil loading of all 49 RPG Maker & prop assets
│   └── assets_orthogonal_test.go    -> Validates rectangular bounds and tile texture integrity
├── ecs/
│   └── components_test.go           -> Validates Player, Zombie, Item, DungeonMaster components
├── game/
│   ├── orthogonal_math_test.go      -> WorldToScreen, ScreenToWorld bijection & fuzzing
│   ├── camera_test.go               -> Camera orthogonal tracking, sub-pixel snapping, viewport corners
│   ├── draw_orthogonal_test.go      -> Seamless tiling, Y-sorted depth ordering, Bezier swoosh
│   ├── dungeon_master_test.go       -> Wave difficulty scaling, perimeter spawning, loot drops, night aggression
│   ├── combat_test.go               -> Shotgun cone, axe cleave, bat reach, armor deflection in orthogonal space
│   ├── armor_test.go                -> Armor durability decay, deflection rolls, HUD calculations
│   └── game_stress_test.go          -> 2500+ frame continuous headless simulation stress test
└── game/world/
    ├── map_test.go                  -> Procedural town generation, orthogonal AABB collision, 2D FOV raycasting
    └── world_empirical_stress_test.go -> 100% non-solid spawn invariants across 50 map generations
```

---

## 5. Verification Method

To independently reproduce observations and verify the test suite:

1. **Run Current Test Suite (Observing Current Failures)**:
   ```bash
   CC=gcc go test ./...
   ```
2. **Run Passing Map and ECS Packages**:
   ```bash
   CC=gcc go test -v ./internal/ecs ./internal/game/world
   ```
3. **Run Headless Game Simulation Smoke Test**:
   ```bash
   CC=gcc go test -v -run TestGameLoopContinuousSimulationStress ./internal/game
   ```
4. **Post-Migration Target Verification**:
   ```bash
   CC=gcc go test -v -race ./...
   ```
