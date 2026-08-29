package game

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestOrthogonal_CoordinateTransformations(t *testing.T) {
	// 1. Identity transform check for backward compatibility
	wx, wy := 123.45, 678.90
	isoX, isoY := WorldToIso(wx, wy)
	if isoX != wx || isoY != wy {
		t.Errorf("WorldToIso(%f, %f) = (%f, %f), want identity (%f, %f)", wx, wy, isoX, isoY, wx, wy)
	}

	backWx, backWy := IsoToWorld(isoX, isoY)
	if backWx != wx || backWy != wy {
		t.Errorf("IsoToWorld(%f, %f) = (%f, %f), want identity (%f, %f)", isoX, isoY, backWx, backWy, wx, wy)
	}

	// 2. Center projection check
	camX, camY := 500.0, 500.0
	sx, sy := WorldToScreen(500.0, 500.0, camX, camY)
	if sx != 640.0 || sy != 360.0 {
		t.Errorf("WorldToScreen at camera center = (%f, %f), want (640, 360)", sx, sy)
	}

	unprojX, unprojY := ScreenToWorld(640.0, 360.0, camX, camY)
	if math.Abs(unprojX-500.0) > 1e-9 || math.Abs(unprojY-500.0) > 1e-9 {
		t.Errorf("ScreenToWorld(640, 360) = (%f, %f), want (500, 500)", unprojX, unprojY)
	}

	// 3. Extensive fuzzing round-trip verification across 10,000 random points
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10000; i++ {
		testWx := (rng.Float64() - 0.5) * 20000.0
		testWy := (rng.Float64() - 0.5) * 20000.0
		testCamX := (rng.Float64() - 0.5) * 20000.0
		testCamY := (rng.Float64() - 0.5) * 20000.0

		scrX, scrY := WorldToScreen(testWx, testWy, testCamX, testCamY)
		recWx, recWy := ScreenToWorld(scrX, scrY, testCamX, testCamY)

		if math.Abs(recWx-testWx) > 1e-9 || math.Abs(recWy-testWy) > 1e-9 {
			t.Fatalf("Roundtrip error at (%f, %f) under cam (%f, %f): recovered (%f, %f)",
				testWx, testWy, testCamX, testCamY, recWx, recWy)
		}
	}
}

func TestOrthogonal_CameraCartesianTracking(t *testing.T) {
	cam := NewCamera()
	if cam.Initialized {
		t.Errorf("New camera should not be initialized")
	}

	// Snap directly to Cartesian position
	cam.Snap(1200.0, 800.0)
	if !cam.Initialized || cam.X != 1200.0 || cam.Y != 800.0 || cam.TargetX != 1200.0 || cam.TargetY != 800.0 {
		t.Errorf("Camera snap failed: %+v", cam)
	}

	// Update towards target
	cam.Update(1300.0, 900.0)
	expectedX := 1200.0 + (1300.0-1200.0)*0.10
	expectedY := 800.0 + (900.0-800.0)*0.10
	if math.Abs(cam.X-expectedX) > 1e-9 || math.Abs(cam.Y-expectedY) > 1e-9 {
		t.Errorf("Camera update mismatch: got (%f, %f), want (%f, %f)", cam.X, cam.Y, expectedX, expectedY)
	}
}

func TestOrthogonal_SeamlessTileAdjacency(t *testing.T) {
	camX, camY := 1000.0, 1000.0

	for tx := 0; tx < 50; tx++ {
		for ty := 0; ty < 50; ty++ {
			// Tile (tx, ty)
			w0x := float64(tx * world.TileSize)
			w0y := float64(ty * world.TileSize)
			s0x, s0y := WorldToScreen(w0x, w0y, camX, camY)

			// Right edge of Tile (tx, ty)
			rightEdge := s0x + float64(world.TileSize)*DefaultZoom
			// Left edge of Tile (tx+1, ty)
			w1x := float64((tx + 1) * world.TileSize)
			s1x, _ := WorldToScreen(w1x, w0y, camX, camY)

			if math.Abs(rightEdge-s1x) > 1e-9 {
				t.Fatalf("Horizontal tile gap detected between tile (%d,%d) and (%d,%d): rightEdge=%f, leftEdge=%f",
					tx, ty, tx+1, ty, rightEdge, s1x)
			}

			// Bottom edge of Tile (tx, ty)
			bottomEdge := s0y + float64(world.TileSize)*DefaultZoom
			// Top edge of Tile (tx, ty+1)
			w1y := float64((ty + 1) * world.TileSize)
			_, s1y := WorldToScreen(w0x, w1y, camX, camY)

			if math.Abs(bottomEdge-s1y) > 1e-9 {
				t.Fatalf("Vertical tile gap detected between tile (%d,%d) and (%d,%d): bottomEdge=%f, topEdge=%f",
					tx, ty, tx, ty+1, bottomEdge, s1y)
			}
		}
	}
}

func TestOrthogonal_TopDownYDepthSorting(t *testing.T) {
	type Renderable struct {
		ID    int
		Depth float64
	}

	items := []Renderable{
		{ID: 1, Depth: 500.0}, // South
		{ID: 2, Depth: 100.0}, // North
		{ID: 3, Depth: 300.0}, // Middle
		{ID: 4, Depth: 250.0}, // North-middle
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Depth < items[j].Depth
	})

	expectedOrder := []int{2, 4, 3, 1}
	for i, it := range items {
		if it.ID != expectedOrder[i] {
			t.Errorf("Depth sort mismatch at index %d: got ID %d, want %d", i, it.ID, expectedOrder[i])
		}
	}
}

func TestOrthogonal_GameResetAndHeadlessDraw(t *testing.T) {
	assets.Load()
	g := NewGame()
	if g == nil {
		t.Fatalf("NewGame returned nil")
	}

	// Verify camera snapped to player spawn
	if math.Abs(g.camera.X-g.gameMap.PlayerSpawn.X) > 1e-9 || math.Abs(g.camera.Y-g.gameMap.PlayerSpawn.Y) > 1e-9 {
		t.Errorf("Camera after Reset = (%f, %f), want PlayerSpawn (%f, %f)",
			g.camera.X, g.camera.Y, g.gameMap.PlayerSpawn.X, g.gameMap.PlayerSpawn.Y)
	}

	// Update game ticks
	for tick := 0; tick < 60; tick++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Update error at tick %d: %v", tick, err)
		}
	}

	// Headless draw to 1280x720 surface
	screen := ebiten.NewImage(1280, 720)
	g.Draw(screen)
}
