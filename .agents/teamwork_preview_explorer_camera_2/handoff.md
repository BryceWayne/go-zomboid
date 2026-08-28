# Handoff Report: Camera Zoom (50%), Mouse Input Inversion Math, and Smooth Camera Lerping

## 1. Observation

Direct code observations from the codebase:

### A. Window Resolution and Layout
- In `internal/game/game.go:146-148`:
  ```go
  func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
      return 1280, 720
  }
  ```
  The game's virtual canvas resolution is $1280 \times 720$. The screen center is at $(cx, cy) = (640.0, 360.0)$.

### B. Current Camera Position & Lack of Persistent State
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
  ```
- Neither `Game`, `UpdateSystem`, nor `DrawSystem` currently stores camera position or velocity. Both systems independently recompute `camX = isoX - 400` and `camY = isoY - 300` per-frame from the player's instantaneous position.
- The offsets $-400$ and $-300$ are legacy offsets from the original $800 \times 600$ viewport ($800/2=400, 600/2=300$) and do not center the player on the $1280 \times 720$ canvas.
- Instant snapping occurs on every movement frame without smoothing or lerping.

### C. Current Mouse Input Unprojection Math
- In `internal/game/game.go:360-364`:
  ```go
  mx, my := ebiten.CursorPosition()
  mouseIsoX := float64(mx) + camX
  mouseIsoY := float64(my) + camY
  mouseWorldX, mouseWorldY := IsoToWorld(mouseIsoX, mouseIsoY)
  ```
- In `internal/game/game.go:750-760`:
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
- In `UpdateSystem`, `mouseIsoX = float64(mx) + camX` and `mouseIsoY = float64(my) + camY` assume a 1:1 scale ($s = 1.0$) and anchor to top-left rather than inverting the $0.5$ zoom scale centered at $(640, 360)$.
- When the world rendering is scaled by $0.5$, clicking with this unscaled math results in clicks targeting coordinates off by a factor of 2.0 relative to the center of the screen.

### D. Current World Rendering Transformations in `DrawSystem`
- Ground Tiles (`internal/game/game.go:830-835`):
  ```go
  isoX, isoY := WorldToIso(worldX, worldY)
  drawX := isoX - 128 - camX
  drawY := isoY - 0 - camY
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(drawX, drawY)
  screen.DrawImage(assets.GrassImage, op)
  ```
- Obstacles/Props (`internal/game/game.go:915-920`):
  ```go
  drawX := isoX - 128 - camX
  drawY := isoY - 128 - camY
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(drawX, drawY)
  ```
- Entities (`internal/game/game.go:1015-1020`):
  ```go
  drawX := isoX - 32 - camX
  drawY := isoY - 128 - camY
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(drawX, drawY)
  ```
- Items (`internal/game/game.go:954-959`):
  ```go
  drawX := isoX - 32 - camX
  drawY := isoY - 32 - camY
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(drawX, drawY)
  ```
- Facing Indicator (`internal/game/game.go:1060-1083`):
  ```go
  isoX, isoY := WorldToIso(targetX, targetY)
  drawX := isoX - 16 - camX
  drawY := isoY - 16 - camY
  op.GeoM.Reset()
  op.GeoM.Scale(0.5, 0.5)
  op.GeoM.Translate(drawX, drawY)
  ```
- Bezier Curve Swooshes (`internal/game/game.go:1238-1240, 1340-1341`):
  ```go
  s0x, s0y := WorldToIso(p0x, p0y)
  s1x, s1y := WorldToIso(p1x, p1y)
  s2x, s2y := WorldToIso(p2x, p2y)
  var path vector.Path
  path.MoveTo(float32(s0x-camX), float32(s0y-camY))
  path.QuadTo(float32(s1x-camX), float32(s1y-camY), float32(s2x-camX), float32(s2y-camY))
  ```
- Day/Night Lighting Overlay (`internal/game/game.go:1109`):
  ```go
  vector.DrawFilledRect(screen, 0, 0, 800, 600, color.RGBA{0, 0, 15, uint8(alpha * 255)}, false)
  ```
- Vision Culling Cutoff (`internal/game/game.go:806, 821`):
  ```go
  visionRadius := 1000.0
  if dx*dx + dy*dy > visionRadius*visionRadius { continue }
  ```
- FOV Raycast Radius (`internal/game/game.go:178`):
  ```go
  s.gameMap.CalculateFOV(pPos.X, pPos.Y, 15) // 15 tile vision radius
  ```

---

## 2. Logic Chain

### A. Coordinate Space Pipeline & Forward Transformation (World -> Screen)
1. **World Space $(wx, wy)$**: Continuous float Cartesian coordinates for simulation, movement physics, colliders, and distances. $1 \text{ tile} = 128 \times 128 \text{ px}$.
2. **Isometric Projection Space $(isoX, isoY)$**: Continuous 2:1 dimetric projection space:
   $$isoX = wx - wy$$
   $$isoY = \frac{wx + wy}{2.0}$$
3. **Camera Centering in Isometric Space**: Let $(camX, camY)$ be the focal point of the camera in isometric projection coordinates (i.e. the isometric coordinates of the point that should appear exactly at the screen center).
   - Relative isometric displacement from camera:
     $$\Delta isoX = isoX - camX$$
     $$\Delta isoY = isoY - camY$$
4. **Global 50% Zoom Scaling & Screen Placement**:
   - Screen canvas dimensions: $W = 1280, H = 720$. Center: $(cx, cy) = (640.0, 360.0)$.
   - Global scale factor: $s = 0.5$.
   - Screen coordinate for any isometric point $(isoX, isoY)$:
     $$sx = (isoX - camX) \cdot s + cx = (isoX - camX) \cdot 0.5 + 640.0$$
     $$sy = (isoY - camY) \cdot s + cy = (isoY - camY) \cdot 0.5 + 360.0$$
5. **Sprite Texture Anchor Geometry (`GeoM`)**:
   For any sprite with texture anchor offset $(ox, oy)$ (e.g. floor: $(-128, 0)$, obstacle: $(-128, -128)$, entity: $(-32, -128)$, item: $(-32, -32)$):
   $$GeoM = \text{Translate}(ox, oy) \to \text{Translate}(isoX - camX, isoY - camY) \to \text{Scale}(0.5, 0.5) \to \text{Translate}(640.0, 360.0)$$
   Under this matrix chain:
   $$\begin{pmatrix} x' \\ y' \end{pmatrix} = 0.5 \cdot \begin{pmatrix} x_{texture} + ox + isoX - camX \\ y_{texture} + oy + isoY - camY \end{pmatrix} + \begin{pmatrix} 640.0 \\ 360.0 \end{pmatrix}$$
   When evaluated at the player's ground anchor $(isoX = camX, isoY = camY)$ with texture feet contact $(32, 128)$ and anchor offset $(-32, -128)$:
   $$x' = 0.5 \cdot (32 - 32 + 0) + 640.0 = 640.0$$
   $$y' = 0.5 \cdot (128 - 128 + 0) + 360.0 = 360.0$$
   The player's feet are positioned at $(640, 360)$ (the exact screen center).

### B. Inverted Mouse-Click Transformation (Screen -> World)
To accurately convert a mouse click $(mx, my) \in [0..1280, 0..720]$ to world Cartesian coordinates $(wx, wy)$:
1. Invert the screen translation and $0.5$ zoom scale to obtain isometric cursor position $(mouseIsoX, mouseIsoY)$:
   $$mx = (mouseIsoX - camX) \cdot 0.5 + 640.0 \implies mouseIsoX = camX + \frac{mx - 640.0}{0.5} = camX + 2.0 \cdot (mx - 640.0)$$
   $$my = (mouseIsoY - camY) \cdot 0.5 + 360.0 \implies mouseIsoY = camY + \frac{my - 360.0}{0.5} = camY + 2.0 \cdot (my - 360.0)$$
2. Convert $(mouseIsoX, mouseIsoY)$ to Cartesian world coordinates using the standard `IsoToWorld` bijective inverse:
   $$mouseWorldX = mouseIsoY + \frac{mouseIsoX}{2.0}$$
   $$mouseWorldY = mouseIsoY - \frac{mouseIsoX}{2.0}$$
3. **Bijective Invertibility Proof**:
   Substituting $(mouseIsoX, mouseIsoY)$ into $(mouseWorldX, mouseWorldY)$:
   $$mouseWorldX = \left[camY + 2(my - 360)\right] + \frac{camX + 2(mx - 640)}{2} = camY + \frac{camX}{2} + 2(my - 360) + (mx - 640)$$
   $$mouseWorldY = \left[camY + 2(my - 360)\right] - \frac{camX + 2(mx - 640)}{2} = camY - \frac{camX}{2} + 2(my - 360) - (mx - 640)$$
   For any world point $(wx, wy)$, forward projecting to screen $(sx, sy)$ and back-projecting yields:
   $$sx = ((wx - wy) - camX) \cdot 0.5 + 640$$
   $$sy = \left(\frac{wx + wy}{2} - camY\right) \cdot 0.5 + 360$$
   $$mouseIsoX = camX + 2(sx - 640) = wx - wy$$
   $$mouseIsoY = camY + 2(sy - 360) = \frac{wx + wy}{2}$$
   $$mouseWorldX = \frac{wx + wy}{2} + \frac{wx - wy}{2} = wx$$
   $$mouseWorldY = \frac{wx + wy}{2} - \frac{wx - wy}{2} = wy$$
   The transformation has zero algebraic drift and perfect mathematical invertibility.

### C. Smooth Camera Tracking (Lerp Dynamics) & Camera Architecture
1. **Camera Position Representation**:
   Define a persistent `Camera` struct:
   ```go
   type Camera struct {
       X, Y             float64 // Current smoothed isometric camera position
       TargetX, TargetY float64 // Target isometric camera position (player isoX, isoY)
       LerpFactor       float64 // Exponential smoothing factor per frame (default: 0.10)
       Initialized      bool    // Guard flag to prevent lerping from (0,0) on spawn
   }
   ```
2. **Per-Frame Lerp Update**:
   In each update frame ($60 \text{ TPS}$):
   $$camX_{t+1} = camX_t + (targetCamX - camX_t) \cdot \text{LerpFactor}$$
   $$camY_{t+1} = camY_t + (targetCamY - camY_t) \cdot \text{LerpFactor}$$
   - When $\text{LerpFactor} = 0.10$, at 60 FPS:
     - 1 frame (~16.6ms): 10% distance closed
     - 10 frames (~166ms): $1 - (0.9)^{10} = 65.1\%$ closed
     - 20 frames (~333ms): $1 - (0.9)^{20} = 87.8\%$ closed
     - 30 frames (~500ms): $1 - (0.9)^{30} = 95.8\%$ closed
     This creates smooth tracking when moving at 12 px/frame without lagging behind or causing jitter.
3. **Spawn / Reset Snapping**:
   On `Game.Reset()` or player spawn:
   `camera.Snap(targetIsoX, targetIsoY)` sets `camX = targetCamX` and `camY = targetCamY` immediately, preventing initial camera sliding from the origin $(0, 0)$.
4. **Synchronization between UpdateSystem & DrawSystem**:
   The `*Camera` instance must be shared between `UpdateSystem` (for mouse cursor unprojection) and `DrawSystem` (for world rendering). This ensures that while the camera is dynamically catching up to the moving player, mouse clicks continue to hit the exact visual tile clicked.

### D. Vision Radius and FOV Culling Expansion
1. At 50% scale on a $1280 \times 720$ display, the visible region covers $2560 \times 1440$ isometric pixels.
2. The viewport half-diagonal in world space is $\approx \sqrt{(1280)^2 + (720)^2} \approx 1468 \text{ px} \approx 11.5 \text{ tiles}$.
3. Expanding `visionRadius` in `DrawSystem.Draw` from $1000.0 \to 2200.0$ prevents edge-culling popping.
4. Expanding `radiusTiles` in `CalculateFOV` from $15 \to 25$ ensures raycasted lighting covers the entire zoomed viewport.
5. Updating the Day/Night lighting rectangle in `DrawSystem.Draw` from $800 \times 600 \to 1280 \times 720$ covers the entire window.

---

## 3. Caveats

1. **Backwards Compatibility with Unit Tests**: Over 15 existing unit tests construct `NewUpdateSystem(w, m)` and `NewDrawSystem(w, m)`. Modifying these constructor signatures to require `*Camera` would break compilation across the test suite. Therefore, constructors should initialize a default `Camera` internally if nil, and `Game.Reset()` should wire a shared `Camera` instance to both systems.
2. **UI / HUD Scale Invariance**: The user requirement specifically states that 50% scaling applies to the *game world rendering* (excluding UI/HUD). The UI bars, inventory hotbar, text strings, and death screen must remain drawn directly to `screen` at 1:1 scale.
3. **Sub-pixel Lerp Snapping**: If the Euclidean distance between `(camX, camY)` and `(TargetX, TargetY)` is below $0.01\text{ px}$, snap directly to target to prevent floating-point asymptotic residual oscillations.

---

## 4. Conclusion & Concrete Recommendations

### A. Recommended Concrete Code Changes

#### 1. Define `Camera` and Coordinate Helper Functions in `internal/game/game.go`
```go
type Camera struct {
	X, Y             float64
	TargetX, TargetY float64
	LerpFactor       float64
	Initialized      bool
}

