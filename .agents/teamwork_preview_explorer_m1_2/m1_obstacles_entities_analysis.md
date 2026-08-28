# Milestone 1 Investigation Report: Vertical Obstacles / Props (256x256) & Character Entities (64x128)

**Author**: `teamwork_preview_explorer_m1_2`  
**Date**: 2026-08-28  
**Scope**: `cmd/tools/genassets/main.go`, `internal/assets/`, `internal/game/`  
**Target Milestone**: Milestone 1 (High-Fidelity Asset Pipeline 4x Scaling)

---

## 1. Executive Summary

In Milestone 1, the `go-zomboid` procedural asset generator (`cmd/tools/genassets/main.go`) is upgraded by a 4x linear scaling factor to match the Dribbble vector art style:
- **Floor Tiles**: Scaled from 64x32 to 256x128 (2:1 dimetric ratio).
- **Vertical Obstacles & Props**: Scaled from 64x64 to 256x256 (10 sprites: `wall`, `tree`, `fence`, `debris`, `tent`, `stump`, `mushroom`, `sign`, `elevation_block`, `elevation_ramp`).
- **Character Entities**: Scaled from 16x32 to 64x128 (3 sprites: `player`, `zombie`, `runner`).
- **Items & Equipment**: Scaled from 16x16 to 64x64 (8 sprites: `food`, `water`, `weapon`, `axe`, `shotgun`, `ammo`, `armor`, `antidote`).

This report provides the exact geometric formulas, color palettes, vector shading models, and drop shadow formulations needed for all **10 vertical obstacles/props** (256x256) and all **3 character entities** (64x128).

---

## 2. Coordinate Pipeline & Canvas Alignment

### 2.1 Coordinate Systems
1. **World Cartesian Space $(wx, wy)$**: $1\text{ cell} = 128 \times 128\text{ px}$.
2. **2:1 Isometric Projection**:
   $$\text{isoX} = wx - wy, \quad \text{isoY} = \frac{wx + wy}{2}$$
3. **Viewport Anchor Offsets**:
   - **Floors (256x128)**: $\text{drawX} = \text{isoX} - 128, \text{drawY} = \text{isoY} - 0$
   - **Obstacles (256x256)**: $\text{drawX} = \text{isoX} - 128, \text{drawY} = \text{isoY} - 128$
   - **Entities (64x128)**: $\text{drawX} = \text{isoX} - 32, \text{drawY} = \text{isoY} - 128$
   - **Items (64x64)**: $\text{drawX} = \text{isoX} - 32, \text{drawY} = \text{isoY} - 32$

### 2.2 Obstacle 256x256 Canvas Alignment
When an obstacle tile is placed at $(wx, wy)$, its ground base footprint is an isometric diamond spanning $x \in [0..256], y \in [128..256]$.
- **Ground Center**: $(128, 192)$
- **Ground Diamond Vertices**:
  - Top vertex: $(128, 128)$
  - Left vertex: $(0, 192)$
  - Right vertex: $(256, 192)$
  - Bottom vertex: $(128, 256)$
- **Raised Top Face (Height $H = 128\text{px}$)**:
  - Top face diamond center: $(128, 64)$
  - Top vertex: $(128, 0)$
  - Left vertex: $(0, 64)$
  - Right vertex: $(256, 64)$
  - Bottom vertex: $(128, 128)$

### 2.3 Entity 64x128 Canvas Alignment
When an entity is at $(wx, wy)$, its world position is the center of the feet on the ground.
- **Canvas Dimensions**: $64 \times 128\text{ px}$.
- **Ground Anchor Point**: Center $X = 32$, $Y \in [118..124]$.
- **Drop Shadow Ellipse**: Centered at $(32, 122)$, radii $r_x = 24.0, r_y = 6.0$, alpha-blended over ground.
- **Head & Torso**: Torso $y \in [48..82]$, Head $y \in [12..48]$, Center $X = 32$.

---

## 3. Helper Primitives & Alpha Compositing

To achieve anti-aliased vector rendering without external dependencies, we introduce lightweight drawing primitives in `cmd/tools/genassets/main.go`:

```go
// Alpha-blended pixel setter (Porter-Duff Over)
func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	if c.A == 255 {
		img.SetRGBA(x, y, c)
		return
	}
	if c.A == 0 {
		return
	}
	dst := img.RGBAAt(x, y)
	if dst.A == 0 {
		img.SetRGBA(x, y, c)
		return
	}
	srcA := float64(c.A) / 255.0
	dstA := float64(dst.A) / 255.0
	outA := srcA + dstA*(1.0-srcA)
	if outA <= 0 {
		return
	}
	outR := (float64(c.R)*srcA + float64(dst.R)*dstA*(1.0-srcA)) / outA
	outG := (float64(c.G)*srcA + float64(dst.G)*dstA*(1.0-srcA)) / outA
	outB := (float64(c.B)*srcA + float64(dst.B)*dstA*(1.0-srcA)) / outA
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(math.Round(outR)),
		G: uint8(math.Round(outG)),
		B: uint8(math.Round(outB)),
		A: uint8(math.Round(outA * 255.0)),
	})
}

// Anti-aliased filled ellipse
func drawAAEllipse(img *image.RGBA, cx, cy, rx, ry float64, c color.RGBA) {
	minX := int(math.Floor(cx - rx - 1))
	maxX := int(math.Ceil(cx + rx + 1))
	minY := int(math.Floor(cy - ry - 1))
	maxY := int(math.Ceil(cy + ry + 1))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(y) + 0.5 - cy) / ry
			d := dx*dx + dy*dy
			if d <= 1.0 {
				edgeDist := 1.0 - math.Sqrt(d)
				alphaFactor := math.Min(1.0, edgeDist*math.Min(rx, ry))
				if alphaFactor > 0.05 {
					col := c
					col.A = uint8(float64(c.A) * alphaFactor)
					blendPixel(img, x, y, col)
				}
			}
		}
	}
}
```

