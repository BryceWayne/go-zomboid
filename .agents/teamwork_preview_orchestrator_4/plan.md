# Plan: Milestone 3 — Camera System QoL Improvements

## Objective
Implement 50% global zoom-out scale in DrawSystem, update IsoToWorld mouse click coordinate math to invert zoom, implement smooth camera tracking (lerping) centering on player (1280x720 screen), and expand visionRadius and FOV culling distance.

## Phase 1: Survey & Exploration
- Spawn 3 Explorers (`teamwork_preview_explorer`) in parallel:
  - Explorer 1: Inspect `DrawSystem` in `internal/game/game.go` and screen scaling matrices, GeoM transformations, HUD separation, and resolution parameters (1280x720).
  - Explorer 2: Inspect `IsoToWorld` and mouse interaction (`UpdateSystem`, input handlers, coordinate transformations) to identify required adjustments for 50% zoom scale.
  - Explorer 3: Inspect `visionRadius`, FOV calculation, tile/entity culling distance (`internal/game/world/map.go`, `internal/game/game.go`), and camera position updates (instant snap vs lerping).
- Collect reports, check consensus, and define exact code modification requirements.

## Phase 2: Implementation (Worker)
- Dispatch `teamwork_preview_worker` to:
  - Apply 0.5 global zoom scale matrix to world rendering in `DrawSystem` while keeping HUD/UI 1:1.
  - Implement smooth camera tracking (lerping) with player dynamic centering on 1280x720 viewport.
  - Update `IsoToWorld` coordinate math and any screen-to-world mouse conversion.
  - Expand `visionRadius` and FOV culling distance to prevent edge pop-in under 50% zoom.
  - Run `CC=gcc go test ./...` and `CC=gcc go run ./cmd/game` verification.

## Phase 3: Review, Challenge & Forensic Audit
- Dispatch 2 `teamwork_preview_reviewer` agents to verify code correctness, boundary conditions, and test passes.
- Dispatch 2 `teamwork_preview_challenger` agents to stress test coordinate transformations, lerp stability, and edge culling.
- Dispatch 1 `teamwork_preview_auditor` for integrity verification (no dummy logic / cheating).

## Phase 4: Gate & Completion
- Verify all gate criteria:
  - 100% test pass (`CC=gcc go test ./...`)
  - Reviewer APPROVE
  - Challenger confirmation
  - Auditor CLEAN
- Report completion back to parent sentinel.
