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
		// Character Entities (64x128)
		{"images/player.png", 64, 128},
		{"images/zombie.png", 64, 128},
		{"images/runner.png", 64, 128},

		// Floor Tiles (256x128)
		{"images/grass.png", 256, 128},
		{"images/dirt.png", 256, 128},
		{"images/wood.png", 256, 128},
		{"images/asphalt.png", 256, 128},
		{"images/concrete.png", 256, 128},
		{"images/tile_floor.png", 256, 128},

		// Vertical Obstacles & Props (256x256)
		{"images/wall.png", 256, 256},
		{"images/tree.png", 256, 256},
		{"images/fence.png", 256, 256},
		{"images/debris.png", 256, 256},
		{"images/tent.png", 256, 256},
		{"images/stump.png", 256, 256},
		{"images/mushroom.png", 256, 256},
		{"images/sign.png", 256, 256},
		{"images/elevation_block.png", 256, 256},
		{"images/elevation_ramp.png", 256, 256},

		// Items, Weapons & Equipment (64x64)
		{"images/food.png", 64, 64},
		{"images/water.png", 64, 64},
		{"images/weapon.png", 64, 64},
		{"images/axe.png", 64, 64},
		{"images/shotgun.png", 64, 64},
		{"images/ammo.png", 64, 64},
		{"images/armor.png", 64, 64},
		{"images/antidote.png", 64, 64},
	}

	if len(expectedAssets) != 27 {
		t.Fatalf("expected 27 assets to be tested, found %d", len(expectedAssets))
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
	Load()

	handles := []struct {
		name  string
		img   *ebiten.Image
		wantW int
		wantH int
	}{
		// 1. Entities (64x128)
		{"PlayerImage", PlayerImage, 64, 128},
		{"ZombieImage", ZombieImage, 64, 128},
		{"RunnerImage", RunnerImage, 64, 128},

		// 2. Floors (256x128)
		{"GrassImage", GrassImage, 256, 128},
		{"DirtImage", DirtImage, 256, 128},
		{"WoodImage", WoodImage, 256, 128},
		{"AsphaltImage", AsphaltImage, 256, 128},
		{"ConcreteImage", ConcreteImage, 256, 128},
		{"TileFloorImage", TileFloorImage, 256, 128},

		// 3. Obstacles & Props (256x256)
		{"WallImage", WallImage, 256, 256},
		{"TreeImage", TreeImage, 256, 256},
		{"FenceImage", FenceImage, 256, 256},
		{"DebrisImage", DebrisImage, 256, 256},
		{"TentImage", TentImage, 256, 256},
		{"StumpImage", StumpImage, 256, 256},
		{"MushroomImage", MushroomImage, 256, 256},
		{"SignImage", SignImage, 256, 256},
		{"ElevationBlockImage", ElevationBlockImage, 256, 256},
		{"ElevationRampImage", ElevationRampImage, 256, 256},

		// 4. Items & Equipment (64x64)
		{"FoodImage", FoodImage, 64, 64},
		{"WaterImage", WaterImage, 64, 64},
		{"WeaponImage", WeaponImage, 64, 64},
		{"AxeImage", AxeImage, 64, 64},
		{"ShotgunImage", ShotgunImage, 64, 64},
		{"AmmoImage", AmmoImage, 64, 64},
		{"ArmorImage", ArmorImage, 64, 64},
		{"AntidoteImage", AntidoteImage, 64, 64},

		// 5. External World Props & Foliage
		{"BenchImage", BenchImage, 52, 37},
		{"ChestImage", ChestImage, 22, 21},
		{"Sculpture1Image", Sculpture1Image, 23, 31},
		{"Sculpture2Image", Sculpture2Image, 29, 32},
		{"SculptureImage", SculptureImage, 23, 31},
		{"Bush1Image", Bush1Image, 24, 18},
		{"Bush2Image", Bush2Image, 19, 15},
		{"Bush3Image", Bush3Image, 25, 19},
		{"Bush4Image", Bush4Image, 28, 19},
		{"BushImage", BushImage, 24, 18},
		{"Flower1Image", Flower1Image, 26, 25},
		{"Flower2Image", Flower2Image, 24, 22},
		{"Flower3Image", Flower3Image, 26, 18},
		{"FlowerImage", FlowerImage, 26, 25},
		{"Stone1Image", Stone1Image, 28, 19},
		{"Stone2Image", Stone2Image, 29, 25},
		{"StoneImage", StoneImage, 28, 19},
		{"ForestStumpImage", ForestStumpImage, 29, 19},
		{"GrassTuft1Image", GrassTuft1Image, 25, 24},
		{"GrassTuft2Image", GrassTuft2Image, 31, 15},

		// 6. External Tilesets
		{"LabTilesetImage", LabTilesetImage, 768, 768},
		{"ZombieTilesetImage", ZombieTilesetImage, 764, 300},
	}

	if len(handles) != 49 {
		t.Fatalf("expected 49 asset pointers to be checked (27 legacy + 22 external), found %d", len(handles))
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
