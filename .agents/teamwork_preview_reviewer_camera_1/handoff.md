# Review & Adversarial Challenge Report: Camera System QoL (Milestone 1)

**Reviewer Agent**: `teamwork_preview_reviewer_camera_1`  
**Roles**: `reviewer`, `critic`  
**Milestone**: Milestone 1 / Milestone 3 — Camera System QoL  
**Verdict**: **APPROVE**  
**Date**: 2026-08-28  

---

## 1. Observation

1. **Global 50% Zoom World Rendering & UI 1:1 Preservation (R1)**:
   - `internal/game/game.go:870-1160`: `DrawSystem.Draw` uniformly applies the $0.5$ zoom and viewport centering transformation:
     ```go
     op.GeoM.Translate(isoX-camX, isoY-camY)
     op.GeoM.Scale(0.5, 0.5)
     op.GeoM.Translate(640, 360)
     ```
     across all world sprite layers:
     - Ground diamond tiles (anchor `(-128, 0)`, lines 903-906)
     - Obstacles, walls, trees, props, ramps, elevation blocks (anchor `(-128, -128)`, lines 987-990)
     - Items and loot drops (anchor `(-32, -32)`, lines 1027-1030)
     - Entities (player, zombies, runners) (anchor `(-32, -128)`, lines 1088-1091)
     - Facing reticle indicator (lines 1149-1153)
     - Dynamic Bezier attack arcs and muzzle rays (lines 1308-1314, 1341-1354, 1419-1425) with halved stroke widths (`outerWidth * 0.5`, `coreWidth * 0.5`).
   - `internal/game/game.go:1184-1264`: UI elements (Health, Hunger, Thirst, Armor bars, Weapon text, 9-slot inventory grid, infection/death overlays) render directly to `screen` at unscaled 1:1 pixel coordinates.
   - `internal/game/game.go:154-156`: `Game.Layout` returns `1280, 720`, matching viewport center $(640.0, 360.0)$.

2. **Inverted Mouse Coordinate Math (R1)**:
   - `internal/game/game.go:198-207`:
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
   - `internal/game/game.go:420-431`: `UpdateSystem.processInputAndCombat` converts cursor `ebiten.CursorPosition()` via `ScreenToWorld(float64(mx), float64(my), camX, camY)` for left-click pathfinding/movement and right-click combat aiming.

3. **Smooth Camera Centering & Exponential Lerp (R2)**:
   - `internal/game/game.go:158-196`:
     ```go
     type Camera struct {
         X, Y             float64
         TargetX, TargetY float64
         LerpFactor       float64
         Initialized      bool
     }
     ```
     `LerpFactor` is initialized to `0.10`. `Update(targetIsoX, targetIsoY)` features sub-pixel snapping when distance $< 0.01$ px to avoid asymptotic float drift.
   - `internal/game/game.go:50-53, 117-120`: In `Game.Reset()`, `g.camera` snaps to player spawn, and the same pointer reference is shared with `g.updateSys.camera` and `g.drawSys.camera`.

4. **Vision Radius & FOV Culling Expansion (R3)**:
   - `internal/game/game.go:877`: `DrawSystem.Draw` culling distance is expanded to `visionRadius := 2200.0` px.
   - `internal/game/game.go:242`: `UpdateSystem.Update` calls `s.gameMap.CalculateFOV(pPos.X, pPos.Y, 22)` (22 tiles $= 2816.0$ world units).
   - `internal/game/game.go:671`: Zombie AI sensory acquisition `visionRadius := 600.0` in `processZombies` is preserved to maintain balanced AI aggro mechanics.

5. **Empirical Test Suite Execution**:
   - `internal/game/camera_test.go`: 12 automated unit and stress tests covering initialization, convergence dynamics, sub-pixel snapping, 5,000 roundtrip coordinate inversions, screen center invariance, shared pointer identity, FOV expansion, headless draw execution, 4-corner unprojection symmetry, movement lag/catchup, and tile click accuracy.
   - Independent test execution `CC=gcc go test -v -count=1 ./...` passed 100% across all packages (`cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`).
   - Independent build `CC=gcc go build -o /tmp/review1_bin ./cmd/game` completed with exit code 0 and zero warnings.

---

## 2. Logic Chain

