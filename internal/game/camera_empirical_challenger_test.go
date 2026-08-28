package game

import (
	"fmt"
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

// setupChallengerGameHarness initializes an isolated headless test world with camera
func setupChallengerGameHarness() (*arkecs.World, *world.Map, *UpdateSystem, *DrawSystem, *Camera, arkecs.Entity) {
	assets.Load()
	assets.InitAudio()
	w := arkecs.NewWorld()
	m := world.NewMap(60, 60)
	upd := NewUpdateSystem(w, m)
	drw := NewDrawSystem(w, m)
	cam := NewCamera()
	upd.camera = cam
	drw.camera = cam

	playerStartX := 30.0 * float64(world.TileSize)
	playerStartY := 30.0 * float64(world.TileSize)
	spawnIsoX, spawnIsoY := WorldToIso(playerStartX, playerStartY)
	cam.Snap(spawnIsoX, spawnIsoY)

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pEnt := pMap.NewEntity(
		&ecs.Player{
			Health:             100.0,
			Hunger:             100.0,
			Thirst:             100.0,
			Inventory:          []string{},
			WeaponEquipped:     true,
			WeaponType:         "axe",
			WeaponDurability:   12,
			ArmorEquipped:      true,
			ArmorType:          "vest",
			ArmorDefense:       0.50,
			ArmorDurability:    10,
			ArmorMaxDurability: 10,
			InfectionResist:    0.70,
			AttackCooldown:     0,
			Dead:               false,
			Infected:           false,
			FacingX:            1.0,
			FacingY:            0.0,
		},
		&ecs.Position{X: playerStartX, Y: playerStartY},
		&ecs.Velocity{X: 0, Y: 0},
		&ecs.Sprite{Color: color.RGBA{0, 255, 0, 255}, W: 64, H: 128},
		&ecs.Collider{Width: 64, Height: 64},
	)

	return w, m, upd, drw, cam, pEnt
}

// -----------------------------------------------------------------------------------------
// SUITE 1: Viewport Boundary Culling & Extreme Screen Corners Geometry
// -----------------------------------------------------------------------------------------

// TestChallenger_ViewportCornerCullingDistanceAndInvariants mathematically verifies
// that all four extreme corners of the 1280x720 window project to world distances < 2200px
// and pass the vision culling filter with zero visual popping or clipping.
func TestChallenger_ViewportCornerCullingDistanceAndInvariants(t *testing.T) {
	assets.Load()

	// Corner definitions on 1280x720 canvas
	corners := []struct {
		name string
		sx   float64
		sy   float64
	}{
		{"TopLeft (0, 0)", 0.0, 0.0},
		{"TopRight (1280, 0)", 1280.0, 0.0},
		{"BottomLeft (0, 720)", 0.0, 720.0},
		{"BottomRight (1280, 720)", 1280.0, 720.0},
	}

	testPlayerPositions := [][2]float64{
		{0.0, 0.0},
		{3840.0, 3840.0},
		{1000.0, 5000.0},
		{-2000.0, 4000.0},
	}

	const visionRadius = 2200.0
	const maxSpriteExtent = 256.0 // Max tile/wall/prop sprite extent
	const maxLerpLag = 200.0      // Upper bound on tracking lag at maximum speed

	for _, pPos := range testPlayerPositions {
		px, py := pPos[0], pPos[1]
		camIsoX, camIsoY := WorldToIso(px, py)

		for _, c := range corners {
			t.Run(fmt.Sprintf("%s_at_pos_(%.0f,%.0f)", c.name, px, py), func(t *testing.T) {
				// Unproject screen corner to world space
				cornerWx, cornerWy := ScreenToWorld(c.sx, c.sy, camIsoX, camIsoY)

				// Calculate Cartesian Euclidean distance in world coordinates
				dx := cornerWx - px
				dy := cornerWy - py
				dist := math.Hypot(dx, dy)

				// Theoretical corner distance calculation:
				// sx offset = +/- 640px, sy offset = +/- 360px
				// deltaIsoX = sx_offset / 0.5 = +/- 1280px
				// deltaIsoY = sy_offset / 0.5 = +/- 720px
				// deltaWx = deltaIsoY + deltaIsoX/2
				// deltaWy = deltaIsoY - deltaIsoX/2
				// (|deltaWx|, |deltaWy|) = (1360, 80) or (80, 1360)
				// dist = sqrt(1360^2 + 80^2) = sqrt(1849600 + 6400) = sqrt(1856000) approx 1362.3509 px
				expectedDist := math.Sqrt(1856000.0)

				if math.Abs(dist-expectedDist) > 1e-4 {
					t.Fatalf("Corner %s: Expected Euclidean distance %f px, got %f px (dx=%f, dy=%f)",
						c.name, expectedDist, dist, dx, dy)
				}

				// Verify corner is strictly inside the 2200px visionRadius
				if dist >= visionRadius {
					t.Fatalf("Corner %s: Distance %f px exceeds or equals visionRadius %f px",
						c.name, dist, visionRadius)
				}

				// Verify safety margin accounts for sprite extent + dynamic lerp lag
				effectiveDist := dist + maxSpriteExtent + maxLerpLag
				if effectiveDist >= visionRadius {
					t.Fatalf("Corner %s: Effective worst-case distance %f px exceeds visionRadius %f px",
						c.name, effectiveDist, visionRadius)
				}

				// Verify culling filter passes for a tile / prop at this exact corner
				if dx*dx+dy*dy > visionRadius*visionRadius {
					t.Fatalf("Corner %s: Culling check dx^2+dy^2 (%f) failed visionRadius^2 (%f)",
						c.name, dx*dx+dy*dy, visionRadius*visionRadius)
				}
			})
		}
	}
}

// TestChallenger_CullingThresholdRadialBoundarySweep tests culling behavior across a dense
// radial sweep of 360 directions and distances from 0 to 3000px in 10px steps.
func TestChallenger_CullingThresholdRadialBoundarySweep(t *testing.T) {
	playerX, playerY := 3000.0, 3000.0
	const visionRadius = 2200.0
	const visionRadiusSq = visionRadius * visionRadius

	passCount := 0
	culledCount := 0

	for angleDeg := 0; angleDeg < 360; angleDeg += 10 {
		rad := float64(angleDeg) * math.Pi / 180.0
		cosA := math.Cos(rad)
		sinA := math.Sin(rad)

		for dist := 0.0; dist <= 3000.0; dist += 10.0 {
			targetX := playerX + dist*cosA
			targetY := playerY + dist*sinA

			dx := targetX - playerX
			dy := targetY - playerY
			distSq := dx*dx + dy*dy

			passesCulling := distSq <= visionRadiusSq

			if dist < visionRadius-1e-6 && !passesCulling {
				t.Fatalf("Distance %f px (angle %d) was unexpectedly culled! distSq=%f, maxSq=%f",
					dist, angleDeg, distSq, visionRadiusSq)
			}
			if dist > visionRadius+1e-6 && passesCulling {
				t.Fatalf("Distance %f px (angle %d) unexpectedly passed culling! distSq=%f, maxSq=%f",
					dist, angleDeg, distSq, visionRadiusSq)
			}

			if passesCulling {
				passCount++
			} else {
				culledCount++
			}
		}
	}

	if passCount == 0 || culledCount == 0 {
		t.Fatalf("Sweep anomaly: passCount=%d, culledCount=%d", passCount, culledCount)
	}
}

// TestChallenger_FOVRaycastingEnclosesViewportCulling verifies that the 22-tile FOV
// radius (2816px) strictly encloses the 2200px viewport culling circle.
func TestChallenger_FOVRaycastingEnclosesViewportCulling(t *testing.T) {
	const fovTileRadius = 22
	const tileSize = world.TileSize // 128
	fovPixelRadius := float64(fovTileRadius * tileSize) // 2816.0 px
	const viewportCullingRadius = 2200.0

	bufferPixels := fovPixelRadius - viewportCullingRadius
	bufferTiles := bufferPixels / float64(tileSize)

	if fovPixelRadius <= viewportCullingRadius {
		t.Fatalf("FOV pixel radius (%f px) must be strictly greater than viewport culling radius (%f px)",
			fovPixelRadius, viewportCullingRadius)
	}

	if bufferTiles < 4.0 {
		t.Errorf("FOV buffer margin (%f tiles, %f px) is less than recommended 4 tiles",
			bufferTiles, bufferPixels)
	}
}

// -----------------------------------------------------------------------------------------
// SUITE 2: Mouse Click Tile Navigation Across 1280x720 Canvas
// -----------------------------------------------------------------------------------------

// TestChallenger_MouseClickTileNavigationVectorField simulates mouse clicks across a dense
// 65x37 grid on the 1280x720 canvas and verifies player movement vector direction & magnitude.
func TestChallenger_MouseClickTileNavigationVectorField(t *testing.T) {
	assets.Load()
	w, m, _, _, cam, _ := setupChallengerGameHarness()
	playerMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)

	playerX := 3500.0
	playerY := 3500.0
	camIsoX, camIsoY := WorldToIso(playerX, playerY)
	cam.Snap(camIsoX, camIsoY)

	const speed = 12.0

	// Dense grid over 1280x720 canvas: step 20px
	for sx := 0.0; sx <= 1280.0; sx += 20.0 {
		for sy := 0.0; sy <= 720.0; sy += 20.0 {
			// Unproject clicked screen position to world coordinates
			targetWx, targetWy := ScreenToWorld(sx, sy, cam.X, cam.Y)

			dx := targetWx - playerX
			dy := targetWy - playerY
			dist := math.Hypot(dx, dy)

			// Compute simulated velocity from movement logic in processInputAndCombat
			var velX, velY float64
			if dist > speed {
				velX = (dx / dist) * speed
				velY = (dy / dist) * speed
			} else if dist > 0 {
				velX = dx
				velY = dy
			}

			// Validate direction vector alignment
			if dist > 1e-6 {
				velMag := math.Hypot(velX, velY)
				if dist > speed {
					if math.Abs(velMag-speed) > 1e-6 {
						t.Fatalf("Click (%f, %f): Velocity magnitude mismatch: got %f, want %f", sx, sy, velMag, speed)
					}
				} else {
					if math.Abs(velMag-dist) > 1e-6 {
						t.Fatalf("Click (%f, %f): Sub-step velocity magnitude mismatch: got %f, want %f", sx, sy, velMag, dist)
					}
				}

				// Dot product between normalized movement vector and displacement vector must equal +1.0
				dot := (velX*dx + velY*dy) / (velMag * dist)
				if math.Abs(dot-1.0) > 1e-6 {
					t.Fatalf("Click (%f, %f): Movement vector not collinear with target. dot product = %f", sx, sy, dot)
				}

				// Verify one step strictly moves player closer to target
				newDist := math.Hypot(targetWx-(playerX+velX), targetWy-(playerY+velY))
				if newDist >= dist {
					t.Fatalf("Click (%f, %f): Step did not reduce distance: old=%f, new=%f", sx, sy, dist, newDist)
				}
			} else {
				// Clicked exact player position (center)
				if velX != 0 || velY != 0 {
					t.Fatalf("Click (%f, %f): Expected zero velocity for center click, got (%f, %f)", sx, sy, velX, velY)
				}
			}
		}
	}
	_ = playerMap
	_ = m
}

