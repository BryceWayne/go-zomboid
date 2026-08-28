# Handoff Report — Engine Isometric Math, World Transforms, and Map Systems Survey

**Agent**: `survey_explorer_2`  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2`  
**Handoff Type**: Hard (Task complete)  
**Survey Report File**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey_report.md`

---

## 1. Observation

Direct code observations from the codebase:

1. **Tile Size & World Grid Definition**:
   - `internal/game/world/map.go:30`: `const TileSize = 32`.
   - `internal/game/world/map.go:341-342`: Player spawn coordinate formula:
     ```go
     m.PlayerSpawn = FloatPoint{
         X: float64(playerTileX)*TileSize + 16.0,
         Y: float64(playerTileY)*TileSize + 16.0,
     }
     ```
   - `internal/game/world/map.go:971-990`: `IsColliding(rectX, rectY, rectW, rectH)` uses `int(rectX) / TileSize` and `int(rectY) / TileSize`.
   - `internal/game/world/map.go:898`: Safe zombie spawn distance `dist < 350.0`.

2. **Isometric Projection & Unprojection Functions**:
   - `internal/game/game.go:744-748`:
     ```go
     func WorldToIso(wx, wy float64) (isoX, isoY float64) {
         isoX = wx - wy
         isoY = (wx + wy) / 2.0
         return
     }
     ```
   - `internal/game/game.go:750-754`:
     ```go
     func IsoToWorld(isoX, isoY float64) (wx, wy float64) {
         wx = isoY + isoX/2.0
         wy = isoY - isoX/2.0
         return
     }
     ```

3. **Camera & Viewport Translation**:
   - `internal/game/game.go:140-142`: Viewport size is `800 x 600`.
   - `internal/game/game.go:352-353, 796-797`: Camera is centered on the player in isometric space:
     ```go
     isoX, isoY := WorldToIso(pPos.X, pPos.Y)
     camX = isoX - 400
     camY = isoY - 300
     ```

4. **Sprite Anchor Offsets in DrawSystem**:
   - `internal/game/game.go:825-826`: Ground tiles ($64 \times 32$ diamond) translated by `drawX := isoX - 32 - camX`, `drawY := isoY - 0 - camY`.
   - `internal/game/game.go:909-910`: Vertical obstacles ($64 \times 64$) translated by `drawX := isoX - 32 - camX`, `drawY := isoY - 32 - camY`.
   - `internal/game/game.go:948-949`: Items ($16 \times 16$) translated by `drawX := isoX - 8 - camX`, `drawY := isoY - 8 - camY`.
   - `internal/game/game.go:1009-1010`: Characters ($16 \times 32$) translated by `drawX := isoX - 8 - camX`, `drawY := isoY - 32 - camY`.

5. **Entity Speeds, Colliders, and Physics**:
   - `internal/game/game.go:65, 98`: Colliders are `Collider{Width: 16, Height: 16}`.
   - `internal/game/game.go:263`: Player movement speed `speed := 3.0` px/frame.
   - `internal/game/game.go:80, 82`: Normal zombie speed $1.0 \sim 1.5$ px/frame, Runner zombie speed $2.2 \sim 2.6$ px/frame.
   - `internal/game/game.go:615-616`: Boid separation radius $20.0$ px, separation force $2.0$.
   - `internal/game/game.go:483, 560`: Unarmed shove knockback velocity $5.0 \times \text{Facing}$.
   - `internal/game/game.go:211`: Item pickup radius $16.0$ px.
   - `internal/game/game.go:637`: Zombie bite contact distance $14.0$ px.
   - `internal/game/game.go:433, 446, 458`: Shotgun max range $160.0$ px, point-blank $24.0$ px, acoustic noise pulse $400.0$ px.
   - `internal/game/game.go:489, 519, 549`: Axe cleave reach $32.0$ px (radius $32.0$), bat/shove reach $24.0$ px (radius $24.0$).
   - `internal/game/game.go:594, 598, 670`: Zombie hearing $50.0 / 200.0$ px, vision $150.0$ px, de-aggro $400.0$ px.
   - `internal/game/game.go:800`: Vision cutoff radius $250.0$ px.

