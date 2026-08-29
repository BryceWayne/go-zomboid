package game

import (
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// 1. Unit Test: Camera Snap and Initialization State
func TestCamera_SnapAndInitialization(t *testing.T) {
	cam := NewCamera()
	if cam.Initialized {
		t.Fatal("New camera should not be initialized before first snap or update")
	}
	if cam.LerpFactor != 0.10 {
		t.Errorf("Expected default LerpFactor = 0.10, got %f", cam.LerpFactor)
	}

	cam.Snap(1280.0, 640.0)
	if !cam.Initialized {
		t.Fatal("Camera should be marked Initialized after Snap")
	}
	if cam.X != 1280.0 || cam.Y != 640.0 {
		t.Errorf("Camera position mismatch: got (%f, %f), want (1280.0, 640.0)", cam.X, cam.Y)
	}
	if cam.TargetX != 1280.0 || cam.TargetY != 640.0 {
		t.Errorf("Camera target mismatch: got (%f, %f), want (1280.0, 640.0)", cam.TargetX, cam.TargetY)
	}
}

// 2. Unit Test: Camera Update on Uninitialized Camera Auto-Snaps
func TestCamera_UpdateUninitializedSnaps(t *testing.T) {
	cam := NewCamera()
	cam.Update(500.0, 300.0)

	if !cam.Initialized {
		t.Fatal("Camera should be initialized on first Update call")
	}
	if cam.X != 500.0 || cam.Y != 300.0 {
		t.Errorf("Uninitialized update should snap to target: got (%f, %f), want (500.0, 300.0)", cam.X, cam.Y)
	}
}

// 3. Unit Test: Smooth Exponential Convergence Dynamics (LerpFactor = 0.10)
func TestCamera_UpdateExponentialConvergence(t *testing.T) {
	cam := NewCamera()
	cam.Snap(0.0, 0.0)

	targetX, targetY := 800.0, 600.0 // Exactly 1000.0px Euclidean distance
	initialDist := math.Hypot(targetX, targetY)

	prevDist := initialDist
	for frame := 1; frame <= 60; frame++ {
		cam.Update(targetX, targetY)

		curDist := math.Hypot(targetX-cam.X, targetY-cam.Y)
		expectedDist := initialDist * math.Pow(1.0-cam.LerpFactor, float64(frame))

		// Check exponential decay within floating point tolerance
		if math.Abs(curDist-expectedDist) > 1e-4 {
			t.Fatalf("Frame %d: Exponential decay mismatch. got %f, want %f", frame, curDist, expectedDist)
		}

		if curDist >= prevDist {
			t.Fatalf("Frame %d: Camera distance did not decrease (prev: %f, cur: %f)", frame, prevDist, curDist)
		}
		prevDist = curDist
	}

	// After 60 frames (~1 second at 60 TPS), 1000px distance reduced to 1000 * 0.9^60 = ~1.797px (< 2.0px)
	if prevDist > 2.0 {
		t.Errorf("Expected distance after 60 frames < 2.0px, got %f", prevDist)
	}
}

// 4. Unit Test: Sub-Pixel Snap Prevents Floating-Point Asymptotic Oscillation
func TestCamera_UpdateSubpixelSnap(t *testing.T) {
	cam := NewCamera()
	cam.Snap(100.0, 100.0)

	// Target 0.005px away (less than 0.01 threshold)
	cam.Update(100.005, 100.004)

	if cam.X != 100.005 || cam.Y != 100.004 {
		t.Errorf("Expected sub-pixel snap to exact target (100.005, 100.004), got (%f, %f)", cam.X, cam.Y)
	}
}

// 5. Unit Test: Mathematical Invertibility & Bijectivity (ScreenToIso & ScreenToWorld)
func TestCamera_ScreenToIsoAndScreenToWorldRoundtrip(t *testing.T) {
	r := rand.New(rand.NewSource(13579))

	for i := 0; i < 5000; i++ {
		// Random world coordinates in [-10000, 10000]
		wx := (r.Float64() - 0.5) * 20000.0
		wy := (r.Float64() - 0.5) * 20000.0

		// Random camera position in [-10000, 10000]
		camX := (r.Float64() - 0.5) * 20000.0
		camY := (r.Float64() - 0.5) * 20000.0

		// Forward projection to Screen space
		isoX, isoY := WorldToIso(wx, wy)
		sx := (isoX-camX)*0.5 + 640.0
		sy := (isoY-camY)*0.5 + 360.0

		// Inverse projection from Screen space
		recoveredIsoX, recoveredIsoY := ScreenToIso(sx, sy, camX, camY)
		recoveredWx, recoveredWy := ScreenToWorld(sx, sy, camX, camY)

		const eps = 1e-9
		if math.Abs(recoveredIsoX-isoX) > eps || math.Abs(recoveredIsoY-isoY) > eps {
			t.Fatalf("Iteration %d: ScreenToIso mismatch. Got (%f, %f), want (%f, %f)", i, recoveredIsoX, recoveredIsoY, isoX, isoY)
		}
		if math.Abs(recoveredWx-wx) > eps || math.Abs(recoveredWy-wy) > eps {
			t.Fatalf("Iteration %d: ScreenToWorld mismatch. Got (%f, %f), want (%f, %f)", i, recoveredWx, recoveredWy, wx, wy)
		}
	}
}

// 6. Unit Test: Screen Center (640, 360) Invariance
func TestCamera_ScreenCenterInvariance(t *testing.T) {
	r := rand.New(rand.NewSource(24680))

	for i := 0; i < 1000; i++ {
		playerWx := (r.Float64() - 0.5) * 5000.0
		playerWy := (r.Float64() - 0.5) * 5000.0

		camIsoX, camIsoY := WorldToIso(playerWx, playerWy)

		// When clicking screen center (640, 360), unprojected coordinate must match player world position
		clickWx, clickWy := ScreenToWorld(640.0, 360.0, camIsoX, camIsoY)

		const eps = 1e-9
		if math.Abs(clickWx-playerWx) > eps || math.Abs(clickWy-playerWy) > eps {
			t.Fatalf("Iteration %d: Center click failed to unproject to player position. Got (%f, %f), want (%f, %f)",
				i, clickWx, clickWy, playerWx, playerWy)
		}
	}
}

// 7. Unit Test: Shared Camera Instance Synchronization in Game.Reset
func TestCamera_GameResetSharedInstance(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()

	if g.camera == nil {
		t.Fatal("Game.camera must not be nil")
	}
	if g.updateSys.camera == nil {
		t.Fatal("UpdateSystem.camera must not be nil")
	}
	if g.drawSys.camera == nil {
		t.Fatal("DrawSystem.camera must not be nil")
	}

	// Verify exact pointer equality (shared instance)
	if g.camera != g.updateSys.camera {
		t.Fatal("Game.camera and UpdateSystem.camera must reference the identical *Camera pointer")
	}
	if g.camera != g.drawSys.camera {
		t.Fatal("Game.camera and DrawSystem.camera must reference the identical *Camera pointer")
	}

	// Verify camera was snapped to player spawn
	spawnIsoX, spawnIsoY := WorldToIso(g.gameMap.PlayerSpawn.X, g.gameMap.PlayerSpawn.Y)
	if math.Abs(g.camera.X-spawnIsoX) > 1e-6 || math.Abs(g.camera.Y-spawnIsoY) > 1e-6 {
		t.Errorf("Camera not snapped to player spawn: got (%f, %f), want (%f, %f)", g.camera.X, g.camera.Y, spawnIsoX, spawnIsoY)
	}
	if !g.camera.Initialized {
		t.Fatal("Camera should be initialized on Game.Reset")
	}
}

// 8. Unit Test: FOV Radius Raycasting Expansion (22 Tiles)
func TestCamera_FOVExpandedRadius(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(60, 60)
	sys := NewUpdateSystem(w, m)
	cam := NewCamera()
	sys.camera = cam

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{},
		&ecs.Position{X: 30 * world.TileSize, Y: 30 * world.TileSize},
		&ecs.Velocity{},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 64, H: 128},
		&ecs.Collider{Width: 64, Height: 64},
	)

	sys.Update(-1)

	// Clear out any walls on horizontal test line to check raycast distance
	for tx := 30; tx <= 50; tx++ {
		m.SetTile(tx, 30, world.TileGrass)
	}

	// Re-run update to cast unobstructed rays
	sys.Update(-1)

	// With 22 tiles FOV radius, tile at distance 20 (tx = 50, ty = 30) must be visible and explored
	idx := 30*m.Width + 50
	if !m.Visible[idx] || !m.Explored[idx] {
		t.Errorf("Tile at distance 20 (idx %d) should be visible and explored with expanded 22-tile FOV", idx)
	}
}

