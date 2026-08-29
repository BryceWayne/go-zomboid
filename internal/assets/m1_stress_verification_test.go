package assets

import (
	"bytes"
	"image"
	_ "image/png"
	"io/fs"
	"runtime"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestChallenger_MassiveConcurrentLoadStress launches hundreds of goroutines
// aggressively calling Load() and asserting on all 49 image pointers simultaneously.
func TestChallenger_MassiveConcurrentLoadStress(t *testing.T) {
	const numGoroutines = 200
	const iterations = 100

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			<-start

			for j := 0; j < iterations; j++ {
				Load()

				// Sample checks on pointers across all categories
				if routineID%4 == 0 {
					if PlayerImage == nil || ZombieImage == nil || RunnerImage == nil {
						t.Errorf("entity pointer nil in routine %d", routineID)
					}
				}
				if routineID%4 == 1 {
					if GrassImage == nil || WallImage == nil || TreeImage == nil {
						t.Errorf("world pointer nil in routine %d", routineID)
					}
				}
				if routineID%4 == 2 {
					if WeaponImage == nil || ArmorImage == nil || AxeImage == nil {
						t.Errorf("item pointer nil in routine %d", routineID)
					}
				}
				if routineID%4 == 3 {
					if BenchImage == nil || ChestImage == nil || SculptureImage == nil ||
						LabTilesetImage == nil || ZombieTilesetImage == nil {
						t.Errorf("external pointer nil in routine %d", routineID)
					}
				}
			}
		}(i)
	}

	close(start)
	wg.Wait()

	// Final verification of all 49 pointers
	pointers := []*ebiten.Image{
		PlayerImage, ZombieImage, RunnerImage,
		GrassImage, DirtImage, WoodImage, AsphaltImage, ConcreteImage, TileFloorImage,
		WallImage, TreeImage, FenceImage, DebrisImage, TentImage, StumpImage, MushroomImage, SignImage, ElevationBlockImage, ElevationRampImage,
		WeaponImage, AxeImage, ShotgunImage, AmmoImage, ArmorImage, AntidoteImage, FoodImage, WaterImage,
		BenchImage, ChestImage, Sculpture1Image, Sculpture2Image, SculptureImage,
		Bush1Image, Bush2Image, Bush3Image, Bush4Image, BushImage,
		Flower1Image, Flower2Image, Flower3Image, FlowerImage,
		Stone1Image, Stone2Image, StoneImage, ForestStumpImage, GrassTuft1Image, GrassTuft2Image,
		LabTilesetImage, ZombieTilesetImage,
	}

	for idx, ptr := range pointers {
		if ptr == nil {
			t.Fatalf("pointer index %d is nil after massive concurrent load stress", idx)
		}
		b := ptr.Bounds()
		if b.Dx() <= 0 || b.Dy() <= 0 {
			t.Fatalf("pointer index %d has invalid bounds %v", idx, b)
		}
	}
}

// TestChallenger_ParallelDecodeAll606Images decodes all 606 embedded images
// using a worker pool matching GOMAXPROCS to stress test PNG decoders.
func TestChallenger_ParallelDecodeAll606Images(t *testing.T) {
	var paths []string
	err := fs.WalkDir(imageFS, "images", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk imageFS: %v", err)
	}

	if len(paths) != 606 {
		t.Fatalf("expected exactly 606 paths in imageFS, found %d", len(paths))
	}

	numWorkers := runtime.GOMAXPROCS(0) * 4
	if numWorkers < 8 {
		numWorkers = 8
	}

	pathCh := make(chan string, len(paths))
	for _, p := range paths {
		pathCh <- p
	}
	close(pathCh)

	var wg sync.WaitGroup
	errCh := make(chan error, len(paths))

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathCh {
				data, err := imageFS.ReadFile(path)
				if err != nil {
					errCh <- err
					return
				}

				img, format, err := image.Decode(bytes.NewReader(data))
				if err != nil {
					errCh <- err
					return
				}
				if format != "png" {
					t.Errorf("path %s format %s != png", path, format)
				}
				b := img.Bounds()
				if b.Dx() <= 0 || b.Dy() <= 0 {
					t.Errorf("path %s has non-positive dimensions %v", path, b)
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("error during parallel decoding: %v", err)
	}
}