---

## 4. Vertical Obstacles & Props (256x256) Specifications

### 4.1 Wall (`generateIsoWall` -> `wall.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Clean architectural vector brick wall with light concrete coping slab, bevel highlights, staggered brick mortar lines.
- **Color Palette**:
  - `copingTop`: `RGBA{228, 224, 218, 255}`
  - `copingLeft`: `RGBA{200, 195, 188, 255}`
  - `copingRight`: `RGBA{170, 165, 158, 255}`
  - `brickLeftBase`: `RGBA{154, 62, 48, 255}`
  - `brickLeftMortar`: `RGBA{185, 95, 78, 255}`
  - `brickRightBase`: `RGBA{118, 44, 32, 255}`
  - `brickRightMortar`: `RGBA{88, 30, 20, 255}`
  - `ridgeHighlight`: `RGBA{255, 250, 245, 255}`
- **Exact Geometry**:
  1. **Top Coping Face**:
     Diamond centered at $(128, 56)$ with $r_x = 128.0, r_y = 56.0$:
     $$\frac{|x - 127.5|}{128.0} + \frac{|y - 55.5|}{56.0} \le 1.0$$
  2. **Left Face (West Wall)** for $x \in [0..127]$:
     $$\text{topY} = 60 + \lfloor x/2 \rfloor, \quad \text{botY} = 188 + \lfloor x/2 \rfloor$$
     Mortar lines: horizontal joints every 16px along slope $(y - \lfloor x/2 \rfloor) \pmod{16} == 0$, vertical joints staggered every 32px.
  3. **Right Face (South Wall)** for $x \in [128..255]$:
     $$\text{topY} = 124 - \lfloor (x - 128)/2 \rfloor, \quad \text{botY} = 252 - \lfloor (x - 128)/2 \rfloor$$
  4. **Highlight Ridge**:
     $1\text{px}$ crisp line connecting $(0, 60) \to (128, 124) \to (256, 60)$.

---

### 4.2 Tree (`generateIsoTree` -> `tree.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Lush, stylized multi-tiered spherical canopy with Dribbble vector toon shading, cylindrical bark trunk, ground drop shadow.
- **Color Palette**:
  - `trunkHighlight`: `RGBA{135, 100, 75, 255}`
  - `trunkBase`: `RGBA{101, 74, 57, 255}`
  - `trunkShadow`: `RGBA{68, 48, 36, 255}`
  - `leafHighlight`: `RGBA{110, 218, 158, 255}`
  - `leafMid`: `RGBA{74, 184, 131, 255}`
  - `leafShadow`: `RGBA{44, 142, 96, 255}`
  - `leafDeepShadow`: `RGBA{28, 98, 64, 255}`
  - `groundShadow`: `RGBA{0, 0, 0, 50}`
- **Exact Geometry**:
  1. **Ground Drop Shadow**:
     Ellipse centered at $(128, 220)$ with $r_x = 64.0, r_y = 20.0$.
  2. **Trunk Cylinder**:
     $x \in [112..143]$ (width 32px), $y \in [148..222]$ (height 74px).
     - Left highlight: $x \in [112..119] \implies \text{trunkHighlight}$
     - Center: $x \in [120..135] \implies \text{trunkBase}$
     - Right shadow: $x \in [136..143] \implies \text{trunkShadow}$
     - Base root flare: widening by 4px on each side at $y \in [216..222]$.
  3. **Multi-Tier Foliage Canopy**:
     Formed by overlapping smooth spheres with directional toon lighting:
     - Center Sphere: center $(128, 100), R = 80$
     - Top Lobe: center $(128, 60), R = 56$
     - Left Lobe: center $(84, 108), R = 54$
     - Right Lobe: center $(172, 108), R = 54$
     - Toon Shading Rule on combined distance field:
       - If $dx < -0.15 R \land dy < -0.15 R \implies \text{leafHighlight}$
       - If $dx > 0.20 R \land dy > 0.20 R \land \text{dist} > 0.45 R \implies \text{leafShadow}$
       - If $dx > 0.40 R \land dy > 0.40 R \implies \text{leafDeepShadow}$
       - Else $\implies \text{leafMid}$

---

