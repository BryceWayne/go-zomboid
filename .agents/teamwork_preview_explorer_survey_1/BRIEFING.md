# BRIEFING — 2026-08-28T17:14:00Z

## Mission
Investigate procedural sprite generation in cmd/tools/genassets and asset handling in internal/assets/images and internal/assets to understand sprite inventory, generation mechanisms, loading/rendering pipeline, improvement opportunities, and constraints.

## 🔒 My Identity
- Archetype: explorer
- Roles: explorer, analyst
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_1
- Original parent: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Milestone: asset procedural generation survey

## 🔒 Key Constraints
- Read-only investigation — do NOT implement
- Purely procedural generation / no external downloads
- Only write in agent's own directory (.agents/teamwork_preview_explorer_survey_1)

## Current Parent
- Conversation ID: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Updated: 2026-08-28T17:14:00Z

## Investigation State
- **Explored paths**:
  - `cmd/tools/genassets/main.go`
  - `internal/assets/assets.go`
  - `internal/assets/audio.go`
  - `internal/ecs/components.go`
  - `internal/game/game.go`
  - `internal/game/game_test.go`
  - `internal/game/world/map.go`
  - `internal/game/world/map_test.go`
  - `cmd/game/main.go`
  - `ORIGINAL_REQUEST.md`
- **Key findings**:
  - Currently 11 sprites generated: player.png (16x32), zombie.png (16x32), runner.png (16x32), weapon.png (16x16), food.png (16x16), water.png (16x16), grass.png (64x32), dirt.png (64x32), wood.png (64x32), wall.png (64x64), tree.png (64x64).
  - All existing sprites are generated using basic shapes / solid fills / simple noise.
  - Assets are embedded via `//go:embed images/*` in `internal/assets/assets.go` and decoded at startup into `*ebiten.Image`.
  - Game engine renders floor tiles at (isoX - 32, isoY), walls/trees at (isoX - 32, isoY - 32), items at (isoX - 8, isoY - 8), and entities at (isoX - 8, isoY - 32) with Y-depth sorting.
  - Upgrading sprites requires purely procedural pixel-art algorithms in Go standard library (color palettes, anatomical layers, shading, noise, dithering, geometric primitives).
  - Adding armor and new weapons requires expanding genassets to produce new sprites (`armor.png`, `axe.png`, etc.), declaring them in `internal/assets/assets.go`, and integrating with ECS components and rendering.
- **Unexplored areas**: None remaining for this survey.

## Key Decisions Made
- Fully documented all 11 sprites, rendering pipeline, coordinate transformations, and algorithmic upgrade designs.

## Artifact Index
- DISPATCH.md — record of initial dispatch
- BRIEFING.md — working memory
- progress.md — liveness heartbeat
- handoff.md — final analysis handoff report
