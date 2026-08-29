# Progress Log

Last visited: 2026-08-29T17:02:48Z

## Completed Tasks
1. Milestone 2 / Requirement R2 (Equip/Unequip Items & Dedicated UI Slot):
   - Implemented weapon equip swapping logic across 1-9 number keys in `internal/game/game.go`.
   - Implemented unequip logic via hotkey 'U' returning active weapon to first available empty slot in the main inventory.
   - Implemented full inventory data loss protection when unequipping.
   - Implemented drag & drop support between inventory slots 0..8 and dedicated equipped slot 9.
   - Added dedicated 'Equipped' UI slot at (1070, 265, 200, 30) displaying weapon type, hit count durability, or empty status.
2. Milestone 3 / Requirement R3 (Storage Chest Interaction):
   - Added `Chests map[Point][]string` to `world.Map` with `GetChestInventory(tx, ty int)` and `SetChestInventory(tx, ty int, inv []string)` in `internal/game/world/map.go`.
   - Populated thematic starter loot in procedural chests (Warehouse, Campsite, Bedroom, Police Armory).
   - Implemented proximity detection for `TileChest` within 192px (1.5 tiles).
   - Added HUD prompt `"[E] Swap Inventory with Chest"` when near a chest.
   - Implemented atomic inventory swapping on pressing 'E' with deep copies and 20-frame debounce cooldown. Equipped weapon remains isolated and equipped during swaps.
3. Testing & Verification:
   - Added comprehensive unit and integration tests in `internal/game/inventory_equip_test.go` and `internal/game/chest_interaction_test.go`.
   - Verified 10,000-cycle rapid swap stress test for data conservation.
   - Verified `go test -v -count=1 ./...` passes (100% pass across all packages).
   - Verified `go build -o bin/game ./cmd/game` builds cleanly.
   - Verified `go vet ./...` clean.
