# E2E Test Infra: go-zomboid Enhancement

## Test Philosophy
- Requirement-driven, opaque-box and headless unit/integration testing.
- Methodology: Category-Partition + Boundary Value Analysis + Pairwise Combinatorial Testing + Real-World Workload Testing.

## Feature Inventory & Test Mapping
| # | Feature | Source | Tier 1 (Feature) | Tier 2 (Boundary) | Tier 3 (Pairwise) | Tier 4 (Scenario) |
|---|---------|--------|:----------------:|:-----------------:|:-----------------:|:-----------------:|
| 1 | Procedural Asset Generator (`genassets`) | R1 | 5 tests | 5 tests | ✓ | ✓ |
| 2 | Asset Loading & Decoding (`assets.go`) | R1 | 5 tests | 5 tests | ✓ | ✓ |
| 3 | Procedural Town & Road Network | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 4 | Building Archetypes & Room Spawns | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 5 | World Collision & FOV | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 6 | Armor Equipping & Inventory Integration | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 7 | Armor Damage & Infection Mitigation | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 8 | Armor Durability Decay & Breakage | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 9 | Weapon Expansion (Axe, Shotgun, Ammo) | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 10 | Combat Range, Cleave & Noise Alert | R2 | 5 tests | 5 tests | ✓ | ✓ |
| 11 | Headless Game Loop & Ebitengine Startup | Acceptance | 5 tests | 5 tests | ✓ | ✓ |

## Test Architecture
- Unit and subsystem test runner: `CC=gcc go test -v ./...`
- Asset generator runner: `go run ./cmd/tools/genassets`
- Game binary build runner: `CC=gcc go build -o bin/game ./cmd/game`
- Interactive engine smoke runner: `CC=gcc go run ./cmd/game`

## Real-World Application Scenarios (Tier 4)
| # | Scenario | Features Exercised | Expected Outcome |
|---|----------|--------------------|------------------|
| 1 | Fresh Game Initialization | Town Gen, Player Spawn, Asset Load | Map initialized, player in house, no collision panic |
| 2 | Armor Scavenge & Combat Defense | Armor Pickup, Equip, Zombie Attack | Damage mitigated, infection deflected, durability decremented |
| 3 | Heavy Zombie Swarm Combat with Axe | Axe Equip, Cleave Attack, Durability | Multiple zombies killed in swing radius, durability reduced |
| 4 | Ranged Combat & Shotgun Noise Alert | Shotgun + Ammo Equip, Blast, Alert | Pellet spread hits target, noise attracts distant zombies |
| 5 | Survival Cycle: Scavenge, Eat, Wear Armor | Inventory management, Hunger/Thirst, Defense | Survival metrics sustained, combat defense verified |
