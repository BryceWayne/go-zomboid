## 2026-08-28T17:38:28Z
You are an Explorer subagent (teamwork_preview_explorer_m4_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_1
Project root: /home/bryce/code/go-zomboid
Original Request: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan: /home/bryce/code/go-zomboid/PROJECT.md
Survey Reference: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3/handoff.md

Scope: Milestone 4 - Melee Weapon Expansion (Fire Axe Cleave, Reach, Durability)
Task:
1. Design the melee combat system in `internal/game/game.go:processInputAndCombat()`:
   - Distinguish `"axe"` (Fire Axe) vs `"weapon"` (Spiked Bat) vs unarmed fists.
   - Axe stats: Durability 12, reach 32.0px (vs bat 24.0px), wide swing sweep angle hitting multiple zombies.
   - Bat stats: Durability 5, reach 24.0px.
   - Equipping from inventory slot (1-9) setting `WeaponEquipped = true`, `WeaponType = "axe"`, `WeaponDurability = 12`.
2. Formulate pure Go implementation code and document in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_1/handoff.md`.
When done, message your parent.
