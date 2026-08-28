# Handoff Report: Requirement 1 — Global Camera Zoom & DrawSystem Rendering Scale

**Agent**: `teamwork_preview_explorer_camera_1`  
**Target Milestone**: Requirement 1 (Global Camera Zoom & DrawSystem Rendering Scale)  
**Date**: 2026-08-28  

---

## 1. Observation

### 1.1 Screen Layout & Dimensions
- In `cmd/game/main.go:12`:
  ```go
  ebiten.SetWindowSize(1280, 720)
  ```
- In `internal/game/game.go:146-148`:
  ```go
  func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
      return 1280, 720
  }
  ```
- The active rendering buffer passed to `Game.Draw(screen *ebiten.Image)` is **1280x720** pixels.

### 1.2 Current Hardcoded 800x600 Centering Offsets
- In `internal/game/game.go:801-804` (`DrawSystem.Draw`):
  ```go
  isoX, isoY := WorldToIso(pPos.X, pPos.Y)
  camX = isoX - 400
  camY = isoY - 300
  ```
- In `internal/game/game.go:357-360` (`UpdateSystem.processInputAndCombat`):
  ```go
  isoX, isoY := WorldToIso(pos.X, pos.Y)
  camX := isoX - 400
  camY := isoY - 300
  mx, my := ebiten.CursorPosition()
  mouseIsoX := float64(mx) + camX
  mouseIsoY := float64(my) + camY
  mouseWorldX, mouseWorldY := IsoToWorld(mouseIsoX, mouseIsoY)
  ```
- In `internal/game/game.go:1109` (Day/Night Lighting Overlay):
  ```go
  vector.DrawFilledRect(screen, 0, 0, 800, 600, color.RGBA{0, 0, 15, uint8(alpha * 255)}, false)
  ```
- *Finding*: `400` and `300` are legacy half-dimensions of an `800x600` screen. On a `1280x720` screen, the center is `(640, 360)`. When a 50% scale ($S = 0.5$) is applied, the unscaled isometric viewport size becomes $1280 / 0.5 = 2560$ by $720 / 0.5 = 1440$, with center offsets $\frac{640}{0.5} = 1280$ and $\frac{360}{0.5} = 720$.

### 1.3 World Sprites & Anchors in `DrawSystem`
- Ground Tiles (`internal/game/game.go:809-856`):
  - Texture size: 256x128. Anchor offset: `(-128, 0)`.
  - Draw position: `drawX := isoX - 128 - camX`, `drawY := isoY - 0 - camY`.
- Walls, Trees, Fences, Props (`internal/game/game.go:866-932`):
  - Texture size: 256x256. Anchor offset: `(-128, -128)`.
  - Draw position: `drawX := isoX - 128 - camX`, `drawY := isoY - 128 - camY`.
- Items (`internal/game/game.go:935-985`):
  - Texture size: 64x64. Anchor offset: `(-32, -32)`.
  - Draw position: `drawX := isoX - 32 - camX`, `drawY := isoY - 32 - camY`.
- Entities (`internal/game/game.go:988-1053`):
  - Texture size: 64x128. Anchor offset: `(-32, -128)`.
  - Draw position: `drawX := isoX - 32 - camX`, `drawY := isoY - 128 - camY`.
- Facing Reticle (`internal/game/game.go:1056-1089`):
  - Target: `(playerX + facingX*80, playerY + facingY*80)`.
  - Draw position: `drawX := isoX - 16 - camX`, `drawY := isoY - 16 - camY`.
- Bezier Attack Curves (`internal/game/game.go:1196-1360` in `DrawAttackSwingArc`):
  - World control points ($P_0, P_1, P_2$) transformed via `WorldToIso` to $(s0, s1, s2)$.
  - Screen points: `s0x - camX`, `s0y - camY`, etc. Rendered via `vector.StrokePath` and `vector.StrokeLine`.

### 1.4 HUD/UI Layer in `DrawSystem`
- In `internal/game/game.go:1112-1193`:
  - Health bar (Y: 10), Hunger bar (Y: 35), Thirst bar (Y: 55), Armor bar (Y: 75), Weapon text (Y: 95), Inventory (X: 550, Y: 30..230), Infected warning (Y: 115), Death banner (X: 350, Y: 280).
  - Drawn directly to `screen` using `vector.DrawFilledRect` and `ebitenutil.DebugPrintAt`.

