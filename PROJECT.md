# Project: go-zomboid 2D Orthogonal Engine & Dungeon Master System

## Architecture
The `go-zomboid` project is an apocalyptic zombie survival 2D game built on Go, Ebitengine, and Ark ECS (`github.com/mlange-42/ark/ecs`).
- `cmd/game`: Binary entry point, window configuration (1280x720), asset loading, and Ebitengine game loop runner.
- `internal/ecs`: Pure data component definitions (`Player`, `Zombie`, `Item`, `Position`, `Velocity`, `Sprite`, `Collider`).
- `internal/assets`: Embedded assets loader (`images/*`, `audio.go`), sprite pointers, and texture slicing.
- `internal/game/world`: Procedural 100x100 tile map generator, AABB collision detection, 2D FOV raycasting, and district building placement.
- `internal/game`:
  - `game.go`: `Game` state, `Camera` tracking, `UpdateSystem` (combat, movement, inventory, AI), and `DrawSystem` (rendering, lighting, UI).
  - `dm.go`: `DungeonMaster` simulation engine managing dynamic wave spawning, difficulty curves, loot replenishment, and day/night aggression modifiers.
  - Coordinate system: Strict 2D Orthogonal (top-down) grid mapping $(wx, wy) \leftrightarrow (sx, sy)$.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | 2D Orthogonal Math | Replace 2:1 dimetric isometric coordinate conversions with 1:1 bijective orthogonal Cartesian math (`WorldToIso`, `IsoToWorld`, `ScreenToWorld`, `WorldToScreen`). | M1 | ORIGINAL_REQUEST §R1 |
| 2 | Orthogonal Camera Controller | Camera snaps and smoothly lerps in Cartesian world coordinates $(wx, wy)$ tracking the player; accurate mouse cursor unprojection. | M1 | ORIGINAL_REQUEST §R1 |
| 3 | Asset Pipeline & Slicing | Support 2D RPG Maker / external tilesets; load non-nil textures with rectangular bounding dimensions. | M2 | ORIGINAL_REQUEST §R1 |
| 4 | Seamless 2D Orthogonal DrawSystem | Render floor tiles top-left aligned $(tx \cdot \text{TileSize}, ty \cdot \text{TileSize})$ with proper scaling to guarantee 0 black gaps or diamond voids. | M2 | ORIGINAL_REQUEST §R1 |
| 5 | Top-Down Y-Depth Sorting | Sort standing props, entities (player, zombies), and items by vertical coordinate ($Y$) for natural occlusion. | M2 | ORIGINAL_REQUEST §R1 |
| 6 | Orthogonal Bezier Combat Swoosh | Project attack swing arcs directly to orthogonal screen space without isometric skewing. | M2 | ORIGINAL_REQUEST §R1 |
| 7 | Dungeon Master Wave Spawning | Dynamic zombie wave spawning scaling difficulty over time ($1.0 \to 3.5+$), spawning at safe off-screen perimeters on non-solid tiles. | M3 | ORIGINAL_REQUEST §R2 |
| 8 | Randomized Dynamic Loot Drops | Weighted loot drops on zombie kill (ammo, food, water, weapons, antidote, armor) + ambient supply restocks. | M3 | ORIGINAL_REQUEST §R2 |
| 9 | Day/Night Lighting & Night Aggression | Multi-phase ambient lighting (dawn, day, dusk, night) and scaling zombie speed, hearing, and vision ranges at night. | M3 | ORIGINAL_REQUEST §R2 |
| 10 | Comprehensive Test Suite Refactoring | Update all legacy isometric test assertions in `internal/assets` and `internal/game` to orthogonal assertions; add DM unit tests. | M4 | ORIGINAL_REQUEST §Acceptance |
| 11 | Final E2E Pass & Adversarial Hardening | Verify `CC=gcc go test ./...` 100% pass, run continuous simulation stress tests, forensic audit, and headless game loop verification. | M5 | ORIGINAL_REQUEST §Acceptance |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | 2D Orthogonal Engine Overhaul (R1) | Coordinate math (`WorldToIso`, `IsoToWorld`, `ScreenToWorld`, `WorldToScreen`), `Camera` tracking, `DrawSystem` top-left seamless rendering, RPG Maker asset scaling/slicing, top-down Y-depth sorting, Bezier combat arcs. | none | DONE |
| M2 | Dungeon Master Simulation (R2) | `internal/game/dm.go`, dynamic zombie wave spawning (scaling threat), randomized loot drops (kill drops + ambient restock), day/night lighting phases and nighttime enemy aggression scaling. | M1 | DONE |
| M3 | Comprehensive Test Suite & E2E Pass | Refactor all asset and game tests to match 2D orthogonal geometry and new assets, add comprehensive Dungeon Master unit & continuous simulation tests, achieve 100% `CC=gcc go test ./...` pass. | M1, M2 | DONE |
| M4 | Adversarial Hardening & Forensic Audit | White-box adversarial testing, edge case stress testing, and forensic integrity audit verification. | M3 | DONE |

## Interface Contracts
### Coordinate Engine ↔ Camera & Input
- `WorldToIso(wx, wy float64) (isoX, isoY float64)`: Returns `(wx, wy)` (orthogonal identity for backwards compatibility).
- `IsoToWorld(isoX, isoY float64) (wx, wy float64)`: Returns `(isoX, isoY)`.
- `ScreenToWorld(screenX, screenY, camX, camY float64) (wx, wy float64)`: Returns `(camX + (screenX-640.0)/zoom, camY + (screenY-360.0)/zoom)`.
- `WorldToScreen(wx, wy, camX, camY float64) (sx, sy float64)`: Returns `((wx-camX)*zoom + 640.0, (wy-camY)*zoom + 360.0)`.

### DrawSystem ↔ Assets & World Map
- Floor rendering: Draws `world.Map.GetTile(tx, ty)` scaled to tile dimensions at `WorldToScreen(tx*TileSize, ty*TileSize)`.
- Prop/Entity rendering: Anchored with depth key `Depth = entity.Y + offset` for strict vertical top-down occlusion.

### Dungeon Master ↔ UpdateSystem & Game State
- `NewDungeonMaster(world *arkecs.World, gameMap *world.Map, config DungeonMasterConfig) *DungeonMaster`
- `Update(timeOfDay float64, playerPos ecs.Position) (newZombies int, newLoot int)`
- `HandleZombieDeath(wx, wy float64)`: Rolls weighted loot drop and instantiates item entity at death location.
- `GetAggressionModifiers(timeOfDay float64) (speedMult, noiseMult, visionMult float64)`: Scales AI parameters based on time.

## Code Layout
- `cmd/game/main.go`: Game entry point.
- `internal/ecs/components.go`: ECS data components.
- `internal/assets/assets.go`: Asset loading and tile image pointers.
- `internal/assets/audio.go`: Audio and sound generator.
- `internal/game/world/map.go`: Map model, procedural generation, AABB collision, FOV.
- `internal/game/game.go`: Game loop, Camera, UpdateSystem, DrawSystem, Coordinate math.
- `internal/game/dm.go`: Dungeon Master simulation logic.
