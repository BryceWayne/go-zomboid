# Project: go-zomboid 2D Orthogonal Engine Enhancement

## Architecture
The `go-zomboid` project is an apocalyptic zombie survival 2D game built on Go, Ebitengine, and Ark ECS (`github.com/mlange-42/ark/ecs`).
- `cmd/game`: Binary entry point, window configuration (1280x720), asset loading, and Ebitengine game loop runner.
- `internal/ecs`: Pure data component definitions (`Player`, `Zombie`, `Item`, `Position`, `Velocity`, `Sprite`, `Collider`).
- `internal/assets`: Embedded assets loader (`images/*`, `audio.go`), sprite pointers, texture slicing, and autotiling transition overlays.
- `internal/game/world`: Procedural 100x100 tile map generator, AABB collision detection, 2D FOV raycasting, chest storage persistence, destructible barrier health, and autotiling bitmask calculators.
- `internal/game`:
  - `game.go`: `Game` state, `Camera` tracking, `UpdateSystem` (combat, movement, inventory, chest interactions, barrier destruction), and `DrawSystem` (autotiling terrain rendering, lighting, equipped HUD slot).
  - `dm.go`: `DungeonMaster` simulation engine managing dynamic wave spawning, difficulty curves, loot replenishment, and day/night aggression modifiers.
  - Coordinate system: Strict 2D Orthogonal (top-down) Cartesian grid mapping $(wx, wy) \leftrightarrow (sx, sy)$.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | 2D Autotiling & Terrain Blending | Multi-quadrant blob / bitmask autotiling eliminating harsh square borders between grass, dirt, concrete, asphalt, and floors. | M1 | ORIGINAL_REQUEST §R1 |
| 2 | Connected Wall & Fence Autotiling | 4-bit cardinal autotiling for walls, facades, corners, T-junctions, and fence connections. | M1 | ORIGINAL_REQUEST §R1 |
| 3 | Dedicated 'Equipped' UI Slot | Dedicated UI slot on HUD displaying active equipped weapon and durability, separate from 9 inventory slots. | M2 | ORIGINAL_REQUEST §R2 |
| 4 | Two-Way Equip/Unequip Mechanics | Equipping from inventory slot moves weapon to equipped slot (returning previous weapon); unequipping ('U'/click) returns to empty inventory slot. | M2 | ORIGINAL_REQUEST §R2 |
| 5 | Storage Chest Persistence & Proximity | Chest inventory storage `Chests map[Point][]string` in `world.Map` with proximity detection ($\le 192\text{px}$). | M3 | ORIGINAL_REQUEST §R3 |
| 6 | Hotkey 'E' Atomic Inventory Swap | Pressing 'E' near a chest swaps player's 9-slot inventory with chest contents atomically without data loss or crashes. | M3 | ORIGINAL_REQUEST §R3 |
| 7 | Destructible Barrier Durability Model | Tile durability tracking (`TileDurability map[Point]int`) for fences, walls, trees, stumps; perimeter walls indestructible. | M4 | ORIGINAL_REQUEST §R4 |
| 8 | Weapon/Axe Chopping & Collision | Melee and axe swings detect wooden barriers, reduce durability, and clear map solidity/vision blocking on 0 HP. | M4 | ORIGINAL_REQUEST §R4 |
| 9 | Wood Resource Item Spawning & Pickup | Destroyed wooden barriers spawn collectible `ecs.Item{Type: "wood"}` entities that player can collect into inventory. | M4 | ORIGINAL_REQUEST §R4 |
| 10 | Comprehensive Test Suite & E2E Verification | Unit, integration, and stress tests for all features; 100% pass on `CC=gcc go test ./...` and `CC=gcc go build -o bin/game ./cmd/game`. | M5 | ORIGINAL_REQUEST §Acceptance |
| 11 | Adversarial Hardening & Forensic Integrity Audit | Multi-reviewer, multi-challenger stress testing and forensic audit for genuine implementations. | M5 | ORIGINAL_REQUEST §Acceptance |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Tile Rendering Upgrade & Autotiling (R1) | Implement 2D orthogonal autotiling, bitmask calculations, terrain blending between grass, dirt, concrete, asphalt, and connected wall/fence pieces to eliminate harsh square borders. | none | DONE |
| M2 | Equip/Unequip Items & Dedicated UI Slot (R2) | Add dedicated 'Equipped' UI slot, implement two-way equip/swap from inventory, unequip hotkey/interaction returning weapon to first empty inventory slot. | none | DONE |
| M3 | Storage Chest Interaction & 'E' Swap (R3) | Chest persistence in `world.Map`, proximity detection, atomic 9-slot inventory swap on 'E' with debounce and HUD prompt. | none | DONE |
| M4 | Environmental Destruction & Resource Drops (R4) | Barrier durability in `world.Map`, weapon/axe attack chopping collision, solidity/vision clearing, `wood` resource drop spawning & pickup. | none | DONE |
| M5 | Comprehensive E2E Verification, Adversarial Hardening & Forensic Audit | Full test suite verification, build check, adversarial stress tests, and forensic integrity audit. | M1, M2, M3, M4 | DONE |