### 1.5 Culling Radius
- In `internal/game/game.go:806`:
  ```go
  visionRadius := 1000.0
  ```
  Render distance culling check at lines 821, 876, 941, 1000:
  ```go
  if dx*dx + dy*dy > visionRadius*visionRadius {
      continue
  }
  ```
  At 50% scale, the viewport covers 2560x1440 isometric pixels ($\approx 1920\text{--}2200$ Cartesian world pixels from center to corners). A 1000.0 radius causes visible world popping well inside the viewport.

---

## 2. Logic Chain

1. **Screen & Viewport Geometry**:
   - Screen resolution = $1280 \times 720$. Screen center = $(640, 360)$.
   - Global scale factor $S = 0.5$ (50% zoom-out).
   - In isometric coordinate space, the visible viewport dimensions are $W_{\text{iso}} = \frac{1280}{0.5} = 2560$ and $H_{\text{iso}} = \frac{720}{0.5} = 1440$.
   - To center the camera focus point $(camFocusIsoX, camFocusIsoY)$ at screen center $(640, 360)$:
     $$\begin{aligned}
     camX &= camFocusIsoX - \frac{640}{0.5} = camFocusIsoX - 1280 \\
     camY &= camFocusIsoY - \frac{360}{0.5} = camFocusIsoY - 720
     \end{aligned}$$

2. **World Rendering Transformation with 50% Zoom**:
   Two architectural designs were analyzed:

   - **Architecture A (Direct GeoM Scaling on Viewport & Vector Coordinates — Recommended)**:
     - For any sprite with isometric position $(isoX, isoY)$ and anchor offset $(ax, ay)$:
       $$drawX = 0.5 \cdot (isoX + ax - camX)$$
       $$drawY = 0.5 \cdot (isoY + ay - camY)$$
       Set `op.GeoM.Scale(0.5, 0.5)` then `op.GeoM.Translate(drawX, drawY)`.
     - When $(isoX, isoY) = (camFocusIsoX, camFocusIsoY)$ and $(ax, ay) = (0, 0)$:
       $$drawX = 0.5 \cdot (camFocusIsoX - (camFocusIsoX - 1280)) = 0.5 \cdot 1280 = 640$$
       $$drawY = 0.5 \cdot (camFocusIsoY - (camFocusIsoY - 720)) = 0.5 \cdot 720 = 360$$
       The focus point maps exactly to the screen center $(640, 360)$.
     - For Bezier curves in `DrawAttackSwingArc`:
       Screen control points:
       $$s0x_{\text{screen}} = float32(0.5 \cdot (s0x - camX)), \quad s0y_{\text{screen}} = float32(0.5 \cdot (s0y - camY))$$
       $$s1x_{\text{screen}} = float32(0.5 \cdot (s1x - camX)), \quad s1y_{\text{screen}} = float32(0.5 \cdot (s1y - camY))$$
       $$s2x_{\text{screen}} = float32(0.5 \cdot (s2x - camX)), \quad s2y_{\text{screen}} = float32(0.5 \cdot (s2y - camY))$$
       Stroke widths: `outerWidth * 0.5`, `coreWidth * 0.5` (or adjusted for visual punch).
       Shotgun muzzle rays: `vector.StrokeLine(screen, px_screen, py_screen, rx_screen, ry_screen, 2.5 * 0.5, rayColor, true)`.

   - **Architecture B (Off-screen World Buffer `worldImage`)**:
     - Render world elements at 1:1 scale onto an intermediate `worldImage` of size $2560 \times 1440$ with $camX = camFocusIsoX - 1280, camY = camFocusIsoY - 720$.
     - Blit `worldImage` to `screen` with `GeoM.Scale(0.5, 0.5)` and `FilterLinear`.
     - Draw HUD/UI on `screen` directly.

