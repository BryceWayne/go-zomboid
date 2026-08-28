# Project: go-zomboid Enhancement

## Architecture
- **Engine**: Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`) isometric 2.5D renderer and input loop.
- **ECS**: Ark ECS (`github.com/mlange-42/ark`) managing entities (Player, Zombie, Item) and components (Position, Velocity, Sprite, Collider, Player, Zombie, Item).
- **Procedural Asset Pipeline**: `cmd/tools/genassets` generates all PNG textures without external dependencies into `internal/assets/images`, embedded via Go `embed.FS` in `internal/assets`.
- **World & Town Generation**: `internal/game/world` procedural generator creating zoned districts, road networks, multi-room building archetypes, fences, obstacles, and collision/FOV maps.
- **Gameplay Systems**: `internal/game` handling input, item pickup/consumption, inventory, weapon combat (melee and ranged), armor defense & infection mitigation, zombie AI & swarm behavior, HUD rendering, and isometric projection.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | Upgraded Base Character Sprites | Anatomical pixel-art generation for player, zombie, and runner with clothes, hair, wounds, shading | M1 | ORIGINAL_REQUEST §R1 |
| 2 | Upgraded Environment Tiles | Procedural isometric floor diamonds (grass blades, dirt pebbles, wood planks, concrete, asphalt, tile floor) and 3D blocks (brick wall courses, multi-tier trees, fences, debris) | M1 | ORIGINAL_REQUEST §R1 |
| 3 | Upgraded & New Item Sprites | Detailed item sprites: canned food, water bottle, spiked club, fire axe, shotgun, ammo box, tactical armor vest | M1 | ORIGINAL_REQUEST §R1, R2 |
| 4 | Asset Pipeline Integration | Export and expose new images in `internal/assets/assets.go` from `embed.FS` | M1 | ORIGINAL_REQUEST §R1 |
| 5 | Expanded Tile & Collision System | New `TileType` constants with solid/vision-blocking flags and collision bounds checking | M2 | ORIGINAL_REQUEST §R2 |
| 6 | Procedural Town Generator | Algorithmic layout with road network, district zoning, lot subdivision, and varied building placements | M2 | ORIGINAL_REQUEST §R2 |
| 7 | Multi-Room Building Archetypes | Diverse buildings (Residential Homes, Grocery Stores, Police Stations, Pharmacies, Warehouses) with interior rooms | M2 | ORIGINAL_REQUEST §R2 |
| 8 | Environmental Props & Boundaries | Fenced yards, outdoor debris, vegetation clusters, and safe boundary generation | M2 | ORIGINAL_REQUEST §R2 |
| 9 | Contextual ECS Spawning | Place player safely in house, distribute thematic loot in rooms, spawn zombies in world without wall trapping | M2 | ORIGINAL_REQUEST §R2 |
| 10 | Armor ECS Data & Equipping | `ecs.Player` armor fields, 9-slot inventory equipping of armor vest items | M3 | ORIGINAL_REQUEST §R2 |
| 11 | Armor Damage & Infection Mitigation | Reduce incoming damage (e.g. 50% defense), deflection chance for zombie bites/infection, durability decay and breakage | M3 | ORIGINAL_REQUEST §R2 |
| 12 | Armor HUD & Visual Feedback | Dedicated armor durability/defense bar on HUD, text status, and character visual tint/indicator | M3 | ORIGINAL_REQUEST §R2 |
| 13 | Weapon Expansion (Melee & Ranged) | Implement new weapon types: Fire Axe (heavy cleave, durability) and Shotgun + Ammo (ranged spread cone, knockback, noise radius) | M4 | ORIGINAL_REQUEST §R2 |
| 14 | Expanded Combat System | Weapon-specific reach, ammo consumption, noise aggregation alerting zombie swarm, durability tracking | M4 | ORIGINAL_REQUEST §R2 |
| 15 | E2E Headless & Integration Test Suite | Full 4-tier requirement-driven test suite validating all systems headlessly and interactively | M5 | Acceptance Criteria |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Procedural Sprite Enhancements | Rewrite `cmd/tools/genassets/main.go` for rich procedural pixel-art sprites (characters, tiles, items, armor, weapons) and register in `internal/assets` | none | DONE |
| M2 | Environment & Town Generation Updates | Expand `internal/game/world` with procedural road network, district zoning, building archetypes, multi-room floorplans, props, and contextual spawns | M1 | DONE |
| M3 | Armor System & Damage Mitigation | Implement armor fields in `internal/ecs`, equipping logic in `internal/game`, damage mitigation & infection deflection in `processZombies`, and Armor HUD | M1 | DONE |
| M4 | Weapon Expansion & Combat Mechanics | Implement new weapon types (Axe, Shotgun + Ammo) in `internal/game`, projectile/cleave mechanics, noise alerts, and loot integration | M1, M2 | DONE |
| M5 | E2E Integration & Verification | Full test suite execution (`CC=gcc go test ./...`), asset generation verification (`genassets`), and game loop launch verification (`cmd/game`) | M1, M2, M3, M4 | DONE |

## Code Layout
- `cmd/tools/genassets/main.go`: Procedural image asset generator producing all PNGs.
- `internal/assets/images/*.png`: Generated PNG texture files.
- `internal/assets/assets.go`: Asset embedding and global `*ebiten.Image` handles.
- `internal/assets/audio.go`: Procedural PCM sound synthesis.
- `internal/ecs/components.go`: ECS components (Position, Velocity, Sprite, Collider, Player, Zombie, Item).
- `internal/game/world/map.go`: World map data structures, tiles, procedural generator, collision, FOV.
- `internal/game/world/map_test.go`: Unit tests for map and world generation.
- `internal/game/game.go`: Game engine, update system, input/combat, zombie AI, draw system, isometric projection.
- `internal/game/game_test.go`: Unit tests for game systems, projection, and mechanics.
- `internal/game/armor_test.go`: Unit tests for armor system and damage mitigation.
- `internal/game/combat_test.go`: Unit tests for weapon expansion and combat mechanics.
- `cmd/game/main.go`: Game entry point.

## Interface Contracts
### `genassets` ↔ `internal/assets`
- `genassets` outputs 16x16 items (`food.png`, `water.png`, `weapon.png`, `axe.png`, `shotgun.png`, `ammo.png`, `armor.png`), 16x32 entities (`player.png`, `zombie.png`, `runner.png`), 64x32 floors (`grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png`), and 64x64 blocks (`wall.png`, `tree.png`, `fence.png`, `debris.png`).
- `internal/assets/assets.go` exports matching `*ebiten.Image` variables loaded via `Load()`.

### `internal/game/world` ↔ `internal/game`
- `world.NewMap(width, height)` returns `*world.Map` with fully populated procedural town layout, buildings, roads, and props.
- `world.Map.IsColliding(x, y, w, h float64) bool`: evaluates solid tiles (`TileWall`, `TileTree`, `TileFence`, `TileDebris`, etc.).
- `world.Map.BlocksVision(x, y int) bool`: evaluates vision blocking for FOV raycasting.
- `world.Map` provides spawn metadata: `PlayerSpawn FloatPoint`, `LootSpawns []LootSpawn{Type, X, Y, RoomType}`, `ZombieSpawns []FloatPoint`.

### `internal/ecs` ↔ `internal/game`
- `ecs.Player`:
  ```go
  type Player struct {
      Health             float64
      Hunger             float64
      Thirst             float64
      Inventory          []string
      WeaponEquipped     bool
      WeaponType         string
      WeaponDurability   int
      ArmorEquipped      bool
      ArmorType          string
      ArmorDefense       float64
      ArmorDurability    int
      ArmorMaxDurability int
      InfectionResist    float64
      AttackCooldown     int
      Dead               bool
      Infected           bool
      FacingX, FacingY   float64
  }
  ```
