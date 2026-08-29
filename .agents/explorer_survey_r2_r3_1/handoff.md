# Explorer Survey & Implementation Blueprint: Requirements R2 & R3

## 1. Observation

### 1.1 Codebase Structure & Component Inventory
- **ECS Player Definition** (`internal/ecs/components.go:28-48`):
  ```go
  type Player struct {
      Health             float64
      Hunger             float64
      Thirst             float64
      Inventory          []string
      WeaponEquipped     bool
      WeaponType         string
      WeaponDurability   int
      ArmorEquipped      bool
      ArmorType          string
      ArmorDefense       float64
      ArmorDurability    int
      ArmorMaxDurability int
      InfectionResist    float64
      AttackCooldown     int
      Dead               bool
      Infected           bool
      FacingX            float64
      FacingY            float64
  }
  ```
  `Inventory` is maintained as a `[]string` initialized with length 9 (`make([]string, 9)` in `internal/game/game.go:62`).

- **Current Inventory Use & Weapon Equipping** (`internal/game/game.go:417-473`):
  ```go
  useItemIdx := -1
  if ebiten.IsKeyPressed(ebiten.Key1) { useItemIdx = 0 }
  // ... Keys 2-9 ...
  if useItemIdx >= 0 && useItemIdx < len(player.Inventory) && player.AttackCooldown <= 0 {
      player.AttackCooldown = 30
      t := player.Inventory[useItemIdx]
      used := false
      // ...
      } else if t == "weapon" {
          player.WeaponEquipped = true
          player.WeaponType = "weapon"
          player.WeaponDurability = 5
          used = true
      } else if t == "axe" {
          player.WeaponEquipped = true
          player.WeaponType = "axe"
          player.WeaponDurability = 12
          used = true
      } else if t == "shotgun" {
          player.WeaponEquipped = true
          player.WeaponType = "shotgun"
          player.WeaponDurability = 15
          used = true
      }
      if used {
          player.Inventory[useItemIdx] = ""
      }
  }
  ```
  *Defect Identified*: When a weapon is equipped, it overwrites any currently equipped weapon without returning the old weapon to inventory. There is no unequip logic or hotkey, and no dedicated 'Equipped' UI slot displayed on screen.

- **Current Inventory UI Rendering** (`internal/game/game.go:1392-1415`):
  ```go
  ebitenutil.DebugPrintAt(screen, "Inventory (Press 1-9 to use):", 1070, 10)
  for i := 0; i < 9; i++ {
      y := 30 + (i * 25)
      colorBg := color.RGBA{50, 50, 50, 200}
      if i == draggingSlot {
          colorBg = color.RGBA{100, 100, 100, 200}
      }
      vector.DrawFilledRect(screen, 1070, float32(y), 200, 20, colorBg, false)
      text := fmt.Sprintf("%d: [Empty]", i+1)
      if i < len(playerInventory) && playerInventory[i] != "" {
          text = fmt.Sprintf("%d: %s", i+1, playerInventory[i])
      }
      if i != draggingSlot {
          ebitenutil.DebugPrintAt(screen, text, 1075, y+2)
      }
  }
  ```
  The HUD currently only renders slots 1 through 9. There is no visual UI slot for the equipped item.

- **World Storage Chest Representation** (`internal/game/world/map.go:29,41,808-822`):
  `TileChest` is defined as `TileType = 17` and marked as solid (`IsSolid() == true` at line 41).
  In `map.go:808-822`, chests are placed procedurally at 4 designated locations:
  1. Warehouse corner: `Point{midX + 22, midY + 8}`
  2. Campsite: `Point{width - 10, 9}`
  3. Residential bedroom: `Point{11, 9}`
  4. Police armory: `Point{11, midY + 7}`
  *Defect Identified*: `world.Map` currently does NOT store chest inventories. There is no chest interaction logic, proximity detection, or inventory swapping on hotkey 'E'.

- **Asset Availability** (`internal/assets/assets.go:57,121`):
  `assets.ChestImage` is loaded from `images/Small Forest/Bench and chest/Chest.png` (22x21 px) and rendered in `DrawSystem.Draw` (`internal/game/game.go:1076-1077`).

---

## 2. Logic Chain

### 2.1 Requirement R2: Equip/Unequip Items Logic Chain
1. *Observation*: The user request specifies:
   > "Add a dedicated 'Equipped' UI slot on the screen. When a player equips an item, it should move from their main inventory into this dedicated slot. Unequipping the item should return it to an empty slot in the main inventory."