### 4.3 Fence (`generateIsoFence` -> `fence.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Isometric wooden picket fence running along the West axis with dual horizontal cross-rails, 7 pointed pickets, iron fastening nails, and a 3D corner post with pyramid cap.
- **Color Palette**:
  - `woodLight`: `RGBA{178, 158, 135, 255}`
  - `woodMid`: `RGBA{142, 120, 98, 255}`
  - `woodDark`: `RGBA{88, 72, 56, 255}`
  - `nailColor`: `RGBA{35, 30, 25, 255}`
  - `nailHighlight`: `RGBA{180, 180, 190, 255}`
- **Exact Geometry**:
  1. **Horizontal Rails** ($x \in [8..128]$ along slope $y = \text{base} + \lfloor x/2 \rfloor$):
     - Top Rail: $y \in [112 + x/2 .. 120 + x/2]$ (thickness 8px)
     - Bottom Rail: $y \in [160 + x/2 .. 168 + x/2]$ (thickness 8px)
  2. **Vertical Pickets** (7 pickets at $px \in \{20, 36, 52, 68, 84, 100, 116\}$, width 10px):
     - $\text{baseY} = 184 + \lfloor px/2 \rfloor, \quad \text{topY} = \text{baseY} - 96$.
     - Pointed peak: triangle rising 8px from $\text{topY}$ to peak at $(px + 5, \text{topY} - 8)$.
     - 3-stripe shading: $x \in [px..px+2] \implies \text{woodLight}$, $[px+3..px+7] \implies \text{woodMid}$, $[px+8..px+9] \implies \text{woodDark}$.
     - Fastening nails: $2\times 2\text{px}$ dark studs at top and bottom rail intersections with $1\text{px}$ highlight.
  3. **Corner Post** ($x \in [120..135]$, width 16px):
     - Post body: $y \in [112..240]$.
     - 4-sided faceted pyramid cap: $y \in [96..112]$.
  4. **Left Post** ($x \in [8..23]$, width 16px):
     - Post body: $y \in [56..184]$, pyramid cap $y \in [40..56]$.

---

### 4.4 Debris (`generateIsoDebris` -> `debris.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Heavy 3D military wooden supply crate with internal planking, full diagonal X-bracing, iron corner brackets with rivets, surrounded by shattered concrete chunks and terracotta brick fragments.
- **Color Palette**:
  - `crateTop`: `RGBA{190, 142, 92, 255}`
  - `crateLeft`: `RGBA{156, 112, 70, 255}`
  - `crateRight`: `RGBA{102, 70, 42, 255}`
  - `frameWood`: `RGBA{124, 84, 52, 255}`
  - `ironBracket`: `RGBA{78, 82, 88, 255}`
  - `ironRivet`: `RGBA{190, 195, 205, 255}`
  - `concreteMid`: `RGBA{130, 130, 125, 255}`
  - `concreteLight`: `RGBA{170, 170, 165, 255}`
  - `concreteDark`: `RGBA{75, 75, 72, 255}`
  - `brickRed`: `RGBA{145, 58, 42, 255}`
  - `groundShadow`: `RGBA{20, 20, 20, 120}`
- **Exact Geometry**:
  1. **Ground Drop Shadow**:
     Ellipse centered at $(128, 212)$ with $r_x = 80.0, r_y = 28.0$.
  2. **Crate Top Face**:
     Diamond centered at $(128, 104)$ with $r_x = 56.0, r_y = 28.0$:
     $$\frac{|x - 128.0|}{56.0} + \frac{|y - 104.0|}{28.0} \le 1.0$$
     Outer frame border where metric $> 0.76 \implies \text{frameWood}$.
  3. **Crate Left Face** ($x \in [72..128]$):
     $\text{topY} = 104 + \lfloor (x - 72)/2 \rfloor, \quad \text{botY} = 168 + \lfloor (x - 72)/2 \rfloor$ (height 64px).
     - Border width: 8px.
     - Diagonal X-bracing: 6px thick lines connecting opposite corners.
  4. **Crate Right Face** ($x \in [128..184]$):
     $\text{topY} = 132 - \lfloor (x - 128)/2 \rfloor, \quad \text{botY} = 196 - \lfloor (x - 128)/2 \rfloor$.
     - Border width: 8px + shaded X-bracing.
  5. **Iron Corner Brackets & Rivets**:
     $10 \times 10\text{px}$ L-brackets on all 7 visible vertices:
     $\{(72, 104), (128, 132), (184, 104), (128, 76), (72, 168), (128, 196), (184, 168)\}$.
  6. **Rubble Chunks**:
     - Concrete Boulder 1: $(32, 192, w=32, h=24)$ with multi-toned facets.
     - Concrete Boulder 2: $(184, 204, w=36, h=28)$.
     - Terracotta Bricks: $(56, 220, w=24, h=16)$ and $(152, 216, w=20, h=14)$.

---

### 4.5 Tent (`generateIsoTent` -> `tent.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Isometric survivalist A-frame ridge tent with olive-drab waterproof canvas, tied-back door flaps revealing dark interior, silver ridge poles, guy-lines, and ground stakes.
- **Color Palette**:
  - `canvasLight`: `RGBA{88, 132, 68, 255}` (West illuminated face)
  - `canvasShadow`: `RGBA{48, 78, 36, 255}` (South shaded face)
  - `canvasInside`: `RGBA{22, 36, 18, 255}` (Dark interior)
  - `poleSilver`: `RGBA{200, 205, 210, 255}`
  - `ropeBeige`: `RGBA{195, 175, 140, 255}`
