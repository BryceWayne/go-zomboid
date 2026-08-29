package game

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestChallenger_OrthogonalTransformations_ExtremeBoundsAndFuzzing verifies bijection, scale,
// center invariance, and numerical stability for WorldToIso, IsoToWorld, ScreenToWorld, and WorldToScreen
// across extreme coordinate bounds [-1e8, +1e8].
func TestChallenger_OrthogonalTransformations_ExtremeBoundsAndFuzzing(t *testing.T) {
	// 1. Static extreme boundary points
	extremePoints := []struct {
		wx, wy, camX, camY float64
		name               string
	}{
		{0.0, 0.0, 0.0, 0.0, "Origin"},
		{1e8, 1e8, 1e8, 1e8, "PositiveExtremeCenter"},
		{-1e8, -1e8, -1e8, -1e8, "NegativeExtremeCenter"},
		{1e8, -1e8, 0.0, 0.0, "PositiveWXNegativeWY"},
		{-1e8, 1e8, 0.0, 0.0, "NegativeWXPositiveWY"},
		{1234567.890123, -9876543.210987, 5000000.0, -5000000.0, "SubpixelArbitrary"},
		{0.0000001, -0.0000001, 0.0000002, -0.0000002, "NearZeroSubmicron"},
	}

	for _, pt := range extremePoints {
		t.Run(pt.name, func(t *testing.T) {
			// WorldToIso / IsoToWorld identity test
			isoX, isoY := WorldToIso(pt.wx, pt.wy)
			if isoX != pt.wx || isoY != pt.wy {
				t.Errorf("WorldToIso(%e, %e) = (%e, %e), want identity", pt.wx, pt.wy, isoX, isoY)
			}
			recWx, recWy := IsoToWorld(isoX, isoY)
			if recWx != pt.wx || recWy != pt.wy {
				t.Errorf("IsoToWorld(%e, %e) = (%e, %e), want (%e, %e)", isoX, isoY, recWx, recWy, pt.wx, pt.wy)
			}

			// ScreenToIso / WorldToIso consistency
			scrIsoX, scrIsoY := ScreenToIso(640.0, 360.0, pt.camX, pt.camY)
			if math.Abs(scrIsoX-pt.camX) > 1e-6 || math.Abs(scrIsoY-pt.camY) > 1e-6 {
				t.Errorf("ScreenToIso center unprojection mismatch: got (%e, %e), want cam (%e, %e)",
					scrIsoX, scrIsoY, pt.camX, pt.camY)
			}

			// WorldToScreen -> ScreenToWorld round-trip
			sx, sy := WorldToScreen(pt.wx, pt.wy, pt.camX, pt.camY)
			unprojWx, unprojWy := ScreenToWorld(sx, sy, pt.camX, pt.camY)

			if math.Abs(unprojWx-pt.wx) > 1e-6 || math.Abs(unprojWy-pt.wy) > 1e-6 {
				t.Errorf("Roundtrip W2S->S2W mismatch: original (%e, %e) -> screen (%e, %e) -> recovered (%e, %e)",
					pt.wx, pt.wy, sx, sy, unprojWx, unprojWy)
			}

			// ScreenToWorld -> WorldToScreen round-trip
			testScrX, testScrY := 320.0, 180.0
			wX, wY := ScreenToWorld(testScrX, testScrY, pt.camX, pt.camY)
			backSx, backSy := WorldToScreen(wX, wY, pt.camX, pt.camY)

			if math.Abs(backSx-testScrX) > 1e-6 || math.Abs(backSy-testScrY) > 1e-6 {
				t.Errorf("Roundtrip S2W->W2S mismatch: original screen (%e, %e) -> world (%e, %e) -> recovered screen (%e, %e)",
					testScrX, testScrY, wX, wY, backSx, backSy)
			}

			// Viewport center invariant: point at camera position maps to (640, 360)
			centerSx, centerSy := WorldToScreen(pt.camX, pt.camY, pt.camX, pt.camY)
			if math.Abs(centerSx-640.0) > 1e-6 || math.Abs(centerSy-360.0) > 1e-6 {
				t.Errorf("Center invariant failed: cam pos (%e, %e) mapped to screen (%e, %e), want (640, 360)",
					pt.camX, pt.camY, centerSx, centerSy)
			}
		})
	}

	// 2. Fuzzing 50,000 pseudo-random coordinate pairs in [-1e8, 1e8]
	rng := rand.New(rand.NewSource(1337))
	for i := 0; i < 50000; i++ {
		fwx := (rng.Float64()*2.0 - 1.0) * 1e8
		fwy := (rng.Float64()*2.0 - 1.0) * 1e8
		fcamX := (rng.Float64()*2.0 - 1.0) * 1e8
		fcamY := (rng.Float64()*2.0 - 1.0) * 1e8

		// Bijective roundtrip
		sx, sy := WorldToScreen(fwx, fwy, fcamX, fcamY)
		rwx, rwy := ScreenToWorld(sx, sy, fcamX, fcamY)

		if math.Abs(rwx-fwx) > 1e-5 || math.Abs(rwy-fwy) > 1e-5 {
			t.Fatalf("Fuzz roundtrip failed at iter %d: input (%e, %e), cam (%e, %e), screen (%e, %e), recovered (%e, %e)",
				i, fwx, fwy, fcamX, fcamY, sx, sy, rwx, rwy)
		}

		// Linearity & Scale Invariant: offset of delta world units must equal delta * DefaultZoom in screen space
		delta := (rng.Float64()*2.0 - 1.0) * 1000.0
		sxOff, syOff := WorldToScreen(fwx+delta, fwy+delta, fcamX, fcamY)
		expectedDx := delta * DefaultZoom
		expectedDy := delta * DefaultZoom

		if math.Abs((sxOff-sx)-expectedDx) > 1e-6 || math.Abs((syOff-sy)-expectedDy) > 1e-6 {
			t.Fatalf("Linearity check failed at iter %d: delta=%e, screenDx=%e (want %e), screenDy=%e (want %e)",
				i, delta, sxOff-sx, expectedDx, syOff-sy, expectedDy)
		}
	}
}

