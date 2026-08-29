package assets

import (
	"image"
	"image/color"
	"math"

	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	// Wall 4-bit cardinal autotile sprites (16 states: 0..15)
	WallAutotileImages [16]*ebiten.Image

	// Wall front facade drop shadow image
	WallFacadeShadowImage *ebiten.Image

	// Fence 4-bit cardinal autotile sprites (16 states: 0..15)
	FenceAutotileImages [16]*ebiten.Image

	// Quadrant Transition Overlays map: [TileType][Quadrant][SubtileState]
	// SubtileState: Full, HorizontalEdge, VerticalEdge, OuterCorner, InnerCorner, plus DiagonalTip
	terrainOverlays map[world.TileType]map[world.Quadrant]map[world.SubtileState]*ebiten.Image
	diagonalOverlays map[world.TileType]map[world.Quadrant]*ebiten.Image
)

// initAutotiles is called once during assets.Load() to generate autotile images.
func initAutotiles() {
	generateWallAutotiles()
	generateFenceAutotiles()
	generateTerrainOverlays()
}

// GetWallAutotileImage returns the autotile image for the given 4-bit wall bitmask (0..15).
func GetWallAutotileImage(mask uint8) *ebiten.Image {
	if int(mask) < len(WallAutotileImages) && WallAutotileImages[mask] != nil {
		return WallAutotileImages[mask]
	}
	return WallImage
}

// GetFenceAutotileImage returns the autotile image for the given 4-bit fence bitmask (0..15).
func GetFenceAutotileImage(mask uint8) *ebiten.Image {
	if int(mask) < len(FenceAutotileImages) && FenceAutotileImages[mask] != nil {
		return FenceAutotileImages[mask]
	}
	return FenceImage
}

// GetTerrainOverlay returns the quadrant transition overlay image for blending terrain.
func GetTerrainOverlay(t world.TileType, q world.Quadrant, s world.SubtileState, isDiag bool) *ebiten.Image {
	if isDiag {
		if diagMap, ok := diagonalOverlays[t]; ok {
			return diagMap[q]
		}
		return nil
	}
	if qMap, ok := terrainOverlays[t]; ok {
		if sMap, ok := qMap[q]; ok {
			return sMap[s]
		}
	}
	return nil
}

func generateWallAutotiles() {
	const size = 256
	// Wall colors: Stylized Slate Masonry
	capColor := color.RGBA{R: 75, G: 85, B: 98, A: 255}
	capHighlight := color.RGBA{R: 110, G: 125, B: 145, A: 255}
	capShadow := color.RGBA{R: 45, G: 52, B: 62, A: 255}
	wallBody := color.RGBA{R: 55, G: 62, B: 72, A: 255}
	wallDark := color.RGBA{R: 38, G: 44, B: 52, A: 255}
	brickLine := color.RGBA{R: 30, G: 34, B: 40, A: 255}

	for mask := 0; mask < 16; mask++ {
		img := image.NewRGBA(image.Rect(0, 0, size, size))

		hasN := (mask & (1 << 0)) != 0
		hasE := (mask & (1 << 1)) != 0
		hasS := (mask & (1 << 2)) != 0
		hasW := (mask & (1 << 3)) != 0

		minX, maxX := 76, 180
		minY := 76

		if hasW {
			minX = 0
		}
		if hasE {
			maxX = size
		}
		if hasN {
			minY = 0
		}

		// 1. Drop shadow beneath wall base
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				// Base shape test
				inCenter := (x >= 76 && x < 180 && y >= 76 && y < 180)
				inN := (hasN && x >= 76 && x < 180 && y < 76)
				inS := (hasS && x >= 76 && x < 180 && y >= 180)
				inW := (hasW && x < 76 && y >= 76 && y < 180)
				inE := (hasE && x >= 180 && y >= 76 && y < 180)

				if inCenter || inN || inS || inW || inE {
					// Inside wall geometry
					// Vertical facade on lower portion (y >= 140) if no South connection or on horizontal wall
					isFrontFacade := (y >= 140 && (!hasS || y >= 200 || (hasE && x >= 180) || (hasW && x < 76)))
					if isFrontFacade {
						img.Set(x, y, wallDark)
						// Stylized brick mortar lines
						if (y == 165 || y == 195 || y == 225) || (x%40 == 0 && ((y/30)%2 == 0)) {
							img.Set(x, y, brickLine)
						}
					} else {
						img.Set(x, y, wallBody)
						// Top Cap bevels and highlight
						isTopBevel := (y <= minY+12)
						isLeftBevel := (x <= minX+12)
						isBottomBevel := (y >= 135 && y <= 142)
						isRightBevel := (x >= maxX-12)

						if isTopBevel || isLeftBevel {
							img.Set(x, y, capHighlight)
						} else if isBottomBevel || isRightBevel {
							img.Set(x, y, capShadow)
						} else {
							img.Set(x, y, capColor)
							// Stylized cap slab divider
							if x == 128 || (hasW && x == 64) || (hasE && x == 192) {
								img.Set(x, y, capShadow)
							}
						}
					}
				}
			}
		}

		WallAutotileImages[mask] = ebiten.NewImageFromImage(img)
	}

	// Create WallFacadeShadowImage (soft shadow on ground south of wall)
	shadowImg := image.NewRGBA(image.Rect(0, 0, size, size/2))
	for y := 0; y < size/2; y++ {
		alpha := uint8(float64(size/2-y) / float64(size/2) * 90.0)
		for x := 0; x < size; x++ {
			shadowImg.Set(x, y, color.RGBA{0, 0, 0, alpha})
		}
	}
	WallFacadeShadowImage = ebiten.NewImageFromImage(shadowImg)
}

