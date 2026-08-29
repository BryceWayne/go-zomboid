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
