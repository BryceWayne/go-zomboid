# Milestone 1 Remediation Handoff Report

## 1. Observation
- **`cmd/tools/genassets/main.go` (Pebble Shadow Blending & Boundary Checks)**:
  - In `drawVectorPebble` (lines 250-290), `dropShadow` (`color.RGBA{0, 0, 0, 45}`) was previously written directly with `setPixel(img, x, y, dropShadow)` without blending over the base background, causing semi-transparent hole punctures ($A=45$) in opaque tiles.
  - In `drawVectorPebble`, there were no isometric boundary checks ($isoDist \le 1.0$) in either the shadow loop or body loop.
  - In `generateDirt` (line 673), pebble index 4 was positioned at `{195, 36}` with radius $r_x=7, r_y=4$, reaching $x=202$ ($isoDist = 1.0117 > 1.0$), bleeding 18 pixels outside the isometric diamond.
- **`internal/assets/assets.go` (Asset Loading Synchronization)**:
  - In `internal/assets/assets.go` (lines 53-88), `Load()` performed un-synchronized writes to 27 exported `*ebiten.Image` pointers on every call, causing write/write and read/write data races when accessed concurrently under `go test -race`.
- **Test Executions**:
  - `CC=gcc go test -p 1 -v ./internal/assets/... ./cmd/tools/genassets/...` -> `PASS`
  - `CC=gcc go test -race -p 1 -v ./internal/assets/... ./cmd/tools/genassets/...` -> `PASS`
  - `CC=gcc go test ./...` -> `PASS`
  - `CC=gcc go test -race ./...` -> `PASS` (all packages passed with 0 data races)

## 2. Logic Chain
1. In `cmd/tools/genassets/main.go`:
   - Updated `drawVectorPebble` to calculate `isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0` in both shadow and body loops, and only draw pixels where `isoDist <= 1.0`.
   - Replaced `setPixel(img, x, y, dropShadow)` with `blendPixel(img, x, y, dropShadow)`. With alpha blending on an opaque surface, the shadow darkens the dirt texture while leaving the alpha channel fully opaque ($A=255$).
   - Relocated dirt pebble from `{195, 36}` inward to `{185, 42}` where $isoDist = 0.7852$, ensuring the entire pebble body and shadow are strictly inside $isoDist \le 0.85$.
2. In `internal/assets/assets.go`:
   - Added `var loadOnce sync.Once` and wrapped the asset loading assignments inside `loadOnce.Do(func() { ... })`.
   - Ensures `Load()` executes exactly once even when invoked concurrently across multiple goroutines, making pointer queries safe and deterministic without race conditions.
3. Executed `go run ./cmd/tools/genassets` to regenerate all 27 assets in `internal/assets/images/`.
4. Validated that all unit tests, empirical challenger tests, determinism checks, and race detector tests pass cleanly across the repository.

## 3. Caveats
- No caveats. All 27 procedural assets compile, load, and satisfy all empirical geometry and contract bounds.

## 4. Conclusion
Milestone 1 fixes are complete and verified. Isometric floor geometry violations and data races in asset loading have been resolved.

## 5. Verification Method
To independently verify the implementation:
```bash
# 1. Regenerate procedural assets
go run ./cmd/tools/genassets

# 2. Run asset and generator tests with race detection
CC=gcc go test -race -p 1 -v ./internal/assets/... ./cmd/tools/genassets/...

# 3. Run full project test suite with race detector
CC=gcc go test -race ./...
```
