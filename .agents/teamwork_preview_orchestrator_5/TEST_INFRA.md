# E2E Test Infra: External Asset Ingestion

## Test Philosophy
- Opaque-box, requirement-driven and unit/integration verification.
- Covers Requirements R1, R2, R3 and all acceptance criteria.

## Feature Inventory & Test Coverage
| # | Feature | Requirement | Tier 1 (Unit) | Tier 2 (Boundary) | Tier 3 (Integration) | Tier 4 (E2E) |
|---|---------|-------------|:-------------:|:-----------------:|:--------------------:|:------------:|
| 1 | Retirement of `cmd/tools/genassets` | R1 | Directory absence | Root binary absence | Clean repo build | No leftover calls |
| 2 | Ingestion of PNG assets | R2 | PNG file presence | Non-empty bytes/dims | Embed FS validity | All image pointers loaded |
| 3 | Asset loader native load | R2 | `assets.Load()` non-nil | Texture dims > 0 | Error handling on missing | Concurrent Load() safety |
| 4 | TileType constants & properties | R3 | `IsSolid`, `BlocksVision` | `IsFloor`, `String` | Map generation counts | 10 legacy types preserved |
| 5 | DrawSystem depth sorting | R3 | Sprite pass generation | Isometric Y depth order | Player/Zombie occlusion | Zero visual hole on ground |
| 6 | Full test suite passing | Acceptance | `go test ./...` exit 0 | Race detector / vet | Map stress tests | Game launch clean |

## Acceptance Verification Criteria
- `cmd/tools/genassets` does NOT exist on disk.
- `CC=gcc go test ./...` passes 100%.
- `CC=gcc go run ./cmd/game` runs and initializes without panic/error.