1. **Bijective Inversion Correctness**:
   - Forward rendering maps isometric coordinates $(isoX, isoY)$ to screen pixels $(sx, sy)$ via:
     $$sx = (isoX - camX) \cdot 0.5 + 640.0$$
     $$sy = (isoY - camY) \cdot 0.5 + 360.0$$
   - Solving for $(isoX, isoY)$ algebraically yields:
     $$isoX = camX + \frac{sx - 640.0}{0.5} = camX + 2.0 \cdot (sx - 640.0)$$
     $$isoY = camY + \frac{sy - 360.0}{0.5} = camY + 2.0 \cdot (sy - 360.0)$$
   - Inverse isometric projection to Cartesian world space:
     $$wx = isoY + \frac{isoX}{2.0}, \quad wy = isoY - \frac{isoX}{2.0}$$
   - Because these linear operations are bijective and mathematically exact, roundtrip error is bounded strictly by IEEE-754 double precision ($< 10^{-9}$), verified across 5,000 randomized iterations in `TestCamera_ScreenToIsoAndScreenToWorldRoundtrip`.

2. **Centering and Screen Center Invariance**:
   - When the camera is centered on the player ($camX = playerIsoX, camY = playerIsoY$), clicking the center of the screen $(640, 360)$ evaluates to:
     $$mouseIsoX = camX + (640 - 640)/0.5 = playerIsoX$$
     $$mouseIsoY = camY + (360 - 360)/0.5 = playerIsoY$$
     $$(mouseWorldX, mouseWorldY) = \text{IsoToWorld}(playerIsoX, playerIsoY) = (playerX, playerY)$$
   - This invariant is preserved under all player coordinates, verified in `TestCamera_ScreenCenterInvariance`.

3. **Smooth Camera Dynamics & Convergence**:
   - The recurrence relation $d_{t+1} = d_t \cdot (1 - \lambda)$ with $\lambda = 0.10$ produces exponential decay $d_t = d_0 \cdot 0.90^t$.
   - At $t = 60$ frames (1 second at 60 TPS), $d_{60} = d_0 \cdot 0.90^{60} \approx 0.001797 \cdot d_0$.
   - A $1000\text{ px}$ offset reduces to $1.797\text{ px}$ in 1 second and snaps cleanly to the target at $< 0.01\text{ px}$, preventing infinite asymptotic oscillation or visual jitter.

4. **Zero-Pop-in Viewport Coverage**:
   - The half-diagonal distance of a $1280 \times 720$ screen is $\sqrt{640^2 + 360^2} \approx 734.3\text{ px}$ on screen, which translates to $1468.6\text{ px}$ in unscaled isometric space.
   - The expanded `visionRadius = 2200.0` px and 22-tile FOV raycast ($2816.0\text{ px}$) provides $> 700\text{ px}$ of buffer beyond the screen edge, ensuring tiles and props do not pop in or out during dynamic camera motion.

5. **Integrity Audit**:
   - No hardcoded test outputs or fake calculations were found.
   - No dummy implementations or facade mocks exist; all logic is fully integrated into the live ECS pipeline and rendering pipeline.
   - No shortcuts or external delegation bypassed the core requirements.

---

## 3. Caveats

- **Zombie AI Sight vs Visual Culling**:
  Zombie AI sensory detection distance is deliberately kept at `600.0` px (increasing to `800.0` px when the player moves). This is an intended gameplay design decision to prevent every zombie visible across the expanded 50% zoom field of view from immediately rushing the player.
- **Headless Context in Tests**:
  Rendering tests (`TestCamera_HeadlessDrawExecution`) execute against headless `ebiten.NewImage(1280, 720)` canvases, which accurately validate the transformation pipeline without requiring physical display hardware.

---

## 4. Conclusion

The camera system QoL overhaul is robust, mathematically sound, clean, and fully tested. All requirements (R1: 50% Zoom & Mouse Inversion, R2: Exponential Lerp & Centering, R3: Expanded Culling & Zero Pop-in) have been implemented and validated against the project specifications.

**Verdict: APPROVE**

---

## 5. Verification Method

To independently reproduce and verify this review:

1. **Run full test suite**:
   ```bash
   CC=gcc go test -v -count=1 ./...
   ```
   *Expected*: All tests pass with 0 failures.

2. **Run dedicated camera test suite**:
   ```bash
   CC=gcc go test -v -run TestCamera ./internal/game/...
   ```
   *Expected*: 12 test functions pass.

3. **Verify compilation**:
   ```bash
   CC=gcc go build -o /tmp/review_bin ./cmd/game
   ```
   *Expected*: Compiles cleanly with exit code 0.

4. **Invalidation conditions**:
   - Inversion error exceeding $10^{-9}$.
   - UI elements scaling down to 50% or drifting off the screen.
   - Camera snapping instantly during gameplay movement instead of smoothly lerping.
   - Ground tiles or walls popping out of existence within the visible $1280 \times 720$ canvas.
