# M1 Asset Pipeline & Race Testing Fix Plan

## 1. Executive Summary & Root Cause Analysis

During execution of the asset generator and test suite under `-race` (`CC=gcc go test -race -v ./cmd/tools/genassets/... ./internal/assets/...`), two distinct failures were identified:

### Root Cause 1: Isometric Floor Geometry Violation in `images/dirt.png`
- **Location**: `cmd/tools/genassets/main.go` (functions `drawVectorPebble` and `generateDirt`).
- **Symptoms**:
  - `TestEmpiricalFloorDiamondGeometry/images/dirt.png` failed with:
    ```
    empirical_challenger_test.go:226: Inner hole at (153, 30): RGBA=(0, 0, 0, 45) [isoDist=0.723]
    empirical_challenger_test.go:235: images/dirt.png has 18 non-transparent pixels outside isometric diamond (isoDist > 1.0)
    empirical_challenger_test.go:242: images/dirt.png has 151 transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)
    ```
- **Mechanism**:
  1. `drawVectorPebble` creates a drop shadow with `dropShadow := color.RGBA{0, 0, 0, 45}` and applies it using `setPixel(img, x, y, dropShadow)` instead of `blendPixel`. This overwrites the opaque dirt pixels ($A = 255$) with semi-transparent pixels ($A = 45$), creating 151 non-opaque "holes" within the inner solid core ($\text{isoDist} \le 0.85$).
  2. Pebble position `{195, 36}` in `generateDirt` is centered at $\text{isoDist} = \frac{|195 - 127.5|}{128} + \frac{|36 - 63.5|}{64} = 0.957$. With horizontal radius $r_x = 7.0$, the right edge reaches $x = 202$ ($\text{isoDist} = 1.012 > 1.0$), bleeding 18 pixels outside the isometric diamond.
  3. `drawVectorPebble` lacked an isometric boundary check (`isoDist <= 1.0`) during pixel placement.

### Root Cause 2: Unsynchronized Global Pointer Mutations in `assets.Load()`
- **Location**: `internal/assets/assets.go` (function `Load()`).
- **Symptoms**:
  - `TestChallenger_MultiThreadedLoadAndPointerRace` in `challenger_stress_test.go` triggered severe `DATA RACE` warnings across all 27 global `*ebiten.Image` pointers (`PlayerImage`, `ZombieImage`, `GrassImage`, etc.), and package tests took ~50 seconds due to 27,000 redundant image decoding passes.
- **Mechanism**:
  1. `Load()` was re-reading embed FS, decoding 27 PNGs, and mutating global pointer variables on every call without synchronization.
  2. When 20 concurrent goroutines executed `Load()` 50 times while 30 reader goroutines concurrently accessed the 27 exported pointers, write-write and write-read races occurred on unprotected pointer addresses.

---

## 2. Exact Proposed Code Changes

### Patch 1: `cmd/tools/genassets/main.go`

#### A. Fix `drawVectorPebble` to blend shadows and clip to diamond boundary
**Target Lines**: 250–288

**Before**:
```go
// Draw rounded vector pebble with highlight and drop shadow
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

	// Pebble body
	minY = int(math.Floor(float64(cy) - ry))
	maxY = int(math.Ceil(float64(cy) + ry))
	minX = int(math.Floor(float64(cx) - rx))
	maxX = int(math.Ceil(float64(cx) + rx))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			normDist := (dx*dx)/(rx*rx) + (dy*dy)/(ry*ry)
			if normDist <= 1.0 {
				c := base
				if dx+dy < -2.0 {
					c = light
				} else if dx+dy > 2.5 {
					c = shadow
				}
				setPixel(img, x, y, c)
			}
		}
	}
}
```

**After**:
```go
// Draw rounded vector pebble with highlight and drop shadow
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
				isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0
				if isoDist <= 1.0 {
					blendPixel(img, x, y, dropShadow)
				}
			}
		}
	}

	// Pebble body
	minY = int(math.Floor(float64(cy) - ry))
	maxY = int(math.Ceil(float64(cy) + ry))
	minX = int(math.Floor(float64(cx) - rx))
	maxX = int(math.Ceil(float64(cx) + rx))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			normDist := (dx*dx)/(rx*rx) + (dy*dy)/(ry*ry)
			if normDist <= 1.0 {
				isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0
				if isoDist <= 1.0 {
					c := base
					if dx+dy < -2.0 {
						c = light
					} else if dx+dy > 2.5 {
						c = shadow
					}
					setPixel(img, x, y, c)
				}
			}
		}
	}
}
```

#### B. Shift pebble position in `generateDirt` safely away from the rim
**Target Lines**: 666–672

