## 2026-08-29T15:53:37Z
You are the Project Orchestrator for go-zomboid.
Your working directory is `/home/bryce/code/go-zomboid/.agents/orchestrator_1`.
Please read the authoritative user request at `/home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md`.

Mission:
Switch the game engine to a 2D Orthogonal perspective to natively support the new RPG Maker assets, and introduce a Dungeon Master system for dynamic gameplay.

Requirements & Acceptance Criteria:
1. R1. Technical Director Engine Overhaul:
   - Rewrite the engine's core coordinate math (e.g., IsoToWorld, WorldToIso) and the DrawSystem to use a strict 2D Orthogonal (top-down) grid rather than an Isometric projection.
   - Map the new 2D external assets so they tile seamlessly on this orthogonal grid without empty black spaces.
2. R2. Dungeon Master Simulation:
   - Implement a Dungeon Master system that injects dynamic gameplay variables.
   - Dynamic zombie wave spawning (scaling difficulty over time).
   - Randomized loot drops across the map.
   - Day/night cycle that darkens the ambient lighting and increases enemy aggression at night.
3. Verification:
   - `CC=gcc go test ./...` passes all map generation and logic tests (updated to reflect orthogonal math).
   - `CC=gcc go run ./cmd/game` successfully launches the game with seamless 2D top-down rendering.
   - Game visibly cycles between day and night (ambient lighting changes).
   - Enemies dynamically spawn into the world over time.

Please coordinate the specialists, maintain your BRIEFING.md, plan.md, and progress.md in `/home/bryce/code/go-zomboid/.agents/orchestrator_1/`, and notify me when complete.
