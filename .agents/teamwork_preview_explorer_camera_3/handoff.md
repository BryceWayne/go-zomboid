# Handoff Report: Requirement 3 — Vision Radius & FOV Culling Expansion Analysis

## 1. Observation

### 1.1 Existing FOV & Raycasting Implementation
- **File**: `internal/game/world/map.go` (lines 907–947)
  - `CalculateFOV(playerX, playerY float64, radiusTiles int)`:
    - Ray count: `rays := radiusTiles * 8` (line 923).
    - Ray traversal: Steps `radiusTiles` times with unit step `dirX, dirY = cos(angle), sin(angle)` (lines 926–932).
    - Grid boundary check: Terminates ray on `tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height` (lines 935–937).
    - Vision occlusion: `m.BlocksVision(tx, ty)` checks `TileType.BlocksVision()` (`t == TileWall`), terminating rays on walls (lines 942–944).
    - Ray angular step: $\Delta \theta = \frac{2\pi}{8 \cdot R} = \frac{\pi}{4 R}$. At radius $R$, arc length between rays is $R \cdot \Delta \theta = \frac{\pi}{4} \approx 0.7854$ tiles $< 1.0$ tile (guaranteeing no gaps between adjacent rays regardless of radius $R$).

### 1.2 Existing Call Sites & FOV Radius
- **File**: `internal/game/game.go` (line 178)
  - In `UpdateSystem.Update()`:
    ```go
    pq := s.playerFilter.Query()
    for pq.Next() {
        _, pPos, _ := pq.Get()
        s.gameMap.CalculateFOV(pPos.X, pPos.Y, 15) // 15 tile vision radius
    }
    ```
  - Current FOV radius is hardcoded to `15` tiles ($15 \times 128\text{px} = 1920.0\text{px}$).

### 1.3 Existing Render Culling & Distance Thresholds
- **File**: `internal/game/game.go` (lines 806–1011)
  - In `DrawSystem.Draw(screen *ebiten.Image, timeOfDay float64)`:
    - Line 806: `visionRadius := 1000.0`
    - Line 821 (Ground diamond tiles):
      ```go
      dx := worldX - playerX
      dy := worldY - playerY
      if dx*dx + dy*dy > visionRadius*visionRadius {
          continue
      }
      ```
    - Line 876 (Vertical props / walls / trees / fences):
      ```go
      dx := worldX - playerX
      dy := worldY - playerY
      if dx*dx + dy*dy > visionRadius*visionRadius {
          continue
      }
      ```
    - Line 941 (Contextual items / loot):
      ```go
      dx := iPos.X - playerX
      dy := iPos.Y - playerY
      if dx*dx + dy*dy > visionRadius*visionRadius {
          continue
      }
      ```
    - Line 1000 (Zombies & runner entities):
      ```go
      dx := pos.X - playerX
      dy := pos.Y - playerY
      if dx*dx + dy*dy > visionRadius*visionRadius {
          continue
      }
      ```
  - Also in `DrawSystem.Draw`, secondary fog-of-war visibility checks:
    - Line 826: `if !s.gameMap.Visible[idx] && !s.gameMap.Explored[idx] { continue }` (Ground tiles)
    - Line 881: `if !s.gameMap.Visible[idx] && !s.gameMap.Explored[idx] { continue }` (Wall / prop tiles)
    - Line 948: `if !s.gameMap.Visible[ty*s.gameMap.Width+tx] { continue }` (Items in darkness)
    - Line 1007: `if !s.gameMap.Visible[ty*s.gameMap.Width+tx] { continue }` (Zombies in darkness)

### 1.4 AI Detection Radius (Separate from Render Culling)
- **File**: `internal/game/game.go` (lines 600–607, 674–678)
  - In `UpdateSystem.processZombies()`:
    - Line 600–604: `noiseRadius := 200.0; if playerMoving { noiseRadius = 800.0 }; visionRadius := 600.0`
    - Line 674: `if dist < noiseRadius || dist < visionRadius { zombie.Chasing = true } else if dist > 1600.0 || playerDead { zombie.Chasing = false }`
    - This `visionRadius := 600.0` is zombie visual perception of the player, NOT screen render culling.