**Before**:
```go
	// 4x Scaled rounded vector pebbles (~14x8px)
	pebbles := [][2]int{
		{80, 40}, {180, 56}, {120, 88}, {60, 80}, {195, 36}, {145, 30},
	}
	for _, pos := range pebbles {
		drawVectorPebble(img, pos[0], pos[1], 7.0, 4.0, pebbleBase, pebbleLight, pebbleShadow)
	}
```

**After**:
```go
	// 4x Scaled rounded vector pebbles (~14x8px)
	pebbles := [][2]int{
		{80, 40}, {180, 56}, {120, 88}, {60, 80}, {185, 42}, {145, 30},
	}
	for _, pos := range pebbles {
		drawVectorPebble(img, pos[0], pos[1], 7.0, 4.0, pebbleBase, pebbleLight, pebbleShadow)
	}
```

---

### Patch 2: `internal/assets/assets.go`

#### A. Add `sync.Once` and wrap `Load()` initialization
**Target Lines**: 1–15 and 53–89

**Before**:
```go
package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed images/*
var imageFS embed.FS
```
...
```go
func Load() {
	// Entities
	PlayerImage = loadEbitenImage("images/player.png")
	ZombieImage = loadEbitenImage("images/zombie.png")
	RunnerImage = loadEbitenImage("images/runner.png")

	// Floor Tiles
	GrassImage = loadEbitenImage("images/grass.png")
	DirtImage = loadEbitenImage("images/dirt.png")
	WoodImage = loadEbitenImage("images/wood.png")
	AsphaltImage = loadEbitenImage("images/asphalt.png")
	ConcreteImage = loadEbitenImage("images/concrete.png")
	TileFloorImage = loadEbitenImage("images/tile_floor.png")

	// Vertical Obstacles
	WallImage = loadEbitenImage("images/wall.png")
	TreeImage = loadEbitenImage("images/tree.png")
	FenceImage = loadEbitenImage("images/fence.png")
	DebrisImage = loadEbitenImage("images/debris.png")
	TentImage = loadEbitenImage("images/tent.png")
	StumpImage = loadEbitenImage("images/stump.png")
	MushroomImage = loadEbitenImage("images/mushroom.png")
	SignImage = loadEbitenImage("images/sign.png")
	ElevationBlockImage = loadEbitenImage("images/elevation_block.png")
	ElevationRampImage = loadEbitenImage("images/elevation_ramp.png")

	// Items / Weapons / Armor
	WeaponImage = loadEbitenImage("images/weapon.png")
	AxeImage = loadEbitenImage("images/axe.png")
	ShotgunImage = loadEbitenImage("images/shotgun.png")
	AmmoImage = loadEbitenImage("images/ammo.png")
	ArmorImage = loadEbitenImage("images/armor.png")
	AntidoteImage = loadEbitenImage("images/antidote.png")
	FoodImage = loadEbitenImage("images/food.png")
	WaterImage = loadEbitenImage("images/water.png")
}
```

**After**:
```go
package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed images/*
var imageFS embed.FS

var loadOnce sync.Once
```
...
```go
func Load() {
	loadOnce.Do(func() {
		// Entities
		PlayerImage = loadEbitenImage("images/player.png")
		ZombieImage = loadEbitenImage("images/zombie.png")
		RunnerImage = loadEbitenImage("images/runner.png")

		// Floor Tiles
		GrassImage = loadEbitenImage("images/grass.png")
		DirtImage = loadEbitenImage("images/dirt.png")
		WoodImage = loadEbitenImage("images/wood.png")
		AsphaltImage = loadEbitenImage("images/asphalt.png")
		ConcreteImage = loadEbitenImage("images/concrete.png")
		TileFloorImage = loadEbitenImage("images/tile_floor.png")

		// Vertical Obstacles
		WallImage = loadEbitenImage("images/wall.png")
		TreeImage = loadEbitenImage("images/tree.png")
		FenceImage = loadEbitenImage("images/fence.png")
		DebrisImage = loadEbitenImage("images/debris.png")
		TentImage = loadEbitenImage("images/tent.png")
		StumpImage = loadEbitenImage("images/stump.png")
		MushroomImage = loadEbitenImage("images/mushroom.png")
		SignImage = loadEbitenImage("images/sign.png")
		ElevationBlockImage = loadEbitenImage("images/elevation_block.png")
		ElevationRampImage = loadEbitenImage("images/elevation_ramp.png")

		// Items / Weapons / Armor
		WeaponImage = loadEbitenImage("images/weapon.png")
		AxeImage = loadEbitenImage("images/axe.png")
		ShotgunImage = loadEbitenImage("images/shotgun.png")
		AmmoImage = loadEbitenImage("images/ammo.png")
		ArmorImage = loadEbitenImage("images/armor.png")
		AntidoteImage = loadEbitenImage("images/antidote.png")
		FoodImage = loadEbitenImage("images/food.png")
		WaterImage = loadEbitenImage("images/water.png")
	})
}
```

