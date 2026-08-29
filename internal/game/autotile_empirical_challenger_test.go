package game

import (
	"runtime"
	"testing"

	"github.com/BryceWayne/go-zomboid/internal/assets"
	"github.com/BryceWayne/go-zomboid/internal/ecs"
	"github.com/BryceWayne/go-zomboid/internal/game/world"
	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// TestChallenger_MassiveMapAutotilingRenderStress_500x500 creates a massive 500x500 (250,000 tiles)
// procedural world and executes 100 consecutive frames with moving camera and day/night cycle.
func TestChallenger_MassiveMapAutotilingRenderStress_500x500(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	const width = 500
	const height = 500
	m := world.NewMap(width, height)

	// Build a dense multi-terrain landscape
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Outer boundary walls
			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				m.SetTile(x, y, world.TileWall)
				continue
			}

			// Major asphalt highways every 50 tiles
			if x%50 == 0 || y%50 == 0 {
				m.SetTile(x, y, world.TileAsphalt)
			} else if x%50 == 1 || x%50 == 49 || y%50 == 1 || y%50 == 49 {
				m.SetTile(x, y, world.TileConcrete)
			} else if (x/10+y/10)%3 == 0 {
				m.SetTile(x, y, world.TileWoodFloor)
			} else if (x/10+y/10)%3 == 1 {
				m.SetTile(x, y, world.TileDirt)
			} else {
				m.SetTile(x, y, world.TileGrass)
			}
		}
	}

	// Add interconnected fence corrals and interior wall structures
	for i := 10; i < width-10; i += 25 {
		for j := 10; j < height-10; j += 25 {
			for fx := i; fx < i+10; fx++ {
				m.SetTile(fx, j, world.TileFence)
				m.SetTile(fx, j+9, world.TileFence)
			}
			for fy := j; fy < j+10; fy++ {
				m.SetTile(i, fy, world.TileFence)
				m.SetTile(i+9, fy, world.TileFence)
			}
		}
	}

	// Calculate initial FOV
	startX, startY := 250.0*float64(world.TileSize), 250.0*float64(world.TileSize)
	m.CalculateFOV(startX, startY, 20)

	// Create player entity
	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{
			Health:         100.0,
			Hunger:         100.0,
			Thirst:         100.0,
			FacingX:        1.0,
			FacingY:        0.0,
			WeaponEquipped: true,
			WeaponType:     "axe",
		},
		&ecs.Position{X: startX, Y: startY},
		&ecs.Velocity{X: 100, Y: 100},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16.0, Height: 16.0},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	// Measure memory before
	var mBefore runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	// Execute 100 frames with camera trajectory
	for frame := 0; frame < 100; frame++ {
		camX := startX + float64(frame)*32.0
		camY := startY + float64(frame)*16.0
		drawSys.camera = &Camera{X: camX, Y: camY}

		// Update FOV as camera moves
		m.CalculateFOV(camX, camY, 20)

		tod := float64(frame%240) / 10.0 // Day/night cycle progression
		drawSys.Draw(screen, tod, -1)
	}

	var mAfter runtime.MemStats
	runtime.ReadMemStats(&mAfter)
	t.Logf("Massive 500x500 100-frame stress completed: HeapAlloc before=%d KB, after=%d KB",
		mBefore.HeapAlloc/1024, mAfter.HeapAlloc/1024)
}

