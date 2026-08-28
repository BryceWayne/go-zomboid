# Milestone 1 Empirical Challenge Report: Asset Generation Pipeline & Image Validity

**Verdict**: **FAIL** (2 Bugs Identified: 1 High Severity, 1 Medium Severity)  
**Agent**: `m1_challenger_2`  
**Date**: 2026-08-28T18:59:30Z  
**Target Scope**: `cmd/tools/genassets`, `internal/assets`

---

## Challenge Summary

**Overall risk assessment**: **HIGH**

Empirical stress testing was conducted against the Milestone 1 asset generation tool and runtime asset loading package. While all 27 exported image pointers match their required bounds and exhibit strong contrast and vector styling, stress testing revealed:
1. **High Severity Bug**: Alpha corruption (151 semi-transparent holes) and isometric diamond boundary bleed (18 pixels) in `images/dirt.png` caused by `drawVectorPebble()` overwriting opaque pixels with `setPixel()` instead of `blendPixel()`.
2. **Medium Severity Bug**: Data race in `internal/assets.Load()` during concurrent multi-threaded execution detected by `go test -race`.

---

## Challenges & Empirical Findings

### [High] Challenge 1: Alpha Hole Corruption & Boundary Bleed in `images/dirt.png`

- **Assumption challenged**: Floor tiles (256x128) must have a 100% solid core within the 2:1 isometric diamond ($d_{iso} \le 0.85$) and zero non-transparent pixels outside the diamond ($d_{iso} > 1.0$).
- **Attack scenario**: Evaluated pixel alpha channels and diamond distance metric $d_{iso} = \frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0}$ on all floor tiles.
- **Root cause**: In `cmd/tools/genassets/main.go:251-265`, `drawVectorPebble()` uses `setPixel(img, x, y, dropShadow)` where `dropShadow := color.RGBA{0, 0, 0, 45}` without bounding checks to $d_{iso} \le 1.0$:
  ```go
  // In cmd/tools/genassets/main.go:262
  if dx*dx+dy*dy <= 1.0 {
      setPixel(img, x, y, dropShadow) // <-- OVERWRITES opaque background with alpha 45!
  }
  ```
- **Empirical evidence**:
  - `images/dirt.png` has **151 transparent/semi-transparent pixels** inside the solid core ($d_{iso} \le 0.85$), e.g. at $(153, 30)$ with `RGBA=(0, 0, 0, 45)`.
  - `images/dirt.png` has **18 non-transparent pixels** bleeding outside the diamond ($d_{iso} > 1.0$).
  - Reproduction: `CC=gcc go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry/images/dirt.png` fails.
- **Blast radius**: When rendered in-game on the isometric grid, underlying layers or black background will show through the dirt tiles, and boundary bleed will create visible tile seams.
- **Mitigation**:
  1. In `drawVectorPebble()`, check that $(x, y)$ satisfies diamond bounds $d_{iso} \le 1.0$.
  2. Use `blendPixel(img, x, y, dropShadow)` instead of `setPixel` so the shadow blends over the earthen background without lowering the alpha channel.

---

### [Medium] Challenge 2: Data Race on `internal/assets.Load()` under Concurrent Invocations

- **Assumption challenged**: `internal/assets.Load()` is thread-safe and idempotent under multi-threaded initialization or concurrent access.
- **Attack scenario**: Spawned 20 loader goroutines calling `Load()` and 30 reader goroutines querying `Bounds()` on all 27 pointers under Go's race detector (`-race`).
- **Root cause**: In `internal/assets/assets.go:53-88`, `Load()` performs direct unsynchronized writes to global package variables (`PlayerImage = loadEbitenImage(...)`, `GrassImage = ...`, etc.).
- **Empirical evidence**:
  - `CC=gcc go test -v -race ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace` output:
    ```
    WARNING: DATA RACE
    Write at 0x000000eb92a8 by goroutine 40:
      github.com/BryceWayne/go-zomboid/internal/assets.Load()
          internal/assets/assets.go:86 +0x84a
    Previous write at 0x000000eb92a8 by goroutine 42:
      github.com/BryceWayne/go-zomboid/internal/assets.Load()
          internal/assets/assets.go:86 +0x84a
    ```
- **Blast radius**: Concurrent background asset loading or parallel test runners cause data races and may observe partially initialized / nil pointer states.
- **Mitigation**: Wrap initialization logic in `internal/assets.Load()` with `sync.Once`:
  ```go
  var loadOnce sync.Once
  func Load() {
      loadOnce.Do(func() {
          PlayerImage = loadEbitenImage("images/player.png")
          // ...
      })
  }
  ```

---

## Stress Test Results