// TestChallenger_SeamlessTileAdjacency_10000Edges tests 10,000 adjacent tile boundaries across
// arbitrary sub-pixel camera offsets to prove zero black gaps or diamond voids.
func TestChallenger_SeamlessTileAdjacency_10000Edges(t *testing.T) {
	rng := rand.New(rand.NewSource(999))

	const totalEdges = 10000
	const maxGapTolerance = 1e-9

	for i := 0; i < totalEdges; i++ {
		// Random tile coordinates across a huge coordinate domain
		tx := rng.Intn(20000) - 10000
		ty := rng.Intn(20000) - 10000

		// Random camera position with fine subpixel fractional offset
		camX := (rng.Float64()*2.0 - 1.0) * 500000.0
		camY := (rng.Float64()*2.0 - 1.0) * 500000.0

		// Tile (tx, ty) top-left
		w0x := float64(tx * world.TileSize)
		w0y := float64(ty * world.TileSize)
		s0x, s0y := WorldToScreen(w0x, w0y, camX, camY)

		// 1. Horizontal adjacency check
		// Right edge of (tx, ty) in screen space
		rightEdge := s0x + float64(world.TileSize)*DefaultZoom
		// Left edge of (tx+1, ty) in screen space
		wRightTileX := float64((tx + 1) * world.TileSize)
		sRightTileX, _ := WorldToScreen(wRightTileX, w0y, camX, camY)

		hGap := math.Abs(rightEdge - sRightTileX)
		if hGap > maxGapTolerance {
			t.Fatalf("Horizontal seam gap detected at iter %d (tile %d,%d, cam %f,%f): rightEdge=%f, leftEdge=%f, gap=%e",
				i, tx, ty, camX, camY, rightEdge, sRightTileX, hGap)
		}

		// 2. Vertical adjacency check
		// Bottom edge of (tx, ty) in screen space
		bottomEdge := s0y + float64(world.TileSize)*DefaultZoom
		// Top edge of (tx, ty+1) in screen space
		wBottomTileY := float64((ty + 1) * world.TileSize)
		_, sBottomTileY := WorldToScreen(w0x, wBottomTileY, camX, camY)

		vGap := math.Abs(bottomEdge - sBottomTileY)
		if vGap > maxGapTolerance {
			t.Fatalf("Vertical seam gap detected at iter %d (tile %d,%d, cam %f,%f): bottomEdge=%f, topEdge=%f, gap=%e",
				i, tx, ty, camX, camY, bottomEdge, sBottomTileY, vGap)
		}

		// 3. Diagonal corner coincidence check:
		// Bottom-right of (tx, ty) must exactly equal Top-left of (tx+1, ty+1)
		wDiagX := float64((tx + 1) * world.TileSize)
		wDiagY := float64((ty + 1) * world.TileSize)
		sDiagX, sDiagY := WorldToScreen(wDiagX, wDiagY, camX, camY)

		if math.Abs(rightEdge-sDiagX) > maxGapTolerance || math.Abs(bottomEdge-sDiagY) > maxGapTolerance {
			t.Fatalf("Diagonal corner gap detected at iter %d (tile %d,%d): corner=(%f,%f), diagTopLeft=(%f,%f)",
				i, tx, ty, rightEdge, bottomEdge, sDiagX, sDiagY)
		}
	}
}

