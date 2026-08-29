# Original User Request

## 2026-08-29T15:53:21Z

# Teamwork Project Prompt — Draft

> Status: Launched
> Goal: Craft prompt → get user approval → delegate to teamwork_preview
> Requested team: dungeon master agent, along with the technical director

Switch the game engine to a 2D Orthogonal perspective to natively support the new RPG Maker assets, and introduce a Dungeon Master system for dynamic gameplay.

Working directory: `/home/bryce/code/go-zomboid`
Integrity mode: demo

## Requirements

### R1. Technical Director Engine Overhaul
Rewrite the engine's core coordinate math (e.g., `IsoToWorld`, `WorldToIso`) and the `DrawSystem` to use a strict 2D Orthogonal (top-down) grid rather than an Isometric projection. Map the new 2D external assets so they tile seamlessly on this orthogonal grid without empty black spaces.

### R2. Dungeon Master Simulation
Implement a Dungeon Master system that injects dynamic gameplay variables. Specifically, implement dynamic zombie wave spawning (scaling difficulty over time), randomized loot drops across the map, and a day/night cycle that darkens the ambient lighting and increases enemy aggression at night.

## Acceptance Criteria

### Verification
- [ ] Running `CC=gcc go test ./...` passes all map generation and logic tests (which should be updated to reflect orthogonal math).
- [ ] Running `CC=gcc go run ./cmd/game` successfully launches the game. The world renders seamlessly in a 2D top-down perspective with no black gaps between tiles.
- [ ] The game visibly cycles between day and night (ambient lighting changes) while playing.
- [ ] Enemies dynamically spawn into the world over time.

## 2026-08-29T16:48:41Z

# Teamwork Project Prompt — Draft

> Status: Launched
> Goal: Craft prompt → get user approval → delegate to teamwork_preview
> Requested team: full team

Enhance the current 2D top-down game engine in `go-zomboid` by upgrading the tile rendering system, implementing equip/unequip mechanics for inventory items, adding improved storage chest interactions, and allowing the player to chop down wooden barriers with weapons. 

**Context:** We recently transitioned the game engine from a 2.5D isometric perspective to a 2D top-down orthogonal grid, and the tile rendering got messed up during the transition. The rendering needs a major upgrade to properly handle 2D top-down autotiling and terrain blending.

Working directory: /home/bryce/code/go-zomboid
Integrity mode: benchmark

## Requirements

### R1. Tile Rendering Upgrade
Fix the messed up tile rendering caused by the 2.5D to 2D transition. Implement 'autotiling' for the new 2D orthogonal grid so that different tile types (such as grass, dirt, and walls) transition seamlessly into one another, eliminating harsh square borders between different terrains.

### R2. Equip/Unequip Items
Add a dedicated 'Equipped' UI slot on the screen. When a player equips an item, it should move from their main inventory into this dedicated slot. Unequipping the item should return it to an empty slot in the main inventory.

### R3. Storage Chest Interaction
Implement an interaction mechanic where pressing a specific hotkey (e.g., 'E') while standing near a storage chest instantly swaps the player's entire inventory with the contents of the chest.

### R4. Environmental Destruction
Allow the player to chop down wooden barriers (fences, walls, etc.) using an axe or suitable weapon. When a barrier is destroyed, it should drop wood or resource items on the ground for the player to collect.

## Acceptance Criteria

### Verification
- [ ] Running `CC=gcc go test ./...` passes all tests.
- [ ] Running `CC=gcc go build -o bin/game ./cmd/game` compiles without errors.
- [ ] The player can successfully equip a weapon, visibly moving it to a new 'Equipped' UI slot.
- [ ] Pressing 'E' near a chest swaps the inventory contents successfully without deleting items or crashing.
- [ ] Attacking a wooden barrier destroys it and spawns a collectible resource item in its place.
- [ ] The terrain rendering correctly handles the 2D orthogonal perspective and utilizes autotiling logic to blend tile edges perfectly.

