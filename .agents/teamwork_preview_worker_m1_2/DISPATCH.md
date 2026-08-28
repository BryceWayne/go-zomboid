## 2026-08-28T19:02:38Z
You are m1_worker_2.
Your working directory is: /home/bryce/code/go-zomboid/.agents/teamwork_preview_worker_m1_2
Original Request File: /home/bryce/code/go-zomboid/.agents/ORIGINAL_REQUEST.md
Project Plan File: /home/bryce/code/go-zomboid/PROJECT.md
Project root: /home/bryce/code/go-zomboid

Explorer fix plans to read:
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_1/fix_plan.md
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_2/fix_plan.md
- /home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_3/fix_plan.md

Task:
Apply fixes for Milestone 1:
1. In `cmd/tools/genassets/main.go`:
   - In `drawVectorPebble`: use `blendPixel(img, x, y, dropShadow)` instead of `setPixel` for dropShadow so it blends over the opaque background without punching holes. Add `isoDist <= 1.0` boundary check to both shadow and body loops.
   - In `generateDirt`: shift pebble at `{195, 36}` inward to `{185, 42}` so it is fully contained inside the 2:1 isometric diamond.
2. In `internal/assets/assets.go`:
   - Add `var loadOnce sync.Once` and wrap `Load()` logic inside `loadOnce.Do(func() { ... })` so that `Load()` is thread-safe and free of data races under `go test -race`.
3. Regenerate all assets: `go run ./cmd/tools/genassets`.
4. Run tests:
   - `CC=gcc go test -v ./internal/assets/... ./cmd/tools/genassets/...`
   - `CC=gcc go test -race -v ./internal/assets/... ./cmd/tools/genassets/...`
   - `CC=gcc go test ./...`
   Verify all tests pass without errors or race conditions.
