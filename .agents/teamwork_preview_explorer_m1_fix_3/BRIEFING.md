# BRIEFING — 2026-08-28T19:01:35Z

## Mission
Investigate test suite assertions in `internal/assets/empirical_challenger_test.go`, `assets_test.go`, and `assets_stress_test.go`, and provide an exact fix plan ensuring `go run ./cmd/tools/genassets` and `CC=gcc go test -race -v ./cmd/tools/genassets/... ./internal/assets/...` pass cleanly.

## 🔒 My Identity
- Archetype: explorer
- Roles: read-only investigation, problem analysis, synthesizing findings, fix plan generation
- Working directory: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_3
- Original parent: f7a8f969-fc3f-4f72-a625-45c03a6444ae
- Milestone: M1 Asset Generation & Asset Management

## 🔒 Key Constraints
- Read-only investigation — do NOT implement directly in source files; write report and fix plan in agent folder
- Output paths: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_3/fix_plan.md` and `handoff.md`
- Inform parent via `send_message`

## Current Parent
- Conversation ID: f7a8f969-fc3f-4f72-a625-45c03a6444ae
- Updated: 2026-08-28T19:01:35Z

## Investigation State
- **Explored paths**: `cmd/tools/genassets/main.go`, `cmd/tools/genassets/genassets_test.go`, `internal/assets/assets.go`, `internal/assets/assets_test.go`, `internal/assets/assets_stress_test.go`, `internal/assets/empirical_challenger_test.go`, `internal/assets/challenger_stress_test.go`.
- **Key findings**:
  1. `drawVectorPebble` in `cmd/tools/genassets/main.go` used `setPixel` with semi-transparent drop shadow, causing 151 inner holes in `dirt.png` (`isoDist <= 0.85`), and pebble `{195, 36}` bled 18 pixels outside the diamond.
  2. `Load()` in `internal/assets/assets.go` mutated global pointer handles on every call without synchronization, causing data races under `-race` and 50-second slowdown.
- **Unexplored areas**: None for M1 scope.

## Key Decisions Made
- Solution 1: Update `drawVectorPebble` with `blendPixel` and `isoDist <= 1.0` check, shift pebble `{195, 36}` to `{185, 42}`.
- Solution 2: Wrap `Load()` in `internal/assets/assets.go` with `sync.Once`.

## Artifact Index
- DISPATCH.md — Task dispatch record
- progress.md — Heartbeat and status tracking
- fix_plan.md — Proposed fix plan and verification steps
- handoff.md — 5-component handoff report