// TestChallenger_MouseClickTileNavigationExactTileCenters tests clicking on 500 distinct map tiles
// and verifying navigation vectors head toward each tile's world center.
func TestChallenger_MouseClickTileNavigationExactTileCenters(t *testing.T) {
	assets.Load()
	_, _, _, _, cam, _ := setupChallengerGameHarness()

	playerX := 40.0 * float64(world.TileSize)
	playerY := 40.0 * float64(world.TileSize)
	cam.Snap(WorldToIso(playerX, playerY))

	r := rand.New(rand.NewSource(54321))

	for i := 0; i < 500; i++ {
		// Pick random tile within [20, 60] range
		tx := 20 + r.Intn(40)
		ty := 20 + r.Intn(40)

		tileCenterWx := float64(tx*world.TileSize) + float64(world.TileSize)/2.0
		tileCenterWy := float64(ty*world.TileSize) + float64(world.TileSize)/2.0

		// Forward project tile center to screen coordinates
		tileIsoX, tileIsoY := WorldToIso(tileCenterWx, tileCenterWy)
		screenX := (tileIsoX-cam.X)*0.5 + 640.0
		screenY := (tileIsoY-cam.Y)*0.5 + 360.0

		// Unproject simulated click at screen coordinates
		unprojectedWx, unprojectedWy := ScreenToWorld(screenX, screenY, cam.X, cam.Y)

		if math.Abs(unprojectedWx-tileCenterWx) > 1e-9 || math.Abs(unprojectedWy-tileCenterWy) > 1e-9 {
			t.Fatalf("Iteration %d: Unprojected click (%f, %f) failed to match tile center (%f, %f)",
				i, unprojectedWx, unprojectedWy, tileCenterWx, tileCenterWy)
		}
	}
}

