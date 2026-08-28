# Handoff Report: Camera System QoL (50% Zoom, Inverted Input Math, Smooth Lerping, and Culling Expansion)

**Agent**: `teamwork_preview_worker_camera_1`  
**Milestone**: Milestone 4 — Camera System QoL  
**Date**: 2026-08-28  

---

## 1. Observation

1. **Window Resolution and Viewport Center**:
   - `internal/game/game.go:146-148`:
     ```go
     func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
         return 1280, 720
     }
     ```
     The virtual canvas is $1280 \times 720$ with screen center at $(cx, cy) = (640.0, 360.0)$.
   - Legacy code previously hardcoded `-400` and `-300` offsets from an $800 \times 600$ viewport.

2. **Camera Architecture and State Management**:
   - In `internal/game/game.go:147-195`:
     Added `Camera` struct with exponential smoothing (`LerpFactor = 0.10`), sub-pixel snap threshold ($< 0.01\text{ px}$), and `Snap(targetIsoX, targetIsoY)` on spawn.
     ```go
     type Camera struct {
         X, Y             float64
         TargetX, TargetY float64
         LerpFactor       float64
         Initialized      bool
     }
     ```
   - In `Game.Reset()` (`internal/game/game.go:48-52, 108-111`):
     Camera is initialized and snapped to player spawn `(playerStartX, playerStartY)` via `WorldToIso`, and the identical `*Camera` pointer is shared across `g.camera`, `g.updateSys.camera`, and `g.drawSys.camera`.

3. **Coordinate Transformation & Mouse Inversion**:
   - In `internal/game/game.go:186-196`:
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
   - In `UpdateSystem.processInputAndCombat` (`internal/game/game.go:420-430`):
     Mouse cursor $(mx, my)$ is unprojected through `ScreenToWorld(float64(mx), float64(my), camX, camY)`, accurately targeting world coordinates under dynamic camera tracking.

4. **World Rendering 50% Zoom Matrix**:
   - In `DrawSystem.Draw` (`internal/game/game.go:865-1180`):
     - Applied `op.GeoM.Translate(ax, ay) -> op.GeoM.Translate(isoX-camX, isoY-camY) -> op.GeoM.Scale(0.5, 0.5) -> op.GeoM.Translate(640, 360)` across all world layers:
       - Ground diamond tiles (anchor: `(-128, 0)`)
       - Walls, props, trees, fences, debris, tents, elevation blocks, ramps, stumps, mushrooms, signs (anchor: `(-128, -128)`)
       - Contextual items and loot (anchor: `(-32, -32)`)
       - Entities: Player, Zombies, Runners (anchor: `(-32, -128)`)
       - Facing reticle indicator (local scale `0.5`, anchor `(-32, -64)`, zoom `0.5`, translated to `(640, 360)`)
     - In `DrawAttackSwingArc` (`internal/game/game.go:1305-1445`):
       Scaled screen points via `(iso - cam) * 0.5 + center` and stroke widths by `0.5` for Bezier motion trails and shotgun muzzle rays.
     - Lighting overlay rectangle updated to `1280, 720`.
     - UI/HUD rendering (health/hunger/thirst/armor bars, inventory, status texts) remains rendered 1:1 directly to `screen`.

5. **Vision Radius and FOV Culling**:
   - `DrawSystem.Draw` `visionRadius` expanded from `1000.0` to `2200.0` px.
   - `UpdateSystem.Update()` `s.gameMap.CalculateFOV` expanded from `15` to `22` tiles ($22 \times 128 = 2816\text{ px}$).
   - Zombie perception `visionRadius := 600.0` in `processZombies` kept unchanged.

---

## 2. Logic Chain

1. **Scale & Center Geometry**:
   At $1280 \times 720$ resolution and global scale factor $S = 0.5$, any continuous isometric world coordinate $(isoX, isoY)$ projects to screen coordinate $(sx, sy)$ via:
   $$sx = (isoX - camX) \cdot 0.5 + 640.0$$
   $$sy = (isoY - camY) \cdot 0.5 + 360.0$$
   When $(isoX, isoY) = (camX, camY)$, $(sx, sy) = (640.0, 360.0)$. Player's feet ground anchor $(-32, -128)$ on texture $(32, 128)$ aligns exactly with the screen center $(640, 360)$.

