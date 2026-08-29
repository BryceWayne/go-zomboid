## 2026-08-29T15:53:56Z

You are Explorer 1 investigating the go-zomboid codebase for a major engine overhaul.
Working directory: /home/bryce/code/go-zomboid/.agents/explorer_survey_1
Original request path: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md

Mission & Scope:
1. Read `/home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md`.
2. Thoroughly investigate the codebase structure, coordinate transformation functions (e.g. IsoToWorld, WorldToIso, TileToScreen, ScreenToTile, etc.), camera system, and DrawSystem.
3. Investigate the asset loading pipeline, tile sizes, sprite sheets, RPG maker assets, and map rendering logic. Determine why/where black gaps occur or how seamless orthogonal top-down rendering must be structured.
4. Document all files, functions, data structures, and mathematical formulas that need to be rewritten to transition from Isometric to 2D Orthogonal (top-down) grid.
5. Write your complete analysis and findings to `/home/bryce/code/go-zomboid/.agents/explorer_survey_1/handoff.md` and send a completion message to parent.
