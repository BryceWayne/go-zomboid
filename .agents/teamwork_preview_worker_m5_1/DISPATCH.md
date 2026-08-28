## 2026-08-28T17:46:14Z

Scope: Milestone 5 - E2E Integration & Full System Verification

Tasks:
1. Verify Asset Generation:
   - Run `go run ./cmd/tools/genassets`.
   - Verify that all 20 PNG asset files exist in `internal/assets/images/` and are non-empty.
2. Verify Full Test Suite:
   - Run `CC=gcc go test -count=1 -v ./...`.
   - Verify 100% pass across all packages (`cmd/tools/genassets`, `internal/assets`, `internal/ecs`, `internal/game`, `internal/game/world`).
3. Verify Binary Compilation:
   - Run `CC=gcc go build -o bin/game ./cmd/game`.
   - Verify `bin/game` executable exists.
4. Verify Game Loop Headless Launch & Stability:
   - Run Ebitengine headless simulations (e.g. `TestGameResetStress`, `TestGameLoopContinuousSimulationStress`) verifying continuous 2000+ frame execution without panics, memory leaks, or NaN velocities.
5. Create `/home/bryce/code/go-zomboid/TEST_READY.md` summarizing the full test suite and pass/fail criteria per `TEST_INFRA.md`.
6. Document your findings in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m5_1/handoff.md`.
