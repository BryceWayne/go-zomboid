# Project: go-zomboid (Milestone 2 - 4x High-Fidelity & Bezier Combat Upgrade)

## Architecture
`go-zomboid` is a pure Go 2.5D isometric survival game built on Ebitengine v2 and Ark ECS.
- Coordinate Pipeline:
  1. Tile Grid Space `(tx, ty)`: Discrete grid cell indices `[0..Map.Width-1, 0..Map.Height-1]`.
  2. World Cartesian Space `(wx, wy)`: Continuous float coordinates for physics, colliders, ranges, and ECS positions. 1 tile cell = `TileSize * TileSize` (128x128).
  3. Isometric Screen Space `(isoX, isoY)`: 2:1 dimetric projection `isoX = wx - wy`, `isoY = (wx + wy) / 2`.
  4. Viewport Screen Space `(drawX, drawY)`: Centered on player camera `drawX = isoX - anchorX - camX`, `drawY = isoY - anchorY - camY`.
- Asset Pipeline: Pure Go procedural asset generator (`cmd/tools/genassets`) rendering vector-style anti-aliased sprites into `internal/assets/images/*.png`, embedded via `embed.FS`.
- Combat & Rendering Pipeline: Real-time attack execution in `UpdateSystem`, visualized via GPU-accelerated Bezier curve vector strokes in `DrawSystem` using `github.com/hajimehoshi/ebiten/v2/vector`.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | 4x Floor Tile Scaling (256x128) | Quadruple base floor tiles (`grass`, `dirt`, `wood`, `asphalt`, `concrete`, `tile_floor`) from 64x32 to 256x128 preserving 2:1 dimetric ratio. | M1 | ORIGINAL_REQUEST §R1 |
| 2 | 4x Obstacles & Props Scaling (256x256) | Proportionally scale vertical obstacles and props (`wall`, `tree`, `fence`, `debris`, `tent`, `stump`, `mushroom`, `sign`, `elevation_block`, `elevation_ramp`) to 256x256. | M1 | ORIGINAL_REQUEST §R1 |
| 3 | 4x Character Entities Scaling (64x128) | Proportionally scale `player`, `zombie`, and `runner` sprites to 64x128 with grounding drop shadows. | M1 | ORIGINAL_REQUEST §R1 |
| 4 | 4x Items & Equipment Scaling (64x64) | Proportionally scale items (`food`, `water`, `weapon`, `axe`, `shotgun`, `ammo`, `armor`, `antidote`) to 64x64 with crisp icons. | M1 | ORIGINAL_REQUEST §R1 |
| 5 | Geometric Vector Overlays | Scale and anti-alias chevron grass tufts, wildflower clusters, pebbles, UV plank lanes & nails, asphalt yellow dashes, concrete joints, and ceramic tile grout. | M1 | ORIGINAL_REQUEST §R1 |
| 6 | Engine TileSize & Math Upgrade | Update `world.TileSize = 128`, coordinate transforms, tile centers (+64), and procedural loot placement offsets. | M2 | ORIGINAL_REQUEST §R2 |
| 7 | DrawSystem Anchors & Camera | Update draw offsets: floors `(-128, 0)`, obstacles `(-128, -128)`, entities `(-32, -128)`, items `(-32, -32)`, camera centering, and FOV render cutoff (`1000px`). | M2 | ORIGINAL_REQUEST §R2 |
| 8 | Entity Physics & Speeds | Update colliders to 64x64, player speed (12.0 px/frame), zombie speeds (4.0-6.0, runners 8.8-10.4), boid separation (80px), shove impulse (20.0). | M2 | ORIGINAL_REQUEST §R2 |
| 9 | Combat & AI Range Scaling | Scale pickup distance (64px), bite contact (56px), axe cleave (128px), bat reach (96px), shotgun range (640px), acoustic alert (1600px), safe spawn (1400px). | M2 | ORIGINAL_REQUEST §R2 |
| 10 | Bezier Attack Curve Calculation | Formulate quadratic/cubic Bezier control points ($P_0, P_1, P_2$) in world space based on player position, facing direction, weapon reach, and mouse click. | M3 | ORIGINAL_REQUEST §R3 |
| 11 | Vector Attack Swoosh Rendering | Implement multi-pass dynamic Bezier stroke rendering (`vector.Path.QuadTo`, `vector.StrokePath`) in `DrawSystem` with outer glow, bright core, and quadratic alpha fade. | M3 | ORIGINAL_REQUEST §R3 |
| 12 | Weapon-Specific Swoosh Styles | Distinct visual trail styles for Fire Axe (fiery orange/red wide cleave), Bat/Club (cool blue/white motion blur), Shove (amber shockwave), Shotgun (radial blast lines). | M3 | ORIGINAL_REQUEST §R3 |
| 13 | E2E Testing Suite (Tiers 1-4) | Comprehensive requirement-driven opaque-box test suite for assets, isometric math, movement, combat, armor, and Bezier rendering. | E2E Track | ORIGINAL_REQUEST §Acceptance Criteria |
| 14 | Integration & Adversarial Hardening | Validate 100% pass of E2E suite and perform Tier 5 white-box adversarial stress testing. | M4 | Project Pattern Final Milestone |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | High-Fidelity Asset Pipeline (4x Scale) | Features 1, 2, 3, 4, 5: `cmd/tools/genassets`, `internal/assets` sprite generation & embedding tests | none | PLANNED |
| M2 | Engine Isometric Math & Coordinate Scaling | Features 6, 7, 8, 9: `internal/game/world/map.go`, `internal/game/game.go`, player/zombie physics, colliders, combat ranges, camera, FOV | M1 | PLANNED |
| M3 | Bezier Curve Combat Dynamics in DrawSystem | Features 10, 11, 12: `internal/game/game.go` (or `render.go`), Bezier curve generation, `ebiten/v2/vector` rendering, attack visual trails | M2 | PLANNED |
| E2E | Parallel E2E Testing Track | Feature 13: Test runner, Tiers 1-4 test cases, `TEST_READY.md` publication | none | PLANNED |
| M4 | Final Validation & Adversarial Hardening | Feature 14: 100% E2E test pass + Tier 5 white-box adversarial coverage hardening | M1, M2, M3, E2E | PLANNED |

