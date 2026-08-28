package assets

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestEmbeddedAssetDimensionsAndValidity(t *testing.T) {
	expectedAssets := []struct {
		path   string
		width  int
		height int
	}{
		// Entities (16x32)
		{"images/player.png", 16, 32},
		{"images/zombie.png", 16, 32},
		{"images/runner.png", 16, 32},

		// Floor Tiles (64x32)
		{"images/grass.png", 64, 32},
		{"images/dirt.png", 64, 32},
		{"images/wood.png", 64, 32},
		{"images/asphalt.png", 64, 32},
		{"images/concrete.png", 64, 32},
		{"images/tile_floor.png", 64, 32},

		// Vertical Obstacles (64x64)
		{"images/wall.png", 64, 64},
		{"images/tree.png", 64, 64},
		{"images/fence.png", 64, 64},
		{"images/debris.png", 64, 64},

		// Items, Weapons & Armor (16x16)
		{"images/weapon.png", 16, 16},
		{"images/axe.png", 16, 16},
		{"images/shotgun.png", 16, 16},
		{"images/ammo.png", 16, 16},
		{"images/armor.png", 16, 16},
		{"images/food.png", 16, 16},
		{"images/water.png", 16, 16},
	}

	if len(expectedAssets) != 20 {
		t.Fatalf("expected 20 assets to be tested, found %d", len(expectedAssets))
	}

	for _, tc := range expectedAssets {
		t.Run(tc.path, func(t *testing.T) {
			data, err := imageFS.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("failed to read embedded file %s: %v", tc.path, err)
			}

			if len(data) == 0 {
				t.Fatalf("embedded file %s is empty", tc.path)
			}

			img, format, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode image %s: %v", tc.path, err)
			}

			if format != "png" {
				t.Errorf("image %s format = %s, want png", tc.path, format)
			}

			bounds := img.Bounds()
			if bounds.Dx() != tc.width || bounds.Dy() != tc.height {
				t.Errorf("image %s dimensions = %dx%d, want %dx%d",
					tc.path, bounds.Dx(), bounds.Dy(), tc.width, tc.height)
			}

			// Verify that image contains non-transparent pixels
			nonTransparentCount := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						nonTransparentCount++
					}
				}
			}

			if nonTransparentCount == 0 {
				t.Errorf("image %s has no non-transparent pixels", tc.path)
			}
		})
	}
}

func TestAssetsLoadAllPointersNonNil(t *testing.T) {
	// Call Load()
	Load()

	handles := []struct {
		name   string
		img    *ebiten.Image
		wantW  int
		wantH  int
	}{
		// Entities
		{"PlayerImage", PlayerImage, 16, 32},
		{"ZombieImage", ZombieImage, 16, 32},
		{"RunnerImage", RunnerImage, 16, 32},

		// Floors
		{"GrassImage", GrassImage, 64, 32},
		{"DirtImage", DirtImage, 64, 32},
		{"WoodImage", WoodImage, 64, 32},
		{"AsphaltImage", AsphaltImage, 64, 32},
		{"ConcreteImage", ConcreteImage, 64, 32},
		{"TileFloorImage", TileFloorImage, 64, 32},

		// Obstacles
		{"WallImage", WallImage, 64, 64},
		{"TreeImage", TreeImage, 64, 64},
		{"FenceImage", FenceImage, 64, 64},
		{"DebrisImage", DebrisImage, 64, 64},

		// Items
		{"WeaponImage", WeaponImage, 16, 16},
		{"AxeImage", AxeImage, 16, 16},
		{"ShotgunImage", ShotgunImage, 16, 16},
		{"AmmoImage", AmmoImage, 16, 16},
		{"ArmorImage", ArmorImage, 16, 16},
		{"FoodImage", FoodImage, 16, 16},
		{"WaterImage", WaterImage, 16, 16},
	}

	if len(handles) != 20 {
		t.Fatalf("expected 20 asset pointers to be checked, found %d", len(handles))
	}

	for _, h := range handles {
		t.Run(h.name, func(t *testing.T) {
			if h.img == nil {
				t.Fatalf("asset pointer %s is nil after Load()", h.name)
			}
			bounds := h.img.Bounds()
			if bounds.Dx() != h.wantW || bounds.Dy() != h.wantH {
				t.Errorf("asset %s dimensions = %dx%d, want %dx%d",
					h.name, bounds.Dx(), bounds.Dy(), h.wantW, h.wantH)
			}
		})
	}
}
