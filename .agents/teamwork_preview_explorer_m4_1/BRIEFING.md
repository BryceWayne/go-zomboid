# BRIEFING — 2026-08-28T17:39:50Z

## Mission
Investigate and design Milestone 4: Melee Weapon Expansion (Fire Axe Cleave, Reach, Durability, Equipping) in `internal/game/game.go` and related systems.

## 🔒 My Identity
- Archetype: Explorer
- Roles: Investigation, Synthesis
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_1
- Original parent: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Milestone: Milestone 4 - Melee Weapon Expansion

## 🔒 Key Constraints
- Read-only investigation — do NOT implement directly in source code
- Produce concrete pure Go design and handoff report

## Current Parent
- Conversation ID: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Updated: 2026-08-28T17:39:50Z

## Investigation State
- **Explored paths**: `internal/ecs/components.go`, `internal/game/game.go`, `internal/game/world/map.go`, `cmd/tools/genassets/main.go`, `internal/assets/assets.go`, `internal/game/armor_test.go`, `internal/game/game_test.go`, `internal/game/game_stress_test.go`.
- **Key findings**:
  - `ecs.Player` requires `WeaponType string` to distinguish `"axe"`, `"weapon"` (bat), and unarmed `""`.
  - Item usage in `processInputAndCombat()` handles 1-9 inventory keys: equips `"axe"` (durability 12) vs `"weapon"` (durability 5).
  - Melee combat distinguishes Fire Axe (reach 32px, hit radius 32px wide cleave sweep hitting multiple zombies) vs Spiked Bat (reach 24px, radius 24px) vs Unarmed Shove (reach 24px, stun 45f + pushback).
  - Durability degrades by 1 per connecting swing; when 0, weapon unequips (`WeaponEquipped = false`, `WeaponType = ""`).
  - Procedural asset generation for `axe.png` already exists in `cmd/tools/genassets` and `internal/assets`.
- **Unexplored areas**: None for M4 melee scope.

## Key Decisions Made
- Formulated complete pure Go implementation for `ecs.Player` component extension, `processInputAndCombat()`, `DrawSystem.Draw()`, and comprehensive unit tests in `proposed_melee_test.go`.
- Generated `proposed_m4_melee.patch` for implementers.

## Artifact Index
- `.agents/teamwork_preview_explorer_m4_1/DISPATCH.md` — Incoming dispatch log
- `.agents/teamwork_preview_explorer_m4_1/progress.md` — Progress and heartbeat log
- `.agents/teamwork_preview_explorer_m4_1/BRIEFING.md` — Working memory
- `.agents/teamwork_preview_explorer_m4_1/proposed_melee_test.go` — Test suite for melee expansion
- `.agents/teamwork_preview_explorer_m4_1/proposed_m4_melee.patch` — Git diff patch
- `.agents/teamwork_preview_explorer_m4_1/handoff.md` — Final 5-component handoff report