### 1.5 Viewport & Camera Parameters
- **Window / Screen Layout**: `1280 x 720` (defined in `Game.Layout`, `internal/game/game.go:147`).
- **Global Zoom**: 50% scale factor ($S = 0.5$) applied in `DrawSystem`.

---

## 2. Logic Chain

### 2.1 Viewport Geometry under 50% Zoom (0.5x Scale)
1. At resolution $1280 \times 720$ with zoom scale $S = 0.5$:
   - Effective visible isometric screen width: $W_{\text{iso}} = 1280 / 0.5 = 2560\text{px}$ ($\Delta \text{isoX} \in [-1280, +1280]$ from camera center).
   - Effective visible isometric screen height: $H_{\text{iso}} = 720 / 0.5 = 1440\text{px}$ ($\Delta \text{isoY} \in [-720, +720]$ from camera center).
2. Using inverse isometric projection math:
   $$\Delta wx = \Delta \text{isoY} + \frac{\Delta \text{isoX}}{2}$$
   $$\Delta wy = \Delta \text{isoY} - \frac{\Delta \text{isoX}}{2}$$
3. Computing the 4 screen corners in Cartesian world pixel space relative to camera center:
   - **Top-Left Corner** $(\Delta \text{isoX} = -1280, \Delta \text{isoY} = -720)$:
     $\Delta wx = -720 - 640 = -1360\text{px}$, $\Delta wy = -720 + 640 = -80\text{px}$.
     Cartesian distance: $\sqrt{(-1360)^2 + (-80)^2} = \sqrt{1,856,000} \approx \mathbf{1362.35\text{px}}$.
   - **Top-Right Corner** $(\Delta \text{isoX} = +1280, \Delta \text{isoY} = -720)$:
     $\Delta wx = -720 + 640 = -80\text{px}$, $\Delta wy = -720 - 640 = -1360\text{px}$.
     Cartesian distance: $\sqrt{(-80)^2 + (-1360)^2} \approx \mathbf{1362.35\text{px}}$.
   - **Bottom-Left Corner** $(\Delta \text{isoX} = -1280, \Delta \text{isoY} = +720)$:
     $\Delta wx = +720 - 640 = +80\text{px}$, $\Delta wy = +720 + 640 = +1360\text{px}$.
     Cartesian distance: $\sqrt{(80)^2 + (1360)^2} \approx \mathbf{1362.35\text{px}}$.
   - **Bottom-Right Corner** $(\Delta \text{isoX} = +1280, \Delta \text{isoY} = +720)$:
     $\Delta wx = +720 + 640 = +1360\text{px}$, $\Delta wy = +720 - 640 = +80\text{px}$.
     Cartesian distance: $\sqrt{(1360)^2 + (80)^2} \approx \mathbf{1362.35\text{px}}$.

### 2.2 Buffer Requirements for Zero-Clipping & Zero-Pop-in
1. **Sprite Dimensions**:
   - Vertical walls and props are $256 \times 256\text{px}$ anchored at $(-128, -128)$.
   - Floor diamonds are $256 \times 128\text{px}$ anchored at $(-128, 0)$.
   - Characters are $64 \times 128\text{px}$ anchored at $(-32, -128)$.
   - Sprites whose origin $(wx, wy)$ is outside the viewport can extend up to $256\text{px}$ into the visible screen.
2. **Smooth Camera Lerping Displacement**:
   - Under dynamic camera tracking (lerping at factor $\approx 0.08$ with player moving at $12\text{px/frame}$), the player position will lead the camera center by up to $150–200\text{px}$ in Cartesian coordinates.
3. **Total Minimum Culling Distance**:
   $$\text{Distance}_{\text{min}} = 1362.35\text{px (corner)} + 256\text{px (sprite bounds)} + 200\text{px (lerp lag)} = 1818.35\text{px}$$
4. **Deficiency of Current Values**:
   - Current `DrawSystem` `visionRadius := 1000.0` is strictly less than the corner distance ($1362.35\text{px}$). At 0.5x zoom, all 4 corners and the outer 30% of the screen will aggressively cull tiles, walls, zombies, and items into black voids.
   - Current `CalculateFOV` radius `15` tiles ($1920\text{px}$) covers the static corners ($1362\text{px}$), but leaves minimal margin ($< 1$ tile) when camera lerping lag and diagonal wall occlusion are combined.