func generateFenceAutotiles() {
	const size = 256
	woodColor := color.RGBA{R: 160, G: 115, B: 75, A: 255}
	woodHighlight := color.RGBA{R: 195, G: 145, B: 95, A: 255}
	woodShadow := color.RGBA{R: 115, G: 80, B: 50, A: 255}
	postColor := color.RGBA{R: 145, G: 100, B: 65, A: 255}

	for mask := 0; mask < 16; mask++ {
		img := image.NewRGBA(image.Rect(0, 0, size, size))

		hasN := (mask & (1 << 0)) != 0
		hasE := (mask & (1 << 1)) != 0
		hasS := (mask & (1 << 2)) != 0
		hasW := (mask & (1 << 3)) != 0

		// Draw Horizontal Rails (Top and Bottom Rails)
		if hasW || hasE {
			minX := 100
			maxX := 156
			if hasW {
				minX = 0
			}
			if hasE {
				maxX = size
			}

			// Top Rail (y: 90..115)
			for y := 90; y < 115; y++ {
				for x := minX; x < maxX; x++ {
					if y <= 95 {
						img.Set(x, y, woodHighlight)
					} else if y >= 110 {
						img.Set(x, y, woodShadow)
					} else {
						img.Set(x, y, woodColor)
					}
				}
			}
			// Bottom Rail (y: 155..180)
			for y := 155; y < 180; y++ {
				for x := minX; x < maxX; x++ {
					if y <= 160 {
						img.Set(x, y, woodHighlight)
					} else if y >= 175 {
						img.Set(x, y, woodShadow)
					} else {
						img.Set(x, y, woodColor)
					}
				}
			}
		}

		// Draw Vertical Posts / Slats
		if hasN {
			for y := 0; y < 100; y++ {
				for x := 108; x < 148; x++ {
					if x <= 116 {
						img.Set(x, y, woodHighlight)
					} else if x >= 140 {
						img.Set(x, y, woodShadow)
					} else {
						img.Set(x, y, woodColor)
					}
				}
			}
		}
		if hasS {
			for y := 156; y < size; y++ {
				for x := 108; x < 148; x++ {
					if x <= 116 {
						img.Set(x, y, woodHighlight)
					} else if x >= 140 {
						img.Set(x, y, woodShadow)
					} else {
						img.Set(x, y, woodColor)
					}
				}
			}
		}

		// Center Main Post (x: 100..156, y: 75..195)
		for y := 75; y < 195; y++ {
			for x := 100; x < 156; x++ {
				if x <= 110 || y <= 85 {
					img.Set(x, y, woodHighlight)
				} else if x >= 146 || y >= 185 {
					img.Set(x, y, woodShadow)
				} else {
					img.Set(x, y, postColor)
				}
			}
		}

		FenceAutotileImages[mask] = ebiten.NewImageFromImage(img)
	}
}