---

## 3. Test Suite Verification Matrix

| Test Case | Test File | Key Assertions Verified | Status Expected |
|---|---|---|---|
| `TestAssetRegenerationDeterminism` | `genassets_test.go` | SHA-256 bit-for-bit identity across 3 iterations of `genassets` | PASS |
| `TestAssetDimensionsAndIntegrity` | `genassets_test.go` | All 27 assets exist, valid PNGs, non-zero fill ratio $\ge 5\%$ | PASS |
| `TestEmbeddedAssetDimensionsAndValidity` | `assets_test.go` | Dimensions (64x128, 256x128, 256x256, 64x64) and non-transparent pixels | PASS |
| `TestAssetsLoadAllPointersNonNil` | `assets_test.go` | All 27 exported pointers non-nil after `Load()` | PASS |
| `TestFloorTileIsometricBounds` | `assets_stress_test.go` | Floor tiles do not bleed beyond $\text{dist} > 1.15$ | PASS |
| `TestCharacterGroundAnchor` | `assets_stress_test.go` | Grounding pixels present in rows 112..127 | PASS |
| `TestItemOutlineContrast` | `assets_stress_test.go` | Item pixel count $\ge 320$ and dark border pixels present | PASS |
| `TestAssetsLoadIdempotency` | `assets_stress_test.go` | Multiple sequential `Load()` calls retain non-nil handles | PASS |
| `TestEmpiricalAssetCatalogCompleteness` | `empirical_challenger_test.go` | Exactly 27 assets with target dimensions | PASS |
| `TestEmpiricalAlphaFillRatios` | `empirical_challenger_test.go` | Alpha fill ratio within $[MinFillRatio, MaxFillRatio]$ | PASS |
| `TestEmpiricalFloorDiamondGeometry` | `empirical_challenger_test.go` | Strict diamond boundary ($\text{isoDist} > 1.0 \Rightarrow A = 0$) and solid core ($\text{isoDist} \le 0.85 \Rightarrow A = 255$) | PASS |
| `TestEmpiricalCharacterGrounding` | `empirical_challenger_test.go` | Grounding pixels in rows 112..127 $\ge 50$ | PASS |
| `TestEmpiricalGenerationDeterminism` | `empirical_challenger_test.go` | SHA-256 consistency across repeated runs | PASS |
| `TestEmpiricalObstacleBoundsAndGrounding` | `empirical_challenger_test.go` | Obstacles have ground contact pixels in lower half ($y \ge 128$) | PASS |
| `TestEmpiricalItemIconQuality` | `empirical_challenger_test.go` | Centroid within $[20..44, 20..44]$ box | PASS |
| `TestChallenger_All27ExportedPointersAndExactBounds` | `challenger_stress_test.go` | Exact 27 pointers, non-nil, exact bounds | PASS |
| `TestChallenger_MultiThreadedLoadAndPointerRace` | `challenger_stress_test.go` | 20 loaders + 30 readers concurrent stress under `-race` | PASS |
| `TestChallenger_RepeatedSequentialLoads` | `challenger_stress_test.go` | 100 sequential `Load()` calls | PASS |
| `TestChallenger_AssetPixelContrastAndColorSaturation` | `challenger_stress_test.go` | RMS contrast $> 10.0$, dyn range $\ge 40.0$, color saturation $> 0.20$ | PASS |
| `TestChallenger_FloorTileGeometryDiamond` | `challenger_stress_test.go` | Diamond fit and interior fill ratio $\ge 98\%$ | PASS |
| `TestChallenger_CharacterGroundDropShadows` | `challenger_stress_test.go` | Head pixels in $[0..40]$, shadow pixels $\ge 50$ in $[112..127]$ | PASS |
| `TestChallenger_ItemIconCenteringAndContour` | `challenger_stress_test.go` | Mass $\ge 200$, centroid centered | PASS |

---

## 4. Verification Commands

To verify the fixes end-to-end:

```bash
# 1. Regenerate procedural assets with the updated generator
go run ./cmd/tools/genassets

# 2. Run the complete test suite with Go Race Detector and GCC C toolchain
CC=gcc go test -race -v ./cmd/tools/genassets/... ./internal/assets/...
```

Expected execution time drops from ~50s down to ~1.5s, with 0 data races and 100% test pass.
