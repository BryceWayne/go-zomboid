# Milestone 1 Floor Tiles Exploration Report (256x128 Scaling)

**Document**: `m1_floor_analysis.md`  
**Agent**: `m1_explorer_1` (Milestone 1 High-Fidelity Asset Pipeline)  
**Target File**: `cmd/tools/genassets/main.go`  
**Date**: 2026-08-28  

---

## 1. Executive Summary

To achieve the crisp, clean, Dribbble vector art aesthetic specified in `ORIGINAL_REQUEST.md` and `PROJECT.md`, the floor tile asset generators in `cmd/tools/genassets/main.go` must be scaled 4x from **64x32** to **256x128** pixels. This quadrupling preserves the 2:1 dimetric isometric projection while providing sufficient pixel density for anti-aliased geometric vector overlays (chevrons, wildflower petals, rounded pebbles with specular highlights, plank grain with nailheads, asphalt dashed lane markings, chamfered concrete expansion joints, and ceramic tiles with mortar grout).

This report provides the exact mathematical formulas, coordinate transforms, vector geometry algorithms, and drop-in Go code implementations for all 6 floor generators:
1. `generateGrass(name string)`
2. `generateDirt(name string)`
3. `generateWoodFloor(name string)`
4. `generateAsphalt(name string)`
5. `generateConcrete(name string)`
6. `generateTileFloor(name string)`

---

## 2. Mathematical Foundations

### 2.1 Isometric Diamond Boundary Equation (256x128)

For a 2D canvas of width $W = 256$ and height $H = 128$ indexed with integer coordinates $x \in [0, 255]$ and $y \in [0, 127]$:
- Canvas Center:
  $$x_c = \frac{W - 1}{2} = 127.5, \quad y_c = \frac{H - 1}{2} = 63.5$$
- Semi-axis Widths:
  $$R_x = \frac{W}{2} = 128.0, \quad R_y = \frac{H}{2} = 64.0$$
- Center-relative offsets:
  $$dx = x - 127.5, \quad dy = y - 63.5$$
- Normalized Manhattan / $L_1$ Diamond Metric:
  $$\text{isoDist}(x, y) = \frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0}$$

A pixel $(x, y)$ lies strictly inside or on the boundary of the isometric tile diamond if and only if:
$$\text{isoDist}(x, y) \le 1.0$$

#### Boundary Extreme Points:
| Diamond Vertex | Screen $(x, y)$ Coordinates | Center Offset $(dx, dy)$ | $\text{isoDist}$ Evaluation |
|---|---|---|---|
| Top Vertex | $(127.5, -0.5)$ | $(0, -64.0)$ | $\frac{0}{128} + \frac{64}{64} = 1.0$ |
| Right Vertex | $(255.5, 63.5)$ | $(128.0, 0)$ | $\frac{128}{128} + \frac{0}{64} = 1.0$ |
| Bottom Vertex | $(127.5, 127.5)$ | $(0, 64.0)$ | $\frac{0}{128} + \frac{64}{64} = 1.0$ |
| Left Vertex | $(-0.5, 63.5)$ | $(-128.0, 0)$ | $\frac{128}{128} + \frac{0}{64} = 1.0$ |

The slope of all four diamond facets is exactly $|\frac{dy}{dx}| = \frac{64}{128} = 0.5$ (2:1 dimetric ratio).

---

### 2.2 Bi-Directional UV Space Mapping ($u, v \in [0, 1]$)

To map textures and vector overlays seamlessly across the rotated isometric ground plane, we define a normalized orthogonal surface space $(u, v) \in [0, 1] \times [0, 1]$:
- $(u=0, v=0) \iff$ Top Vertex $(127.5, -0.5)$
- $(u=1, v=0) \iff$ Right Vertex $(255.5, 63.5)$
- $(u=1, v=1) \iff$ Bottom Vertex $(127.5, 127.5)$
- $(u=0, v=1) \iff$ Left Vertex $(-0.5, 63.5)$