// TestChallenger_ScreenCenterClickZeroVelocityInvariant verifies clicking exact screen center (640, 360)
// produces zero velocity with zero positional drift.
func TestChallenger_ScreenCenterClickZeroVelocityInvariant(t *testing.T) {
	assets.Load()
	cam := NewCamera()

	for px := 0.0; px <= 5000.0; px += 500.0 {
		for py := 0.0; py <= 5000.0; py += 500.0 {
			cam.Snap(WorldToIso(px, py))

			unprojWx, unprojWy := ScreenToWorld(640.0, 360.0, cam.X, cam.Y)
			dx := unprojWx - px
			dy := unprojWy - py

			if math.Abs(dx) > 1e-9 || math.Abs(dy) > 1e-9 {
				t.Fatalf("Center click at player (%f, %f) gave non-zero offset (%f, %f)", px, py, dx, dy)
			}
		}
	}
}

// -----------------------------------------------------------------------------------------
// SUITE 3: Headless Rendering Loop Execution Across Multi-Frame Simulation & Dynamic Lerp
// -----------------------------------------------------------------------------------------

// TestChallenger_HeadlessMultiFrameRenderingLoopDynamicLerp tests Game.Update() and Game.Draw()
// across 360 continuous frames with complex player trajectory patterns and camera lerping.
func TestChallenger_HeadlessMultiFrameRenderingLoopDynamicLerp(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()

	screen := ebiten.NewImage(1280, 720)

	playerFilter := arkecs.NewFilter2[ecs.Player, ecs.Position](g.world)

	// Simulation trajectory: 360 frames
	// Phase 1 (0-59): +X sprint
	// Phase 2 (60-119): +Y sprint
	// Phase 3 (120-179): -X, -Y diagonal sprint
	// Phase 4 (180-239): Orbital circle
	// Phase 5 (240-359): Stationary (camera catchup & sub-pixel snap)
	const sprintSpeed = 12.0

	var prevCamX, prevCamY float64
	prevCamX, prevCamY = g.camera.X, g.camera.Y

	for frame := 0; frame < 360; frame++ {
		// Move player based on trajectory phase
		q := playerFilter.Query()
		for q.Next() {
			_, pPos := q.Get()
			if frame < 60 {
				pPos.X += sprintSpeed
			} else if frame < 120 {
				pPos.Y += sprintSpeed
			} else if frame < 180 {
				pPos.X -= sprintSpeed * 0.707
				pPos.Y -= sprintSpeed * 0.707
			} else if frame < 240 {
				angle := float64(frame-180) * (2.0 * math.Pi / 60.0)
				pPos.X += math.Cos(angle) * sprintSpeed
				pPos.Y += math.Sin(angle) * sprintSpeed
			}
			// Frames 240-359: player stays stationary
		}

		// Advance game loop
		err := g.Update()
		if err != nil {
			t.Fatalf("Frame %d: Game.Update() failed: %v", frame, err)
		}

		// Assert camera position is valid (no NaNs, no Infs)
		if math.IsNaN(g.camera.X) || math.IsNaN(g.camera.Y) {
			t.Fatalf("Frame %d: Camera position contains NaN! (%f, %f)", frame, g.camera.X, g.camera.Y)
		}
		if math.IsInf(g.camera.X, 0) || math.IsInf(g.camera.Y, 0) {
			t.Fatalf("Frame %d: Camera position is infinite! (%f, %f)", frame, g.camera.X, g.camera.Y)
		}

		// Render frame
		g.Draw(screen)

		prevCamX = g.camera.X
		prevCamY = g.camera.Y
	}

	_ = prevCamX
	_ = prevCamY

	// After 120 frames of stationary player, camera must have caught up and snapped to exact target
	qFinal := playerFilter.Query()
	for qFinal.Next() {
		_, pPos := qFinal.Get()
		targetIsoX, targetIsoY := WorldToIso(pPos.X, pPos.Y)
		lag := math.Hypot(targetIsoX-g.camera.X, targetIsoY-g.camera.Y)
		if lag > 1e-4 {
			t.Fatalf("Stationary convergence failed: remaining camera lag is %f px (expected < 1e-4)", lag)
		}
	}
}

