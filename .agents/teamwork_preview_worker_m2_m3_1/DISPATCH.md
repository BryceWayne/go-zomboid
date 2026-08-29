## 2026-08-29T16:58:37Z
You are Worker 2 implementing Milestone 2: Requirement R2 (Equip/Unequip Items & Dedicated UI Slot) and Milestone 3: Requirement R3 (Storage Chest Interaction).
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1.
Read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md (especially section ## 2026-08-29T16:48:41Z) and /home/bryce/code/go-zomboid/PROJECT.md.
Also read the survey findings and blueprints in /home/bryce/code/go-zomboid/.agents/explorer_survey_r2_r3_1/handoff.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Tasks for Milestones 2 & 3:
1. Implement Requirement R2 (Equip/Unequip Items):
   - Add a dedicated 'Equipped' UI slot on the HUD in `internal/game/game.go` (e.g. at screen coordinates 1070, 265, 200, 30) displaying the active weapon and durability.
   - When a player equips an item (via number keys 1-9 or UI click/drag), move it from the main inventory slot into the dedicated equipped slot. If a weapon is already equipped, swap the old weapon back into that inventory slot.
   - Implement unequip functionality (hotkey 'U' and/or UI interaction) that returns the equipped weapon to the first available empty slot in the main inventory. If the inventory is full (all 9 slots occupied), protect against data loss.
2. Implement Requirement R3 (Storage Chest Interaction):
   - In `internal/game/world/map.go`, implement chest persistence via `Chests map[Point][]string` with `GetChestInventory(tx, ty int) []string` and `SetChestInventory(tx, ty int, inv []string)` ensuring 9-slot inventories. Populate starter loot in procedural chests (Warehouse, Campsite, Bedroom, Police Armory).
   - In `internal/game/game.go`, implement proximity detection for `TileChest` (within 192px / 1.5 tiles).
   - Display a HUD interaction prompt `"[E] Swap Inventory with Chest"` when near a chest.
   - On pressing 'E' near a chest, atomically swap the player's entire 9-slot inventory with the chest's 9-slot inventory using deep copies, with input debounce cooldown (20 frames) and sound feedback. The equipped item in the dedicated equipped slot must remain equipped during the swap.
3. Write comprehensive unit and integration tests for R2 and R3 in `internal/game/inventory_equip_test.go` and `internal/game/chest_interaction_test.go`:
   - Test equipping items, swapping equipped items, unequipping to first empty slot, unequip rejection when full.
   - Test chest proximity detection, atomic inventory swapping, 10,000-cycle rapid swap stress test for data conservation.
4. Verify with:
   `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...`
   and
   `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game`
5. Write your handoff report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_m3_1/handoff.md` and send a message back when complete.
