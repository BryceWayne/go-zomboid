# Milestone 1 Investigation Handoff Report: Asset Generator & Assets Loading Fixes

**Target Milestone**: Milestone 1 Remediation  
**Author**: `m1_explorer_fix_2`  
**Date**: 2026-08-28T19:01:00Z  

---

## 1. Observation

Direct observations and evidence collected from code inspection and test execution:

1. **`images/dirt.png` Alpha Hole Punctures & Diamond Boundary Bleed**:
   - In `cmd/tools/genassets/main.go:250-265`:
     ```go
     func drawVectorPebble(img *image.RGBA, cx, cy int, rx, ry float64, base, light, shadow color.RGBA) {
         dropShadow := color.RGBA{0, 0, 0, 45}
         ...
         for y := minY; y <= maxY; y++ {
             for x := minX; x <= maxX; x++ {
                 dx := float64(x-(cx+2)) / rx
                 dy := float64(y-(cy+2)) / ry
                 if dx*dx+dy*dy <= 1.0 {
                     setPixel(img, x, y, dropShadow)
                 }
             }
         }
     ```
     `setPixel` overwrites opaque dirt pixels (`RGBA{151, 103, 81, 255}`) with `RGBA{0, 0, 0, 45}`.
   - In `cmd/tools/genassets/main.go:667`:
     `pebbles := [][2]int{{80, 40}, {180, 56}, {120, 88}, {60, 80}, {195, 36}, {145, 30}}`
     Pebble 5 is centered at `{195, 36}`. Center distance $d_{iso} = \frac{|195-127.5|}{128} + \frac{|36-63.5|}{64} = 0.9570$. With $r_x=7, r_y=4$, point $(202, 36)$ reaches $d_{iso}=1.0117$ and drop shadow reaches $(203, 36)$ with $d_{iso}=1.0195 > 1.0$.
   - Test execution `CC=gcc go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry` verbatim failure:
     ```
     --- FAIL: TestEmpiricalFloorDiamondGeometry/images/dirt.png (0.00s)
         empirical_challenger_test.go:235: images/dirt.png has 18 non-transparent pixels outside isometric diamond (isoDist > 1.0)
         empirical_challenger_test.go:242: images/dirt.png has 151 transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)
     ```

2. **Audit of All Other Floor Generators**:
   - `generateGrass`: 8 chevrons and 5 wildflowers. All detail coordinates stay strictly inside $d_{iso} \le 0.9180 < 1.0$. All colors are fully opaque ($A=255$). Test PASSES.
   - `generateWoodFloor`: Plank seams, end joints, and 8 nailheads. Nails guarded by $d_{iso} \le 0.88$. All colors are fully opaque ($A=255$). Test PASSES.
   - `generateAsphalt`: Yellow dashed markings strictly generated inside `if isoDist <= 1.0`. All colors opaque. Test PASSES.
   - `generateConcrete`: Slab quadrants, expansion joints, bevels generated inside `if isoDist <= 1.0`. All colors opaque. Test PASSES.
   - `generateTileFloor`: 4x4 checkered sub-tiles, grout lines, bevels generated inside `if isoDist <= 1.0`. All colors opaque. Test PASSES.

3. **Data Race in `internal/assets.Load()`**:
   - In `internal/assets/assets.go:53-88`: `Load()` assigns to global pointer variables `PlayerImage = loadEbitenImage(...)`, `GrassImage = ...`, etc., without synchronization.
   - Running `CC=gcc go test -v -race ./internal/assets/... -run TestChallenger_MultiThreadedLoadAndPointerRace` outputs data race warnings on writes at `internal/assets/assets.go:85-87`.

---

## 2. Logic Chain

1. **Step 1 (Alpha Holes)**: In `drawVectorPebble()`, `setPixel(img, x, y, dropShadow)` with `dropShadow.A = 45` replaces the solid floor pixels with semi-transparent pixels. Because the subsequent pebble body only covers pixels centered at `(cx, cy)`, unoccluded shadow pixels at `(cx+2, cy+2)` retain alpha 45. Replacing `setPixel` with `blendPixel` preserves the underlying solid alpha (Porter-Duff Over: $1.0 - 0.176(1.0) + \dots = 1.0$).
2. **Step 2 (Boundary Bleed)**: `drawVectorPebble()` does not verify whether destination coordinates satisfy $d_{iso}(x, y) \le 1.0$. Adding `if isoDist <= 1.0` and moving pebble `{195, 36}` inward to `{185, 42}` completely eliminates all 18 bleed pixels outside the 2:1 isometric diamond.
3. **Step 3 (Floor Generators Conformance)**: The other 5 floor generators (`grass`, `wood`, `asphalt`, `concrete`, `tile_floor`) either confine all pixel operations to `if isoDist <= 1.0` or strictly check bounds ($d_{iso} \le 0.88$), and only use opaque colors ($A=255$). Thus, no other floor generator has alpha or diamond bleed defects.
4. **Step 4 (Concurrency Safety)**: Wrapping the loading logic in `internal/assets.Load()` with `var loadOnce sync.Once; loadOnce.Do(func() { ... })` ensures that the 27 global image pointers are initialized exactly once safely and idempotently across multiple goroutines.

---

## 3. Caveats

- Investigation is read-only. Source changes in `cmd/tools/genassets/main.go` and `internal/assets/assets.go` must be applied by the implementation worker, followed by executing `go run ./cmd/tools/genassets` to regenerate the sprite PNGs.
- Semi-transparent colors in item icons (`water.png`, `antidote.png`) are deliberate glass highlight / bottle effects that pass all item contrast tests and do not affect floor tile geometry.

---

## 4. Conclusion

- The root causes for both Milestone 1 failure points are fully identified and verified with reproducible test cases.
- The remediation plan in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_2/fix_plan.md` provides exact before/after snippets for `cmd/tools/genassets/main.go` and `internal/assets/assets.go`.
- No other floor generators require fixes.

---

## 5. Verification Method

To independently verify the fixes after implementation and asset regeneration:

1. **Verify Asset Generation & Floor Tile Diamond Tests**:
   ```bash
   go run ./cmd/tools/genassets
   CC=gcc go test -v ./cmd/tools/genassets/...
   CC=gcc go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry
   ```
   *Expected*: `TestEmpiricalFloorDiamondGeometry/images/dirt.png` passes with 0 bleed pixels and 0 punctured alpha holes.

2. **Verify Concurrency & Race Detector**:
   ```bash
   CC=gcc go test -v -race ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace
   ```
   *Expected*: Passes with `PASS` and zero race warnings.

3. **Verify Full Project Test Suite**:
   ```bash
   CC=gcc go test ./...
   ```
   *Expected*: All packages pass without errors.
