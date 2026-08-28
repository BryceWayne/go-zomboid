# Milestone 1 Empirical Challenge Report: Asset Pipeline 4x Scaling

**Target Milestone**: Milestone 1 (Procedural Sprite 4x Scaling)
**Reviewer**: `m1_challenger_1` (Empirical Challenger: Critic / Specialist)
**Verdict**: **FAIL**
**Overall Risk Assessment**: **HIGH**

---

## 1. Executive Summary

Milestone 1 introduces 4x resolution sprites (256x128 floor tiles, 256x256 vertical obstacles/props, 64x128 character entities, and 64x64 items). An exhaustive empirical verification and stress testing suite (`internal/assets/empirical_challenger_test.go`) was authored and executed directly against all 27 generated assets and the generator binary `cmd/tools/genassets`.

While 26 of 27 assets meet dimension, grounding, and alpha fill specifications, empirical stress testing revealed **two critical geometric and alpha integrity defects** in the procedural generation of `images/dirt.png`:
1. **Alpha Perforation / Semi-Transparent Puncture**: `drawVectorPebble` in `cmd/tools/genassets/main.go:262` uses `setPixel(img, x, y, dropShadow)` with `dropShadow := color.RGBA{0, 0, 0, 45}` instead of `blendPixel`. This overwrites opaque dirt floor pixels with semi-transparent pixels, leaving **151 transparent holes (`Alpha = 45`)** inside the core diamond of the floor tile.
2. **Isometric Boundary Bleed**: In `generateDirt` (`cmd/tools/genassets/main.go:667`), pebble `{195, 36}` with radius $(r_x=7, r_y=4)$ is placed too close to the perimeter ($isoDist = 0.957$), causing **18 non-transparent pixels to spill across the boundary ($isoDist > 1.0$)** into the transparent outer corner quadrant.

Because floor tiles are rendered directly on screen in 2:1 isometric projection, alpha perforation causes background rendering bleed/flicker, and outer diamond bleeding causes visual seams when neighboring tiles are drawn.

---

## 2. Empirical Verification Matrix (All 27 Assets)

All 27 asset files were decoded, analyzed pixel-by-pixel, and tested for geometric conformance, fill density, and grounding:

| # | Asset File | Category | Target Dimensions | Actual Dimensions | Total Pixels | Non-Zero Alpha Pixels | Fill Ratio | Grounding / Bounds Status | Verdict |
|---|------------|----------|-------------------|-------------------|--------------|-----------------------|------------|---------------------------|---------|
| 1 | `images/player.png` | Entity | 64x128 | 64x128 | 8,192 | 3,118 | 38.06% | 559 ground px (rows 112..127) | **PASS** |
| 2 | `images/zombie.png` | Entity | 64x128 | 64x128 | 8,192 | 3,348 | 40.87% | 525 ground px (rows 112..127) | **PASS** |
| 3 | `images/runner.png` | Entity | 64x128 | 64x128 | 8,192 | 3,246 | 39.62% | 615 ground px (rows 112..127) | **PASS** |
| 4 | `images/grass.png` | Floor | 256x128 | 256x128 | 32,768 | 16,384 | 50.00% | 0 bleed px, 100% solid core | **PASS** |
| 5 | `images/dirt.png` | Floor | 256x128 | 256x128 | 32,768 | 16,396 | 50.04% | **18 bleed px, 151 punctured alpha=45 px** | **FAIL** |
| 6 | `images/wood.png` | Floor | 256x128 | 256x128 | 32,768 | 16,384 | 50.00% | 0 bleed px, 100% solid core | **PASS** |
| 7 | `images/asphalt.png` | Floor | 256x128 | 256x128 | 32,768 | 16,384 | 50.00% | 0 bleed px, 100% solid core | **PASS** |
| 8 | `images/concrete.png` | Floor | 256x128 | 256x128 | 32,768 | 16,384 | 50.00% | 0 bleed px, 100% solid core | **PASS** |
| 9 | `images/tile_floor.png` | Floor | 256x128 | 256x128 | 32,768 | 16,384 | 50.00% | 0 bleed px, 100% solid core | **PASS** |
| 10 | `images/wall.png` | Obstacle | 256x256 | 256x256 | 65,536 | 32,768 | 50.00% | Anchored at lower base | **PASS** |
| 11 | `images/tree.png` | Obstacle | 256x256 | 256x256 | 65,536 | 27,872 | 42.53% | Anchored at trunk base | **PASS** |
| 12 | `images/fence.png` | Obstacle | 256x256 | 256x256 | 65,536 | 13,056 | 19.92% | Grounded posts | **PASS** |
| 13 | `images/debris.png` | Obstacle | 256x256 | 256x256 | 65,536 | 10,752 | 16.40% | Grounded rubble | **PASS** |
| 14 | `images/tent.png` | Obstacle | 256x256 | 256x256 | 65,536 | 18,432 | 28.12% | Grounded skirt | **PASS** |
| 15 | `images/stump.png` | Obstacle | 256x256 | 256x256 | 65,536 | 10,240 | 15.62% | Grounded base | **PASS** |
| 16 | `images/mushroom.png` | Obstacle | 256x256 | 256x256 | 65,536 | 12,800 | 19.53% | Grounded stem | **PASS** |
| 17 | `images/sign.png` | Obstacle | 256x256 | 256x256 | 65,536 | 7,680 | 11.72% | Grounded post | **PASS** |
| 18 | `images/elevation_block.png` | Obstacle | 256x256 | 256x256 | 65,536 | 32,768 | 50.00% | Anchored base | **PASS** |
| 19 | `images/elevation_ramp.png` | Obstacle | 256x256 | 256x256 | 65,536 | 24,576 | 37.50% | Anchored slope | **PASS** |
| 20 | `images/food.png` | Item | 64x64 | 64x64 | 4,096 | 1,024 | 25.00% | Centered icon (32.0, 32.0) | **PASS** |
| 21 | `images/water.png` | Item | 64x64 | 64x64 | 4,096 | 1,088 | 26.56% | Centered icon (32.0, 32.0) | **PASS** |
| 22 | `images/weapon.png` | Item | 64x64 | 64x64 | 4,096 | 896 | 21.88% | Centered icon (32.0, 32.0) | **PASS** |
| 23 | `images/axe.png` | Item | 64x64 | 64x64 | 4,096 | 1,152 | 28.12% | Centered icon (32.0, 32.0) | **PASS** |
| 24 | `images/shotgun.png` | Item | 64x64 | 64x64 | 4,096 | 960 | 23.44% | Centered icon (32.0, 32.0) | **PASS** |
| 25 | `images/ammo.png` | Item | 64x64 | 64x64 | 4,096 | 1,024 | 25.00% | Centered icon (32.0, 32.0) | **PASS** |
| 26 | `images/armor.png` | Item | 64x64 | 64x64 | 4,096 | 1,280 | 31.25% | Centered icon (32.0, 32.0) | **PASS** |
| 27 | `images/antidote.png` | Item | 64x64 | 64x64 | 4,096 | 1,088 | 26.56% | Centered icon (32.0, 32.0) | **PASS** |

