## 2026-08-28T19:28:44Z

You are a Challenger agent (teamwork_preview_challenger_camera_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_challenger_camera_1
Project root: /home/bryce/code/go-zomboid

MANDATORY INSTRUCTIONS:
1. Read the original user request at /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and the milestone scope at /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_4/SCOPE.md.
2. Read the worker's handoff report at /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_camera_1/handoff.md.
3. Write empirical stress tests, fuzzing harnesses, or mathematical invariant checks in your working directory or in temporary test files (e.g., in `internal/game/`) to empirically stress-test:
   - `ScreenToWorld` and `ScreenToIso` over millions of randomized floating-point coordinates, canvas boundaries, negative coordinates, and extreme camera offsets.
   - Camera lerp stability under rapid direction changes, extreme target jumps, zero-distance stability, and sub-pixel snapping.
   - Verify zero NaN/Inf or precision drifts.
4. Run the stress tests with `CC=gcc go test -v ...`. Clean up any temporary test files before writing handoff.
5. Write your findings and verdict to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_challenger_camera_1/handoff.md`.
6. Send a message to the orchestrator when finished.
