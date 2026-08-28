# Handoff Report: Environment Tile Procedural Sprites (Milestone 1)

**Agent ID**: `teamwork_preview_explorer_m1_2`  
**Working Directory**: `/home/bryce/code/go-zomboid/.agents/teamwork_preview_explorer_m1_2`  
**Target Scope**: Milestone 1 - Environment Tile Procedural Sprites in `cmd/tools/genassets`

---

## 1. Observation

1. **Existing Asset Generator**:
   - `cmd/tools/genassets/main.go` currently produces crude, low-fidelity placeholder tiles (`grass.png`, `dirt.png`, `wood.png`, `wall.png`, `tree.png`) using basic solid colors and uniform random noise.
   - It is missing several required environment tiles specified in `PROJECT.md` §Interface Contracts and `ORIGINAL_REQUEST.md`: `asphalt.png`, `concrete.png`, `tile_floor.png`, `fence.png`, and `debris.png`.

2. **Engine Isometric Projection Math** (`internal/game/game.go:546-550`, `610-613`, `658-669`):
   - World tiles are square grid cells of size `world.TileSize = 32` (`internal/game/world/map.go:20`).
   - World-to-isometric screen projection is defined as:
     $$\text{isoX} = wx - wy, \quad \text{isoY} = \frac{wx + wy}{2}$$
   - Floor diamonds (64x32) are drawn at:
     $$\text{drawX} = \text{isoX} - 32 - \text{camX}, \quad \text{drawY} = \text{isoY} - 0 - \text{camY}$$
   - Vertical obstacles (64x64) are drawn at:
     $$\text{drawX} = \text{isoX} - 32 - \text{camX}, \quad \text{drawY} = \text{isoY} - 32 - \text{camY}$$
   - This establishes that the bottom 32 vertical pixels ($y \in [32, 64]$) of a 64x64 obstacle match the exact diamond footprint of a 64x32 floor tile, while the top 32 vertical pixels ($y \in [0, 32]$) represent the 3D height extrusion.

3. **Interface Contract & Assets Loading** (`internal/assets/assets.go:17-42`):
   - Floor sprites must be exported as 64x32 PNGs: `grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png`.
   - Vertical obstacle sprites must be exported as 64x64 PNGs: `wall.png`, `tree.png`, `fence.png`, `debris.png`.
   - Pure Go standard libraries (`image`, `image/color`, `image/png`, `math`, `math/rand`) must be used without external C/dynamic graphics libraries.

---

## 2. Logic Chain

1. **Pixel-Perfect 2:1 Isometric Diamond Formulation**:
   - For a 64x32 tile ($x \in [0, 63], y \in [0, 31]$), the diamond center is at $cx = 31.5, cy = 15.5$.
   - A pixel is inside the diamond if:
     $$\text{isoDist} = \frac{|x - 31.5|}{32.0} + \frac{|y - 15.5|}{16.0} \le 1.0$$
   - Invertible local world coordinate parameterization $(u, v) \in [0.0, 1.0]^2$:
     $$u = \frac{x - 31.5}{64.0} + \frac{y - 15.5}{32.0} + 0.5$$
     $$v = \frac{y - 15.5}{32.0} - \frac{x - 31.5}{64.0} + 0.5$$
   - This mapping allows mapping textures, wood planks, lane markings, mortar joints, and expansion seams onto the isometric plane without perspective distortion.

2. **Floor Diamond Shading & Texture Pipeline**:
   - **`grass.png`**: Multi-octave sinusoidal/value noise for patchy field variations, darker soil undergrowth base, ~36 seeded vertical/slanted grass blade streaks with root shadow, clover/wildflower accents, and edge bevel darkening at $\text{isoDist} > 0.90$.
   - **`dirt.png`**: Uneven soil density noise, embedded pebbles with top-left highlight and bottom-right cast shadow, and edge depression shading.
   - **`wood.png`**: Divides $v$ into 4 distinct plank lanes with recessed seam lines, staggered end joints across lanes at $u = [0.60, 0.30, 0.75, 0.45]$, fine grain striations along $u$, and iron nail heads at board ends.
   - **`asphalt.png`**: Dark granular aggregate stippling, weathered/chipped yellow dashed lane marking down the center line ($v \in [0.43, 0.57]$), and hairline tar cracks.
   - **`concrete.png`**: 4 quadrant sidewalk slabs with subtle per-slab brightness variation, 1px dark expansion joint seams along $u = 0.5$ and $v = 0.5$ with edge groove highlights, and aggregate speckles.
   - **`tile_floor.png`**: $4 \times 4$ checkerboard ceramic grid with recessed dark grout lines, light cream vs dark slate alternating tiles, and top-left specular sheen bevels.

