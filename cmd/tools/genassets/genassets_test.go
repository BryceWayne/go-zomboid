package main

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var expectedAssetFiles = []struct {
	name   string
	width  int
	height int
	isIso  bool
}{
	// Character Entities (64x128)
	{"player.png", 64, 128, false},
	{"zombie.png", 64, 128, false},
	{"runner.png", 64, 128, false},

	// Floor Tiles (256x128)
	{"grass.png", 256, 128, true},
	{"dirt.png", 256, 128, true},
	{"wood.png", 256, 128, true},
	{"asphalt.png", 256, 128, true},
	{"concrete.png", 256, 128, true},
	{"tile_floor.png", 256, 128, true},

	// Vertical Obstacles & Props (256x256)
	{"wall.png", 256, 256, false},
	{"tree.png", 256, 256, false},
	{"fence.png", 256, 256, false},
	{"debris.png", 256, 256, false},
	{"tent.png", 256, 256, false},
	{"stump.png", 256, 256, false},
	{"mushroom.png", 256, 256, false},
	{"sign.png", 256, 256, false},
	{"elevation_block.png", 256, 256, false},
	{"elevation_ramp.png", 256, 256, false},

	// Items & Equipment (64x64)
	{"food.png", 64, 64, false},
	{"water.png", 64, 64, false},
	{"weapon.png", 64, 64, false},
	{"axe.png", 64, 64, false},
	{"shotgun.png", 64, 64, false},
	{"ammo.png", 64, 64, false},
	{"armor.png", 64, 64, false},
	{"antidote.png", 64, 64, false},
}

func getFileHash(t *testing.T, path string) string {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file %s: %v", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("failed to compute hash for %s: %v", path, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func getProjectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root containing go.mod from %s", dir)
		}
		dir = parent
	}
}

func TestAssetRegenerationDeterminism(t *testing.T) {
	projRoot := getProjectRoot(t)
	assetsDir := filepath.Join(projRoot, "internal/assets/images")

	// Phase 1: Capture initial hashes
	initialHashes := make(map[string]string)
	for _, asset := range expectedAssetFiles {
		path := filepath.Join(assetsDir, asset.name)
		initialHashes[asset.name] = getFileHash(t, path)
	}

	// Phase 2: Run generation multiple iterations
	for iter := 1; iter <= 3; iter++ {
		cmd := exec.Command("go", "run", "./cmd/tools/genassets")
		cmd.Dir = projRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Iteration %d: go run ./cmd/tools/genassets failed: %v\nOutput: %s", iter, err, string(output))
		}

		// Verify every asset matches initial hash
		for _, asset := range expectedAssetFiles {
			path := filepath.Join(assetsDir, asset.name)
			currentHash := getFileHash(t, path)
			if currentHash != initialHashes[asset.name] {
				t.Fatalf("Iteration %d: Non-deterministic output detected for %s: got hash %s, expected %s",
					iter, asset.name, currentHash, initialHashes[asset.name])
			}
		}
	}
}

func TestAssetDimensionsAndIntegrity(t *testing.T) {
	projRoot := getProjectRoot(t)
	assetsDir := filepath.Join(projRoot, "internal/assets/images")

	if len(expectedAssetFiles) != 27 {
		t.Fatalf("expected 27 assets to be registered, found %d", len(expectedAssetFiles))
	}

	for _, tc := range expectedAssetFiles {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(assetsDir, tc.name)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to open asset %s: %v", tc.name, err)
			}
			defer f.Close()

			img, format, err := image.Decode(f)
			if err != nil {
				t.Fatalf("failed to decode PNG %s: %v", tc.name, err)
			}
			if format != "png" {
				t.Fatalf("expected format png, got %s", format)
			}

			bounds := img.Bounds()
			if bounds.Dx() != tc.width || bounds.Dy() != tc.height {
				t.Fatalf("dimensions mismatch for %s: got %dx%d, want %dx%d",
					tc.name, bounds.Dx(), bounds.Dy(), tc.width, tc.height)
			}

			// Pixel transparency and density checks
			nonTransparent := 0
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						nonTransparent++
					}
				}
			}

			totalPixels := bounds.Dx() * bounds.Dy()
			if nonTransparent == 0 {
				t.Fatalf("asset %s has zero non-transparent pixels", tc.name)
			}

			// Minimum density check: at least 5% non-transparent
			fillRatio := float64(nonTransparent) / float64(totalPixels)
			if fillRatio < 0.05 {
				t.Errorf("asset %s fill ratio too low: %.2f%%", tc.name, fillRatio*100)
			}
		})
	}
}
