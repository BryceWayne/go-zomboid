# Handoff Report: Combat Dynamics, Bezier Attack Arcs, DrawSystem, and Test Infrastructure

**Agent**: `survey_explorer_3`  
**Date**: 2026-08-28  
**Report Type**: Hard Handoff (Investigation Complete)  
**Detailed Survey File**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3/survey_report.md`

---

## 1. Observation

1. **Player Attack & Mouse Handling (`internal/game/game.go:350-395`)**:
   - Attack trigger: `isAttacking := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyX) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)`
   - Mouse aim:
     ```go
     isoX, isoY := WorldToIso(pos.X, pos.Y)
     camX := isoX - 400
     camY := isoY - 300
     mx, my := ebiten.CursorPosition()
     mouseIsoX := float64(mx) + camX
     mouseIsoY := float64(my) + camY
     mouseWorldX, mouseWorldY := IsoToWorld(mouseIsoX, mouseIsoY)
     ```
   - Facing direction vector is updated from movement velocity (`vel.X/speed, vel.Y/speed`) or right-click aim (`(mouseWorldX - pos.X)/dist, (mouseWorldY - pos.Y)/dist`).
   - Cooldown timer: `player.AttackCooldown = 30` (30 frames = 0.5s at 60 FPS).

2. **Weapons & Combat Mechanics (`internal/game/game.go:397-565`)**:
   - `"axe"`: Reach 32.0px, multi-target cleave radius 32.0px, 12 durability hits (`lines 487-516`).
   - `"weapon"` (bat/club): Reach 24.0px, single-target hit radius 24.0px, 5 durability hits (`lines 517-546`).
   - `"shotgun"`: Reach 160.0px, cone spread $\pm 22.5^\circ$ ($\cos \ge 0.9238$), point-blank $<24$px, 400px noise alert, 15 durability hits, requires `"ammo"` (`lines 397-486`).
   - Unarmed shove: 24.0px reach, stuns zombie for 45 frames (`StunTimer = 45`), knockback velocity $5.0 \times \text{Facing}$ (`lines 548-564`).

3. **Current Visual Rendering (`internal/game/game.go:756-1183`)**:
   - `DrawSystem.Draw()` performs:
     1. Camera centering: `camX = isoX - 400`, `camY = isoY - 300`.
     2. Ground tile pass (2:1 diamond projection).
     3. Sprite collection and depth sorting (`sort.SliceStable` by `Depth = worldX + worldY`).
     4. Current attack visual: `if attackCooldown > 20 { op.ColorScale.Scale(2, 2, 2, 1) }` (white player flash) and a 4x4 indicator dot drawn at `playerX + playerFacingX*20, playerY + playerFacingY*20` (`lines 1027-1084`).
     5. Fullscreen day/night lighting rectangle (`lines 1094-1100`).
     6. UI bars for Health, Hunger, Thirst, Armor, Weapon, and Inventory (`lines 1102-1182`).

4. **Ebitengine Vector Graphics Support (`github.com/hajimehoshi/ebiten/v2/vector`)**:
   - `vector.Path` supports `MoveTo(x, y)`, `LineTo(x, y)`, `QuadTo(x1, y1, x2, y2)`, `CubicTo(x1, y1, x2, y2, x3, y3)`, `Close()`, and `Reset()`.
   - `vector.StrokePath(dst, path, &vector.StrokeOptions{ Width, LineCap, LineJoin, MiterLimit }, clr, antialias)` draws smooth anti-aliased curves.
   - `vector.FillPath(dst, path, &vector.FillOptions{ FillRule }, clr, antialias)` fills closed vector paths.

5. **Test Infrastructure Execution (`CC=gcc go test ./...`)**:
   - `github.com/BryceWayne/go-zomboid/internal/game/world`: 96.0% coverage (15 test functions).
   - `github.com/BryceWayne/go-zomboid/internal/game`: 69.2% coverage (30+ test functions).
   - `github.com/BryceWayne/go-zomboid/internal/assets`: 53.3% coverage.
   - `cmd/tools/genassets`: Asset determinism test passes across 3 generation passes.
   - 100% of existing tests pass cleanly in headless mode without display hardware.

---

## 2. Logic Chain

1. **Requirement R3 Analysis**: The user requested dynamic weapon swing trails/arcs (swoosh effect, especially for axe swinging) using Bezier Curves in the `DrawSystem`, tracing dynamically based on player facing direction and mouse click.
2. **Mathematical Feasibility**: An attack sweep in world space is a circular arc of angle $\Delta\theta$ at radius $R$ centered at $(P_x, P_y)$. Because affine transformations preserve Bezier curves, computing world-space control points $W_0, W_1, W_2$ and projecting them via `WorldToIso(W_k) - camOffset` produces the exact elliptical isometric Bezier arc on screen.
3. **Ebitengine API Match**: `vector.Path.QuadTo` and `vector.StrokePath` / `vector.FillPath` in `ebiten/v2/vector` provide built-in, hardware-accelerated Bezier curve evaluation and rendering without external libraries.
4. **Visual Quality Enhancement**:
   - A multi-layered stroke (broad low-alpha glow + narrow high-alpha core) creates a vibrant blade trail.
   - Fading alpha quadratically $\alpha = (1 - t)^2$ over the 14-frame swing duration creates a smooth dissipation effect.
   - Weapon-specific styling (fire axe = fiery orange/crimson, bat = cool blue/white, shove = shockwave gold, shotgun = conical blast lines) provides distinct feedback for all combat interactions.
5. **Regression Safety & Testability**: Headless testing of `DrawSystem.Draw()` on an `ebiten.NewImage` validates that vector drawing operations execute safely without panicking or leaking memory.

---

## 3. Caveats

- **Scale Coupling with R1/R2**: If `TileSize` or sprite dimensions are quadrupled (64x32 $\to$ 256x128), weapon reach radii ($R=32$ px) and stroke widths (e.g. 2.5–7.0 px) should be scaled proportionally to maintain matching visual weight.
- **Render Layer Ordering**: Bezier swing trails can either be drawn before the day/night lighting overlay (so they are shaded by night darkness) or after the lighting overlay (so they act as emissive light trails). Drawing after depth-sorted entities and before lighting is recommended for standard atmospheric consistency.

---

## 4. Conclusion

The combat and rendering systems in `go-zomboid` are well-structured and ready for the Bezier curve combat dynamics upgrade (R3). The implementation can be cleanly added to `DrawSystem` using Ebitengine's `vector` package without disrupting existing ECS components or headless test harnesses.

---

## 5. Verification Method

To independently verify all findings and test suite integrity:
```bash
# 1. Run all unit and stress tests
CC=gcc go test -v ./...

# 2. Check test coverage
CC=gcc go test -cover ./internal/game ./internal/game/world ./internal/assets

# 3. Test asset generator determinism
go test -v ./cmd/tools/genassets
```
