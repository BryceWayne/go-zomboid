# Comprehensive Survey Report: Combat Dynamics, Bezier Attack Arcs, DrawSystem Rendering, and Test Infrastructure

**Agent**: `survey_explorer_3`  
**Date**: 2026-08-28  
**Project**: `go-zomboid` (Go + Ebitengine v2 + Ark ECS)  
**Target Focus**: Combat dynamics, attack arc geometry, `DrawSystem` rendering pipeline, Bezier curve swoosh trails, and test infrastructure.

---

## 1. Executive Summary

This investigation surveys the combat dynamics, attack arcs, isometric rendering loop (`DrawSystem`), and automated test infrastructure of `go-zomboid`. 

Key findings:
1. **Combat System Architecture**: Combat is driven by `UpdateSystem.processInputAndCombat()` using an Ark ECS entity-component architecture. Currently, 3 weapons (`"weapon"` club/bat, `"axe"` fire axe, `"shotgun"`) plus unarmed shove (`""`) are fully implemented with distinctive range, spread angle, durability lifecycles, sound triggers, and zombie interaction rules.
2. **Current Visual Feedback**: Attack visualization is currently minimal: player sprite tinting during `AttackCooldown > 20` and a static 4x4 indicator dot rendered in front of the player.
3. **Bezier Curve Swing Trails (R3 Requirement)**: Ebitengine v2's `vector` package (`github.com/hajimehoshi/ebiten/v2/vector`) provides native `Path` operations (`MoveTo`, `QuadTo`, `CubicTo`, `LineTo`, `Close`) along with `StrokePath` and `FillPath`. By projecting world-space circular arc control points through `WorldToIso`, smooth quadratic and cubic Bezier swoosh arcs (with layered strokes, glowing cores, crescent fills, and alpha fading) can be rendered directly to the screen with zero external dependencies.
4. **Test Infrastructure**: The codebase features an extensive test suite across 4 packages (`internal/game`, `internal/game/world`, `internal/assets`, `internal/ecs`, plus `cmd/tools/genassets`). All tests execute headlessly via `CC=gcc go test ./...` in ~2.4 seconds with high coverage (96.0% in `world`, 69.2% in `game`).

---

## 2. Player Input & Combat Dynamics

### 2.1 Attack Triggering & Mouse Input
In `internal/game/game.go` (`UpdateSystem.processInputAndCombat`):
- **Input Sources**:
  - Keyboard: `ebiten.IsKeyPressed(ebiten.KeySpace)` or `ebiten.IsKeyPressed(ebiten.KeyX)`
  - Mouse: `ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)`
- **Mouse Movement vs. Aiming**:
  - `MouseButtonLeft`: Moves player toward cursor position in world space.
  - `MouseButtonRight`: Aims facing vector toward cursor position in world space and immediately triggers attack.
- **Coordinate Conversion for Mouse Aiming**:
  ```go
  isoX, isoY := WorldToIso(pos.X, pos.Y)
  camX := isoX - 400
  camY := isoY - 300
  mx, my := ebiten.CursorPosition()
  mouseIsoX := float64(mx) + camX
  mouseIsoY := float64(my) + camY
  mouseWorldX, mouseWorldY := IsoToWorld(mouseIsoX, mouseIsoY)
  ```

### 2.2 Facing Direction Calculation
The player's unit facing vector `(FacingX, FacingY)` in `ecs.Player` is updated via:
1. **Movement Direction**: When moving via WASD or Left-Click:
   ```go
   if vel.X != 0 || vel.Y != 0 {
       player.FacingX = vel.X / speed
       player.FacingY = vel.Y / speed
   }
   ```
2. **Right-Click Aiming Override**:
   ```go
   if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
       dx := mouseWorldX - pos.X
       dy := mouseWorldY - pos.Y
       dist := math.Hypot(dx, dy)
       if dist > 0.001 {
           player.FacingX = dx / dist
           player.FacingY = dy / dist
       }
   }
   ```

### 2.3 Weapon Catalog & Mechanics