// 9. Unit Test: Headless DrawSystem.Draw with Camera at 1280x720
func TestCamera_HeadlessDrawExecution(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()

	screen := ebiten.NewImage(1280, 720)

	// Simulate 10 frames of Update + Draw
	for frame := 0; frame < 10; frame++ {
		err := g.Update()
		if err != nil {
			t.Fatalf("Game.Update failed on frame %d: %v", frame, err)
		}
		g.Draw(screen)
	}
}

// 10. Unit Test: Viewport Four Corners Mathematical Symmetry
func TestCamera_ViewportCornersUnprojection(t *testing.T) {
	camIsoX, camIsoY := 1000.0, 500.0

	// Corners of 1280x720 viewport
	corners := []struct {
		name string
		sx   float64
		sy   float64
	}{
		{"TopLeft", 0.0, 0.0},
		{"TopRight", 1280.0, 0.0},
		{"BottomLeft", 0.0, 720.0},
		{"BottomRight", 1280.0, 720.0},
	}

	for _, c := range corners {
		t.Run(c.name, func(t *testing.T) {
			isoX, isoY := ScreenToIso(c.sx, c.sy, camIsoX, camIsoY)
			deltaIsoX := isoX - camIsoX
			deltaIsoY := isoY - camIsoY

			// sx = 0 -> deltaIsoX = (0 - 640)/0.5 = -1280
			// sx = 1280 -> deltaIsoX = (1280 - 640)/0.5 = +1280
			// sy = 0 -> deltaIsoY = (0 - 360)/0.5 = -720
			// sy = 720 -> deltaIsoY = (720 - 360)/0.5 = +720
			expectedDeltaIsoX := (c.sx - 640.0) / 0.5
			expectedDeltaIsoY := (c.sy - 360.0) / 0.5

			if math.Abs(deltaIsoX-expectedDeltaIsoX) > 1e-9 {
				t.Errorf("deltaIsoX mismatch: got %f, want %f", deltaIsoX, expectedDeltaIsoX)
			}
			if math.Abs(deltaIsoY-expectedDeltaIsoY) > 1e-9 {
				t.Errorf("deltaIsoY mismatch: got %f, want %f", deltaIsoY, expectedDeltaIsoY)
			}

			// Forward projection must return exact original screen corner
			fwdSx := deltaIsoX*0.5 + 640.0
			fwdSy := deltaIsoY*0.5 + 360.0
			if math.Abs(fwdSx-c.sx) > 1e-9 || math.Abs(fwdSy-c.sy) > 1e-9 {
				t.Errorf("Forward projection roundtrip failed: got (%f, %f), want (%f, %f)", fwdSx, fwdSy, c.sx, c.sy)
			}
		})
	}
}