func NewCamera() *Camera {
	return &Camera{
		LerpFactor: 0.10,
	}
}

func (c *Camera) Snap(isoX, isoY float64) {
	c.X = isoX
	c.Y = isoY
	c.TargetX = isoX
	c.TargetY = isoY
	c.Initialized = true
}

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

func ScreenToIso(screenX, screenY, camX, camY float64) (isoX, isoY float64) {
	isoX = camX + (screenX - 640.0) / 0.5
	isoY = camY + (screenY - 360.0) / 0.5
	return
}

func ScreenToWorld(screenX, screenY, camX, camY float64) (wx, wy float64) {
	isoX, isoY := ScreenToIso(screenX, screenY, camX, camY)
	return IsoToWorld(isoX, isoY)
}
```

#### 2. Update `UpdateSystem` and `DrawSystem` Structs
```go
type UpdateSystem struct {
	world        *arkecs.World
	gameMap      *world.Map
	camera       *Camera
	playerFilter *arkecs.Filter3[ecs.Player, ecs.Position, ecs.Velocity]
	zombieFilter *arkecs.Filter3[ecs.Zombie, ecs.Position, ecs.Velocity]
	moveFilter   *arkecs.Filter3[ecs.Position, ecs.Velocity, ecs.Collider]
	itemFilter   *arkecs.Filter2[ecs.Item, ecs.Position]
}

