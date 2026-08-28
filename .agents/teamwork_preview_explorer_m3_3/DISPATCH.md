## 2026-08-28T17:31:35Z

You are an Explorer subagent (teamwork_preview_explorer_m3_3).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m3_3
Project root: /home/bryce/code/go-zomboid
Original Request: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan: /home/bryce/code/go-zomboid/PROJECT.md
Survey Handoff: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3/handoff.md

Scope: Milestone 3 - Armor HUD, Visual Feedback & Test Suite
Task:
1. Formulate HUD updates in `internal/game/game.go:DrawSystem.Draw()`:
   - Armor durability bar (Steel Blue, below Thirst bar, e.g. Y=75, W=200, H=15).
   - Armor status text (`Armor: %d/%d (Def: %d%%)`).
   - Repositioning subsequent HUD elements (Weapon, Infected, Day/Night text).
   - Visual player indicator (chest plate highlight or armor tint when equipped).
2. Design comprehensive unit tests in `internal/game/armor_test.go` covering equip, defense reduction, infection deflection, durability decay, and break on 0.
3. Document in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m3_3/handoff.md`.
When done, message your parent.
