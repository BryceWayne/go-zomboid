# Original User Request

## Initial Request — 2026-08-29T15:12:15Z

User Goal & Requirements:
Completely replace the procedural asset generation system with the new external PNG assets located in the `context/` directory.
Integrity mode: demo
Requested team: full team

Requirements:
- R1. Retire Procedural Generation: Completely delete the `cmd/tools/genassets` directory and its contents. The procedural asset generation pipeline is permanently retired.
- R2. External Asset Ingestion: Copy the external PNG files from the `context/` directory into `internal/assets/images/`. Update `internal/assets/assets.go` to load these specific images into `ebiten.Image` variables.
- R3. Infer and Implement New Logic: Analyze the imported assets (e.g., Benches, Chests, Sculptures) and automatically infer their mapping into the game world. Create new `TileType` constants in `internal/game/world/map.go` and update the `DrawSystem` in `internal/game/game.go` to properly render and depth-sort any objects that did not previously exist.

Acceptance Criteria:
- The `cmd/tools/genassets` directory no longer exists on disk.
- `internal/assets/assets.go` successfully loads the new PNG files natively.
- Running `CC=gcc go test ./...` passes all existing map and loading tests.
- Running `CC=gcc go run ./cmd/game` successfully launches the game without crashing, and the new world objects are visibly rendered on the map.
