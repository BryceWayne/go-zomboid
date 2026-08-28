# Empirical Challenge & Verification Report: Camera System QoL

**Agent**: `teamwork_preview_challenger_camera_2`  
**Milestone**: Milestone 4 — Camera System QoL (50% Zoom, Inverted Input Math, Smooth Lerping, Viewport Culling, Headless Rendering Loop)  
**Date**: 2026-08-28  
**Verdict**: **VERIFIED & PASS**

---

## 1. Observation

1. **Viewport Corner Geometry & Culling Limits**:
   - `internal/game/game.go:154-156`:
     ```go
     func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
         return 1280, 720
     }
     ```
   - In `internal/game/game.go:877, 892-894, 947-949, 1012-1014, 1072-1074`:
     ```go
     visionRadius := 2200.0
     ...
     dx := worldX - playerX
     dy := worldY - playerY
     if dx*dx + dy*dy > visionRadius*visionRadius {
         continue
     }
     ```
   - In `TestChallenger_ViewportCornerCullingDistanceAndInvariants`:
     Evaluated forward projection and inverse unprojection for all 4 viewport corners across multiple player positions:
     - Top-Left: $(0, 0) \implies (\Delta wx, \Delta wy) = (-1360.0, -80.0) \implies R = \sqrt{1856000} \approx 1362.3509\text{ px}$.
     - Top-Right: $(1280, 0) \implies (\Delta wx, \Delta wy) = (-80.0, -1360.0) \implies R = \sqrt{1856000} \approx 1362.3509\text{ px}$.
     - Bottom-Left: $(0, 720) \implies (\Delta wx, \Delta wy) = (80.0, 1360.0) \implies R = \sqrt{1856000} \approx 1362.3509\text{ px}$.
     - Bottom-Right: $(1280, 720) \implies (\Delta wx, \Delta wy) = (1360.0, 80.0) \implies R = \sqrt{1856000} \approx 1362.3509\text{ px}$.
     Each corner satisfies $R = 1362.35\text{ px} < 2200.0\text{ px}$. With a maximal $256\text{ px}$ sprite extent and $200\text{ px}$ dynamic tracking lag, the worst-case boundary distance is $1818.35\text{ px} < 2200.0\text{ px}$ (safety buffer $> 381.65\text{ px}$, $\approx 3$ tiles).
   - In `TestChallenger_CullingThresholdRadialBoundarySweep`:
     Swept 36 radial angles $\times 300$ distance steps ($0\text{ to }3000\text{ px}$). Confirmed 100% boundary fidelity: points at $d \le 2200.0\text{ px}$ pass culling; points at $d > 2200.0\text{ px}$ are culled.
   - In `TestChallenger_FOVRaycastingEnclosesViewportCulling`:
     FOV raycasting radius ($22\text{ tiles} \times 128\text{ px} = 2816.0\text{ px}$) strictly encloses the $2200.0\text{ px}$ viewport culling circle by $616.0\text{ px}$ ($4.81$ tiles).

2. **Mouse Click Tile Navigation Vector Field**:
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
   - In `TestChallenger_MouseClickTileNavigationVectorField`:
     Simulated mouse clicks over a dense $65 \times 37$ grid across the entire $[0, 1280] \times [0, 720]$ viewport canvas. Verified for every click:
     - The unprojected target coordinate $(twx, twy)$ produces a movement velocity vector $\vec{v}$ that is perfectly collinear with $(twx - px, twy - py)$ with dot product $\vec{u}_v \cdot \vec{u}_d = 1.000000 \pm 10^{-9}$.
     - Magnitude $||\vec{v}|| = \min(speed, dist)$.
     - Position step $(px + vel.X, py + vel.Y)$ strictly decreases Euclidean distance to target.
   - In `TestChallenger_MouseClickTileNavigationExactTileCenters`:
     Tested 500 distinct map tiles $(tx, ty)$. Computed their isometric screen positions, unprojected them back, and confirmed exact match to tile centers $(tx \cdot 128 + 64, ty \cdot 128 + 64)$ with error $< 10^{-9}$.
   - In `TestChallenger_ScreenCenterClickZeroVelocityInvariant`:
     Clicking the screen center $(640, 360)$ unprojects to exact player position with 0.0000 displacement and produces zero velocity (no drift or jitter).

