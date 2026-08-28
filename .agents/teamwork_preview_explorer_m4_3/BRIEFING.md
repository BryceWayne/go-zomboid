# BRIEFING — 2026-08-28T17:40:42Z

## Mission
Design weapon HUD status updates in `internal/game/game.go:DrawSystem.Draw()` and comprehensive unit tests in `internal/game/combat_test.go` for Milestone 4 (Weapon HUD, UI & Combat Test Suite).

## 🔒 My Identity
- Archetype: explorer
- Roles: investigator, analyzer, synthesizer
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3
- Original parent: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Milestone: Milestone 4 - Weapon HUD, UI & Combat Test Suite

## 🔒 Key Constraints
- Read-only investigation — do NOT implement directly in source code, write design & reports in agent folder
- Accurate references to line numbers, data structures, and function signatures

## Current Parent
- Conversation ID: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Updated: 2026-08-28T17:40:21Z

## Investigation State
- **Explored paths**: `internal/game/game.go`, `internal/ecs/components.go`, `internal/game/game_test.go`, `internal/game/armor_test.go`, `internal/game/world/map.go`, `internal/assets/assets.go`
- **Key findings**:
  - `DrawSystem.Draw()` requires `playerWeaponType` extraction and formatted display for `"Weapon: %s (%d hits)"` or `"Weapon: NONE (Fists)"`.
  - Shotgun requires counting `"ammo"` items in `player.Inventory` to display `"Weapon: SHOTGUN (%d hits | Ammo: %d)"`.
  - Comprehensive unit test suite with 12 tests authored in `proposed_combat_test.go` covering axe cleave, shotgun ammo consumption, cone spread reach, dry fire, 400px noise alert, and durability breakdown.
- **Unexplored areas**: None for M4 scope.

## Key Decisions Made
- Authored drop-in test suite in `proposed_combat_test.go`
- Authored patch file in `proposed_m4_hud.patch`
- Authored full 5-component handoff report in `handoff.md`

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/DISPATCH.md — incoming dispatch records
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/progress.md — liveness heartbeat
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/proposed_combat_test.go — 12 unit tests for combat
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/proposed_m4_hud.patch — diff patch for HUD and ECS
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/handoff.md — 5-component handoff report