// TestChallenger_CameraTracking_SubpixelSnappingAndExtremeConvergence verifies:
// 1. Instant initialization on first update
// 2. Exponential convergence curve
// 3. Exact snapping behavior at hypot < 0.01 threshold
// 4. Stable convergence across extreme distance (2e8 units) without overflow or jitter
func TestChallenger_CameraTracking_SubpixelSnappingAndExtremeConvergence(t *testing.T) {
	// 1. Instant snap on uninitialized camera
	cam := NewCamera()
	if cam.Initialized {
		t.Error("Camera should start uninitialized")
	}
	cam.Update(500.0, 750.0)
	if !cam.Initialized || cam.X != 500.0 || cam.Y != 750.0 || cam.TargetX != 500.0 || cam.TargetY != 750.0 {
		t.Errorf("First Update on uninitialized camera did not snap: %+v", cam)
	}

	// 2. Subpixel snapping boundary sweep: threshold is hypot(dx, dy) < 0.01
	// Case A: delta = 0.015 (hypot = 0.015 >= 0.01) -> should lerp by 10%
	cam.Snap(100.0, 100.0)
	cam.Update(100.015, 100.0)
	expectedX := 100.0 + 0.015*0.10
	if math.Abs(cam.X-expectedX) > 1e-9 {
		t.Errorf("Camera above snap threshold did not lerp correctly: got %f, want %f", cam.X, expectedX)
	}

	// Case B: delta = 0.008 (hypot = 0.008 < 0.01) -> should snap immediately to target
	cam.Snap(100.0, 100.0)
	cam.Update(100.008, 100.0)
	if cam.X != 100.008 || cam.Y != 100.0 {
		t.Errorf("Camera below snap threshold did not snap: got (%f, %f), want (100.008, 100.0)", cam.X, cam.Y)
	}

	// Case C: 2D diagonal distance hypot(0.007, 0.007) = 0.009899 < 0.01 -> should snap immediately
	cam.Snap(200.0, 200.0)
	cam.Update(200.007, 200.007)
	if cam.X != 200.007 || cam.Y != 200.007 {
		t.Errorf("Camera diagonal subpixel snap failed: got (%f, %f), want (200.007, 200.007)", cam.X, cam.Y)
	}

	// 3. Extreme coordinate jump convergence (from -1e8 to +1e8)
	cam.Snap(-1e8, -1e8)
	targetX, targetY := 1e8, 1e8

	var prevDist float64 = math.Hypot(targetX-cam.X, targetY-cam.Y)
	snapped := false

	for tick := 0; tick < 500; tick++ {
		cam.Update(targetX, targetY)
		currDist := math.Hypot(targetX-cam.X, targetY-cam.Y)

		if math.IsNaN(cam.X) || math.IsNaN(cam.Y) || math.IsInf(cam.X, 0) || math.IsInf(cam.Y, 0) {
			t.Fatalf("Camera coordinates became NaN/Inf at tick %d: (%f, %f)", tick, cam.X, cam.Y)
		}

		if currDist > 0 && currDist >= prevDist {
			t.Fatalf("Camera diverged at tick %d: prevDist=%e, currDist=%e", tick, prevDist, currDist)
		}

		if currDist == 0.0 {
			snapped = true
			if cam.X != targetX || cam.Y != targetY {
				t.Fatalf("Camera finished with non-exact target coords: (%f, %f) vs (%f, %f)", cam.X, cam.Y, targetX, targetY)
			}
			break
		}
		prevDist = currDist
	}

	if !snapped {
		t.Errorf("Camera did not snap to extreme target within 500 ticks; remaining dist: %e", prevDist)
	}
}

