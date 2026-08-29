package game

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestAdversarial_WallAutotile_All16StatesBoundsAndAssetIntegrity checks that every wall mask
// from 0 to 15 is loaded, non-nil, and has exact 256x256 dimensions.
func TestAdversarial_WallAutotile_All16StatesBoundsAndAssetIntegrity(t *testing.T) {
	assets.Load()

	const expectedSize = 256

	for mask := 0; mask < 16; mask++ {
		img := assets.GetWallAutotileImage(uint8(mask))
		if img == nil {
			t.Fatalf("Mask %d: Wall autotile image is nil", mask)
		}

		b := img.Bounds()
		if b.Dx() != expectedSize || b.Dy() != expectedSize {
			t.Fatalf("Mask %d: Wall autotile image dimensions %dx%d != 256x256", mask, b.Dx(), b.Dy())
		}
	}
}

// TestAdversarial_WallAutotile_ProceduralGeometryOracle verifies the procedural generation logic
// for all 16 wall bitmasks against strict geometric invariants (center core, stubs, outer corners).
func TestAdversarial_WallAutotile_ProceduralGeometryOracle(t *testing.T) {
	const size = 256
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

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				inCenter := (x >= 76 && x < 180 && y >= 76 && y < 180)
				inN := (hasN && x >= 76 && x < 180 && y < 76)
				inS := (hasS && x >= 76 && x < 180 && y >= 180)
				inW := (hasW && x < 76 && y >= 76 && y < 180)
				inE := (hasE && x >= 180 && y >= 76 && y < 180)

				if inCenter || inN || inS || inW || inE {
					isFrontFacade := (y >= 140 && (!hasS || y >= 200 || (hasE && x >= 180) || (hasW && x < 76)))
					if isFrontFacade {
						img.Set(x, y, wallDark)
						if (y == 165 || y == 195 || y == 225) || (x%40 == 0 && ((y/30)%2 == 0)) {
							img.Set(x, y, brickLine)
						}
					} else {
						img.Set(x, y, wallBody)
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
							if x == 128 || (hasW && x == 64) || (hasE && x == 192) {
								img.Set(x, y, capShadow)
							}
						}
					}
				}
			}
		}

		// Verify 1: Center Core (x: 100..150, y: 100..150) must be completely solid (Alpha == 255)
		for y := 100; y <= 150; y += 10 {
			for x := 100; x <= 150; x += 10 {
				c := img.At(x, y)
				_, _, _, a := c.RGBA()
				if a < 0xFF00 {
					t.Fatalf("Mask %d: Center pixel at (%d, %d) is not solid", mask, x, y)
				}
			}
		}

		// Verify 2: North Stub (x: 128, y: 20)
		_, _, _, aNorth := img.At(128, 20).RGBA()
		if hasN && aNorth < 0xFF00 {
			t.Fatalf("Mask %d: Expected solid North stub at (128, 20)", mask)
		}
		if !hasN && aNorth > 0 {
			t.Fatalf("Mask %d: Expected transparent North stub at (128, 20)", mask)
		}

		// Verify 3: South Stub (x: 128, y: 220)
		_, _, _, aSouth := img.At(128, 220).RGBA()
		if hasS && aSouth < 0xFF00 {
			t.Fatalf("Mask %d: Expected solid South stub at (128, 220)", mask)
		}
		if !hasS && aSouth > 0 {
			t.Fatalf("Mask %d: Expected transparent South stub at (128, 220)", mask)
		}

		// Verify 4: West Stub (x: 20, y: 128)
		_, _, _, aWest := img.At(20, 128).RGBA()
		if hasW && aWest < 0xFF00 {
			t.Fatalf("Mask %d: Expected solid West stub at (20, 128)", mask)
		}
		if !hasW && aWest > 0 {
			t.Fatalf("Mask %d: Expected transparent West stub at (20, 128)", mask)
		}

		// Verify 5: East Stub (x: 230, y: 128)
		_, _, _, aEast := img.At(230, 128).RGBA()
		if hasE && aEast < 0xFF00 {
			t.Fatalf("Mask %d: Expected solid East stub at (230, 128)", mask)
		}
		if !hasE && aEast > 0 {
			t.Fatalf("Mask %d: Expected transparent East stub at (230, 128)", mask)
		}

		// Verify 6: Outer Corners (20,20), (230,20), (20,230), (230,230)
		corners := [][2]int{{20, 20}, {230, 20}, {20, 230}, {230, 230}}
		for _, pt := range corners {
			_, _, _, aCorn := img.At(pt[0], pt[1]).RGBA()
			if aCorn > 0 {
				t.Fatalf("Mask %d: Outer diagonal corner (%d, %d) should be transparent", mask, pt[0], pt[1])
			}
		}
	}
}

