## 2026-08-28T17:38:28Z
Scope: Milestone 4 - Ranged Weapon System (Shotgun, Ammo Consumption & Noise Alert)
Task:
1. Design the ranged combat system in `internal/game/game.go:processInputAndCombat()`:
   - Shotgun equipping: `WeaponType = "shotgun"`, `WeaponDurability = 15`.
   - Ammo requirement: checks for `"ammo"` in `player.Inventory`; if found, consumes 1 `"ammo"` item on fire.
   - Ranged spread cone: 3 pellet vectors or cone angle $\pm 22.5^\circ$ up to distance 160.0px in facing direction `(FacingX, FacingY)`.
   - Zombie kills & knockback in cone.
   - Acoustic Noise Pulse: Alerts all wandering zombies within radius 400.0px (`z.Chasing = true`, `z.WanderTimer = 0`), drawing the horde towards player position.
   - Dry fire / out-of-ammo behavior.
2. Formulate pure Go implementation code and document in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m4_2/handoff.md`.
When done, message your parent.