type DrawSystem struct {
	world      *arkecs.World
	gameMap    *world.Map
	camera     *Camera
	itemFilter *arkecs.Filter2[ecs.Item, ecs.Position]
}
```

#### 3. Update `Game.Reset()`
```go
func (g *Game) Reset() {
	g.timeOfDay = 8.0

	w := arkecs.NewWorld()
	gameMap := world.NewMap(100, 100)

	playerStartX := gameMap.PlayerSpawn.X
	playerStartY := gameMap.PlayerSpawn.Y

	// Initialize and snap camera immediately to player spawn
	playerIsoX, playerIsoY := WorldToIso(playerStartX, playerStartY)
	g.camera = NewCamera()
	g.camera.Snap(playerIsoX, playerIsoY)

	// ... entity spawning ...

	g.world = w
	g.gameMap = gameMap
	g.updateSys = NewUpdateSystem(w, gameMap)
	g.updateSys.camera = g.camera
	g.drawSys = NewDrawSystem(w, gameMap)
	g.drawSys.camera = g.camera
}
```

#### 4. Update `UpdateSystem.Update()` and `processInputAndCombat()`
```go
func (s *UpdateSystem) Update() {
	pq := s.playerFilter.Query()
	for pq.Next() {
		_, pPos, _ := pq.Get()
		targetIsoX, targetIsoY := WorldToIso(pPos.X, pPos.Y)
		if s.camera != nil {
			s.camera.Update(targetIsoX, targetIsoY)
		}
		s.gameMap.CalculateFOV(pPos.X, pPos.Y, 25) // Expanded 25 tile vision radius
	}

	s.processItems()
	s.processInputAndCombat()
	s.processZombies()
	s.processMovement()
}