2. *State Modeling*:
   - Player's active weapon is tracked in `ecs.Player` via `WeaponEquipped bool`, `WeaponType string`, and `WeaponDurability int`.
   - When equipping a weapon from slot $i \in [0..8]$:
     - If player already has a weapon equipped (`player.WeaponEquipped == true`):
       - Preserve the old equipped item by putting it back into slot $i$: `player.Inventory[i] = player.WeaponType`.
     - If no weapon is currently equipped:
       - Empty slot $i$: `player.Inventory[i] = ""`.
     - Assign new weapon: `player.WeaponEquipped = true`, `player.WeaponType = itemType`, `player.WeaponDurability = baseDurability`.
3. *Unequipping Mechanism*:
   - Triggered via hotkey 'U' (or clicking/interacting with the Equipped slot).
   - If `player.WeaponEquipped`:
     - Scan `player.Inventory` for the first empty slot $k$ (`player.Inventory[k] == ""`).
     - If an empty slot is found ($k \ge 0$):
       - Transfer weapon to inventory: `player.Inventory[k] = player.WeaponType`.
       - Reset equipped state: `player.WeaponEquipped = false`, `player.WeaponType = ""`, `player.WeaponDurability = 0`.
     - If inventory has no empty slots ($k = -1$):
       - Unequip safely aborts (no item deletion or data loss).
4. *Dedicated 'Equipped' UI Slot Layout*:
   - Positioned in top-right HUD directly below the 9 main inventory slots:
     - Coordinates: $X = 1070$, $Y = 265$, Width = 200, Height = 30.
     - Background: Highlighted border/fill (e.g., `RGBA{40, 60, 90, 220}`).
     - Text display:
       - If equipped: `fmt.Sprintf("Equipped: %s (%d hits)", playerWeaponType, playerDurability)`
       - If unequipped: `"Equipped: [Empty]"`
       - Tooltip/label: `"(Press 'U' to unequip)"`
   - Mouse interaction: Clicking or dragging to/from slot index 9 or the Equipped rectangle allows drag-and-drop equipping and unequipping.

### 2.2 Requirement R3: Storage Chest Interaction Logic Chain
1. *Observation*: The user request specifies:
   > "Implement an interaction mechanic where pressing a specific hotkey (e.g., 'E') while standing near a storage chest instantly swaps the player's entire inventory with the contents of the chest."
   > "Verification: Pressing 'E' near a chest swaps the inventory contents successfully without deleting items or crashing."
2. *Chest Inventory Storage & Architecture*:
   - Add `Chests map[Point][]string` to `world.Map` (`internal/game/world/map.go`).
   - In `NewMap()`:
     - Initialize `m.Chests = make(map[Point][]string)`.
     - When procedural chests are placed, populate thematic starter loot:
       - Warehouse: `[]string{"axe", "ammo", "ammo", "food", "", "", "", "", ""}`
       - Campsite: `[]string{"food", "water", "weapon", "antidote", "", "", "", "", ""}`
       - Bedroom: `[]string{"armor", "water", "food", "", "", "", "", "", ""}`
       - Police Armory: `[]string{"shotgun", "ammo", "ammo", "armor", "", "", "", "", ""}`
   - Provide safe accessor methods:
     ```go
     func (m *Map) GetChestInventory(tx, ty int) []string
     func (m *Map) SetChestInventory(tx, ty int, inv []string)
     ```
     These methods guarantee non-nil slices with length 9, preventing nil-pointer or out-of-bounds panics.
3. *Proximity Detection Algorithm*:
   - Given player world position $(P_x, P_y)$, player tile coordinates are $(p_x, p_y) = (\lfloor P_x / 128 \rfloor, \lfloor P_y / 128 \rfloor)$.
   - Check all 8 neighboring tiles + current tile: $(t_x, t_y) \in [p_x-1..p_x+1] \times [p_y-1..p_y+1]$.
   - For each tile with `m.GetTile(t_x, t_y) == world.TileChest`:
     - Chest center: $(C_x, C_y) = (t_x \cdot 128 + 64, t_y \cdot 128 + 64)$.
     - Euclidean distance $d = \sqrt{(P_x - C_x)^2 + (P_y - C_y)^2}$.
     - If $d \le 192.0$ px (1.5 tile radius), the chest is in interaction range.
4. *Atomic Deep Copy Swapping*:
   - Triggered when `ebiten.IsKeyPressed(ebiten.KeyE)` is detected within proximity and debounce cooldown $\le 0$.
   - Swapping operation:
     ```go
     // Normalize player inventory to 9 slots
     for len(player.Inventory) < 9 {
         player.Inventory = append(player.Inventory, "")
     }
     chestInv := gameMap.GetChestInventory(chestTileX, chestTileY)
     for len(chestInv) < 9 {
         chestInv = append(chestInv, "")
     }

     // Value swap with independent memory allocation
     newPlayerInv := make([]string, 9)
     copy(newPlayerInv, chestInv[:9])

     newChestInv := make([]string, 9)
     copy(newChestInv, player.Inventory[:9])

     player.Inventory = newPlayerInv
     gameMap.SetChestInventory(chestTileX, chestTileY, newChestInv)
     ```
   - *Data Integrity Guarantee*: Because both inventories are cloned via `copy()`, future mutations in player inventory (eating food, firing ammo) do NOT corrupt chest storage.
   - *Equipped Weapon Isolation*: Swapping affects only `player.Inventory` (the 9 main backpack slots). The player's equipped weapon in the dedicated 'Equipped' slot remains in hand, preventing accidental disarming.