// 11. Unit Test: Dynamic Tracking Lag and Catchup
func TestCamera_DynamicTrackingLagAndCatchup(t *testing.T) {
	cam := NewCamera()
	cam.Snap(0.0, 0.0)

	// Simulate player moving at 12px/frame along X axis for 60 frames
	playerX := 0.0
	playerY := 0.0
	const speed = 12.0

	for frame := 0; frame < 60; frame++ {
		playerX += speed
		targetIsoX, targetIsoY := WorldToIso(playerX, playerY)
		cam.Update(targetIsoX, targetIsoY)

		// During constant velocity movement, camera smoothly trails target
		if cam.X >= targetIsoX {
			t.Errorf("Frame %d: Camera (%f) should lag behind moving target (%f)", frame, cam.X, targetIsoX)
		}
	}

	// Player stops moving; camera must catch up
	stopIsoX, stopIsoY := WorldToIso(playerX, playerY)
	for frame := 0; frame < 60; frame++ {
		cam.Update(stopIsoX, stopIsoY)
	}

	lagAfterStop := math.Hypot(stopIsoX-cam.X, stopIsoY-cam.Y)
	if lagAfterStop > 1.0 {
		t.Errorf("Camera did not catch up within 60 frames after stopping: remaining lag = %f px", lagAfterStop)
	}
}

// 12. Unit Test: Tile Click Movement Targeting Accuracy
func TestCamera_TileClickMovementTargetingAccuracy(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(60, 60)
	sys := NewUpdateSystem(w, m)
	cam := NewCamera()
	sys.camera = cam

	playerStartWx := 20.0 * float64(world.TileSize)
	playerStartWy := 20.0 * float64(world.TileSize)
	cam.Snap(WorldToIso(playerStartWx, playerStartWy))

	// Click on destination tile (25, 20)
	destWx := 25.0 * float64(world.TileSize)
	destWy := 20.0 * float64(world.TileSize)
	destIsoX, destIsoY := WorldToIso(destWx, destWy)

	// Compute screen coordinates where that tile is rendered
	destScreenX := (destIsoX-cam.X)*0.5 + 640.0
	destScreenY := (destIsoY-cam.Y)*0.5 + 360.0

	// Unproject screen coordinates back
	unprojectedWx, unprojectedWy := ScreenToWorld(destScreenX, destScreenY, cam.X, cam.Y)

	if math.Abs(unprojectedWx-destWx) > 1e-9 || math.Abs(unprojectedWy-destWy) > 1e-9 {
		t.Fatalf("Unprojected click position (%f, %f) does not match destination tile (%f, %f)",
			unprojectedWx, unprojectedWy, destWx, destWy)
	}
}