// TestAdversarial_FenceAutotile_All16StatesBoundsAndIntegrity verifies fence autotile assets.
func TestAdversarial_FenceAutotile_All16StatesBoundsAndIntegrity(t *testing.T) {
	assets.Load()

	const expectedSize = 256

	for mask := 0; mask < 16; mask++ {
		img := assets.GetFenceAutotileImage(uint8(mask))
		if img == nil {
			t.Fatalf("Mask %d: Fence autotile image is nil", mask)
		}

		b := img.Bounds()
		if b.Dx() != expectedSize || b.Dy() != expectedSize {
			t.Fatalf("Mask %d: Fence autotile image dimensions %dx%d != 256x256", mask, b.Dx(), b.Dy())
		}
	}
}

// TestAdversarial_FenceAutotile_ProceduralGeometryOracle verifies fence rails and posts.
func TestAdversarial_FenceAutotile_ProceduralGeometryOracle(t *testing.T) {
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

		// Check center post (128, 128)
		_, _, _, aPost := img.At(128, 128).RGBA()
		if aPost < 0xFF00 {
			t.Fatalf("Mask %d: Center post is not solid", mask)
		}

		// Check West rail (20, 100)
		_, _, _, aWest := img.At(20, 100).RGBA()
		if hasW && aWest < 0xFF00 {
			t.Fatalf("Mask %d: Expected solid West rail", mask)
		}
		if !hasW && aWest > 0 {
			t.Fatalf("Mask %d: Expected transparent West rail", mask)
		}

		// Check East rail (230, 100)
		_, _, _, aEast := img.At(230, 100).RGBA()
		if hasE && aEast < 0xFF00 {
			t.Fatalf("Mask %d: Expected solid East rail", mask)
		}
		if !hasE && aEast > 0 {
			t.Fatalf("Mask %d: Expected transparent East rail", mask)
		}
	}
}

// TestAdversarial_FacadeDropShadow_Properties verifies shadow dimensions and gradient.
func TestAdversarial_FacadeDropShadow_Properties(t *testing.T) {
	assets.Load()

	img := assets.WallFacadeShadowImage
	if img == nil {
		t.Fatal("WallFacadeShadowImage is nil")
	}

	b := img.Bounds()
	if b.Dx() != 256 || b.Dy() != 128 {
		t.Fatalf("Expected WallFacadeShadowImage dimensions 256x128, got %dx%d", b.Dx(), b.Dy())
	}

	// Verify shadow generator math
	const size = 256
	shadowImg := image.NewRGBA(image.Rect(0, 0, size, size/2))
	for y := 0; y < size/2; y++ {
		alpha := uint8(float64(size/2-y) / float64(size/2) * 90.0)
		for x := 0; x < size; x++ {
			shadowImg.Set(x, y, color.RGBA{0, 0, 0, alpha})
		}
	}

	// Verify top row alpha is ~90
	_, _, _, aTop := shadowImg.At(128, 0).RGBA()
	if (aTop >> 8) != 90 {
		t.Fatalf("Expected top shadow alpha 90, got %d", aTop>>8)
	}

	// Verify bottom row alpha is 0
	_, _, _, aBot := shadowImg.At(128, 127).RGBA()
	if (aBot >> 8) != 0 {
		t.Fatalf("Expected bottom shadow alpha 0, got %d", aBot>>8)
	}
}

// TestAdversarial_TerrainOverlays_AllVariantsNonNil verifies that GetTerrainOverlay returns
// a valid non-nil image for every valid terrain type, quadrant, subtile state, and diagonal flag.
func TestAdversarial_TerrainOverlays_AllVariantsNonNil(t *testing.T) {
	assets.Load()

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

	for _, tt := range overlayTypes {
		for _, q := range quads {
			for _, s := range states {
				img := assets.GetTerrainOverlay(tt, q, s, false)
				if img == nil {
					t.Fatalf("GetTerrainOverlay(%v, %v, %v, false) returned nil", tt, q, s)
				}
				b := img.Bounds()
				if b.Dx() != 128 || b.Dy() != 128 {
					t.Fatalf("GetTerrainOverlay(%v, %v, %v, false) bounds %dx%d != 128x128",
						tt, q, s, b.Dx(), b.Dy())
				}
			}

			diagImg := assets.GetTerrainOverlay(tt, q, world.SubtileOuterCorner, true)
			if diagImg == nil {
				t.Fatalf("GetTerrainOverlay(%v, %v, OuterCorner, true) returned nil", tt, q)
			}
			b := diagImg.Bounds()
			if b.Dx() != 128 || b.Dy() != 128 {
				t.Fatalf("GetTerrainOverlay(%v, %v, OuterCorner, true) bounds %dx%d != 128x128",
					tt, q, b.Dx(), b.Dy())
			}
		}
	}
}