3. **3D Vertical Obstacle (64x64) Geometry & Shading Pipeline**:
   - **`wall.png`**: 3D brick masonry prism:
     * Top coping stone slab centered at $(32, 14)$ with 2px stone lip and highlights.
     * Left face ($x \in [0, 32]$): 8 running bond brick courses along $+0.5$ slope, staggered vertical joints, lit at $88\%$ luminance.
     * Right face ($x \in [32, 64]$): 8 running bond courses along $-0.5$ slope, in shadow at $65\%$ luminance.
     * Ambient occlusion contact shadow along the base.
   - **`tree.png`**: 3D layered evergreen pine:
     * Textured trunk with root flares at $y \in [42, 60]$.
     * 3 overlapping conical canopy tiers ($y = [30..50], [16..36], [4..22]$) with scalloped needle edges.
     * Directional top-left sunlit highlights and deep underside shadows.
   - **`fence.png`**: 3D wooden picket boundary fence:
     * Corner posts with pyramid caps, 2 horizontal support rails along isometric slope, 7 vertical pickets with pointed tips, and iron fasteners.
   - **`debris.png`**: Supply crate with rubble:
     * 3D wooden crate with outer frame, diagonal $X$-braces, and metal corner brackets.
     * Scattered concrete boulders, broken red brick fragments, and soft ground contact shadow.

---

## 3. Recommended Code Architecture & Implementation

Here is the complete, drop-in implementation code designed for `cmd/tools/genassets/main.go`:

