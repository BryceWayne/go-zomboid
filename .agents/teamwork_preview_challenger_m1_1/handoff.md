# Handoff Report — Milestone 1 (Asset Pipeline 4x Scaling Empirical Challenge)

**Reviewer**: `m1_challenger_1` (Empirical Challenger: Critic / Specialist)
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_challenger_m1_1`
**Verdict**: **FAIL**

---

## 1. Observation

1. **Test Execution Command & Failure Output**:
   Command run: `CC=gcc go test -p 1 -v -count=1 ./internal/assets/... ./cmd/tools/genassets/...`
   Verbatim output from `TestEmpiricalFloorDiamondGeometry`:
   ```
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
       --- PASS: TestEmpiricalFloorDiamondGeometry/images/grass.png (0.00s)
       --- FAIL: TestEmpiricalFloorDiamondGeometry/images/dirt.png (0.00s)
       --- PASS: TestEmpiricalFloorDiamondGeometry/images/wood.png (0.00s)
       --- PASS: TestEmpiricalFloorDiamondGeometry/images/asphalt.png (0.00s)
       --- PASS: TestEmpiricalFloorDiamondGeometry/images/concrete.png (0.00s)
       --- PASS: TestEmpiricalFloorDiamondGeometry/images/tile_floor.png (0.00s)
   FAIL
   FAIL	github.com/BryceWayne/go-zomboid/internal/assets	0.032s
   ```

2. **Source Code Implementation in `cmd/tools/genassets/main.go`**:
   - Lines 250–265:
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
   - Lines 666–671:
     ```go
     // 4x Scaled rounded vector pebbles (~14x8px)
     pebbles := [][2]int{
         {80, 40}, {180, 56}, {120, 88}, {60, 80}, {195, 36}, {145, 30},
     }
     for _, pos := range pebbles {
         drawVectorPebble(img, pos[0], pos[1], 7.0, 4.0, pebbleBase, pebbleLight, pebbleShadow)
     }
     ```

3. **Asset Dimensions and Determinism Measurements**:
   - All 27 assets in `internal/assets/images/*.png` match required dimensions:
     - 6 Floors: 256x128
     - 10 Obstacles/Props: 256x256
     - 3 Entities: 64x128
     - 8 Items: 64x64
   - Character entities possess solid grounding pixels in rows 112..127 (`player.png`: 559 px / 272 solid; `zombie.png`: 525 px / 238 solid; `runner.png`: 615 px / 276 solid).
   - Asset regeneration is 100% deterministic (identical SHA-256 hashes across consecutive runs of `cmd/tools/genassets`).

---

## 2. Logic Chain

1. From Observation 2 (lines 250–265), `drawVectorPebble` defines `dropShadow` as `color.RGBA{0, 0, 0, 45}` and calls `setPixel(img, x, y, dropShadow)`.
2. `setPixel` performs a direct unblended assignment `img.SetRGBA(x, y, c)`. When applied to the dirt tile base, it replaces the fully opaque ground color `(151, 103, 81, 255)` with `(0, 0, 0, 45)`.
3. Because the pebble body (`normDist <= 1.0` centered at `(cx, cy)`) is offset by `(-2, -2)` from the shadow (`(cx+2, cy+2)`), unoccluded shadow pixels are not painted over by the opaque pebble body.
4. From Observation 1, this directly results in **151 pixels with alpha = 45** inside the solid diamond core of `images/dirt.png`, punching translucent holes in the ground sprite.
5. From Observation 2 (lines 666–671), pebble 5 is placed at `cx = 195, cy = 36` with radii $r_x = 7.0, r_y = 4.0$. At $x = 202, y = 36$, the normalized distance is $\frac{|202 - 127.5|}{128} + \frac{|36 - 63.5|}{64} = 0.5820 + 0.4297 = 1.0117 > 1.0$.
6. Because `drawVectorPebble` lacks diamond boundary clipping, this directly causes **18 non-transparent pixels to bleed outside the 2:1 isometric diamond boundary** into the transparent upper-right corner.
7. Consequently, `images/dirt.png` fails floor tile geometric and alpha integrity contracts, leading to verdict **FAIL**.

---

## 3. Caveats

- Milestone 1 verification focuses strictly on asset generation (`cmd/tools/genassets`), embedded assets (`internal/assets`), and image mathematical geometry. Subsequent milestones (M2 engine coordinate math and M3 Bezier combat curves) were not evaluated in this handoff.
- The existing test `TestFloorTileIsometricBounds` in `internal/assets/assets_stress_test.go` did not catch the pebble bleed because its tolerance was set too loosely (`dist > 1.15`).

---

## 4. Conclusion

Milestone 1 is **FAILED** due to two reproducible defects in `cmd/tools/genassets/main.go` affecting `images/dirt.png`:
1. `drawVectorPebble` calling `setPixel` with semi-transparent `dropShadow` instead of `blendPixel`, creating 151 alpha=45 holes.
2. Pebble `{195, 36}` causing 18 pixels to bleed across the isometric diamond boundary ($isoDist > 1.0$).

**Action Required from Implementer**:
1. In `cmd/tools/genassets/main.go:262`, change `setPixel(img, x, y, dropShadow)` to `blendPixel(img, x, y, dropShadow)`.
2. In `cmd/tools/genassets/main.go:667`, adjust the pebble coordinate `{195, 36}` inward (e.g. `{185, 42}`) or enforce $isoDist \le 0.92$ clipping.
3. Re-run `go run ./cmd/tools/genassets` to regenerate the sprite files.

---

## 5. Verification Method

To independently reproduce the failure and verify fixes:
```bash
# 1. Run the empirical challenger test suite
CC=gcc go test -p 1 -v -count=1 -run TestEmpiricalFloorDiamondGeometry ./internal/assets/...

# 2. Run the full project test suite
CC=gcc go test -p 1 ./...
```
Expected invalidation condition for this FAIL verdict: When the above commands exit with code 0 and `TestEmpiricalFloorDiamondGeometry/images/dirt.png` passes with 0 bleeding pixels and 0 punctured alpha pixels.
