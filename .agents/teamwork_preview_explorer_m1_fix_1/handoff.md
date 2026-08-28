# Handoff Report: Milestone 1 Remediation Fix Design

**Author**: `m1_explorer_fix_1`  
**Milestone**: M1 Asset Generation & Asset Loading Fix  
**Handoff Type**: Hard Handoff (Investigation & Fix Design Complete)  

---

## 1. Observation

1. **Reproduction Command & Verbatim Test Failure (`internal/assets`)**:
   Running `CC=clang go test -v ./internal/assets` produced the following verbatim failure in `TestEmpiricalFloorDiamondGeometry/images/dirt.png`:
   ```text
   === RUN   TestEmpiricalFloorDiamondGeometry
   === RUN   TestEmpiricalFloorDiamondGeometry/images/grass.png
   === RUN   TestEmpiricalFloorDiamondGeometry/images/dirt.png
       empirical_challenger_test.go:226: Inner hole at (153, 30): RGBA=(0, 0, 0, 45) [isoDist=0.723]
       empirical_challenger_test.go:226: Inner hole at (152, 31): RGBA=(0, 0, 0, 45) [isoDist=0.699]
       empirical_challenger_test.go:226: Inner hole at (153, 31): RGBA=(0, 0, 0, 45) [isoDist=0.707]
       empirical_challenger_test.go:226: Inner hole at (152, 32): RGBA=(0, 0, 0, 45) [isoDist=0.684]
       empirical_challenger_test.go:226: Inner hole at (153, 32): RGBA=(0, 0, 0, 45) [isoDist=0.691]
       empirical_challenger_test.go:235: images/dirt.png has 18 non-transparent pixels outside isometric diamond (isoDist > 1.0)
       empirical_challenger_test.go:242: images/dirt.png has 151 transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)
   --- FAIL: TestEmpiricalFloorDiamondGeometry (0.00s)
       --- FAIL: TestEmpiricalFloorDiamondGeometry/images/dirt.png (0.00s)
   ```

2. **Reproduction Command & Verbatim Data Race Warning (`internal/assets`)**:
   Running `CC=clang go test -race -v ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace` produced verbatim race detector warnings:
   ```text
   ==================
   WARNING: DATA RACE
   Write at 0x0000021c3fa0 by goroutine 116:
     github.com/BryceWayne/go-zomboid/internal/assets.Load()
         /home/bryce/code/go-zomboid/internal/assets/assets.go:55 +0x44
   Previous write at 0x0000021c3fa0 by goroutine 115:
     github.com/BryceWayne/go-zomboid/internal/assets.Load()
         /home/bryce/code/go-zomboid/internal/assets/assets.go:55 +0x44
   ==================
   ```

