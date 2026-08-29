# BRIEFING — 2026-08-29T15:29:00Z

## Mission
Implement DrawSystem rendering & depth-sorting for external prop tiles (TileBench, TileChest, TileSculpture, TileBush, TileFlower, TileStone) in internal/game/game.go with ground diamond pass and tests.

## 🔒 My Identity
- Archetype: worker
- Roles: implementer, qa, specialist
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m3
- Original parent: 2341cac8-3fc5-4c81-8832-e3f9a5a870ba
- Milestone: Milestone 3 (DrawSystem Rendering & Depth-Sorting)

## 🔒 Key Constraints
- Multi-pass rendering pipeline: Ground pass (Pass 1) must draw underlying terrain diamond under new props to prevent black/transparent holes.
- Depth-sorted sprite pass (Pass 2): Collect sprites with `Depth = worldX + worldY` and apply geometric anchor transform (`-imgW/2.0`, `128.0-imgH`).
- Stably sorted depth rendering with FOV explored/visible memory tinting.
- No dummy/facade implementations or hardcoded test results.

## Current Parent
- Conversation ID: 2341cac8-3fc5-4c81-8832-e3f9a5a870ba
- Updated: 2026-08-29T15:29:00Z

## Task Summary
- **What to build**: Update `DrawSystem.Draw` in `internal/game/game.go` for prop tiles (`TileBench`, `TileChest`, `TileSculpture`, `TileBush`, `TileFlower`, `TileStone`), update `game_stress_test.go`, and add `draw_depth_test.go`.
- **Success criteria**: All tests in `internal/game/...` and repository `./...` pass cleanly; `./cmd/game` builds without error.

## Change Tracker
- **Files modified**:
  - `internal/game/game.go`: Updated Ground Pass to draw `GrassImage` under prop tiles, and Depth-Sorted Sprite Pass to collect prop sprites with dynamic geometric anchoring (`-imgW/2`, `128.0-imgH`) and `Depth = worldX + worldY`.
  - `internal/game/game_stress_test.go`: Added new prop tiles to `allTiles` in `TestIsometricRenderingAllTileTypesAndPropsStress`.
  - `internal/game/draw_depth_test.go`: Added comprehensive test suite for rendering, depth-sorting, ground pass, and geometric anchoring.
- **Build status**: Pass (`CC=gcc go test ./...` and `CC=gcc go build ./cmd/game`).
- **Pending issues**: None.

## Quality Status
- **Build/test result**: All tests passing across all packages.
- **Lint status**: Clean.
- **Tests added/modified**: `internal/game/draw_depth_test.go`, `internal/game/game_stress_test.go`.
