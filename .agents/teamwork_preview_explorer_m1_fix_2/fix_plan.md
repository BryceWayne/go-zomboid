# Fix Plan: Procedural Asset Generation & Assets Loading Concurrency Fixes

**Target Milestone**: Milestone 1 Remediation  
**Author**: `m1_explorer_fix_2`  
**Date**: 2026-08-28  
**Scope**: `cmd/tools/genassets/main.go`, `internal/assets/assets.go`, and floor tile generators (`grass`, `dirt`, `wood`, `asphalt`, `concrete`, `tile_floor`).

---

## 1. Executive Summary of Defects

Two distinct defects were identified during challenger stress testing of Milestone 1:

1. **`images/dirt.png` Alpha Hole Punctures & Isometric Diamond Bleed**:
   - `drawVectorPebble()` in `cmd/tools/genassets/main.go` sets semi-transparent pixels (`RGBA{0, 0, 0, 45}`) directly with `setPixel()` instead of alpha blending via `blendPixel()`. This punches 151 semi-transparent holes into the opaque dirt tile.
   - `drawVectorPebble()` lacks diamond boundary clipping ($d_{iso} \le 1.0$), and pebble 5 in `generateDirt()` is centered at `{195, 36}` with radius $(r_x=7, r_y=4)$, causing 18 non-transparent pixels to spill past the 2:1 isometric diamond perimeter ($d_{iso} = 1.0195 > 1.0$).
2. **Data Race in `internal/assets.Load()`**:
   - `internal/assets.Load()` mutates 27 global `*ebiten.Image` pointers without synchronization, causing data races when invoked concurrently across goroutines (detected by Go's `-race` detector in `TestChallenger_MultiThreadedLoadAndPointerRace`).

---

## 2. Floor Generators Audit Matrix

A complete mathematical and source inspection of all 6 floor tile generators was conducted against the 2:1 isometric diamond formula:
$$d_{iso}(x, y) = \frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0}$$

| Floor Generator | Detail Elements | Alpha Integrity ($\alpha=255$) | Diamond Boundary ($d_{iso} \le 1.0$) | Status | Details / Findings |
|---|---|---|---|---|---|
| `generateGrass` | Base tile, 8 Chevrons, 5 Wildflowers | **100% Solid** | **Max $d_{iso} \le 0.918$** | **CLEAN** | All blade and petal coordinates strictly inside bounds; all colors opaque ($A=255$). |
| `generateDirt` | Base tile, 6 Rounded Vector Pebbles | **FAILED** (151 holes with $A=45$) | **FAILED** (18 pixels with $d_{iso} > 1.0$) | **DEFECTIVE** | `drawVectorPebble` uses `setPixel` with drop shadow $A=45$, and pebble `{195, 36}` bleeds past $d_{iso}=1.0$. |
| `generateWoodFloor` | Base tile, Plank seams, Stepped bevels, 8 Nailheads | **100% Solid** | **Max $d_{iso} \le 0.919$** | **CLEAN** | Nails explicitly guarded by $d_{iso} \le 0.88$; colors opaque ($A=255$). |
| `generateAsphalt` | Base tile, Yellow dashed lane markings, Stepped bevels | **100% Solid** | **Max $d_{iso} \le 1.000$** | **CLEAN** | Pixel iteration completely enclosed within `if isoDist <= 1.0`; all colors opaque. |
| `generateConcrete` | Base tile, 4 Slab quadrants, 3px Expansion joints, Bevels | **100% Solid** | **Max $d_{iso} \le 1.000$** | **CLEAN** | Pixel iteration completely enclosed within `if isoDist <= 1.0`; all colors opaque. |
| `generateTileFloor` | Base tile, 4x4 Checkerboard, Grout lines, Tile bevels | **100% Solid** | **Max $d_{iso} \le 1.000$** | **CLEAN** | Pixel iteration completely enclosed within `if isoDist <= 1.0`; all colors opaque. |

**Audit Conclusion**: `generateDirt` is the **only** floor generator with alpha corruption and boundary bleed defects. All other floor generators (`grass`, `wood`, `asphalt`, `concrete`, `tile_floor`) strictly enforce $d_{iso} \le 1.0$ and 100% alpha opacity.

---

## 3. Root Cause Analysis & Proposed Fixes

### Issue 1: `drawVectorPebble` & `generateDirt` Fix

