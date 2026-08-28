## 2026-08-28T19:22:41Z
Implement Quality of Life (QoL) improvements for the camera system. Specifically, apply a global zoom-out scale to increase the visible world area around the player, and implement smooth camera tracking (lerping) to keep the player centered.
Integrity mode: development
Requested team: full team

Requirements:
- R1. Global Camera Zoom: Modify the DrawSystem in internal/game/game.go to apply a global 50% scale matrix to the entire game world rendering (excluding the UI/HUD). This will make the high-resolution assets appear smaller, vastly increasing the player's field of view. Ensure the IsoToWorld mouse-click math is updated to account for this new zoom scale so that clicking to move/attack still works accurately.
- R2. Smooth Camera Centering: Implement smooth camera tracking (lerping) so the camera smoothly drifts towards the player's position rather than instantly snapping. The player should remain dynamically centered on the 1280x720 screen.
- R3. Vision Radius Culling Expansion: Increase the visionRadius and FOV culling distance to account for the new 50% zoom scale, ensuring tiles and entities do not pop out of existence at the edges of the newly expanded view.

Acceptance Criteria:
- Running `CC=gcc go test ./...` passes all existing map and input tests.
- Running `CC=gcc go run ./cmd/game` successfully launches the game. The world appears zoomed out, and the player is smoothly tracked in the center of the screen.
- Left-clicking on a tile correctly moves the player to that tile (verifying the inverted zoom math in `IsoToWorld`).
