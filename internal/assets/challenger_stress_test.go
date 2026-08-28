package assets

import (
	"bytes"
	"image"
	_ "image/png"
	"math"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// All 27 exported pointers descriptor
type assetDescriptor struct {
	name   string
	ptr    **ebiten.Image
	path   string
	wantW  int
	wantH  int
	cat    string
}

func getAssetDescriptors() []assetDescriptor {
	return []assetDescriptor{
		// Character Entities (3) - 64x128
		{"PlayerImage", &PlayerImage, "images/player.png", 64, 128, "Entity"},
		{"ZombieImage", &ZombieImage, "images/zombie.png", 64, 128, "Entity"},
		{"RunnerImage", &RunnerImage, "images/runner.png", 64, 128, "Entity"},

		// Floor Tiles (6) - 256x128
		{"GrassImage", &GrassImage, "images/grass.png", 256, 128, "Floor"},
		{"DirtImage", &DirtImage, "images/dirt.png", 256, 128, "Floor"},
		{"WoodImage", &WoodImage, "images/wood.png", 256, 128, "Floor"},
		{"AsphaltImage", &AsphaltImage, "images/asphalt.png", 256, 128, "Floor"},
		{"ConcreteImage", &ConcreteImage, "images/concrete.png", 256, 128, "Floor"},
		{"TileFloorImage", &TileFloorImage, "images/tile_floor.png", 256, 128, "Floor"},

		// Vertical Obstacles / Props (10) - 256x256
		{"WallImage", &WallImage, "images/wall.png", 256, 256, "Obstacle/Prop"},
		{"TreeImage", &TreeImage, "images/tree.png", 256, 256, "Obstacle/Prop"},
		{"FenceImage", &FenceImage, "images/fence.png", 256, 256, "Obstacle/Prop"},
		{"DebrisImage", &DebrisImage, "images/debris.png", 256, 256, "Obstacle/Prop"},
		{"TentImage", &TentImage, "images/tent.png", 256, 256, "Obstacle/Prop"},
		{"StumpImage", &StumpImage, "images/stump.png", 256, 256, "Obstacle/Prop"},
		{"MushroomImage", &MushroomImage, "images/mushroom.png", 256, 256, "Obstacle/Prop"},
		{"SignImage", &SignImage, "images/sign.png", 256, 256, "Obstacle/Prop"},
		{"ElevationBlockImage", &ElevationBlockImage, "images/elevation_block.png", 256, 256, "Obstacle/Prop"},
		{"ElevationRampImage", &ElevationRampImage, "images/elevation_ramp.png", 256, 256, "Obstacle/Prop"},

		// Item / Weapon / Equipment (8) - 64x64
		{"FoodImage", &FoodImage, "images/food.png", 64, 64, "Item"},
		{"WaterImage", &WaterImage, "images/water.png", 64, 64, "Item"},
		{"WeaponImage", &WeaponImage, "images/weapon.png", 64, 64, "Item"},
		{"AxeImage", &AxeImage, "images/axe.png", 64, 64, "Item"},
		{"ShotgunImage", &ShotgunImage, "images/shotgun.png", 64, 64, "Item"},
		{"AmmoImage", &AmmoImage, "images/ammo.png", 64, 64, "Item"},
		{"ArmorImage", &ArmorImage, "images/armor.png", 64, 64, "Item"},
		{"AntidoteImage", &AntidoteImage, "images/antidote.png", 64, 64, "Item"},
	}
}

// TestChallenger_All27ExportedPointersAndExactBounds rigorously validates
// count, non-nil status, and exact pixel boundaries for all 27 assets.
func TestChallenger_All27ExportedPointersAndExactBounds(t *testing.T) {
	Load()

	descriptors := getAssetDescriptors()
	if len(descriptors) != 27 {
		t.Fatalf("expected exactly 27 exported image descriptors, got %d", len(descriptors))
	}

	for _, d := range descriptors {
		t.Run(d.name, func(t *testing.T) {
			img := *d.ptr
			if img == nil {
				t.Fatalf("Pointer %s is nil after Load()", d.name)
			}

			bounds := img.Bounds()
			if bounds.Min.X != 0 || bounds.Min.Y != 0 {
				t.Errorf("Pointer %s Bounds().Min = (%d, %d), want (0, 0)", d.name, bounds.Min.X, bounds.Min.Y)
			}
			if bounds.Dx() != d.wantW || bounds.Dy() != d.wantH {
				t.Errorf("Pointer %s Bounds() = %dx%d, want %dx%d (%s category)",
					d.name, bounds.Dx(), bounds.Dy(), d.wantW, d.wantH, d.cat)
			}
		})
	}
}

// TestChallenger_MultiThreadedLoadAndPointerRace stress-tests concurrent
// Load() calls and concurrent pointer reads with multiple goroutines.
func TestChallenger_MultiThreadedLoadAndPointerRace(t *testing.T) {
	const numLoaders = 20
	const numReaders = 30
	const iterations = 50

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	// 1. Concurrent Load() goroutines
	for i := 0; i < numLoaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-startSignal
			for j := 0; j < iterations; j++ {
				Load()
			}
		}(i)
	}

	// 2. Concurrent Reader goroutines reading all 27 pointers
	descriptors := getAssetDescriptors()
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-startSignal
			for j := 0; j < iterations*2; j++ {
				for _, d := range descriptors {
					img := *d.ptr
					if img != nil {
						b := img.Bounds()
						if b.Dx() <= 0 || b.Dy() <= 0 {
							t.Errorf("reader %d: %s invalid bounds: %v", id, d.name, b)
						}
					}
				}
			}
		}(i)
	}

	// Release all goroutines simultaneously
	close(startSignal)
	wg.Wait()

	// Final verification after concurrent stress
	for _, d := range descriptors {
		img := *d.ptr
		if img == nil {
			t.Fatalf("post-concurrency: pointer %s is nil", d.name)
		}
		b := img.Bounds()
		if b.Dx() != d.wantW || b.Dy() != d.wantH {
			t.Fatalf("post-concurrency: %s bounds mismatch: got %dx%d, want %dx%d", d.name, b.Dx(), b.Dy(), d.wantW, d.wantH)
		}
	}
}

