# Comprehensive Fix Plan: Procedural Asset Generation & Thread-Safe Asset Loading

**Target Milestone**: Milestone 1 Remediation  
**Author**: `m1_explorer_fix_1`  
**Date**: 2026-08-28  
**Target Files**:
1. `cmd/tools/genassets/main.go`
2. `internal/assets/assets.go`
3. Regenerated asset: `internal/assets/images/dirt.png`

---

## 1. Executive Summary & Problem Formulation

Challenger adversarial testing and `-race` testing identified two distinct blocking issues in Milestone 1:

1. **`images/dirt.png` Core Alpha Corruption & Diamond Boundary Bleed** (`cmd/tools/genassets/main.go:250-288, 667`):
   - `drawVectorPebble` writes drop shadow pixels (`RGBA{0, 0, 0, 45}`) directly using `setPixel()` instead of `blendPixel()`. This overwrites opaque dirt pixels with semi-transparent pixels ($A=45$), creating 151 punctured translucent holes inside the solid diamond core ($isoDist \le 0.85$).
   - `drawVectorPebble` lacks diamond boundary bounds clipping ($isoDist \le 1.0$), and `generateDirt()` places a pebble at `{195, 36}` with radius $r_x=7, r_y=4$. The center has $isoDist = \frac{|195-127.5|}{128} + \frac{|36-63.5|}{64} = 0.9570$, causing 18 non-transparent pixels to spill past the 2:1 isometric diamond perimeter ($isoDist = 1.0195 > 1.0$).

2. **Data Race in `internal/assets.Load()`** (`internal/assets/assets.go:53-88`):
   - `Load()` assigns 27 global `*ebiten.Image` pointers without synchronization. When invoked concurrently across goroutines (e.g. during concurrent scene transitions or under `go test -race` in `TestChallenger_MultiThreadedLoadAndPointerRace`), Go's race detector flags read/write and write/write data races.

---

## 2. Mathematical Analysis & Geometry Audit

The standard 2:1 isometric diamond metric on a $256 \times 128$ bounding box with center $(127.5, 63.5)$ and semi-axes $r_x = 128.0, r_y = 64.0$ is:
$$d_{iso}(x, y) = \frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0}$$

### Pebble Position Analysis:
- Original pebble 5 at `{195, 36}`:
  - Center: $d_{iso}(195, 36) = \frac{67.5}{128.0} + \frac{27.5}{64.0} = 0.5273 + 0.4297 = 0.9570$.
  - Extreme body pixel $(202, 36)$: $d_{iso} = \frac{74.5}{128.0} + \frac{27.5}{64.0} = 0.5820 + 0.4297 = 1.0117 > 1.0$.
  - Extreme shadow pixel $(204, 38)$: $d_{iso} = \frac{76.5}{128.0} + \frac{25.5}{64.0} = 0.5976 + 0.3984 = 0.9960$.
  - Result: 18 non-transparent pixels violate the bounding diamond.
- Repositioned pebble 5 at `{185, 42}`:
  - Center: $d_{iso}(185, 42) = \frac{57.5}{128.0} + \frac{21.5}{64.0} = 0.4492 + 0.3359 = 0.7852$.
  - Extreme body pixel $(192, 42)$: $d_{iso} = \frac{64.5}{128.0} + \frac{21.5}{64.0} = 0.5039 + 0.3359 = 0.8398 \le 0.85$.
  - Extreme shadow pixel $(194, 44)$: $d_{iso} = \frac{66.5}{128.0} + \frac{19.5}{64.0} = 0.5195 + 0.3047 = 0.8242 \le 0.85$.
  - Result: 0 pixels bleed outside $d_{iso} \le 1.0$; pebble resides fully in the floor surface without clipping.

### Drop Shadow Blending Formula:
Using `blendPixel(img, x, y, dropShadow)` with $c = \{0, 0, 0, 45\}$ on an opaque background $dst = \{R, G, B, 255\}$:
- $srcA = 45 / 255.0 \approx 0.1765$
- $dstA = 255 / 255.0 = 1.0$
- $outA = srcA + dstA \cdot (1 - srcA) = 0.1765 + 1.0 \cdot (1 - 0.1765) = 1.0$
- $outR = \frac{0 \cdot srcA + dst.R \cdot dstA \cdot (1 - srcA)}{1.0} = dst.R \cdot (1 - 0.1765) \approx 0.8235 \cdot dst.R$
- Result: Background is darkened by $\approx 17.6\%$, while resulting alpha remains $100\%$ opaque ($A=255$).

---

## 3. Exact Code Fixes

### Fix 1: `cmd/tools/genassets/main.go`

#### A. Modify `drawVectorPebble` (lines 250-288)
Replace `setPixel(img, x, y, dropShadow)` with `blendPixel(img, x, y, dropShadow)` and add explicit $isoDist \le 1.0$ diamond boundary checks:

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

#### B. Modify `generateDirt` (line 667)
Adjust pebble position `{195, 36}` inward to `{185, 42}`:

```go
	// 4x Scaled rounded vector pebbles (~14x8px)
	pebbles := [][2]int{
		{80, 40}, {180, 56}, {120, 88}, {60, 80}, {185, 42}, {145, 30},
	}
```

---

### Fix 2: `internal/assets/assets.go`

Add `sync.Once` synchronization so that `Load()` safely runs initialization once across all concurrent goroutines.

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

	// Floor Tiles (64x32)
	GrassImage     *ebiten.Image
	DirtImage      *ebiten.Image
	WoodImage      *ebiten.Image
	AsphaltImage   *ebiten.Image
	ConcreteImage  *ebiten.Image
	TileFloorImage *ebiten.Image

	// Vertical Obstacles / Props (64x64)
	WallImage           *ebiten.Image
	TreeImage           *ebiten.Image
	FenceImage          *ebiten.Image
	DebrisImage         *ebiten.Image
	TentImage           *ebiten.Image
	StumpImage          *ebiten.Image
	MushroomImage       *ebiten.Image
	SignImage           *ebiten.Image
	ElevationBlockImage *ebiten.Image
	ElevationRampImage  *ebiten.Image

	// Item / Weapon / Armor Sprites (16x16)
	WeaponImage  *ebiten.Image
	AxeImage     *ebiten.Image
	ShotgunImage *ebiten.Image
	AmmoImage    *ebiten.Image
	ArmorImage   *ebiten.Image
	AntidoteImage *ebiten.Image
	FoodImage    *ebiten.Image
	WaterImage   *ebiten.Image
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

func loadEbitenImage(path string) *ebiten.Image {
	data, err := imageFS.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read embedded image %s: %v", path, err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to decode image %s: %v", path, err)
	}

	return ebiten.NewImageFromImage(img)
}
```

---

## 4. Implementation Steps & Verification Protocol

1. Apply modifications to `cmd/tools/genassets/main.go`.
2. Apply modifications to `internal/assets/assets.go`.
3. Regenerate all 27 assets:
   ```bash
   go run ./cmd/tools/genassets
   ```
4. Verify asset generator tests:
   ```bash
   go test -v ./cmd/tools/genassets
   ```
5. Verify adversarial challenger tests:
   ```bash
   CC=clang go test -v ./internal/assets -run TestEmpiricalFloorDiamondGeometry
   ```
6. Verify race-free concurrency:
   ```bash
   CC=clang go test -race -v ./internal/assets -run TestChallenger_MultiThreadedLoadAndPointerRace
   ```
7. Verify full test suite across the project:
   ```bash
   CC=clang go test -v ./...
   ```
