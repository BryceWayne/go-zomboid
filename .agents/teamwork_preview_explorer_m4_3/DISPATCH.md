## 2026-08-28T17:38:28Z
You are an Explorer subagent (teamwork_preview_explorer_m4_3).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3
Project root: /home/bryce/code/go-zomboid
Original Request: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan: /home/bryce/code/go-zomboid/PROJECT.md
Survey Reference: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_survey_3/handoff.md

Scope: Milestone 4 - Weapon HUD, UI & Combat Test Suite
Task:
1. Design weapon HUD status updates in `internal/game/game.go:DrawSystem.Draw()`:
   - Formatted weapon text: `fmt.Sprintf("Weapon: %s (%d hits)", strings.ToUpper(player.WeaponType), player.WeaponDurability)` or `"Weapon: NONE (Fists)"`.
   - Ammo count indicator when shotgun equipped (counting `"ammo"` items in inventory).
2. Design comprehensive unit tests in `internal/game/combat_test.go` covering:
   - Axe cleave multi-target kill and durability loss.
   - Shotgun ammo check, consumption, cone reach hit, and out-of-ammo dry fire.
   - Noise pulse triggering zombie chase within 400px.
   - Weapon durability breakdown on 0 hits.
3. Document in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_3/handoff.md`.
When done, message your parent.

## 2026-08-28T17:40:21Z
Sender: efb9db38-c509-4c3c-ad0a-53ad2f86b201
**Context**: Milestone 4 Weapon HUD & Combat Test Suite
**Content**: Checking on status of your handoff report for Weapon HUD status display and combat test suite.
**Action**: Please complete and write handoff.md.
