## 2026-08-28T17:26:58Z

You are a Worker subagent (teamwork_preview_worker_m2_1).
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_1
Project root: /home/bryce/code/go-zomboid
Original Request: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan: /home/bryce/code/go-zomboid/PROJECT.md

Scope: Milestone 2 - Environment & Town Generation Updates

Explorer Handoff References:
- Town Layout & Zoning: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_1/handoff.md and /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_1/proposed_map.go
- Building Archetypes: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_2/handoff.md
- Map Integration, Tests & Game Rendering: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3/handoff.md, /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3/proposed_map.go, /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3/proposed_map_test.go, and /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m2_3/proposed_game_patch.go

Write Ownership:
You own and may modify:
- `internal/game/world/map.go`
- `internal/game/world/map_test.go`
- `internal/game/game.go`
- `internal/game/game_test.go`

Tasks:
1. Implement the expanded Tile System (10 TileTypes, `IsSolid()`, `BlocksVision()`, `IsFloor()`) and Procedural Town Generation in `internal/game/world/map.go`:
   - Multi-tier road network: asphalt avenues with concrete sidewalks, connector streets.
   - 4 thematic districts with multi-room building archetypes: Residential Houses (Living Room, Kitchen, Bedroom, Bathroom), Grocery Store, Pharmacy/Clinic, Police Station (Lobby, Office, Armory, Cells), Industrial Warehouse (Bay, Foreman Office, Loading Dock).
   - Fenced yards, outdoor debris, and vegetation clusters.
   - Collision detection (`IsColliding`) checking all solid tiles (`TileWall`, `TileTree`, `TileFence`, `TileDebris`).
   - FOV raycasting (`CalculateFOV`) occluded by `BlocksVision()` (`TileWall`).
   - Spawn metadata (`PlayerSpawn`, `Buildings`, `LootSpawns`, `ZombieSpawns`).
2. Update `internal/game/world/map_test.go` with comprehensive tests verifying tile properties, town generation, safe player spawn in house, contextual loot spawns, zombie non-trapping, collision AABB, and FOV occlusion.
3. Update `internal/game/game.go`:
   - `Reset()` to use `gameMap.PlayerSpawn`, `gameMap.LootSpawns`, and `gameMap.ZombieSpawns`.
   - `DrawSystem.Draw` to render all 6 ground floor diamond types (`Grass`, `Dirt`, `Wood`, `Asphalt`, `Concrete`, `TileFloor`), 4 vertical obstacle types (`Wall`, `Tree`, `Fence`, `Debris`), and all 7 item types (`food`, `water`, `weapon`, `axe`, `shotgun`, `ammo`, `armor`).
4. Update `internal/game/game_test.go` to test isometric coordinate transformations and game reset initialization.
5. Run `CC=gcc go test -v ./...` and `CC=gcc go build -o bin/game ./cmd/game` to verify all builds and tests pass cleanly.
6. Document your implementation and verification in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m2_1/handoff.md`.