## Interface Contracts
### Autotiling & Map Rendering
- `GetCardinalBitmask4(m *Map, x, y int, matchType TileType) uint8`: Computes cardinal neighbor mask $\in [0..15]$.
- `GetQuadrantSubtile(m *Map, x, y int, quad Quadrant, primaryType TileType) SubtileState`: Computes sub-tile corner/edge state for 4-quadrant terrain blending.
- `DrawSystem.Draw`: Renders base ground layer, transition overlays, and Y-sorted connected walls/fences.

### Inventory & Equipped Slot
- `Player.WeaponEquipped bool`, `Player.WeaponType string`, `Player.WeaponDurability int`: Reflects current item in dedicated equipped slot.
- `Player.Inventory []string`: Fixed length 9 slice for main backpack slots.
- HUD layout: 9 backpack slots at $(1070, 30+i\cdot 25)$ and Dedicated 'Equipped' slot at $(1070, 265, 200, 30)$.

### Storage Chest Interaction
- `Map.GetChestInventory(tx, ty int) []string`: Returns 9-slot inventory for chest at $(tx, ty)$.
- `Map.SetChestInventory(tx, ty int, inv []string)`: Stores 9-slot inventory for chest at $(tx, ty)$.
- Proximity check: Euclidean distance between player center and chest center $\le 192.0$ px.
- Interaction: Hotkey 'E' swaps player inventory with chest inventory atomically with 20-frame debounce cooldown.

### Environmental Destruction
- `Map.IsDestructible(x, y int) bool`: Returns true for fences, interior walls, trees, stumps; false for perimeter boundaries.
- `Map.DamageTile(x, y int, amount int) (destroyed bool, dropType string)`: Decrements durability, replaces with floor on 0 HP, returns drop resource `"wood"`.
- Attack routine: Weapon and axe swings query destructible tiles in swing radius, apply damage, decrement weapon durability, and spawn `ecs.Item{Type: "wood"}` on destruction.

## Code Layout
- `cmd/game/main.go`: Game entry point.
- `internal/ecs/components.go`: ECS data components (`Player`, `Item`, `Position`, etc.).
- `internal/assets/assets.go`: Asset loading, sprite pointers, autotile image slices.
- `internal/assets/audio.go`: Audio and sound generator.
- `internal/game/world/map.go`: Map model, tile types, procedural generation, chest storage, barrier durability, AABB collision, FOV.
- `internal/game/game.go`: Game loop, Camera, UpdateSystem (combat, chopping, chests, equip/unequip), DrawSystem (autotiling terrain, equipped HUD slot, chest prompts).
- `internal/game/dm.go`: Dungeon Master simulation logic.