// TestChallenger_RepeatedSequentialLoads tests repeated consecutive Load() calls
func TestChallenger_RepeatedSequentialLoads(t *testing.T) {
	descriptors := getAssetDescriptors()
	for iter := 1; iter <= 100; iter++ {
		Load()
		for _, d := range descriptors {
			if *d.ptr == nil {
				t.Fatalf("iteration %d: %s is nil", iter, d.name)
			}
		}
	}
}

// TestChallenger_AssetPixelContrastAndColorSaturation performs statistical
// color analysis on all 27 embedded assets.
func TestChallenger_AssetPixelContrastAndColorSaturation(t *testing.T) {
	descriptors := getAssetDescriptors()

	type stats struct {
		name          string
		totalPixels   int
		nonTransCount int
		fillRatio     float64
		meanLum       float64
		rmsContrast   float64
		minLum        float64
		maxLum        float64
		meanSat       float64
		maxSat        float64
	}

	results := make([]stats, 0, len(descriptors))

	for _, d := range descriptors {
		t.Run(d.name, func(t *testing.T) {
			data, err := imageFS.ReadFile(d.path)
			if err != nil {
				t.Fatalf("failed to read embedded %s: %v", d.path, err)
			}

			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", d.path, err)
			}

			bounds := img.Bounds()
			w, h := bounds.Dx(), bounds.Dy()
			total := w * h

			var nonTransCount int
			var luminances []float64
			var saturations []float64

			minL := 255.0
			maxL := 0.0
			maxS := 0.0

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r16, g16, b16, a16 := img.At(x, y).RGBA()
					if a16 > 0 {
						nonTransCount++
						// Unmultiply alpha to get standard 8-bit RGB
						aNorm := float64(a16) / 65535.0
						r8 := (float64(r16) / 65535.0) / aNorm * 255.0
						g8 := (float64(g16) / 65535.0) / aNorm * 255.0
						b8 := (float64(b16) / 65535.0) / aNorm * 255.0

						if r8 > 255 {
							r8 = 255
						}
						if g8 > 255 {
							g8 = 255
						}
						if b8 > 255 {
							b8 = 255
						}

						// Perceived Luminance (ITU-R BT.601)
						lum := 0.299*r8 + 0.587*g8 + 0.114*b8
						luminances = append(luminances, lum)

						if lum < minL {
							minL = lum
						}
						if lum > maxL {
							maxL = lum
						}

						// HSV Saturation: S = (max(R,G,B) - min(R,G,B)) / max(R,G,B)
						maxC := math.Max(r8, math.Max(g8, b8))
						minC := math.Min(r8, math.Min(g8, b8))
						sat := 0.0
						if maxC > 0 {
							sat = (maxC - minC) / maxC
						}
						saturations = append(saturations, sat)
						if sat > maxS {
							maxS = sat
						}
					}
				}
			}

			if nonTransCount == 0 {
				t.Fatalf("asset %s has NO non-transparent pixels!", d.name)
			}

			fillRatio := float64(nonTransCount) / float64(total)

			// Minimum fill ratios based on asset type
			minExpectedFill := 0.05
			if d.cat == "Floor" {
				minExpectedFill = 0.40 // Isometric diamonds occupy ~50% of bounding box
			} else if d.cat == "Entity" {
				minExpectedFill = 0.15
			}

			if fillRatio < minExpectedFill {
				t.Errorf("%s fill ratio %.2f%% is below minimum expected %.2f%%",
					d.name, fillRatio*100, minExpectedFill*100)
			}

			// Mean Luminance and RMS Contrast
			var sumLum float64
			for _, l := range luminances {
				sumLum += l
			}
			meanLum := sumLum / float64(len(luminances))

			var sumSqDiff float64
			for _, l := range luminances {
				diff := l - meanLum
				sumSqDiff += diff * diff
			}
			rmsContrast := math.Sqrt(sumSqDiff / float64(len(luminances)))

			// Mean Saturation
			var sumSat float64
			for _, s := range saturations {
				sumSat += s
			}
			meanSat := sumSat / float64(len(saturations))

			// Record stats
			st := stats{
				name:          d.name,
				totalPixels:   total,
				nonTransCount: nonTransCount,
				fillRatio:     fillRatio,
				meanLum:       meanLum,
				rmsContrast:   rmsContrast,
				minLum:        minL,
				maxLum:        maxL,
				meanSat:       meanSat,
				maxSat:        maxS,
			}
			results = append(results, st)

			// Assertions on Contrast & Saturation
			// 1. RMS Contrast must be > 10.0 (prevents flat solid monochromatic textures)
			if rmsContrast < 10.0 {
				t.Errorf("%s has very low RMS contrast (%.2f < 10.0), image may be flat/unshaded", d.name, rmsContrast)
			}

			// 2. Dynamic Luminance Range must be >= 40.0
			dynRange := maxL - minL
			if dynRange < 40.0 {
				t.Errorf("%s dynamic range too small (%.2f < 40.0)", d.name, dynRange)
			}

			// 3. Colored assets (like grass, wood, tree, axe, etc.) should have non-zero saturation
			// Grayscale textures (concrete, asphalt, wall) might have lower saturation
			if d.name != "ConcreteImage" && d.name != "AsphaltImage" && d.name != "AmmoImage" {
				if maxS < 0.20 {
					t.Errorf("%s lacks color saturation (maxS = %.2f < 0.20)", d.name, maxS)
				}
			}
		})
	}

	// Print summary table in test logs for empirical documentation
	t.Logf("\n%-20s | %-10s | %-9s | %-8s | %-12s | %-12s | %-8s",
		"Asset", "FillRatio", "MeanLum", "Contrast", "DynRange", "MeanSat", "MaxSat")
	t.Logf("-----------------------------------------------------------------------------------------------")
	for _, r := range results {
		t.Logf("%-20s | %9.2f%% | %9.2f | %8.2f | %12.2f | %8.2f | %8.2f",
			r.name, r.fillRatio*100, r.meanLum, r.rmsContrast, r.maxLum-r.minLum, r.meanSat, r.maxSat)
	}
}