3. **Headless Multi-Frame Simulation & Dynamic Camera Lerping**:
   - `internal/game/game.go:180-196`:
     ```go
     func (c *Camera) Update(targetIsoX, targetIsoY float64) {
         c.TargetX = targetIsoX
         c.TargetY = targetIsoY
         if !c.Initialized {
             c.Snap(targetIsoX, targetIsoY)
             return
         }
         dx := c.TargetX - c.X
         dy := c.TargetY - c.Y
         if math.Hypot(dx, dy) < 0.01 {
             c.X = c.TargetX
             c.Y = c.TargetY
             return
         }
         c.X += dx * c.LerpFactor
         c.Y += dy * c.LerpFactor
     }
     ```
   - In `TestChallenger_HeadlessMultiFrameRenderingLoopDynamicLerp`:
     Simulated 360 continuous frames of `Game.Update()` and `Game.Draw(screen)` with multi-phase dynamic trajectories (+X sprint, +Y sprint, diagonal sprint, orbital circle, and stationary catch-up).
     - Verified no panics across all 360 frames.
     - Verified zero NaN or infinite coordinates in camera state.
     - Verified exponential damping $\Delta cam \le |target - cam| \cdot 0.10$.
     - Verified that after stationary phase, camera smoothly converges and snaps to exact target ($< 10^{-4}\text{ px}$ lag).
   - In `TestChallenger_HeadlessRenderingAllCombatArcsUnderDynamicCamera`:
     Rendered axe swings, shotgun rays, club swings, and shove reticles into offscreen buffer while camera was offset/lerping across all attack cooldown frames (24 down to 16). All render operations completed without errors.

4. **Adversarial Edge Cases & Fuzzing**:
   - In `TestChallenger_SubpixelThresholdBifurcation`:
     - Target offset $= 0.008485\text{ px} < 0.01\text{ px}$: instantly snapped to target.
     - Target offset $= 0.050000\text{ px} > 0.01\text{ px}$: lerped by 10% ($X_{t+1} = X_t + \Delta X \cdot 0.10$).
   - In `TestChallenger_AdversarialFuzzingExtremeCoordinates`:
     10,000 randomized astronomical coordinates in $[-10^7, 10^7]$ confirmed exact roundtrip invertibility with residual error $< 10^{-6}$.
   - In `TestChallenger_DayNightLightingOverlayInvariance`:
     Verified lighting overlay across 48 continuous fractional hours ($0.0$ to $24.0$).

---

## 2. Logic Chain

1. **Screen Projection & Corner Derivation**:
   - Screen transformation maps world coordinate $(wx, wy)$ to $(sx, sy)$ via:
     $$isoX = wx - wy, \quad isoY = \frac{wx + wy}{2}$$
     $$sx = (isoX - camX) \cdot 0.5 + 640.0, \quad sy = (isoY - camY) \cdot 0.5 + 360.0$$
   - When camera is centered on player $(camX = playerIsoX, camY = playerIsoY)$, the four screen corners $(0, 0), (1280, 0), (0, 720), (1280, 720)$ correspond to screen offsets $(\Delta sx, \Delta sy) \in \{(\pm 640, \pm 360)\}$.
   - In isometric coordinate space: $\Delta isoX = \frac{\Delta sx}{0.5} = \pm 1280$, $\Delta isoY = \frac{\Delta sy}{0.5} = \pm 720$.
   - Solving for Cartesian world displacement:
     $$\Delta wx = \Delta isoY + \frac{\Delta isoX}{2} = \pm 720 \pm 640 = \pm 1360 \text{ or } \mp 80$$
     $$\Delta wy = \Delta isoY - \frac{\Delta isoX}{2} = \pm 720 \mp 640 = \mp 80 \text{ or } \pm 1360$$
   - The Euclidean world distance from player to any screen corner is:
     $$R = \sqrt{1360^2 + 80^2} = \sqrt{1849600 + 6400} = \sqrt{1856000} \approx 1362.3509\text{ px}$$
   - Because $1362.35\text{ px} < 2200.0\text{ px}$, any visible point within the screen boundaries is guaranteed to have distance $d \le 1362.35\text{ px} < 2200.0\text{ px}$.
   - Even when adding maximum sprite extent ($256\text{ px}$) and dynamic camera lerp lag ($200\text{ px}$), the worst-case boundary distance is $1818.35\text{ px} < 2200.0\text{ px}$.
   - Therefore, tiles, walls, props, entities, and loot never pop in or out at the screen edges.