3. **Inverted Mouse Coordinate Math**:
   - Screen cursor $(mx, my)$ in $[0..1280, 0..720]$ maps to world isometric coordinates:
     $$mx = 640 + 0.5 \cdot (mouseIsoX - camFocusIsoX) \implies mouseIsoX = camFocusIsoX + 2.0 \cdot (mx - 640) = 2.0 \cdot mx + camX$$
     $$my = 360 + 0.5 \cdot (mouseIsoY - camFocusIsoY) \implies mouseIsoY = camFocusIsoY + 2.0 \cdot (my - 360) = 2.0 \cdot my + camY$$
   - World Cartesian coordinates are obtained via standard bijective `IsoToWorld`:
     $$(mouseWorldX, mouseWorldY) = \text{IsoToWorld}(mouseIsoX, mouseIsoY)$$
     where $\text{IsoToWorld}(ix, iy) = (iy + ix/2.0, iy - ix/2.0)$.

4. **HUD / UI Separation (1:1 Scale)**:
   - All HUD elements (health/hunger/thirst/armor bars, weapon status, inventory, death screen) continue to be rendered directly onto `screen` at 1:1 scale with zero matrix scaling.
   - Day/night lighting rectangle must be updated from `800x600` to `1280x720` to ensure full screen coverage.

5. **Culling Expansion**:
   - `visionRadius` in `DrawSystem` must be expanded from `1000.0` to `2200.0` px.
   - FOV tile radius in `UpdateSystem.Update()` (`s.gameMap.CalculateFOV`) should be expanded from 15 to 22 tiles ($22 \times 128 = 2816$ px).

---

## 3. Caveats

1. **Camera Synchronization Between Systems**:
   - `UpdateSystem` (which processes mouse click movement & aiming) and `DrawSystem` (which renders the world) must use the identical camera $(camX, camY)$ coordinates.
   - When smooth camera tracking (lerping) is implemented, the camera position should be stored in `Game` or `DrawSystem`/`UpdateSystem` so both systems read the exact same interpolated camera position each frame.
2. **Headless Unit Tests**:
   - Unit tests running in CI (e.g. `combat_empirical_challenger_m4_test.go`) pass test image buffers of varying sizes (e.g. `800x600`).
   - Using dynamic screen bounds `cx := float64(screen.Bounds().Dx()) / 2.0` and `cy := float64(screen.Bounds().Dy()) / 2.0` ensures backward compatibility with all test harnesses.

---

## 4. Conclusion

- **Camera Center Calculation**:
  For 50% scale ($S = 0.5$) and 1280x720 resolution, camera top-left offsets are:
  $$camX = playerIsoX - 1280.0, \quad camY = playerIsoY - 720.0$$
- **DrawSystem Transformation**:
  Apply `op.GeoM.Scale(0.5, 0.5)` and `op.GeoM.Translate(0.5*(isoX + ax - camX), 0.5*(isoY + ay - camY))` to all ground tiles, props, items, and entities.
- **Bezier Trails Transformation**:
  Scale transformed isometric coordinates $(s0, s1, s2)$ and ray endpoints by $0.5$ relative to $(camX, camY)$.
- **Mouse Coordinate Conversion**:
  $$mouseIsoX = 2.0 \cdot mx + camX, \quad mouseIsoY = 2.0 \cdot my + camY$$
  $$mouseWorldX, mouseWorldY = \text{IsoToWorld}(mouseIsoX, mouseIsoY)$$
- **Vision Radius**:
  Expand `visionRadius` in `DrawSystem` to `2200.0` and `CalculateFOV` radius to `22` tiles.

---

## 5. Verification Method

1. **Unit Test Suite**:
   ```bash
   CC=gcc go test -v ./internal/game/... ./cmd/... ./internal/assets/...
   ```
2. **Key Math Invariant Checks**:
   - Center click test: $(mx, my) = (640, 360)$ unprojects to exact player position $(pos.X, pos.Y)$.
   - Edge click test: $(mx, my) = (1280, 720)$ unprojects to $(playerIsoX + 1280, playerIsoY + 720) \to \text{IsoToWorld}$.
   - Projection round-trip test: For any world coordinate $(wx, wy)$, mapping $(wx, wy) \to \text{WorldToIso} \to \text{ScreenPos} \to \text{MouseUnproject} \to \text{IsoToWorld}$ reconstructs $(wx, wy)$ with error $< 10^{-9}$.
3. **Rendering Verification**:
   ```bash
   CC=gcc go run ./cmd/game
   ```
   Verify 1280x720 window displays zoomed-out world with crisp 1:1 HUD/UI and accurate mouse-click targeting.