// TestChallenger_FloorTileGeometryDiamond ensures all floor tiles fit cleanly
// within the standard 2:1 isometric diamond without bleeding.
func TestChallenger_FloorTileGeometryDiamond(t *testing.T) {
	floorTiles := []struct {
		name string
		path string
	}{
		{"GrassImage", "images/grass.png"},
		{"DirtImage", "images/dirt.png"},
		{"WoodImage", "images/wood.png"},
		{"AsphaltImage", "images/asphalt.png"},
		{"ConcreteImage", "images/concrete.png"},
		{"TileFloorImage", "images/tile_floor.png"},
	}

	for _, ft := range floorTiles {
		t.Run(ft.name, func(t *testing.T) {
			data, err := imageFS.ReadFile(ft.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", ft.path, err)
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", ft.path, err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 256 || bounds.Dy() != 128 {
				t.Fatalf("%s bounds = %dx%d, want 256x128", ft.name, bounds.Dx(), bounds.Dy())
			}

			centerX := 127.5
			centerY := 63.5
			radiusX := 128.5
			radiusY := 64.5

			bleedPixels := 0
			interiorPixels := 0
			var totalInsideDiamond int

			for y := 0; y < 128; y++ {
				for x := 0; x < 256; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					normDist := math.Abs(float64(x)-centerX)/radiusX + math.Abs(float64(y)-centerY)/radiusY

					if normDist <= 0.95 {
						totalInsideDiamond++
						if a > 0 {
							interiorPixels++
						}
					}

					// Beyond 1.15 is considered external bleeding
					if normDist > 1.15 && a > 0 {
						bleedPixels++
					}
				}
			}

			if bleedPixels > 0 {
				t.Errorf("%s has %d pixels bleeding significantly outside isometric diamond boundary", ft.name, bleedPixels)
			}

			interiorFill := float64(interiorPixels) / float64(totalInsideDiamond)
			if interiorFill < 0.98 {
				t.Errorf("%s interior diamond fill ratio is %.2f%%, want >= 98%% (holes detected)", ft.name, interiorFill*100)
			}
		})
	}
}

