# BRIEFING — 2026-08-29T16:00:00Z

## Mission
Implement Milestone 1: 2D Orthogonal Engine Overhaul (R1) converting the rendering, camera, coordinate math, and asset pipeline to strict 2D top-down orthogonal.

## 🔒 My Identity
- Archetype: worker
- Roles: implementer, qa, specialist
- Working directory: /home/bryce/code/go-zomboid/.agents/worker_m1
- Original parent: d24acf99-20c6-4e30-b7be-668df332bc88
- Milestone: Milestone 1: 2D Orthogonal Engine Overhaul (R1)

## 🔒 Key Constraints
- DO NOT CHEAT. All implementations must be genuine.
- Strict 2D Orthogonal (top-down) grid Cartesian math.
- Screen/world projection:
  screenX = (wx - camX) * zoom + 640.0
  screenY = (wy - camY) * zoom + 360.0
  wx = camX + (screenX - 640.0) / zoom
  wy = camY + (screenY - 360.0) / zoom
- Camera operated on Cartesian world coordinates (wx, wy).
- DrawSystem: remove 2:1 dimetric isometric diamond translations, draw rectangular tiles at (tx*TileSize, ty*TileSize) scaled to fill cell with zero gaps, props anchored top-down with depth worldY+TileSize, entities centered at (pos.X, pos.Y) depth pos.Y, Y depth sorting, bezier combat swoosh converted directly to screen coords.
- Asset pipeline: all image pointers load non-nil textures and are scaled/tiled seamlessly.

## Current Parent
- Conversation ID: d24acf99-20c6-4e30-b7be-668df332bc88
- Updated: 2026-08-29T16:00:00Z

## Task Summary
- **What to build**: 2D Orthogonal Engine Overhaul in `internal/game/game.go` and `internal/assets/assets.go`
- **Success criteria**: Go build passes, tests pass, coordinate math and rendering fully orthogonal top-down
- **Interface contracts**: PROJECT.md
- **Code layout**: internal/game, internal/assets, internal/ecs, internal/game/world

## Key Decisions Made
- Implemented bijective 1:1 orthogonal coordinate transformations (`WorldToIso`, `IsoToWorld`, `ScreenToWorld`, `WorldToScreen`) using `DefaultZoom = 0.5`.
- Updated `Camera` to snap and lerp in world Cartesian coordinates $(wx, wy)$.
- Refactored `DrawSystem.Draw` ground pass to render rectangular tiles top-left aligned $(tx \cdot TileSize, ty \cdot TileSize)$ scaled to tile dimensions, eliminating diamond translations and black voids.
- Updated props/obstacles pass with depth key `worldY + TileSize`, entity/item pass centered at position with depth key `pos.Y` / `iPos.Y`, and vertical top-down Y depth sorting.
- Updated Bezier attack arcs and shotgun radial blasts to project directly from world Cartesian to screen coordinates.
- Updated `internal/assets/assets.go` to load all 49 non-nil image handles cleanly.
- Added comprehensive unit and property tests in `internal/game/orthogonal_engine_test.go`.

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/worker_m1/DISPATCH.md — Dispatch instructions
- /home/bryce/code/go-zomboid/.agents/worker_m1/progress.md — Progress tracker
- /home/bryce/code/go-zomboid/.agents/worker_m1/handoff.md — Final handoff report

## Change Tracker
- **Files modified**: `internal/game/game.go`, `internal/assets/assets.go`, `internal/game/orthogonal_engine_test.go`
- **Build status**: PASS (exit code 0)
- **Pending issues**: None

## Quality Status
- **Build/test result**: `CC=gcc go build ./...` PASS, `CC=gcc go test -v ./internal/ecs ./internal/game/world ./internal/assets` PASS, `CC=gcc go test -v -run "TestOrthogonal" ./internal/game` PASS
- **Lint status**: Clean
- **Tests added/modified**: `internal/game/orthogonal_engine_test.go` (5 new tests covering coordinate transforms, camera Cartesian tracking, seamless tile adjacency, Y-depth sorting, and headless game loop)

## Loaded Skills
- None
