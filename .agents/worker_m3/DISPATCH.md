## 2026-08-29T16:04:24Z

You are the QA & Test Specialist Worker implementing Milestone 3: Comprehensive Test Suite Refactoring & E2E Pass.
Working directory: /home/bryce/code/go-zomboid/.agents/worker_m3
Original request path: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project plan path: /home/bryce/code/go-zomboid/PROJECT.md
Test infra path: /home/bryce/code/go-zomboid/TEST_INFRA.md
Survey report: /home/bryce/code/go-zomboid/.agents/explorer_survey_3/handoff.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. An auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & Implementation Details:
1. Refactor Test Suites in `internal/assets/`:
   - Update `assets_test.go`, `assets_stress_test.go`, `challenger_stress_test.go`, `empirical_challenger_test.go`, `m1_stress_verification_test.go` to remove legacy 256x128 diamond / diamond alpha mask assertions.
   - Assert all 49 exported image pointers are non-nil with valid rectangular dimensions matching the active RPG Maker and game sprites.
2. Refactor Test Suites in `internal/game/`:
   - Update `game_test.go`, `camera_test.go`, `camera_empirical_challenger_test.go`, `draw_depth_test.go`, `bezier_combat_test.go`, `game_stress_test.go`, `game_empirical_stress_test.go`, `challenger_tile_render_test.go`:
     - Update `WorldToIso` tests to verify orthogonal Cartesian identity mapping (wx, wy) <-> (sx, sy).
     - Update `ScreenToWorld` and `WorldToScreen` tests to verify orthogonal roundtrip bijection across fuzzed coordinates.
     - Update camera tests (`Snap`, `Update`, viewport corner deltas \Delta wx = \pm 640/Z, \Delta wy = \pm 360/Z, click targeting).
     - Update `DrawSystem` depth sorting to assert strict vertical Y-depth monotonicity (pos.Y).
     - Update `DrawSystem` sprite anchor tests to assert top-down rectangular origins and obstacle anchors.
     - Update `Bezier` combat arc tests to verify control point calculations and screen projection in orthogonal space.
     - Ensure headless simulation stress tests (`TestGameLoopContinuousSimulationStress`) pass smoothly.
3. Verification & Acceptance Criteria:
   - Execute `CC=gcc go test -v ./...`
   - Assert all packages (`internal/ecs`, `internal/assets`, `internal/game/world`, `internal/game`) pass 100% with 0 failures.
   - Verify `CC=gcc go build ./...` passes.
4. Publish `TEST_READY.md` at `/home/bryce/code/go-zomboid/TEST_READY.md` containing test counts across Tiers 1-4 and feature checklist.
5. Write detailed handoff report to `/home/bryce/code/go-zomboid/.agents/worker_m3/handoff.md` and send message to parent.