// In processInputAndCombat:
camX := 0.0
camY := 0.0
if s.camera != nil {
	camX = s.camera.X
	camY = s.camera.Y
} else {
	camX, camY = WorldToIso(pos.X, pos.Y)
}

mx, my := ebiten.CursorPosition()
mouseWorldX, mouseWorldY := ScreenToWorld(float64(mx), float64(my), camX, camY)
```

#### 5. Update `DrawSystem.Draw()`
- Camera Position:
  ```go
  camX := 0.0
  camY := 0.0
  if s.camera != nil {
      camX = s.camera.X
      camY = s.camera.Y
  } else {
      camX, camY = WorldToIso(playerX, playerY)
  }
  visionRadius := 2200.0
  ```
- GeoM Matrix for Ground Tiles, Obstacles, Items, and Entities:
  ```go
  // Ground tiles:
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(-128, 0)
  op.GeoM.Translate(isoX-camX, isoY-camY)
  op.GeoM.Scale(0.5, 0.5)
  op.GeoM.Translate(640, 360)

  // Obstacles / Props:
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(-128, -128)
  op.GeoM.Translate(isoX-camX, isoY-camY)
  op.GeoM.Scale(0.5, 0.5)
  op.GeoM.Translate(640, 360)

  // Items:
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(-32, -32)
  op.GeoM.Translate(isoX-camX, isoY-camY)
  op.GeoM.Scale(0.5, 0.5)
  op.GeoM.Translate(640, 360)

  // Entities:
  op := &ebiten.DrawImageOptions{}
  op.GeoM.Translate(-32, -128)
  op.GeoM.Translate(isoX-camX, isoY-camY)
  op.GeoM.Scale(0.5, 0.5)
  op.GeoM.Translate(640, 360)
  ```
- Bezier Curve Swooshes in `DrawAttackSwingArc`:
  ```go
  screen0X := float32((s0x-camX)*0.5 + 640.0)
  screen0Y := float32((s0y-camY)*0.5 + 360.0)
  screen1X := float32((s1x-camX)*0.5 + 640.0)
  screen1Y := float32((s1y-camY)*0.5 + 360.0)
  screen2X := float32((s2x-camX)*0.5 + 640.0)
  screen2Y := float32((s2y-camY)*0.5 + 360.0)

  path.MoveTo(screen0X, screen0Y)
  path.QuadTo(screen1X, screen1Y, screen2X, screen2Y)
  ```
- Lighting Overlay:
  ```go
  if alpha > 0.05 {
      vector.DrawFilledRect(screen, 0, 0, 1280, 720, color.RGBA{0, 0, 15, uint8(alpha * 255)}, false)
  }
  ```

---

## 5. Verification Method

### A. Independent Automated Test Suite Verification
Run all existing unit and stress tests:
```bash
CC=gcc go test -v ./...
```
Expected: All tests pass with zero regressions.

### B. New Camera & Mouse Inversion Unit Test Specifications
Add dedicated unit test verifying the mathematical accuracy of `ScreenToWorld`, `ScreenToIso`, and `Camera.Update`:
1. **Mathematical Bijectivity Test**:
   - For arbitrary test coordinates $(wx, wy) \in [-10000, 10000]$ and random camera positions $(camX, camY)$:
   - Compute forward screen projection $(sx, sy) = ((wx - wy - camX)*0.5 + 640, ((wx+wy)/2 - camY)*0.5 + 360)$.
   - Compute `recoveredWx, recoveredWy := ScreenToWorld(sx, sy, camX, camY)`.
   - Assert $|recoveredWx - wx| < 10^{-9}$ and $|recoveredWy - wy| < 10^{-9}$.
2. **Camera Lerp Convergence Test**:
   - Initialize camera at $(0, 0)$. Set target to $(1000, 500)$.
   - Simulate 60 frames of `camera.Update(1000, 500)`.
   - Assert `camera.X` and `camera.Y` smoothly converge towards $(1000, 500)$ with error $< 2.0\text{ px}$.
3. **Screen Center Centering Test**:
   - When player is at $(wx, wy)$ and camera is at `WorldToIso(wx, wy)`:
   - Player's ground anchor must project to exactly $(640, 360)$.

### C. Invalidation Conditions
- Clicking on a screen tile fails to move the player toward that tile.
- Camera instantly snaps rather than lerping when player changes velocity.
- Tiles near the edge of the $1280 \times 720$ window clip/pop out abruptly.
- Any UI elements scale down to 50% rather than remaining 1:1.
