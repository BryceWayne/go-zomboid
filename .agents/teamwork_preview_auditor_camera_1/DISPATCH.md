## 2026-08-28T14:28:45-05:00

You are a Forensic Auditor agent (teamwork_preview_auditor_camera_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_auditor_camera_1
Project root: /home/bryce/code/go-zomboid

MANDATORY INSTRUCTIONS:
1. Read the original user request at /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md and milestone scope at /home/bryce/code/go-zomboid/.agents/teamwork_preview_orchestrator_4/SCOPE.md.
2. Read the worker's handoff report at /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_camera_1/handoff.md.
3. Perform a rigorous forensic integrity audit on all changes made:
   - Check `internal/game/game.go`, `internal/game/camera_test.go`, and git diffs for:
     - Hardcoded test inputs/outputs or cheating logic
     - Dummy or facade implementations
     - Fabricated test assertions or bypasses
     - Circumvention of 50% scale matrix, smooth lerping, inverted mouse coordinate math, or FOV/culling expansion
     - Unintended side-effects or regressions
4. Run static analysis and verification tests: `CC=gcc go test -v ./...` and verify genuine code execution.
5. Deliver a clear, binary verdict: **CLEAN** or **INTEGRITY VIOLATION / CHEATING DETECTED** with detailed evidence.
6. Write your audit report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_auditor_camera_1/handoff.md`.
7. Send a message to the orchestrator with your verdict.
