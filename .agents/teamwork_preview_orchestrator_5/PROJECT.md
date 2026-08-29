# Project: External Asset Ingestion & World Integration

## Architecture
- **Asset Ingestion Pipeline (`internal/assets`)**: Embeds discrete PNGs and sprite sheets from `internal/assets/images/` via Go `embed.FS`. Native `Load()` decodes PNGs to `*ebiten.Image`.
- **World Map System (`internal/game/world`)**: 100x100 tile grid. `TileType` constants define collision (`IsSolid`), vision occlusion (`BlocksVision`), and floor vs obstacle status (`IsFloor`).
- **Rendering & Depth-Sorting System (`internal/game`)**: Multi-pass rendering. Isometric projection with $isoY = (worldX + worldY)/2$. Sprites collected with `Depth = worldX + worldY` and stably sorted by depth.
- **Game Lifecycle (`cmd/game`)**: Standard Ebiten entry point initializing assets, audio, ECS world, and game loop.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | Retire Procedural Asset Tool | Completely delete `cmd/tools/genassets` directory, root binary, and clean up direct invocation tests | M1 | R1 |
| 2 | Ingest External PNG Assets | Copy PNG assets from `context/` into `internal/assets/images/` (ignoring `.DS_Store`, PSDs, zone identifiers) | M1 | R2 |
| 3 | Native Asset Loader Extension | Update `internal/assets/assets.go` to load new `*ebiten.Image` pointers (`BenchImage`, `ChestImage`, `SculptureImage`, etc.) | M1 | R2 |
| 4 | Define World TileType Constants | Add `TileBench` (16), `TileChest` (17), `TileSculpture` (18), `TileBush` (19), `TileFlower` (20), `TileStone` (21) in `internal/game/world/map.go` | M2 | R3 |
| 5 | Tile Physical Properties | Implement `IsSolid()`, `BlocksVision()`, `IsFloor()`, `String()` for all new tile types | M2 | R3 |
| 6 | World Generation Placement | Update `map.go` generation to place new prop tiles across town while preserving original 10 tile types | M2 | R3 |
| 7 | DrawSystem Depth-Sorting & Render | Update `internal/game/game.go` `DrawSystem.Draw` to render base ground in Pass 1 and depth-sorted prop sprites in Pass 2 | M3 | R3 |
| 8 | Comprehensive E2E Verification | Ensure `CC=gcc go test ./...` passes across all packages and game launches cleanly | M4 | Acceptance Criteria |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Asset Ingestion & Retirement | Delete `cmd/tools/genassets`, copy PNGs from `context/` to `internal/assets/images/`, update `assets.go` & tests | None | DONE |
| M2 | World TileType & Map Logic | Add tile constants, properties, world generation in `internal/game/world/map.go` & tests | M1 | DONE |
| M3 | DrawSystem Depth Sorting | Update `internal/game/game.go` `DrawSystem` rendering and depth sorting for new tiles | M1, M2 | DONE |
| M4 | E2E & Acceptance Verification | Verify `CC=gcc go test ./...` and `cmd/game` launch | M1, M2, M3 | DONE |

## Interface Contracts
### `internal/assets` -> `internal/game` & `internal/game/world`
- Exported variables in `internal/assets`:
  - `BenchImage *ebiten.Image`
  - `ChestImage *ebiten.Image`
  - `SculptureImage *ebiten.Image` (and `Sculpture1Image`, `Sculpture2Image`)
  - `BushImage *ebiten.Image`
  - `FlowerImage *ebiten.Image`
  - `StoneImage *ebiten.Image`
- Method: `assets.Load() error` initializes all legacy and new image pointers without error.

### `internal/game/world` -> `internal/game`
- Exported constants in `internal/game/world`:
  - `TileBench TileType = 16`
  - `TileChest TileType = 17`
  - `TileSculpture TileType = 18`
  - `TileBush TileType = 19`
  - `TileFlower TileType = 20`
  - `TileStone TileType = 21`
- Methods on `TileType`:
  - `IsSolid() bool` -> true for Bench, Chest, Sculpture, Stone; false for Bush, Flower
  - `BlocksVision() bool` -> false for all new props (only Wall blocks vision)
  - `IsFloor() bool` -> false for all new props (drawn as vertical sprites)
  - `String() string` -> "bench", "chest", "sculpture", "bush", "flower", "stone"

## Code Layout
- `internal/assets/`: Asset loading and embedded image files (`internal/assets/images/`). Owned by M1.
- `internal/game/world/`: World generation, tile properties, coordinate system. Owned by M2.
- `internal/game/`: ECS systems, `game.go` `DrawSystem`, depth sorting. Owned by M3.
- `cmd/game/`: Game executable entry point. Owned by M4.