// TestChallenger_CharacterGroundDropShadows checks that player, zombie, and runner
// have grounded drop shadows and full vertical posture.
func TestChallenger_CharacterGroundDropShadows(t *testing.T) {
	characters := []struct {
		name string
		path string
	}{
		{"PlayerImage", "images/player.png"},
		{"ZombieImage", "images/zombie.png"},
		{"RunnerImage", "images/runner.png"},
	}

	for _, c := range characters {
		t.Run(c.name, func(t *testing.T) {
			data, err := imageFS.ReadFile(c.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", c.path, err)
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", c.path, err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 128 {
				t.Fatalf("%s bounds = %dx%d, want 64x128", c.name, bounds.Dx(), bounds.Dy())
			}

			// 1. Head pixels in y in [0..40]
			headPixels := 0
			for y := 0; y <= 40; y++ {
				for x := 0; x < 64; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						headPixels++
					}
				}
			}
			if headPixels == 0 {
				t.Errorf("%s missing head/torso pixels in top region [0..40]", c.name)
			}

			// 2. Ground/shadow pixels in y in [112..127]
			groundPixels := 0
			for y := 112; y < 128; y++ {
				for x := 0; x < 64; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						groundPixels++
					}
				}
			}
			if groundPixels < 50 {
				t.Errorf("%s insufficient grounding/shadow pixels in bottom region [112..127] (count: %d)", c.name, groundPixels)
			}
		})
	}
}

// TestChallenger_ItemIconCenteringAndContour verifies 64x64 item sprites
func TestChallenger_ItemIconCenteringAndContour(t *testing.T) {
	items := []string{
		"images/food.png",
		"images/water.png",
		"images/weapon.png",
		"images/axe.png",
		"images/shotgun.png",
		"images/ammo.png",
		"images/armor.png",
		"images/antidote.png",
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

			var count int
			var sumX, sumY float64

			for y := 0; y < 64; y++ {
				for x := 0; x < 64; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						count++
						sumX += float64(x)
						sumY += float64(y)
					}
				}
			}

			if count < 200 {
				t.Errorf("%s has too few solid pixels: %d", path, count)
			}

			centroidX := sumX / float64(count)
			centroidY := sumY / float64(count)

			// Centroid should be reasonably centered in 64x64 icon box (between 20 and 44)
			if centroidX < 20.0 || centroidX > 44.0 || centroidY < 20.0 || centroidY > 44.0 {
				t.Errorf("%s centroid (%.2f, %.2f) is off-center (expected within [20, 44])",
					path, centroidX, centroidY)
			}
		})
	}
}