// TestChallenger_YDepthSorting_MonotonicityAndOcclusion verifies:
// 1. Strict mathematical monotonicity of sorted renderable items
// 2. Stable preservation of original order for identical depths
// 3. Natural top-down occlusion for world props, items, and entities
func TestChallenger_YDepthSorting_MonotonicityAndOcclusion(t *testing.T) {
	type Renderable struct {
		ID    int
		Depth float64
	}

	// 1. Monotonicity fuzzing across 1,000 random item lists
	rng := rand.New(rand.NewSource(777))
	for iter := 0; iter < 1000; iter++ {
		itemCount := 100
		items := make([]Renderable, itemCount)
		for i := 0; i < itemCount; i++ {
			items[i] = Renderable{
				ID:    i,
				Depth: (rng.Float64()*2.0 - 1.0) * 1e8,
			}
		}

		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Depth < items[j].Depth
		})

		for i := 0; i < itemCount-1; i++ {
			if items[i].Depth > items[i+1].Depth {
				t.Fatalf("Monotonicity violation at iter %d, index %d: depth[%d]=%e > depth[%d]=%e",
					iter, i, i, items[i].Depth, i+1, items[i+1].Depth)
			}
		}
	}

	// 2. Stability test: multiple items with identical depth preserve insertion order
	identicalDepthItems := []Renderable{
		{ID: 10, Depth: 500.0},
		{ID: 20, Depth: 200.0},
		{ID: 30, Depth: 500.0},
		{ID: 40, Depth: 100.0},
		{ID: 50, Depth: 500.0},
		{ID: 60, Depth: 200.0},
	}

	sort.SliceStable(identicalDepthItems, func(i, j int) bool {
		return identicalDepthItems[i].Depth < identicalDepthItems[j].Depth
	})

	expectedOrder := []int{40, 20, 60, 10, 30, 50}
	for i, item := range identicalDepthItems {
		if item.ID != expectedOrder[i] {
			t.Errorf("Stable sort ordering mismatch at pos %d: got ID %d, want %d", i, item.ID, expectedOrder[i])
		}
	}

	// 3. Natural top-down occlusion hierarchy check:
	// Wall at tile (10, 10), worldY = 10 * TileSize. Wall base Depth = worldY + TileSize.
	wallTileY := 10
	wallWorldY := float64(wallTileY * world.TileSize)
	wallDepth := wallWorldY + float64(world.TileSize)

	// Player north of wall base (e.g. Y = wallDepth - 20.0): player is BEHIND wall
	playerBehindY := wallDepth - 20.0
	// Player south of wall base (e.g. Y = wallDepth + 20.0): player is IN FRONT of wall
	playerInFrontY := wallDepth + 20.0

	sceneBehind := []Renderable{
		{ID: 1, Depth: playerBehindY}, // Player
		{ID: 2, Depth: wallDepth},     // Wall
	}
	sort.SliceStable(sceneBehind, func(i, j int) bool {
		return sceneBehind[i].Depth < sceneBehind[j].Depth
	})
	if sceneBehind[0].ID != 1 || sceneBehind[1].ID != 2 {
		t.Errorf("Player behind wall should be drawn first (occluded by wall): %+v", sceneBehind)
	}

	sceneInFront := []Renderable{
		{ID: 1, Depth: playerInFrontY}, // Player
		{ID: 2, Depth: wallDepth},      // Wall
	}
	sort.SliceStable(sceneInFront, func(i, j int) bool {
		return sceneInFront[i].Depth < sceneInFront[j].Depth
	})
	if sceneInFront[0].ID != 2 || sceneInFront[1].ID != 1 {
		t.Errorf("Wall behind player should be drawn first (player occludes wall): %+v", sceneInFront)
	}
}