6. **Depth Sorting**:
   - `internal/game/game.go:921, 976, 1044, 1076`: `Depth: worldX + worldY` (or `pos.X + pos.Y`).
   - `internal/game/game.go:1086-1088`: Stable sorting in ascending depth order.

---

## 2. Logic Chain

1. **Projection Geometry**:
   - In 2:1 isometric projection, a Cartesian square of size $S \times S$ in world space projects into a 2D diamond of width $2S$ and height $S$.
   - For $S = 32$, the projected diamond is $64 \times 32$, matching the current floor textures in `internal/assets/images`.
   - When the floor textures are scaled to $256 \times 128$ (4x higher resolution), the corresponding Cartesian world tile size must scale proportionally to $S = 128$ ($32 \times 4$).

2. **Mathematical Invariance of Coordinate Transforms**:
   - `WorldToIso` and `IsoToWorld` perform a linear basis transformation $\begin{pmatrix} 1 & -1 \\ 0.5 & 0.5 \end{pmatrix}$.
   - Because this linear transformation is homogeneous, scaling the world coordinates $(wx, wy)$ by $4\times$ causes the isometric screen coordinates $(isoX, isoY)$ to scale by exactly $4\times$.
   - Consequently, the mathematical formulas for `WorldToIso` and `IsoToWorld` remain structurally unchanged; they automatically scale.

3. **Speed and Physics Invariance**:
   - The time required for an entity to traverse one tile at velocity $v$ across tile size $S$ is $T = S / v$.
   - For $S = 32$ and $v_{player} = 3.0$, $T \approx 10.67$ frames per tile.
   - When $S = 128$, setting $v_{player} = 12.0$ ($3.0 \times 4$) keeps $T = 128 / 12.0 \approx 10.67$ frames per tile, ensuring identical gameplay pacing.
   - Similarly, all physical radii (colliders, attack reaches, hearing/vision radii, spawn distances) scale linearly by $4\times$.

4. **Depth Sorting Invariance**:
   - Monotonic sorting on $\text{Depth} = wx + wy$ is preserved under positive linear scaling: if $w_1 + h_1 < w_2 + h_2$, then $4(w_1 + h_1) < 4(w_2 + h_2)$.

---

## 3. Caveats

- **Test Suite Updates**: 15+ unit and stress tests across `internal/game`, `internal/game/world`, and `internal/assets` contain hardcoded numeric assertions based on the 32px tile size (e.g. `TestWorldToIso` test table, safe distance 350.0, collider bounds 16x16, asset dimensions 64x32). When implementing the 4x upgrade, these test assertions must be updated synchronously with code changes.
- **Bezier Curves (Milestone 3)**: Weapon swing curves will take control points in world/isometric coordinates; scaling to 128px tile size will naturally increase the pixel radius of the swing arc ($32\text{px} \to 128\text{px}$), which provides ample screen space for smooth quadratic/cubic Bezier curve visualization.

---

## 4. Conclusion

The engine's isometric coordinate system is mathematically clean, decoupled, and ready for the 256x128 (4x) resolution upgrade.

To complete the upgrade seamlessly without breaking map generation, entity physics, or camera tracking:
1. Update `const TileSize = 128` in `internal/game/world/map.go` and scale spawn offsets to $+64.0$ and safe perimeter to $1400.0$.
2. Update texture generation in `cmd/tools/genassets/main.go` to output $256 \times 128$ floor diamonds, $256 \times 256$ obstacle cubes, $64 \times 128$ character entities, and $64 \times 64$ items.
3. Update `DrawSystem` sprite anchors to $(-128, 0)$ for floors, $(-128, -128)$ for walls, $(-32, -128)$ for characters, and $(-32, -32)$ for items.
4. Scale player speed to $12.0$, normal zombie speed to $4.0 \sim 6.0$, runner speed to $8.8 \sim 10.4$, colliders to $64 \times 64$, and combat/AI ranges by $4\times$.
5. Update test expectations across `internal/assets/assets_test.go`, `internal/game/*_test.go`, and `internal/game/world/*_test.go`.

Full survey details are documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey_report.md`.

---

## 5. Verification Method

To verify the existing test suite and codebase integrity:
```bash
CC=gcc go test ./...
```
Expected result: All tests in all packages pass (`cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`).
