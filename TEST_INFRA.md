# E2E Test Infra: go-zomboid

## Test Philosophy
- Opaque-box and requirement-driven testing.
- Headless execution with CGO enabled (`CC=gcc go test ./...`).
- Verification across 4 distinct test tiers ensuring orthogonal mathematical correctness, seamless rendering, dynamic Dungeon Master behavior, and combat/game loop robustness.

## Feature Inventory & Test Coverage
| # | Feature | Source (Requirement) | Tier 1 | Tier 2 | Tier 3 | Tier 4 |
|---|---------|----------------------|:------:|:------:|:------:|:------:|
| 1 | 2D Orthogonal Coordinate Math | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 2 | Orthogonal Camera Controller | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 3 | Asset Pipeline & Slicing | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 4 | Seamless 2D Orthogonal DrawSystem | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 5 | Top-Down Y-Depth Sorting | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 6 | Bezier Combat Swoosh | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 7 | Dynamic Wave Spawning | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 8 | Randomized Dynamic Loot Drops | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 9 | Day/Night Lighting & Aggression | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 10 | Full Package Suite & Simulation | ORIGINAL_REQUEST §Verification | 5 | 5 | ✓ | ✓ |

## Test Architecture
- **Runner**: `CC=gcc go test -v ./...`
- **Headless Context**: `ebiten.NewImage(1280, 720)` used as render targets in tests.
- **Coverage Tiers**:
  - **Tier 1 (Feature Coverage)**: Isolated unit tests for each coordinate function, camera snap/lerp, asset pointer, DM wave calculation, loot drop roll, day/night lighting.
  - **Tier 2 (Boundary & Corner Cases)**: Extreme coordinate fuzzing ($\pm 10^7$), zero/max zombie caps, edge-of-map spawns, midnight/noon extremes, empty inventory.
  - **Tier 3 (Cross-Feature Combinations)**: Camera movement + mouse unprojection, wave spawning + night aggression scaling + combat cleave, loot drop + player pickup.
  - **Tier 4 (Real-World Application Scenarios)**: Headless continuous multi-minute game simulation (2500+ ticks) exercising day/night cycles, wave spawns, combat, and inventory.

## Real-World Application Scenarios (Tier 4)
| # | Scenario | Features Exercised | Complexity |
|---|----------|--------------------|------------|
| 1 | Continuous Multi-Day Simulation | F1, F2, F7, F8, F9, F10 | High |
| 2 | Midnight Horde Attack & Combat | F1, F6, F7, F9 | High |
| 3 | Scavenge & Dynamic Loot Restock | F1, F4, F8 | Medium |
| 4 | Seamless Map Traversal & FOV | F1, F2, F4, F5 | Medium |
| 5 | Game Reset & State Cleanliness | F1, F2, F7, F8, F10 | Medium |
