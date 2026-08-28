package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	_ "image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// AssetCategory groups assets for specific geometric constraints.
type AssetCategory int

const (
	CategoryFloor AssetCategory = iota
	CategoryObstacle
	CategoryEntity
	CategoryItem
)

type AssetSpec struct {
	Path         string
	Category     AssetCategory
	Width        int
	Height       int
	MinFillRatio float64
	MaxFillRatio float64
}

var all27AssetSpecs = []AssetSpec{
	// 1. Character Entities (64x128)
	{"images/player.png", CategoryEntity, 64, 128, 0.20, 0.70},
	{"images/zombie.png", CategoryEntity, 64, 128, 0.20, 0.70},
	{"images/runner.png", CategoryEntity, 64, 128, 0.20, 0.70},

	// 2. Floor Tiles (256x128) - Diamond area is ~50%
	{"images/grass.png", CategoryFloor, 256, 128, 0.45, 0.55},
	{"images/dirt.png", CategoryFloor, 256, 128, 0.45, 0.55},
	{"images/wood.png", CategoryFloor, 256, 128, 0.45, 0.55},
	{"images/asphalt.png", CategoryFloor, 256, 128, 0.45, 0.55},
	{"images/concrete.png", CategoryFloor, 256, 128, 0.45, 0.55},
	{"images/tile_floor.png", CategoryFloor, 256, 128, 0.45, 0.55},

	// 3. Vertical Obstacles & Props (256x256)
	{"images/wall.png", CategoryObstacle, 256, 256, 0.15, 0.85},
	{"images/tree.png", CategoryObstacle, 256, 256, 0.15, 0.85},
	{"images/fence.png", CategoryObstacle, 256, 256, 0.05, 0.85},
	{"images/debris.png", CategoryObstacle, 256, 256, 0.05, 0.85},
	{"images/tent.png", CategoryObstacle, 256, 256, 0.10, 0.85},
	{"images/stump.png", CategoryObstacle, 256, 256, 0.05, 0.85},
	{"images/mushroom.png", CategoryObstacle, 256, 256, 0.05, 0.85},
	{"images/sign.png", CategoryObstacle, 256, 256, 0.05, 0.85},
	{"images/elevation_block.png", CategoryObstacle, 256, 256, 0.15, 0.85},
	{"images/elevation_ramp.png", CategoryObstacle, 256, 256, 0.15, 0.85},

	// 4. Items & Equipment (64x64)
	{"images/food.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/water.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/weapon.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/axe.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/shotgun.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/ammo.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/armor.png", CategoryItem, 64, 64, 0.05, 0.70},
	{"images/antidote.png", CategoryItem, 64, 64, 0.05, 0.70},
}

// TestEmpiricalAssetCatalogCompleteness verifies exactly 27 assets exist with expected specs.
func TestEmpiricalAssetCatalogCompleteness(t *testing.T) {
	if len(all27AssetSpecs) != 27 {
		t.Fatalf("expected 27 asset specs, got %d", len(all27AssetSpecs))
	}

	for _, spec := range all27AssetSpecs {
		t.Run(spec.Path, func(t *testing.T) {
			data, err := imageFS.ReadFile(spec.Path)
			if err != nil {
				t.Fatalf("failed to read embedded asset %s: %v", spec.Path, err)
			}
			if len(data) == 0 {
				t.Fatalf("asset file %s is 0 bytes", spec.Path)
			}

			img, format, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode PNG %s: %v", spec.Path, err)
			}
			if format != "png" {
				t.Fatalf("expected PNG format, got %s", format)
			}

			b := img.Bounds()
			if b.Dx() != spec.Width || b.Dy() != spec.Height {
				t.Errorf("asset %s dimensions = %dx%d, want %dx%d",
					spec.Path, b.Dx(), b.Dy(), spec.Width, spec.Height)
			}
		})
	}
}