5. *HUD Prompt & Feedback*:
   - When in range of a chest, display an interaction prompt on screen:
     `ebitenutil.DebugPrintAt(screen, "[E] Swap Inventory with Chest", 540, 650)`
   - Play audio feedback `assets.PlaySound(assets.ShoveSound)` upon successful swap.

---

## 3. Caveats

1. **Armor vs Weapon Equip Slot**: Requirement R2 explicitly requests a dedicated slot for equipping weapons/items. Armor is currently tracked in `ArmorEquipped`, `ArmorType`, `ArmorDefense`, `ArmorDurability`. If desired, armor can either have its own dedicated armor bar/status (as currently implemented at Y: 75) or share the equipped mechanics. Weapon equipping is the primary focus of R2.
2. **Keyboard Polling & Headless Tests**: In headless unit testing, keyboard presses are verified using headless simulation functions where input keys or direct update system methods are exercised without needing active X11/Wayland display server windows.
3. **Multi-Chest Proximity**: If a player is positioned between two adjacent chests (rare in procedural map), proximity detection deterministically selects the nearest chest by euclidean distance.

---

## 4. Conclusion & Proposed Implementation Blueprint

### 4.1 Changes to `internal/ecs/components.go`
No breaking changes needed. `ecs.Player` already contains all necessary fields (`Inventory`, `WeaponEquipped`, `WeaponType`, `WeaponDurability`). An optional `EquippedSlot string` can be added if desired, or existing fields can be utilized.

### 4.2 Changes to `internal/game/world/map.go`
1. Add `Chests map[Point][]string` to `Map` struct.
2. Initialize `Chests` map in `NewMap()`.
3. Add starter loot to procedural chests in `placeEnvironmentalProps`.
4. Add `GetChestInventory(tx, ty int) []string` and `SetChestInventory(tx, ty int, inv []string)`.

### 4.3 Changes to `internal/game/game.go`
1. **Input & Equip Handling in `processInputAndCombat`**:
   - Update 1-9 key handlers to swap old equipped weapon back into `player.Inventory[useItemIdx]` when equipping a new weapon.
   - Add 'U' key handler to unequip active weapon into the first empty inventory slot.
   - Add 'E' key handler to detect nearby `TileChest` ($d \le 192.0$ px) and perform atomic inventory swap.
   - Implement interaction debounce cooldown (20 frames) to prevent rapid churn on key hold.
2. **UI & Rendering in `DrawSystem.Draw`**:
   - Render dedicated 'Equipped' UI slot at $(1070, 265, 200, 30)$ with item name and durability count.
   - Render proximity prompt `"[E] Swap Inventory with Chest"` when player is near a chest.
   - Update drag-and-drop logic to support dragging to/from the Equipped slot.

---

## 5. Verification Method

### 5.1 Compilation & Existing Test Suite
Run the full test suite to verify zero regressions:
```bash
CC=gcc go test ./...
CC=gcc go build -o bin/game ./cmd/game
```

### 5.2 Unit Test Plan (New Tests to Add)
1. `TestEquip_WeaponMovesToEquippedSlot`: Verifies pressing '1' with an axe in slot 0 clears slot 0 and sets `WeaponEquipped = true, WeaponType = "axe"`.
2. `TestEquip_SwappingEquippedWeapons`: Equipping a shotgun while holding an axe places the axe into the shotgun's former inventory slot and equips the shotgun.
3. `TestUnequip_HotkeyReturnsWeaponToInventory`: Pressing 'U' while holding a weapon transfers the weapon to the first empty slot in `player.Inventory` and sets `WeaponEquipped = false`.
4. `TestUnequip_FullInventorySafety`: When all 9 inventory slots are occupied, pressing 'U' does not delete the equipped weapon or overwrite inventory items.
5. `TestChest_ProximityDetection`: Tests player standing at distances 64px, 128px, 180px (in range) vs 250px, 400px (out of range) from `TileChest`.
6. `TestChest_InventorySwapAtomic`: Verifies pressing 'E' near a chest swaps all 9 slots between player and chest without mutating unrelated fields or dropping items.
7. `TestChest_10000RapidSwapStress`: Executes 10,000 continuous swaps between player and chest, asserting exact conservation of item counts across both inventories.
8. `TestEquippedSlot_HUDDrawing`: Executes headless `DrawSystem.Draw` verifying the 'Equipped' slot and chest interaction prompt render without panic.