```go
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
)

// Color manipulation helpers
func darken(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

func lighten(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(c.R)*factor)),
		G: uint8(math.Min(255, float64(c.G)*factor)),
		B: uint8(math.Min(255, float64(c.B)*factor)),
		A: c.A,
	}
}

func blend(c1, c2 color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		A: uint8(float64(c1.A)*(1-t) + float64(c2.A)*t),
	}
}

// -------------------------------------------------------------
// FLOOR TILES (64x32)
// -------------------------------------------------------------

func generateGrass(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(42))

	baseDark := color.RGBA{38, 77, 36, 255}
	baseMid := color.RGBA{48, 98, 45, 255}
	baseLight := color.RGBA{68, 128, 55, 255}
	soilDark := color.RGBA{24, 50, 22, 255}
	weedColor := color.RGBA{92, 115, 48, 255}
	flowerColor := color.RGBA{225, 210, 110, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				n := math.Sin(u*18.0)*math.Cos(v*18.0)*0.5 + 0.5
				n2 := math.Sin(u*37.0+v*23.0)*0.5 + 0.5
				val := n*0.7 + n2*0.3

				var c color.RGBA
				if val < 0.35 {
					c = blend(soilDark, baseDark, val/0.35)
				} else if val < 0.75 {
					c = blend(baseDark, baseMid, (val-0.35)/0.40)
				} else {
					c = blend(baseMid, baseLight, (val-0.75)/0.25)
				}

				if rng.Float64() < 0.25 {
					c = blend(c, weedColor, 0.3)
				}

				if isoDist > 0.90 {
					if y >= 16 {
						c = darken(c, 0.75)
					} else {
						c = darken(c, 0.88)
					}
				}

				img.Set(x, y, c)
			}
		}
	}

	// Grass Blades
	for i := 0; i < 36; i++ {
		bx := 6 + rng.Intn(52)
		by := 4 + rng.Intn(24)
		dx := float64(bx) - 31.5
		dy := float64(by) - 15.5
		if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.85 {
			img.Set(bx, by, soilDark)
			tipX := bx + (rng.Intn(3) - 1)
			tipY := by - (1 + rng.Intn(2))
			if tipY >= 0 {
				img.Set(tipX, tipY, baseLight)
				if rng.Float64() < 0.3 {
					img.Set(tipX, tipY-1, weedColor)
				}
			}
		}
	}

	// Wildflower Accents
	for i := 0; i < 4; i++ {
		fx := 10 + rng.Intn(44)
		fy := 6 + rng.Intn(20)
		dx := float64(fx) - 31.5
		dy := float64(fy) - 15.5
		if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.75 {
			img.Set(fx, fy, flowerColor)
			img.Set(fx, fy+1, soilDark)
		}
	}

	saveImg(name, img)
}

func generateDirt(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(101))

	baseDark := color.RGBA{68, 46, 32, 255}
	baseMid := color.RGBA{92, 64, 45, 255}
	baseLight := color.RGBA{118, 85, 60, 255}
	pebbleBase := color.RGBA{120, 115, 105, 255}
	pebbleHigh := color.RGBA{165, 160, 150, 255}
	pebbleShadow := color.RGBA{40, 28, 20, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				n := math.Sin(u*24.0)*math.Cos(v*24.0)*0.5 + 0.5
				n2 := (rng.Float64() - 0.5) * 0.3
				val := math.Max(0, math.Min(1, n+n2))

				var c color.RGBA
				if val < 0.4 {
					c = blend(baseDark, baseMid, val/0.4)
				} else {
					c = blend(baseMid, baseLight, (val-0.4)/0.6)
				}

				if isoDist > 0.90 {
					c = darken(c, 0.78)
				}

				img.Set(x, y, c)
			}
		}
	}

	for i := 0; i < 14; i++ {
		px := 8 + rng.Intn(48)
		py := 4 + rng.Intn(24)
		dx := float64(px) - 31.5
		dy := float64(py) - 15.5
		if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.80 {
			if rng.Intn(2) == 0 {
				img.Set(px, py, pebbleHigh)
				img.Set(px, py+1, pebbleShadow)
			} else {
				img.Set(px, py, pebbleHigh)
				img.Set(px+1, py, pebbleBase)
				img.Set(px, py+1, pebbleBase)
				img.Set(px+1, py+1, pebbleShadow)
			}
		}
	}

	saveImg(name, img)
}

func generateWoodFloor(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(202))

	plankColors := []color.RGBA{
		{142, 92, 54, 255},
		{128, 80, 46, 255},
		{156, 104, 62, 255},
		{136, 88, 50, 255},
	}
	seamDark := color.RGBA{45, 26, 14, 255}
	seamLight := color.RGBA{180, 130, 85, 255}
	nailColor := color.RGBA{30, 22, 18, 255}
	endJoints := []float64{0.60, 0.30, 0.75, 0.45}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				laneFloat := v * 4.0
				lane := int(math.Floor(laneFloat))
				if lane < 0 { lane = 0 }
				if lane > 3 { lane = 3 }
				vInLane := laneFloat - float64(lane)

				baseC := plankColors[lane%len(plankColors)]
				grain := math.Sin(u*45.0+float64(lane)*10.0)*8.0 + (rng.Float64()-0.5)*6.0
				c := color.RGBA{
					R: uint8(math.Max(0, math.Min(255, float64(baseC.R)+grain))),
					G: uint8(math.Max(0, math.Min(255, float64(baseC.G)+grain*0.8))),
					B: uint8(math.Max(0, math.Min(255, float64(baseC.B)+grain*0.6))),
					A: 255,
				}

				if vInLane < 0.09 {
					c = seamDark
				} else if vInLane > 0.91 {
					c = blend(c, seamLight, 0.4)
				}

				endU := endJoints[lane]
				if math.Abs(u-endU) < 0.025 {
					c = seamDark
				}

				if isoDist > 0.92 {
					c = darken(c, 0.82)
				}

				img.Set(x, y, c)
			}
		}
	}

	for lane := 0; lane < 4; lane++ {
		endU := endJoints[lane]
		for _, offset := range []float64{-0.05, 0.05} {
			nu := endU + offset
			nv := (float64(lane) + 0.5) / 4.0
			nx := int(math.Round(31.5 + (nu-nv)*32.0))
			ny := int(math.Round(15.5 + (nu+nv-1.0)*16.0))
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				dx := float64(nx) - 31.5
				dy := float64(ny) - 15.5
				if math.Abs(dx)/32.0+math.Abs(dy)/16.0 <= 0.85 {
					img.Set(nx, ny, nailColor)
				}
			}
		}
	}

	saveImg(name, img)
}

func generateAsphalt(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(303))

	baseDark := color.RGBA{38, 40, 44, 255}
	baseMid := color.RGBA{48, 50, 55, 255}
	stippleLight := color.RGBA{68, 70, 76, 255}
	yellowMarking := color.RGBA{220, 180, 45, 255}
	yellowDark := color.RGBA{160, 130, 30, 255}
	crackDark := color.RGBA{22, 23, 26, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				n := rng.Float64()
				var c color.RGBA
				if n < 0.20 {
					c = baseDark
				} else if n < 0.80 {
					c = baseMid
				} else {
					c = stippleLight
				}

				if v >= 0.43 && v <= 0.57 && (u <= 0.38 || u >= 0.62) {
					if rng.Float64() > 0.15 {
						if v < 0.46 {
							c = yellowMarking
						} else if v > 0.54 {
							c = yellowDark
						} else {
							c = yellowMarking
						}
					}
				}

				if math.Abs(u-0.5+math.Sin(v*12.0)*0.08) < 0.015 && rng.Float64() < 0.7 {
					c = crackDark
				}

				if isoDist > 0.92 {
					c = darken(c, 0.80)
				}

				img.Set(x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateConcrete(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(404))

	slabBase := color.RGBA{145, 145, 142, 255}
	slabLight := color.RGBA{168, 168, 165, 255}
	jointDark := color.RGBA{50, 50, 50, 255}
	jointLight := color.RGBA{185, 185, 182, 255}
	stipple := color.RGBA{100, 100, 98, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				quadX := 0
				if u >= 0.5 { quadX = 1 }
				quadY := 0
				if v >= 0.5 { quadY = 1 }
				slabShift := (float64((quadX*2+quadY)*7%11) - 5.0) * 2.0

				n := rng.Float64()
				var c color.RGBA
				if n < 0.15 {
					c = stipple
				} else if n < 0.70 {
					c = slabBase
				} else {
					c = slabLight
				}

				c.R = uint8(math.Max(0, math.Min(255, float64(c.R)+slabShift)))
				c.G = uint8(math.Max(0, math.Min(255, float64(c.G)+slabShift)))
				c.B = uint8(math.Max(0, math.Min(255, float64(c.B)+slabShift)))

				distU := math.Abs(u - 0.5)
				distV := math.Abs(v - 0.5)
				if distU < 0.025 || distV < 0.025 {
					c = jointDark
				} else if (u > 0.5 && distU < 0.05) || (v > 0.5 && distV < 0.05) {
					c = jointLight
				}

				if isoDist > 0.92 {
					c = darken(c, 0.82)
				}

				img.Set(x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateTileFloor(name string) {
	w, h := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(505))

	tileALight := color.RGBA{225, 225, 220, 255}
	tileABase := color.RGBA{200, 200, 195, 255}
	tileADark := color.RGBA{175, 175, 170, 255}

	tileBLight := color.RGBA{85, 98, 110, 255}
	tileBBase := color.RGBA{65, 75, 85, 255}
	tileBDark := color.RGBA{45, 52, 60, 255}

	groutDark := color.RGBA{35, 38, 42, 255}
	groutLight := color.RGBA{150, 150, 150, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 15.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/16.0
			if isoDist <= 1.0 {
				u := dx/64.0 + dy/32.0 + 0.5
				v := dy/32.0 - dx/64.0 + 0.5

				gridU := u * 4.0
				gridV := v * 4.0
				tileU := int(math.Floor(gridU))
				tileV := int(math.Floor(gridV))
				if tileU < 0 { tileU = 0 }
				if tileU > 3 { tileU = 3 }
				if tileV < 0 { tileV = 0 }
				if tileV > 3 { tileV = 3 }

				subU := gridU - float64(tileU)
				subV := gridV - float64(tileV)

				isTileA := (tileU+tileV)%2 == 0
				var c color.RGBA

				if subU < 0.08 || subV < 0.08 {
					c = groutDark
				} else if subU > 0.92 || subV > 0.92 {
					c = blend(groutDark, groutLight, 0.4)
				} else {
					if isTileA {
						if subU < 0.25 || subV < 0.25 {
							c = tileALight
						} else if subU > 0.75 || subV > 0.75 {
							c = tileADark
						} else {
							c = tileABase
						}
					} else {
						if subU < 0.25 || subV < 0.25 {
							c = tileBLight
						} else if subU > 0.75 || subV > 0.75 {
							c = tileBDark
						} else {
							c = tileBBase
						}
					}
					if rng.Float64() < 0.15 {
						c = lighten(c, 1.05)
					}
				}

				if isoDist > 0.92 {
					c = darken(c, 0.80)
				}

				img.Set(x, y, c)
			}
		}
	}

	saveImg(name, img)
}

// -------------------------------------------------------------
// VERTICAL OBSTACLE BLOCKS (64x64)
// -------------------------------------------------------------

func generateIsoWall(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(606))

	copingTop := color.RGBA{185, 180, 175, 255}
	copingHigh := color.RGBA{220, 215, 210, 255}
	copingShadow := color.RGBA{115, 110, 105, 255}

	brickBaseRed := color.RGBA{148, 56, 42, 255}
	brickDarkRed := color.RGBA{115, 40, 28, 255}
	brickLightRed := color.RGBA{175, 75, 55, 255}

	mortarLeft := color.RGBA{190, 185, 175, 255}
	mortarRight := color.RGBA{110, 105, 100, 255}

	// 1. Top Coping Stone
	for y := 0; y < 28; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 31.5
			dy := float64(y) - 13.5
			isoDist := math.Abs(dx)/32.0 + math.Abs(dy)/14.0
			if isoDist <= 1.0 {
				c := copingTop
				if isoDist > 0.88 {
					if y < 14 {
						c = copingHigh
					} else {
						c = copingShadow
					}
				} else if rng.Float64() < 0.2 {
					c = lighten(copingTop, 1.08)
				}
				img.Set(x, y, c)
			}
		}
	}

	for x := 0; x < 32; x++ {
		edgeY := int(math.Round(14.0 + float64(x)*0.5))
		img.Set(x, edgeY+1, copingShadow)
	}
	for x := 32; x < 64; x++ {
		edgeY := int(math.Round(30.0 - float64(x-32)*0.5))
		img.Set(x, edgeY+1, darken(copingShadow, 0.8))
	}

	// 2. Left Face (West Wall)
	for x := 0; x < 32; x++ {
		topY := 15 + x/2
		botY := 47 + x/2
		for y := topY + 1; y <= botY && y < h; y++ {
			hRel := y - topY
			course := hRel / 4
			courseY := hRel % 4

			if courseY == 0 {
				img.Set(x, y, mortarLeft)
			} else {
				jointOffset := 0
				if course%2 == 1 { jointOffset = 4 }
				if (x+jointOffset)%8 == 0 {
					img.Set(x, y, mortarLeft)
				} else {
					brickSeed := (course*13 + (x+jointOffset)/8*7) % 5
					c := brickBaseRed
					if brickSeed == 1 || brickSeed == 3 {
						c = brickLightRed
					} else if brickSeed == 2 {
						c = brickDarkRed
					}
					c = darken(c, 0.90)
					if rng.Float64() < 0.15 {
						c = lighten(c, 1.08)
					}
					img.Set(x, y, c)
				}
			}
		}
	}

	// 3. Right Face (South Wall)
	for x := 32; x < 64; x++ {
		topY := 31 - (x-32)/2
		botY := 63 - (x-32)/2
		for y := topY + 1; y <= botY && y < h; y++ {
			hRel := y - topY
			course := hRel / 4
			courseY := hRel % 4

			if courseY == 0 {
				img.Set(x, y, mortarRight)
			} else {
				jointOffset := 0
				if course%2 == 1 { jointOffset = 4 }
				if ((x-32)+jointOffset)%8 == 0 {
					img.Set(x, y, mortarRight)
				} else {
					brickSeed := (course*17 + ((x-32)+jointOffset)/8*11) % 5
					c := brickBaseRed
					if brickSeed == 1 || brickSeed == 3 {
						c = brickLightRed
					} else if brickSeed == 2 {
						c = brickDarkRed
					}
					c = darken(c, 0.65)
					if rng.Float64() < 0.15 {
						c = lighten(c, 1.06)
					}
					img.Set(x, y, c)
				}
			}
		}
	}

	// Base contact shadow
	for x := 0; x < 32; x++ {
		botY := 47 + x/2
		if botY < h { img.Set(x, botY, color.RGBA{20, 15, 15, 255}) }
	}
	for x := 32; x < 64; x++ {
		botY := 63 - (x-32)/2
		if botY < h { img.Set(x, botY, color.RGBA{15, 10, 10, 255}) }
	}

	saveImg(name, img)
}

func generateIsoTree(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(707))

	barkDark := color.RGBA{58, 36, 20, 255}
	barkMid := color.RGBA{88, 56, 32, 255}
	barkLight := color.RGBA{118, 78, 46, 255}

	leafDeepShadow := color.RGBA{16, 42, 20, 255}
	leafShadow := color.RGBA{24, 66, 30, 255}
	leafMid := color.RGBA{38, 98, 44, 255}
	leafLight := color.RGBA{62, 142, 68, 255}
	leafHighlight := color.RGBA{95, 185, 88, 255}

	// 1. Trunk and Root Flares
	for y := 42; y <= 60; y++ {
		trunkW := 3
		if y >= 54 {
			trunkW = 3 + (y - 54)
		}
		for x := 32 - trunkW; x <= 32 + trunkW; x++ {
			if x >= 0 && x < w {
				dx := float64(x - 32)
				var c color.RGBA
				if dx < -1 {
					c = barkLight
				} else if dx > 1 {
					c = barkDark
				} else {
					c = barkMid
				}
				if rng.Float64() < 0.2 {
					c = darken(c, 0.85)
				}
				img.Set(x, y, c)
			}
		}
	}

	// 2. 3-Tier Layered Foliage Domes
	type Tier struct {
		topY, botY int
		maxR       float64
		cY         int
	}
	tiers := []Tier{
		{topY: 30, botY: 50, maxR: 22.0, cY: 42},
		{topY: 16, botY: 36, maxR: 17.0, cY: 28},
		{topY: 4, botY: 22, maxR: 12.0, cY: 14},
	}

	for _, tier := range tiers {
		for y := tier.topY; y <= tier.botY; y++ {
			progress := float64(y-tier.topY) / float64(tier.botY-tier.topY)
			radius := math.Sin(progress*math.Pi*0.85) * tier.maxR

			for x := int(32.0 - radius); x <= int(32.0 + radius); x++ {
				if x < 0 || x >= w { continue }
				dx := float64(x - 32)
				dy := float64(y - tier.cY)
				distNorm := math.Sqrt(dx*dx+dy*dy) / (radius + 0.1)

				lightFactor := (-dx/radius)*0.4 + (-dy/(float64(tier.botY-tier.topY)*0.5))*0.4

				var c color.RGBA
				if lightFactor > 0.4 {
					c = leafHighlight
				} else if lightFactor > 0.1 {
					c = leafLight
				} else if lightFactor > -0.2 {
					c = leafMid
				} else if lightFactor > -0.5 {
					c = leafShadow
				} else {
					c = leafDeepShadow
				}

				n := rng.Float64()
				if n < 0.18 {
					c = lighten(c, 1.15)
				} else if n > 0.85 {
					c = darken(c, 0.80)
				}

				if y >= tier.botY-2 {
					c = leafDeepShadow
				}

				if distNorm > 0.88 && rng.Float64() < 0.3 {
					continue
				}

				img.Set(x, y, c)
			}
		}
	}

	saveImg(name, img)
}

func generateIsoFence(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	woodLight := color.RGBA{168, 148, 125, 255}
	woodMid := color.RGBA{135, 115, 95, 255}
	woodDark := color.RGBA{85, 70, 55, 255}
	nailColor := color.RGBA{35, 30, 25, 255}

	// 1. Horizontal Rails
	for x := 2; x <= 32; x++ {
		tRailY := int(math.Round(28.0 + float64(x)*0.5))
		bRailY := int(math.Round(40.0 + float64(x)*0.5))
		for dy := 0; dy < 2; dy++ {
			img.Set(x, tRailY+dy, woodMid)
			img.Set(x, bRailY+dy, woodDark)
		}
	}

	// 2. Vertical Pickets
	picketPositions := []int{5, 9, 13, 17, 21, 25, 29}
	for _, px := range picketPositions {
		baseY := int(math.Round(46.0 + float64(px)*0.5))
		topY := baseY - 24

		img.Set(px+1, topY-2, woodLight)
		img.Set(px, topY-1, woodLight)
		img.Set(px+1, topY-1, woodMid)
		img.Set(px+2, topY-1, woodDark)

		for y := topY; y <= baseY; y++ {
			img.Set(px, y, woodLight)
			img.Set(px+1, y, woodMid)
			img.Set(px+2, y, woodDark)
		}

		tRailY := int(math.Round(28.0 + float64(px+1)*0.5))
		bRailY := int(math.Round(40.0 + float64(px+1)*0.5))
		img.Set(px+1, tRailY, nailColor)
		img.Set(px+1, bRailY, nailColor)
	}

	// 3. Main Corner Post
	for y := 28; y <= 60; y++ {
		img.Set(30, y, woodLight)
		img.Set(31, y, woodLight)
		img.Set(32, y, woodMid)
		img.Set(33, y, woodDark)
		img.Set(34, y, woodDark)
	}
	img.Set(32, 25, woodLight)
	img.Set(31, 26, woodLight)
	img.Set(32, 26, woodMid)
	img.Set(33, 26, woodDark)
	img.Set(30, 27, woodLight)
	img.Set(31, 27, woodLight)
	img.Set(32, 27, woodMid)
	img.Set(33, 27, woodDark)
	img.Set(34, 27, woodDark)

	// Left post
	for y := 16; y <= 48; y++ {
		img.Set(2, y, woodLight)
		img.Set(3, y, woodMid)
		img.Set(4, y, woodDark)
	}
	img.Set(3, 14, woodLight)
	img.Set(2, 15, woodLight)
	img.Set(3, 15, woodMid)
	img.Set(4, 15, woodDark)

	saveImg(name, img)
}

func generateIsoDebris(name string) {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(909))

	woodTop := color.RGBA{185, 135, 85, 255}
	woodLeft := color.RGBA{150, 105, 65, 255}
	woodRight := color.RGBA{95, 65, 38, 255}
	frameWood := color.RGBA{120, 80, 48, 255}
	metalBracket := color.RGBA{75, 80, 85, 255}

	concreteLight := color.RGBA{160, 160, 155, 255}
	concreteMid := color.RGBA{125, 125, 120, 255}
	concreteDark := color.RGBA{70, 70, 68, 255}
	brickChunk := color.RGBA{140, 55, 40, 255}
	shadowColor := color.RGBA{20, 20, 20, 150}

	// 1. Drop shadow
	for y := 46; y <= 60; y++ {
		for x := 12; x <= 52; x++ {
			dx := float64(x) - 32.0
			dy := float64(y) - 53.0
			if (dx*dx)/(20.0*20.0)+(dy*dy)/(7.0*7.0) <= 1.0 {
				img.Set(x, y, shadowColor)
			}
		}
	}

	// 2. 3D Wooden Crate Top Face
	for y := 19; y <= 33; y++ {
		for x := 18; x <= 46; x++ {
			dx := float64(x) - 32.0
			dy := float64(y) - 26.0
			if math.Abs(dx)/14.0+math.Abs(dy)/7.0 <= 1.0 {
				c := woodTop
				if math.Abs(dx)/14.0+math.Abs(dy)/7.0 > 0.80 {
					c = frameWood
				} else if rng.Float64() < 0.2 {
					c = lighten(c, 1.08)
				}
				img.Set(x, y, c)
			}
		}
	}

	// Left Face
	for x := 18; x <= 32; x++ {
		topY := 26 + (x-18)/2
		botY := 42 + (x-18)/2
		for y := topY; y <= botY; y++ {
			c := woodLeft
			relX := x - 18
			relY := y - topY
			if relX <= 1 || relX >= 13 || relY <= 1 || relY >= 15 {
				c = frameWood
			} else if relX == relY || relX == (16-relY) {
				c = frameWood
			}
			img.Set(x, y, c)
		}
	}

	// Right Face
	for x := 32; x <= 46; x++ {
		topY := 33 - (x-32)/2
		botY := 49 - (x-32)/2
		for y := topY; y <= botY; y++ {
			c := woodRight
			relX := x - 32
			relY := y - topY
			if relX <= 1 || relX >= 13 || relY <= 1 || relY >= 15 {
				c = darken(frameWood, 0.75)
			} else if relX == relY || relX == (16-relY) {
				c = darken(frameWood, 0.75)
			}
			img.Set(x, y, c)
		}
	}

	corners := [][2]int{
		{18, 26}, {32, 33}, {46, 26}, {32, 19},
		{18, 42}, {32, 49}, {46, 42},
	}
	for _, pt := range corners {
		img.Set(pt[0], pt[1], metalBracket)
		img.Set(pt[0]+1, pt[1], metalBracket)
		img.Set(pt[0], pt[1]+1, metalBracket)
	}

	// 3. Rubble Chunks
	drawRock := func(rx, ry, rw, rh int, baseC, highC, darkC color.RGBA) {
		for y := ry; y < ry+rh; y++ {
			for x := rx; x < rx+rw; x++ {
				dx := float64(x - rx - rw/2)
				dy := float64(y - ry - rh/2)
				if (dx*dx)/float64(rw*rw/4)+(dy*dy)/float64(rh*rh/4) <= 1.0 {
					var c color.RGBA
					if dy < -0.5 {
						c = highC
					} else if dy > 0.5 {
						c = darkC
					} else {
						c = baseC
					}
					img.Set(x, y, c)
				}
			}
		}
	}

	drawRock(8, 48, 8, 6, concreteMid, concreteLight, concreteDark)
	drawRock(46, 51, 9, 7, concreteMid, concreteLight, concreteDark)
	drawRock(14, 55, 6, 4, brickChunk, lighten(brickChunk, 1.2), darken(brickChunk, 0.7))
	drawRock(38, 54, 5, 4, brickChunk, lighten(brickChunk, 1.2), darken(brickChunk, 0.7))
	img.Set(24, 56, concreteLight)
	img.Set(25, 56, concreteDark)
	img.Set(44, 46, concreteLight)
	img.Set(45, 46, concreteDark)

	saveImg(name, img)
}
```