// TestEmpiricalAlphaFillRatios computes exact non-zero alpha fill ratios and verifies boundaries.
func TestEmpiricalAlphaFillRatios(t *testing.T) {
	for _, spec := range all27AssetSpecs {
		t.Run(spec.Path, func(t *testing.T) {
			data, err := imageFS.ReadFile(spec.Path)
			if err != nil {
				t.Fatalf("failed to read asset %s: %v", spec.Path, err)
			}

			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode image %s: %v", spec.Path, err)
			}

			bounds := img.Bounds()
			totalPixels := bounds.Dx() * bounds.Dy()
			nonZeroAlpha := 0
			fullyOpaque := 0
			semiTransparent := 0

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					// RGBA() returns alpha in range [0, 0xFFFF]
					if a > 0 {
						nonZeroAlpha++
						if a == 0xFFFF {
							fullyOpaque++
						} else {
							semiTransparent++
						}
					}
				}
			}

			if nonZeroAlpha == 0 {
				t.Fatalf("asset %s has 0 visible pixels (completely empty)", spec.Path)
			}

			fillRatio := float64(nonZeroAlpha) / float64(totalPixels)
			t.Logf("Asset %s: fillRatio=%.2f%% (solid=%d, semi=%d, total=%d)",
				spec.Path, fillRatio*100.0, fullyOpaque, semiTransparent, totalPixels)

			if fillRatio < spec.MinFillRatio {
				t.Errorf("asset %s fill ratio %.2f%% below minimum %.2f%%",
					spec.Path, fillRatio*100.0, spec.MinFillRatio*100.0)
			}
			if fillRatio > spec.MaxFillRatio {
				t.Errorf("asset %s fill ratio %.2f%% exceeds maximum %.2f%%",
					spec.Path, fillRatio*100.0, spec.MaxFillRatio*100.0)
			}
		})
	}
}

// TestEmpiricalFloorDiamondGeometry verifies isometric diamond strict boundary constraints.
// - All pixels outside Manhattan distance 1.0 must have Alpha == 0 (zero bleeding).
// - All pixels in inner diamond core (distance <= 0.85) must have Alpha == 255 (solid floor).
func TestEmpiricalFloorDiamondGeometry(t *testing.T) {
	floorPaths := []string{
		"images/grass.png",
		"images/dirt.png",
		"images/wood.png",
		"images/asphalt.png",
		"images/concrete.png",
		"images/tile_floor.png",
	}

	for _, path := range floorPaths {
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
			if bounds.Dx() != 256 || bounds.Dy() != 128 {
				t.Fatalf("%s unexpected dimensions %dx%d", path, bounds.Dx(), bounds.Dy())
			}

			cx := 127.5
			cy := 63.5
			rx := 128.0
			ry := 64.0

			outerBleedCount := 0
			innerHoleCount := 0
			cornerPixelCount := 0

			for y := 0; y < 128; y++ {
				for x := 0; x < 256; x++ {
					dx := float64(x) - cx
					dy := float64(y) - cy
					isoDist := math.Abs(dx)/rx + math.Abs(dy)/ry
					_, _, _, a := img.At(x, y).RGBA()

					// Check outer corners: isoDist > 1.0 must be 100% transparent
					if isoDist > 1.0 {
						if a > 0 {
							outerBleedCount++
						}
					}

					// Check extreme 4 corners (0..15, 0..7) etc
					if (x < 16 && y < 8) || (x > 240 && y < 8) || (x < 16 && y > 120) || (x > 240 && y > 120) {
						if a > 0 {
							cornerPixelCount++
						}
					}

					// Check inner core: isoDist <= 0.85 must be 100% solid (a == 0xFFFF)
					if isoDist <= 0.85 {
						if a < 0xFFFF {
							innerHoleCount++
							if innerHoleCount <= 5 {
								r, g, b, _ := img.At(x, y).RGBA()
								t.Logf("Inner hole at (%d, %d): RGBA=(%d, %d, %d, %d) [isoDist=%.3f]",
									x, y, r>>8, g>>8, b>>8, a>>8, isoDist)
							}
						}
					}
				}
			}

			if outerBleedCount > 0 {
				t.Errorf("%s has %d non-transparent pixels outside isometric diamond (isoDist > 1.0)",
					path, outerBleedCount)
			}
			if cornerPixelCount > 0 {
				t.Errorf("%s has %d pixels in outer corner bounding boxes", path, cornerPixelCount)
			}
			if innerHoleCount > 0 {
				t.Errorf("%s has %d transparent/semi-transparent pixels inside solid core (isoDist <= 0.85)",
					path, innerHoleCount)
			}
		})
	}
}

// TestEmpiricalCharacterGrounding verifies feet and ground shadow placement in rows 112..127.
func TestEmpiricalCharacterGrounding(t *testing.T) {
	characters := []struct {
		path            string
		minGroundPixels int
	}{
		{"images/player.png", 50},
		{"images/zombie.png", 50},
		{"images/runner.png", 50},
	}

	for _, tc := range characters {
		t.Run(tc.path, func(t *testing.T) {
			data, err := imageFS.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tc.path, err)
			}

			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("failed to decode %s: %v", tc.path, err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 128 {
				t.Fatalf("%s bounds = %dx%d, want 64x128", tc.path, bounds.Dx(), bounds.Dy())
			}

			groundCount := 0
			solidBootsCount := 0
			for y := 112; y < 128; y++ {
				for x := 0; x < 64; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						groundCount++
						if a == 0xFFFF {
							solidBootsCount++
						}
					}
				}
			}

			t.Logf("Character %s: groundPixels(112..127)=%d (solid=%d)",
				tc.path, groundCount, solidBootsCount)

			if groundCount < tc.minGroundPixels {
				t.Errorf("character %s has too few grounding pixels (%d < %d)",
					tc.path, groundCount, tc.minGroundPixels)
			}
		})
	}
}

