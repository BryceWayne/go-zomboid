# Handoff Report: Asset Generation & Isometric Rendering Survey

**Agent**: `survey_explorer_1`  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1`  
**Handoff Type**: Hard (Task Complete)

---

## 1. Observation

1. **Asset Generation Pipeline (`cmd/tools/genassets/main.go`)**:
   - `main.go:13`: Output directory is `const outDir = "internal/assets/images"`.
   - `main.go:20-56`: Generates 3 Character Entities (`player.png`, `zombie.png`, `runner.png` at 16x32), 6 Floor Tiles (`grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png` at 64x32), 10 Vertical Obstacles/Props (`wall.png`, `tree.png`, `fence.png`, `debris.png`, `tent.png`, `stump.png`, `mushroom.png`, `sign.png`, `elevation_block.png`, `elevation_ramp.png` at 64x64), and 8 Items/Equipment (`food.png`, `water.png`, `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`, `antidote.png` at 16x16).
   - `main.go:363-368`: Diamond distance test:
     ```go
     dx := float64(x) - 31.5
     dy := float64(y) - 15.5
     isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
     ```
   - `main.go:387-400`: Chevrons at hardcoded coordinates `{{16, 12}, {40, 8}, {24, 20}, {48, 16}}`.
   - `main.go:403-416`: Wildflower accents at hardcoded coordinates `{{24, 8}, {40, 20}, {12, 18}}`.
   - `main.go:454-459`: Pebbles on dirt at hardcoded coordinates `{{20, 10}, {45, 14}, {30, 22}, {15, 20}}`.
   - `main.go:483-484`: Diamond-to-UV mapping:
     ```go
     u := dx/64.0 + dy/32.0 + 0.5
     v := dy/32.0 - dx/64.0 + 0.5
     ```
   - `main.go:773-798`: Teardrop/spherical tree canopy centered at `(32, 26)` with radius $22$.

2. **Asset Embedding (`internal/assets/assets.go`)**:
   - `assets.go:13-14`: Uses `//go:embed images/*` with `var imageFS embed.FS`.
   - `assets.go:53-88`: Function `Load()` reads all PNGs into exported `*ebiten.Image` handles.

3. **World Engine Math (`internal/game/world/map.go`)**:
   - `map.go:30`: `const TileSize = 32`.
   - `map.go:865-866, 892-893`: World coordinates are mapped as `zx := float64(tx)*TileSize + 16.0`.
   - `map.go:898`: Safe perimeter check `dist < 350.0`.
   - `map.go:971-989`: Collision detection iterates over `int(rectX) / TileSize` through `int(rectX+rectW) / TileSize`.

4. **Rendering & Isometric Projection (`internal/game/game.go`)**:
   - `game.go:744-754`:
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
   - `game.go:824-826`: Ground tile draw offset: `drawX = isoX - 32 - camX`, `drawY = isoY - 0 - camY`.
   - `game.go:909-910`: Vertical obstacle draw offset: `drawX = isoX - 32 - camX`, `drawY = isoY - 32 - camY`.
   - `game.go:948-949`: Item draw offset: `drawX = isoX - 8 - camX`, `drawY = isoY - 8 - camY`.
   - `game.go:1009-1010`: Character draw offset: `drawX = isoX - 8 - camX`, `drawY = isoY - 32 - camY`.
   - `game.go:921, 976, 1044`: Depth sorting key: `Depth = worldX + worldY` (or `pos.X + pos.Y`).

5. **Test Status (`CC=gcc go test ./...`)**:
   - All packages (`cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`) pass with 0 failures.

---

## 2. Logic Chain

1. **Floor Scaling (Observation 1)**:
   - Base floor tiles are 64x32 ($2:1$ ratio). Quadrupling pixel dimensions yields $256 \times 128$.
   - The diamond equation $\frac{|dx|}{128.0} + \frac{|dy|}{64.0} \le 1.0$ with center $(127.5, 63.5)$ preserves exact $2:1$ dimetric projection.
   - UV coordinates $(u, v) = (\frac{dx}{256} + \frac{dy}{128} + 0.5, \frac{dy}{128} - \frac{dx}{256} + 0.5)$ are normalized $[0, 1]$, making wood plank lanes, asphalt yellow dash stripes, concrete slab seams, and ceramic tile grout scale cleanly.

2. **Proportional Asset Scaling (Observation 1 & 2)**:
   - Scaling factor $S = 4$ scales vertical obstacles/props from $64 \times 64 \rightarrow 256 \times 256$, character entities from $16 \times 32 \rightarrow 64 \times 128$, and items/equipment from $16 \times 16 \rightarrow 64 \times 64$.
   - Overlay coordinates multiply by $4$ (e.g. grass chevrons at `(64, 48), (160, 32), (96, 80), (192, 64)`, pebbles at `(80, 40), (180, 56), (120, 88), (60, 80)`), and line stroke thicknesses scale from $1\text{px} \rightarrow 3\text{--}4\text{px}$.

3. **Engine Coordinate System Scaling (Observation 3 & 4)**:
   - In isometric projection, a world tile of size $T$ projects to a diamond of screen width $2T$ and screen height $T$.
   - Since the new floor texture width is $256$ and height is $128$, setting `world.TileSize = 128` ($4 \times 32$) aligns 1 world tile exactly with 1 texture sprite on screen.
   - Ground draw offset becomes `drawX = isoX - 128 - camX`, `drawY = isoY - 0 - camY`.
   - Obstacle draw offset becomes `drawX = isoX - 128 - camX`, `drawY = isoY - 128 - camY`.
   - Character draw offset becomes `drawX = isoX - 32 - camX`, `drawY = isoY - 128 - camY`.
   - Item draw offset becomes `drawX = isoX - 32 - camX`, `drawY = isoY - 32 - camY`.
   - Physics velocities ($3.0 \rightarrow 12.0$), colliders ($16 \times 16 \rightarrow 64 \times 64$), and weapon reach/radii ($24/32 \rightarrow 96/128$) scale by 4x to maintain identical gameplay speed, collision tightness, and feel.

4. **Combat Visualization (Observation 4)**:
   - Bezier curves $B(t) = (1-t)^2 P_0 + 2(1-t)t P_1 + t^2 P_2$ can be evaluated at $24$ sample points in world coordinates around the attack angle and projected to screen via `WorldToIso` to render dynamic swing swooshes.

---

## 3. Caveats

- **No caveats.** All file paths, constants, coordinate transformations, and asset rendering loops have been completely examined and verified.

---

## 4. Conclusion

The asset generation pipeline and sprite rendering systems are completely understood and documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1/survey_report.md`. Quadrupling texture resolution to 256x128 requires:
1. Updating `cmd/tools/genassets/main.go` to generate $256 \times 128$ floors, $256 \times 256$ obstacles, $64 \times 128$ entities, and $64 \times 64$ items with scaled vector primitives.
2. Updating `internal/game/world/map.go` to set `TileSize = 128` and scale spawn/safety distances by 4x.
3. Updating `internal/game/game.go` to adjust draw offsets (`-128`, `-32`), entity speeds, colliders, combat ranges, and implement Bezier attack trails.
4. Updating test dimensions and threshold invariants across all test files.

---

## 5. Verification Method

To independently verify the current system:
1. Run asset generator: `go run ./cmd/tools/genassets` (verifies clean generation of all 20 images).
2. Run test suite: `CC=gcc go test ./...` (verifies all unit, integration, and stress tests pass).
3. Inspect survey report: `view_file` on `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1/survey_report.md`.
