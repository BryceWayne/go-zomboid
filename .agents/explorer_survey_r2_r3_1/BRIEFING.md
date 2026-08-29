# BRIEFING — 2026-08-29T16:51:30Z

## Mission
Survey codebase for Requirements R2 (Equip/Unequip Items) and R3 (Storage Chest Interaction) and produce comprehensive survey report.

## 🔒 My Identity
- Archetype: explorer
- Roles: survey, analysis, synthesis
- Working directory: /home/bryce/code/go-zomboid/.agents/explorer_survey_r2_r3_1
- Original parent: 8fd0f6a8-cb46-4ae5-8024-c99358e741e1
- Milestone: survey

## 🔒 Key Constraints
- Read-only investigation — do NOT implement in source code
- Produce detailed 5-component handoff report at /home/bryce/code/go-zomboid/.agents/explorer_survey_r2_r3_1/handoff.md
- Send message back to parent when done

## Current Parent
- Conversation ID: 8fd0f6a8-cb46-4ae5-8024-c99358e741e1
- Updated: 2026-08-29T16:51:30Z

## Investigation State
- **Explored paths**: `internal/ecs/components.go`, `internal/game/game.go`, `internal/game/world/map.go`, `internal/assets/assets.go`, all tests in `internal/game`
- **Key findings**:
  1. R2: Current equipping logic in `game.go:444-467` overwrites equipped weapons upon equipping a new item from inventory and provides no unequip mechanism or UI slot. Proposed design introduces dedicated 'Equipped' UI slot at (1070, 265), two-way swap when equipping, hotkey 'U' (and UI drag/click) to return equipped weapon to first empty inventory slot, and full inventory overflow protection.
  2. R3: `TileChest` (17) is solid and procedurally placed in `map.go:808-822`, but has no inventory storage or interaction handling. Proposed design adds `Chests map[Point][]string` to `world.Map`, proximity detection within 192px (1.5 tiles), atomic deep-copy swap on hotkey 'E' with debounce, and on-screen HUD prompt.
- **Unexplored areas**: None for R2/R3 scope.

## Key Decisions Made
- Completed comprehensive survey of R2 (Equip/Unequip) and R3 (Storage Chest Interaction)
- Documented complete architecture, logic chains, UI layout specifications, edge case defenses, and test plans in `handoff.md`

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/explorer_survey_r2_r3_1/handoff.md — Final survey and synthesis report
