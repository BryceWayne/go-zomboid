# BRIEFING — 2026-08-28T17:26:45Z

## Mission
Design the Milestone 2 integration in `internal/game/world/map.go` and `internal/game/game.go`: expanded tile types, collision/FOV logic, ground diamond and Y-depth sorted prop rendering, contextual thematic spawning (player, loot, zombies), and comprehensive unit/integration test cases.

## 🔒 My Identity
- Archetype: explorer
- Roles: investigator, designer
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3
- Original parent: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Milestone: M2 (Environment & Town Generation Updates)

## 🔒 Key Constraints
- Read-only investigation — do NOT implement directly in source files; propose designs, code, and test cases in reports/handoff.
- Align with assets generated in M1 (`internal/assets/images/*.png` and `internal/assets/assets.go`).
- Pure Go implementation, compatible with Ebitengine v2 and Ark ECS.
- Complete 5-component handoff report.

## Current Parent
- Conversation ID: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Updated: 2026-08-28T17:26:45Z

## Investigation State
- **Explored paths**: `PROJECT.md`, `.agents/ORIGINAL_REQUEST.md`, `.agents/teamwork_preview_spec_miner_survey_2/handoff.md`, `internal/assets/assets.go`, `internal/game/world/map.go`, `internal/game/game.go`, `internal/ecs/components.go`, `internal/game/world/map_test.go`, `internal/game/game_test.go`.
- **Key findings**:
  - Expanded `TileType` with 10 constants: `TileGrass`, `TileWall`, `TileDirt`, `TileWoodFloor`, `TileTree`, `TileAsphalt`, `TileConcrete`, `TileTileFloor`, `TileFence`, `TileDebris`.
  - Added methods `IsSolid() bool`, `BlocksVision() bool`, `IsFloor() bool`, and `String() string`.
  - Designed full procedural town layout with asphalt/concrete roads, zoned districts (residential, commercial grocery/pharmacy, police armory, industrial warehouse), multi-room interiors, fenced yards, debris piles, and tree groves.
  - Implemented structured spawn points: Safe player house spawn, thematic room loot spawns (kitchen food/water, police armory guns/ammo/armor, pharmacy supplies), and pre-validated non-trapped zombie spawns with safe perimeter.
  - Formulated ground diamond floor pass and Y-depth sorted obstacle/prop/item pass for `DrawSystem`.
  - Prototyped and verified code via unit test suite passing with 100% success.
- **Unexplored areas**: None for M2.

## Key Decisions Made
- `TileFence` and `TileDebris` are set as solid (`IsSolid() == true`) but allow vision (`BlocksVision() == false`) for tactical gameplay.
- Attached `PlayerSpawn`, `Buildings`, `LootSpawns`, and `ZombieSpawns` metadata directly onto `world.Map`.
- Validated negative coordinates handling in AABB collision detection to prevent edge truncation bugs.

## Artifact Index
- `.agents/teamwork_preview_explorer_m2_3/handoff.md` — 5-component handoff report.
- `.agents/teamwork_preview_explorer_m2_3/proposed_map.go` — Complete verified replacement code for `internal/game/world/map.go`.
- `.agents/teamwork_preview_explorer_m2_3/proposed_map_test.go` — Complete verified test suite for `internal/game/world/map_test.go`.
- `.agents/teamwork_preview_explorer_m2_3/proposed_game_patch.go` — Step-by-step integration code for `internal/game/game.go` and `internal/game/game_test.go`.
