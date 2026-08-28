# BRIEFING — 2026-08-28T12:43:20-05:00

## Mission
Milestone 4 - Implement Weapon Expansion (Axe, Shotgun, Bat, Unarmed) & Combat Mechanics & HUD & Unit Tests

## 🔒 My Identity
- Archetype: worker
- Roles: implementer, qa, specialist
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m4_1
- Original parent: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Milestone: Milestone 4 - Weapon Expansion & Combat Mechanics Implementation

## 🔒 Key Constraints
- Write ownership: internal/ecs/components.go, internal/game/game.go, internal/game/combat_test.go
- Do not hardcode test results or create facade implementations
- Genuine implementation with thorough behavior-driven tests
- CC=gcc go test -v ./... and CC=gcc go build -o bin/game ./cmd/game must pass

## Current Parent
- Conversation ID: efb9db38-c509-4c3c-ad0a-53ad2f86b201
- Updated: not yet

## Task Summary
- **What to build**: Weapon expansion supporting Fire Axe (32px cleave reach/radius, 12 durability), Shotgun (ammo consumption, 160px spread cone +-22.5 deg, 400px noise pulse, dry fire shove), Spiked Bat (24px reach, 5 durability), Unarmed fist shove, Weapon HUD text & reticle styling, and comprehensive unit tests in `combat_test.go`.
- **Success criteria**: All unit tests pass, compilation succeeds, game launches cleanly.
- **Interface contracts**: PROJECT.md
- **Code layout**: internal/ecs, internal/game

## Change Tracker
- **Files modified**:
  - `internal/ecs/components.go`: Added `WeaponType string` to `ecs.Player`.
  - `internal/game/game.go`: Added weapon equipping (keys 1-9 for weapon, axe, shotgun), combat handling for shotgun (ammo consumption, spread cone, 400px noise pulse, dry fire shove), fire axe (32px cleave reach/radius), bat (24px reach), unarmed shove, reticle color coding, and HUD weapon text at Y=95.
  - `internal/game/combat_test.go`: Created comprehensive 16-test suite covering all combat mechanics.
- **Build status**: CC=gcc go test ./... PASS, CC=gcc go build -o bin/game ./cmd/game PASS
- **Pending issues**: None

## Quality Status
- **Build/test result**: PASS (16 new unit tests in combat_test.go + all existing project tests pass)
- **Lint status**: Clean
- **Tests added/modified**: `internal/game/combat_test.go` (16 unit tests)

## Key Decisions Made
- Use exact geometric spread cone cos(theta) >= 0.9238795325112867 (approx math.Cos(22.5 deg)) with point-blank range < 24px.
- Use 400px noise pulse to set Chasing=true and WanderTimer=0 for all zombies within 400px on shotgun firing with ammo.
- Display weapon status at Y=95 with weapon name, durability, and ammo count when shotgun equipped.
- Dry fire on shotgun when out of ammo performs defensive butt shove without deducting durability or alerting horde.

## Artifact Index
- .agents/teamwork_preview_worker_m4_1/DISPATCH.md
- .agents/teamwork_preview_worker_m4_1/BRIEFING.md
- .agents/teamwork_preview_worker_m4_1/progress.md
- .agents/teamwork_preview_worker_m4_1/handoff.md