---

## 3. Empirical Stress Test Findings & Challenges

### [HIGH] Challenge 1: `images/dirt.png` Alpha Perforation via `setPixel` Drop Shadow Overwrite
- **Assumption Challenged**: All floor tiles are 100% opaque ($\alpha = 255$) across their entire inner 2:1 isometric diamond core ($\frac{|x - 127.5|}{128} + \frac{|y - 63.5|}{64} \le 0.85$).
- **Attack Scenario**: Evaluated pixel alpha values across the inner solid core of `images/dirt.png`.
- **Empirical Observation**: 151 pixels in `dirt.png` have $\alpha = 45$ ($17.6\%$ opacity) inside the core diamond.
  - Verbatim log:
    ```
    Inner hole at (153, 30): RGBA=(0, 0, 0, 45) [isoDist=0.723]
    Inner hole at (152, 31): RGBA=(0, 0, 0, 45) [isoDist=0.699]
    Inner hole at (153, 31): RGBA=(0, 0, 0, 45) [isoDist=0.707]
    Inner hole at (152, 32): RGBA=(0, 0, 0, 45) [isoDist=0.684]
    Inner hole at (153, 32): RGBA=(0, 0, 0, 45) [isoDist=0.691]
    ```
- **Root Cause**: In `cmd/tools/genassets/main.go:250-265`:
  ```go
  func drawVectorPebble(img *image.RGBA, cx, cy int, rx, ry float64, base, light, shadow color.RGBA) {
      dropShadow := color.RGBA{0, 0, 0, 45}
      ...
      for y := minY; y <= maxY; y++ {
          for x := minX; x <= maxX; x++ {
              ...
              if dx*dx+dy*dy <= 1.0 {
                  setPixel(img, x, y, dropShadow) // <-- Overwrites base opaque pixel!
              }
          }
      }
  ```
  `setPixel` overwrites the underlying opaque dirt color `(151, 103, 81, 255)` with `(0, 0, 0, 45)`. Because the pebble body is shifted relative to the shadow, unoccluded shadow pixels leave translucent holes in the ground tile.
- **Blast Radius**: In the Ebitengine render pass, any object or black background behind the dirt floor tile shows through these 151 translucent holes, creating visual artifacting and flicker.
- **Recommended Mitigation**: Replace `setPixel(img, x, y, dropShadow)` with `blendPixel(img, x, y, dropShadow)` in `drawVectorPebble` so alpha blending preserves the 100% opaque alpha of the floor tile.

---