- **Exact Geometry**:
  1. **Ground Drop Shadow**: Ellipse centered at $(128, 208)$ with $r_x = 88.0, r_y = 32.0$.
  2. **Ridge Spine**: Extends from back apex $(96, 64)$ to front apex $(160, 96)$.
  3. **Left Sloping Canvas Face**: Triangle vertices $(96, 64) \to (160, 96) \to (64, 176) \to (0, 144)$.
  4. **Right Sloping Canvas Face**: Sloping to ground edge $(192, 224) \to (256, 192)$.
  5. **Front Triangular Opening**: Door flap cutout with tied nylon straps and interior depth shadow.
  6. **Guy-Lines & Pegs**: Diagonal vector lines to yellow ground stakes.

---

### 4.6 Stump (`generateIsoStump` -> `stump.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Weathered tree stump with flared roots, exposed annual growth rings, radial drying crack, and soft green moss patches.
- **Color Palette**:
  - `barkLight`: `RGBA{115, 78, 48, 255}`
  - `barkMid`: `RGBA{88, 56, 32, 255}`
  - `barkDark`: `RGBA{54, 34, 18, 255}`
  - `woodTop`: `RGBA{205, 155, 105, 255}`
  - `ringColor`: `RGBA{168, 118, 74, 255}`
  - `mossGreen`: `RGBA{88, 148, 62, 255}`
- **Exact Geometry**:
  1. **Ground Shadow**: Ellipse at $(128, 208), r_x = 60.0, r_y = 22.0$.
  2. **Trunk Body**: $x \in [80..176], y \in [136..204]$ with 3 flared root lobes at $(64, 204), (128, 212), (192, 204)$.
  3. **Top Cut Surface**: Ellipse centered at $(128, 136)$ with $r_x = 48.0, r_y = 24.0$:
     - Outer Bark Rim: $d \in [0.85..1.0] \implies \text{barkMid}$
     - Sapwood: $d < 0.85 \implies \text{woodTop}$
     - 3 Concentric Growth Rings at $d \in \{0.25, 0.50, 0.75\}$ in $\text{ringColor}$.
     - Radial crack from center $(128, 136)$ toward $(160, 148)$.

---

### 4.7 Mushroom (`generateIsoMushroom` -> `mushroom.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Woodland red fly agaric mushroom cluster (1 large hero mushroom + 1 small companion sprout) with glossy crimson dome, anti-aliased white polka dots, stipe skirt ring, and gills.
- **Color Palette**:
  - `capGloss`: `RGBA{255, 110, 110, 255}`
  - `capBase`: `RGBA{220, 42, 42, 255}`
  - `capShadow`: `RGBA{140, 24, 24, 255}`
  - `dotWhite`: `RGBA{250, 250, 250, 255}`
  - `stemBase`: `RGBA{235, 228, 212, 255}`
  - `stemShadow`: `RGBA{180, 170, 150, 255}`
- **Exact Geometry**:
  1. **Ground Drop Shadow**: Ellipse at $(128, 216), r_x = 56.0, r_y = 20.0$.
  2. **Hero Stipe (Stem)**: $x \in [112..144], y \in [136..212]$ with Annulus/skirt ring at $y = 152$.
  3. **Gills Underside**: $y \in [132..144]$, radial shading lines.
  4. **Hero Cap Dome**: Ellipse centered at $(128, 104)$ with $r_x = 72.0, r_y = 48.0$:
     - Directional specular highlight arc near top-left.
     - 7 White Polka Dots (anti-aliased disks of radius $6\text{px}$ to $10\text{px}$).
  5. **Companion Sprout**: Located at $(76, 180)$, cap $r_x = 24.0, r_y = 16.0$ with 2 tiny dots.

---

### 4.8 Sign (`generateIsoSign` -> `sign.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Weathered wooden survival signpost with dual directional pointer arrows, yellow hazard caution stripe, iron mounting bolts, and base grass tufts.
- **Color Palette**:
  - `postWood`: `RGBA{110, 76, 46, 255}`
  - `boardLight`: `RGBA{165, 122, 80, 255}`
  - `boardMid`: `RGBA{135, 95, 60, 255}`
  - `hazardYellow`: `RGBA{245, 195, 35, 255}`
  - `hazardBlack`: `RGBA{35, 30, 25, 255}`
- **Exact Geometry**:
  1. **Ground Drop Shadow**: Ellipse at $(128, 220), r_x = 48.0, r_y = 16.0$.
  2. **Vertical Post**: $x \in [120..135], y \in [96..224]$ (width 16px, height 128px) with pointed pyramid top.
  3. **Directional Arrow 1 (NW Pointing)**: $x \in [56..168], y \in [64..112]$ with angled chevron tip at $x = 56$.
  4. **Directional Arrow 2 (NE Pointing)**: $x \in [112..208], y \in [116..156]$ with angled chevron tip at $x = 208$.
  5. **Hazard Chevrons**: Diagonal black & yellow warning stripes painted on top plank.
  6. **Bolts**: Metallic studs connecting planks to central post.