2. **Inverted Input Math & Movement Directivity**:
   - The inverse projection is an affine bijection:
     $$mouseIsoX = camX + 2.0 \cdot (mx - 640.0)$$
     $$mouseIsoY = camY + 2.0 \cdot (my - 360.0)$$
     $$mouseWorldX = mouseIsoY + \frac{mouseIsoX}{2.0}, \quad mouseWorldY = mouseIsoY - \frac{mouseIsoX}{2.0}$$
   - For any mouse click $(mx, my)$, the displacement vector $\vec{d} = (mouseWorldX - pos.X, mouseWorldY - pos.Y)$ is aligned with the line connecting the player to the clicked world position.
   - The movement velocity $\vec{v} = \frac{\vec{d}}{||\vec{d}||} \cdot speed$ moves the player directly toward the target with unit correlation $\vec{u}_v \cdot \vec{u}_d = 1.000000$.

3. **Camera Damping & Asymptotic Convergence**:
   - Exponential smoothing with $\alpha = 0.10$ reduces distance to target by factor of $(1 - 0.10) = 0.90$ per tick.
   - For an initial displacement $D_0$, the distance after $N$ ticks is $D_N = D_0 \cdot 0.90^N$.
   - When $D_N < 0.01\text{ px}$, sub-pixel snapping terminates the tail and sets $(c.X, c.Y) = (c.TargetX, c.TargetY)$.
   - All empirical test runs confirm stable convergence without overshoot, oscillation, or NaN propagation.

---

## 3. Caveats

- **Zombie AI Vision**:
  `visionRadius := 600.0` in `processZombies()` is decoupled from the rendering viewport culling radius ($2200.0\text{ px}$). This is an intentional gameplay balancing decision so zombies do not swarm the player from across the entire expanded screen.
- No other caveats.

---

## 4. Conclusion

- **Verdict**: **PASS**
- Viewport boundary culling, mouse click tile navigation vector alignment, dynamic camera lerping, and multi-frame headless rendering have been empirically stress-tested and verified.
- All 8 empirical challenger test suites in `internal/game/camera_empirical_challenger_test.go` and all 12 unit tests in `internal/game/camera_test.go` pass 100%.
- Full repository test suite passes with 0 failures (`CC=gcc go test ./...`).
- Full binary compilation succeeds (`CC=gcc go build ./cmd/game`).

---

## 5. Verification Method

To independently reproduce and verify all empirical challenger tests:

```bash
# 1. Run all tests in the repository
CC=gcc go test -v -count=1 ./cmd/... ./internal/...

# 2. Run dedicated camera empirical challenger test suite
CC=gcc go test -v -count=1 -run TestChallenger_ internal/game/camera_empirical_challenger_test.go internal/game/game.go internal/game/bezier_combat_test.go

# 3. Build executable
CC=gcc go build -o /tmp/game_test ./cmd/game && rm -f /tmp/game_test
```

### Invalidation Conditions
- Viewport corner world distance $> 2200.0\text{ px}$.
- Click movement velocity vector dot product with target displacement $< 0.999999$.
- Panics during `Game.Draw(screen)` during camera lerping.
- Camera position containing NaN or Inf.