type terrainStyle struct {
	base      color.RGBA
	highlight color.RGBA
	shadow    color.RGBA
	accent    color.RGBA
	hasGrass  bool
	hasCurb   bool
	hasPlanks bool
	hasTiles  bool
}

func getTerrainStyle(t world.TileType) terrainStyle {
	switch t {
	case world.TileGrass:
		return terrainStyle{
			base:      color.RGBA{R: 110, G: 148, B: 42, A: 255},
			highlight: color.RGBA{R: 135, G: 175, B: 55, A: 255},
			shadow:    color.RGBA{R: 85, G: 118, B: 30, A: 255},
			accent:    color.RGBA{R: 145, G: 190, B: 60, A: 255},
			hasGrass:  true,
		}
	case world.TileConcrete:
		return terrainStyle{
			base:      color.RGBA{R: 185, G: 190, B: 195, A: 255},
			highlight: color.RGBA{R: 215, G: 220, B: 225, A: 255},
			shadow:    color.RGBA{R: 145, G: 150, B: 155, A: 255},
			accent:    color.RGBA{R: 130, G: 135, B: 140, A: 255},
			hasCurb:   true,
		}
	case world.TileAsphalt:
		return terrainStyle{
			base:      color.RGBA{R: 52, G: 56, B: 62, A: 255},
			highlight: color.RGBA{R: 70, G: 76, B: 84, A: 255},
			shadow:    color.RGBA{R: 35, G: 38, B: 42, A: 255},
			accent:    color.RGBA{R: 240, G: 200, B: 50, A: 255}, // Road line yellow
		}
	case world.TileWoodFloor:
		return terrainStyle{
			base:      color.RGBA{R: 160, G: 110, B: 65, A: 255},
			highlight: color.RGBA{R: 190, G: 135, B: 85, A: 255},
			shadow:    color.RGBA{R: 120, G: 80, B: 45, A: 255},
			accent:    color.RGBA{R: 90, G: 60, B: 35, A: 255},
			hasPlanks: true,
		}
	case world.TileTileFloor:
		return terrainStyle{
			base:      color.RGBA{R: 215, G: 220, B: 225, A: 255},
			highlight: color.RGBA{R: 240, G: 245, B: 250, A: 255},
			shadow:    color.RGBA{R: 175, G: 180, B: 185, A: 255},
			accent:    color.RGBA{R: 140, G: 145, B: 150, A: 255},
			hasTiles:  true,
		}
	default:
		return terrainStyle{
			base:      color.RGBA{R: 110, G: 148, B: 42, A: 255},
			highlight: color.RGBA{R: 135, G: 175, B: 55, A: 255},
			shadow:    color.RGBA{R: 85, G: 118, B: 30, A: 255},
		}
	}
}