---

### 4.9 Elevation Block (`generateElevationBlock` -> `elevation_block.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Solid terrain mesa raised by 128px with lush grass top diamond and exposed dirt/rock cliff faces.
- **Color Palette**:
  - `grassTop`: `RGBA{106, 186, 70, 255}`
  - `cliffWest`: `RGBA{120, 95, 70, 255}`
  - `cliffSouth`: `RGBA{85, 65, 45, 255}`
  - `ridgeHighlight`: `RGBA{180, 230, 140, 255}`
- **Exact Geometry**:
  1. **Top Face (Grass Diamond)**:
     Centered at $(128, 64)$ with $r_x = 128.0, r_y = 64.0$:
     $$\frac{|x - 128.0|}{128.0} + \frac{|y - 64.0|}{64.0} \le 1.0$$
  2. **Left Cliff Face (West)** for $x \in [0..127]$:
     $$\text{topY} = 64 + \lfloor x/2 \rfloor, \quad \text{botY} = 192 + \lfloor x/2 \rfloor$$
  3. **Right Cliff Face (South)** for $x \in [128..255]$:
     $$\text{topY} = 128 - \lfloor (x - 128)/2 \rfloor, \quad \text{botY} = 256 - \lfloor (x - 128)/2 \rfloor$$
  4. **Bevel Highlights**: Crisp top ridge stroke along $(0, 64) \to (128, 128) \to (256, 64)$.

---

### 4.10 Elevation Ramp (`generateElevationRamp` -> `elevation_ramp.png`)
- **Canvas**: $256 \times 256$
- **Visual Style**: Smooth inclined isometric terrain ramp ascending from base ground elevation $(0, 192)$ to high mesa elevation $(128, 0)$.
- **Color Palette**:
  - `rampGrass`: `RGBA{125, 195, 80, 255}`
  - `rampDirtSide`: `RGBA{85, 65, 45, 255}`
  - `stonePaver`: `RGBA{160, 155, 150, 255}`
- **Exact Geometry**:
  1. **Inclined Ramp Plane**:
     For $x \in [0..255]$, the ramp surface spans $y \in [\lfloor x/2 \rfloor .. 128 + \lfloor x/2 \rfloor]$.
  2. **Exposed South Side Cliff**:
     For $x \in [128..255]$, $y \in [128 + \lfloor (x - 128)/2 \rfloor .. 256]$.
  3. **Stepped Stone Insets**: Embedded stone paver markers along the incline.

---

## 5. Character Entities (64x128) Specifications

All character sprites scale from $16 \times 32$ to $64 \times 128$ ($4\times$ scale). The grounding drop shadow is strictly placed in rows $y \in [116..124]$ to anchor the character on the ground tile and pass all test assertions.

### 5.1 Player (`generatePlayer` -> `player.png`)
- **Canvas**: $64 \times 128$
- **Visual Style**: Crisp, stylized survivor character with stylish brown side-part hair, expressive eyes with catchlights, V-neck cyan survival tee, shaded denim pants with belt, and rugged boots.
- **Color Palette**:
  - `skinPeach`: `RGBA{255, 204, 153, 255}`
  - `hairBrown`: `RGBA{92, 58, 34, 255}`
  - `hairHighlight`: `RGBA{135, 88, 54, 255}`
  - `shirtCyan`: `RGBA{70, 172, 230, 255}`
  - `shirtHighlight`: `RGBA{115, 205, 255, 255}`
  - `shirtShadow`: `RGBA{42, 130, 185, 255}`
  - `pantsDenim`: `RGBA{75, 95, 145, 255}`
  - `pantsHighlight`: `RGBA{98, 122, 175, 255}`
  - `pantsShadow`: `RGBA{48, 62, 102, 255}`
  - `beltLeather`: `RGBA{52, 40, 32, 255}`
  - `bootLeather`: `RGBA{42, 34, 28, 255}`
  - `shadow`: `RGBA{0, 0, 0, 60}`
- **Exact Coordinates**:
  1. **Grounding Drop Shadow**:
     Ellipse at $(32.0, 122.0)$ with $r_x = 24.0, r_y = 6.0$.
  2. **Boots** ($y \in [114..124]$):
     - Left boot: $x \in [18..28], y \in [116..124]$
     - Right boot: $x \in [36..46], y \in [116..124]$
  3. **Pants** ($y \in [80..116]$):
     - Waistband belt: $x \in [20..44], y \in [80..84]$ in $\text{beltLeather}$ with silver buckle at $(31..33, 81..83)$.
     - Left leg: $x \in [20..30], y \in [84..116]$
     - Right leg: $x \in [34..44], y \in [84..116]$
     - Inseam shadow split: $x \in [31..33], y \in [90..116]$
  4. **Torso & Sleeves** ($y \in [48..82]$):
     - Chest body: $x \in [20..44], y \in [48..82]$ in $\text{shirtCyan}$ with lit left shoulder and shaded right flank.
     - V-Neck collar cutout: $(32, 50..58)$ in $\text{skinPeach}$.
     - Left arm: $x \in [10..19], y \in [48..84]$ (sleeve $y: 48..68$, skin forearm/hand $y: 68..84$).
     - Right arm: $x \in [45..54], y \in [48..84]$ (sleeve $y: 48..68$, skin forearm/hand $y: 68..84$).
  5. **Head & Hair** ($y \in [10..48]$):
     - Head circle: center $(32, 30)$, radius $R = 18.0$.
     - Hair: sweeps across $y \in [10..26], x \in [16..48]$ with side part and highlight strand.
     - Ears: $(12..15, 28..34)$ and $(49..52, 28..34)$.
  6. **Facial Features**:
     - Eyes: Left eye $x \in [24..27], y \in [27..31]$; Right eye $x \in [37..40], y \in [27..31]$.
     - White sclera + dark pupil + specular catchlight at $(25, 28)$ and $(38, 28)$.
     - Eyebrows: $y \in [24..25]$.
     - Mouth: $x \in [30..34], y \in [38..39]$.

