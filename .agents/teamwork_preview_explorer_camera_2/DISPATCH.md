## 2026-08-28T19:23:10Z

You are an Explorer agent (teamwork_preview_explorer_camera_2).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_2
Project root: /home/bryce/code/go-zomboid

MANDATORY INSTRUCTIONS:
1. Read the original request at /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and project spec at /home/bryce/code/go-zomboid/PROJECT.md.
2. Investigate the codebase with a focus on Requirement 1 (IsoToWorld Mouse-Click Inversion & Input Math) & Requirement 2 (Smooth Camera Centering / Lerping):
   - Inspect `UpdateSystem`, mouse input handling, and coordinate conversions in `internal/game/game.go` and `internal/game/world/map.go` (`WorldToIso`, `IsoToWorld`, screen-to-world conversion).
   - Trace how mouse clicks (screen coordinates `mx, my`) are converted to world coordinates `(wx, wy)` for player movement and attacks.
   - Determine the exact mathematical transformation required when the world rendering is scaled by 50% (0.5) centered on the camera/screen.
   - Inspect the current camera position logic (`camX, camY` or camera struct). How is it updated? What is needed to implement smooth camera lerping (e.g., `camX += (targetCamX - camX) * lerpFactor * dt` or per-frame lerping) so the camera smoothly tracks and centers the player on the 1280x720 screen without snapping or jittering?
3. Write your complete analysis and recommendations to /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_camera_2/handoff.md.
4. Send a message to the orchestrator when finished.