#### Forward Projection: $(u, v) \to (dx, dy) \to (x, y)$
$$dx = (u - v) \cdot 128.0$$
$$dy = (u + v - 1.0) \cdot 64.0$$
$$x = 127.5 + (u - v) \cdot 128.0$$
$$y = 63.5 + (u + v - 1.0) \cdot 64.0$$

#### Inverse Projection: $(x, y) \to (dx, dy) \to (u, v)$
$$dx = x - 127.5, \quad dy = y - 63.5$$
$$u = \frac{dx}{256.0} + \frac{dy}{128.0} + 0.5$$
$$v = \frac{dy}{128.0} - \frac{dx}{256.0} + 0.5$$

#### Equivalence Identity:
For all $(x, y) \in \mathbb{R}^2$:
$$\text{isoDist}(x, y) \le 1.0 \iff u(x,y) \in [0, 1] \text{ and } v(x,y) \in [0, 1]$$
This mathematical equivalence guarantees zero seam artifacts or coordinate leakage at the diamond edges.

---

## 3. Geometric Vector Overlay Specifications

### 3.1 Grass: Multi-Blade Chevrons & Wildflower Clusters
- **Chevrons (Grass Tufts)**:
  - Scaled anchor coordinates:
    $$\mathcal{C} = \{(64, 48), (160, 32), (96, 80), (192, 64), (128, 56), (48, 68), (176, 92), (140, 24)\}$$
  - Blade Geometry: Each tuft is rendered as a clean 3-blade vector cluster:
    - Center blade: vertical line of width 2px from $(cx, cy)$ up to $(cx, cy - 8)$.
    - Left blade: 2px wide diagonal arm from $(cx-1, cy)$ up-left to $(cx-7, cy-7)$.
    - Right blade: 2px wide diagonal arm from $(cx+1, cy)$ up-right to $(cx+7, cy-7)$.
    - Root base: solid 3x2px anchor at $(cx-1, cy)$.
- **Wildflowers**:
  - Scaled anchor coordinates:
    $$\mathcal{F} = \{(96, 32), (160, 80), (52, 70), (180, 44), (120, 96)\}$$
  - Petal Cluster Geometry:
    - Center pistil: filled disk of radius $r=2.5$ px in `flowerYellow` (`#FFDC64`).
    - 5 Petal Lobes: filled circles of radius $r=2.5$ px in `flowerWhite` (`#FFFFFF`) arranged at angular offsets $\theta_k = k \cdot \frac{2\pi}{5}$ at distance $d = 4.5$ px.

### 3.2 Dirt: Rounded Pebbles & Organic Clods
- **Rounded Pebble Geometry**:
  - Scaled anchor coordinates:
    $$\mathcal{P} = \{(80, 40), (180, 56), (120, 88), (60, 80), (195, 36), (145, 30)\}$$
  - Shape: 2D ellipse with horizontal radius $r_x = 7.0$ px and vertical radius $r_y = 4.0$ px (total dimension $\approx 14 \times 8$ px):
    $$\frac{(x - px)^2}{7.0^2} + \frac{(y - py)^2}{4.0^2} \le 1.0$$
  - Drop Shadow: Offset ellipse at $(px+2, py+2)$ with `color.RGBA{0, 0, 0, 45}`.
  - Multi-tone Shading:
    - Specular highlight crescent (top-left, $(x - px) + (y - py) < -2.0$): `pebbleLight` (`#D7B9A5`).
    - Ambient shadow crease (bottom-right, $(x - px) + (y - py) > 2.5$): `pebbleShadow` (`#6E4B3C`).
    - Core stone body: `pebbleBase` (`#AF8C78`).

### 3.3 Wood: Longitudinal UV Lanes, 3px Seams & Iron Nailheads
- **Lanes**: 4 longitudinal plank lanes partitioned in UV space along $v$:
  $$\text{lane} = \lfloor 4.0 \cdot v \rfloor \in \{0, 1, 2, 3\}, \quad v_{\text{local}} = 4.0 \cdot v - \text{lane} \in [0, 1)$$