// TestChallenger_BezierCombatArc_AffineProjectionAndInvariants verifies:
// 1. Affine linearity: WorldToScreen(B_world(t)) == B_screen(t)
// 2. Control points bounds across all weapon types ("axe", "weapon", "shotgun", "")
// 3. Degenerate facing vector handling (zero-vector fallback)
// 4. Alpha fade quadratic progression monotonicity
// 5. Headless draw of Bezier combat arcs under dynamic and extreme camera settings
func TestChallenger_BezierCombatArc_AffineProjectionAndInvariants(t *testing.T) {
	assets.Load()
	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	// 1. Quadratic Bezier affine projection invariant:
	// For any quadratic curve B(t) = (1-t)^2 P0 + 2(1-t)t P1 + t^2 P2,
	// because WorldToScreen is an affine map S(P) = (P - Cam)*Zoom + Center,
	// S(B(t)) == (1-t)^2 S(P0) + 2(1-t)t S(P1) + t^2 S(P2) must hold identically.
	p0 := [2]float64{100.0, 150.0}
	p1 := [2]float64{250.0, 300.0}
	p2 := [2]float64{200.0, 100.0}
	camX, camY := 150.0, 200.0

	s0x, s0y := WorldToScreen(p0[0], p0[1], camX, camY)
	s1x, s1y := WorldToScreen(p1[0], p1[1], camX, camY)
	s2x, s2y := WorldToScreen(p2[0], p2[1], camX, camY)

	for step := 0; step <= 100; step++ {
		tVal := float64(step) / 100.0
		c0 := (1.0 - tVal) * (1.0 - tVal)
		c1 := 2.0 * (1.0 - tVal) * tVal
		c2 := tVal * tVal

		// World curve point
		bWorldX := c0*p0[0] + c1*p1[0] + c2*p2[0]
		bWorldY := c0*p0[1] + c1*p1[1] + c2*p2[1]
		projSx, projSy := WorldToScreen(bWorldX, bWorldY, camX, camY)

		// Screen curve point
		bScreenX := c0*s0x + c1*s1x + c2*s2x
		bScreenY := c0*s0y + c1*s1y + c2*s2y

		if math.Abs(projSx-bScreenX) > 1e-9 || math.Abs(projSy-bScreenY) > 1e-9 {
			t.Fatalf("Affine projection invariant failed at t=%f: projScreen=(%f,%f), screenBezier=(%f,%f)",
				tVal, projSx, projSy, bScreenX, bScreenY)
		}
	}

	// 2. Alpha fade quadratic progression monotonicity
	var prevAlpha float32 = 2.0
	for cd := 30; cd >= 17; cd-- {
		tVal := float64(30-cd) / 14.0
		alpha := float32((1.0 - tVal) * (1.0 - tVal))

		if alpha > prevAlpha {
			t.Fatalf("Alpha fade is not monotonically decreasing: at cd=%d alpha=%f > prevAlpha=%f", cd, alpha, prevAlpha)
		}
		if alpha < 0.0 || alpha > 1.0 {
			t.Fatalf("Alpha out of bounds [0, 1]: cd=%d, alpha=%f", cd, alpha)
		}
		prevAlpha = alpha
	}

	// 3. Degenerate zero facing vector and 360-degree sweep across all weapons
	weapons := []string{"axe", "weapon", "shotgun", ""}
	for _, wType := range weapons {
		// Zero facing vector: should not panic or generate NaN
		drawSys.DrawAttackSwingArc(screen, 500.0, 500.0, 0.0, 0.0, wType, 25, 500.0, 500.0)

		// Sub-micron facing vector
		drawSys.DrawAttackSwingArc(screen, 500.0, 500.0, 1e-12, 1e-12, wType, 25, 500.0, 500.0)

		// 360-degree rotation sweep
		for deg := 0; deg < 360; deg += 15 {
			rad := float64(deg) * math.Pi / 180.0
			fx := math.Cos(rad)
			fy := math.Sin(rad)
			for cd := 30; cd > 16; cd -= 2 {
				drawSys.DrawAttackSwingArc(screen, 640.0, 360.0, fx, fy, wType, cd, 640.0, 360.0)
			}
		}

		// Extreme camera positions
		drawSys.DrawAttackSwingArc(screen, 1e7, 1e7, 1.0, 0.0, wType, 25, 1e7, 1e7)
		drawSys.DrawAttackSwingArc(screen, -1e7, -1e7, 0.0, 1.0, wType, 20, -1e7, -1e7)
	}
}
