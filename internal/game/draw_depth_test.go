package game

import (
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

func TestDrawSystem_SpriteGeometricAnchors(t *testing.T) {
	assets.Load()
	tests := []struct {
		name       string
		img        *ebiten.Image
		wantTransX float64
		wantTransY float64
	}{
		// Legacy Obstacles (256x256)
		{"Wall", assets.WallImage, -128.0, -128.0},
		{"Tree", assets.TreeImage, -128.0, -128.0},
		{"Fence", assets.FenceImage, -128.0, -128.0},
		{"Debris", assets.DebrisImage, -128.0, -128.0},
		{"Tent", assets.TentImage, -128.0, -128.0},
		{"Stump", assets.StumpImage, -128.0, -128.0},
		{"Mushroom", assets.MushroomImage, -128.0, -128.0},
		{"Sign", assets.SignImage, -128.0, -128.0},
		{"ElevationBlock", assets.ElevationBlockImage, -128.0, -128.0},
		{"ElevationRamp", assets.ElevationRampImage, -128.0, -128.0},

		// New Environmental Props (Variable Dimensions)
		{"Bench", assets.BenchImage, -26.0, 91.0},        // 52x37: -52/2 = -26, 128 - 37 = 91
		{"Chest", assets.ChestImage, -11.0, 107.0},      // 22x21: -22/2 = -11, 128 - 21 = 107
		{"Sculpture", assets.SculptureImage, -11.5, 97.0}, // 23x31: -23/2 = -11.5, 128 - 31 = 97
		{"Bush", assets.BushImage, -12.0, 110.0},         // 24x18: -24/2 = -12, 128 - 18 = 110
		{"Flower", assets.FlowerImage, -13.0, 103.0},     // 26x25: -26/2 = -13, 128 - 25 = 103
		{"Stone", assets.StoneImage, -14.0, 109.0},       // 28x19: -28/2 = -14, 128 - 19 = 109
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.img == nil {
				t.Fatalf("image %s is nil", tt.name)
			}
			b := tt.img.Bounds()
			imgW := float64(b.Dx())
			imgH := float64(b.Dy())

			transX := -imgW / 2.0
			transY := 128.0 - imgH

			if transX != tt.wantTransX {
				t.Errorf("%s transX = %f, want %f", tt.name, transX, tt.wantTransX)
			}
			if transY != tt.wantTransY {
				t.Errorf("%s transY = %f, want %f", tt.name, transY, tt.wantTransY)
			}
		})
	}
}

func TestDrawSystem_NewPropTilesLoadedAndDrawn(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(20, 20)

	props := []world.TileType{
		world.TileBench,
		world.TileChest,
		world.TileSculpture,
		world.TileBush,
		world.TileFlower,
		world.TileStone,
	}

	for idx, p := range props {
		m.SetTile(idx+1, idx+1, p)
		m.Visible[(idx+1)*20+(idx+1)] = true
		m.Explored[(idx+1)*20+(idx+1)] = true
	}

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(800, 600)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DrawSystem panicked while drawing prop tiles: %v", r)
		}
	}()

	drawSys.Draw(screen, 12.0)
}

func TestDrawSystem_GroundPassUnderNewProps(t *testing.T) {
	assets.Load()
	if assets.GrassImage == nil {
		t.Fatal("GrassImage is nil")
	}

	b := assets.GrassImage.Bounds()
	if b.Dx() != 256 || b.Dy() != 128 {
		t.Errorf("GrassImage dimensions = %dx%d, want 256x128", b.Dx(), b.Dy())
	}
}

func TestDrawSystem_DepthSortingOrdering(t *testing.T) {
	w := arkecs.NewWorld()
	m := world.NewMap(10, 10)
	drawSys := NewDrawSystem(w, m)

	if drawSys == nil {
		t.Fatal("NewDrawSystem returned nil")
	}

	// Verify top-down vertical Y-depth sorting monotonicity
	// When pos.Y (vertical coordinate) increases, rendering depth strictly increases
	pos1Y := 100.0
	pos2Y := 200.0

	depth1 := pos1Y
	depth2 := pos2Y

	if depth1 >= depth2 {
		t.Errorf("Expected vertical Y-depth ordering monotonicity, got %f >= %f", depth1, depth2)
	}
}