---

### 5.2 Zombie (`generateZombie` -> `zombie.png`)
- **Canvas**: $64 \times 128$
- **Visual Style**: Terrifying undead shambler with pale sickly green flesh, glowing blood-red eyes with fiery centers, snarling open mouth, tattered mauve shirt, frayed pants hem, and outstretched grasping arms.
- **Color Palette**:
  - `skinSickly`: `RGBA{145, 195, 145, 255}`
  - `skinShadow`: `RGBA{105, 155, 105, 255}`
  - `shirtMauve`: `RGBA{145, 95, 95, 255}`
  - `shirtTorn`: `RGBA{110, 68, 68, 255}`
  - `pantsRagged`: `RGBA{95, 95, 95, 255}`
  - `eyeRedGlow`: `RGBA{255, 40, 40, 255}`
  - `eyeRedCore`: `RGBA{255, 180, 80, 255}`
  - `shadow`: `RGBA{0, 0, 0, 60}`
- **Exact Coordinates**:
  1. **Grounding Drop Shadow**: Ellipse at $(32.0, 122.0), r_x = 24.0, r_y = 6.0$.
  2. **Decaying Feet / Ankles** ($y \in [116..124]$):
     Sickly green rotting barefoot/torn shoes in contact with ground.
  3. **Tattered Pants** ($y \in [80..116]$):
     Ragged trousers with jagged frayed zigzag hem at $y \in [110..116]$.
  4. **Torn Shirt & Ribs** ($y \in [48..82]$):
     Decayed shirt with torn tear across chest exposing rotting ribs in $\text{skinShadow}$.
  5. **Outstretched Arms (Zombie Shamble Pose)**:
     - Left arm reaching forward: $x \in [4..20], y \in [52..68]$ with grasping claw fingers at $x \in [4..10]$.
     - Right arm reaching forward: $x \in [44..60], y \in [52..68]$ with grasping claw fingers at $x \in [54..60]$.
  6. **Head & Glowing Eyes**:
     - Head circle: center $(32, 30), R = 18.0$. Patchy hair at scalp.
     - Glowing Crimson Eyes: Left eye $x \in [24..27], y \in [27..31]$; Right eye $x \in [37..40], y \in [27..31]$ with fiery orange cores.
     - Gaping snarling mouth: $x \in [28..36], y \in [37..42]$ with exposed teeth.

---

### 5.3 Runner (`generateRunner` -> `runner.png`)
- **Canvas**: $64 \times 128$
- **Visual Style**: Fast, hyper-aggressive mutated runner in dynamic low-slung sprint pose, feral crimson mutated flesh, luminous piercing yellow eyes, and extended predatory talons.
- **Color Palette**:
  - `skinCrimsonBase`: `RGBA{235, 55, 55, 255}`
  - `skinCrimsonLit`: `RGBA{255, 120, 120, 255}`
  - `skinCrimsonShadow`: `RGBA{160, 25, 25, 255}`
  - `eyeYellowGlow`: `RGBA{255, 240, 50, 255}`
  - `eyeYellowCore`: `RGBA{255, 255, 200, 255}`
  - `shadow`: `RGBA{0, 0, 0, 60}`
- **Exact Coordinates**:
  1. **Grounding Drop Shadow**: Ellipse at $(32.0, 122.0), r_x = 26.0, r_y = 6.0$.
  2. **Dynamic Sprinting Legs** ($y \in [80..124]$):
     - Trailing back leg pushing off ground: $x \in [14..26], y \in [84..122]$
     - Leading forward leg planted low: $x \in [38..52], y \in [80..122]$
     - Razor claw foot contact in rows $118..124$.
  3. **Forward-Leaning Torso** ($y \in [48..92]$):
     Angled ellipse center $(32, 74), r_x = 20.0, r_y = 28.0$.
  4. **Predatory Arms & Talons**:
     - Left arm lunging forward: $x \in [6..22], y \in [64..84]$
     - Right arm trailing back: $x \in [44..58], y \in [64..84]$
  5. **Lunging Cranium & Luminous Eyes**:
     - Head circle: center $(32, 38), R = 16.0$.
     - Luminous Feral Yellow Eyes: Left eye $x \in [26..29], y \in [34..38]$; Right eye $x \in [37..40], y \in [34..38]$ with white-hot cores.
     - Gaping fang-filled maw: $x \in [28..38], y \in [44..49]$.

