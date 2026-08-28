## 2026-08-28T19:23:10Z

You are an Explorer agent (teamwork_preview_explorer_camera_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1
Project root: /home/bryce/code/go-zomboid

MANDATORY INSTRUCTIONS:
1. Read the original request at /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and project spec at /home/bryce/code/go-zomboid/PROJECT.md.
2. Investigate the codebase with a focus on Requirement 1 (Global Camera Zoom & DrawSystem rendering scale):
   - Examine `DrawSystem` and rendering pipeline in `internal/game/game.go` (and any related rendering files).
   - Determine how world sprites (floors, walls/props, entities, items, Bezier curves) are drawn vs HUD/UI elements.
   - Determine where and how to apply the global 50% scale matrix (e.g. `GeoM.Scale(0.5, 0.5)` or equivalent viewport scaling/matrix transform) to game world elements while leaving HUD/UI at 1:1 scale.
   - Examine the screen dimensions (1280x720) and how camera offset / viewport centering interacts with a 50% world scale.
3. Write your complete analysis and recommendations to /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_1/handoff.md.
4. Send a message to the orchestrator when finished.