| Weapon | Type ID | Reach / Range | Arc / Spread | Max Durability | Special Mechanics | Audio Trigger |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Unarmed Shove** | `""` | 24.0 px | 24.0 px radius | $\infty$ | Stuns zombie (`StunTimer = 45`), knocks back 5.0 speed | `assets.ShoveSound` |
| **Spiked Bat / Club**| `"weapon"` | 24.0 px | 24.0 px single-target | 5 hits | Single-target kill, decrements durability by 1 | `assets.HitSound` (hit) / `assets.ShoveSound` (miss) |
| **Fire Axe** | `"axe"` | 32.0 px | 32.0 px multi-target cleave | 12 hits | Multi-target cleave (all zombies in radius), decrements durability by 1 | `assets.HitSound` (hit) / `assets.ShoveSound` (miss) |
| **Shotgun** | `"shotgun"` | 160.0 px | $\pm 22.5^\circ$ cone ($\cos \ge 0.9238$) | 15 hits | Requires `"ammo"`, point-blank kill $<24$px, emits 400px acoustic noise pulse alerting zombies; dry-fire fallback shoves | `assets.HitSound` (shot) / `assets.ShoveSound` (dry click) |

### 2.4 Attack Timers & Animation States
- `AttackCooldown` is set to `30` frames (0.5 seconds at 60 FPS) when an attack is initiated.
- Each tick: `if player.AttackCooldown > 0 { player.AttackCooldown-- }`.
- Current visual animation:
  - When `attackCooldown > 20` (first 10 frames of attack): Player sprite receives white highlight `op.ColorScale.Scale(2, 2, 2, 1)`.
  - A small 4x4 indicator dot is placed at `playerX + playerFacingX*20.0, playerY + playerFacingY*20.0` with weapon-specific color tint (shotgun: orange, axe: red-orange, weapon: red, shove: yellow).

---

## 3. `DrawSystem` & Rendering Loop Analysis

### 3.1 Architecture of `DrawSystem`
`DrawSystem` is located in `internal/game/game.go` (lines 730–1183) and executes sequentially on each `Game.Draw(screen)` call:

```
+-------------------------------------------------------------+
|                      DrawSystem.Draw()                      |
+-------------------------------------------------------------+
                              |
                              v
   1. Camera Centering: camX = isoX(Player) - 400, camY = isoY(Player) - 300
                              |
                              v
   2. Ground Tiles Pass: Diamond projection, memory tint for fog of war
                              |
                              v
   3. Sprite Collection & Depth Sorting:
      - Walls, trees, fences, props (Depth = worldX + worldY)
      - Ground items (Depth = itemX + itemY)
      - Entities (Player, Zombies, Runners) (Depth = posX + posY)
      - Facing indicator (Depth = targetX + targetY)
      - sort.SliceStable by Depth -> Draw in sorted order
                              |
                              v
   4. Day/Night Lighting Overlay: Fullscreen semi-transparent rectangle
                              |
                              v
   5. UI Rendering: Health, Hunger, Thirst, Armor, Weapon, Inventory (1-9)
```

### 3.2 Isometric Math & Transformations
The engine uses standard 2:1 isometric projection:
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

---

## 4. Bezier Curve Combat Dynamics Design (R3 Specification)

### 4.1 Mathematical Formulation of Weapon Swing Trails

A melee attack (such as an axe swing or bat strike) sweeps across an angular sector around the player in world space:
- Center: Player position $(P_x, P_y)$.
- Facing angle: $\theta = \text{atan2}(\text{FacingY}, \text{FacingX})$.
- Sweep span: $\Delta\theta$ (e.g., $\Delta\theta = 120^\circ = \frac{2\pi}{3}$ for Fire Axe, $90^\circ = \frac{\pi}{2}$ for Bat).
- Arc radius: $R$ (e.g., $R = 36.0$ for Axe reach).

#### 4.1.1 Quadratic Bezier Curve ($B_2$)
For a circular arc centered at $(P_x, P_y)$ with radius $R$ spanning from $\theta_0 = \theta - \frac{\Delta\theta}{2}$ to $\theta_1 = \theta + \frac{\Delta\theta}{2}$:
1. **Start Point $W_0$**:
   $$W_0 = (P_x + R \cos\theta_0, \; P_y + R \sin\theta_0)$$
2. **End Point $W_2$**:
   $$W_2 = (P_x + R \cos\theta_1, \; P_y + R \sin\theta_1)$$
3. **Control Point $W_1$**:
   $$W_1 = (P_x + R_{ctrl} \cos\theta, \; P_y + R_{ctrl} \sin\theta)$$
   where $R_{ctrl} = \frac{R}{\cos(\Delta\theta/2)}$. For $\Delta\theta = 120^\circ$, $\cos(60^\circ) = 0.5 \implies R_{ctrl} = 2.0 \cdot R$ (or $1.3 \cdot R$ for an aerodynamic flared swoosh).