---

## 4. Caveats

1. **Draw Offset Alignment**:
   - In `internal/game/game.go:610-613`, ground tiles are rendered with `drawX = isoX - 32`, `drawY = isoY - 0`.
   - In `internal/game/game.go:658-669`, vertical blocks are rendered with `drawX = isoX - 32`, `drawY = isoY - 32`.
   - Any new obstacle tile (like `fence.png` and `debris.png`) must place its base diamond at $y \in [32, 64]$ to align identically with `wall.png` and `tree.png`.
2. **Deterministic Texture Generation**:
   - All procedural functions use explicit, fixed seed `rand.NewSource(...)` instances (e.g., seed 42 for grass, 101 for dirt, 202 for wood, 606 for wall). This ensures identical deterministic sprite output across builds.
3. **No External Asset Dependencies**:
   - Every sprite is generated strictly mathematically through `image.RGBA` pixel writes. No external image assets or font files are required.

---

## 5. Conclusion

- We have formulated and documented the exact procedural generation algorithms and complete Go implementations for all 6 floor tiles (`grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png`) and all 4 vertical obstacles (`wall.png`, `tree.png`, `fence.png`, `debris.png`).
- The algorithms respect pixel-perfect 2:1 isometric projection geometry, depth sorting, and directional shading.
- The implementer subagent can directly adopt these modular generator functions in `cmd/tools/genassets/main.go`.

---

## 6. Verification Method

1. **Verify Asset Generator Execution**:
   ```bash
   go run ./cmd/tools/genassets
   ```
   Must output without errors and generate all 10 environment PNGs in `internal/assets/images/`.

2. **Verify Image Dimensions**:
   Inspect generated PNG headers or use Go image decode:
   - Floor diamonds (`grass.png`, `dirt.png`, `wood.png`, `asphalt.png`, `concrete.png`, `tile_floor.png`): exactly $64 \times 32$.
   - Vertical obstacles (`wall.png`, `tree.png`, `fence.png`, `debris.png`): exactly $64 \times 64$.

3. **Verify Engine Compilation and Tests**:
   ```bash
   CC=gcc go test ./...
   ```
   Must pass all tests cleanly.
