# Reviewer Handoff Report — Milestones 2 & 3 (Requirements R2 & R3)

## 1. Observation

### 1.1 Requirement R2: Equip/Unequip Items & Dedicated UI Slot
- **Dedicated UI Slot**: Located at HUD coordinates `(1070, 265, 200, 30)` (`internal/game/game.go:1618-1633`).
  - When unarmed, correctly renders `"Equipped: [Empty]"`.
  - When holding a weapon with durability $> 0$, renders `"Equipped: %s (%d hits)"` with weapon name and remaining hits.
  - Highlights background from `color.RGBA{40, 60, 90, 220}` to `color.RGBA{80, 110, 160, 220}` when `draggingSlot == 9`.
  - Tooltip rendering at `internal/game/game.go:1636-1644` safely guards against out-of-bounds index access (`draggingSlot >= 0 && draggingSlot < len(playerInventory)` vs `draggingSlot == 9`).
- **Equip Mechanics**: Implemented in `internal/game/game.go:188-208` (drag & drop) and `internal/game/game.go:515-532` (hotkeys 1-9).
  - When equipping a weapon while already armed, `oldWeapon` is preserved and placed back into the source inventory slot (`player.Inventory[useItemIdx] = oldWeapon`).
  - Correct initial durability is applied per weapon type: `"weapon"` (5 hits), `"axe"` (12 hits), `"shotgun"` (15 hits).
- **Unequip Mechanics & Data Loss Protection**: Implemented in `internal/game/game.go:545-564` (hotkey `'U'`) and `internal/game/game.go:209-244` (drag & drop).
  - Searches for the first empty slot (`player.Inventory[idx] == ""`).
  - If all 9 slots are full (`emptyIdx == -1`), the unequip request is safely rejected without data loss or slice truncation.
  - Drag & drop from slot 9 to an occupied non-weapon slot searches for the first empty slot, preserving existing items.

### 1.2 Requirement R3: Storage Chest Interaction
- **Chest Persistence & Starter Loot**: Implemented in `internal/game/world/map.go:191,205,810-831,1115-1149`.
  - `world.Map.Chests` maps `Point` coordinates to 9-slot inventory slices.
  - `GetChestInventory` and `SetChestInventory` normalize slice lengths to 9 and perform defensive deep copy slice allocations (`make([]string, 9)` and `copy()`).
  - Starter loot properly populated across 4 contextual buildings:
    - Warehouse `(midX + 22, midY + 8)`: `["axe", "ammo", "ammo", "food", "", "", "", "", ""]`
    - Campsite `(width - 10, 9)`: `["food", "water", "weapon", "antidote", "", "", "", "", ""]`
    - House 1 Bedroom `(11, 9)`: `["armor", "water", "food", "", "", "", "", "", ""]`
    - Police Armory `(11, midY + 7)`: `["shotgun", "ammo", "ammo", "armor", "", "", "", "", ""]`
- **Proximity Detection & HUD Prompt**: Implemented in `internal/game/game.go:566-593,1646-1674`.
  - Iterates over a $5 \times 5$ tile neighborhood around the player position and computes Euclidean distance to chest center $(tx \cdot 128 + 64, ty \cdot 128 + 64)$.
  - Activates when $d \le 192.0\text{px}$ (1.5 tiles).
  - Renders interaction prompt `"[E] Swap Inventory with Chest"` at $(520, 650)$ inside background box $(490, 645, 300, 25)$ when alive and in proximity.
- **Atomic Deep-Copy Inventory Swap**: Implemented in `internal/game/game.go:594-618`.
  - Pressing hotkey `'E'` allocates independent slices `make([]string, 9)` and deep-copies elements between `player.Inventory` and `chestInv`.
  - Sets `s.interactCooldown = 20` frames for debounce cooldown protection.
  - Active equipped weapon (`player.WeaponEquipped`, `player.WeaponType`, `player.WeaponDurability`) remains isolated and unaffected during chest swaps.

### 1.3 Independent Verification & Execution Results
- **Full Test Suite (`go test -v -count=1 ./...`)**:
  - `github.com/BryceWayne/go-zomboid/cmd/game`: PASS
  - `github.com/BryceWayne/go-zomboid/internal/assets`: PASS
  - `github.com/BryceWayne/go-zomboid/internal/ecs`: PASS
  - `github.com/BryceWayne/go-zomboid/internal/game`: PASS (including all equip, unequip, chest swap, HUD drawing, and 50,000-cycle stress tests)
  - `github.com/BryceWayne/go-zomboid/internal/game/world`: PASS
- **Go Race Detector (`go test -race ./internal/game ./internal/game/world`)**:
  - `github.com/BryceWayne/go-zomboid/internal/game`: PASS (57.187s, 0 data races)
  - `github.com/BryceWayne/go-zomboid/internal/game/world`: PASS (1.740s, 0 data races)
- **Go Vet (`go vet ./...`)**:
  - Exited with code 0 (0 warnings or errors).
- **Binary Build (`go build -o bin/game ./cmd/game`)**:
  - Exited with code 0 (clean binary compilation to `bin/game`).

---

## 2. Logic Chain

1. **Memory Reference Isolation**:
   - `world.Map.GetChestInventory` and `SetChestInventory` both create fresh slices (`make([]string, 9)`) and perform element copies.
   - In `game.go:607-614`, the swap operation allocates `newPlayerInv` and `newChestInv` before calling `copy()`.
   - Modifying items in the player's inventory (e.g. consuming food or expending ammo) has zero effect on the chest's storage, eliminating memory reference leaks.
2. **Equip/Unequip Invariance**:
   - Equipping swaps an existing weapon back into the source slot, ensuring item count conservation.
   - Unequipping scans for the first empty slot. If none exists (full inventory), the action is rejected, preventing weapon deletion or silent overwrites.
   - Durability tracking and weapon breaking on 0 durability properly reset player combat state (`WeaponEquipped = false, WeaponType = "", WeaponDurability = 0`).
3. **Boundary and Edge Case Robustness**:
   - Proximity boundary test at $191.9\text{px}$ vs $192.1\text{px}$ confirms exact precision.
   - Rapid hammering of `'E'` (100 frames) and `'U'` (1,000 frames) tests confirm that debounce cooldown and full inventory checks prevent unintended behavior.
   - Drag & drop bounds checking cleanly handles slot index 9 without out-of-bounds panics.
4. **Integrity Confirmation**:
   - Source code exhibits genuine business logic and data structures rather than hardcoded returns or facades.
   - All tests pass with real calculations, 100% genuine execution.

---

## 3. Caveats

No caveats. All Milestone 2 (R2) and Milestone 3 (R3) requirements have been rigorously reviewed, stress-tested, and verified.

---

## 4. Conclusion

**Verdict: APPROVE**

The implementations of Milestone 2 (Requirement R2: Equip/Unequip Items & Dedicated UI Slot) and Milestone 3 (Requirement R3: Storage Chest Interaction) satisfy all design specifications, performance requirements, and acceptance criteria. Zero data races, memory leaks, or integrity violations were detected.

---

## 5. Verification Method

To independently reproduce and verify this review:

1. **Run Full Test Suite**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...
   ```
   *Expected result: 0 test failures across all packages.*

2. **Run Concurrency / Race Detector**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -race ./internal/game ./internal/game/world
   ```
   *Expected result: 0 race conditions detected.*

3. **Verify Build**:
   ```bash
   C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game
   ```
   *Expected result: Clean binary output at `bin/game` with exit code 0.*