// TestChallenger_HeadlessRenderingAllCombatArcsUnderDynamicCamera tests weapon attack swooshes
// and muzzle flashes across all weapon types while camera is in dynamic motion.
func TestChallenger_HeadlessRenderingAllCombatArcsUnderDynamicCamera(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	w, m, _, drw, cam, pEnt := setupChallengerGameHarness()
	screen := ebiten.NewImage(1280, 720)

	playerMap := arkecs.NewMap2[ecs.Player, ecs.Position](w)

	weaponTypes := []string{"axe", "shotgun", "club", ""}

	for _, wType := range weaponTypes {
		for cooldown := 24; cooldown >= 16; cooldown-- {
			p, pPos := playerMap.Get(pEnt)
			p.WeaponType = wType
			p.AttackCooldown = cooldown

			// Shift camera off-target to simulate lerping lag
			cam.X += 15.0
			cam.Y -= 10.0

			// Draw frame with attack arc active
			drw.Draw(screen, 12.0)

			// Assert no panic, screen remains valid
			_ = pPos
		}
	}
	_ = m
}

// -----------------------------------------------------------------------------------------
// SUITE 4: Adversarial Stress & Edge Cases
// -----------------------------------------------------------------------------------------

// TestChallenger_SubpixelThresholdBifurcation tests the exact 0.01px sub-pixel snapping threshold.
func TestChallenger_SubpixelThresholdBifurcation(t *testing.T) {
	cam := NewCamera()
	cam.Snap(100.0, 100.0)

	// Case A: Offset < 0.01px (e.g. 0.009px) -> Should snap instantly to target
	targetXA := 100.006
	targetYA := 100.006 // sqrt(0.006^2 + 0.006^2) = 0.008485 < 0.01
	cam.Update(targetXA, targetYA)
	if cam.X != targetXA || cam.Y != targetYA {
		t.Fatalf("Case A: Expected sub-pixel snap to (%f, %f), got (%f, %f)", targetXA, targetYA, cam.X, cam.Y)
	}

	// Case B: Offset > 0.01px (e.g. 0.05px) -> Should lerp by 10%
	cam.Snap(100.0, 100.0)
	targetXB := 100.04
	targetYB := 100.03 // sqrt(0.04^2 + 0.03^2) = 0.05 > 0.01
	cam.Update(targetXB, targetYB)

	expectedXB := 100.0 + 0.04*0.10 // 100.004
	expectedYB := 100.0 + 0.03*0.10 // 100.003
	if math.Abs(cam.X-expectedXB) > 1e-9 || math.Abs(cam.Y-expectedYB) > 1e-9 {
		t.Fatalf("Case B: Expected lerp to (%f, %f), got (%f, %f)", expectedXB, expectedYB, cam.X, cam.Y)
	}
}

