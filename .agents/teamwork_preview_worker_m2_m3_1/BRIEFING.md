# BRIEFING — 2026-08-29T17:02:50Z

## Mission
Implement Milestone 2 (R2: Equip/Unequip Items & Dedicated UI Slot) and Milestone 3 (R3: Storage Chest Interaction) in go-zomboid.

## 🔒 My Identity
- Archetype: teamwork_preview_worker_m2_m3
- Roles: implementer, qa, specialist
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1
- Original parent: 8fd0f6a8-cb46-4ae5-8024-c99358e741e1
- Milestone: Milestone 2 (R2) & Milestone 3 (R3)

## 🔒 Key Constraints
- Genuine implementation only; no dummy/facade implementations or hardcoded test returns.
- Minimal change principle.
- Co-locate tests in `internal/game/`.
- Ensure durability and weapon stats are preserved during equip/unequip/chest swap.
- Dedicated equipped weapon slot must remain equipped during chest swap.
- Inventory full protection on unequip.
- 10,000-cycle rapid swap stress test for data conservation.
- All tests must pass with `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...`
- Game must build cleanly with `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game`

## Current Parent
- Conversation ID: 8fd0f6a8-cb46-4ae5-8024-c99358e741e1
- Updated: 2026-08-29T17:02:50Z

## Task Summary
- **What to build**: Equip/unequip functionality with dedicated equipped UI slot on HUD, and Storage Chest Interaction with map persistence, proximity detection, HUD prompt, debounce cooldown, atomic inventory swapping, and procedural chest loot.
- **Success criteria**: All R2 and R3 requirements fulfilled, unit & integration tests pass, stress test passes, build passes.
- **Interface contracts**: PROJECT.md & ORIGINAL_REQUEST.md
- **Code layout**: `internal/game/` for Game logic and UI, `internal/game/world/` for Map & Chests.

## Change Tracker
- **Files modified**:
  - `internal/game/world/map.go`: added `Chests map[Point][]string`, starter loot in procedural chests, and `GetChestInventory`/`SetChestInventory` accessors.
  - `internal/game/game.go`: added equipped slot UI rendering at (1070, 265, 200, 30), equip swapping on keys 1-9, unequip on 'U', chest proximity detection (192px), chest interaction prompt on HUD, and atomic deep copy swap on 'E' with debounce cooldown.
  - `internal/game/inventory_equip_test.go`: comprehensive unit and integration tests for R2.
  - `internal/game/chest_interaction_test.go`: comprehensive unit and integration tests for R3 including 10,000-cycle rapid swap stress test.
- **Build status**: PASS (`go build -o bin/game ./cmd/game`)
- **Pending issues**: None

## Quality Status
- **Build/test result**: PASS (100% pass across all packages in `go test -v -count=1 ./...`)
- **Lint status**: PASS (`go vet ./...` clean)
- **Tests added/modified**: 15 new test cases across `inventory_equip_test.go` and `chest_interaction_test.go`.

## Loaded Skills
- None specified.

## Key Decisions Made
- Used deep copying (`copy()`) for chest inventory swaps to isolate memory references and prevent data corruption.
- Isolated equipped weapon state from chest inventory swapping so player is not disarmed during chest interaction.
- Protected unequip against full inventory data loss by rejecting unequip when all 9 slots are filled.
- Dedicated equipped UI slot placed at (1070, 265, 200, 30) with drag & drop support (slot index 9).

## Artifact Index
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1/DISPATCH.md
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1/BRIEFING.md
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1/progress.md
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1/handoff.md