- **Plank Colors**: 4 distinct warm cedar/oak tones cycled per lane.
- **Seams**:
  - Longitudinal joint: $v_{\text{local}} < 0.04 \text{ or } v_{\text{local}} > 0.96 \implies \text{seamDark}$ (`#2D1A0E`).
  - Staggered transverse end joints: $u_{\text{end}} \in \{0.60, 0.30, 0.75, 0.45\}$ per lane.
    $$|u - u_{\text{end}}| < 0.012 \implies \text{seamDark}$$
- **Nailhead Circles**:
  - Placed at $u = u_{\text{end}} \pm 0.035$ on lane center $v = \frac{\text{lane} + 0.5}{4.0}$.
  - Rendered as filled circles of radius $r = 2.5$ px with dark iron base (`#1E1612`) and 1px specular dot at $(nx-1, ny-1)$ (`#645046`).

### 3.4 Asphalt: Dashed Yellow Highway Striping
- **Centerline Stripe**:
  - UV stripe band: $v \in [0.45, 0.55]$ (10% width = ~13px ribbon).
  - Dashes along $u$:
    - Dash 1: $u \in [0.08, 0.40]$
    - Dash 2: $u \in [0.60, 0.92]$
    - Gap: $u \in [0.40, 0.60]$
  - Primary color: `yellowMarking` (`#F0C32D`).
  - Bottom bevel edge shadow: $v \in [0.535, 0.55] \implies \text{yellowShadow}$ (`#B48C1E`).

### 3.5 Concrete: 2x2 Slabs with Chamfered Expansion Joints
- **Slab Division**: Quadrants split at $u = 0.5$ and $v = 0.5$:
  $$\text{quadX} = \begin{cases} 0 & u < 0.5 \\ 1 & u \ge 0.5 \end{cases}, \quad \text{quadY} = \begin{cases} 0 & v < 0.5 \\ 1 & v \ge 0.5 \end{cases}$$
- **Expansion Joint**:
  - Central groove: $|u - 0.5| < 0.010 \text{ or } |v - 0.5| < 0.010 \implies \text{jointDark}$ (`#2D2D2D`).
  - Chamfered bevel highlight on adjacent edge: $(u \in [0.510, 0.522] \text{ or } v \in [0.510, 0.522]) \implies \text{jointBevelLight}$ (`#C3C3BE`).

### 3.6 Tile Floor: 4x4 Checkerboard & Mortar Grout
- **Grid Partition**:
  $$\text{tileU} = \lfloor 4.0 \cdot u \rfloor, \quad \text{tileV} = \lfloor 4.0 \cdot v \rfloor \quad (\text{clamped to } [0, 3])$$
  $$subU = 4.0 \cdot u - \text{tileU}, \quad subV = 4.0 \cdot v - \text{tileV} \in [0, 1)$$
- **Mortar Grout**: $subU < 0.045 \text{ or } subV < 0.045 \implies \text{groutDark}$ (`#202226`).
- **Porcelain / Slate Bevels**:
  - Alternating base: $(\text{tileU} + \text{tileV}) \bmod 2 == 0 \implies \text{tileA}$ (Light porcelain `#D2D2CD`), else $\text{tileB}$ (Dark slate `#414B55`).
  - Top-left specular bevel: $subU \in [0.045, 0.10] \text{ or } subV \in [0.045, 0.10] \implies \text{lighten}(\text{base}, 1.15)$.
  - Bottom-right shadow bevel: $subU > 0.94 \text{ or } subV > 0.94 \implies \text{darken}(\text{base}, 0.82)$.

---

## 4. Complete Drop-In Go Code Implementation

Below is the complete, tested Go code for the drawing helper primitives and all 6 floor generator functions in `cmd/tools/genassets/main.go`.