### 2.3 Required Expansions
1. **`DrawSystem.Draw` `visionRadius`**:
   - Expand from `1000.0` to **`2000.0`** (or **`2200.0`** with extra safety buffer).
   - At `2000.0` px, safety margin over screen corner is $2000 - 1362.35 = 637.65\text{px}$ ($> 4.9$ tiles), completely eliminating all pop-in.
2. **`CalculateFOV` `radiusTiles` in `UpdateSystem.Update()`**:
   - Expand from `15` to **`20`** (or **`22`** tiles).
   - $20\text{ tiles} \times 128\text{px/tile} = 2560.0\text{px}$.
   - $22\text{ tiles} \times 128\text{px/tile} = 2816.0\text{px}$.
   - $20$ tiles comfortably covers the full $2560\times 1440$ isometric viewport even with maximum camera tracking lag.
   - Performance: $20 \times 8 = 160$ rays $\times 20$ steps $= 3200$ integer iterations per frame ($< 0.04\text{ms}$ in Go).

---

## 3. Caveats

1. **Zombie AI Sight vs Render Culling**:
   - `internal/game/game.go:604` (`visionRadius := 600.0`) in `processZombies()` controls when zombies visually acquire the player. Do **NOT** expand this AI radius to `2000.0`, as doing so would cause all zombies across the entire zoomed-out screen to instantly aggro and swarm the player on frame 1, ruining game balance.
2. **Test Independence**:
   - Unit tests in `internal/game/world/*_test.go` call `CalculateFOV` with explicit local parameters (`radius = 10`), testing raycasting logic independently. They do not depend on `game.go`'s default `UpdateSystem` radius.
3. **Map Boundary Safety**:
   - `CalculateFOV` and `GetTile` already contain out-of-bounds guards (`tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height`), so increasing `radiusTiles` to 20 or 22 is completely safe on maps of any size (including fallback $20\times 20$ maps).

---

## 4. Conclusion & Recommendations

### 4.1 Summary of Exact Changes Recommended

| File | Line(s) | Current Code | Recommended Replacement | Purpose |
|---|---|---|---|---|
| `internal/game/game.go` | 178 | `s.gameMap.CalculateFOV(pPos.X, pPos.Y, 15)` | `s.gameMap.CalculateFOV(pPos.X, pPos.Y, 20)` | Expands FOV raycasting radius from 15 to 20 tiles ($2560\text{px}$) to illuminate entire 0.5x viewport |
| `internal/game/game.go` | 806 | `visionRadius := 1000.0` | `visionRadius := 2000.0` | Expands screen culling threshold from 1000px to 2000px, preventing tile/entity edge pop-in |

### 4.2 Detailed Code Snippets

#### Edit 1: `internal/game/game.go` (FOV calculation in UpdateSystem)
```go
// Line 177-179:
pq := s.playerFilter.Query()
for pq.Next() {
	_, pPos, _ := pq.Get()
	s.gameMap.CalculateFOV(pPos.X, pPos.Y, 20) // 20 tile vision radius for 50% zoom
}
```

#### Edit 2: `internal/game/game.go` (Culling radius in DrawSystem)
```go
// Line 806:
visionRadius := 2000.0 // Expanded from 1000.0 to prevent pop-in on 1280x720 50% zoom viewport
```

---

## 5. Verification Method

### 5.1 Automated Unit & Integration Test Commands
Run the standard test suite:
```bash
CC=gcc go test ./...
```
All existing tests across `cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, and `internal/game/world` will pass without modification.

### 5.2 Specific Test Assertions to Verify Requirement 3
1. **FOV Visibility at Viewport Extremities**:
   - Given a player at $(10 \cdot 128, 10 \cdot 128)$ on an open map, after `CalculateFOV(px, py, 20)`, tile at $(10 \pm 15, 10)$ is `Visible == true` and `Explored == true`.
2. **Render Culling Distance**:
   - Entities and tiles at Cartesian distance $d = 1600.0\text{px}$ ($< \text{visionRadius} = 2000.0$) pass the distance check and are submitted to `sprites` Renderable slice.
   - Entities and tiles at Cartesian distance $d = 2500.0\text{px}$ ($> 2000.0$) are culled.
3. **Headless Draw Execution**:
   - Running `g.Draw(screen)` with `ebiten.NewImage(1280, 720)` completes cleanly without out-of-bounds panics.
