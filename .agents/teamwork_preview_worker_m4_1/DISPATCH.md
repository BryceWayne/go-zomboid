## 2026-08-29T17:05:37Z

You are Worker 3 implementing Milestone 4: Requirement R4 (Environmental Destruction & Resource Drops).
Your working directory is /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m4_1.
Read /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md (especially section ## 2026-08-29T16:48:41Z) and /home/bryce/code/go-zomboid/PROJECT.md.
Also read the survey findings and blueprints in /home/bryce/code/go-zomboid/.agents/explorer_survey_r4_1/handoff.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Tasks for Milestone 4:
1. Implement Requirement R4 (Environmental Destruction):
   - In `internal/game/world/map.go`:
     - Implement tile durability tracking (e.g. `TileDurability map[Point]int` in `world.Map`).
     - Implement `IsDestructible(x, y int) bool` (returns true for fences, interior walls, trees, stumps, benches; returns false for perimeter boundary walls).
     - Implement `GetTileMaxDurability(t TileType) int`, `GetTileDurability(x, y int) int`.
     - Implement `DamageTile(x, y int, amount int) (destroyed bool, dropType string)`: Decrements durability, replaces destroyed tile with walkable ground (`TileGrass` or `TileWoodFloor`), immediately clearing collision and vision blocking, and returns dropType `"wood"`.
   - In `internal/game/game.go`:
     - In `processInputAndCombat`, integrate barrier chopping into melee attack routines (Axe with dmg 2, Club/Weapon with dmg 1, Shotgun with dmg 2; Unarmed shove with dmg 0 cannot damage barriers).
     - Attack swings detect destructible tiles in swing reach/radius, apply damage, decrement weapon durability, play hit sound, and when a tile is destroyed, spawn an `ecs.Item{Type: "wood"}` entity at the destroyed tile center.
     - Ensure player stepping within 64px collects the dropped `"wood"` item into their inventory.
     - Ensure `"wood"` item rendering and HUD display function properly.
2. Write comprehensive unit and integration tests in `internal/game/world/destruction_test.go` and `internal/game/destruction_combat_test.go`:
   - Test tile durability degradation, perimeter wall indestructibility, collision & vision clearing on destruction.
   - Test weapon attack chops (Axe, Club, Shotgun vs Unarmed), weapon durability consumption on chop, item drop spawning and player pickup into inventory.
   - Multi-barrier breach simulation where player chops down consecutive fences and walks through the cleared gap.
3. Verify with:
   `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go test -v -count=1 ./...`
   and
   `C_INCLUDE_PATH=/usr/include CGO_CFLAGS="-I/usr/include" CC=gcc go build -o bin/game ./cmd/game`
4. Write your handoff report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m4_1/handoff.md` and send a message back when complete.