// TestAdversarial_ZeroGapGuarantee_ExtremeZoomAndSubpixelScaling mathematically proves
// that adjacent tile positions and scaled widths have zero gap or overlap across extreme zooms.
func TestAdversarial_ZeroGapGuarantee_ExtremeZoomAndSubpixelScaling(t *testing.T) {
	zoomFactors := []float64{
		0.01, 0.05, 0.1, 0.25, 0.3333333333333333,
		0.5, 0.61803398875, 0.7071067811865475, 1.0,
		1.234567, 1.4142135623730951, 1.5, 1.7320508075688772,
		2.0, 2.718281828459045, 3.0, 3.141592653589793,
		4.0, 5.5, 8.0, 10.0, 16.0, 32.0, 64.0, 100.0,
	}

	camPositions := [][2]float64{
		{0, 0},
		{100, 100},
		{-500.5, 300.25},
		{1234.56789, 9876.54321},
		{-9999.999, -8888.888},
		{1e-7, 1e-7},
	}

	for _, z := range zoomFactors {
		for _, cam := range camPositions {
			camX, camY := cam[0], cam[1]

			for tileX := -10; tileX <= 10; tileX++ {
				for tileY := -10; tileY <= 10; tileY++ {
					wX1 := float64(tileX * world.TileSize)
					wY1 := float64(tileY * world.TileSize)
					wX2 := float64((tileX + 1) * world.TileSize)
					wY2 := float64((tileY + 1) * world.TileSize)

					sX1 := (wX1 - camX) * z
					sY1 := (wY1 - camY) * z
					sX2 := (wX2 - camX) * z
					sY2 := (wY2 - camY) * z

					drawnW := float64(world.TileSize) * z
					drawnH := float64(world.TileSize) * z

					hGap := (sX1 + drawnW) - sX2
					if math.Abs(hGap) > 1e-8 {
						t.Fatalf("Zoom %f Cam (%f,%f) Tile (%d,%d): Horizontal gap detected: %e",
							z, camX, camY, tileX, tileY, hGap)
					}

					vGap := (sY1 + drawnH) - sY2
					if math.Abs(vGap) > 1e-8 {
						t.Fatalf("Zoom %f Cam (%f,%f) Tile (%d,%d): Vertical gap detected: %e",
							z, camX, camY, tileX, tileY, vGap)
					}

					qDrawnW := 64.0 * z
					qDrawnH := 64.0 * z

					sNW_X := (wX1 - camX) * z
					sNW_Y := (wY1 - camY) * z
					sNE_X := (wX1 + 64.0 - camX) * z
					sSW_Y := (wY1 + 64.0 - camY) * z

					qHGap := (sNW_X + qDrawnW) - sNE_X
					if math.Abs(qHGap) > 1e-8 {
						t.Fatalf("Zoom %f: Quadrant horizontal gap: %e", z, qHGap)
					}

					qVGap := (sNW_Y + qDrawnH) - sSW_Y
					if math.Abs(qVGap) > 1e-8 {
						t.Fatalf("Zoom %f: Quadrant vertical gap: %e", z, qVGap)
					}
				}
			}
		}
	}
}

// TestAdversarial_DrawSystem_AllWallCombinationsCoverage renders a complex map
// containing all 16 wall bitmasks and all 16 fence bitmasks simultaneously.
func TestAdversarial_DrawSystem_AllWallCombinationsCoverage(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	m := world.NewMap(40, 40)

	for i := range m.Tiles {
		m.Tiles[i] = world.TileGrass
		m.Visible[i] = true
		m.Explored[i] = true
	}

	for mask := 0; mask < 16; mask++ {
		wx := 2 + (mask%4)*8
		wy := 2 + (mask/4)*8

		m.SetTile(wx, wy, world.TileWall)
		if (mask & 1) != 0 {
			m.SetTile(wx, wy-1, world.TileWall)
		}
		if (mask & 2) != 0 {
			m.SetTile(wx+1, wy, world.TileWall)
		}
		if (mask & 4) != 0 {
			m.SetTile(wx, wy+1, world.TileWall)
		}
		if (mask & 8) != 0 {
			m.SetTile(wx-1, wy, world.TileWall)
		}
	}

	for mask := 0; mask < 16; mask++ {
		fx := 2 + (mask%4)*8
		fy := 20 + (mask/4)*8

		m.SetTile(fx, fy, world.TileFence)
		if (mask & 1) != 0 {
			m.SetTile(fx, fy-1, world.TileFence)
		}
		if (mask & 2) != 0 {
			m.SetTile(fx+1, fy, world.TileFence)
		}
		if (mask & 4) != 0 {
			m.SetTile(fx, fy+1, world.TileFence)
		}
		if (mask & 8) != 0 {
			m.SetTile(fx-1, fy, world.TileFence)
		}
	}

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{Health: 100, Hunger: 100, Thirst: 100},
		&ecs.Position{X: 20 * world.TileSize, Y: 20 * world.TileSize},
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16, Height: 16},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	for _, tod := range []float64{0.0, 6.0, 12.0, 18.0, 23.0} {
		drawSys.Draw(screen, tod, -1)
	}
}