### [MEDIUM] Challenge 2: `images/dirt.png` Isometric Diamond Boundary Bleed
- **Assumption Challenged**: Floor tile sprites strictly conform to the 2:1 isometric diamond boundary ($\frac{|x - 127.5|}{128} + \frac{|y - 63.5|}{64} \le 1.0$), with all outer corner pixels having $\alpha = 0$.
- **Attack Scenario**: Evaluated pixel alpha values for all coordinates with $isoDist > 1.0$.
- **Empirical Observation**: `images/dirt.png` has **18 non-transparent pixels** with $isoDist > 1.0$.
- **Root Cause**: In `generateDirt` (`cmd/tools/genassets/main.go:667`), pebble 5 is placed at `{195, 36}` with radius $r_x=7, r_y=4$. The center is at $isoDist = 0.957$, and its right edge extends to $x=202, y=36$ where $isoDist = 1.0117$. Unlike `generateWoodFloor` (which checks $isoDist \le 0.88$ before placing nails) or `generateGrass` (which keeps flowers/chevrons well inside the diamond), `drawVectorPebble` lacks diamond boundary clipping.
- **Blast Radius**: Spills tile pixels into adjacent tile space during isometric grid rendering, creating edge seams and visual overlap artifacts.
- **Recommended Mitigation**: Either reposition the pebble at `{195, 36}` inward (e.g. `{185, 42}`), or add an `isoDist <= 0.92` bounds check in `drawVectorPebble` or `generateDirt`.

---

## 4. Test Suite Execution Results

Executed command: `CC=gcc go test -p 1 -v -count=1 ./internal/assets/... ./cmd/tools/genassets/...`

```
=== RUN   TestEmbeddedAssetDimensionsAndValidity (internal/assets)
--- PASS: TestEmbeddedAssetDimensionsAndValidity (0.01s)
=== RUN   TestAssetsLoadAllPointersNonNil (internal/assets)
--- PASS: TestAssetsLoadAllPointersNonNil (0.01s)
=== RUN   TestFloorTileIsometricBounds (internal/assets)
--- PASS: TestFloorTileIsometricBounds (0.00s) [Note: Loose tolerance dist > 1.15 masked pebble bleed]
=== RUN   TestCharacterGroundAnchor (internal/assets)
--- PASS: TestCharacterGroundAnchor (0.00s)
=== RUN   TestItemOutlineContrast (internal/assets)
--- PASS: TestItemOutlineContrast (0.00s)
=== RUN   TestAssetsLoadIdempotency (internal/assets)
--- PASS: TestAssetsLoadIdempotency (0.01s)
=== RUN   TestEmpiricalAssetCatalogCompleteness (internal/assets)
--- PASS: TestEmpiricalAssetCatalogCompleteness (0.01s)
=== RUN   TestEmpiricalAlphaFillRatios (internal/assets)
--- PASS: TestEmpiricalAlphaFillRatios (0.01s)
=== RUN   TestEmpiricalFloorDiamondGeometry (internal/assets)
    --- PASS: TestEmpiricalFloorDiamondGeometry/images/grass.png (0.00s)
    --- FAIL: TestEmpiricalFloorDiamondGeometry/images/dirt.png (0.00s)
        empirical_challenger_test.go:235: images/dirt.png has 18 non-transparent pixels outside isometric diamond (isoDist > 1.0)
        empirical_challenger_test.go:242: images/dirt.png has 151 transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)
    --- PASS: TestEmpiricalFloorDiamondGeometry/images/wood.png (0.00s)
    --- PASS: TestEmpiricalFloorDiamondGeometry/images/asphalt.png (0.00s)
    --- PASS: TestEmpiricalFloorDiamondGeometry/images/concrete.png (0.00s)
    --- PASS: TestEmpiricalFloorDiamondGeometry/images/tile_floor.png (0.00s)
--- FAIL: TestEmpiricalFloorDiamondGeometry (0.00s)
=== RUN   TestEmpiricalCharacterGrounding (internal/assets)
--- PASS: TestEmpiricalCharacterGrounding (0.00s)
=== RUN   TestEmpiricalGenerationDeterminism (internal/assets)
--- PASS: TestEmpiricalGenerationDeterminism (0.25s)
=== RUN   TestEmpiricalObstacleBoundsAndGrounding (internal/assets)
--- PASS: TestEmpiricalObstacleBoundsAndGrounding (0.01s)
=== RUN   TestEmpiricalItemIconQuality (internal/assets)
--- PASS: TestEmpiricalItemIconQuality (0.00s)
=== RUN   TestAssetRegenerationDeterminism (cmd/tools/genassets)
--- PASS: TestAssetRegenerationDeterminism (0.32s)
=== RUN   TestAssetDimensionsAndIntegrity (cmd/tools/genassets)
--- PASS: TestAssetDimensionsAndIntegrity (0.02s)
```

---

## 5. Unchallenged Areas
- **Ebitengine In-Game Rendering of Bezier curves (Milestone 3)**: Out of scope for Milestone 1 asset pipeline verification.
- **Audio file embedding (`internal/assets/audio.go`)**: Milestone 1 scope is restricted to 4x scaling of 27 sprite assets.

---

## 6. Conclusion & Recommendation

The asset scaling architecture is solid across 26 of 27 assets. However, due to the two verified empirical defects in `images/dirt.png` (151 punctured semi-transparent pixels and 18 outer-diamond bleed pixels caused by `drawVectorPebble`), Milestone 1 is **FAILED** until these generator defects are addressed by the implementation agent and regenerated.
