# Scope: Milestone 3 — Camera System QoL (Zoom, Lerp, Inverted Input Math, Culling)

## Architecture
- Resolution: 1280x720 window and virtual canvas (`Game.Layout` returns 1280, 720; Center: `(640.0, 360.0)`).
- Camera: Persistent `Camera` struct with smoothed `(X, Y)`, target `(TargetX, TargetY)`, `LerpFactor = 0.10`, and `Snap(targetIsoX, targetIsoY)` on spawn/reset.
- Rendering Scale: Global 50% scale matrix (`0.5, 0.5`) in `DrawSystem.Draw` applied to ground tiles, obstacles, items, entities, reticle, and Bezier arcs with screen centering offset `(640, 360)`. HUD/UI remains unscaled 1:1.
- Input Inversion: `ScreenToIso` & `ScreenToWorld` inverts `0.5` scale and camera translation:
  $$mouseIsoX = camX + (mx - 640.0) / 0.5$$
  $$mouseIsoY = camY + (my - 360.0) / 0.5$$
  $$(mouseWorldX, mouseWorldY) = \text{IsoToWorld}(mouseIsoX, mouseIsoY)$$
- Culling & FOV: `DrawSystem` `visionRadius = 2200.0`, `UpdateSystem` `CalculateFOV` radius `22` (or `25`) tiles. Zombie AI sight `visionRadius = 600.0` unchanged.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | Camera Data Structure & Lerping | `Camera` struct with exponential lerping (`LerpFactor = 0.10`), spawn snapping, and synchronized shared reference in `Game`, `UpdateSystem`, and `DrawSystem`. | M3 | ORIGINAL_REQUEST §R2 |
| 2 | Global 50% Zoom Rendering in DrawSystem | Apply `0.5` scale matrix and `(640, 360)` centering to ground tiles, props, items, entities, reticle, and Bezier curves while keeping HUD/UI 1:1. | M3 | ORIGINAL_REQUEST §R1 |
| 3 | Inverted Mouse Click Coordinate Math | Implement `ScreenToIso` and `ScreenToWorld` in `internal/game/game.go` to invert 50% zoom scale and camera offset for accurate movement and combat targeting. | M3 | ORIGINAL_REQUEST §R1 |
| 4 | Vision Radius & FOV Culling Expansion | Increase `DrawSystem` `visionRadius` to `2200.0` px and `UpdateSystem` `CalculateFOV` radius to `22` tiles to prevent edge popping on 1280x720 canvas. | M3 | ORIGINAL_REQUEST §R3 |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M3.1 | Implementation of Camera QoL | Features 1, 2, 3, 4: `internal/game/game.go`, `internal/game/camera_test.go` | none | DONE |
| M3.2 | Review & Empirical Challenge | Independent verification by Reviewers & Challengers | M3.1 | DONE |
| M3.3 | Forensic Integrity Audit | Integrity validation by Forensic Auditor | M3.1, M3.2 | DONE |
| M3.4 | Gate Verification & Completion | Acceptance criteria check & parent reporting | M3.3 | DONE |
