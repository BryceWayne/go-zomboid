package assets

import (
	"bytes"
	"image"
	_ "image/png"
	"math"
	"testing"
)

// TestFloorTileIsometricBounds verifies that floor tiles (64x32)
// do not bleed heavily outside the 2:1 isometric diamond.
func TestFloorTileIsometricBounds(t *testing.T) {
	floorTiles := []string{
		"images/grass.png",
		"images/dirt.png",
		"images/wood.png",
		"images/asphalt.png",
		"images/concrete.png",
		"images/tile_floor.png",
	}

	for _, path := range floorTiles {
		t.Run(path, func(t *testing.T) {
			data, err := imageFS.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", path, err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 32 {
				t.Fatalf("%s bounds = %dx%d, want 64x32", path, bounds.Dx(), bounds.Dy())
			}

			// Center of diamond is (31.5, 15.5)
			centerX := 31.5
			centerY := 15.5
			radiusX := 32.5
			radiusY := 16.5

			invalidBleedPixels := 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						// Normalized Manhattan distance in diamond coordinates
						dist := math.Abs(float64(x)-centerX)/radiusX + math.Abs(float64(y)-centerY)/radiusY
						// Allow slight anti-aliasing / pixel boundary tolerance up to 1.15
						if dist > 1.15 {
							invalidBleedPixels++
						}
					}
				}
			}

			if invalidBleedPixels > 0 {
				t.Errorf("%s has %d pixels bleeding outside isometric diamond", path, invalidBleedPixels)
			}
		})
	}
}

// TestCharacterGroundAnchor verifies that character entities (16x32)
// have feet anchored at the lower boundary (y in [28..31]) to prevent floating.
func TestCharacterGroundAnchor(t *testing.T) {
	characters := []string{
		"images/player.png",
		"images/zombie.png",
		"images/runner.png",
	}

	for _, path := range characters {
		t.Run(path, func(t *testing.T) {
			data, err := imageFS.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", path, err)
			}

			// Check bottom 4 rows for grounding pixels
			groundPixels := 0
			for y := 28; y < 32; y++ {
				for x := 0; x < 16; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						groundPixels++
					}
				}
			}

			if groundPixels == 0 {
				t.Errorf("character %s has no grounding pixels in rows 28-31 (appears floating)", path)
			}
		})
	}
}

// TestItemOutlineContrast verifies that 16x16 item sprites have adequate
// pixel density and contrast against transparent backgrounds.
func TestItemOutlineContrast(t *testing.T) {
	items := []string{
		"images/food.png",
		"images/water.png",
		"images/weapon.png",
		"images/axe.png",
		"images/shotgun.png",
		"images/ammo.png",
		"images/armor.png",
	}

	for _, path := range items {
		t.Run(path, func(t *testing.T) {
			data, err := imageFS.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", path, err)
			}

			solidCount := 0
			darkContourCount := 0

			for y := 0; y < 16; y++ {
				for x := 0; x < 16; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					if a > 0 {
						solidCount++
						// 16-bit color values in RGBA()
						r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
						// Dark border check: luminance < 80
						lum := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
						if lum < 80 {
							darkContourCount++
						}
					}
				}
			}

			if solidCount < 20 {
				t.Errorf("item %s has too few solid pixels (%d / 256)", path, solidCount)
			}
			if darkContourCount == 0 {
				t.Errorf("item %s lacks dark outline contrast pixels", path)
			}
		})
	}
}

// TestAssetsLoadIdempotency verifies that calling Load() multiple times
// does not panic, leak nil handles, or corrupt state.
func TestAssetsLoadIdempotency(t *testing.T) {
	for i := 0; i < 3; i++ {
		Load()

		if PlayerImage == nil || ZombieImage == nil || RunnerImage == nil ||
			GrassImage == nil || DirtImage == nil || WoodImage == nil ||
			AsphaltImage == nil || ConcreteImage == nil || TileFloorImage == nil ||
			WallImage == nil || TreeImage == nil || FenceImage == nil || DebrisImage == nil ||
			WeaponImage == nil || AxeImage == nil || ShotgunImage == nil ||
			AmmoImage == nil || ArmorImage == nil || FoodImage == nil || WaterImage == nil {
			t.Fatalf("iteration %d: one or more asset pointers is nil after Load()", i)
		}
	}
}