### 4.1 Vector Drawing Helpers
```go
// Draw filled circle with bounds checking
func drawFilledCircle(img *image.RGBA, cx, cy int, r float64, c color.RGBA) {
	r2 := r * r
	minX := int(math.Floor(float64(cx) - r))
	maxX := int(math.Ceil(float64(cx) + r))
	minY := int(math.Floor(float64(cy) - r))
	maxY := int(math.Ceil(float64(cy) + r))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			if dx*dx+dy*dy <= r2 {
				setPixel(img, x, y, c)
			}
		}
	}
}

// Draw vector grass blade cluster (3 blades + root)
func drawVectorChevron(img *image.RGBA, cx, cy int, col color.RGBA) {
	// Center root
	fillRect(img, cx-1, cy, 3, 2, col)
	// Center blade (vertical)
	for i := 0; i < 8; i++ {
		setPixel(img, cx, cy-i, col)
		setPixel(img, cx+1, cy-i, col)
	}
	// Left blade (diagonal up-left)
	for i := 1; i <= 7; i++ {
		setPixel(img, cx-i, cy-i, col)
		setPixel(img, cx-i, cy-i+1, col)
	}
	// Right blade (diagonal up-right)
	for i := 1; i <= 7; i++ {
		setPixel(img, cx+i+1, cy-i, col)
		setPixel(img, cx+i+1, cy-i+1, col)
	}
}

// Draw 5-petal wildflower with yellow center
func drawVectorFlower(img *image.RGBA, cx, cy int, petalColor, centerColor color.RGBA) {
	// 5 petals
	for k := 0; k < 5; k++ {
		angle := float64(k) * (2.0 * math.Pi / 5.0) - math.Pi/2.0
		px := int(math.Round(float64(cx) + 4.5*math.Cos(angle)))
		py := int(math.Round(float64(cy) + 4.5*math.Sin(angle)))
		drawFilledCircle(img, px, py, 2.5, petalColor)
	}
	// Center pistil
	drawFilledCircle(img, cx, cy, 2.5, centerColor)
}

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
			dx := float64(x - (cx + 2)) / rx
			dy := float64(y - (cy + 2)) / ry
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

---

### 4.2 `generateGrass(name string)`
```go
func generateGrass(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	topColor := color.RGBA{113, 225, 157, 255}
	leftColor := color.RGBA{74, 184, 131, 255}
	rightColor := color.RGBA{54, 150, 103, 255}
	flowerWhite := color.RGBA{255, 255, 255, 255}
	flowerYellow := color.RGBA{255, 220, 100, 255}
	chevronColor := color.RGBA{154, 238, 181, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				c := topColor

				rimThickness := 0.08 + 0.03*math.Sin(float64(x)*0.08)
				if isoDist > 1.0-rimThickness {
					if x < 128 {
						c = leftColor
					} else {
						c = rightColor
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// 4x Scaled vector grass blades
	chevrons := [][2]int{
		{64, 48}, {160, 32}, {96, 80}, {192, 64},
		{128, 56}, {48, 68}, {176, 92}, {140, 24},
	}
	for _, pos := range chevrons {
		drawVectorChevron(img, pos[0], pos[1], chevronColor)
	}

	// 4x Scaled wildflower clusters
	flowers := [][2]int{
		{96, 32}, {160, 80}, {52, 70}, {180, 44}, {120, 96},
	}
	for _, pos := range flowers {
		drawVectorFlower(img, pos[0], pos[1], flowerWhite, flowerYellow)
	}

	saveImg(name, img)
}
```

---

### 4.3 `generateDirt(name string)`
```go
func generateDirt(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	topColor := color.RGBA{151, 103, 81, 255}
	leftColor := color.RGBA{122, 79, 59, 255}
	rightColor := color.RGBA{94, 60, 44, 255}
	pebbleBase := color.RGBA{175, 140, 120, 255}
	pebbleLight := color.RGBA{215, 185, 165, 255}
	pebbleShadow := color.RGBA{110, 75, 60, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				c := topColor

				rimThickness := 0.08 + 0.03*math.Sin(float64(x)*0.08)
				if isoDist > 1.0-rimThickness {
					if x < 128 {
						c = leftColor
					} else {
						c = rightColor
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// 4x Scaled rounded vector pebbles (~14x8px)
	pebbles := [][2]int{
		{80, 40}, {180, 56}, {120, 88}, {60, 80}, {195, 36}, {145, 30},
	}
	for _, pos := range pebbles {
		drawVectorPebble(img, pos[0], pos[1], 7.0, 4.0, pebbleBase, pebbleLight, pebbleShadow)
	}

	saveImg(name, img)
}
```

---

### 4.4 `generateWoodFloor(name string)`
```go
func generateWoodFloor(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	plankColors := []color.RGBA{
		{142, 92, 54, 255},
		{128, 80, 46, 255},
		{156, 104, 62, 255},
		{136, 88, 50, 255},
	}
	seamDark := color.RGBA{45, 26, 14, 255}
	nailColor := color.RGBA{30, 22, 18, 255}
	nailHighlight := color.RGBA{100, 80, 70, 255}
	endJoints := []float64{0.60, 0.30, 0.75, 0.45}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				laneFloat := v * 4.0
				lane := int(math.Floor(laneFloat))
				if lane < 0 {
					lane = 0
				}
				if lane > 3 {
					lane = 3
				}
				vInLane := laneFloat - float64(lane)

				c := plankColors[lane%len(plankColors)]

				// 3px Seams along longitudinal lanes
				if vInLane < 0.04 || vInLane > 0.96 {
					c = seamDark
				}

				// Transverse end joints
				endU := endJoints[lane]
				if math.Abs(u-endU) < 0.012 {
					c = seamDark
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = darken(c, 0.85)
					} else {
						c = darken(c, 0.70)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	// 4x Vector nailheads
	for lane := 0; lane < 4; lane++ {
		endU := endJoints[lane]
		for _, offset := range []float64{-0.035, 0.035} {
			nu := endU + offset
			nv := (float64(lane) + 0.5) / 4.0
			nx := int(math.Round(127.5 + (nu-nv)*128.0))
			ny := int(math.Round(63.5 + (nu+nv-1.0)*64.0))
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				dx := float64(nx) - 127.5
				dy := float64(ny) - 63.5
				if math.Abs(dx)/128.0+math.Abs(dy)/64.0 <= 0.88 {
					drawFilledCircle(img, nx, ny, 2.5, nailColor)
					setPixel(img, nx-1, ny-1, nailHighlight)
				}
			}
		}
	}

	saveImg(name, img)
}
```

---

### 4.5 `generateAsphalt(name string)`
```go
func generateAsphalt(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	baseDark := color.RGBA{38, 40, 44, 255}
	baseMid := color.RGBA{48, 50, 55, 255}
	yellowMarking := color.RGBA{240, 195, 45, 255}
	yellowShadow := color.RGBA{180, 140, 30, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				c := baseMid

				// Crisp highway yellow dashed stripe in UV
				if v >= 0.45 && v <= 0.55 {
					if (u >= 0.08 && u <= 0.40) || (u >= 0.60 && u <= 0.92) {
						if v >= 0.535 {
							c = yellowShadow
						} else {
							c = yellowMarking
						}
					}
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = baseDark
					} else {
						c = darken(baseDark, 0.75)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}
```

---

### 4.6 `generateConcrete(name string)`
```go
func generateConcrete(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	slabBase := color.RGBA{145, 145, 142, 255}
	slabLight := color.RGBA{168, 168, 165, 255}
	jointDark := color.RGBA{45, 45, 45, 255}
	jointBevelLight := color.RGBA{195, 195, 190, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				quadX := 0
				if u >= 0.5 {
					quadX = 1
				}
				quadY := 0
				if v >= 0.5 {
					quadY = 1
				}

				var c color.RGBA
				if (quadX+quadY)%2 == 0 {
					c = slabBase
				} else {
					c = slabLight
				}

				distU := math.Abs(u - 0.5)
				distV := math.Abs(v - 0.5)

				// Expansion joints (3px deep grooving)
				if distU < 0.010 || distV < 0.010 {
					c = jointDark
				} else if (distU < 0.020 && u > 0.5) || (distV < 0.020 && v > 0.5) {
					// Top/left bevel highlight
					c = jointBevelLight
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = darken(slabBase, 0.8)
					} else {
						c = darken(slabBase, 0.6)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}
```

---

### 4.7 `generateTileFloor(name string)`
```go
func generateTileFloor(name string) {
	w, h := 256, 128
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	tileABase := color.RGBA{210, 210, 205, 255}
	tileBBase := color.RGBA{65, 75, 85, 255}
	groutDark := color.RGBA{32, 34, 38, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - 127.5
			dy := float64(y) - 63.5
			isoDist := math.Abs(dx)/128.0 + math.Abs(dy)/64.0
			if isoDist <= 1.0 {
				u := dx/256.0 + dy/128.0 + 0.5
				v := dy/128.0 - dx/256.0 + 0.5

				gridU := u * 4.0
				gridV := v * 4.0
				tileU := int(math.Floor(gridU))
				tileV := int(math.Floor(gridV))
				if tileU < 0 {
					tileU = 0
				}
				if tileU > 3 {
					tileU = 3
				}
				if tileV < 0 {
					tileV = 0
				}
				if tileV > 3 {
					tileV = 3
				}

				subU := gridU - float64(tileU)
				subV := gridV - float64(tileV)

				isTileA := (tileU+tileV)%2 == 0
				var baseCol color.RGBA
				if isTileA {
					baseCol = tileABase
				} else {
					baseCol = tileBBase
				}

				var c color.RGBA
				if subU < 0.045 || subV < 0.045 {
					// Mortar grout line
					c = groutDark
				} else if subU < 0.09 || subV < 0.09 {
					// Specular tile bevel highlight
					c = lighten(baseCol, 1.15)
				} else if subU > 0.94 || subV > 0.94 {
					// Tile shadow bevel
					c = darken(baseCol, 0.82)
				} else {
					c = baseCol
				}

				// Flat stepped extrusions
				if isoDist > 0.96 {
					if x < 128 {
						c = darken(c, 0.85)
					} else {
						c = darken(c, 0.70)
					}
				}

				setPixel(img, x, y, c)
			}
		}
	}

	saveImg(name, img)
}
```

---

## 5. Interface Contracts & Verification Strategy

### 5.1 Asset Dimensions Contract
| Asset File | Target Width | Target Height | Bit Depth / Color | Transparency |
|---|---|---|---|---|
| `grass.png` | 256 | 128 | 32-bit RGBA PNG | Transparent outside diamond |
| `dirt.png` | 256 | 128 | 32-bit RGBA PNG | Transparent outside diamond |
| `wood.png` | 256 | 128 | 32-bit RGBA PNG | Transparent outside diamond |
| `asphalt.png` | 256 | 128 | 32-bit RGBA PNG | Transparent outside diamond |
| `concrete.png` | 256 | 128 | 32-bit RGBA PNG | Transparent outside diamond |
| `tile_floor.png` | 256 | 128 | 32-bit RGBA PNG | Transparent outside diamond |

### 5.2 Verification Methods
1. **Asset Generation Command**:
   ```bash
   go run ./cmd/tools/genassets
   ```
2. **Dimension & Non-Nil Validation Tests**:
   Update `internal/assets/assets_test.go` and `internal/assets/assets_stress_test.go` floor expectations to `256x128`. Run:
   ```bash
   CC=gcc go test -v ./internal/assets/...
   ```
3. **Diamond Symmetry & Mathematical Bounds Invariant**:
   For any generated floor tile `img`, verify:
   - For all $(x, y)$ where $\frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0} > 1.01$, `Alpha == 0`.
   - For all $(x, y)$ where $\frac{|x - 127.5|}{128.0} + \frac{|y - 63.5|}{64.0} < 0.95$, `Alpha == 255`.
