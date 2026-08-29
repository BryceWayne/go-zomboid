## 2026-08-29T16:00:28Z
You are the Dungeon Master Worker implementing Milestone 2: Dungeon Master Simulation (R2).
Working directory: /home/bryce/code/go-zomboid/.agents/worker_m2
Original request path: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project plan path: /home/bryce/code/go-zomboid/PROJECT.md
Survey report: /home/bryce/code/go-zomboid/.agents/explorer_survey_2/handoff.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. An auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Implementation Details:
1. Implement Dungeon Master Core (`internal/game/dm.go`):
   - Define `DungeonMasterConfig` and `DungeonMaster` struct.
   - Dynamic Zombie Wave Spawning:
     - Spawns waves periodically (e.g. every 1800 ticks / 30s) or when zombie count drops below threshold.
     - Threat scaling formula: base zombies scaled by elapsed ticks + day count + night bonus.
     - Spawns candidate zombies at perimeter distance (700px - 1600px) from player on valid non-solid walkable tiles (`!gameMap.GetTile(tx, ty).IsSolid()`).
     - Scales runner probability: 15% day, 45% night.
   - Dynamic & Randomized Loot Drops:
     - `HandleZombieDeath(wx, wy float64)`: 25% chance of item drop upon kill, weighted drop table across 8 items (ammo 30%, food 25%, water 20%, weapon 10%, antidote 8%, axe 4%, armor 2%, shotgun 1%). Spawns item entity at death position.
     - Periodic ambient supply drops in rooms/plazas when item count is below cap.
   - Day/Night Cycle & Aggression Scaling:
     - `GetAggressionModifiers(timeOfDay float64) (speedMult, noiseMult, visionMult float64)`:
       - Daytime (08:00 - 17:00): 1.0, 1.0, 1.0
       - Night (20:00 - 05:00): speedMult >= 1.25 (up to 1.35 at midnight), noiseMult >= 1.50, visionMult >= 1.25.
     - `GetAmbientLighting(timeOfDay float64) (color.RGBA, float64)`:
       - Dawn (05:00-08:00): Warm rose/gold tint.
       - Day (08:00-17:00): Clear sunlight (alpha 0.0).
       - Dusk (17:00-20:00): Amber twilight.
       - Night (20:00-05:00): Midnight navy tint peaking at alpha ~0.85-0.90.
2. Integrate with Game Loop (`internal/game/game.go`):
   - Add `dm *DungeonMaster` to `Game` struct and initialize in `NewGame()` / `Reset()`.
   - In `UpdateSystem.Update()`, call `dm.Update(timeOfDay, playerPos)`.
   - In `UpdateSystem.processInputAndCombat()`, call `dm.HandleZombieDeath(pos.X, pos.Y)` whenever a zombie is killed.
   - In `UpdateSystem.processZombies()`, apply `dm.GetAggressionModifiers(timeOfDay)` to zombie speed, noise detection radius, and vision radius.
   - In `DrawSystem.Draw()`, use `dm.GetAmbientLighting(timeOfDay)` for ambient day/night overlay.
3. Unit Tests (`internal/game/dm_test.go`):
   - Write tests for wave scaling, perimeter spawn validity (non-solid, >=700px), death drops, ambient restock, day/night lighting, and night aggression scaling.
4. Verification:
   - Run `CC=gcc go test -v -run "TestDungeonMaster|TestOrthogonal" ./internal/game`
   - Run `CC=gcc go build ./...`
   - Document changes and verification results in `/home/bryce/code/go-zomboid/.agents/worker_m2/handoff.md`.
   - Send completion message to parent.