#### 4.1.2 Cubic Bezier Curve ($B_3$) for Dynamic Asymmetric Swooshes
To model acceleration during the swing (fast startup, sweeping blade, tapering follow-through):
1. **Start Point $W_0$**: $(P_x + R \cos\theta_0, \; P_y + R \sin\theta_0)$
2. **First Control Handle $W_1$**: $(P_x + 1.2 R \cos(\theta - 0.2 \Delta\theta), \; P_y + 1.2 R \sin(\theta - 0.2 \Delta\theta))$
3. **Second Control Handle $W_2$**: $(P_x + 1.5 R \cos(\theta + 0.1 \Delta\theta), \; P_y + 1.5 R \sin(\theta + 0.1 \Delta\theta))$
4. **End Point $W_3$**: $(P_x + R \cos\theta_1, \; P_y + R \sin\theta_1)$

#### 4.1.3 Isometric Projection of Control Points
Because affine transformations preserve Bezier curves, projecting control points to screen coordinates preserves exact geometric fidelity:
$$S_k = \text{WorldToIso}(W_k) - \text{CameraOffset}$$

```
World Space (Circular Arc)              Screen Space (Isometric Elliptical Bezier)
          W1 (Control Point)                             S1 (Projected Control Point)
              /\                                               /\
             /  \                                             /  \
            /    \                  =======>                 /    \
           /  .   \                                         /   .  \
          / (Px,Py)\                                       / (IsoPlayer)\
        W0----------W2                                   S0--------------S2
```

---

### 4.2 Ebitengine `vector` Package Implementation

Ebitengine v2 provides high-performance GPU-accelerated vector drawing in `github.com/hajimehoshi/ebiten/v2/vector`.

#### 4.2.1 Multi-Layered Stroked Swoosh (Outer Glow + Core Blade)
```go
func drawBezierSwingTrail(screen *ebiten.Image, s0, s1, s2 [2]float64, progress float64, weaponType string) {
    // progress in [0.0, 1.0], alpha fades out quadratically
    alpha := float32(math.Max(0, 1.0 - progress))
    if alpha <= 0.01 {
        return
    }

    var path vector.Path
    path.MoveTo(float32(s0[0]), float32(s0[1]))
    path.QuadTo(float32(s1[0]), float32(s1[1]), float32(s2[0]), float32(s2[1]))

    // Weapon-specific color palettes
    var glowColor, coreColor color.RGBA
    var glowWidth, coreWidth float32

    switch weaponType {
    case "axe":
        glowColor = color.RGBA{R: 255, G: 80, B: 20, A: uint8(alpha * 160)}   // Fiery orange-red
        coreColor = color.RGBA{R: 255, G: 240, B: 200, A: uint8(alpha * 240)} // Bright sharp core
        glowWidth, coreWidth = 7.0, 2.5
    case "weapon": // Bat / Club
        glowColor = color.RGBA{R: 160, G: 200, B: 255, A: uint8(alpha * 140)} // Cool motion blur
        coreColor = color.RGBA{R: 255, G: 255, B: 255, A: uint8(alpha * 230)} // White blade
        glowWidth, coreWidth = 5.0, 2.0
    default: // Unarmed Shove
        glowColor = color.RGBA{R: 255, G: 220, B: 80, A: uint8(alpha * 120)}  // Shockwave amber
        coreColor = color.RGBA{R: 255, G: 255, B: 255, A: uint8(alpha * 200)}
        glowWidth, coreWidth = 4.0, 1.5
    }

    // 1. Draw Outer Glow / Motion Blur Pass
    vector.StrokePath(screen, &path, &vector.StrokeOptions{
        Width:    glowWidth,
        LineCap:  vector.LineCapRound,
        LineJoin: vector.LineJoinRound,
    }, glowColor, true)

    // 2. Draw Sharp Core Blade Pass
    vector.StrokePath(screen, &path, &vector.StrokeOptions{
        Width:    coreWidth,
        LineCap:  vector.LineCapRound,
        LineJoin: vector.LineJoinRound,
    }, coreColor, true)
}
```

