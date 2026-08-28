## 2026-08-28T18:59:34Z
You are m1_explorer_fix_1.
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_1
Original Request File: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan File: /home/bryce/code/go-zomboid/PROJECT.md
Project root: /home/bryce/code/go-zomboid

Failure Context from Challenger Reports:
1. `cmd/tools/genassets/main.go:250-265` (`drawVectorPebble`):
   - `drawVectorPebble` uses `setPixel(img, x, y, dropShadow)` with `dropShadow := color.RGBA{0, 0, 0, 45}` instead of `blendPixel(img, x, y, dropShadow)`, overwriting the opaque dirt background and leaving 151 semi-transparent holes in the core diamond.
   - Pebble `{195, 36}` with radius $r_x=7, r_y=4$ spills 18 non-transparent pixels past the isometric diamond boundary ($isoDist > 1.0$).
2. `internal/assets/assets.go`:
   - `Load()` needs thread-safe synchronization (e.g., `sync.Once` or mutex) to prevent data races when called concurrently under `go test -race`.

Mission:
Investigate and design the exact code fixes for `cmd/tools/genassets/main.go` and `internal/assets/assets.go` so that `empirical_challenger_test.go` and `-race` tests pass 100%.
Write your report to `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_1/fix_plan.md` and `handoff.md`.
Send a message to your parent when complete.