func generateTerrainOverlays() {
	terrainOverlays = make(map[world.TileType]map[world.Quadrant]map[world.SubtileState]*ebiten.Image)
	diagonalOverlays = make(map[world.TileType]map[world.Quadrant]*ebiten.Image)

	overlayTypes := []world.TileType{
		world.TileGrass,
		world.TileConcrete,
		world.TileAsphalt,
		world.TileWoodFloor,
		world.TileTileFloor,
	}

	quads := []world.Quadrant{world.QuadNW, world.QuadNE, world.QuadSW, world.QuadSE}
	states := []world.SubtileState{
		world.SubtileFull,
		world.SubtileHorizontalEdge,
		world.SubtileVerticalEdge,
		world.SubtileOuterCorner,
		world.SubtileInnerCorner,
	}

	const qSize = 128 // 128x128 high-res quadrant, drawn at 64x64

	for _, tt := range overlayTypes {
		style := getTerrainStyle(tt)
		terrainOverlays[tt] = make(map[world.Quadrant]map[world.SubtileState]*ebiten.Image)
		diagonalOverlays[tt] = make(map[world.Quadrant]*ebiten.Image)

		for _, q := range quads {
			terrainOverlays[tt][q] = make(map[world.SubtileState]*ebiten.Image)

			for _, s := range states {
				img := image.NewRGBA(image.Rect(0, 0, qSize, qSize))
				renderQuadrantOverlay(img, q, s, false, style, qSize)
				terrainOverlays[tt][q][s] = ebiten.NewImageFromImage(img)
			}

			// Diagonal corner tip overlay
			diagImg := image.NewRGBA(image.Rect(0, 0, qSize, qSize))
			renderQuadrantOverlay(diagImg, q, world.SubtileOuterCorner, true, style, qSize)
			diagonalOverlays[tt][q] = ebiten.NewImageFromImage(diagImg)
		}
	}
}

func renderQuadrantOverlay(img *image.RGBA, q world.Quadrant, s world.SubtileState, isDiag bool, style terrainStyle, size int) {
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Normalize coordinates (u, v) so (0,0) is always the exterior corner and (1,1) is the quadrant center
			var u, v float64
			switch q {
			case world.QuadNW:
				u = float64(x) / float64(size)
				v = float64(y) / float64(size)
			case world.QuadNE:
				u = float64(size-1-x) / float64(size)
				v = float64(y) / float64(size)
			case world.QuadSW:
				u = float64(x) / float64(size)
				v = float64(size-1-y) / float64(size)
			case world.QuadSE:
				u = float64(size-1-x) / float64(size)
				v = float64(size-1-y) / float64(size)
			}

			inside := false
			edgeFactor := 0.0

			if isDiag {
				// Only diagonal corner active: small rounded triangular cap at (0,0)
				dist := math.Hypot(u, v)
				if dist <= 0.65 {
					inside = true
					edgeFactor = dist / 0.65
				}
			} else {
				switch s {
				case world.SubtileFull:
					inside = true
					edgeFactor = 0.0
				case world.SubtileHorizontalEdge:
					// Horizontal border from top (v = 0) extending into quadrant
					// Scalloped / wave edge
					fringe := 0.55 + 0.12*math.Sin(u*math.Pi*4.0)
					if v <= fringe {
						inside = true
						edgeFactor = v / fringe
					}
				case world.SubtileVerticalEdge:
					// Vertical border from left (u = 0) extending into quadrant
					fringe := 0.55 + 0.12*math.Sin(v*math.Pi*4.0)
					if u <= fringe {
						inside = true
						edgeFactor = u / fringe
					}
				case world.SubtileOuterCorner:
					// Convex outer corner cap coming from exterior corner (u=0, v=0)
					dist := math.Hypot(u, v)
					fringe := 0.85 + 0.10*math.Cos(math.Atan2(v, u)*6.0)
					if dist <= fringe {
						inside = true
						edgeFactor = dist / fringe
					}
				case world.SubtileInnerCorner:
					// Concave inner notch at (0,0) — fills whole quadrant EXCEPT near (0,0)
					dist := math.Hypot(u, v)
					fringe := 0.45 + 0.08*math.Sin(math.Atan2(v, u)*4.0)
					if dist >= fringe {
						inside = true
						edgeFactor = 1.0 - (dist-fringe)/0.55
					}
				}
			}

			if inside {
				pixColor := style.base
				if edgeFactor < 0.25 {
					pixColor = style.highlight
				} else if edgeFactor > 0.80 {
					pixColor = style.shadow
				}

				// Stylized chevron blades for grass overlay
				if style.hasGrass && ((x+y)%16 == 0 || (x-y)%16 == 0) && edgeFactor < 0.6 {
					pixColor = style.accent
				}

				// Stylized curb border for concrete
				if style.hasCurb && edgeFactor > 0.70 {
					pixColor = style.accent
				}

				img.Set(x, y, pixColor)
			}
		}
	}
}