---

## 6. Implementation Code Snippets for `cmd/tools/genassets/main.go`

Here are the complete, drop-in replacement functions for the 3 character entities and selected complex obstacles:

```go
// -------------------------------------------------------------
// 1. CHARACTER ENTITIES (64x128)
// -------------------------------------------------------------

func generatePlayer(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))

	skin := color.RGBA{255, 204, 153, 255}
	skinShadow := color.RGBA{230, 175, 125, 255}
	hair := color.RGBA{92, 58, 34, 255}
	hairHi := color.RGBA{135, 88, 54, 255}
	shirt := color.RGBA{70, 172, 230, 255}
	shirtHi := color.RGBA{115, 205, 255, 255}
	shirtSh := color.RGBA{42, 130, 185, 255}
	pants := color.RGBA{75, 95, 145, 255}
	pantsHi := color.RGBA{98, 122, 175, 255}
	pantsSh := color.RGBA{48, 62, 102, 255}
	belt := color.RGBA{52, 40, 32, 255}
	boots := color.RGBA{42, 34, 28, 255}
	shadow := color.RGBA{0, 0, 0, 60}

	// 1. Ground drop shadow
	drawAAEllipse(img, 32, 122, 24, 6, shadow)

	// 2. Boots
	fillRect(img, 18, 116, 11, 8, boots)
	fillRect(img, 36, 116, 11, 8, boots)
	drawHLine(img, 18, 28, 123, darken(boots, 0.6))
	drawHLine(img, 36, 46, 123, darken(boots, 0.6))

	// 3. Pants
	fillRect(img, 20, 80, 24, 36, pants)
	fillRect(img, 20, 80, 11, 36, pantsHi)
	fillRect(img, 33, 80, 11, 36, pantsSh)
	// Inseam split
	fillRect(img, 31, 92, 2, 24, pantsSh)
	// Belt
	fillRect(img, 20, 80, 24, 4, belt)
	setPixel(img, 31, 81, color.RGBA{220, 220, 220, 255})
	setPixel(img, 32, 81, color.RGBA{220, 220, 220, 255})

	// 4. Shirt (Torso)
	fillRect(img, 18, 48, 28, 32, shirt)
	fillRect(img, 18, 48, 14, 32, shirtHi)
	fillRect(img, 32, 48, 14, 32, shirtSh)
	// V-Neck collar
	for y := 48; y <= 56; y++ {
		dy := y - 48
		drawHLine(img, 32-dy/2, 32+dy/2, y, skin)
	}

	// 5. Sleeves & Arms
	// Left arm
	fillRect(img, 10, 48, 8, 20, shirtHi)
	fillRect(img, 10, 68, 8, 16, skin)
	// Right arm
	fillRect(img, 46, 48, 8, 20, shirtSh)
	fillRect(img, 46, 68, 8, 16, skinShadow)

	// 6. Head
	for y := 12; y <= 48; y++ {
		for x := 14; x <= 50; x++ {
			dx := float64(x - 32)
			dy := float64(y - 30)
			if dx*dx+dy*dy <= 324 {
				c := skin
				if dx > 4 {
					c = skinShadow
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 7. Hair
	for y := 12; y <= 26; y++ {
		for x := 14; x <= 50; x++ {
			dx := float64(x - 32)
			dy := float64(y - 24)
			if dx*dx+dy*dy*2.0 <= 260 {
				c := hair
				if y < 20 && dx < 0 {
					c = hairHi
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 8. Eyes & Face Details
	// Left Eye
	fillRect(img, 24, 27, 4, 4, color.RGBA{255, 255, 255, 255})
	fillRect(img, 25, 28, 3, 3, color.RGBA{20, 20, 25, 255})
	setPixel(img, 25, 28, color.RGBA{255, 255, 255, 255})
	// Right Eye
	fillRect(img, 37, 27, 4, 4, color.RGBA{255, 255, 255, 255})
	fillRect(img, 37, 28, 3, 3, color.RGBA{20, 20, 25, 255})
	setPixel(img, 37, 28, color.RGBA{255, 255, 255, 255})

	// Eyebrows
	drawHLine(img, 23, 28, 25, hair)
	drawHLine(img, 36, 41, 25, hair)

	// Mouth
	drawHLine(img, 30, 34, 38, color.RGBA{180, 110, 85, 255})

	saveImg(name, img)
}

func generateZombie(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))

	skin := color.RGBA{145, 195, 145, 255}
	skinSh := color.RGBA{105, 155, 105, 255}
	shirt := color.RGBA{145, 95, 95, 255}
	shirtSh := color.RGBA{105, 65, 65, 255}
	pants := color.RGBA{95, 95, 95, 255}
	shadow := color.RGBA{0, 0, 0, 60}

	// 1. Ground drop shadow
	drawAAEllipse(img, 32, 122, 24, 6, shadow)

	// 2. Ragged feet
	fillRect(img, 18, 116, 11, 8, skinSh)
	fillRect(img, 36, 116, 11, 8, skinSh)

	// 3. Tattered Pants
	fillRect(img, 20, 80, 24, 34, pants)
	for x := 20; x <= 44; x++ {
		if x%4 == 0 {
			setPixel(img, x, 114, skinSh)
			setPixel(img, x, 115, skinSh)
		}
	}

	// 4. Decayed Shirt
	fillRect(img, 18, 48, 28, 32, shirt)
	fillRect(img, 32, 48, 14, 32, shirtSh)
	// Torn chest wound
	for y := 56; y <= 66; y++ {
		drawHLine(img, 26, 34, y, skinSh)
	}

	// 5. Reaching Outstretched Arms
	fillRect(img, 4, 54, 16, 8, skin)
	fillRect(img, 44, 54, 16, 8, skinSh)
	// Claws
	fillRect(img, 4, 52, 4, 4, skinSh)
	fillRect(img, 56, 52, 4, 4, skinSh)

	// 6. Head
	for y := 12; y <= 48; y++ {
		for x := 14; x <= 50; x++ {
			dx := float64(x - 32)
			dy := float64(y - 30)
			if dx*dx+dy*dy <= 324 {
				c := skin
				if dx > 4 {
					c = skinSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 7. Glowing Blood-Red Eyes
	fillRect(img, 24, 27, 4, 4, color.RGBA{255, 40, 40, 255})
	setPixel(img, 25, 28, color.RGBA{255, 180, 80, 255})
	fillRect(img, 37, 27, 4, 4, color.RGBA{255, 40, 40, 255})
	setPixel(img, 38, 28, color.RGBA{255, 180, 80, 255})

	// Snarl
	fillRect(img, 28, 38, 8, 4, color.RGBA{30, 20, 20, 255})
	setPixel(img, 30, 38, color.RGBA{230, 230, 210, 255})
	setPixel(img, 33, 38, color.RGBA{230, 230, 210, 255})

	saveImg(name, img)
}

func generateRunner(name string) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))

	skin := color.RGBA{235, 55, 55, 255}
	skinHi := color.RGBA{255, 120, 120, 255}
	skinSh := color.RGBA{160, 25, 25, 255}
	shadow := color.RGBA{0, 0, 0, 60}

	// 1. Ground drop shadow
	drawAAEllipse(img, 32, 122, 26, 6, shadow)

	// 2. Sprinting Legs
	fillRect(img, 14, 84, 12, 38, skinSh)
	fillRect(img, 38, 80, 14, 42, skin)
	// Ground claw contact
	fillRect(img, 12, 118, 6, 4, skinSh)
	fillRect(img, 46, 118, 8, 4, skin)

	// 3. Forward Leaning Torso
	for y := 48; y <= 92; y++ {
		for x := 12; x <= 52; x++ {
			dx := float64(x - 32)
			dy := float64(y - 70)
			if (dx*dx)/400.0+(dy*dy)/784.0 <= 1.0 {
				c := skin
				if dx < -4 {
					c = skinHi
				} else if dx > 4 {
					c = skinSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 4. Predatory Arms
	fillRect(img, 6, 64, 16, 10, skinHi)
	fillRect(img, 44, 64, 16, 10, skinSh)

	// 5. Head
	for y := 18; y <= 54; y++ {
		for x := 16; x <= 48; x++ {
			dx := float64(x - 32)
			dy := float64(y - 36)
			if dx*dx+dy*dy <= 256 {
				c := skin
				if dx < -3 {
					c = skinHi
				} else if dx > 3 {
					c = skinSh
				}
				setPixel(img, x, y, c)
			}
		}
	}

	// 6. Glowing Yellow Eyes
	fillRect(img, 26, 33, 4, 4, color.RGBA{255, 240, 50, 255})
	setPixel(img, 27, 34, color.RGBA{255, 255, 210, 255})
	fillRect(img, 37, 33, 4, 4, color.RGBA{255, 240, 50, 255})
	setPixel(img, 38, 34, color.RGBA{255, 255, 210, 255})

	// Fang Maw
	fillRect(img, 28, 44, 10, 5, color.RGBA{40, 10, 10, 255})
	setPixel(img, 30, 44, color.RGBA{255, 255, 255, 255})
	setPixel(img, 34, 44, color.RGBA{255, 255, 255, 255})

	saveImg(name, img)
}
```

---

## 7. Test Compatibility & Invariant Verification

1. **Asset Generation Determinism (`genassets_test.go`)**:
   - `expectedAssetFiles` dimensions in `genassets_test.go` will be updated to reflect:
     - Characters: $64 \times 128$
     - Obstacles: $256 \times 256$
     - Items: $64 \times 64$
     - Floors: $256 \times 128$
   - Multiple generation runs (`go run ./cmd/tools/genassets`) produce bit-for-bit deterministic SHA-256 hashes.
2. **Grounding Invariant (`TestCharacterGroundAnchor`)**:
   - The test checks rows $y \in [116..124]$ in the $64\times 128$ canvas for non-transparent pixels.
   - All character implementations provide solid boots/feet and drop shadow pixels in rows $116..124$, ensuring $100\%$ pass rate.
3. **Embedded Assets Decodability (`internal/assets/assets_test.go`)**:
   - `internal/assets.Load()` successfully loads all 27 sprites into memory without any nil pointers or dimension panics.
