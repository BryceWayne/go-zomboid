## 2026-08-28T18:47:56Z
You are survey_explorer_2.
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2
Original Request File: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project root: /home/bryce/code/go-zomboid

Mission:
Investigate the engine isometric math, world coordinate transforms, movement, camera, and map systems.
Specifically:
1. Read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md.
2. Examine `internal/game/world` and `internal/game/game.go`, `internal/game/render.go`, `internal/game/player.go`, `internal/game/zombie.go`, etc.
3. Identify all constants and math: `TileSize`, `WorldToIso`, `IsoToWorld`, speed coefficients, tile width/height, half widths/heights, camera offsets, screen centering, chunk sizing, collision bounding boxes.
4. How do current math functions convert between world coordinates, isometric screen coordinates, and tile grid indexes?
5. How should `TileSize` and coordinate math be upgraded for 256x128 (4x texture resolution) so that map generation, tile positioning, sorting, entity positions, collision detection, movement speed, and camera tracking remain seamless and non-broken?
6. Write a comprehensive survey report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_2/survey_report.md` and `handoff.md`.
7. Send a message to your parent when complete.
