# BRIEFING — 2026-08-29T17:08:20Z

## Mission
Implement Milestone 4: Requirement R4 (Environmental Destruction & Resource Drops) for go-zomboid.

## 🔒 My Identity
- Archetype: Worker
- Roles: implementer, qa, specialist
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m4_1
- Original parent: 8fd0f6a8-cb46-4ae5-8024-c99358e741e1
- Milestone: Milestone 4 (Requirement R4)

## 🔒 Key Constraints
- Tile durability tracking in `world.Map`.
- `IsDestructible(x, y int) bool` (fences, interior walls, trees, stumps, benches destructible; perimeter boundary walls indestructible).
- `GetTileMaxDurability(t TileType) int`, `GetTileDurability(x, y int) int`.
- `DamageTile(x, y int, amount int) (destroyed bool, dropType string)` clearing collision & vision blocking, replacing with walkable ground, returning "wood".
- In `internal/game/game.go`, integrate barrier chopping into melee attacks:
  - Axe: dmg 2, Club/Weapon: dmg 1, Shotgun: dmg 2, Unarmed: dmg 0 (cannot damage barriers).
  - Melee attack reach checks destructible tiles, damages them, decrements weapon durability, plays hit sound.
  - Spawns `ecs.Item{Type: "wood"}` entity at tile center on destruction.
  - Player stepping within 64px collects "wood" item.
  - "wood" rendering and HUD inventory display supported.
- Unit and integration tests in `internal/game/world/destruction_test.go` and `internal/game/destruction_combat_test.go`.
- Real logic only, no cheating or facades.

## Current Parent
- Conversation ID: 8fd0f6a8-cb46-4ae5-8024-c99358e741e1
- Updated: 2026-08-29T17:08:20Z

## Task Summary
- **What to build**: Environmental destruction mechanics, tile durability, resource drops (wood), melee chopping interactions, item pickup, and comprehensive test suite.
- **Success criteria**: All requirements implemented, all tests pass, binary builds cleanly.
- **Interface contracts**: PROJECT.md, ORIGINAL_REQUEST.md, survey blueprint.
- **Code layout**: internal/game, internal/game/world.

## Key Decisions Made
- `world.Map.TileDurability` implemented as `map[Point]int` tracking degraded tile health.
- `IsDestructible(x, y)` protects map perimeter boundaries while allowing interior fences, walls, trees, stumps, and benches to be chopped.
- `DamageTile` resets tile to `TileWoodFloor` for walls, `TileGrass` for fences/trees/stumps/benches, removing solidity and vision blocking.
- Combat routine in `game.go` calculates barrier chopping for Axe (dmg 2), Club/Weapon (dmg 1), Shotgun (dmg 2), and Unarmed (dmg 0).
- Destructed tiles instantiate `ecs.Item{Type: "wood"}` via ECS map, collected into player inventory when stepping within 64px.

## Artifact Index
- DISPATCH.md — assignment details
- BRIEFING.md — situational awareness
- progress.md — liveness and step progress
- handoff.md — comprehensive handoff report

## Change Tracker
- **Files modified**:
  - `internal/game/world/map.go`: Durability tracking, `IsDestructible`, `GetTileMaxDurability`, `GetTileDurability`, `DamageTile`.
  - `internal/game/game.go`: Barrier chopping integration in attack routines, `assets.WoodImage` item rendering in DrawSystem.
  - `internal/game/world/destruction_test.go`: 9 comprehensive unit tests for map durability, destruction, and perimeter invariants.
  - `internal/game/destruction_combat_test.go`: 7 comprehensive combat & integration tests for weapon chopping, wood drops, and breach traversal.
- **Build status**: All unit & integration tests pass (exit code 0); `bin/game` compiles cleanly.
- **Pending issues**: None

## Quality Status
- **Build/test result**: Pass (100% tests passing across repository)
- **Lint status**: Clean (no unused variables or compile warnings)
- **Tests added/modified**: 16 new test functions across `world/destruction_test.go` and `destruction_combat_test.go`

## Loaded Skills
- None
