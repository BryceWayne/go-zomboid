# Progress Log

Last visited: 2026-08-28T19:04:25Z

## Status: All Milestone 1 Fixes Applied & Verified
- [x] Initialized DISPATCH.md and BRIEFING.md
- [x] Read fix plans from explorers (`m1_fix_1`, `m1_fix_2`, `m1_fix_3`)
- [x] Inspected `cmd/tools/genassets/main.go` and `internal/assets/assets.go`
- [x] Applied fixes to `cmd/tools/genassets/main.go`:
  - In `drawVectorPebble`: replaced `setPixel` with `blendPixel(img, x, y, dropShadow)` for drop shadow rendering; added `isoDist <= 1.0` boundary checks to both drop shadow and pebble body loops.
  - In `generateDirt`: adjusted pebble coordinate from `{195, 36}` inward to `{185, 42}` so it is fully contained inside the 2:1 isometric diamond.
- [x] Applied fixes to `internal/assets/assets.go`:
  - Added `var loadOnce sync.Once` and wrapped `Load()` initialization within `loadOnce.Do(func() { ... })`.
- [x] Regenerated all 27 assets via `go run ./cmd/tools/genassets`.
- [x] Verified tests:
  - `CC=gcc go test -p 1 -v ./internal/assets/... ./cmd/tools/genassets/...` -> PASS
  - `CC=gcc go test -race -p 1 -v ./internal/assets/... ./cmd/tools/genassets/...` -> PASS
  - `CC=gcc go test ./...` -> PASS
  - `CC=gcc go test -race ./...` -> PASS
- [x] Prepared handoff report and notified parent.
