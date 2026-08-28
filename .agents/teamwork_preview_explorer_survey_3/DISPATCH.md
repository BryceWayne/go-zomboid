## 2026-08-28T18:47:56Z
You are survey_explorer_3.
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3
Original Request File: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project root: /home/bryce/code/go-zomboid

Mission:
Investigate combat dynamics, attack arcs, DrawSystem / rendering loop, and test infrastructure.
Specifically:
1. Read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md.
2. Examine player attack triggering, weapon types (axe, baseball bat, knife, etc.), attack animations / timers, facing direction calculation, and mouse input handling.
3. Examine `DrawSystem` / rendering logic (e.g., in `internal/game/render.go` or `internal/game/`): How are weapons, entities, and effects currently drawn? How can dynamic weapon swing trails/arcs (swoosh effect) using Bezier Curves (quadratic/cubic bezier with vector points, alpha fading, stroke width, color gradient/fill) be implemented in Ebitengine?
4. Inspect existing test suites across the project (`go test ./...`), test coverage, and test structure.
5. Write a comprehensive survey report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3/survey_report.md` and `handoff.md`.
6. Send a message to your parent when complete.
