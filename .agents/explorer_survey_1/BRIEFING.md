# BRIEFING — 2026-08-29T15:56:00Z

## Mission
Investigate go-zomboid codebase for isometric to 2D orthogonal (top-down) engine overhaul, focusing on coordinate transforms, camera, DrawSystem, asset pipeline/tile sizes, black gaps causes, and complete rewrite mapping.

## 🔒 My Identity
- Archetype: explorer
- Roles: codebase investigation, synthesis, transition mapping
- Working directory: /home/bryce/code/go-zomboid/.agents/explorer_survey_1
- Original parent: d24acf99-20c6-4e30-b7be-668df332bc88
- Milestone: Engine Architecture Survey & Top-Down Migration Plan

## 🔒 Key Constraints
- Read-only investigation — do NOT implement
- Produce exhaustive, verified handoff report with exact formulas, file paths, line numbers

## Current Parent
- Conversation ID: d24acf99-20c6-4e30-b7be-668df332bc88
- Updated: 2026-08-29T15:56:00Z

## Investigation State
- **Explored paths**:
  - `cmd/game/main.go`
  - `internal/assets/assets.go`, `internal/assets/*_test.go`, `internal/assets/audio.go`
  - `internal/ecs/components.go`
  - `internal/game/game.go`, `internal/game/*_test.go`
  - `internal/game/world/map.go`, `internal/game/world/map_test.go`, `internal/game/world/world_empirical_stress_test.go`
  - `context/Lab/Inside_C.png`, `context/Small Forest/`, `context/Zombie Apocalypse Tileset/`
- **Key findings**:
  - Black gaps caused by: 1) dimetric isometric projection skew $(wx-wy, (wx+wy)/2)$, 2) `TileSize=128` vs non-uniform asset dimensions (e.g. 25x24, 31x15), 3) isometric diamond anchor offset `(-128, 0)`, 4) lack of tileset slicing, 5) camera scale mismatch.
  - Coordinate refactoring requires converting `WorldToIso`/`IsoToWorld` to 1:1 orthogonal Cartesian transformations $(wx, wy)$, centering camera directly on $(wx, wy)$, rewriting `DrawSystem` ground pass to render rectangular cells $[x \cdot S, y \cdot S]$ with top-left origin $(0,0)$, depth sorting by `Depth = pos.Y`, and transforming Bezier combat swooshes directly to orthogonal screen space.
- **Unexplored areas**: None.

## Key Decisions Made
- Documented exact mathematical transformations, file locations, line numbers, and test overhaul requirements for handoff.

## Artifact Index
- handoff.md — Comprehensive findings and rewrite mapping