2. **Bijective Algebraic Inversion**:
   Inverting the screen transformation:
   $$mouseIsoX = camX + \frac{mx - 640.0}{0.5} = camX + 2.0 \cdot (mx - 640.0)$$
   $$mouseIsoY = camY + \frac{my - 360.0}{0.5} = camY + 2.0 \cdot (my - 360.0)$$
   $$mouseWorldX = mouseIsoY + \frac{mouseIsoX}{2.0}, \quad mouseWorldY = mouseIsoY - \frac{mouseIsoX}{2.0}$$
   Because forward and inverse transformations are linear bijections without loss of precision, roundtrip error is bounded by machine epsilon ($< 10^{-9}$).

3. **Smooth Camera Tracking Dynamics**:
   Using exponential lerp $cam_{t+1} = cam_t + (target - cam_t) \cdot 0.10$ updates camera position by 10% of remaining distance each frame (60 TPS).
   - Distance reduces by $(1 - 0.10)^{60} = 0.90^{60} \approx 0.001797$ after 60 frames.
   - A $1000\text{ px}$ target displacement converges to $< 1.8\text{ px}$ within 1 second.
   - Snapping on spawn prevents lerping from origin $(0, 0)$.
   - Sub-pixel snapping at $< 0.01\text{ px}$ eliminates floating-point residual oscillations.

4. **Zero-Pop-in Viewport Coverage**:
   - Screen half-diagonal in world Cartesian space is $\approx 1362.35\text{ px}$.
   - Accounting for $256\text{ px}$ sprite bounds and dynamic lerping lag ($\approx 200\text{ px}$), the required culling radius is $> 1818.35\text{ px}$.
   - Expanding `visionRadius` to `2200.0` px and FOV raycasting radius to `22` tiles ($2816\text{ px}$) provides a 4+ tile buffer, completely eliminating pop-in artifacts.

---

## 3. Caveats

1. **Zombie AI Sight**:
   `visionRadius := 600.0` in `processZombies()` governs zombie sensory acquisition of the player and was deliberately retained at 600.0 px. Expanding this AI value would cause all zombies across the 50% zoomed-out screen to instantly aggro and swarm the player on spawn.
2. **Headless Test Harness Image Dimensions**:
   Unit tests running with arbitrary offscreen image buffers pass cleanly because the world transformation is grounded on the screen coordinate center $(640, 360)$.

---

## 4. Conclusion

- All requirements (R1: Global Camera Zoom & Mouse Inversion, R2: Smooth Camera Tracking & Centering, R3: Vision Radius & FOV Culling Expansion) have been fully implemented with genuine mathematical logic.
- 12 comprehensive unit tests in `internal/game/camera_test.go` verify mathematical invertibility (5,000 roundtrips with $< 10^{-9}$ error), exponential convergence, sub-pixel snapping, center invariance, viewport symmetry, tracking lag & catchup, tile click accuracy, and FOV expansion.
- All tests pass across the entire repository (`CC=gcc go test ./...`).
- Binary build succeeds (`CC=gcc go build -o /tmp/game_test ./cmd/game`).

---

## 5. Verification Method

### Automated Test Commands
1. Run all unit and empirical stress tests across all packages:
   ```bash
   CC=gcc go test -v -count=1 ./...
   ```
   **Expected**: 100% pass across all packages (`cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`).

2. Run dedicated camera unit test suite:
   ```bash
   CC=gcc go test -v -run TestCamera ./internal/game/...
   ```
   **Expected**: All 12 camera unit tests pass.

3. Verify game compilation:
   ```bash
   CC=gcc go build -o /tmp/game_test ./cmd/game
   ```
   **Expected**: Exit code 0 with zero warnings.

### Invalidation Conditions
- Screen-to-world roundtrip error $> 10^{-9}$.
- Player contact feet anchor does not project to $(640, 360)$ when camera is centered on player.
- Camera instantly snaps rather than smoothly lerping when target changes.
- Tiles pop out within the $1280 \times 720$ viewport.
- UI/HUD scales down to 50% instead of remaining 1:1.
