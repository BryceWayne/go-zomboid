## 2026-08-29T15:12:30Z
You are teamwork_preview_spec_miner_survey_3.
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_spec_miner_survey_3.
Please read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md.

Task:
Conduct an in-depth specification mining and survey of the rendering system and test suite:
1. Explore `/home/bryce/code/go-zomboid/internal/game/game.go` and related rendering code. How does `DrawSystem` / `Draw` work? How are tiles, entities, and objects currently rendered?
2. Investigate depth-sorting: is there Y-sorting, layer sorting, or custom depth ordering? How should newly introduced objects (Benches, Chests, Sculptures, etc.) be properly depth-sorted and rendered?
3. Inspect `cmd/game/main.go` and game lifecycle.
4. Survey all tests across the repository: run `CC=gcc go test ./...` or inspect test files to catalog all existing tests, map tests, asset loading tests, and test dependencies.
5. Identify any edge cases, rendering pitfalls, or test requirements needed to meet all acceptance criteria.

Write your comprehensive survey report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_spec_miner_survey_3/survey.md`.
Also write a structured `handoff.md` and update `progress.md` with your status. Send a message when done.