#### 4.2.2 Filled Vector Crescent Ribbon (`FillPath`)
For a solid stylized energy swoosh ribbon, two concentric Bezier curves (outer radius $R_{out}$ and inner radius $R_{in}$) are joined:
```go
func drawFilledCrescentSwoosh(screen *ebiten.Image, out0, out1, out2, in0, in1, in2 [2]float64, clr color.RGBA) {
    var crescent vector.Path
    // Outer arc: out0 -> out1 -> out2
    crescent.MoveTo(float32(out0[0]), float32(out0[1]))
    crescent.QuadTo(float32(out1[0]), float32(out1[1]), float32(out2[0]), float32(out2[1]))
    
    // Connect to inner arc: out2 -> in2 -> in1 -> in0
    crescent.LineTo(float32(in2[0]), float32(in2[1]))
    crescent.QuadTo(float32(in1[0]), float32(in1[1]), float32(in0[0]), float32(in0[1]))
    crescent.Close()

    vector.FillPath(screen, &crescent, &vector.FillOptions{
        FillRule: vector.FillRuleNonZero,
    }, clr, true)
}
```

### 4.3 Animation Timing & Lifecycle
- When an attack occurs, `AttackCooldown` is set to 30.
- The active visual swing window is defined as `30 >= AttackCooldown >= 16` (duration = 14 frames, ~233ms).
- Parameter $progress = \frac{30 - \text{AttackCooldown}}{14.0} \in [0.0, 1.0]$.
- During $progress \in [0.0, 0.4]$, the blade rapidly extends; from $0.4 \to 1.0$, the swoosh expands and alpha fades to 0.

---

## 5. Test Infrastructure Survey

### 5.1 Test Suite Inventory

| Package | Test Files | Key Test Areas | Statements Coverage |
| :--- | :--- | :--- | :---: |
| `internal/game/world` | `map_test.go`, `world_empirical_stress_test.go` | Procedural town generation, building archetypes, FOV raycasting, AABB collision, safe spawn, zombie non-solid invariant | **96.0%** |
| `internal/game` | `game_test.go`, `game_stress_test.go`, `game_empirical_stress_test.go`, `combat_test.go`, `combat_empirical_stress_test.go`, `combat_empirical_challenger_m4_test.go`, `armor_test.go`, `armor_empirical_stress_test.go`, `armor_empirical_challenge_test.go`, `adversarial_challenger_m5_test.go` | Projection math fuzzing, 24h rendering loop stress, weapon durability lifecycles, shotgun spread cones & noise pulses, armor deflection & degradation, continuous simulation | **69.2%** |
| `internal/assets` | `assets_test.go`, `assets_stress_test.go` | Image decoding, audio initialization, memory stability | **53.3%** |
| `internal/ecs` | `components_test.go` | Component initialization and data layout | *(data structs)* |
| `cmd/tools/genassets`| `genassets_test.go` | 3-iteration asset generation determinism and SHA-256 hash validation | *(exec harness)* |

### 5.2 Test Execution Performance
- Command: `CC=gcc go test ./...`
- Total execution time: **~2.4 seconds**
- All tests execute 100% headlessly without requiring a display server (`$DISPLAY`), as Ebitengine images (`ebiten.NewImage`) and `vector` operations operate in headless memory buffer mode during tests.

---

## 6. Recommendations & Implementation Roadmap for R3

1. **Implement `DrawAttackSwingArc` in `DrawSystem` (`internal/game/game.go` or `internal/game/render.go`)**:
   - Compute the swing arc control points $(W_0, W_1, W_2)$ from `player.FacingX, player.FacingY` and player coordinates.
   - Transform control points to screen coordinates with `WorldToIso`.
   - Render multi-pass `vector.StrokePath` / `vector.FillPath` for `"axe"`, `"weapon"`, and unarmed shove.
2. **Shotgun Conical Tracers**:
   - For `"shotgun"`, render dynamic radial blast line vectors along the cone boundaries ($\pm 22.5^\circ$) with fading muzzle flash arcs during the first 6 frames of attack.
3. **Compatibility with 4x High-Resolution Tiles (R1/R2)**:
   - Ensure reach radii and stroke widths scale proportionally if `TileSize` scales from 32 to 128 (4x) so that visual weight remains crisp and aligned with high-res assets.
4. **Automated Testing**:
   - Add unit tests verifying Bezier control point calculations, angle sweep boundaries, and headless rendering execution without panics.