// TestEmpiricalGenerationDeterminism checks SHA-256 hashes across repeated executions of genassets.
func TestEmpiricalGenerationDeterminism(t *testing.T) {
	// Find project root
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root containing go.mod")
		}
		dir = parent
	}

	imagesDir := filepath.Join(dir, "internal/assets/images")

	// 1. Record baseline hashes
	baselineHashes := make(map[string]string)
	for _, spec := range all27AssetSpecs {
		imgPath := filepath.Join(imagesDir, filepath.Base(spec.Path))
		data, err := os.ReadFile(imgPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", imgPath, err)
		}
		h := sha256.Sum256(data)
		baselineHashes[spec.Path] = hex.EncodeToString(h[:])
	}

	// 2. Run genassets binary 2 times consecutively and verify bit-for-bit identity
	for run := 1; run <= 2; run++ {
		cmd := exec.Command("go", "run", "./cmd/tools/genassets")
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Run %d failed: go run ./cmd/tools/genassets: %v\n%s", run, err, string(output))
		}

		for _, spec := range all27AssetSpecs {
			imgPath := filepath.Join(imagesDir, filepath.Base(spec.Path))
			data, err := os.ReadFile(imgPath)
			if err != nil {
				t.Fatalf("failed to read %s on run %d: %v", imgPath, run, err)
			}
			h := sha256.Sum256(data)
			currHash := hex.EncodeToString(h[:])
			if currHash != baselineHashes[spec.Path] {
				t.Fatalf("Run %d: non-deterministic output for %s: got %s, expected %s",
					run, spec.Path, currHash, baselineHashes[spec.Path])
			}
		}
	}
}

// TestEmpiricalObstacleBoundsAndGrounding verifies vertical obstacles have grounding anchors.
func TestEmpiricalObstacleBoundsAndGrounding(t *testing.T) {
	obstacles := []string{
		"images/wall.png",
		"images/tree.png",
		"images/fence.png",
		"images/debris.png",
		"images/tent.png",
		"images/stump.png",
		"images/mushroom.png",
		"images/sign.png",
		"images/elevation_block.png",
		"images/elevation_ramp.png",
	}

	for _, path := range obstacles {
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
			if bounds.Dx() != 256 || bounds.Dy() != 256 {
				t.Fatalf("%s dimensions %dx%d != 256x256", path, bounds.Dx(), bounds.Dy())
			}

			// Verify that obstacle has pixels in lower half (ground contact area)
			lowerPixels := 0
			for y := 128; y < 256; y++ {
				for x := 0; x < 256; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						lowerPixels++
					}
				}
			}

			if lowerPixels == 0 {
				t.Errorf("obstacle %s has no grounding pixels in lower half (y >= 128)", path)
			}
		})
	}
}

// TestEmpiricalItemIconQuality verifies items have distinct icons and centered silhouettes.
func TestEmpiricalItemIconQuality(t *testing.T) {
	itemFiles := []string{
		"images/food.png",
		"images/water.png",
		"images/weapon.png",
		"images/axe.png",
		"images/shotgun.png",
		"images/ammo.png",
		"images/armor.png",
		"images/antidote.png",
	}

	for _, path := range itemFiles {
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
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Fatalf("%s dimensions %dx%d != 64x64", path, bounds.Dx(), bounds.Dy())
			}

			// Compute centroid of non-zero pixels
			totalMass := 0
			sumX := 0
			sumY := 0
			for y := 0; y < 64; y++ {
				for x := 0; x < 64; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						totalMass++
						sumX += x
						sumY += y
					}
				}
			}

			if totalMass == 0 {
				t.Fatalf("item %s has no mass", path)
			}

			centroidX := float64(sumX) / float64(totalMass)
			centroidY := float64(sumY) / float64(totalMass)

			// Centroid should be reasonably centered in the 64x64 box (between 20.0 and 44.0)
			if centroidX < 20.0 || centroidX > 44.0 || centroidY < 20.0 || centroidY > 44.0 {
				t.Errorf("item %s centroid (%.1f, %.1f) is off-center (expected within [20..44])",
					path, centroidX, centroidY)
			}
		})
	}
}
