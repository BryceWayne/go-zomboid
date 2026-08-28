## 2026-08-28T14:28:45-05:00
You are a Challenger agent (teamwork_preview_challenger_camera_2).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_challenger_camera_2
Project root: /home/bryce/code/go-zomboid

MANDATORY INSTRUCTIONS:
1. Read the original user request at /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and the milestone scope at /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_4/SCOPE.md.
2. Read the worker's handoff report at /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_camera_1/handoff.md.
3. Write empirical integration tests to verify:
   - Viewport boundary culling: test that tiles, entities, and props at distance < 2200px pass culling, and verify no visual clipping at the extreme screen corners (0,0), (1280,0), (0,720), (1280,720).
   - Mouse click tile navigation: simulate mouse clicks at screen coordinates across the 1280x720 window and verify player movement vector heads directly toward the exact clicked world tile.
   - Headless rendering loop execution with `Game.Draw()` across multiple frames with dynamic camera lerping.
4. Run tests with `CC=gcc go test -v ...`. Clean up any temporary test files before writing handoff.
5. Write your findings and verdict to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_challenger_camera_2/handoff.md`.
6. Send a message to the orchestrator when finished.