## Interface Contracts
### `cmd/tools/genassets` ↔ `internal/assets`
- Asset generation binary generates 25 PNG files into `internal/assets/images/*.png`:
  - Floors: `grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png` (Dimensions: 256x128)
  - Obstacles/Props: `wall.png`, `tree.png`, `fence.png`, `debris.png`, `tent.png`, `stump.png`, `mushroom.png`, `sign.png`, `elevation_block.png`, `elevation_ramp.png` (Dimensions: 256x256)
  - Characters: `player.png`, `zombie.png`, `runner.png` (Dimensions: 64x128)
  - Items: `food.png`, `water.png`, `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`, `antidote.png` (Dimensions: 64x64)
- `internal/assets.Load()` decodes all 25 images and creates non-nil `*ebiten.Image` pointers.

### `internal/game/world` ↔ `internal/game`
- `world.TileSize = 128` (Cartesian cell width and height).
- `world.Map.IsColliding(rectX, rectY, rectW, rectH)` tests world AABB against solid tiles using `rectX / TileSize`.
- `world.Map.PlayerSpawn` provides initial coordinates `(float64(tx)*TileSize + 64.0, float64(ty)*TileSize + 64.0)`.
- `WorldToIso(wx, wy)`: `isoX = wx - wy, isoY = (wx + wy) / 2.0`.
- `IsoToWorld(isoX, isoY)`: `wx = isoY + isoX / 2.0, wy = isoY - isoX / 2.0`.

### `internal/game.DrawSystem` ↔ Bezier Combat Swoosh
- When `player.AttackCooldown > 16` (active swing frames, duration 14 frames):
  - Normalized progress $t = (30 - \text{AttackCooldown}) / 14.0 \in [0.0, 1.0]$.
  - Quadratic Bezier control points in world space:
    - $P_0 = (P_x + R_{in} \cos(\theta - \Delta\theta/2), P_y + R_{in} \sin(\theta - \Delta\theta/2))$
    - $P_1 = (P_x + R_{apex} \cos(\theta), P_y + R_{apex} \sin(\theta))$
    - $P_2 = (P_x + R_{out} \cos(\theta + \Delta\theta/2), P_y + R_{out} \sin(\theta + \Delta\theta/2))$
  - Transformed to screen via `WorldToIso` and camera offset.
  - Rendered with `vector.Path.QuadTo` and `vector.StrokePath` using weapon-specific colors and alpha fade $\alpha = (1 - t)^2$.

## Code Layout
- `cmd/tools/genassets/main.go`: Procedural image generator for 4x resolution sprites.
- `cmd/tools/genassets/genassets_test.go`: Asset generation determinism tests.
- `internal/assets/assets.go`: Embedded FS texture and sound loader.
- `internal/assets/assets_test.go`: Asset dimensions and non-nil validation tests.
- `internal/assets/assets_stress_test.go`: Texture bounds and pixel contrast stress tests.
- `internal/game/world/map.go`: World map grid, procedural town generator, FOV, AABB collision.
- `internal/game/world/map_test.go`: Map generation, FOV, and collision unit tests.
- `internal/game/world/world_empirical_stress_test.go`: World coordinate stress tests.
- `internal/game/game.go`: Game loop, ECS systems (`UpdateSystem`, `DrawSystem`), physics, combat, Bezier drawing.
- `internal/game/combat_test.go`: Combat dynamics, durability, reach, and noise pulse tests.
- `internal/game/armor_test.go`: Armor deflection and durability decay tests.
- `internal/game/game_stress_test.go`: Projection math fuzzing and rendering stress tests.