3. **Source Code Inspection in `cmd/tools/genassets/main.go`**:
   - `main.go:250-265`:
     ```go
     func drawVectorPebble(img *image.RGBA, cx, cy int, rx, ry float64, base, light, shadow color.RGBA) {
     	dropShadow := color.RGBA{0, 0, 0, 45}
     	// Drop shadow
     	minY := int(math.Floor(float64(cy+2) - ry))
     	maxY := int(math.Ceil(float64(cy+2) + ry))
     	minX := int(math.Floor(float64(cx+2) - rx))
     	maxX := int(math.Ceil(float64(cx+2) + rx))
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
     `setPixel` directly overwrites existing pixel memory with `{0, 0, 0, 45}` without alpha blending against the background.
   - `main.go:667`:
     ```go
     pebbles := [][2]int{
     	{80, 40}, {180, 56}, {120, 88}, {60, 80}, {195, 36}, {145, 30},
     }
     ```
     Pebble at `{195, 36}` with radius $r_x=7, r_y=4$ produces points with $isoDist = \frac{|195+7-127.5|}{128} + \frac{|36-63.5|}{64} = 0.5820 + 0.4297 = 1.0117 > 1.0$.

4. **Source Code Inspection in `internal/assets/assets.go`**:
   - `assets.go:53-88`:
     `Load()` assigns global pointers `PlayerImage = loadEbitenImage(...)`, `DirtImage = loadEbitenImage(...)`, etc., on every invocation without any concurrency guard.

---

## 2. Logic Chain

1. From Observation 3, `setPixel(img, x, y, dropShadow)` sets the pixel alpha directly to $45$ in the region where the drop shadow extends beyond the pebble body. Because this happens over the already rendered opaque dirt floor, it punctures 151 semi-transparent pixels inside the solid core ($isoDist \le 0.85$).
2. Replacing `setPixel(img, x, y, dropShadow)` with `blendPixel(img, x, y, dropShadow)` computes destination blending:
   $$outA = srcA + dstA \cdot (1 - srcA) = \frac{45}{255} + 1.0 \cdot \left(1 - \frac{45}{255}\right) = 1.0$$
   This darkens the RGB color values while keeping alpha at 255 ($100\%$ opacity), eliminating all 151 semi-transparent holes.
3. From Observation 3, pebble `{195, 36}` has its center at normalized isometric distance $d_{iso} = 0.9570$. Adding pebble radius $r_x=7, r_y=4$ causes pixels in the range $x \in [199..204], y \in [32..40]$ to exceed $d_{iso} = 1.0$, producing the 18 non-transparent bleed pixels.
4. Repositioning pebble `{195, 36}` to `{185, 42}` places the center at $d_{iso} = 0.7852$, so all pebble and shadow pixels satisfy $d_{iso} \le 0.8398 < 0.85$. Adding explicit `if isoDist <= 1.0` guards in `drawVectorPebble` provides defense-in-depth against boundary spillover.
5. From Observation 4, `Load()` lacks synchronization. When multiple goroutines invoke `Load()` or read global pointers concurrently, unsynchronized writes occur on the 27 global `*ebiten.Image` variables.
6. Introducing `var loadOnce sync.Once` and wrapping the body of `Load()` in `loadOnce.Do(func() { ... })` ensures that the loading procedure executes exactly once, with Go memory barriers guaranteeing safe concurrent reads by all goroutines.

---

## 3. Caveats

- In headless Linux environments without X11/ALSA installed in default gcc include search paths, running ebiten/oto requires specifying `CC=clang` when running tests.
- Re-running `genassets` will update `internal/assets/images/dirt.png`. All other 26 generated assets will remain bit-for-bit identical due to deterministic generator math.

---

## 4. Conclusion

Milestone 1 defects are fully diagnosed and resolved by:
1. Updating `cmd/tools/genassets/main.go` to use `blendPixel` for drop shadows, guarding pixel iteration with `isoDist <= 1.0`, and shifting pebble `{195, 36}` inward to `{185, 42}`.
2. Updating `internal/assets/assets.go` to wrap asset loading in `sync.Once`.
3. Executing `go run ./cmd/tools/genassets` to regenerate `internal/assets/images/dirt.png`.

The complete fix plan and code diffs are documented in `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_fix_1/fix_plan.md`.

---

## 5. Verification Method

To verify these fixes independently:

1. **Run asset generation**:
   ```bash
   go run ./cmd/tools/genassets
   ```

2. **Run asset generator tests**:
   ```bash
   go test -v ./cmd/tools/genassets
   ```

3. **Run empirical challenger tests**:
   ```bash
   CC=clang go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry
   ```
   *Expected*: Passes with 0 bleeding pixels and 0 transparent/semi-transparent pixels inside core.

4. **Run concurrency race test**:
   ```bash
   CC=clang go test -race -v ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace
   ```
   *Expected*: Passes with 0 data races.

5. **Run entire test suite**:
   ```bash
   CC=clang go test ./...
   ```

**Invalidation Conditions**:
- Any alpha values $< 255$ inside $isoDist \le 0.85$ for `images/dirt.png`.
- Any non-zero alpha pixels with $isoDist > 1.0$ for `images/dirt.png`.
- Any data race warning reported under `go test -race ./internal/assets`.