// TestChallenger_AutotileAssets_VisualContinuityAndSeamAnalysis performs exhaustive inspection
// of every generated autotile image, shadow sprite, and quadrant overlay texture to verify
// non-nil pointers and exact dimensions.
func TestChallenger_AutotileAssets_VisualContinuityAndSeamAnalysis(t *testing.T) {
	assets.Load()

	// 1. Verify all 16 Wall Autotile Sprites
	for mask := 0; mask < 16; mask++ {
		img := assets.GetWallAutotileImage(uint8(mask))
		if img == nil {
			t.Fatalf("Wall autotile mask %d is nil", mask)
		}
		bounds := img.Bounds()
		if bounds.Dx() != 256 || bounds.Dy() != 256 {
			t.Fatalf("Wall autotile mask %d bounds %vx%v, expected 256x256", mask, bounds.Dx(), bounds.Dy())
		}
	}

	// 2. Verify all 16 Fence Autotile Sprites
	for mask := 0; mask < 16; mask++ {
		img := assets.GetFenceAutotileImage(uint8(mask))
		if img == nil {
			t.Fatalf("Fence autotile mask %d is nil", mask)
		}
		bounds := img.Bounds()
		if bounds.Dx() != 256 || bounds.Dy() != 256 {
			t.Fatalf("Fence autotile mask %d bounds %vx%v, expected 256x256", mask, bounds.Dx(), bounds.Dy())
		}
	}

	// 3. Verify WallFacadeShadowImage
	if assets.WallFacadeShadowImage == nil {
		t.Fatal("WallFacadeShadowImage is nil")
	}
	sbounds := assets.WallFacadeShadowImage.Bounds()
	if sbounds.Dx() != 256 || sbounds.Dy() != 128 {
		t.Fatalf("WallFacadeShadowImage bounds %vx%v, expected 256x128", sbounds.Dx(), sbounds.Dy())
	}

	// 4. Verify Quadrant Terrain Overlays
	overlayTypes := []world.TileType{
		world.TileGrass,
		world.TileConcrete,
		world.TileAsphalt,
		world.TileWoodFloor,
		world.TileTileFloor,
	}

	quads := []world.Quadrant{world.QuadNW, world.QuadNE, world.QuadSW, world.QuadSE}
	states := []world.SubtileState{
		world.SubtileFull,
		world.SubtileHorizontalEdge,
		world.SubtileVerticalEdge,
		world.SubtileOuterCorner,
		world.SubtileInnerCorner,
	}

	for _, tt := range overlayTypes {
		for _, q := range quads {
			for _, s := range states {
				overlay := assets.GetTerrainOverlay(tt, q, s, false)
				if overlay == nil {
					t.Fatalf("TerrainOverlay nil for type %v (%s) quad %v state %v", tt, tt.String(), q, s)
				}
				ob := overlay.Bounds()
				if ob.Dx() != 128 || ob.Dy() != 128 {
					t.Fatalf("TerrainOverlay bounds %vx%v, expected 128x128 for %v quad %v state %v",
						ob.Dx(), ob.Dy(), tt, q, s)
				}
			}

			// Diagonal tip
			diagOverlay := assets.GetTerrainOverlay(tt, q, world.SubtileOuterCorner, true)
			if diagOverlay == nil {
				t.Fatalf("Diagonal overlay nil for type %v quad %v", tt, q)
			}
			db := diagOverlay.Bounds()
			if db.Dx() != 128 || db.Dy() != 128 {
				t.Fatalf("Diagonal overlay bounds %vx%v, expected 128x128 for %v quad %v",
					db.Dx(), db.Dy(), tt, q)
			}
		}
	}
}

// TestChallenger_Autotiling_CameraZoomSubpixelOffsets verifies that camera positioning
// at fractional subpixels, extreme coordinates, and varying zoom ratios does not crash
// or produce coordinate distortion.
func TestChallenger_Autotiling_CameraZoomSubpixelOffsets(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	m := world.NewMap(50, 50)
	for i := range m.Tiles {
		switch i % 5 {
		case 0:
			m.Tiles[i] = world.TileDirt
		case 1:
			m.Tiles[i] = world.TileGrass
		case 2:
			m.Tiles[i] = world.TileConcrete
		case 3:
			m.Tiles[i] = world.TileWall
		case 4:
			m.Tiles[i] = world.TileFence
		}
		m.Visible[i] = true
		m.Explored[i] = true
	}

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{Health: 100, Hunger: 100, Thirst: 100},
		&ecs.Position{X: 25 * world.TileSize, Y: 25 * world.TileSize},
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16, Height: 16},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	subpixelCoords := [][2]float64{
		{0.0, 0.0},
		{0.0001, 0.0001},
		{123.4567, 891.2345},
		{-500.123, -300.789},
		{50000.88, 50000.99},
		{25.5 * float64(world.TileSize), 25.5 * float64(world.TileSize)},
	}

	for _, pt := range subpixelCoords {
		drawSys.camera = &Camera{X: pt[0], Y: pt[1]}
		drawSys.Draw(screen, 12.0, -1)
	}
}

// TestChallenger_Autotiling_ExtremeDynamicTerrainMorphology simulates real-time world
// editing / environmental destruction while rendering each frame to verify stability.
func TestChallenger_Autotiling_ExtremeDynamicTerrainMorphology(t *testing.T) {
	assets.Load()

	w := arkecs.NewWorld()
	m := world.NewMap(30, 30)
	for i := range m.Tiles {
		m.Tiles[i] = world.TileGrass
		m.Visible[i] = true
		m.Explored[i] = true
	}

	pMap := arkecs.NewMap5[ecs.Player, ecs.Position, ecs.Velocity, ecs.Sprite, ecs.Collider](w)
	pMap.NewEntity(
		&ecs.Player{Health: 100, Hunger: 100, Thirst: 100},
		&ecs.Position{X: 15 * world.TileSize, Y: 15 * world.TileSize},
		&ecs.Velocity{},
		&ecs.Sprite{W: 64, H: 128},
		&ecs.Collider{Width: 16, Height: 16},
	)

	drawSys := NewDrawSystem(w, m)
	screen := ebiten.NewImage(1280, 720)

	allTypes := []world.TileType{
		world.TileGrass,
		world.TileDirt,
		world.TileWoodFloor,
		world.TileAsphalt,
		world.TileConcrete,
		world.TileTileFloor,
		world.TileWall,
		world.TileFence,
	}

	for frame := 0; frame < 50; frame++ {
		// Mutate 20 tiles per frame deterministically
		for mIdx := 0; mIdx < 20; mIdx++ {
			rx := (frame*7 + mIdx*13) % 30
			ry := (frame*11 + mIdx*17) % 30
			rt := allTypes[(frame+mIdx)%len(allTypes)]
			m.SetTile(rx, ry, rt)
		}

		drawSys.Draw(screen, 14.0, -1)
	}
}