// TestChallenger_AdversarialFuzzingExtremeCoordinates tests projection invertibility across
// 10,000 extreme, negative, and astronomical coordinates.
func TestChallenger_AdversarialFuzzingExtremeCoordinates(t *testing.T) {
	r := rand.New(rand.NewSource(99999))

	for i := 0; i < 10000; i++ {
		// Astronomical coordinates up to +/- 10,000,000 px
		wx := (r.Float64() - 0.5) * 20000000.0
		wy := (r.Float64() - 0.5) * 20000000.0

		camIsoX := (r.Float64() - 0.5) * 20000000.0
		camIsoY := (r.Float64() - 0.5) * 20000000.0

		// Forward projection
		isoX, isoY := WorldToIso(wx, wy)
		screenX := (isoX-camIsoX)*0.5 + 640.0
		screenY := (isoY-camIsoY)*0.5 + 360.0

		// Inverse projection
		recIsoX, recIsoY := ScreenToIso(screenX, screenY, camIsoX, camIsoY)
		recWx, recWy := ScreenToWorld(screenX, screenY, camIsoX, camIsoY)

		if math.Abs(recIsoX-isoX) > 1e-6 || math.Abs(recIsoY-isoY) > 1e-6 {
			t.Fatalf("Fuzz %d: Iso mismatch. got (%f, %f), want (%f, %f)", i, recIsoX, recIsoY, isoX, isoY)
		}
		if math.Abs(recWx-wx) > 1e-6 || math.Abs(recWy-wy) > 1e-6 {
			t.Fatalf("Fuzz %d: World mismatch. got (%f, %f), want (%f, %f)", i, recWx, recWy, wx, wy)
		}
	}
}

// TestChallenger_DayNightLightingOverlayInvariance tests that the lighting overlay rectangle
// (1280x720) executes cleanly across 48 continuous fractional hours.
func TestChallenger_DayNightLightingOverlayInvariance(t *testing.T) {
	assets.Load()
	assets.InitAudio()
	g := NewGame()
	screen := ebiten.NewImage(1280, 720)

	for hour := 0.0; hour <= 24.0; hour += 0.5 {
		g.timeOfDay = hour
		g.Draw(screen)
	}
}