| Test Case | Target | Expected Behavior | Actual Behavior | Result |
|-----------|--------|-------------------|-----------------|--------|
| `TestChallenger_All27ExportedPointersAndExactBounds` | 27 Image Pointers | All 27 pointers non-nil with exact dimensions | All 27 pointers valid: 3 Entities (64x128), 6 Floors (256x128), 10 Props (256x256), 8 Items (64x64) | **PASS** |
| `TestChallenger_AssetPixelContrastAndColorSaturation` | All 27 PNG Assets | RMS Contrast > 10.0, Dyn Range $\ge$ 40.0, Mean Sat > 0.2 | RMS Contrast 25.26–75.31, Dyn Range 92.26–250.00, Max Sat 0.30–0.91 | **PASS** |
| `TestChallenger_CharacterGroundDropShadows` | `player`, `zombie`, `runner` | Drop shadows at $y \in [112..127]$, head at $y \le 40$ | Ground shadow pixels > 500, head pixels > 100 | **PASS** |
| `TestChallenger_ItemIconCenteringAndContour` | 8 Item Icons | Centroid within $[20..44]$, dark contour outlines | Centroids centered, contour pixels present | **PASS** |
| `TestAssetRegenerationDeterminism` | `genassets` tool | 100% byte-for-byte SHA256 hash stability across runs | All 27 hashes identical over 3 iterations | **PASS** |
| `TestEmpiricalFloorDiamondGeometry` | Floor tiles (6) | $d_{iso} \le 1.0$, core $d_{iso} \le 0.85$ solid ($a = 255$) | `dirt.png` has 151 alpha<255 holes and 18 bleed pixels | **FAIL** |
| `TestChallenger_MultiThreadedLoadAndPointerRace` | `assets.Load()` | Race-free concurrent loading and reading | Data race detected on global pointer assignments | **FAIL** |

---

## Statistical Asset Metrics Summary

| Asset | Category | Dimensions | Fill Ratio | Mean Lum | RMS Contrast | Dynamic Range | Max Sat |
|---|---|---|---|---|---|---|---|
| `PlayerImage` | Entity | 64x128 | 21.64% | 118.89 | 70.08 | 244.60 | 0.88 |
| `ZombieImage` | Entity | 64x128 | 21.14% | 138.83 | 50.94 | 234.33 | 0.87 |
| `RunnerImage` | Entity | 64x128 | 21.90% | 116.71 | 75.31 | 249.52 | 0.91 |
| `GrassImage` | Floor | 256x128 | 50.00% | 139.11 | 50.77 | 180.59 | 0.76 |
| `DirtImage` | Floor | 256x128 | 50.00% | 126.96 | 40.54 | 148.96 | 0.51 |
| `WoodImage` | Floor | 256x128 | 50.00% | 91.17 | 25.70 | 93.90 | 0.71 |
| `AsphaltImage` | Floor | 256x128 | 50.00% | 57.49 | 33.82 | 161.61 | 0.83 |
| `ConcreteImage` | Floor | 256x128 | 50.00% | 147.57 | 29.30 | 149.43 | 0.03 |
| `TileFloorImage` | Floor | 256x128 | 50.00% | 128.62 | 72.23 | 217.27 | 0.24 |
| `WallImage` | Prop | 256x256 | 72.26% | 123.06 | 69.67 | 204.72 | 0.77 |
| `TreeImage` | Prop | 256x256 | 44.32% | 126.80 | 50.38 | 178.87 | 0.71 |
| `FenceImage` | Prop | 256x256 | 17.31% | 119.20 | 34.42 | 150.21 | 0.36 |
| `DebrisImage` | Prop | 256x256 | 27.89% | 83.96 | 37.75 | 242.05 | 0.71 |
| `TentImage` | Prop | 256x256 | 34.99% | 55.91 | 47.11 | 204.07 | 0.82 |
| `StumpImage` | Prop | 256x256 | 18.22% | 80.49 | 54.28 | 164.25 | 0.67 |
| `MushroomImage` | Prop | 256x256 | 26.08% | 107.71 | 75.58 | 250.00 | 0.91 |
| `SignImage` | Prop | 256x256 | 17.71% | 86.43 | 65.06 | 191.71 | 0.86 |
| `ElevationBlockImage` | Prop | 256x256 | 75.29% | 106.81 | 34.35 | 136.09 | 0.62 |
| `ElevationRampImage` | Prop | 256x256 | 56.54% | 130.10 | 43.10 | 92.26 | 0.59 |
| `FoodImage` | Item | 64x64 | 46.00% | 139.47 | 71.64 | 206.74 | 0.89 |
| `WaterImage` | Item | 64x64 | 38.26% | 136.21 | 69.96 | 207.81 | 0.90 |
| `WeaponImage` | Item | 64x64 | 20.46% | 123.84 | 64.75 | 217.47 | 0.87 |
| `AxeImage` | Item | 64x64 | 20.34% | 112.78 | 69.44 | 221.50 | 0.89 |
| `ShotgunImage` | Item | 64x64 | 14.62% | 66.57 | 46.58 | 183.94 | 0.77 |
| `AmmoImage` | Item | 64x64 | 55.08% | 95.20 | 55.10 | 205.33 | 0.84 |
| `ArmorImage` | Item | 64x64 | 48.44% | 51.44 | 25.26 | 133.75 | 0.30 |
| `AntidoteImage` | Item | 64x64 | 36.11% | 131.44 | 66.29 | 227.29 | 0.87 |

---

## Recommendation

Block Milestone 1 sign-off until the two empirical bugs are remediated:
1. Fix `drawVectorPebble()` in `cmd/tools/genassets/main.go` to use `blendPixel()` and enforce $d_{iso} \le 1.0$.
2. Protect `internal/assets.Load()` with `sync.Once`.