#### Root Cause:
1. In `cmd/tools/genassets/main.go:250-288`:
   ```go
   func drawVectorPebble(img *image.RGBA, cx, cy int, rx, ry float64, base, light, shadow color.RGBA) {
       dropShadow := color.RGBA{0, 0, 0, 45}
       ...
       for y := minY; y <= maxY; y++ {
           for x := minX; x <= maxX; x++ {
               ...
               if dx*dx+dy*dy <= 1.0 {
                   setPixel(img, x, y, dropShadow) // <-- OVERWRITES base pixel!
               }
           }
       }
   ```
   `setPixel` overwrites the destination pixel RGBA values directly. Because `dropShadow` has $A=45$, the destination pixel becomes semi-transparent ($A=45$), creating 151 translucent holes where the drop shadow is not occluded by the pebble body.
2. `drawVectorPebble` performs no check against the isometric diamond boundary ($d_{iso} \le 1.0$).
3. In `generateDirt()`, pebble 5 is placed at `{195, 36}`. The center has $d_{iso} = \frac{67.5}{128} + \frac{27.5}{64} = 0.9570$. With $r_x=7, r_y=4$, pixel $(202, 36)$ has $d_{iso} = 1.0117$ and drop shadow $(203, 36)$ has $d_{iso} = 1.0195$, spilling 18 pixels into the transparent outer quadrant.

#### Proposed Remediation:
1. Update `drawVectorPebble` in `cmd/tools/genassets/main.go`:
   - Replace `setPixel(img, x, y, dropShadow)` with `blendPixel(img, x, y, dropShadow)`.
   - Add boundary check `isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0; if isoDist <= 1.0` to both drop shadow and pebble body loops.
2. In `generateDirt()`:
   - Adjust pebble position `{195, 36}` inward to `{185, 42}` (where $d_{iso} = 0.7852$, fully containing the pebble within bounds without visual clipping).

```go
// Proposed replacement for drawVectorPebble:
func drawVectorPebble(img *image.RGBA, cx, cy int, rx, ry float64, base, light, shadow color.RGBA) {
	dropShadow := color.RGBA{0, 0, 0, 45}
	// Drop shadow
	minY := int(math.Floor(float64(cy+2) - ry))
	maxY := int(math.Ceil(float64(cy+2) + ry))
	minX := int(math.Floor(float64(cx+2) - rx))
	maxX := int(math.Ceil(float64(cx+2) + rx))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0
			if isoDist <= 1.0 {
				dx := float64(x-(cx+2)) / rx
				dy := float64(y-(cy+2)) / ry
				if dx*dx+dy*dy <= 1.0 {
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
			isoDist := math.Abs(float64(x)-127.5)/128.0 + math.Abs(float64(y)-63.5)/64.0
			if isoDist <= 1.0 {
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
}
```

---

### Issue 2: `internal/assets.Load()` Concurrency Fix

#### Root Cause:
In `internal/assets/assets.go:53-88`, `Load()` performs direct assignments to global package variables (`PlayerImage = loadEbitenImage(...)`, etc.) on every call without mutex or `sync.Once` guards. When multiple goroutines invoke `Load()` or query pointers concurrently, Go's race detector flags read/write and write/write conflicts.

#### Proposed Remediation:
1. Import `"sync"` in `internal/assets/assets.go`.
2. Add package-level variable `var loadOnce sync.Once`.
3. Wrap all asset loading in `loadOnce.Do(func() { ... })`.

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

var (
	// Entity Sprites (16x32)
	PlayerImage *ebiten.Image
	ZombieImage *ebiten.Image
	RunnerImage *ebiten.Image
	// ... (all 27 pointers unchanged)
)

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

## 4. Implementation & Verification Workflow for Worker Agent

1. **Apply Code Edits**:
   - Edit `cmd/tools/genassets/main.go` to update `drawVectorPebble` and `generateDirt`.
   - Edit `internal/assets/assets.go` to add `sync.Once`.
2. **Regenerate Assets**:
   - Execute: `go run ./cmd/tools/genassets`
   - Verify `images/dirt.png` is updated.
3. **Execute Verification Tests**:
   - `CC=gcc go test -v ./internal/assets/... ./cmd/tools/genassets/...` (All empirical challenger tests must PASS).
   - `CC=gcc go test -v -race ./internal/assets/... -run TestChallenger_MultiThreadedLoadAndPointerRace` (Must PASS with zero data race warnings).
   - `CC=gcc go test ./...` (Whole project test suite passes).
