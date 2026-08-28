# E2E Test Infra: go-zomboid (Milestone 2)

## Test Philosophy
- Opaque-box, requirement-driven testing based directly on `ORIGINAL_REQUEST.md`.
- No dependency on implementation design.
- Complete coverage across all 4 Tiers:
  - **Tier 1: Feature Coverage** (≥5 tests per feature covering representative happy paths).
  - **Tier 2: Boundary & Corner Cases** (≥5 tests per feature covering limits, edge coordinates, empty inputs, extreme ranges).
  - **Tier 3: Cross-Feature Interactions** (Pairwise coverage across major feature interactions: attack + movement + armor, 4x coordinate scaling + FOV + collision, Bezier swoosh + weapon durability + facing angle).
  - **Tier 4: Real-World Application Scenarios** (Realistic end-to-end survival scenarios exercising combat, exploration, day/night cycles, and high-res rendering loops).

## Feature Inventory & Test Coverage Goals
| # | Feature | Source (Requirement) | Tier 1 | Tier 2 | Tier 3 | Tier 4 |
|---|---------|----------------------|:------:|:------:|:------:|:------:|
| 1 | 4x Floor Tiles (256x128) | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 2 | 4x Obstacles/Props (256x256) | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 3 | 4x Entities (64x128) | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 4 | 4x Items/Weapons (64x64) | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 5 | Geometric Vector Overlays | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ | ✓ |
| 6 | Engine TileSize (128) & Math | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 7 | DrawSystem Anchors & Camera | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 8 | Entity Physics & Speeds | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 9 | Combat & AI Range Scaling | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ | ✓ |
| 10 | Bezier Attack Curve Math | ORIGINAL_REQUEST §R3 | 5 | 5 | ✓ | ✓ |
| 11 | Vector Attack Swoosh Rendering | ORIGINAL_REQUEST §R3 | 5 | 5 | ✓ | ✓ |
| 12 | Weapon-Specific Swoosh Styles | ORIGINAL_REQUEST §R3 | 5 | 5 | ✓ | ✓ |

## Test Architecture
- Test Runner Command: `CC=gcc go test -v ./...`
- Determinism Command: `go test -v ./cmd/tools/genassets`
- Game Launch Verification: `CC=gcc go test -run TestHeadlessGameLoop ./internal/game` (or `CC=gcc go run ./cmd/game` with headless/smoke verification).
- Pass/Fail Semantics: Zero non-zero exit codes, 100% test pass rate, all assets non-nil, dimensions verified.

## Real-World Application Scenarios (Tier 4)
| # | Scenario | Features Exercised | Complexity |
|---|----------|--------------------|------------|
| 1 | High-Res Isometric Town Exploration & FOV Raycasting | F1, F2, F6, F7, F8 | High |
| 2 | Fire Axe Cleave Combat & Bezier Swoosh Arc under Right-Click Aim | F3, F4, F9, F10, F11, F12 | High |
| 3 | Multi-Weapon Durability Lifecycle & Armor Mitigation with 4x Physics | F4, F8, F9, F10, F11 | High |
| 4 | Night Survival Horde Encounter with Shotgun Blast & Acoustic Pulse | F3, F4, F8, F9, F11, F12 | High |
| 5 | Full Procedural Asset Regeneration Determinism & Hash Stability | F1, F2, F3, F4, F5 | Medium |

## Coverage Thresholds
- Tier 1: ≥ 60 test cases (5 × 12 features)
- Tier 2: ≥ 60 test cases (5 × 12 features)
- Tier 3: ≥ 12 pairwise cross-feature test cases
- Tier 4: ≥ 5 realistic end-to-end workload scenarios
- **Total Suite**: ≥ 137 verified test assertions
